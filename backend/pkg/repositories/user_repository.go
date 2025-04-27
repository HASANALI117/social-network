package repositories

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/HASANALI117/social-network/pkg/models"
	"github.com/mattn/go-sqlite3" // Import sqlite3 driver for error checking
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(user *models.User) error
	GetByID(id string) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	Delete(id string) error
	List(limit, offset int) ([]*models.User, error)
	UpdatePrivacy(userID string, isPrivate bool) error // Added method
}

// userRepository implements UserRepository interface
type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Create(user *models.User) error {
	query := `
INSERT INTO users (id, username, email, password_hash, first_name, last_name, avatar_url, about_me, birth_date, is_private, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`
	// Added is_private to INSERT

	_, err := r.db.Exec(
		query,
		user.ID,
		user.Username,
		user.Email,
		user.Password, // Assuming Password field holds the hash
		user.FirstName,
		user.LastName,
		user.AvatarURL,
		user.AboutMe,
		user.BirthDate,
		user.IsPrivate, // Added is_private value
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			// Check for unique constraint violation (code 19 -> SQLITE_CONSTRAINT)
			if sqliteErr.Code == sqlite3.ErrConstraint && strings.Contains(sqliteErr.Error(), "UNIQUE constraint failed: users.email") {
				return ErrUserAlreadyExists // Return specific error for duplicate email
			}
			if sqliteErr.Code == sqlite3.ErrConstraint && strings.Contains(sqliteErr.Error(), "UNIQUE constraint failed: users.username") {
				return ErrUserAlreadyExists // Could define a more specific ErrUsernameExists
			}
		}
		// Return generic error for other DB issues
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *userRepository) GetByID(id string) (*models.User, error) {
	query := `
SELECT id, username, email, password_hash, first_name, last_name, avatar_url, about_me, birth_date, is_private, created_at, updated_at
FROM users
WHERE id = ?
`
	// Added is_private to SELECT

	var user models.User
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password, // Scans password_hash into Password field
		&user.FirstName,
		&user.LastName,
		&user.AvatarURL,
		&user.AboutMe,
		&user.BirthDate,
		&user.IsPrivate, // Added is_private scan target
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

func (r *userRepository) GetByUsername(username string) (*models.User, error) {
	query := `
SELECT id, username, email, password_hash, first_name, last_name, avatar_url, about_me, birth_date, is_private, created_at, updated_at
FROM users
WHERE username = ?
`
	// Added is_private to SELECT

	var user models.User
	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password, // Scans password_hash into Password field
		&user.FirstName,
		&user.LastName,
		&user.AvatarURL,
		&user.AboutMe,
		&user.BirthDate,
		&user.IsPrivate, // Added is_private scan target
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	query := `
SELECT id, username, email, password_hash, first_name, last_name, avatar_url, about_me, birth_date, is_private, created_at, updated_at
FROM users
WHERE email = ?
`
	// Added is_private to SELECT

	var user models.User
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password, // Scans password_hash into Password field
		&user.FirstName,
		&user.LastName,
		&user.AvatarURL,
		&user.AboutMe,
		&user.BirthDate,
		&user.IsPrivate, // Added is_private scan target
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

func (r *userRepository) Update(user *models.User) error {
	// Note: This updates all fields based on the provided user struct.
	// Consider if partial updates are needed. Password hash should only be updated
	// if a new password is provided (logic likely belongs in the service layer).
	// is_private is included here, but might be better handled by UpdatePrivacy.
	query := `
UPDATE users
SET username = ?, email = ?, password_hash = ?, first_name = ?, last_name = ?, avatar_url = ?, about_me = ?, birth_date = ?, is_private = ?, updated_at = ?
WHERE id = ?
`
	// Added is_private to SET clause

	result, err := r.db.Exec(
		query,
		user.Username,
		user.Email,
		user.Password, // Assumes user.Password contains the hash to update
		user.FirstName,
		user.LastName,
		user.AvatarURL,
		user.AboutMe,
		user.BirthDate,
		user.IsPrivate, // Added is_private value
		time.Now(),
		user.ID,
	)

	if err != nil {
		// Add check for constraint violations if email/username are updated
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				// Could check specific constraint message if needed
				return ErrUserAlreadyExists // Or a more specific error
			}
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// Log this error but don't necessarily fail the operation
		fmt.Printf("Warning: failed to get rows affected after updating user %s: %v\n", user.ID, err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound // No user found with the given ID
	}

	return nil
}

func (r *userRepository) Delete(id string) error {
	query := "DELETE FROM users WHERE id = ?"

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after delete: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *userRepository) List(limit, offset int) ([]*models.User, error) {
	query := `
SELECT id, username, email, password_hash, first_name, last_name, avatar_url, about_me, birth_date, is_private, created_at, updated_at
FROM users
ORDER BY created_at DESC
LIMIT ? OFFSET ?
`
	// Added is_private to SELECT

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.Password, // Scans password_hash into Password field
			&user.FirstName,
			&user.LastName,
			&user.AvatarURL,
			&user.AboutMe,
			&user.BirthDate,
			&user.IsPrivate, // Added is_private scan target
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			// Log error and continue? Or return immediately?
			return nil, fmt.Errorf("failed to scan user during list: %w", err)
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user list rows: %w", err)
	}

	return users, nil
}

// UpdatePrivacy updates the is_private status for a user
func (r *userRepository) UpdatePrivacy(userID string, isPrivate bool) error {
	query := `
UPDATE users
SET is_private = ?, updated_at = ?
WHERE id = ?
`
	result, err := r.db.Exec(query, isPrivate, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update user privacy: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// Log the error but potentially continue, as the update might have succeeded
		// depending on the DB driver's behavior.
		fmt.Printf("Warning: failed to get rows affected after updating privacy for user %s: %v\n", userID, err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound // User ID did not match any rows
	}

	return nil
}
