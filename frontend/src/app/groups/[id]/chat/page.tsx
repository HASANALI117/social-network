'use client';

import { useState, useEffect, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useUserStore } from '@/store/useUserStore';
import { useRequest } from '@/hooks/useRequest';
import { Message } from '@/types/Message';
import { Group } from '@/types/Group';
import { UserBasicInfo } from '@/types/User';

import { useGlobalWebSocket, WebSocketMessage } from '@/contexts/GlobalWebSocketContext';
import ChatHeader from '@/components/chat/ChatHeader';
import MessageList from '@/components/chat/MessageList';
import MessageInput from '@/components/chat/MessageInput';
import toast from 'react-hot-toast';

const MESSAGES_PER_PAGE = 20;

// Interface for the actual API response for group members
interface ApiMember {
  id: string;
  username: string;
  first_name: string;
  last_name: string;
  avatar_url?: string;
  // Add other fields if necessary, based on actual API response
}

interface ApiGroupMembersResponse {
  members: ApiMember[];
  count?: number;
}

export default function GroupChatPage() {
  const params = useParams();
  const router = useRouter();
  const groupId = params.id as string;

  const currentUserId = useUserStore((state) => state.user?.id);
  const currentUserAvatarUrl = useUserStore((state) => state.user?.avatar_url);
  const currentUserUsername = useUserStore((state) => state.user?.username);
  const isAuthenticated = useUserStore((state) => state.isAuthenticated);
  const hydrated = useUserStore((state) => state.hydrated);

  const [messages, setMessages] = useState<Message[]>([]);
  const [group, setGroup] = useState<Group | null>(null);
  const [isLoadingHistory, setIsLoadingHistory] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [hasMoreMessages, setHasMoreMessages] = useState(true);
  const [offset, setOffset] = useState(0);
  const [error, setError] = useState<string | null>(null);
  const [groupMembers, setGroupMembers] = useState<Map<string, UserBasicInfo>>(new Map());
  const [isLoadingMembers, setIsLoadingMembers] = useState(false);
  const [onlineMemberCount, setOnlineMemberCount] = useState<number | undefined>(undefined);
  const [totalMemberCount, setTotalMemberCount] = useState<number | undefined>(undefined);

  const [isSendingMessage, setIsSendingMessage] = useState(false);

  const messagesRequest = useRequest<Message[]>();
  const groupRequest = useRequest<Group>();
  const membersRequest = useRequest<ApiGroupMembersResponse>(); // Use the new interface
  const { onlineUserIds, sendMessage, subscribeToGroupMessages, connectionStatus } = useGlobalWebSocket();

  useEffect(() => {
    useUserStore.persist.rehydrate();
  }, []);

  // Fetch group details
  useEffect(() => {
    if (groupId && hydrated && isAuthenticated) {
      const fetchGroupDetails = async () => {
        const data = await groupRequest.get(`/api/groups/${groupId}`);
        if (data) {
          setGroup(data);
        }
      };
      fetchGroupDetails();
    }
  }, [groupId, groupRequest.get, hydrated, isAuthenticated]);

  // Fetch group members
  useEffect(() => {
    if (groupId && hydrated && isAuthenticated) {
      const fetchGroupMembers = async () => {
        setIsLoadingMembers(true);
        const data = await membersRequest.get(`/api/groups/${groupId}/members`);
        console.log('Group members API response:', data); // Log API response
        if (data && data.members) {
          const membersMap = new Map<string, UserBasicInfo>();
          data.members.forEach(apiMember => {
            const userBasic: UserBasicInfo = {
              user_id: apiMember.id, // Map API's id to UserBasicInfo's user_id
              username: apiMember.username,
              first_name: apiMember.first_name,
              last_name: apiMember.last_name,
              avatar_url: apiMember.avatar_url,
            };
            membersMap.set(apiMember.id, userBasic); // Key the map with the API's id
          });
          console.log('Populated groupMembers map:', membersMap); // Log populated map
          setGroupMembers(membersMap);
        } else if (membersRequest.error) {
          toast.error(`Failed to load group members: ${membersRequest.error.message}`);
        }
        setIsLoadingMembers(false);
      };
      fetchGroupMembers();
    }
  }, [groupId, membersRequest.get, hydrated, isAuthenticated]);

  // Effect to calculate online and total member counts
  useEffect(() => {
    if (groupMembers.size > 0) {
      const total = groupMembers.size;
      setTotalMemberCount(total);

      let onlineCount = 0;
      groupMembers.forEach(member => {
        if (onlineUserIds.includes(member.user_id)) {
          onlineCount++;
        }
      });
      setOnlineMemberCount(onlineCount);
    } else {
      setTotalMemberCount(undefined);
      setOnlineMemberCount(undefined);
    }
  }, [groupMembers, onlineUserIds]);

  // Effect to handle group loading errors
  useEffect(() => {
    if (groupRequest.isLoading === false && !group && groupRequest.error) {
      setError(`Failed to load group: ${groupRequest.error.message}`);
      toast.error(`Error loading group: ${groupRequest.error.message}`);
    }
  }, [groupRequest.isLoading, group, groupRequest.error]);

  const fetchMessageHistory = useCallback(async (currentOffset: number, loadMore = false) => {
    if (!groupId || !currentUserId || groupMembers.size === 0 && !isLoadingMembers) return; // Wait for members if not loading
    if (loadMore) setIsLoadingMore(true);
    else setIsLoadingHistory(true);

    const data = await messagesRequest.get(`/api/groups/${groupId}/messages?limit=${MESSAGES_PER_PAGE}&offset=${currentOffset}`);
    if (data) {
      const rawMessages: Message[] = (data as any)?.messages || [];
      const messagesWithAvatars = rawMessages.map(msg => {
        const senderInfo = groupMembers.get(msg.sender_id);
        console.log('[History] Looking up sender_id:', msg.sender_id, 'Found info:', senderInfo); // Log historical message enrichment
        return {
          ...msg,
          sender_username: msg.sender_id === currentUserId ? currentUserUsername : senderInfo?.username,
          sender_avatar_url: msg.sender_id === currentUserId
            ? currentUserAvatarUrl
            : senderInfo?.avatar_url
        };
      });

      if (loadMore) {
        setMessages(prev => [...messagesWithAvatars.reverse(), ...prev]);
      } else {
        setMessages(messagesWithAvatars.reverse());
      }
      setHasMoreMessages(messagesWithAvatars.length === MESSAGES_PER_PAGE);
      setOffset(currentOffset + messagesWithAvatars.length);
    } else {
      setHasMoreMessages(false);
    }
    if (loadMore) setIsLoadingMore(false);
    else setIsLoadingHistory(false);
  }, [groupId, currentUserId, currentUserAvatarUrl, currentUserUsername, messagesRequest.get, groupMembers, isLoadingMembers]);

  // Effect to handle messages loading errors
  useEffect(() => {
    if (messagesRequest.isLoading === false && messages.length === 0 && messagesRequest.error) {
      setError(`Failed to fetch messages: ${messagesRequest.error.message}`);
      toast.error(`Failed to load messages: ${messagesRequest.error.message}`);
    }
  }, [messagesRequest.isLoading, messages, messagesRequest.error]);

  // Initial message load
  useEffect(() => {
    if (groupId && currentUserId && hydrated) {
      setMessages([]);
      setOffset(0);
      setHasMoreMessages(true);
      fetchMessageHistory(0);
    }
  }, [groupId, currentUserId, fetchMessageHistory, hydrated, groupMembers]); // Added groupMembers dependency

  // Subscribe to group messages via Global WebSocket Context
  useEffect(() => {
    if (!groupId || !currentUserId || !hydrated || !isAuthenticated || groupMembers.size === 0) {
      return;
    }

    const handleNewMessage = (rawMessage: WebSocketMessage) => {
      console.log('Global WebSocket message received for group chat:', rawMessage);
      // Ensure the message is for this group
      console.log('[GroupChatPage] Comparing for message processing:');
      console.log('[GroupChatPage] rawMessage.type:', rawMessage.type, '(Expected: "group")');
      console.log('[GroupChatPage] rawMessage.receiver_id:', rawMessage.receiver_id);
      console.log('[GroupChatPage] page groupId:', groupId);
      console.log('[GroupChatPage] Comparison result (type === "group"):', rawMessage.type === 'group');
      console.log('[GroupChatPage] Comparison result (receiver_id === groupId):', rawMessage.receiver_id === groupId);

      if (rawMessage.type === 'group' && rawMessage.receiver_id === groupId) {
        console.log('[GroupChatPage] Entered message processing block.');

        // Check for essential fields before processing
        if (rawMessage.sender_id && rawMessage.content && rawMessage.created_at) {
          console.log('[GroupChatPage] rawMessage.sender_id:', rawMessage.sender_id);

          const senderInfo = groupMembers.get(rawMessage.sender_id);
          console.log('[GroupChatPage] senderInfo from groupMembers:', senderInfo);

          console.log('[GroupChatPage] rawMessage.content:', rawMessage.content);
          console.log('[GroupChatPage] rawMessage.created_at:', rawMessage.created_at);
          console.log('[GroupChatPage] rawMessage.id (if expected):', rawMessage.id);

          const messageWithAvatar: Message = {
            id: rawMessage.id || Math.random().toString(36).substring(2, 15),
            type: 'group',
            sender_id: rawMessage.sender_id,
            receiver_id: rawMessage.receiver_id,
            content: rawMessage.content,
            created_at: rawMessage.created_at,
            sender_username: rawMessage.sender_id === currentUserId ? currentUserUsername : senderInfo?.username,
            sender_avatar_url: rawMessage.sender_id === currentUserId
              ? currentUserAvatarUrl
              : senderInfo?.avatar_url,
          };
          console.log('[GroupChatPage] Constructed messageWithAvatar:', messageWithAvatar);

          console.log('[GroupChatPage] About to call setMessages.');
          setMessages(prevMessages => [...prevMessages, messageWithAvatar]);
        } else {
          console.log('[GroupChatPage] Message for this group received, but missing essential fields (sender_id, content, created_at, or id). Message:', rawMessage);
        }
      } else if (rawMessage.type === 'group_message_echo' && rawMessage.group_id === groupId && rawMessage.sender_id === currentUserId) {
        // Handle echo for the sender if needed, or rely on optimistic update
        // For now, we assume optimistic update handles the sender's view,
        // and this echo confirms delivery or provides the server-assigned ID.
        // If not doing optimistic updates, this is where you'd add the message.
        console.log('Received echo for sent group message:', rawMessage);
      }
    };

    const unsubscribe = subscribeToGroupMessages(groupId, handleNewMessage);
    console.log(`Subscribed to group messages for group ID: ${groupId}`);

    return () => {
      if (unsubscribe) {
        unsubscribe();
        console.log(`Unsubscribed from group messages for group ID: ${groupId}`);
      }
    };
  }, [groupId, currentUserId, hydrated, isAuthenticated, groupMembers, subscribeToGroupMessages, currentUserUsername, currentUserAvatarUrl]);


  const handleSendMessage = (content: string) => {
    if (connectionStatus !== 'connected') {
      toast.error('Not connected to chat server. Please wait or try refreshing.');
      return;
    }
    if (!currentUserId || !groupId) {
      toast.error('User or group information missing.');
      return;
    }

    setIsSendingMessage(true);
    const messageToSend = { // Structure according to GlobalWebSocketContext's expected format
      type: 'group', // This should match what GlobalWebSocketContext expects
      sender_id: currentUserId,
      receiver_id: groupId, // Ensure this is populated correctly
      content: content,
      // created_at will likely be set by the server or context
    };

    try {
      sendMessage(messageToSend);
      // If sendMessage is async and returns a promise, you might want to await it
      // and handle success/failure.
    } catch (e) {
      console.error('Failed to send message via Global WebSocket:', e);
      toast.error('Failed to send message.');
    } finally {
      setIsSendingMessage(false);
    }
  };

  const handleLoadMore = () => {
    if (!isLoadingMore && hasMoreMessages) {
      fetchMessageHistory(offset, true);
    }
  };

  if (!hydrated) {
    return <div className="flex items-center justify-center min-h-screen bg-gray-900 text-white">Loading session...</div>;
  }
  if (!isAuthenticated) {
    router.push('/login');
    return <div className="flex items-center justify-center min-h-screen bg-gray-900 text-white">Redirecting to login...</div>;
  }
  if (groupRequest.isLoading && !group) {
    return <div className="flex items-center justify-center min-h-screen bg-gray-900 text-white">Loading group...</div>;
  }
  if (isLoadingMembers && groupMembers.size === 0) {
    return <div className="flex items-center justify-center min-h-screen bg-gray-900 text-white">Loading members...</div>;
  }

  return (
    <div className="flex flex-col h-screen max-w-2xl mx-auto bg-gray-900 text-white">
      <ChatHeader
        type="group"
        target={group}
        onlineMemberCount={onlineMemberCount}
        totalMemberCount={totalMemberCount}
      />
      {error && <div className="p-2 text-center text-red-400 bg-red-900">{error}</div>}
      <div className="flex-grow overflow-y-auto">
        <MessageList
          messages={messages}
          currentUserId={currentUserId!}
          onLoadMore={handleLoadMore}
          hasMoreMessages={hasMoreMessages}
          isLoadingMore={isLoadingMore || (messagesRequest.isLoading && offset > 0)}
          type="group"
          emptyMessage={`No messages in ${group?.name || 'this group'} yet. Start the conversation!`}
        />
      </div>
      {(isLoadingHistory && messages.length === 0) &&
        <div className="text-center py-4 text-gray-400">Loading messages...</div>
      }
      <MessageInput
        onSendMessage={handleSendMessage}
        isSending={isSendingMessage}
        canSendMessage={true}
      />
    </div>
  );
}