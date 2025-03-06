package helpers

import (
	"fmt"

	"github.com/HASANALI117/social-network/pkg/db"
	"github.com/HASANALI117/social-network/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

// AuthenticateUser verifies user credentials and returns the user if valid
func AuthenticateUser(identifier string, password string) (*models.User, error) {
	query := `
        SELECT id, username, email, password_hash, first_name, last_name, avatar_url, about_me, birth_date, created_at, updated_at
        FROM users
        WHERE username = ? OR email = ?
    `

	var user models.User
	var hashedPassword string

	err := db.GlobalDB.QueryRow(query, identifier, identifier).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&hashedPassword,
		&user.FirstName,
		&user.LastName,
		&user.AvatarURL,
		&user.AboutMe,
		&user.BirthDate,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return nil, fmt.Errorf("incorrect password")
	}

	return &user, nil
}
