'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useUserStore } from '@/store/useUserStore';
import { useRequest } from '@/hooks/useRequest';
import { Message } from '@/types/Message';
import { Group } from '@/types/Group';
import { UserBasicInfo } from '@/types/User';

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

  const [isWsConnected, setIsWsConnected] = useState(false);
  const [isSendingMessage, setIsSendingMessage] = useState(false);
  const webSocketRef = useRef<WebSocket | null>(null);
  const [retryCount, setRetryCount] = useState(0);
  const MAX_RETRIES = 3;
  const RETRY_DELAY = 3000;

  const messagesRequest = useRequest<Message[]>();
  const groupRequest = useRequest<Group>();
  const membersRequest = useRequest<ApiGroupMembersResponse>(); // Use the new interface

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

  // WebSocket connection
  useEffect(() => {
    if (!currentUserId || !groupId || !hydrated || !isAuthenticated) {
      console.log('WebSocket connection prerequisites not met.');
      if (webSocketRef.current) {
        webSocketRef.current.onclose = null;
        webSocketRef.current.onerror = null;
        webSocketRef.current.close();
        webSocketRef.current = null;
        setIsWsConnected(false);
      }
      return;
    }

    const setupWebSocket = () => {
      const wsScheme = window.location.protocol === "https:" ? "wss:" : "ws:";
      const wsHost = process.env.NEXT_PUBLIC_WEBSOCKET_URL || (window.location.hostname === 'localhost' ? 'localhost:8080' : window.location.host);
      const wsUrl = `${wsScheme}//${wsHost}/ws?id=${currentUserId}`;

      console.log(`Attempting to connect to WebSocket: ${wsUrl} (Attempt: ${retryCount + 1}/${MAX_RETRIES + 1})`);
      const ws = new WebSocket(wsUrl);
      webSocketRef.current = ws;

      ws.onopen = () => {
        console.log('WebSocket connected successfully.');
        setIsWsConnected(true);
        setRetryCount(0);
        toast.success('Chat connected!');
      };

      ws.onmessage = (event) => {
        try {
          const rawMessageData = JSON.parse(event.data as string) as Message;
          console.log('WebSocket message received:', rawMessageData);

          if (rawMessageData.type === 'group' && rawMessageData.receiver_id === groupId) {
            const senderInfo = groupMembers.get(rawMessageData.sender_id);
            console.log('[WebSocket] Looking up sender_id:', rawMessageData.sender_id, 'Found info:', senderInfo); // Log WebSocket message enrichment
            const messageWithAvatar: Message = {
              ...rawMessageData,
              sender_username: rawMessageData.sender_id === currentUserId ? currentUserUsername : senderInfo?.username,
              sender_avatar_url: rawMessageData.sender_id === currentUserId
                ? currentUserAvatarUrl
                : senderInfo?.avatar_url
            };
            setMessages(prevMessages => [...prevMessages, messageWithAvatar]);
          }
        } catch (e) {
          console.error('Error processing WebSocket message:', e);
        }
      };

      ws.onerror = (errorEvent) => {
        console.error('WebSocket error event:', errorEvent);
        setIsWsConnected(false);
        
        if (retryCount < MAX_RETRIES) {
          const delay = RETRY_DELAY * (retryCount + 1);
          console.log(`WebSocket connection error. Retrying in ${delay / 1000}s... (Attempt ${retryCount + 1})`);
          toast.error(`Chat connection error. Retrying (attempt ${retryCount + 1})...`);
          setTimeout(() => {
            setRetryCount(prev => prev + 1);
          }, delay);
        } else {
          console.error('WebSocket connection failed after max retries.');
          setError('Failed to connect to chat server after multiple attempts. Real-time updates disabled.');
          toast.error('Failed to connect to chat after multiple retries.');
        }
      };

      ws.onclose = (closeEvent) => {
        console.log(`WebSocket disconnected. Code: ${closeEvent.code}, Reason: "${closeEvent.reason}", Clean: ${closeEvent.wasClean}`);
        setIsWsConnected(false);
        if (!closeEvent.wasClean && webSocketRef.current === ws) {
          // Handle unclean disconnection if needed
        }
      };
    };

    setupWebSocket();

    return () => {
      if (webSocketRef.current) {
        console.log('Cleaning up WebSocket connection.');
        webSocketRef.current.onopen = null;
        webSocketRef.current.onmessage = null;
        webSocketRef.current.onerror = null;
        webSocketRef.current.onclose = null;
        webSocketRef.current.close();
        webSocketRef.current = null;
      }
    };
  }, [currentUserId, groupId, hydrated, isAuthenticated, currentUserAvatarUrl, currentUserUsername, retryCount, groupMembers]);

  const handleSendMessage = (content: string) => {
    if (!webSocketRef.current || webSocketRef.current.readyState !== WebSocket.OPEN) {
      toast.error('Not connected to chat server. Please wait or try refreshing.');
      return;
    }
    if (!currentUserId || !groupId) {
      toast.error('User information missing.');
      return;
    }

    setIsSendingMessage(true);
    const message: Message = {
      type: 'group',
      sender_id: currentUserId,
      receiver_id: groupId,
      content: content,
      created_at: new Date().toISOString(),
      sender_username: currentUserUsername!,
      sender_avatar_url: currentUserAvatarUrl
    };

    try {
      webSocketRef.current.send(JSON.stringify(message));
    } catch (e) {
      console.error('Failed to send message via WebSocket:', e);
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