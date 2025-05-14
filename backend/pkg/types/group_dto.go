package types

import "time"

// UserBasicInfo represents basic information about a user.
type UserBasicInfo struct {
	UserID    string `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
}

// GroupDetailResponse represents the detailed information for a group,
// including creator details and counts of members, posts, and events.
type GroupDetailResponse struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	AvatarURL    string        `json:"avatar_url"`
	CreatorInfo  UserBasicInfo `json:"creator_info"`
	MembersCount int           `json:"members_count"`
	PostsCount   int           `json:"posts_count"`
	EventsCount  int           `json:"events_count"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}
