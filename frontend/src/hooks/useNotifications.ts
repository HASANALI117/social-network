import { useState, useEffect, useCallback } from 'react';
import { toast } from 'react-hot-toast';
import { useRequest } from '@/hooks/useRequest';
import { Notification, NotificationsResponse } from '@/types/Notification';
import { useGlobalWebSocket } from '@/contexts/GlobalWebSocketContext'; // For real-time updates
import { useUserStore } from '@/store/useUserStore';

export function useNotifications() {
  const [notifications, setNotifications] = useState<Notification[]>([]);
  const [unreadCount, setUnreadCount] = useState<number>(0);
  const [hasMore, setHasMore] = useState<boolean>(false);
  const [offset, setOffset] = useState<number>(0);
  const limit = 20; // Or make configurable

  const { isAuthenticated, hydrated } = useUserStore();
  const { lastMessageData } = useGlobalWebSocket();

  const {
    data: fetchedNotificationsData,
    error: fetchError,
    isLoading: isLoadingFetch,
    get: getNotificationsRequest,
  } = useRequest<NotificationsResponse>();

  const {
    error: markAsReadError,
    isLoading: isMarkingAsRead,
    post: markAsReadRequest,
  } = useRequest<Notification>(); // Assuming API returns the updated notification

  const {
    error: markAllAsReadError,
    isLoading: isMarkingAllAsRead,
    post: markAllAsReadRequest,
  } = useRequest<{ message: string }>(); // Assuming API returns a success message

  const fetchNotifications = useCallback(
    async (loadMore = false) => {
      // Don't fetch if user is not authenticated
      if (!isAuthenticated || !hydrated) {
        console.log('useNotifications: Skipping fetch - user not authenticated or store not hydrated');
        return;
      }

      const currentOffset = loadMore ? offset : 0;
      const url = `/api/notifications?limit=${limit}&offset=${currentOffset}`;

      getNotificationsRequest(url, (data) => {
        if (data) {
          const newNotifications = Array.isArray(data.notifications)
            ? data.notifications
            : [];
          setNotifications((prev) =>
            loadMore ? [...prev, ...newNotifications] : newNotifications
          );
          setUnreadCount(data.unread_count || 0);
          setHasMore(data.has_more || false);
          
          if (!loadMore) {
            // Reset offset when fetching from beginning
            setOffset(newNotifications.length);
          } else {
            setOffset(currentOffset + newNotifications.length);
          }
        } else {
          // Handle cases where data itself might be null or undefined
          setNotifications((prev) => (loadMore ? prev : []));
          setUnreadCount(0);
          setHasMore(false);
          if (!loadMore) {
            setOffset(0);
          }
        }
      });
    },
    [offset, limit, getNotificationsRequest, isAuthenticated, hydrated]
  );

  useEffect(() => {
    // Only fetch notifications if user is authenticated and store is hydrated
    if (isAuthenticated && hydrated) {
      fetchNotifications();
    } else {
      // Clear notifications if user is not authenticated
      setNotifications([]);
      setUnreadCount(0);
      setHasMore(false);
      setOffset(0);
    }
  }, [isAuthenticated, hydrated, fetchNotifications]);

  useEffect(() => {
    if (fetchError) {
      console.error('Failed to fetch notifications:', fetchError);
      toast.error(fetchError.message || 'Could not load notifications.');
    }
  }, [fetchError]);

  // Real-time updates via WebSocket
  useEffect(() => {
    if (lastMessageData) {
      try {
        const message =
          typeof lastMessageData === 'string'
            ? JSON.parse(lastMessageData)
            : lastMessageData;

        if (
          message &&
          (message.type === 'notification_created' || message.type === 'new_notification') &&
          (message.data || message.payload)
        ) {
          const newNotification = message.data || message.payload;

          console.log(
            'useNotifications: Received new notification via WebSocket',
            newNotification
          );

          // Add the new notification to the beginning of the list
          setNotifications((prev) => {
            // Remove any existing notification with the same ID to avoid duplicates
            const filtered = prev.filter((n) => n.id !== newNotification.id);
            return [newNotification, ...filtered];
          });
          
          // Only increment unread count if the notification is actually unread
          if (!newNotification.is_read) {
            setUnreadCount((prev) => prev + 1);
          }

          // Show a subtle toast notification (the main toast is handled by GlobalWebSocketContext)
          console.log(`New notification received: ${newNotification.message}`);
        }
      } catch (error) {
        // Ignore parsing errors for non-notification messages
        console.debug('useNotifications: WebSocket message parsing error (likely not a notification):', error);
      }
    }
  }, [lastMessageData]);

  const markAsRead = async (notificationId: string) => {
    // Don't attempt to mark as read if user is not authenticated
    if (!isAuthenticated) {
      console.log('useNotifications: Skipping markAsRead - user not authenticated');
      return;
    }

    const url = `/api/notifications/${notificationId}/read`;
    try {
      await markAsReadRequest(url, {}, (updatedNotification) => {
        if (updatedNotification) {
          setNotifications((prev) =>
            prev.map((n) =>
              n.id === notificationId ? { ...n, is_read: true } : n
            )
          );
          // Optimistically decrement or refetch unread count
          const notification = notifications.find(
            (n) => n.id === notificationId
          );
          if (notification && !notification.is_read) {
            setUnreadCount((prev) => Math.max(0, prev - 1));
          }
          // Or use updatedNotification.unread_count if the API provides it
        }
      });
    } catch (error) {
      // Error is handled by the useRequest hook's error state
      console.error(
        `Failed to mark notification ${notificationId} as read:`,
        markAsReadError
      );
      toast.error(
        markAsReadError?.message || 'Failed to mark notification as read.'
      );
      throw markAsReadError || error;
    }
  };

  const markAllAsRead = async () => {
    // Don't attempt to mark all as read if user is not authenticated
    if (!isAuthenticated) {
      console.log('useNotifications: Skipping markAllAsRead - user not authenticated');
      return;
    }

    const url = `/api/notifications/read-all`;
    try {
      await markAllAsReadRequest(url, {}, (response) => {
        if (response) {
          setNotifications((prev) =>
            prev.map((n) => ({ ...n, is_read: true }))
          );
          setUnreadCount(0);
          // toast.success(response.message || "All notifications marked as read.");
        }
      });
    } catch (error) {
      console.error(
        'Failed to mark all notifications as read:',
        markAllAsReadError
      );
      toast.error(
        markAllAsReadError?.message ||
          'Failed to mark all notifications as read.'
      );
      throw markAllAsReadError || error;
    }
  };

  const refreshNotifications = useCallback(() => {
    // Force refresh by resetting offset and fetching from beginning
    setOffset(0);
    fetchNotifications(false);
  }, [fetchNotifications]);

  return {
    notifications,
    unreadCount,
    isLoading: isLoadingFetch || isMarkingAsRead || isMarkingAllAsRead,
    hasMore,
    fetchNotifications,
    refreshNotifications, // New function for manual refresh
    markAsRead,
    markAllAsRead,
    error: fetchError || markAsReadError || markAllAsReadError, // Consolidate errors
  };
}
