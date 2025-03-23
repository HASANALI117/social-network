package helpers

import (
	"time"

	"social-network/pkg/db"
	"github.com/google/uuid"
)

type Message struct {
	ID         string `json:"id"`
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"receiver_id"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
}

func SaveMessage(msg *Message) error {
	query := `
        INSERT INTO messages (id, sender_id, receiver_id, content, created_at)
        VALUES (?, ?, ?, ?, ?)
    `
	msg.ID = uuid.New().String()
	if msg.CreatedAt == "" {
		msg.CreatedAt = time.Now().Format(time.RFC3339)
	}
	_, err := db.GlobalDB.Exec(query, msg.ID, msg.SenderID, msg.ReceiverID, msg.Content, msg.CreatedAt)
	return err
}
