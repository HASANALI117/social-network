package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/HASANALI117/social-network/pkg/apperrors"
	"github.com/HASANALI117/social-network/pkg/helpers"
)

// TODO: Replace placeholder object with actual request/response models later
type signInRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type signInResponse struct {
	Message string      `json:"message"`
	User    interface{} `json:"user"` // Use a proper User model DTO later
}

type signOutResponse struct {
	Message string `json:"message"`
}

// SignIn godoc
// @Summary User login
// @Description Authenticate a user and create a session
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body object true "Login credentials"
// @Success 200 {object} map[string]interface{} "Login successful"
// @Failure 400 {object} map[string]string "Invalid credentials or request body"
// @Failure 401 {object} map[string]string "Incorrect password"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Failed to create session or internal error"
// @Router /auth/signin [post]
func SignIn(w http.ResponseWriter, r *http.Request) error {
	// Method check (can be handled by router later, but good for defense)
	if r.Method != http.MethodPost {
		return apperrors.ErrMethodNotAllowed("", nil)
	}

	var req signInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apperrors.ErrBadRequest("Invalid request body", err)
	}
	defer r.Body.Close()

	// Validate input (basic example)
	if req.Identifier == "" || req.Password == "" {
		return apperrors.ErrBadRequest("Identifier and password are required", nil)
	}

	// Authenticate user
	// TODO: Replace helpers.AuthenticateUser when service layer is added
	user, err := helpers.AuthenticateUser(req.Identifier, req.Password)
	if err != nil {
		// Distinguish between "not found" and "wrong password" if possible from AuthenticateUser
		// For now, assume AuthenticateUser returns specific errors or a generic one
		// Example check (adjust based on actual AuthenticateUser error types):
		if err.Error() == "sql: no rows in result set" { // Example check
			return apperrors.ErrNotFound("User not found", err)
		}
		if err.Error() == "incorrect password" { // Example check
			return apperrors.ErrUnauthorized("Incorrect password", err)
		}
		// Generic fallback for other auth errors
		return apperrors.ErrBadRequest("Authentication failed", err)
	}

	// TODO: Replace helpers.CreateSession when service layer is added
	sessionToken, err := helpers.CreateSession(user.ID, 24*time.Hour)
	if err != nil {
		return apperrors.ErrInternalServer("Failed to create session", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Recommended for production
		SameSite: http.SameSiteLaxMode, // Recommended
		MaxAge:   86400, // 1 day
	})

	// Prepare response DTO (Data Transfer Object)
	response := signInResponse{
		Message: "User logged in successfully",
		User: map[string]interface{}{ // Replace with a proper User DTO later
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"avatar_url": user.AvatarURL,
			"about_me":   user.AboutMe,
			"birth_date": user.BirthDate,
		},
	}

	return helpers.RespondJSON(w, http.StatusOK, response)
}

// SignOut godoc
// @Summary User logout
// @Description End user session
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} signOutResponse "Logged out successfully"
// @Failure 405 {object} map[string]string "Method not allowed"
// @Failure 500 {object} map[string]string "Internal server error during session deletion"
// @Router /auth/signout [post]
func SignOut(w http.ResponseWriter, r *http.Request) error {
	// Method check
	if r.Method != http.MethodPost {
		return apperrors.ErrMethodNotAllowed("", nil)
	}

	// Delete session from database
	cookie, err := r.Cookie("session_token")
	if err == nil && cookie.Value != "" {
		// TODO: Replace helpers.DeleteSession when service layer is added
		// The DeleteSession function should ideally return an error
		deleteErr := helpers.DeleteSession(cookie.Value)
		if deleteErr != nil {
			// Log the error but proceed to clear the cookie anyway
			log.Printf("WARN: Failed to delete session from DB: %v", deleteErr)
			// Optionally return an internal server error if session deletion failure is critical
			// return apperrors.ErrInternalServer("Failed to clear session", deleteErr)
		}
	}
	// If cookie doesn't exist or is empty, no server-side action needed, just clear client cookie.

	// Clear the cookie on the client side
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Match settings used in SignIn
		SameSite: http.SameSiteLaxMode, // Match settings used in SignIn
		MaxAge:   -1, // Instructs browser to delete cookie immediately
		Expires:  time.Unix(0, 0), // Explicitly set expiry in the past
	})

	response := signOutResponse{
		Message: "Logged out successfully",
	}

	return helpers.RespondJSON(w, http.StatusOK, response)
}
