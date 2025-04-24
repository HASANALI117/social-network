package helpers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/HASANALI117/social-network/pkg/db"
	"github.com/HASANALI117/social-network/pkg/models"
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

// GetSession retrieves user by session
func GetUserFromSession(r *http.Request) (*models.User, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return nil, fmt.Errorf("no session cookie found")
	}

	userID, err := GetSession(cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("invalid session: %v", err)
	}

	// Verify user exists in database
	user, err := GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	return user, nil
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
