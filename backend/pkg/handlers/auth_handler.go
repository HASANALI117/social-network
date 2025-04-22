package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/HASANALI117/social-network/pkg/helpers"
	"github.com/HASANALI117/social-network/pkg/httperr"
)

// SignIn godoc
// @Summary User login
// @Description Authenticate a user and create a session
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body object true "Login credentials"
// @Success 200 {object} map[string]interface{} "Login successful"
// @Failure 400 {object} httperr.ErrorResponse "Invalid credentials"
// @Failure 405 {object} httperr.ErrorResponse "Method not allowed"
// @Failure 500 {object} httperr.ErrorResponse "Failed to create session"
// @Router /auth/signin [post]
func SignIn(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return httperr.NewMethodNotAllowed(nil, "")
	}

	var credentials struct {
		Identifier string `json:"identifier"`
		Password   string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		return httperr.NewBadRequest(err, "Invalid request body")
	}

	// Authenticate user
	user, err := helpers.AuthenticateUser(credentials.Identifier, credentials.Password)
	if err != nil {
		if errors.Is(err, helpers.ErrUserNotFound) {
			return httperr.NewBadRequest(nil, "Email Not Found")
		}
		return httperr.NewBadRequest(err, "Invalid credentials")
	}

	sessionToken, err := helpers.CreateSession(user.ID, 24*time.Hour)
	if err != nil {
		return httperr.NewInternalServerError(err, "Failed to create session")
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400, // 1 day
	})

	// Return created user (excluding password)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User logged in successfully",
		"user": map[string]interface{}{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"avatar_url": user.AvatarURL,
			"about_me":   user.AboutMe,
			"birth_date": user.BirthDate,
		},
	})
	return nil
}

// SignOut godoc
// @Summary User logout
// @Description End user session
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Logged out successfully"
// @Failure 405 {object} httperr.ErrorResponse "Method not allowed"
// @Router /auth/signout [post]
func SignOut(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return httperr.NewMethodNotAllowed(nil, "")
	}

	// Delete session from database
	cookie, err := r.Cookie("session_token")
	if err == nil {
		// Ignore errors from DeleteSession as we want to proceed with cookie deletion anyway
		helpers.DeleteSession(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1, // Delete cookie
	})

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Logged out successfully",
	})
	return nil
}
