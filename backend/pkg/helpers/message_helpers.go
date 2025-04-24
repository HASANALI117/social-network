package helpers

import (
	"time"

	"github.com/HASANALI117/social-network/pkg/db"
	"github.com/HASANALI117/social-network/pkg/models"
	"github.com/google/uuid"
)

func SaveMessage(message *models.Message) error {
	query := `
        INSERT INTO messages (id, sender_id, receiver_id, content, created_at)
        VALUES (?, ?, ?, ?, ?)
    `

	message.ID = uuid.NewString()
	message.CreatedAt = time.Now().Format(time.RFC3339)

	_, err := db.GlobalDB.Exec(query, message.ID, message.SenderID, message.ReceiverID, message.Content, message.CreatedAt)
	return err
}

func GetUserMessages(senderID string, receiverID string, limit, offset int) ([]*models.Message, error) {
	query := `
        SELECT id, sender_id, receiver_id, content, created_at 
        FROM messages 
        WHERE (sender_id = ? AND receiver_id = ?) 
        OR (sender_id = ? AND receiver_id = ?)
        ORDER BY created_at DESC
		LIMIT ? OFFSET ?
    `

	rows, err := db.GlobalDB.Query(query, senderID, receiverID, receiverID, senderID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		msg := &models.Message{}
		err := rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.ReceiverID,
			&msg.Content,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}
