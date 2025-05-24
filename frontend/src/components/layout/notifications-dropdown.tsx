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
import { UserPlusIcon, UserGroupIcon, CalendarDaysIcon, CheckCircleIcon, EnvelopeOpenIcon, EyeIcon, XCircleIcon } from '@heroicons/react/24/solid'
import { toast } from 'react-hot-toast'
import { formatDistanceToNow } from 'date-fns' // For relative timestamps
import { useRequest } from '@/hooks/useRequest';
import { Notification } from '@/types/Notification'; // Import from centralized types

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
    case 'follow_accept':
      return <UserPlusIcon className="h-5 w-5 text-blue-500" />
    case 'group_invite':
    case 'group_join_request':
      return <UserGroupIcon className="h-5 w-5 text-purple-500" />
    case 'group_event_created':
    case 'event_reminder':
      return <CalendarDaysIcon className="h-5 w-5 text-green-500" />
    default:
      return <EyeIcon className="h-5 w-5 text-gray-500" /> // Generic icon
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

  useEffect(() => {
    setNotifications(initialNotifications)
  }, [initialNotifications])

  useEffect(() => {
    setUnreadCount(initialUnreadCount)
  }, [initialUnreadCount])


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
    <DropdownMenu className="min-w-96 max-h-96 overflow-y-auto" anchor={anchor}> {/* Increased min-width */}
      <DropdownHeader className="flex justify-between items-center p-3 border-b"> {/* Increased padding */}
        <h3 className="text-lg font-semibold">Notifications</h3>
        {unreadCount > 0 && (
          <button
            onClick={handleMarkAllRead}
            disabled={processingAction !== null || isActionLoading}
            className="text-sm text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 flex items-center disabled:opacity-50"
            title="Mark all as read"
          >
            <EnvelopeOpenIcon className="h-4 w-4 mr-1" />
            Mark all read
          </button>
        )}
      </DropdownHeader>

      {notifications.length === 0 ? (
        <DropdownItem className="text-center text-zinc-500 dark:text-zinc-400 py-4">
          No new notifications.
        </DropdownItem>
      ) : (
        notifications.map((notification) => {
          const { acceptUrl, declineUrl } = getActionUrls(notification);
          const canPerformAction = (acceptUrl || declineUrl) && !notification.is_read;

          return (
            <Fragment key={notification.id}>
              <div
                className={`p-3 ${
                  !notification.is_read ? 'bg-blue-50 dark:bg-blue-900/30' : ''
                } hover:bg-zinc-50 dark:hover:bg-zinc-700/30`}
              >
                <div className="flex items-start gap-3 cursor-pointer" onClick={() => handleItemClick(notification)}>
                  {!notification.is_read && (
                    <div className="w-2 h-2 bg-blue-500 rounded-full mt-1.5 shrink-0" title="Unread"></div>
                  )}
                  <div className={`flex-1 space-y-1 ${notification.is_read ? 'pl-4' : ''}`}>
                    <p className={`text-sm ${!notification.is_read ? 'font-semibold text-zinc-800 dark:text-zinc-100' : 'text-zinc-700 dark:text-zinc-300'}`}>
                      {notification.message}
                    </p>
                    <p className="text-xs text-zinc-500 dark:text-zinc-400">
                      {formatDistanceToNow(new Date(notification.created_at), { addSuffix: true })}
                    </p>
                  </div>
                  <div className="shrink-0 self-center">
                    {getNotificationIcon(notification.type)}
                  </div>
                </div>

                {/* Action Buttons */}
                {canPerformAction && (
                  <div className="mt-2 flex space-x-2 justify-end">
                    {acceptUrl && (
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleAction(acceptUrl, notification.id, "Request accepted");
                        }}
                        disabled={processingAction === notification.id || isActionLoading}
                        className="px-3 py-1 text-xs font-medium text-white bg-green-600 rounded-md hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500 disabled:opacity-50 flex items-center"
                      >
                        <CheckCircleIcon className="h-4 w-4 mr-1" /> Accept
                      </button>
                    )}
                    {declineUrl && (
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleAction(declineUrl, notification.id, "Request declined");
                        }}
                        disabled={processingAction === notification.id || isActionLoading}
                        className="px-3 py-1 text-xs font-medium text-white bg-red-600 rounded-md hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:opacity-50 flex items-center"
                      >
                        <XCircleIcon className="h-4 w-4 mr-1" /> Decline
                      </button>
                    )}
                  </div>
                )}
              </div>
              <DropdownDivider />
            </Fragment>
          )
        })
      )}
      <DropdownDivider />
      <DropdownItem href="/notifications" className="text-center text-blue-600 hover:text-blue-800 dark:text-blue-400 dark:hover:text-blue-300 py-3"> {/* Increased padding */}
        View all notifications
      </DropdownItem>
    </DropdownMenu>
  )
}
