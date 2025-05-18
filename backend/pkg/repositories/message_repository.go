package repositories

import (
	"database/sql"
	"fmt" // Added import
	"log"
	"github.com/HASANALI117/social-network/pkg/models"
)

// MessageRepository defines the interface for message data operations.
// It now includes methods previously expected from ChatMessageRepository.
type MessageRepository interface {
	GetChatPartners(currentUserID string) ([]models.ChatPartner, error)
	GetDirectMessagesBetweenUsers(user1ID, user2ID string, limit, offset int) ([]models.Message, int64, error)
	GetGroupMessages(groupID string, limit, offset int, requestingUserID string) ([]*models.GroupMessage, error) // Corrected parameter name
	// Add other message-related methods here if any, e.g., CreateMessage
}

type sqliteMessageRepository struct {
	db *sql.DB
}

// NewMessageRepository creates a new instance of MessageRepository.
func NewMessageRepository(db *sql.DB) MessageRepository {
	return &sqliteMessageRepository{db: db}
}

// GetChatPartners retrieves a list of users with whom the current user has a chat history,
// sorted by the most recent message.
func (r *sqliteMessageRepository) GetChatPartners(currentUserID string) ([]models.ChatPartner, error) {
	query := `
    SELECT
        u.id,
        u.first_name,
        u.last_name,
        u.username,
        COALESCE(u.avatar_url, '') AS avatar_url, -- Handle NULL avatar_url
        m_last.content AS last_message,
        m_last.created_at AS last_message_at
    FROM users u
    JOIN (
        -- Subquery to find distinct chat partners
        SELECT DISTINCT
            CASE
                WHEN sender_id = ?1 THEN receiver_id
                ELSE sender_id
            END AS partner_id
        FROM messages
        WHERE sender_id = ?1 OR receiver_id = ?1
    ) partners ON u.id = partners.partner_id
    JOIN (
        -- Subquery to get the last message for each partner
        SELECT
            CASE
                WHEN sender_id = ?1 THEN receiver_id
                ELSE sender_id
            END AS partner_id,
            content,
            created_at,
            ROW_NUMBER() OVER (PARTITION BY
                CASE
                    WHEN sender_id = ?1 THEN receiver_id
                    ELSE sender_id
                END
                ORDER BY created_at DESC
            ) as rn
        FROM messages
        WHERE (sender_id = ?1 OR receiver_id = ?1)
    ) m_last ON u.id = m_last.partner_id AND m_last.rn = 1
    ORDER BY m_last.created_at DESC;
    `

	rows, err := r.db.Query(query, currentUserID)
	if err != nil {
		log.Printf("Error querying chat partners for user %s: %v", currentUserID, err)
		return nil, err
	}
	defer rows.Close()

	var chatPartners []models.ChatPartner
	for rows.Next() {
		var cp models.ChatPartner
		if err := rows.Scan(
			&cp.ID,
			&cp.FirstName,
			&cp.LastName,
			&cp.Username,
			&cp.AvatarURL,
			&cp.LastMessage,
			&cp.LastMessageAt,
		); err != nil {
			log.Printf("Error scanning chat partner row: %v", err)
			return nil, err
		}
		chatPartners = append(chatPartners, cp)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Error iterating chat partner rows: %v", err)
		return nil, err
	}

	return chatPartners, nil
}

// GetDirectMessagesBetweenUsers is a stub implementation to satisfy the MessageRepository interface.
// TODO: Replace with actual logic if this repository is meant to handle direct messages.
func (r *sqliteMessageRepository) GetDirectMessagesBetweenUsers(user1ID, user2ID string, limit, offset int) ([]models.Message, int64, error) {
	log.Printf("STUB: GetDirectMessagesBetweenUsers called for user1: %s, user2: %s, limit: %d, offset: %d", user1ID, user2ID, limit, offset)
	// This is a placeholder. Real implementation would query the database.
	// If another repository (e.g., an actual ChatMessageRepository implementation) handles this,
	// this sqliteMessageRepository might not be the correct one to use in init.go,
	// or it needs to be properly implemented.
	return []models.Message{}, 0, fmt.Errorf("GetDirectMessagesBetweenUsers not implemented in this version of sqliteMessageRepository")
}

// GetGroupMessages is a stub implementation to satisfy the MessageRepository interface.
// TODO: Replace with actual logic if this repository is meant to handle group messages.
func (r *sqliteMessageRepository) GetGroupMessages(groupID string, limit, offset int, requestingUserID string) ([]*models.GroupMessage, error) {
	log.Printf("STUB: GetGroupMessages called for groupID: %s, limit: %d, offset: %d, user: %s", groupID, limit, offset, requestingUserID)
	// Placeholder
	return []*models.GroupMessage{}, fmt.Errorf("GetGroupMessages not implemented in this version of sqliteMessageRepository")
}

// Note: Ensure your 'messages' table has 'sender_id', 'receiver_id', 'content', 'created_at'
// and 'users' table has 'id', 'first_name', 'last_name', 'username', 'avatar_url'.
// The COALESCE(u.avatar_url, '') is added to prevent errors if avatar_url is NULL.