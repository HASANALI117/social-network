"use client"

import { Fragment, useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import {
  DropdownDivider,
  DropdownItem,
  DropdownLabel,
  DropdownMenu,
  DropdownHeader
} from '@/components/ui/dropdown'
import { UserPlusIcon, UserGroupIcon, CalendarDaysIcon, CheckCircleIcon, EnvelopeOpenIcon, EyeIcon, XCircleIcon, BellIcon } from '@heroicons/react/24/solid'
import { toast } from 'react-hot-toast'
import { formatDistanceToNow } from 'date-fns' // For relative timestamps
import { useRequest } from '@/hooks/useRequest';
import { Notification } from '@/types/Notification'; // Import from centralized types
import { useGlobalWebSocket } from '@/contexts/GlobalWebSocketContext';

interface NotificationsDropdownProps {
  anchor: 'top start' | 'bottom start' | 'bottom end'
  notifications: Notification[]
  unreadCount: number
  onNotificationClick: (notification: Notification) => Promise<void> // To handle marking as read and navigation
  onMarkAllAsRead: () => Promise<void>
  fetchNotifications: () => void // To refetch notifications after an action
  currentUserId?: string // Optional: pass current user ID if needed for actions
}

const getNotificationIcon = (type: string) => {
  switch (type) {
    case 'follow_request':
      return <UserPlusIcon className="h-4 w-4 text-blue-600 dark:text-blue-400" />
    case 'follow_accept':
      return <UserPlusIcon className="h-4 w-4 text-green-600 dark:text-green-400" />
    case 'group_invite':
      return <UserGroupIcon className="h-4 w-4 text-purple-600 dark:text-purple-400" />
    case 'group_join_request':
      return <UserGroupIcon className="h-4 w-4 text-indigo-600 dark:text-indigo-400" />
    case 'group_event_created':
      return <CalendarDaysIcon className="h-4 w-4 text-emerald-600 dark:text-emerald-400" />
    case 'event_reminder':
      return <CalendarDaysIcon className="h-4 w-4 text-amber-600 dark:text-amber-400" />
    default:
      return <EyeIcon className="h-4 w-4 text-zinc-500 dark:text-zinc-400" />
  }
}


export function NotificationsDropdown({
  anchor,
  notifications: initialNotifications,
  unreadCount: initialUnreadCount,
  onNotificationClick,
  onMarkAllAsRead,
  fetchNotifications,
  currentUserId // Added currentUserId
}: NotificationsDropdownProps) {
  const router = useRouter()
  const [notifications, setNotifications] = useState<Notification[]>(initialNotifications)
  const [unreadCount, setUnreadCount] = useState<number>(initialUnreadCount)
  const [processingAction, setProcessingAction] = useState<string | null>(null); // To disable buttons during action
  const [hideReadNotifications, setHideReadNotifications] = useState<boolean>(false); // Toggle for hiding read notifications
  
  // Listen for WebSocket updates to refresh dropdown in real-time
  const { lastMessageData } = useGlobalWebSocket();

  useEffect(() => {
    setNotifications(initialNotifications)
  }, [initialNotifications])

  useEffect(() => {
    setUnreadCount(initialUnreadCount)
  }, [initialUnreadCount])

  // Handle real-time WebSocket updates
  useEffect(() => {
    if (lastMessageData) {
      try {
        const message = typeof lastMessageData === 'string' 
          ? JSON.parse(lastMessageData) 
          : lastMessageData;

        if (message && 
            (message.type === 'notification_created' || message.type === 'new_notification') && 
            (message.data || message.payload)) {
          
          console.log('NotificationsDropdown: Received WebSocket notification update, refreshing...');
          
          // Refresh the notifications list when a new notification is received
          fetchNotifications();
        }
      } catch (error) {
        // Ignore parsing errors for non-notification messages
        console.debug('NotificationsDropdown: WebSocket message parsing error (likely not a notification):', error);
      }
    }
  }, [lastMessageData, fetchNotifications]);


  const handleItemClick = async (notification: Notification) => {
    try {
      await onNotificationClick(notification) // Parent handles API call and state update
      
      // Optimistically update UI or rely on parent to refetch/pass new props
      // For now, parent will handle refetching via fetchNotifications prop
      fetchNotifications() // Ask parent to refetch

      // Navigate based on notification type, only if not an action button click
      // Action buttons will handle their own logic and won't trigger this navigation
      if (!['follow_request', 'group_invite', 'group_join_request'].includes(notification.type)) {
        switch (notification.type) {
          case 'follow_accept': // User who sent request gets this
            if (notification.entity_type === 'user' && notification.entity_id) {
              router.push(`/profile/${notification.entity_id}`) // Navigate to the profile of user who accepted
            }
            break
          case 'group_event_created':
            if (notification.entity_type === 'event' && notification.entity_id) {
              // Assuming entity_id is eventID. We need groupID to navigate to event page.
              // This might require changes in backend to include groupID in notification for events,
              // or a way to fetch event details here to get groupID.
              // For now, let's assume entity_id for group_event_created is the group_id and we need event_id from somewhere else
              // Or, if entity_id IS the event_id, we need a way to get the group_id.
              // Let's assume for now entity_id is event_id and we need to fetch group_id.
              // This is a placeholder, actual navigation might be more complex.
              // A better approach: backend sends entity_id as eventID and a secondary_entity_id as groupID for this type.
              // For now, we'll just log it.
              console.log(`Navigate to event: ${notification.entity_id} - Group ID needed.`);
              // Example: router.push(`/groups/${GROUP_ID_HERE}/events/${notification.entity_id}`);
            }
            break;
          // For group_invite and group_join_request, navigation happens if user clicks the main body,
          // not the accept/decline buttons.
          case 'group_invite': // Click on body navigates to group
            if (notification.entity_type === 'group' && notification.entity_id) {
              router.push(`/groups/${notification.entity_id}`);
            }
            break;
          case 'group_join_request': // For group creator, click on body navigates to user profile
            if (notification.entity_type === 'user' && notification.entity_id) {
               router.push(`/profile/${notification.entity_id}`);
            }
            break;
          default:
            // No specific navigation for this type, or could navigate to a general notifications page
            break
        }
      }
    } catch (error) {
      console.error("Error handling notification click:", error)
      toast.error("Could not process notification action.")
    }
  }

  const handleMarkAllRead = async () => {
    try {
      await onMarkAllAsRead()
      fetchNotifications() // Ask parent to refetch
    } catch (error) {
      console.error("Error marking all as read:", error)
      toast.error("Could not mark all notifications as read.")
    }
  }

  const { post: performAction, isLoading: isActionLoading, error: actionError } = useRequest<any>();

  const handleAction = async (
    url: string,
    notificationId: string,
    successMessage: string = "Action completed successfully!"
  ) => {
    setProcessingAction(notificationId);
    try {
      await performAction(url, {}, () => {
        toast.success(successMessage);
        fetchNotifications(); // Refetch to update the list
      });
    } catch (error: any) {
      // Error is now primarily handled by actionError state from useRequest
      console.error("Error performing action:", actionError || error);
      toast.error(actionError?.message || error?.message || "Could not complete action.");
    } finally {
      setProcessingAction(null);
    }
  };
  
  useEffect(() => {
    if (actionError) {
        toast.error(actionError.message || "An error occurred while performing the action.");
    }
  }, [actionError]);


  const getActionUrls = (notification: Notification): { acceptUrl?: string, declineUrl?: string } => {
    switch (notification.type) {
      case 'follow_request':
        return {
          acceptUrl: `/api/followers/accept/${notification.entity_id}`,
          declineUrl: `/api/followers/reject/${notification.entity_id}`,
        };
      case 'group_invite':
        // Assuming notification.id is the invitationId for group invites
        return {
          acceptUrl: `/api/groups/invitations/${notification.id}/accept`,
          declineUrl: `/api/groups/invitations/${notification.id}/reject`,
        };
      case 'group_join_request':
        // Assuming notification.id is the requestId for group join requests
        return {
          acceptUrl: `/api/groups/join-requests/${notification.id}/accept`,
          declineUrl: `/api/groups/join-requests/${notification.id}/reject`,
        };
      default:
        return {};
    }
  };

  return (
    <DropdownMenu 
      className="notifications-dropdown !w-96 !min-w-96 !max-w-96 max-h-96 overflow-hidden shadow-xl" 
      anchor={anchor}
      style={{ width: '384px', minWidth: '384px', maxWidth: '384px', overflow: 'hidden' }}
    >
      <DropdownHeader className="flex justify-between items-center w-full p-4 bg-gradient-to-r from-blue-50 to-purple-50 dark:from-zinc-800 dark:to-zinc-700 border-b border-zinc-200 dark:border-zinc-600" style={{ width: '384px', minWidth: '384px', maxWidth: '384px', boxSizing: 'border-box' }}>
        <div className="flex items-center gap-2">
          <BellIcon className="h-5 w-5 text-blue-600 dark:text-blue-400" />
          <h3 className="text-lg font-semibold text-zinc-900 dark:text-zinc-100">Notifications</h3>
          {unreadCount > 0 && (
            <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-red-100 text-red-800 dark:bg-red-900/50 dark:text-red-200">
              {unreadCount}
            </span>
          )}
        </div>
        <div className="flex items-center gap-2">
          {/* Toggle to hide read notifications */}
          <button
            onClick={() => setHideReadNotifications(!hideReadNotifications)}
            className={`inline-flex items-center px-2 py-1.5 text-xs font-medium rounded-lg transition-all duration-200 ${
              hideReadNotifications 
                ? 'bg-blue-100 text-blue-800 dark:bg-blue-900/50 dark:text-blue-200' 
                : 'text-zinc-600 hover:text-zinc-800 dark:text-zinc-400 dark:hover:text-zinc-300 hover:bg-zinc-100 dark:hover:bg-zinc-700/30'
            }`}
            title={hideReadNotifications ? "Show read notifications" : "Hide read notifications"}
          >
            <EyeIcon className="h-3.5 w-3.5 mr-1" />
            {hideReadNotifications ? 'Show Read' : 'Hide Read'}
          </button>

          {/* View all notifications */}
          <button
            onClick={() => {
              console.log('Navigating to /notifications');
              router.push('/notifications');
            }}
            className="inline-flex items-center px-2 py-1.5 text-xs font-medium text-zinc-600 hover:text-zinc-800 dark:text-zinc-400 dark:hover:text-zinc-300 hover:bg-zinc-100 dark:hover:bg-zinc-700/30 rounded-lg transition-all duration-200"
            title="View all notifications"
          >
            <EyeIcon className="h-3.5 w-3.5 mr-1" />
            View All
          </button>

          {/* Mark all as read */}
          {unreadCount > 0 && (
            <button
              onClick={handleMarkAllRead}
              disabled={processingAction !== null || isActionLoading}
              className="inline-flex items-center px-2 py-1.5 text-xs font-medium text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 hover:bg-blue-100 dark:hover:bg-blue-900/30 rounded-lg transition-all duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
              title="Mark all as read"
            >
              <EnvelopeOpenIcon className="h-3.5 w-3.5 mr-1" />
              Mark Read
            </button>
          )}
        </div>
      </DropdownHeader>

      <div className="overflow-y-auto max-h-80" style={{ width: '384px', boxSizing: 'border-box' }}>{/* Scrollable content area */}

      {(() => {
        const filteredNotifications = hideReadNotifications 
          ? notifications.filter(notification => !notification.is_read)
          : notifications;

        return filteredNotifications.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12 px-6">
            <div className="w-16 h-16 bg-gradient-to-br from-zinc-100 to-zinc-200 dark:from-zinc-800 dark:to-zinc-700 rounded-full flex items-center justify-center mb-4 shadow-inner">
              <BellIcon className="w-8 h-8 text-zinc-400 dark:text-zinc-500" />
            </div>
            <p className="text-sm font-medium text-zinc-600 dark:text-zinc-300 text-center">
              {hideReadNotifications ? 'No unread notifications' : 'No new notifications'}
            </p>
            <p className="text-xs text-zinc-500 dark:text-zinc-400 text-center mt-1 max-w-48">
              {hideReadNotifications 
                ? 'All caught up! Check back later for updates.' 
                : 'We\'ll notify you when something interesting happens'
              }
            </p>
          </div>
        ) : (
          filteredNotifications.map((notification, index) => {
          const { acceptUrl, declineUrl } = getActionUrls(notification);
          const canPerformAction = (acceptUrl || declineUrl) && !notification.is_read;
          const isProcessing = processingAction === notification.id;

          return (
            <Fragment key={notification.id}>
              <div
                className={`notification-item relative w-full transition-all duration-200 ${
                  !notification.is_read 
                    ? 'bg-blue-50/50 dark:bg-blue-950/30 hover:bg-blue-50 dark:hover:bg-blue-950/50' 
                    : 'bg-transparent hover:bg-zinc-50 dark:hover:bg-zinc-800/50'
                } ${isProcessing ? 'opacity-60 pointer-events-none' : ''}`}
                style={{ 
                  width: '100%', 
                  maxWidth: '384px',
                  boxSizing: 'border-box',
                  overflow: 'hidden'
                }}
              >
                {/* Blue indicator for unread notifications */}
                {!notification.is_read && (
                  <div className="absolute left-0 top-0 bottom-0 w-1 bg-blue-500"></div>
                )}
                
                {/* Loading overlay for processing actions */}
                {isProcessing && (
                  <div className="absolute inset-0 bg-white/20 dark:bg-zinc-900/20 flex items-center justify-center rounded-lg backdrop-blur-sm z-10">
                    <div className="w-5 h-5 border-2 border-blue-600 border-t-transparent rounded-full animate-spin"></div>
                  </div>
                )}

                <div 
                  className="flex items-start gap-3 cursor-pointer group p-4" 
                  onClick={() => !isProcessing && handleItemClick(notification)}
                >
                  {/* Notification Icon */}
                  <div className={`shrink-0 p-2 rounded-full transition-all duration-200 group-hover:scale-105 ${
                    !notification.is_read 
                      ? 'bg-blue-100 dark:bg-blue-900/50 shadow-sm' 
                      : 'bg-zinc-100 dark:bg-zinc-800'
                  }`}>
                    {getNotificationIcon(notification.type)}
                  </div>

                  {/* Content */}
                  <div className="flex-1 min-w-0 space-y-1">
                    <div className="flex items-start justify-between gap-2">
                      <p className={`text-sm leading-relaxed ${
                        !notification.is_read 
                          ? 'font-medium text-zinc-900 dark:text-zinc-100' 
                          : 'text-zinc-700 dark:text-zinc-300'
                      } group-hover:text-zinc-900 dark:group-hover:text-zinc-100 transition-colors`}>
                        {notification.message}
                      </p>
                      {!notification.is_read && (
                        <div className="relative">
                          <div className="w-2 h-2 bg-blue-500 rounded-full" title="Unread"></div>
                          <div className="absolute inset-0 w-2 h-2 bg-blue-400 rounded-full animate-ping opacity-75"></div>
                        </div>
                      )}
                    </div>
                    
                    <div className="flex items-center gap-2">
                      <p className="text-xs text-zinc-500 dark:text-zinc-400">
                        {formatDistanceToNow(new Date(notification.created_at), { addSuffix: true })}
                      </p>
                      {!notification.is_read && (
                        <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900/50 dark:text-blue-200">
                          New
                        </span>
                      )}
                    </div>
                  </div>
                </div>

                {/* Action Buttons */}
                {canPerformAction && (
                  <div className="w-full px-4 pb-4 flex space-x-2 justify-end">
                    {acceptUrl && (
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleAction(acceptUrl, notification.id, "Request accepted");
                        }}
                        disabled={isProcessing || isActionLoading}
                        className="inline-flex items-center px-3 py-1.5 text-xs font-medium text-white bg-green-600 rounded-lg hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200 shadow-sm hover:shadow-md transform hover:scale-105 active:scale-95"
                      >
                        <CheckCircleIcon className="h-3.5 w-3.5 mr-1.5" /> 
                        Accept
                      </button>
                    )}
                    {declineUrl && (
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleAction(declineUrl, notification.id, "Request declined");
                        }}
                        disabled={isProcessing || isActionLoading}
                        className="inline-flex items-center px-3 py-1.5 text-xs font-medium text-white bg-red-600 rounded-lg hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-200 shadow-sm hover:shadow-md transform hover:scale-105 active:scale-95"
                      >
                        <XCircleIcon className="h-3.5 w-3.5 mr-1.5" /> 
                        Decline
                      </button>
                    )}
                  </div>
                )}
              </div>
              {index < filteredNotifications.length - 1 && (
                <div className="w-full border-b border-zinc-100 dark:border-zinc-800" />
              )}
            </Fragment>
          )
        })
        );
      })()}
      </div>{/* End scrollable content area */}
    </DropdownMenu>
  )
}
