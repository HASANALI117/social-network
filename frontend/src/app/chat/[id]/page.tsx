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
  const [retryCount, setRetryCount] = useState(0);
  const MAX_RETRIES = 3; // Max number of retry attempts
  const RETRY_DELAY = 3000; // Initial delay in ms, increases with each retry

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
  }, [targetUserId, currentUser, targetUser, messagesRequest.get]);

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
    // Ensure all dependencies for connection are met
    if (!currentUser?.id || !targetUserId || !hydrated || !isAuthenticated) {
      console.log('WebSocket connection prerequisites not met (currentUser, targetUserId, hydrated, or isAuthenticated missing).');
      // If there's an existing WebSocket connection, close it as prerequisites are no longer met.
      if (webSocketRef.current) {
        console.log('Closing existing WebSocket due to unmet prerequisites.');
        webSocketRef.current.onclose = null; // Prevent onclose handler from firing during this cleanup
        webSocketRef.current.onerror = null; // Prevent onerror handler from firing
        webSocketRef.current.close();
        webSocketRef.current = null;
        setIsWsConnected(false); // Update connection status
      }
      return; // Do not attempt to connect
    }

    // Function to setup WebSocket connection
    const setupWebSocket = () => {
      const wsScheme = window.location.protocol === "https:" ? "wss:" : "ws:";
      const wsHost = process.env.NEXT_PUBLIC_WEBSOCKET_URL || (window.location.hostname === 'localhost' ? 'localhost:8080' : window.location.host);
      const wsUrl = `${wsScheme}//${wsHost}/ws?id=${currentUser.id}`;

      console.log(`Attempting to connect to WebSocket: ${wsUrl} (Attempt: ${retryCount + 1}/${MAX_RETRIES + 1})`);
      const ws = new WebSocket(wsUrl);
      webSocketRef.current = ws;

      ws.onopen = () => {
        console.log('WebSocket connected successfully.');
        setIsWsConnected(true);
        setRetryCount(0); // Reset retry count on successful connection
        toast.success('Chat connected!');
      };

      ws.onmessage = (event) => {
        try {
          const rawMessageData = JSON.parse(event.data as string) as Message;
          console.log('WebSocket message received:', rawMessageData);

          if (rawMessageData.type === 'direct' &&
              ((rawMessageData.sender_id === currentUser.id && rawMessageData.receiver_id === targetUserId) ||
               (rawMessageData.sender_id === targetUserId && rawMessageData.receiver_id === currentUser.id))) {
            
            const messageWithAvatar: Message = {
              ...rawMessageData,
              sender_avatar_url: rawMessageData.sender_id === currentUser.id
                ? currentUser.avatar_url
                : rawMessageData.sender_id === targetUserId
                  ? targetUser?.avatar_url
                  : undefined,
            };
            setMessages(prevMessages => [...prevMessages, messageWithAvatar]);
          }
        } catch (e) {
          console.error('Error processing WebSocket message:', e);
        }
      };

      ws.onerror = (errorEvent) => {
        console.error('WebSocket error event. Check browser console for more details. Event:', errorEvent);
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
        // Note: Retries for initial connection failures are handled in onerror.
        // This onclose might indicate an established connection was lost.
        // If !closeEvent.wasClean and webSocketRef.current === ws (meaning it's not an old, cleaned-up instance),
        // one might consider further retry logic here, but it can get complex.
        // For now, we rely on onerror for connection retries.
        if (!closeEvent.wasClean && webSocketRef.current === ws) {
             // toast.error('Chat disconnected unexpectedly.'); // Optional: notify user
        }
      };
    };

    // Attempt to setup WebSocket
    setupWebSocket();

    // Cleanup function: This runs when dependencies change or component unmounts.
    return () => {
      if (webSocketRef.current) {
        console.log('Cleaning up WebSocket connection (useEffect cleanup).');
        // Remove event listeners to prevent them from firing after cleanup
        webSocketRef.current.onopen = null;
        webSocketRef.current.onmessage = null;
        webSocketRef.current.onerror = null;
        webSocketRef.current.onclose = null;
        
        webSocketRef.current.close();
        webSocketRef.current = null; // Ensure the ref is cleared
      }
    };
  }, [currentUser, targetUserId, hydrated, isAuthenticated, targetUser, retryCount]); // Added retryCount to dependencies

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