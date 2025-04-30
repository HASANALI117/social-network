package models

import "time"

// GroupEventResponse represents a user's response to a group event
type GroupEventResponse struct {
	ID        string    `json:"id"`
	EventID   string    `json:"event_id"`
	UserID    string    `json:"user_id"`
	Response  string    `json:"response"` // "going" or "not_going"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
