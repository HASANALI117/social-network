'use client';

import React, {
  createContext,
  useContext,
  useEffect,
  useRef,
  useState,
  useCallback,
} from 'react';
import { useUserStore } from '@/store/useUserStore';
import { toast } from 'react-hot-toast';
import { usePathname, useRouter } from 'next/navigation';

type ConnectionStatus =
  | 'idle'
  | 'connecting'
  | 'connected'
  | 'disconnected'
  | 'error';

interface GlobalWebSocketContextType {
  connectionStatus: ConnectionStatus;
  onlineUserIds: string[];
  error: string | null;
  lastMessageData: string | null; // Added to expose last raw message data
  messageCount: number; // Added for message count
  connectWebSocket: () => void;
  disconnectWebSocket: () => void;
  clearMessageCount: () => void; // Added to clear message count
  sendMessage: (message: object) => void;
  subscribeToMessages: (chatId: string, callback: MessageCallback) => () => void;
  subscribeToDirectMessages: (callback: MessageCallback) => () => void;
  subscribeToGroupMessages: (
    groupId: string,
    callback: MessageCallback
  ) => () => void;
}

const GlobalWebSocketContext = createContext<
  GlobalWebSocketContextType | undefined
>(undefined);

export const useGlobalWebSocket = () => {
  const context = useContext(GlobalWebSocketContext);
  if (!context) {
    throw new Error(
      'useGlobalWebSocket must be used within a GlobalWebSocketProvider'
    );
  }
  return context;
};

const MAX_RETRIES = 5;
const INITIAL_RETRY_DELAY = 3000; // 3 seconds

export type MessageCallback = (message: WebSocketMessage) => void;

export interface WebSocketMessage { // Add export here
  type: string;
  userIds?: string[];
  // For direct messages
  id?: string; // Added for message identification
  sender_id?: string;
  receiver_id?: string; // For group_id or user_id in direct
  group_id?: string;
  content?: string;
  created_at?: string; // Added for timestamp
  // For notifications
  payload?: {
    id: string;
    user_id: string;
    type: string;
    entity_type: string;
    message: string;
    entity_id: string;
    is_read: boolean;
    created_at: string;
  };
  // For notification_created messages from hub
  data?: {
    id: string;
    user_id: string;
    type: string;
    entity_type: string;
    message: string;
    entity_id: string;
    is_read: boolean;
    created_at: string;
  };
}

export const GlobalWebSocketProvider: React.FC<{
  children: React.ReactNode;
}> = ({ children }) => {
  const webSocketRef = useRef<WebSocket | null>(null);
  const [connectionStatus, setConnectionStatus] =
    useState<ConnectionStatus>('idle');
  const [onlineUserIds, setOnlineUserIds] = useState<string[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [lastMessageData, setLastMessageData] = useState<string | null>(null); // State for last message data
  const [messageCount, setMessageCount] = useState<number>(0); // State for message count
  const [retryCount, setRetryCount] = useState(0);
  const retryTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const subscriptionsRef = useRef<Record<string, MessageCallback[]>>({});

  const currentUserId = useUserStore((state) => state.user?.id);
  const isAuthenticated = useUserStore((state) => state.isAuthenticated);
  const hydrated = useUserStore((state) => state.hydrated);
  const pathname = usePathname();
  const router = useRouter();

  const getWebSocketUrl = useCallback((): string | null => {
    if (!currentUserId) return null;

    const wsBaseUrl = process.env.NEXT_PUBLIC_WEBSOCKET_URL;
    if (!wsBaseUrl) {
      console.error('Global WebSocket: NEXT_PUBLIC_WEBSOCKET_URL is not set.');
      return null;
    }

    // Determine scheme based on browser protocol, but allow override if wsBaseUrl includes it
    let scheme = window.location.protocol === 'https:' ? 'wss' : 'ws';
    // let domainAndPath = wsBaseUrl; // This variable is not used in the provided snippet

    if (wsBaseUrl.startsWith('ws://') || wsBaseUrl.startsWith('wss://')) {
      // If scheme is already in env var, use it directly
      return `${wsBaseUrl}/ws?id=${currentUserId}`;
    }

    // Otherwise, prepend the determined scheme
    return `${scheme}://${wsBaseUrl}/ws?id=${currentUserId}`;
  }, [currentUserId]);

  const disconnectWebSocket = useCallback(() => {
    if (retryTimeoutRef.current) {
      clearTimeout(retryTimeoutRef.current);
      retryTimeoutRef.current = null;
    }
    if (webSocketRef.current) {
      console.log('Global WebSocket: Disconnecting...');
      webSocketRef.current.onclose = null; // Prevent onclose handler from triggering retries
      webSocketRef.current.onerror = null;
      webSocketRef.current.close();
      webSocketRef.current = null;
    }
    setConnectionStatus('disconnected');
    setOnlineUserIds([]);
    setRetryCount(0); // Reset retries on manual disconnect
    // setError(null); // Optionally clear error on disconnect
    console.log('Global WebSocket: Disconnected');
  }, []);

  const clearMessageCount = useCallback(() => {
    setMessageCount(0);
  }, []);

  const connectWebSocket = useCallback(() => {
    if (!hydrated || !isAuthenticated || !currentUserId) {
      console.log(
        'Global WebSocket: Conditions not met for connection (hydrated, isAuthenticated, user.id). Current state:',
        { hydrated, isAuthenticated, userId: currentUserId }
      );
      if (
        webSocketRef.current &&
        webSocketRef.current.readyState === WebSocket.OPEN
      ) {
        disconnectWebSocket();
      }
      return;
    }

    if (
      webSocketRef.current &&
      webSocketRef.current.readyState === WebSocket.OPEN
    ) {
      console.log('Global WebSocket: Already connected or connecting.');
      return;
    }

    if (retryTimeoutRef.current) {
      clearTimeout(retryTimeoutRef.current);
      retryTimeoutRef.current = null;
    }

    const wsUrl = getWebSocketUrl();
    if (!wsUrl) {
      setError('User ID not available to construct WebSocket URL.');
      setConnectionStatus('error');
      return;
    }

    console.log(
      `Global WebSocket: Connecting to ${wsUrl}... (Attempt ${retryCount + 1})`
    );
    setConnectionStatus('connecting');
    setError(null);

    try {
      const ws = new WebSocket(wsUrl);
      webSocketRef.current = ws;

      ws.onopen = () => {
        console.log('Global WebSocket: Connected');
        setConnectionStatus('connected');
        setRetryCount(0);
        setError(null);
        setLastMessageData(null); // Clear last message on new connection
      };

      ws.onmessage = (event) => {
        const rawData = event.data as string;
        setLastMessageData(rawData); // Store the raw message data

        try {
          const message = JSON.parse(rawData) as WebSocketMessage;
          console.log('Global WebSocket: Message received:', message);

          // Invoke generic subscriptions based on chatId
          if (message.sender_id && subscriptionsRef.current[message.sender_id]) {
            subscriptionsRef.current[message.sender_id].forEach((cb) =>
              cb(message)
            );
          }
          if (
            message.receiver_id &&
            subscriptionsRef.current[message.receiver_id]
          ) {
            subscriptionsRef.current[message.receiver_id].forEach((cb) =>
              cb(message)
            );
          }
          if (message.group_id && subscriptionsRef.current[message.group_id]) {
            subscriptionsRef.current[message.group_id].forEach((cb) =>
              cb(message)
            );
          }

          if (message.type === 'online_users' && message.userIds) {
            setOnlineUserIds((prevUserIds) => {
              if (message.userIds === undefined) {
                if (prevUserIds.length > 0) {
                  console.log(
                    'Global WebSocket: message.userIds is undefined, clearing online users.'
                  );
                  return [];
                }
                return prevUserIds;
              }
              const sortedPrevUserIds = [...prevUserIds].sort();
              const sortedNewUserIds = [...message.userIds].sort();
              if (
                JSON.stringify(sortedPrevUserIds) !==
                JSON.stringify(sortedNewUserIds)
              ) {
                console.log(
                  'Global WebSocket: Updating online users:',
                  message.userIds
                );
                return message.userIds;
              }
              return prevUserIds;
            });
          } else if (
            message.type === 'direct' &&
            message.sender_id &&
            message.content
          ) {
            const senderId = message.sender_id;
            const currentChatPath = `/chat/${senderId}`;
            setMessageCount((prev) => prev + 1);

            // Invoke direct message subscriptions
            if (subscriptionsRef.current['direct']) {
              subscriptionsRef.current['direct'].forEach((cb) => cb(message));
            }

            if (pathname !== currentChatPath) {
              const senderDisplayName = message.sender_id; // Placeholder
              const messageSnippet =
                message.content.substring(0, 50) +
                (message.content.length > 50 ? '...' : '');
              toast.custom(
                (t) => (
                  <div
                    className={`${
                      t.visible ? 'animate-enter' : 'animate-leave'
                    } max-w-md w-full bg-gray-800 shadow-lg rounded-lg pointer-events-auto flex ring-1 ring-gray-700 border border-gray-600`}
                  >
                    <div className="flex-1 w-0 p-4">
                      <div className="flex items-start">
                        <div className="flex-shrink-0 mt-0.5">
                          <div className="w-8 h-8 rounded-full bg-gray-700 flex items-center justify-center">
                            <svg
                              className="w-4 h-4 text-gray-300"
                              fill="none"
                              stroke="currentColor"
                              viewBox="0 0 24 24"
                            >
                              <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                strokeWidth={2}
                                d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
                              />
                            </svg>
                          </div>
                        </div>
                        <div className="ml-3 flex-1">
                          <p className="text-sm font-medium text-gray-100">
                            New Direct Message
                          </p>
                          <p className="mt-1 text-sm text-gray-300 leading-relaxed">
                            {messageSnippet}
                          </p>
                          <p className="mt-1 text-xs text-gray-500">
                            From {senderDisplayName}
                          </p>
                        </div>
                      </div>
                    </div>
                    <div className="flex border-l border-gray-700">
                      <button
                        onClick={() => {
                          router.push(`/chat/${senderId}`);
                          toast.dismiss(t.id);
                        }}
                        className="border border-transparent rounded-none p-3 flex items-center justify-center text-xs font-medium text-indigo-400 hover:text-indigo-300 hover:bg-gray-700/50 focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-colors duration-200"
                        title="Open Chat"
                      >
                        <svg
                          className="w-4 h-4"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
                          />
                        </svg>
                      </button>
                      <button
                        onClick={() => toast.dismiss(t.id)}
                        className="border border-transparent rounded-none rounded-r-lg p-3 flex items-center justify-center text-xs font-medium text-gray-400 hover:text-gray-300 hover:bg-gray-700/50 focus:outline-none focus:ring-2 focus:ring-gray-500 transition-colors duration-200"
                        title="Dismiss"
                      >
                        <svg
                          className="w-4 h-4"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M6 18L18 6M6 6l12 12"
                          />
                        </svg>
                      </button>
                    </div>
                  </div>
                ),
                { duration: 6000 }
              );
            }
          } else if (message.type === 'group') {
            const groupId = message.receiver_id; // Use receiver_id as groupId for group messages

            if (groupId && message.sender_id && message.content) {
              setMessageCount((prev) => prev + 1);

              // 1. Dispatch to specific group subscribers (subscribed via subscribeToGroupMessages)
              const groupSpecificKey = `group-${groupId}`;
              const groupSubscribers = subscriptionsRef.current[groupSpecificKey];
              console.log(
                '[GlobalWebSocketContext] Dispatching to group subscribers for groupId:',
                groupId,
                'Subscribers found:',
                !!(groupSubscribers && groupSubscribers.length > 0)
              );
              if (groupSubscribers) {
                groupSubscribers.forEach((cb) => cb(message));
              }

              // 2. Dispatch to generic subscribers (subscribed via subscribeToMessages with chatId === groupId)
              const genericSubscribers = subscriptionsRef.current[groupId];
              console.log(
                '[GlobalWebSocketContext] Dispatching to generic subscribers for chatId (groupId):',
                groupId,
                'Subscribers found:',
                !!(genericSubscribers && genericSubscribers.length > 0)
              );
              if (genericSubscribers) {
                genericSubscribers.forEach((cb) => cb(message));
              }
              
              // Toast notification logic
              const currentGroupChatPath = `/groups/${groupId}/chat`;
              if (pathname !== currentGroupChatPath) {
                const groupName = groupId; // Using the determined groupId (from message.receiver_id)
                const senderDisplayName = message.sender_id; // Placeholder
                const messageSnippet =
                  message.content.substring(0, 50) +
                  (message.content.length > 50 ? '...' : '');

                toast.custom(
                  (t) => (
                    <div
                      className={`${
                        t.visible ? 'animate-enter' : 'animate-leave'
                      } max-w-md w-full bg-gray-800 shadow-lg rounded-lg pointer-events-auto flex ring-1 ring-gray-700 border border-gray-600`}
                    >
                      <div className="flex-1 w-0 p-4">
                        <div className="flex items-start">
                          <div className="flex-shrink-0 mt-0.5">
                            <div className="w-8 h-8 rounded-full bg-gray-700 flex items-center justify-center">
                              {/* Group Icon */}
                              <svg
                                className="w-4 h-4 text-gray-300"
                                fill="none"
                                stroke="currentColor"
                                viewBox="0 0 24 24"
                                xmlns="http://www.w3.org/2000/svg"
                              >
                                <path
                                  strokeLinecap="round"
                                  strokeLinejoin="round"
                                  strokeWidth="2"
                                  d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
                                ></path>
                              </svg>
                            </div>
                          </div>
                          <div className="ml-3 flex-1">
                            <p className="text-sm font-medium text-gray-100">
                              New Group Message in {groupName}
                            </p>
                            <p className="mt-1 text-sm text-gray-300 leading-relaxed">
                              {senderDisplayName}: {messageSnippet}
                            </p>
                          </div>
                        </div>
                      </div>
                      <div className="flex border-l border-gray-700">
                        <button
                          onClick={() => {
                            router.push(`/groups/${groupId}/chat`); // Using the determined groupId
                            toast.dismiss(t.id);
                          }}
                          className="border border-transparent rounded-none p-3 flex items-center justify-center text-xs font-medium text-indigo-400 hover:text-indigo-300 hover:bg-gray-700/50 focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-colors duration-200"
                          title="Open Group Chat"
                        >
                          <svg
                            className="w-4 h-4"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                          >
                            <path
                              strokeLinecap="round"
                              strokeLinejoin="round"
                              strokeWidth={2}
                              d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
                            />
                          </svg>
                        </button>
                        <button
                          onClick={() => toast.dismiss(t.id)}
                          className="border border-transparent rounded-none rounded-r-lg p-3 flex items-center justify-center text-xs font-medium text-gray-400 hover:text-gray-300 hover:bg-gray-700/50 focus:outline-none focus:ring-2 focus:ring-gray-500 transition-colors duration-200"
                          title="Dismiss"
                        >
                          <svg
                            className="w-4 h-4"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                          >
                            <path
                              strokeLinecap="round"
                              strokeLinejoin="round"
                              strokeWidth={2}
                              d="M6 18L18 6M6 6l12 12"
                            />
                          </svg>
                        </button>
                      </div>
                    </div>
                  ),
                  { duration: 6000 }
                );
              }
            } else {
                console.warn('[GlobalWebSocketContext] Group message missing receiver_id (for groupId), sender_id, or content:', message);
            }
          } else if (
            (message.type === 'new_notification' && message.payload) ||
            (message.type === 'notification_created' && message.data)
          ) {
            // Handle notification messages from backend
            const notificationData = message.payload || message.data;
            if (notificationData) {
              const getNotificationIcon = (type: string) => {
                switch (type) {
                  case 'follow_request':
                    return (
                      <svg
                        className="w-5 h-5 text-blue-400"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
                        />
                      </svg>
                    );
                  case 'follow_accept':
                    return (
                      <svg
                        className="w-5 h-5 text-green-400"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                        />
                      </svg>
                    );
                  case 'group_invite':
                    return (
                      <svg
                        className="w-5 h-5 text-purple-400"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
                        />
                      </svg>
                    );
                  case 'group_join_request':
                    return (
                      <svg
                        className="w-5 h-5 text-yellow-400"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197m13.5-9a2.25 2.25 0 11-4.5 0 2.25 2.25 0 014.5 0z"
                        />
                      </svg>
                    );
                  case 'group_event_created':
                    return (
                      <svg
                        className="w-5 h-5 text-orange-400"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
                        />
                      </svg>
                    );
                  default:
                    return (
                      <svg
                        className="w-5 h-5 text-gray-400"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M15 17h5l-5 5-5-5h5v-6h-3l4-4 4 4h-3v6z"
                        />
                      </svg>
                    );
                }
              };

              const getNotificationAction = (
                type: string,
                entityId: string
              ) => {
                switch (type) {
                  case 'follow_request':
                    return {
                      label: 'View Profile',
                      path: `/profile/${entityId}`,
                    };
                  case 'follow_accept':
                    return {
                      label: 'View Profile',
                      path: `/profile/${entityId}`,
                    };
                  case 'group_invite':
                  case 'group_join_request':
                  case 'group_event_created':
                    return {
                      label: 'View Group',
                      path: `/groups/${entityId}`,
                    };
                  default:
                    return {
                      label: 'View',
                      path: '/notifications',
                    };
                }
              };

              const action = getNotificationAction(
                notificationData.type,
                notificationData.entity_id
              );
              const icon = getNotificationIcon(notificationData.type);

              toast.custom(
                (t) => (
                  <div
                    className={`${
                      t.visible ? 'animate-enter' : 'animate-leave'
                    } max-w-md w-full bg-gray-800 shadow-lg rounded-lg pointer-events-auto flex ring-1 ring-gray-700 border border-gray-600`}
                  >
                    <div className="flex-1 w-0 p-4">
                      <div className="flex items-start">
                        <div className="flex-shrink-0 mt-0.5">
                          <div className="w-8 h-8 rounded-full bg-gray-700 flex items-center justify-center">
                            {icon}
                          </div>
                        </div>
                        <div className="ml-3 flex-1">
                          <p className="text-sm font-medium text-gray-100">
                            New Notification
                          </p>
                          <p className="mt-1 text-sm text-gray-300 leading-relaxed">
                            {notificationData.message}
                          </p>
                          <p className="mt-1 text-xs text-gray-500">
                            {new Date(
                              notificationData.created_at
                            ).toLocaleTimeString()}
                          </p>
                        </div>
                      </div>
                    </div>
                    <div className="flex border-l border-gray-700">
                      <button
                        onClick={() => {
                          router.push(action.path);
                          toast.dismiss(t.id);
                        }}
                        className="border border-transparent rounded-none p-3 flex items-center justify-center text-xs font-medium text-blue-400 hover:text-blue-300 hover:bg-gray-700/50 focus:outline-none focus:ring-2 focus:ring-blue-500 transition-colors duration-200"
                        title={action.label}
                      >
                        <svg
                          className="w-4 h-4"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"
                          />
                        </svg>
                      </button>
                      <button
                        onClick={() => toast.dismiss(t.id)}
                        className="border border-transparent rounded-none rounded-r-lg p-3 flex items-center justify-center text-xs font-medium text-gray-400 hover:text-gray-300 hover:bg-gray-700/50 focus:outline-none focus:ring-2 focus:ring-gray-500 transition-colors duration-200"
                        title="Dismiss"
                      >
                        <svg
                          className="w-4 h-4"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            strokeWidth={2}
                            d="M6 18L18 6M6 6l12 12"
                          />
                        </svg>
                      </button>
                    </div>
                  </div>
                ),
                { duration: 8000 } // Slightly longer duration for notifications
              );
            }
          }
          // Handle other global message types here
        } catch (e) {
          console.error('Global WebSocket: Error parsing message:', e);
          setError('Failed to parse incoming message.');
        }
      };

      ws.onerror = (event) => {
        console.error('Global WebSocket: Error', event);
        setError('WebSocket connection error occurred.');
        setConnectionStatus('error');
        // onclose will handle retry logic for errors that also trigger close
      };

      ws.onclose = (event) => {
        console.log(
          `Global WebSocket: Closed. Code: ${event.code}, Reason: ${event.reason}, Clean: ${event.wasClean}`
        );
        // Check if the current ref is the one that closed to avoid issues with rapid connect/disconnect
        if (webSocketRef.current !== ws) {
          console.log(
            'Global WebSocket: onclose event for a stale WebSocket instance. Ignoring.'
          );
          return;
        }

        webSocketRef.current = null; // Clear the ref as it's closed

        if (!event.wasClean && isAuthenticated) {
          // Only retry if not a clean close and user is still authenticated
          setConnectionStatus('disconnected'); // Show disconnected before retrying
          if (retryCount < MAX_RETRIES) {
            const delay = INITIAL_RETRY_DELAY * Math.pow(2, retryCount);
            console.log(
              `Global WebSocket: Attempting to reconnect in ${
                delay / 1000
              }s... (Retry ${retryCount + 1}/${MAX_RETRIES})`
            );
            retryTimeoutRef.current = setTimeout(() => {
              setRetryCount((prev) => prev + 1);
              connectWebSocket();
            }, delay);
          } else {
            console.error('Global WebSocket: Max retries reached. Giving up.');
            setError('Failed to connect after multiple retries.');
            setConnectionStatus('error');
          }
        } else {
          // If it was a clean close, or user logged out, just set to disconnected.
          setConnectionStatus('disconnected');
          if (!isAuthenticated) {
            setOnlineUserIds([]); // Clear online users if logged out
          }
        }
      };
    } catch (e) {
      console.error('Global WebSocket: Instantiation failed', e);
      setError('Failed to instantiate WebSocket.');
      setConnectionStatus('error');
      if (retryCount < MAX_RETRIES && isAuthenticated) {
        const delay = INITIAL_RETRY_DELAY * Math.pow(2, retryCount);
        console.log(
          `Global WebSocket: Retrying instantiation in ${delay / 1000}s...`
        );
        retryTimeoutRef.current = setTimeout(() => {
          setRetryCount((prev) => prev + 1);
          connectWebSocket();
        }, delay);
      } else if (isAuthenticated) {
        setError('Failed to instantiate WebSocket after multiple retries.');
      }
    }
  }, [
    hydrated,
    isAuthenticated,
    currentUserId,
    getWebSocketUrl,
    retryCount,
    disconnectWebSocket,
    pathname, // Added pathname to dependencies as it's used in onmessage
    router, // Added router to dependencies
  ]);

  const sendMessage = useCallback(
    (message: object) => {
      if (
        webSocketRef.current &&
        webSocketRef.current.readyState === WebSocket.OPEN
      ) {
        try {
          webSocketRef.current.send(JSON.stringify(message));
          console.log('Global WebSocket: Message sent:', message);
        } catch (e) {
          console.error('Global WebSocket: Error sending message:', e);
          setError('Failed to send message.');
        }
      } else {
        console.warn(
          'Global WebSocket: Attempted to send message, but connection is not open.'
        );
        setError('Cannot send message: WebSocket is not connected.');
        // Optionally, queue the message or attempt to reconnect
      }
    },
    [webSocketRef]
  ); // webSocketRef is stable

  const subscribe = useCallback(
    (key: string, callback: MessageCallback): (() => void) => {
      subscriptionsRef.current[key] = [
        ...(subscriptionsRef.current[key] || []),
        callback,
      ];
      return () => {
        subscriptionsRef.current[key] = (
          subscriptionsRef.current[key] || []
        ).filter((cb) => cb !== callback);
        if (subscriptionsRef.current[key].length === 0) {
          delete subscriptionsRef.current[key];
        }
      };
    },
    []
  );

  const subscribeToMessages = useCallback(
    (chatId: string, callback: MessageCallback) => {
      return subscribe(chatId, callback);
    },
    [subscribe]
  );

  const subscribeToDirectMessages = useCallback(
    (callback: MessageCallback) => {
      return subscribe('direct', callback);
    },
    [subscribe]
  );

  const subscribeToGroupMessages = useCallback(
    (groupId: string, callback: MessageCallback) => {
      return subscribe(`group-${groupId}`, callback);
    },
    [subscribe]
  );

  useEffect(() => {
    if (!hydrated) {
      console.log('Global WebSocket: Waiting for Zustand store hydration...');
      return;
    }

    console.log(
      'Global WebSocket: Hydration complete. Auth state:',
      isAuthenticated,
      'User ID:',
      currentUserId
    );
    console.log(
      'Global WebSocket: NEXT_PUBLIC_WEBSOCKET_URL:',
      process.env.NEXT_PUBLIC_WEBSOCKET_URL
    );

    if (isAuthenticated && currentUserId) {
      if (
        !webSocketRef.current ||
        webSocketRef.current.readyState === WebSocket.CLOSED
      ) {
        console.log('Global WebSocket: Auth detected, attempting to connect.');
        connectWebSocket();
      } else {
        console.log(
          'Global WebSocket: Auth detected, connection already open or opening.'
        );
      }
    } else {
      console.log(
        'Global WebSocket: No auth or user ID, ensuring disconnection.'
      );
      disconnectWebSocket();
    }

    return () => {
      // Cleanup on component unmount or if dependencies change causing re-run
      // This specific cleanup might be redundant if disconnectWebSocket is called when auth changes,
      // but good for safety if provider unmounts for other reasons.
      console.log(
        'Global WebSocket: Provider useEffect cleanup. Ensuring disconnection.'
      );
      disconnectWebSocket();
    };
  }, [
    isAuthenticated,
    currentUserId,
    hydrated,
    connectWebSocket,
    disconnectWebSocket,
  ]);

  const contextValue = React.useMemo(
    () => ({
      connectionStatus,
      onlineUserIds,
      error,
      lastMessageData, // Include in context
      messageCount, // Include message count
      connectWebSocket,
      disconnectWebSocket,
      clearMessageCount,
      sendMessage,
      subscribeToMessages,
      subscribeToDirectMessages,
      subscribeToGroupMessages,
    }),
    [
      connectionStatus,
      onlineUserIds,
      error,
      lastMessageData,
      messageCount,
      connectWebSocket,
      disconnectWebSocket,
      clearMessageCount,
      sendMessage,
      subscribeToMessages,
      subscribeToDirectMessages,
      subscribeToGroupMessages,
    ]
  );

  return (
    <GlobalWebSocketContext.Provider value={contextValue}>
      {children}
    </GlobalWebSocketContext.Provider>
  );
};
