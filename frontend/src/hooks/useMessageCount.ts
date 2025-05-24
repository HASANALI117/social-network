'use client';

import { useState, useCallback, useEffect } from 'react';
import { useGlobalWebSocket } from '@/contexts/GlobalWebSocketContext';
import { useUserStore } from '@/store/useUserStore';
import { usePathname } from 'next/navigation';

export function useMessageCount() {
  const [totalUnreadCount, setTotalUnreadCount] = useState(0);
  const [conversationCounts, setConversationCounts] = useState<
    Map<string, number>
  >(new Map());

  const { isAuthenticated, hydrated } = useUserStore();
  const { lastMessageData } = useGlobalWebSocket();
  const pathname = usePathname();

  // Function to increment message count for a specific sender
  const incrementMessageCount = useCallback((senderId: string) => {
    setTotalUnreadCount((prev) => prev + 1);
    setConversationCounts((prev) => {
      const newMap = new Map(prev);
      const currentCount = newMap.get(senderId) || 0;
      newMap.set(senderId, currentCount + 1);
      return newMap;
    });
  }, []);

  // Function to mark messages as read for a specific conversation
  const markConversationAsRead = useCallback(
    (senderId: string) => {
      const previousCount = conversationCounts.get(senderId) || 0;
      setTotalUnreadCount((prev) => Math.max(0, prev - previousCount));
      setConversationCounts((prev) => {
        const newMap = new Map(prev);
        newMap.delete(senderId);
        return newMap;
      });
    },
    [conversationCounts]
  );

  // Function to get unread count for a specific conversation
  const getConversationUnreadCount = useCallback(
    (senderId: string) => {
      return conversationCounts.get(senderId) || 0;
    },
    [conversationCounts]
  );

  // Function to reset all counts (when user logs out or signs in)
  const resetMessageCounts = useCallback(() => {
    setTotalUnreadCount(0);
    setConversationCounts(new Map());
  }, []);

  // Clear counts if user is not authenticated
  useEffect(() => {
    if (!isAuthenticated || !hydrated) {
      resetMessageCounts();
    }
  }, [isAuthenticated, hydrated, resetMessageCounts]);

  // Listen for WebSocket messages and update counts in real-time
  useEffect(() => {
    if (lastMessageData) {
      try {
        const message = JSON.parse(lastMessageData);
        // Handle direct messages
        if (message.type === 'direct' && message.sender_id && message.content) {
          const senderId = message.sender_id;
          const currentChatPath = `/chat/${senderId}`;

          // Only increment count if user is not currently viewing this chat
          if (pathname !== currentChatPath) {
            incrementMessageCount(senderId);
          }
        }
      } catch (error) {
        console.error('Error parsing WebSocket message:', error);
      }
    }
  }, [lastMessageData, pathname, incrementMessageCount]);

  // Mark conversation as read when user enters a chat
  useEffect(() => {
    const chatMatch = pathname.match(/^\/chat\/(.+)$/);
    if (chatMatch && chatMatch[1]) {
      const senderId = chatMatch[1];
      const currentUnreadCount = getConversationUnreadCount(senderId);

      // Only mark as read if there are unread messages
      if (currentUnreadCount > 0) {
        markConversationAsRead(senderId);
      }
    }
  }, [pathname, getConversationUnreadCount, markConversationAsRead]);

  return {
    totalUnreadCount,
    conversationCounts,
    incrementMessageCount,
    markConversationAsRead,
    getConversationUnreadCount,
    resetMessageCounts,
  };
}
