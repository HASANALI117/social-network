// frontend/src/app/chat/[id]/page.tsx
'use client';

import { useState, useEffect, useCallback } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useUserStore } from '@/store/useUserStore';
import { useRequest } from '@/hooks/useRequest';
import { Message } from '@/types/Message';
import { UserProfile } from '@/types/User'; // For fetching target user details
import { useGlobalWebSocket, MessageCallback } from '@/contexts/GlobalWebSocketContext'; // Added import

import ChatHeader from '@/components/chat/ChatHeader';
import MessageList from '@/components/chat/MessageList';
import MessageInput from '@/components/chat/MessageInput';
import toast from 'react-hot-toast';

const MESSAGES_PER_PAGE = 20;

export default function ChatPage() {
  const params = useParams();
  const router = useRouter();
  const targetUserId = params.id as string;

  const currentUserId = useUserStore((state) => state.user?.id);
  const currentUserAvatarUrl = useUserStore((state) => state.user?.avatar_url);
  const currentUserUsername = useUserStore((state) => state.user?.username);
  const isAuthenticated = useUserStore((state) => state.isAuthenticated);
  const hydrated = useUserStore((state) => state.hydrated);

  const {
    sendMessage: globalSendMessage,
    subscribeToDirectMessages,
    connectionStatus,
  } = useGlobalWebSocket(); // Added global WebSocket context

  const [messages, setMessages] = useState<Message[]>([]);
  const [targetUser, setTargetUser] = useState<UserProfile | null>(null);
  const [isLoadingHistory, setIsLoadingHistory] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [hasMoreMessages, setHasMoreMessages] = useState(true);
  const [offset, setOffset] = useState(0);
  const [error, setError] = useState<string | null>(null);

  const [isSendingMessage, setIsSendingMessage] = useState(false);

  // For checking if chat is allowed (simplified frontend check, backend enforces)
  const [canChat, setCanChat] = useState(true); // Assume true initially, update based on targetUser profile

  const messagesRequest = useRequest<Message[]>();
  const profileRequest = useRequest<UserProfile>();

  useEffect(() => {
    // Ensure Zustand is hydrated before proceeding
    useUserStore.persist.rehydrate();
  }, []);

  // Fetch target user's profile
  useEffect(() => {
    if (targetUserId && hydrated && isAuthenticated) {
      const fetchUserProfile = async () => {
        try {
          const data = await profileRequest.get(`/api/users/${targetUserId}`);
          if (data) {
            setTargetUser(data);
            // Simplified frontend check for chat restrictions
            // Backend will ultimately enforce this.
            if (data.is_private && !data.is_followed && data.id !== currentUserId) {
              // More complex logic might be needed if "followed_by_target_user" is available
              // For now, if private and not followed by current user, assume restricted.
              // setCanChat(false); // This might be too restrictive, rely on backend errors for now
            }
          }
        } catch (error) {
          // Error is handled by the useEffect hook watching profileRequest.error
          console.error('[ChatPage] Error fetching user profile:', error); // Keep one log for actual errors
        }
      };
      fetchUserProfile();
    }
  }, [targetUserId, profileRequest.get, hydrated, isAuthenticated, currentUserId]);

  // Effect to handle profile loading errors
  useEffect(() => {
    if (profileRequest.isLoading === false && !targetUser && profileRequest.error) {
        setError(`Failed to load user: ${profileRequest.error.message}`);
        toast.error(`Error loading user: ${profileRequest.error.message}`);
    }
  }, [profileRequest.isLoading, targetUser, profileRequest.error]);


  const fetchMessageHistory = useCallback(async (currentOffset: number, loadMore = false) => {
    if (!targetUserId || !currentUserId) return;
    if (loadMore) setIsLoadingMore(true);
    else setIsLoadingHistory(true);

    const data = await messagesRequest.get(`/api/messages?targetUserId=${targetUserId}&limit=${MESSAGES_PER_PAGE}&offset=${currentOffset}`);
    if (data) {
      const rawMessages: Message[] = (data as any)?.messages || [];
      const messagesWithAvatars = rawMessages.map(msg => ({
        ...msg,
        sender_avatar_url: msg.sender_id === currentUserId
          ? currentUserAvatarUrl
          : msg.sender_id === targetUser?.id
            ? targetUser?.avatar_url
            : undefined, // Or a default avatar
      }));

      if (loadMore) {
        // New batch is newest-first from API, reverse it to be oldest-first, then prepend.
        setMessages(prev => [...messagesWithAvatars.reverse(), ...prev]);
      } else {
        // Initial load, API returns newest-first, reverse to display oldest-first at top.
        setMessages(messagesWithAvatars.reverse());
      }
      setHasMoreMessages(messagesWithAvatars.length === MESSAGES_PER_PAGE);
      setOffset(currentOffset + messagesWithAvatars.length);
    } else {
      setHasMoreMessages(false);
    }
    if (loadMore) setIsLoadingMore(false);
    else setIsLoadingHistory(false);
  }, [targetUserId, currentUserId, currentUserAvatarUrl, targetUser, messagesRequest.get]);

  // Effect to handle messages loading errors
  useEffect(() => {
    if (messagesRequest.isLoading === false && messages.length === 0 && messagesRequest.error) {
        setError(`Failed to fetch messages: ${messagesRequest.error.message}`);
        toast.error(`Failed to load messages: ${messagesRequest.error.message}`);
    }
  }, [messagesRequest.isLoading, messages, messagesRequest.error]);

  // Initial message load
  useEffect(() => {
    if (targetUserId && currentUserId && hydrated) {
      setMessages([]); // Clear previous messages if targetUserId changes
      setOffset(0);
      setHasMoreMessages(true);
      fetchMessageHistory(0);
    }
  }, [targetUserId, currentUserId, fetchMessageHistory, hydrated]); // fetchMessageHistory depends on messagesRequest.get

  // Subscribe to messages via Global WebSocket Context
  useEffect(() => {
    if (!currentUserId || !targetUserId || !hydrated || !isAuthenticated) {
      return;
    }

    const handleNewMessage: MessageCallback = (rawMessage) => { // Renamed to rawMessage for clarity
      console.log('Global WebSocket message received in ChatPage:', rawMessage);

      if (
        rawMessage.type === 'direct' &&
        rawMessage.sender_id && // sender_id is on WebSocketMessage
        rawMessage.receiver_id // receiver_id is on WebSocketMessage
      ) {
        // Cast to Message type after confirming it's a direct message
        // Assumes the backend sends a payload compatible with the Message type for direct messages
        const incomingDirectMessage = rawMessage as Message;

        if (
          ((incomingDirectMessage.sender_id === currentUserId && incomingDirectMessage.receiver_id === targetUserId) ||
           (incomingDirectMessage.sender_id === targetUserId && incomingDirectMessage.receiver_id === currentUserId))
        ) {
          const messageWithAvatar: Message = {
            id: incomingDirectMessage.id || Math.random().toString(36).substring(2, 15), // Now uses casted type
            type: 'direct',
            sender_id: incomingDirectMessage.sender_id,
            receiver_id: incomingDirectMessage.receiver_id,
            content: incomingDirectMessage.content || '', // content is on Message type
            created_at: incomingDirectMessage.created_at || new Date().toISOString(), // Now uses casted type
            sender_username: incomingDirectMessage.sender_id === currentUserId ? currentUserUsername : targetUser?.username,
            sender_avatar_url: incomingDirectMessage.sender_id === currentUserId
              ? currentUserAvatarUrl
              : targetUser?.avatar_url,
            // Add other fields if present in incomingDirectMessage and Message type
          };
          setMessages(prevMessages => [...prevMessages, messageWithAvatar]);
        }
      }
    };

    // Using subscribeToDirectMessages as it seems more appropriate for a 1-on-1 chat page.
    // This assumes subscribeToDirectMessages will call the callback for all direct messages
    // and we filter client-side.
    // Alternatively, if subscribeToMessages(chatId, callback) is designed for direct chats
    // where chatId can be targetUserId, that could be used.
    // For now, using the more general direct message subscription and filtering.
    const unsubscribe = subscribeToDirectMessages(handleNewMessage);

    return () => {
      unsubscribe();
    };
  }, [currentUserId, targetUserId, hydrated, isAuthenticated, currentUserAvatarUrl, currentUserUsername, targetUser, subscribeToDirectMessages]);


  const handleSendMessage = (content: string) => {
    if (connectionStatus !== 'connected') {
      toast.error('Not connected to chat server. Please wait or try refreshing.');
      return;
    }
    if (!currentUserId || !targetUserId) {
      toast.error('User information missing.');
      return;
    }

    setIsSendingMessage(true);
    const message: Partial<Message> & { type: string, sender_id: string, receiver_id: string, content: string } = {
      type: 'direct',
      sender_id: currentUserId!,
      receiver_id: targetUserId,
      content: content,
      created_at: new Date().toISOString(),
      sender_username: currentUserUsername!,
      sender_avatar_url: currentUserAvatarUrl
    };

    try {
      globalSendMessage(message);
      // Optimistic UI update can still be done here if desired,
      // though the message will also arrive via the subscription.
      // For simplicity, we'll rely on the subscription to update the message list.
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
    router.push('/login'); // Redirect if not authenticated
    return <div className="flex items-center justify-center min-h-screen bg-gray-900 text-white">Redirecting to login...</div>;
  }
  if (profileRequest.isLoading && !targetUser) {
     return <div className="flex items-center justify-center min-h-screen bg-gray-900 text-white">Loading chat participant...</div>;
  }


  return (
    <div className="flex flex-col h-screen max-w-2xl mx-auto bg-gray-900 text-white">
      <ChatHeader type="direct" target={targetUser} />
      {error && <div className="p-2 text-center text-red-400 bg-red-900">{error}</div>}
      <div className="flex-grow overflow-y-auto">
        <MessageList
          messages={messages}
          currentUserId={currentUserId!} // currentUserId is checked by isAuthenticated logic path
          onLoadMore={handleLoadMore}
          hasMoreMessages={hasMoreMessages}
          isLoadingMore={isLoadingMore || (messagesRequest.isLoading && offset > 0)}
        />
      </div>
      {(isLoadingHistory && messages.length === 0) &&
        <div className="text-center py-4 text-gray-400">Loading messages...</div>
      }
      <MessageInput
        onSendMessage={handleSendMessage}
        isSending={isSendingMessage}
        canSendMessage={canChat} // Use the state variable
      />
    </div>
  );
}