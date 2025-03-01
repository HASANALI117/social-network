package helpers

import (
	"fmt"

	"github.com/HASANALI117/social-network/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

// AuthenticateUser verifies user credentials and returns the user if valid
func (udb *UserDB) AuthenticateUser(identifier string, password string) (*models.User, error) {
	fmt.Println("identifier is: " + identifier)
	fmt.Println("password is: " + password)

	query := `
        SELECT id, username, email, password_hash, first_name, last_name, avatar_url, about_me, birth_date, created_at, updated_at
        FROM users
        WHERE username = ? OR email = ?
    `

	var user models.User
	var hashedPassword string

	fmt.Println(user)
	fmt.Println("hash is: " + hashedPassword)

	err := udb.db.QueryRow(query, identifier, identifier).Scan(
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
