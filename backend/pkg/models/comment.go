package models

import "time"

// Comment represents a comment on a post
type Comment struct {
	ID        string    `json:"id"`
	PostID    string    `json:"post_id"`
	UserID    string    `json:"user_id"`
	Content   string    `json:"content"`
	ImageURL  string    `json:"image_url,omitempty"` // New field
	CreatedAt time.Time `json:"created_at"`
}
