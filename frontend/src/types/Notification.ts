// Defines the structure for a single notification
export interface Notification {
  id: string;
  user_id: string;
  type: string; // e.g., "follow_request", "follow_accept", "group_invite"
  entity_type: string; // e.g., "user", "group"
  message: string;
  entity_id: string; // ID of the entity related to the notification (e.g., user ID for follow request, group ID for invite)
  is_read: boolean;
  created_at: string; // ISO string date
}

// Defines the structure for the API response when fetching multiple notifications
export interface NotificationsResponse {
  notifications: Notification[];
  unread_count: number;
  limit: number;
  offset: number;
  has_more: boolean;
}