package models

import "time"

// Comment represents a comment on a post
type Comment struct {
	ID        string    `json:"id"`
	PostID    string    `json:"post_id"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	// TODO: Consider adding User details (username, avatar) if needed for responses
}
