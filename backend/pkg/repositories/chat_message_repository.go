package repositories

import (
	"database/sql"
	"fmt" // Added for error wrapping

	"github.com/HASANALI117/social-network/pkg/models"
)

// ChatMessageRepository defines the interface for chat message data operations.
type ChatMessageRepository interface {
	SaveDirectMessage(message *models.Message) error
	SaveGroupMessage(message *models.GroupMessage) error
	// GetDirectMessagesBetweenUsers retrieves paginated direct messages between two users and the total count.
	GetDirectMessagesBetweenUsers(user1ID, user2ID string, limit, offset int) ([]models.Message, int64, error)
	GetGroupMessages(groupID string, limit int, offset int, currentUserID string) ([]*models.GroupMessage, error)
	GetChatPartners(currentUserID string) ([]models.ChatPartner, error) // Added method
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
func (r *chatMessageRepository) SaveDirectMessage(message *models.Message) error {
	query := `INSERT INTO messages (id, sender_id, receiver_id, content, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(query, message.ID, message.SenderID, message.ReceiverID, message.Content, message.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save direct message: %w", err) // Added error wrapping
	}
	return nil
}

// SaveGroupMessage saves a group message to the database.
func (r *chatMessageRepository) SaveGroupMessage(message *models.GroupMessage) error {
	query := `INSERT INTO group_messages (id, group_id, sender_id, content, created_at) VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.Exec(query, message.ID, message.GroupID, message.SenderID, message.Content, message.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save group message: %w", err) // Added error wrapping
	}
	return nil
}

// GetDirectMessagesBetweenUsers retrieves paginated direct messages between two users and the total count.
func (r *chatMessageRepository) GetDirectMessagesBetweenUsers(user1ID, user2ID string, limit, offset int) ([]models.Message, int64, error) {
	var totalCount int64
	var messages []models.Message

	// Query to get the total count
	countQuery := `
		SELECT COUNT(*)
		FROM messages
		WHERE (sender_id = $1 AND receiver_id = $2)
		   OR (sender_id = $3 AND receiver_id = $4)
	`
	err := r.db.QueryRow(countQuery, user1ID, user2ID, user2ID, user1ID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count direct messages: %w", err)
	}

	// If count is 0, no need to query for messages
	if totalCount == 0 {
		return messages, 0, nil
	}

	// Query to get the paginated messages
	messagesQuery := `
		SELECT id, sender_id, receiver_id, content, created_at
		FROM messages
		WHERE (sender_id = $1 AND receiver_id = $2)
		   OR (sender_id = $3 AND receiver_id = $4)
		ORDER BY created_at DESC
		LIMIT $5 OFFSET $6
	`
	rows, err := r.db.Query(messagesQuery, user1ID, user2ID, user2ID, user1ID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query direct messages: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var msg models.Message // Use value type as returned by service
		err := rows.Scan(
			&msg.ID,
			&msg.SenderID,
			&msg.ReceiverID,
			&msg.Content,
			&msg.CreatedAt,
		)
		if err != nil {
			// Log or handle scan error appropriately
			return nil, 0, fmt.Errorf("failed to scan direct message row: %w", err)
		}
		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating direct message rows: %w", err)
	}

	return messages, totalCount, nil
}

// GetGroupMessages retrieves messages for a group with pagination.
func (r *chatMessageRepository) GetGroupMessages(groupID string, limit int, offset int, currentUserID string) ([]*models.GroupMessage, error) {
	// TODO: Implement actual logic for fetching group messages for ChatMessageRepository if needed
	// This is a stub to satisfy the interface.
	// The primary implementation is likely in sqlite_message_repository.go
	// log.Printf("ChatMessageRepository: GetGroupMessages called with groupID: %s, limit: %d, offset: %d, currentUserID: %s (STUB)", groupID, limit, offset, currentUserID)
	// return []*models.GroupMessage{}, nil
	query := `
	       SELECT id, group_id, sender_id, content, created_at
	       FROM group_messages
	       WHERE group_id = $1
	       ORDER BY created_at DESC
	       LIMIT $2 OFFSET $3
	   ` // Using $ placeholders for PostgreSQL compatibility

	rows, err := r.db.Query(query, groupID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query group messages: %w", err) // Added error wrapping
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
			return nil, fmt.Errorf("failed to scan group message row: %w", err) // Added error wrapping
		}

		// Assign the ID only if it's not NULL in the database
		if nullableID.Valid {
			msg.ID = nullableID.String
		} else {
			msg.ID = "" // Or handle NULL ID case as appropriate (e.g., log warning)
		}

		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating group message rows: %w", err) // Added error wrapping
	}

	return messages, nil
}

// GetChatPartners is a stub implementation to satisfy the ChatMessageRepository interface
// which now aligns with MessageRepository.
// TODO: This method is fully implemented in sqliteMessageRepository.
// If chatMessageRepository is intended to be the primary message repository,
// this stub should be replaced with a proper implementation or delegate to
// the one in sqliteMessageRepository if appropriate.
func (r *chatMessageRepository) GetChatPartners(currentUserID string) ([]models.ChatPartner, error) {
    query := `
        WITH LastMessages AS (
            SELECT
                CASE
                    WHEN sender_id = ? THEN receiver_id
                    ELSE sender_id
                END AS partner_id,
                content as last_message,
                created_at as last_message_at,
                ROW_NUMBER() OVER (
                    PARTITION BY
                        CASE
                            WHEN sender_id = ? THEN receiver_id
                            ELSE sender_id
                        END
                    ORDER BY created_at DESC
                ) as rn
            FROM messages
            WHERE sender_id = ? OR receiver_id = ?
        )
        SELECT
            u.id,
            u.first_name,
            u.last_name,
            u.username,
            u.avatar_url,
            lm.last_message,
            lm.last_message_at
        FROM LastMessages lm
        JOIN users u ON u.id = lm.partner_id
        WHERE lm.rn = 1
        ORDER BY lm.last_message_at DESC;
    `

    rows, err := r.db.Query(query, currentUserID, currentUserID, currentUserID, currentUserID)
    if err != nil {
        return nil, fmt.Errorf("failed to get chat partners: %w", err)
    }
    defer rows.Close()

    var partners []models.ChatPartner
    for rows.Next() {
        var partner models.ChatPartner
        err := rows.Scan(
            &partner.ID,
            &partner.FirstName,
            &partner.LastName,
            &partner.Username,
            &partner.AvatarURL,
            &partner.LastMessage,
            &partner.LastMessageAt,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan chat partner: %w", err)
        }
        partners = append(partners, partner)
    }

    if err = rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating chat partners: %w", err)
    }

    return partners, nil
}
