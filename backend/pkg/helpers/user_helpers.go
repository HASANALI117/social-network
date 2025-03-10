package helpers

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/HASANALI117/social-network/pkg/db"
	"github.com/HASANALI117/social-network/pkg/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

// Create a new user to the database
func CreateUser(user *models.User) error {
	query := `
        INSERT INTO users (id, username, email, password_hash, first_name, last_name, avatar_url, about_me, birth_date, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `

	// Generate password hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.ID = uuid.New().String()
	user.Password = string(hashedPassword)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err = db.GlobalDB.Exec(
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
		// Check if it's a duplicate entry error
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func GetUserByID(id string) (*models.User, error) {
	query := `
        SELECT id, username, email, password_hash, first_name, last_name, avatar_url, about_me, birth_date, created_at, updated_at
        FROM users
        WHERE id = ?
    `

	var user models.User
	var createdAt, updatedAt string

	err := db.GlobalDB.QueryRow(query, id).Scan(
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

	// Parse timestamps
	user.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	user.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return &user, nil
}

// GetByUsername retrieves a user by username
func GetUserByUsername(username string) (*models.User, error) {
	query := `
        SELECT id, username, email, password_hash, first_name, last_name, avatar_url, about_me, birth_date, created_at, updated_at
        FROM users
        WHERE username = ?
    `

	var user models.User
	var createdAt, updatedAt string

	err := db.GlobalDB.QueryRow(query, username).Scan(
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

	// Parse timestamps
	user.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	user.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return &user, nil
}

// GetByEmail retrieves a user by email
func GetUserByEmail(email string) (*models.User, error) {
	query := `
        SELECT id, username, email, password_hash, first_name, last_name, avatar_url, about_me, birth_date, created_at, updated_at
        FROM users
        WHERE email = ?
    `

	var user models.User
	var createdAt, updatedAt string

	err := db.GlobalDB.QueryRow(query, email).Scan(
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

	// Parse timestamps
	user.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	user.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return &user, nil
}

// Update an existing user
func UpdateUser(user *models.User) error {
	query := `
        UPDATE users
        SET username = ?, email = ?, password_hash = ?, first_name = ?, last_name = ?, avatar_url = ?, about_me = ?, birth_date = ?, updated_at = ?
        WHERE id = ?
    `

	user.UpdatedAt = time.Now()

	result, err := db.GlobalDB.Exec(
		query,
		user.Username,
		user.Email,
		user.Password,
		user.FirstName,
		user.LastName,
		user.AvatarURL,
		user.AboutMe,
		user.BirthDate,
		user.UpdatedAt,
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

// Delete removes a user
func DeleteUser(id string) error {
	query := "DELETE FROM users WHERE id = ?"

	result, err := db.GlobalDB.Exec(query, id)
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

// ListUsers returns a list of all users
func ListUsers(limit, offset int) ([]*models.User, error) {
	query := `
        SELECT id, username, email, password_hash, first_name, last_name, avatar_url, about_me, birth_date, created_at, updated_at
        FROM users
        ORDER BY created_at DESC
        LIMIT ? OFFSET ?
    `

	rows, err := db.GlobalDB.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	users := make([]*models.User, 0)
	for rows.Next() {
		var user models.User
		var createdAt, updatedAt string

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

		// Parse timestamps
		user.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		user.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}
