package models

import "time"

// Group represents a chat group
type Group struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatorID   string    `json:"creator_id"`
	AvatarURL   string    `json:"avatar_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GroupMember represents a user membership in a group
type GroupMember struct {
	GroupID  string    `json:"group_id"`
	UserID   string    `json:"user_id"`
	Role     string    `json:"role"` // admin or member
	JoinedAt time.Time `json:"joined_at"`
}

// GroupMessage represents a message in a group chat
type GroupMessage struct {
	ID        string    `json:"id"`
	GroupID   string    `json:"group_id"`
	SenderID  string    `json:"sender_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"` // Changed to time.Time for consistency
}

// GroupInvitation represents an invitation for a user to join a group
type GroupInvitation struct {
	ID        string    `json:"id"`
	GroupID   string    `json:"group_id"`
	InviterID string    `json:"inviter_id"` // User who sent the invitation
	InviteeID string    `json:"invitee_id"` // User who received the invitation
	Status    string    `json:"status"`     // e.g., "pending", "accepted", "rejected"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GroupJoinRequest represents a request from a user to join a group
type GroupJoinRequest struct {
	ID          string    `json:"id"`
	GroupID     string    `json:"group_id"`
	RequesterID string    `json:"requester_id"` // User requesting to join
	Status      string    `json:"status"`       // e.g., "pending", "accepted", "rejected"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
