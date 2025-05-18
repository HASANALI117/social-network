package models

import (
	"time"
)

// NotificationType defines the type of notification.
type NotificationType string

// EntityType defines the type of entity a notification might refer to.
type EntityType string

const (
	FollowRequestNotification NotificationType = "follow_request"
	FollowAcceptNotification  NotificationType = "follow_accept" // Added for when a follow request is accepted
	GroupInviteNotification   NotificationType = "group_invite"
	GroupJoinRequestNotification NotificationType = "group_join_request"
	GroupEventCreatedNotification NotificationType = "group_event_created"
	// Add other notification types here in the future
)

const (
	UserEntityType  EntityType = "user"
	GroupEntityType EntityType = "group"
	EventEntityType EntityType = "event"
	// Add other entity types here
)

// Notification represents a notification in the system.
type Notification struct {
	ID         string           `json:"id" db:"id"`
	UserID     string           `json:"user_id" db:"user_id"`           // Recipient of the notification
	Type       NotificationType `json:"type" db:"type"`                 // Type of notification (e.g., "follow_request")
	EntityType EntityType       `json:"entity_type" db:"entity_type"`   // Type of the entity this notification refers to
	Message    string           `json:"message" db:"message"`             // User-friendly message
	EntityID   string           `json:"entity_id" db:"entity_id"`         // ID of the related entity (e.g., follower's UserID, GroupID, EventID)
	IsRead     bool             `json:"is_read" db:"is_read"`             // Whether the notification has been read
	CreatedAt  time.Time        `json:"created_at" db:"created_at"`       // Timestamp of creation
}