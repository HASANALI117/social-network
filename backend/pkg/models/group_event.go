package models

import "time"

// GroupEvent represents a group event with details like title, description, and scheduled time
type GroupEvent struct {
	ID          string    `json:"id"`
	GroupID     string    `json:"group_id"`
	CreatorID   string    `json:"creator_id"` // Changed from int to string
	Title       string    `json:"title"`
	Description string    `json:"description"`
	EventTime   time.Time `json:"event_time"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// EventResponseAPI represents the enriched response data for an event attendee
type EventResponseAPI struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	AvatarURL string `json:"avatar_url"`
	Response  string `json:"response"`
	UpdatedAt string `json:"updated_at"` // Or time.Time if handled as such
}

// GroupEventAPI represents the group event details along with enriched responses
type GroupEventAPI struct {
	GroupEvent
	Responses []EventResponseAPI `json:"responses,omitempty"`
}
