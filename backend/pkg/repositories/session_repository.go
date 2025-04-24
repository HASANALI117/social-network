package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/HASANALI117/social-network/pkg/models"
)

var (
	// ErrSessionNotFound indicates that a session token is invalid or expired.
	ErrSessionNotFound = errors.New("session not found or expired")
)

// SessionRepository defines the interface for session data access
type SessionRepository interface {
	Create(session *models.Session) error
	GetByToken(token string) (*models.Session, error)
	DeleteByToken(token string) error
	CleanExpired() error
}

// sessionRepository implements SessionRepository interface
type sessionRepository struct {
	db *sql.DB
}

// NewSessionRepository creates a new SessionRepository
func NewSessionRepository(db *sql.DB) SessionRepository {
	return &sessionRepository{
		db: db,
	}
}

// Create inserts a new session record into the database
func (r *sessionRepository) Create(session *models.Session) error {
	query := `
        INSERT INTO sessions (token, user_id, expires_at)
        VALUES (?, ?, ?)
    `
	_, err := r.db.Exec(query, session.Token, session.UserID, session.ExpiresAt)
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	return nil
}

// GetByToken retrieves an active session by its token
func (r *sessionRepository) GetByToken(token string) (*models.Session, error) {
	query := `
        SELECT token, user_id, expires_at 
        FROM sessions 
        WHERE token = ? AND expires_at > ?
    `
	var session models.Session
	// Use time.Now() for comparison against expires_at
	err := r.db.QueryRow(query, token, time.Now()).Scan(&session.Token, &session.UserID, &session.ExpiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session by token: %w", err)
	}
	return &session, nil
}

// DeleteByToken removes a session record by its token
func (r *sessionRepository) DeleteByToken(token string) error {
	query := `DELETE FROM sessions WHERE token = ?`
	result, err := r.db.Exec(query, token)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// Log this error but don't necessarily fail the operation,
		// as the session might have already been deleted or expired.
		fmt.Printf("Warning: failed to get rows affected after deleting session: %v\n", err)
	}
	if rowsAffected == 0 {
		// This isn't necessarily an error, could just mean the token didn't exist.
		// Depending on requirements, you might return ErrSessionNotFound or just nil.
		// Returning nil is often simpler for logout scenarios.
	}
	return nil // Return nil even if no rows affected, as the goal (session gone) is achieved
}

// CleanExpired removes all expired sessions from the database
func (r *sessionRepository) CleanExpired() error {
	query := `DELETE FROM sessions WHERE expires_at <= ?`
	_, err := r.db.Exec(query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to clean expired sessions: %w", err)
	}
	return nil
}
