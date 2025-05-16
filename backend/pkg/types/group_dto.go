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

// PostSummary represents key details of a post.
type PostSummary struct {
	PostID         string        `json:"post_id"`
	Title          string        `json:"title,omitempty"` // Assuming posts might not always have a title
	ContentSnippet string        `json:"content_snippet"`
	CreatorInfo    UserBasicInfo `json:"creator_info"`
	CreatedAt      time.Time     `json:"created_at"`
	CommentsCount  int           `json:"comments_count"`
}

// EventSummary represents key details of an event.
type EventSummary struct {
	EventID            string    `json:"event_id"`
	Title              string    `json:"title"`
	DescriptionSnippet string    `json:"description_snippet,omitempty"`
	StartTime          time.Time `json:"start_time"`
	// Location           string    `json:"location,omitempty"` // Removed as its existence in DB is uncertain
}

// GroupDetailResponse represents the detailed information for a group,
// including creator details and counts of members, posts, and events.
// For group members, it also includes lists of members, posts, and events.
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

	// Fields for group members only
	Members []UserBasicInfo `json:"members,omitempty"`
	Posts   []PostSummary   `json:"posts,omitempty"`
	Events  []EventSummary  `json:"events,omitempty"`
}
