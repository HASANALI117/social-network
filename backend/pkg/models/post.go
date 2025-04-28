package models

import (
	"database/sql"
	"time"
)

// Privacy levels constants
const (
	PrivacyPublic        = "public"
	PrivacyAlmostPrivate = "semi_private" // Followers only
	PrivacyPrivate       = "private"      // Specific users only
)

type Post struct {
	ID           string         `json:"id"`
	UserID       string         `json:"user_id"`
	Title        string         `json:"title"`
	Content      string         `json:"content"`
	ImageURL     string         `json:"image_url,omitempty"`
	Privacy      string         `json:"privacy"`            // Should be one of the constants above
	GroupID      sql.NullString `json:"group_id,omitempty"` // Nullable foreign key to groups table
	CreatedAt    time.Time      `json:"created_at"`
	AllowedUsers []string       `json:"-" db:"-"` // Not stored in posts table, populated separately for private posts
}
