package repositories

import (
	"database/sql"

	"github.com/HASANALI117/social-network/pkg/models"
)

// ChatMessageRepository defines the interface for chat message data operations.
type ChatMessageRepository interface {
	SaveDirectMessage(message *models.Message) error
	SaveGroupMessage(message *models.GroupMessage) error
	GetUserMessages(senderID string, receiverID string, limit, offset int) ([]*models.Message, error)
	GetGroupMessages(groupID string, limit, offset int) ([]*models.GroupMessage, error) // Added
}

// chatMessageRepository implements the ChatMessageRepository interface.
type chatMessageRepository struct {
	db *sql.DB
}

// NewChatMessageRepository creates a new instance of ChatMessageRepository.
func NewChatMessageRepository(db *sql.DB) ChatMessageRepository {
	return &chatMessageRepository{db: db}
}

// SaveDirectMessage saves a direct message to the database.
// TODO: Implement the actual SQL query to insert the message.
func (r *chatMessageRepository) SaveDirectMessage(message *models.Message) error {
	// Placeholder implementation - replace with actual DB logic
	// Insert the message ID along with other fields into the correct 'messages' table.
	query := `INSERT INTO messages (id, sender_id, receiver_id, content, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(query, message.ID, message.SenderID, message.ReceiverID, message.Content, message.CreatedAt)
	if err != nil {
		// Add proper error handling/logging
		return err
	}
	return nil
}

// SaveGroupMessage saves a group message to the database.
// TODO: Implement the actual SQL query to insert the message.
func (r *chatMessageRepository) SaveGroupMessage(message *models.GroupMessage) error {
	// Placeholder implementation - replace with actual DB logic
	query := `INSERT INTO group_messages (id, group_id, sender_id, content, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(query, message.ID, message.GroupID, message.SenderID, message.Content, message.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}

// GetUserMessages retrieves direct messages between two users with pagination.
func (r *chatMessageRepository) GetUserMessages(senderID string, receiverID string, limit, offset int) ([]*models.Message, error) {
	query := `
	       SELECT id, sender_id, receiver_id, content, created_at
	       FROM messages
	       WHERE (sender_id = ? AND receiver_id = ?)
	       OR (sender_id = ? AND receiver_id = ?)
	       ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	   `

	rows, err := r.db.Query(query, senderID, receiverID, receiverID, senderID, limit, offset)
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

// GetGroupMessages retrieves messages for a group with pagination.
func (r *chatMessageRepository) GetGroupMessages(groupID string, limit, offset int) ([]*models.GroupMessage, error) {
	query := `
	       SELECT id, group_id, sender_id, content, created_at
	       FROM group_messages
	       WHERE group_id = ?
	       ORDER BY created_at DESC
	       LIMIT ? OFFSET ?
	   `

	rows, err := r.db.Query(query, groupID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	messages := make([]*models.GroupMessage, 0)
	for rows.Next() {
		msg := &models.GroupMessage{}
		var nullableID sql.NullString // Use sql.NullString for the ID

		err := rows.Scan(
			&nullableID, // Scan into the NullString
			&msg.GroupID,
			&msg.SenderID,
			&msg.Content,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, err // Return scan errors immediately
		}

		// Assign the ID only if it's not NULL in the database
		if nullableID.Valid {
			msg.ID = nullableID.String
		} else {
			msg.ID = "" // Or handle NULL ID case as appropriate (e.g., log warning)
		}

		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}
