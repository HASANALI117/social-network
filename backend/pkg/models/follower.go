package models

import "time"

// Follower represents a follow relationship between users
type Follower struct {
	FollowerID  string    `json:"follower_id" db:"follower_id"`
	FollowingID string    `json:"following_id" db:"following_id"`
	Status      string    `json:"status" db:"status"` // "pending" or "accepted"
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
