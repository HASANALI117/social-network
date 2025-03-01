package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/HASANALI117/social-network/pkg/helpers"
	"github.com/google/uuid"
)

// AuthHandler handles HTTP requests for users
type AuthHandler struct {
	userDB *helpers.UserDB
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(userDB *helpers.UserDB) *AuthHandler {
	return &AuthHandler{userDB: userDB}
}

// SignIn handles user authentication
func (h *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var credentials struct {
		Identifier string `json:"identifier"`
		Password   string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Authenticate user
	user, err := h.userDB.AuthenticateUser(credentials.Identifier, credentials.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sessionToken := uuid.New().String()

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
}

// Signout handles user logout
func (h *AuthHandler) SignOut(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1, // Delete cookie
	})

	// Return created user (excluding password)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Logged out successfully",
	})
}
