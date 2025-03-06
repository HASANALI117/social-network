package helpers

import (
	"database/sql"
	"errors"
	"time"

	"github.com/HASANALI117/social-network/pkg/db"
	"github.com/google/uuid"
)

var (
	ErrInvalidSession = errors.New("invalid or expired session")
)

// CreateSession creates a new session for a user
func CreateSession(userID string, duration time.Duration) (string, error) {
	token := uuid.New().String()
	expiresAt := time.Now().Add(duration)

	query := `
        INSERT INTO sessions (token, user_id, expires_at)
        VALUES (?, ?, ?)
    `

	_, err := db.GlobalDB.Exec(query, token, userID, expiresAt)
	if err != nil {
		return "", err
	}

	return token, nil
}

// GetSession retrieves a session by token
func GetSession(token string) (string, error) {
	query := `
        SELECT user_id FROM sessions 
        WHERE token = ? AND expires_at > datetime('now')
    `

	var userID string
	err := db.GlobalDB.QueryRow(query, token).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrInvalidSession
		}
		return "", err
	}

	return userID, nil
}

// DeleteSession removes a session by token
func DeleteSession(token string) error {
	query := `DELETE FROM sessions WHERE token = ?`
	_, err := db.GlobalDB.Exec(query, token)
	return err
}

// CleanExpiredSessions removes all expired sessions
func CleanExpiredSessions() error {
	query := `DELETE FROM sessions WHERE expires_at <= datetime('now')`
	_, err := db.GlobalDB.Exec(query)
	return err
}
