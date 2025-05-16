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
  connectWebSocket: () => void;
  disconnectWebSocket: () => void;
}

const GlobalWebSocketContext = createContext<GlobalWebSocketContextType | undefined>(
  undefined
);

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

interface WebSocketMessage {
  type: string;
  userIds?: string[];
  // For direct messages
  sender_id?: string;
  content?: string;
  // Add other potential message fields here
}

export const GlobalWebSocketProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const webSocketRef = useRef<WebSocket | null>(null);
  const [connectionStatus, setConnectionStatus] =
    useState<ConnectionStatus>('idle');
  const [onlineUserIds, setOnlineUserIds] = useState<string[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [retryCount, setRetryCount] = useState(0);
  const retryTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const currentUserId = useUserStore((state) => state.user?.id);
  const isAuthenticated = useUserStore((state) => state.isAuthenticated);
  const hydrated = useUserStore((state) => state.hydrated);
  const pathname = usePathname();
  const router = useRouter();

  const getWebSocketUrl = useCallback((): string | null => {
    if (!currentUserId) return null;

    const wsBaseUrl = process.env.NEXT_PUBLIC_WEBSOCKET_URL;
    if (!wsBaseUrl) {
      console.error("Global WebSocket: NEXT_PUBLIC_WEBSOCKET_URL is not set.");
      return null;
    }

    // Determine scheme based on browser protocol, but allow override if wsBaseUrl includes it
    let scheme = window.location.protocol === "https:" ? "wss" : "ws";
    // let domainAndPath = wsBaseUrl; // This variable is not used in the provided snippet

    if (wsBaseUrl.startsWith("ws://") || wsBaseUrl.startsWith("wss://")) {
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


  const connectWebSocket = useCallback(() => {
    if (!hydrated || !isAuthenticated || !currentUserId) {
      console.log('Global WebSocket: Conditions not met for connection (hydrated, isAuthenticated, user.id). Current state:', { hydrated, isAuthenticated, userId: currentUserId });
      if (webSocketRef.current && webSocketRef.current.readyState === WebSocket.OPEN) {
         disconnectWebSocket();
      }
      return;
    }

    if (webSocketRef.current && webSocketRef.current.readyState === WebSocket.OPEN) {
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

    console.log(`Global WebSocket: Connecting to ${wsUrl}... (Attempt ${retryCount + 1})`);
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
      };

      ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data as string) as WebSocketMessage;
          console.log('Global WebSocket: Message received:', message);
          if (message.type === 'online_users' && message.userIds) {
            setOnlineUserIds(prevUserIds => {
              if (message.userIds === undefined) {
                // If message.userIds is undefined, decide on behavior.
                // Option 1: Don't change state
                // console.log('Global WebSocket: message.userIds is undefined, not updating state.');
                // return prevUserIds;
                // Option 2: Set to empty array if it's different from current
                if (prevUserIds.length > 0) {
                    console.log('Global WebSocket: message.userIds is undefined, clearing online users.');
                    return [];
                }
                return prevUserIds;
              }

              // Sort arrays before stringifying to ensure order doesn't cause false positives,
              // but return the original message.userIds to preserve server order if it's meaningful.
              const sortedPrevUserIds = [...prevUserIds].sort();
              const sortedNewUserIds = [...message.userIds].sort(); // message.userIds is now guaranteed to be string[]

              if (JSON.stringify(sortedPrevUserIds) !== JSON.stringify(sortedNewUserIds)) {
                console.log('Global WebSocket: Updating online users:', message.userIds);
                return message.userIds; // Return the new, potentially unsorted, list
              }
              // console.log('Global WebSocket: Online users list is effectively the same, not updating state.');
              return prevUserIds; // Explicitly return previous state if no change
            });
          } else if (message.type === 'direct' && message.sender_id && message.content) {
            const senderId = message.sender_id;
            const currentChatPath = `/chat/${senderId}`;

            if (pathname !== currentChatPath) {
              const senderDisplayName = message.sender_id; // Placeholder for actual name
              const messageSnippet = message.content.substring(0, 50) + (message.content.length > 50 ? '...' : '');

              toast.custom(
                (t) => (
                  <div
                    className={`${
                      t.visible ? 'animate-enter' : 'animate-leave'
                    } max-w-md w-full bg-gray-800 shadow-lg rounded-lg pointer-events-auto flex ring-1 ring-gray-700`}
                  >
                    <div className="flex-1 w-0 p-4">
                      <div className="flex items-start">
                        <div className="ml-3 flex-1">
                          <p className="text-sm font-medium text-gray-100">
                            New Direct Message
                          </p>
                          <p className="mt-1 text-sm text-gray-400">
                            {messageSnippet}
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
                        className="w-full border border-transparent rounded-none rounded-r-lg p-4 flex items-center justify-center text-sm font-medium text-indigo-400 hover:text-indigo-300 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                      >
                        Open Chat
                      </button>
                      <button
                          onClick={() => toast.dismiss(t.id)}
                          className="w-full border border-transparent rounded-none p-4 flex items-center justify-center text-sm font-medium text-gray-300 hover:text-gray-400 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                      >
                          Dismiss
                      </button>
                    </div>
                  </div>
                ),
                { duration: 6000 }
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
        console.log(`Global WebSocket: Closed. Code: ${event.code}, Reason: ${event.reason}, Clean: ${event.wasClean}`);
        // Check if the current ref is the one that closed to avoid issues with rapid connect/disconnect
        if (webSocketRef.current !== ws) {
            console.log("Global WebSocket: onclose event for a stale WebSocket instance. Ignoring.");
            return;
        }

        webSocketRef.current = null; // Clear the ref as it's closed

        if (!event.wasClean && isAuthenticated) { // Only retry if not a clean close and user is still authenticated
          setConnectionStatus('disconnected'); // Show disconnected before retrying
          if (retryCount < MAX_RETRIES) {
            const delay = INITIAL_RETRY_DELAY * Math.pow(2, retryCount);
            console.log(`Global WebSocket: Attempting to reconnect in ${delay / 1000}s... (Retry ${retryCount + 1}/${MAX_RETRIES})`);
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
        console.log(`Global WebSocket: Retrying instantiation in ${delay / 1000}s...`);
        retryTimeoutRef.current = setTimeout(() => {
          setRetryCount((prev) => prev + 1);
          connectWebSocket();
        }, delay);
      } else if (isAuthenticated) {
        setError('Failed to instantiate WebSocket after multiple retries.');
      }
    }
  }, [hydrated, isAuthenticated, currentUserId, getWebSocketUrl, retryCount, disconnectWebSocket]);


  useEffect(() => {
    if (!hydrated) {
      console.log("Global WebSocket: Waiting for Zustand store hydration...");
      return;
    }

    console.log("Global WebSocket: Hydration complete. Auth state:", isAuthenticated, "User ID:", currentUserId);
    console.log("Global WebSocket: NEXT_PUBLIC_WEBSOCKET_URL:", process.env.NEXT_PUBLIC_WEBSOCKET_URL);

    if (isAuthenticated && currentUserId) {
      if (!webSocketRef.current || webSocketRef.current.readyState === WebSocket.CLOSED) {
        console.log("Global WebSocket: Auth detected, attempting to connect.");
        connectWebSocket();
      } else {
        console.log("Global WebSocket: Auth detected, connection already open or opening.");
      }
    } else {
      console.log("Global WebSocket: No auth or user ID, ensuring disconnection.");
      disconnectWebSocket();
    }

    return () => {
      // Cleanup on component unmount or if dependencies change causing re-run
      // This specific cleanup might be redundant if disconnectWebSocket is called when auth changes,
      // but good for safety if provider unmounts for other reasons.
      console.log("Global WebSocket: Provider useEffect cleanup. Ensuring disconnection.");
      disconnectWebSocket();
    };
  }, [isAuthenticated, currentUserId, hydrated, connectWebSocket, disconnectWebSocket]);

  const contextValue = React.useMemo(() => ({
    connectionStatus,
    onlineUserIds,
    error,
    connectWebSocket,
    disconnectWebSocket,
  }), [connectionStatus, onlineUserIds, error, connectWebSocket, disconnectWebSocket]);

  return (
    <GlobalWebSocketContext.Provider value={contextValue}>
      {children}
    </GlobalWebSocketContext.Provider>
  );
};