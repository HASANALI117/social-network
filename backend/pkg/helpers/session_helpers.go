package helpers

import (
	"errors"
	"net/http"

	"github.com/HASANALI117/social-network/pkg/services"
)

var (
	ErrInvalidSession = errors.New("invalid or expired session")
)

// GetUserFromSession retrieves the user from the session cookie using AuthService
func GetUserFromSession(r *http.Request, authService services.AuthService) (*services.UserResponse, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			return nil, ErrInvalidSession
		}
		return nil, err
	}

	// Use AuthService to get user by session token
	user, err := authService.GetUserBySessionToken(cookie.Value)
	if err != nil {
		return nil, ErrInvalidSession
	}

	return user, nil
}
