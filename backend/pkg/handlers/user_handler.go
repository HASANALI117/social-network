package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/HASANALI117/social-network/pkg/helpers"
	"github.com/HASANALI117/social-network/pkg/httperr"
	"github.com/HASANALI117/social-network/pkg/models"
	"golang.org/x/crypto/bcrypt"
)

// Register godoc
// @Summary Register a new user
// @Description Register a new user in the system
// @Tags users
// @Accept json
// @Produce json
// @Param user body models.User true "User registration details"
// @Success 201 {object} map[string]interface{} "User created successfully"
// @Failure 400 {object} httperr.ErrorResponse "Invalid request body"
// @Failure 500 {object} httperr.ErrorResponse "Failed to register user"
// @Router /users/register [post]
func RegisterUser(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return httperr.NewMethodNotAllowed(nil, "")
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		return httperr.NewBadRequest(err, "Invalid request body")
	}

	// Create and Save user to database
	if err := helpers.CreateUser(&user); err != nil {
		return httperr.NewInternalServerError(err, "Failed to register user")
	}

	// Return created user (excluding password)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"avatar_url": user.AvatarURL,
		"about_me":   user.AboutMe,
		"birth_date": user.BirthDate,
		"created_at": user.CreatedAt,
	})
	return nil
}

// GetUser godoc
// @Summary Get user by ID
// @Description Get user details by user ID
// @Tags users
// @Accept json
// @Produce json
// @Param id query string true "User ID"
// @Success 200 {object} map[string]interface{} "User details"
// @Failure 400 {object} httperr.ErrorResponse "User ID is required"
// @Failure 404 {object} httperr.ErrorResponse "User not found"
// @Failure 500 {object} httperr.ErrorResponse "Failed to get user"
// @Router /users/get [get]
func GetUser(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return httperr.NewMethodNotAllowed(nil, "")
	}

	userID := r.URL.Query().Get("id")
	if userID == "" {
		return httperr.NewBadRequest(nil, "User ID is required")
	}

	user, err := helpers.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, helpers.ErrUserNotFound) {
			return httperr.NewNotFound(err, "User not found")
		}
		return httperr.NewInternalServerError(err, "Failed to get user")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"avatar_url": user.AvatarURL,
		"about_me":   user.AboutMe,
		"birth_date": user.BirthDate,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	})
	return nil
}

// ListUsers godoc
// @Summary List users
// @Description Get a paginated list of users
// @Tags users
// @Accept json
// @Produce json
// @Param limit query int false "Number of users to return (default 10)"
// @Param offset query int false "Number of users to skip (default 0)"
// @Success 200 {object} map[string]interface{} "List of users"
// @Failure 500 {object} httperr.ErrorResponse "Failed to list users"
// @Router /users/list [get]
func ListUsers(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return httperr.NewMethodNotAllowed(nil, "")
	}

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // Default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0 // Default offset
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	users, err := helpers.ListUsers(limit, offset)
	if err != nil {
		return httperr.NewInternalServerError(err, "Failed to list users")
	}

	// Sanitize user data (remove password hash)
	result := make([]map[string]interface{}, len(users))
	for i, user := range users {
		result[i] = map[string]interface{}{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"avatar_url": user.AvatarURL,
			"about_me":   user.AboutMe,
			"birth_date": user.BirthDate,
			"created_at": user.CreatedAt,
			"updated_at": user.UpdatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users":  result,
		"limit":  limit,
		"offset": offset,
		"count":  len(users),
	})
	return nil
}

// UpdateUser godoc
// @Summary Update user
// @Description Update user details
// @Tags users
// @Accept json
// @Produce json
// @Param id query string true "User ID"
// @Param user body object true "User update details"
// @Success 200 {object} map[string]interface{} "Updated user details"
// @Failure 400 {object} httperr.ErrorResponse "Invalid request"
// @Failure 401 {object} httperr.ErrorResponse "Unauthorized"
// @Failure 404 {object} httperr.ErrorResponse "User not found"
// @Failure 500 {object} httperr.ErrorResponse "Failed to update user"
// @Router /users/update [put]
func UpdateUser(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPut {
		return httperr.NewMethodNotAllowed(nil, "")
	}

	// Get current user from session for authorization
	currentUser, err := helpers.GetUserFromSession(r)
	if err != nil {
		return httperr.NewUnauthorized(err, "")
	}

	// Get existing user
	user, err := helpers.GetUserByID(currentUser.ID)
	if err != nil {
		if errors.Is(err, helpers.ErrUserNotFound) {
			return httperr.NewNotFound(err, "User not found")
		}
		return httperr.NewInternalServerError(err, "Failed to get user")
	}

	// Parse request body
	var req struct {
		Username  string `json:"username"`
		Email     string `json:"email"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		AvatarURL string `json:"avatar_url"`
		AboutMe   string `json:"about_me"`
		BirthDate string `json:"birth_date"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return httperr.NewBadRequest(err, "Invalid request body")
	}

	// Basic validation
	if req.Username == "" || req.Email == "" {
		return httperr.NewBadRequest(nil, "Username and email are required")
	}

	// Update user data
	user.Username = req.Username
	user.Email = req.Email
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.AvatarURL = req.AvatarURL
	user.AboutMe = req.AboutMe
	user.BirthDate = req.BirthDate

	// Only update password if provided
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return httperr.NewInternalServerError(err, "Failed to hash password")
		}
		user.Password = string(hashedPassword)
	}

	// Save updated user to database
	if err := helpers.UpdateUser(user); err != nil {
		return httperr.NewInternalServerError(err, "Failed to update user")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"avatar_url": user.AvatarURL,
		"about_me":   user.AboutMe,
		"birth_date": user.BirthDate,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	})
	return nil
}

// DeleteUser godoc
// @Summary Delete user
// @Description Delete a user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id query string true "User ID"
// @Success 200 {object} map[string]string "User deleted successfully"
// @Failure 400 {object} httperr.ErrorResponse "User ID is required"
// @Failure 404 {object} httperr.ErrorResponse "User not found"
// @Failure 500 {object} httperr.ErrorResponse "Failed to delete user"
// @Router /users/delete [delete]
func DeleteUser(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodDelete {
		return httperr.NewMethodNotAllowed(nil, "")
	}

	userID := r.URL.Query().Get("id")
	if userID == "" {
		return httperr.NewBadRequest(nil, "User ID is required")
	}

	if err := helpers.DeleteUser(userID); err != nil {
		if errors.Is(err, helpers.ErrUserNotFound) {
			return httperr.NewNotFound(err, "User not found")
		}
		return httperr.NewInternalServerError(err, "Failed to delete user")
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User deleted successfully",
	})
	return nil
}

// OnlineUsers godoc
// @Summary Get online users
// @Description Get a list of currently online users
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {array} map[string]string "List of online users"
// @Failure 401 {object} httperr.ErrorResponse "Unauthorized"
// @Router /users/online [get]
func OnlineUsers(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return httperr.NewMethodNotAllowed(nil, "")
	}

	currentUser, err := helpers.GetUserFromSession(r)
	if err != nil {
		return httperr.NewUnauthorized(err, "")
	}

	// Get online users from WebSocket hub
	onlineUsers := WebSocketHub.GetUsersWithStatus()

	// Filter out current user from the list
	filteredUsers := make([]map[string]string, 0)
	for _, user := range onlineUsers {
		if user["id"] != currentUser.ID {
			filteredUsers = append(filteredUsers, user)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"online_users": filteredUsers,
		"count":        len(filteredUsers),
	})
	return nil
}
