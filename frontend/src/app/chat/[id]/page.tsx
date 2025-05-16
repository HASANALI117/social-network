// frontend/src/app/chat/[id]/page.tsx
'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { useUserStore } from '@/store/useUserStore';
import { useRequest } from '@/hooks/useRequest';
import { Message } from '@/types/Message';
import { UserProfile } from '@/types/User'; // For fetching target user details

import ChatHeader from '@/components/chat/ChatHeader';
import MessageList from '@/components/chat/MessageList';
import MessageInput from '@/components/chat/MessageInput';
import toast from 'react-hot-toast';

const MESSAGES_PER_PAGE = 20;

export default function ChatPage() {
  const params = useParams();
  const router = useRouter();
  const targetUserId = params.id as string;

  const currentUser = useUserStore((state) => state.user);
  const isAuthenticated = useUserStore((state) => state.isAuthenticated);
  const hydrated = useUserStore((state) => state.hydrated);

  const [messages, setMessages] = useState<Message[]>([]);
  const [targetUser, setTargetUser] = useState<UserProfile | null>(null);
  const [isLoadingHistory, setIsLoadingHistory] = useState(false);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [hasMoreMessages, setHasMoreMessages] = useState(true);
  const [offset, setOffset] = useState(0);
  const [error, setError] = useState<string | null>(null);

  const [isWsConnected, setIsWsConnected] = useState(false);
  const [isSendingMessage, setIsSendingMessage] = useState(false);
  const webSocketRef = useRef<WebSocket | null>(null);

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
        const data = await profileRequest.get(`/api/users/${targetUserId}`);
        if (data) {
          setTargetUser(data);
          // Simplified frontend check for chat restrictions
          // Backend will ultimately enforce this.
          if (data.is_private && !data.is_followed && data.id !== currentUser?.id) {
            // More complex logic might be needed if "followed_by_target_user" is available
            // For now, if private and not followed by current user, assume restricted.
            // setCanChat(false); // This might be too restrictive, rely on backend errors for now
          }
        }
      };
      fetchUserProfile();
    }
  }, [targetUserId, profileRequest.get, hydrated, isAuthenticated, currentUser?.id]);

  // Effect to handle profile loading errors
  useEffect(() => {
    if (profileRequest.isLoading === false && !targetUser && profileRequest.error) {
        setError(`Failed to load user: ${profileRequest.error.message}`);
        toast.error(`Error loading user: ${profileRequest.error.message}`);
    }
  }, [profileRequest.isLoading, targetUser, profileRequest.error]);


  const fetchMessageHistory = useCallback(async (currentOffset: number, loadMore = false) => {
    if (!targetUserId || !currentUser?.id) return;
    if (loadMore) setIsLoadingMore(true);
    else setIsLoadingHistory(true);

    const data = await messagesRequest.get(`/api/messages?targetUserId=${targetUserId}&limit=${MESSAGES_PER_PAGE}&offset=${currentOffset}`);
    if (data) {
      const rawMessages: Message[] = (data as any)?.messages || [];
      const messagesWithAvatars = rawMessages.map(msg => ({
        ...msg,
        sender_avatar_url: msg.sender_id === currentUser?.id
          ? currentUser?.avatar_url
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
  }, [targetUserId, currentUser?.id, messagesRequest.get]);

  // Effect to handle messages loading errors
  useEffect(() => {
    if (messagesRequest.isLoading === false && messages.length === 0 && messagesRequest.error) {
        setError(`Failed to fetch messages: ${messagesRequest.error.message}`);
        toast.error(`Failed to load messages: ${messagesRequest.error.message}`);
    }
  }, [messagesRequest.isLoading, messages, messagesRequest.error]);

  // Initial message load
  useEffect(() => {
    if (targetUserId && currentUser?.id && hydrated) {
      setMessages([]); // Clear previous messages if targetUserId changes
      setOffset(0);
      setHasMoreMessages(true);
      fetchMessageHistory(0);
    }
  }, [targetUserId, currentUser?.id, fetchMessageHistory, hydrated]); // fetchMessageHistory depends on messagesRequest.get

  // WebSocket connection
  useEffect(() => {
    if (!targetUserId || !currentUser?.id || !hydrated || !isAuthenticated) return;

    // Construct WebSocket URL. Adjust if your backend URL is different.
    // Ensure `NEXT_PUBLIC_API_BASE_URL` or similar is configured for WebSocket.
    // For local dev, it might be 'ws://localhost:8080/ws?id='
    const wsScheme = window.location.protocol === "https:" ? "wss:" : "ws:";
    const wsHost = process.env.NEXT_PUBLIC_WEBSOCKET_URL || (window.location.hostname === 'localhost' ? 'localhost:8080' : window.location.host);
    const wsUrl = `${wsScheme}//${wsHost}/ws?id=${currentUser.id}`;


    webSocketRef.current = new WebSocket(wsUrl);
    console.log(`Attempting to connect to WebSocket: ${wsUrl}`);


    webSocketRef.current.onopen = () => {
      console.log('WebSocket connected');
      setIsWsConnected(true);
      toast.success('Chat connected');
    };

    webSocketRef.current.onmessage = (event) => {
      try {
        const rawMessageData = JSON.parse(event.data as string) as Message;
        console.log('WebSocket message received:', rawMessageData);

        // Ensure it's a direct message and relevant to this chat
        if (rawMessageData.type === 'direct' &&
            ((rawMessageData.sender_id === currentUser.id && rawMessageData.receiver_id === targetUserId) ||
             (rawMessageData.sender_id === targetUserId && rawMessageData.receiver_id === currentUser.id))) {
          
          const messageWithAvatar: Message = {
            ...rawMessageData,
            sender_avatar_url: rawMessageData.sender_id === currentUser.id
              ? currentUser.avatar_url
              : rawMessageData.sender_id === targetUserId
                ? targetUser?.avatar_url // targetUser might not be loaded yet, though unlikely for WS message
                : undefined, // Or a default avatar
          };
          // New messages are appended to the end (bottom of the screen)
          setMessages(prevMessages => [...prevMessages, messageWithAvatar]);
        }
      } catch (e) {
        console.error('Error processing WebSocket message:', e);
      }
    };

    webSocketRef.current.onerror = (errorEvent) => {
      console.error('WebSocket error event. Check browser console for more details. Event:', errorEvent);
      setError('WebSocket connection error. Real-time updates may not work.');
      toast.error('Chat connection error.');
      setIsWsConnected(false);
    };

    webSocketRef.current.onclose = (closeEvent) => {
      console.log('WebSocket disconnected:', closeEvent.reason, closeEvent.code);
      setIsWsConnected(false);
      if (!closeEvent.wasClean) {
        // toast.error('Chat disconnected. Attempting to reconnect...');
        // Implement reconnection logic if desired
      }
    };

    return () => {
      if (webSocketRef.current) {
        console.log('Closing WebSocket connection');
        webSocketRef.current.close();
      }
    };
  }, [targetUserId, currentUser?.id, hydrated, isAuthenticated]); // currentUser.id ensures it reconnects if user changes

  const handleSendMessage = (content: string) => {
    if (!webSocketRef.current || webSocketRef.current.readyState !== WebSocket.OPEN) {
      toast.error('Not connected to chat server. Please wait or try refreshing.');
      return;
    }
    if (!currentUser?.id || !targetUserId) {
      toast.error('User information missing.');
      return;
    }

    setIsSendingMessage(true);
    const message: Message = {
      type: 'direct',
      sender_id: currentUser.id,
      receiver_id: targetUserId,
      content: content,
      created_at: new Date().toISOString(),
      // Optional: Add sender_username and sender_avatar_url if available client-side
      // Or rely on backend to populate these if needed for the receiver
      sender_username: currentUser.username,
      sender_avatar_url: currentUser.avatar_url
    };

    try {
      webSocketRef.current.send(JSON.stringify(message));
      // Optimistically add to UI. New messages are appended to the end.
      // The message object already includes sender_avatar_url from currentUser.
      // This will be displayed at the bottom.
      setMessages(prevMessages => [...prevMessages, message]);
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
    router.push('/login'); // Redirect if not authenticated
    return <div className="flex items-center justify-center min-h-screen bg-gray-900 text-white">Redirecting to login...</div>;
  }
  if (profileRequest.isLoading && !targetUser) {
     return <div className="flex items-center justify-center min-h-screen bg-gray-900 text-white">Loading chat participant...</div>;
  }


  return (
    <div className="flex flex-col h-screen max-w-2xl mx-auto bg-gray-900 text-white">
      <ChatHeader targetUser={targetUser} />
      {error && <div className="p-2 text-center text-red-400 bg-red-900">{error}</div>}
      <MessageList
        messages={messages}
        currentUserId={currentUser!.id} // currentUser is checked by isAuthenticated
        onLoadMore={handleLoadMore}
        hasMoreMessages={hasMoreMessages}
        isLoadingMore={isLoadingMore || (messagesRequest.isLoading && offset > 0)}
      />
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