package models

import "time"

// Session represents a user session in the database
type Session struct {
	Token     string    `db:"token"`
	UserID    string    `db:"user_id"`
	ExpiresAt time.Time `db:"expires_at"`
}
