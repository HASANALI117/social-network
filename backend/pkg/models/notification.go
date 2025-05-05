package models

import "time"

// NotificationType represents the type of notification.
type NotificationType string

const (
	NotificationTypeFollowRequest    NotificationType = "follow_request"
	NotificationTypeGroupInvite      NotificationType = "group_invite"
	NotificationTypeGroupJoinRequest NotificationType = "group_join_request"
	NotificationTypeNewGroupEvent    NotificationType = "new_group_event"
	// Add other notification types as needed
)

// EntityType represents the type of the entity related to the notification.
type EntityType string

const (
	EntityTypeGroup EntityType = "group"
	EntityTypeEvent EntityType = "event"
	EntityTypeUser  EntityType = "user"
	// Add other entity types as needed
)

/*
-- Database Schema for notifications table
CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,                 -- Foreign Key to the user receiving the notification
    actor_id INT,                         -- Foreign Key to the user who triggered the notification (nullable)
    type VARCHAR(50) NOT NULL,            -- Type of notification (e.g., 'follow_request', 'group_invite')
    entity_id INT,                        -- ID of the related entity (nullable)
    entity_type VARCHAR(50),              -- Type of the related entity (nullable)
    is_read BOOLEAN NOT NULL DEFAULT FALSE, -- Whether the notification has been read
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), -- Timestamp when the notification was created

    -- Optional: Add foreign key constraints if users table exists
    -- FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    -- FOREIGN KEY (actor_id) REFERENCES users(id) ON DELETE SET NULL

    -- Optional: Add indexes for performance
    INDEX idx_notifications_user_id (user_id),
    INDEX idx_notifications_is_read (is_read)
);
*/

// Notification represents a notification entry in the database.
type Notification struct {
	ID         int              `db:"id" json:"id"`
	UserID     int              `db:"user_id" json:"user_id"`
	ActorID    *int             `db:"actor_id" json:"actor_id,omitempty"` // Pointer for nullable field
	Type       NotificationType `db:"type" json:"type"`
	EntityID   *int             `db:"entity_id" json:"entity_id,omitempty"`     // Pointer for nullable field
	EntityType *EntityType      `db:"entity_type" json:"entity_type,omitempty"` // Pointer for nullable field
	IsRead     bool             `db:"is_read" json:"is_read"`
	CreatedAt  time.Time        `db:"created_at" json:"created_at"`
}
