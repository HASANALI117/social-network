package models

import "time"

// ChatPartner represents a user with whom the current user has exchanged messages
type ChatPartner struct {
    ID            string    `json:"id"`
    FirstName     string    `json:"first_name"`
    LastName      string    `json:"last_name"`
    Username      string    `json:"username"`
    AvatarURL     string    `json:"avatar_url"`
    LastMessage   string    `json:"last_message"`
    LastMessageAt time.Time `json:"last_message_at"`
}