package handlers

import (
"encoding/json"
"errors"
"net/http"
"time"

// "github.com/HASANALI117/social-network/pkg/helpers" // No longer needed directly here
"github.com/HASANALI117/social-network/pkg/httperr"
"github.com/HASANALI117/social-network/pkg/services" // Import services
)

// AuthHandler handles authentication requests
type AuthHandler struct {
authService services.AuthService
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authService services.AuthService) *AuthHandler {
return &AuthHandler{
authService: authService,
}
}

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
// @Failure 401 {object} httperr.ErrorResponse "Invalid credentials"
// @Failure 500 {object} httperr.ErrorResponse "Internal server error"
// @Router /auth/signin [post]
func (h *AuthHandler) SignIn(w http.ResponseWriter, r *http.Request) error {
if r.Method != http.MethodPost {
return httperr.NewMethodNotAllowed(nil, "")
}

var creds services.AuthCredentials // Use AuthCredentials from service
if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
return httperr.NewBadRequest(err, "Invalid request body")
}

// Call AuthService to sign in
session, userResponse, err := h.authService.SignIn(creds)
if err != nil {
if errors.Is(err, services.ErrInvalidCredentials) {
return httperr.NewUnauthorized(err, "Invalid credentials") // Use 401 Unauthorized
}
// Handle other potential errors from service (e.g., DB errors)
return httperr.NewInternalServerError(err, "Failed to sign in")
}

// Set session cookie
http.SetCookie(w, &http.Cookie{
Name:     "session_token",
Value:    session.Token,
Path:     "/",
Expires:  session.ExpiresAt, // Use Expires field
HttpOnly: true,
// Secure: true, // Add Secure flag in production (HTTPS)
// SameSite: http.SameSiteLaxMode, // Consider SameSite attribute
})

// Return sanitized user data from service
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(map[string]interface{}{
"message": "User logged in successfully",
"user":    userResponse, // Directly encode the UserResponse DTO
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
// @Failure 500 {object} httperr.ErrorResponse "Internal server error"
// @Router /auth/signout [post]
func (h *AuthHandler) SignOut(w http.ResponseWriter, r *http.Request) error {
if r.Method != http.MethodPost {
return httperr.NewMethodNotAllowed(nil, "")
}

// Get token from cookie
cookie, err := r.Cookie("session_token")
if err == nil && cookie.Value != "" {
// Call AuthService to sign out (delete session)
// Ignore errors here as we want to clear the cookie regardless
_ = h.authService.SignOut(cookie.Value)
}

// Clear the session cookie
http.SetCookie(w, &http.Cookie{
Name:     "session_token",
Value:    "",
Path:     "/",
Expires:  time.Unix(0, 0), // Set expiry to the past
HttpOnly: true,
// Secure: true, // Add Secure flag in production (HTTPS)
// SameSite: http.SameSiteLaxMode, // Consider SameSite attribute
MaxAge:   -1, // Explicitly tell browser to delete cookie
})

// Return success response
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(map[string]interface{}{
"message": "Logged out successfully",
})
return nil
}
