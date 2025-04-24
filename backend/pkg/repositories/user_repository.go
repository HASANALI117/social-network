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
INSERT INTO users (id, username, email, password_hash, first_name, last_name, avatar_url, about_me, birth_date, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`

	_, err := r.db.Exec(
		query,
		user.ID,
		user.Username,
		user.Email,
		user.Password,
		user.FirstName,
		user.LastName,
		user.AvatarURL,
		user.AboutMe,
		user.BirthDate,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			// Check for unique constraint violation (code 19 -> SQLITE_CONSTRAINT)
			// Further check the error message for specific constraint (e.g., users.email)
			if sqliteErr.Code == sqlite3.ErrConstraint && strings.Contains(sqliteErr.Error(), "UNIQUE constraint failed: users.email") {
				return ErrUserAlreadyExists // Return specific error for duplicate email
			}
			// Can add checks for other constraints like username if needed
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
		SELECT id, username, email, password_hash, first_name, last_name, avatar_url, about_me, birth_date, created_at, updated_at
		FROM users
		WHERE id = ?
	`

	var user models.User
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.AvatarURL,
		&user.AboutMe,
		&user.BirthDate,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *userRepository) GetByUsername(username string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name, avatar_url, about_me, birth_date, created_at, updated_at
		FROM users
		WHERE username = ?
	`

	var user models.User
	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.AvatarURL,
		&user.AboutMe,
		&user.BirthDate,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name, avatar_url, about_me, birth_date, created_at, updated_at
		FROM users
		WHERE email = ?
	`

	var user models.User
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.AvatarURL,
		&user.AboutMe,
		&user.BirthDate,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *userRepository) Update(user *models.User) error {
	query := `
		UPDATE users
		SET username = ?, email = ?, password_hash = ?, first_name = ?, last_name = ?, avatar_url = ?, about_me = ?, birth_date = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.Exec(
		query,
		user.Username,
		user.Email,
		user.Password,
		user.FirstName,
		user.LastName,
		user.AvatarURL,
		user.AboutMe,
		user.BirthDate,
		time.Now(),
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
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
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *userRepository) List(limit, offset int) ([]*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, first_name, last_name, avatar_url, about_me, birth_date, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

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
			&user.Password,
			&user.FirstName,
			&user.LastName,
			&user.AvatarURL,
			&user.AboutMe,
			&user.BirthDate,
			&user.CreatedAt,
			&user.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}
