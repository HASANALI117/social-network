package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/HASANALI117/social-network/pkg/helpers"
)

// SignIn godoc
// @Summary User login
// @Description Authenticate a user and create a session
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body object true "Login credentials"
// @Success 200 {object} map[string]interface{} "Login successful"
// @Failure 400 {string} string "Invalid credentials"
// @Router /auth/signin [post]
func SignIn(w http.ResponseWriter, r *http.Request) {
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
	user, err := helpers.AuthenticateUser(credentials.Identifier, credentials.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sessionToken, err := helpers.CreateSession(user.ID, 24*time.Hour)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
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
}

// SignOut godoc
// @Summary User logout
// @Description End user session
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Logged out successfully"
// @Router /auth/signout [post]
func SignOut(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Delete session from database
	cookie, err := r.Cookie("session_token")
	if err == nil {
		helpers.DeleteSession(cookie.Value)
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
