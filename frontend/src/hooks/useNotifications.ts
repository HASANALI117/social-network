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
          setUnreadCount(data.unread_count || 0); // Ensure unread_count is also handled if potentially undefined
          setHasMore(data.has_more || false); // Ensure has_more is also handled
          setOffset(currentOffset + newNotifications.length);
        } else {
          // Handle cases where data itself might be null or undefined, though useRequest might already guard this
          setNotifications((prev) => (loadMore ? prev : []));
          setUnreadCount(0);
          setHasMore(false);
          // setOffset(currentOffset); // Offset might not need to change or reset depending on desired behavior
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
          message.type === 'notification_created' &&
          message.data
        ) {
          const newNotification = message.data as Notification;

          console.log(
            'WebSocket: Received new notification event.',
            newNotification
          );

          setNotifications((prev) => [
            newNotification,
            ...prev.filter((n) => n.id !== newNotification.id),
          ]);
          setUnreadCount((prev) => prev + 1);
          toast.success(`New notification: ${newNotification.message}`, {
            duration: 5000,
          });
        }
      } catch (error) {
        // console.warn("WebSocket: Received message that is not a targeted notification event or failed to parse:", error);
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

  return {
    notifications,
    unreadCount,
    isLoading: isLoadingFetch || isMarkingAsRead || isMarkingAllAsRead,
    hasMore,
    fetchNotifications,
    markAsRead,
    markAllAsRead,
    error: fetchError || markAsReadError || markAllAsReadError, // Consolidate errors
  };
}
