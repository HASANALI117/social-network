package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/HASANALI117/social-network/pkg/helpers"
	"github.com/HASANALI117/social-network/pkg/models"
)

// Register godoc
// @Summary Register a new user
// @Description Register a new user in the system
// @Tags users
// @Accept json
// @Produce json
// @Param user body models.User true "User registration details"
// @Success 201 {object} map[string]interface{} "User created successfully"
// @Failure 400 {string} string "Invalid request body"
// @Failure 500 {string} string "Failed to register user"
// @Router /users/register [post]
func RegisterUser(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Basic validation
	// if req.Username == "" || req.Email == "" || req.Password == "" {
	// 	http.Error(w, "Username, email, and password are required", http.StatusBadRequest)
	// 	return
	// }

	// Create and Save user to database
	if err := helpers.CreateUser(&user); err != nil {
		http.Error(w, "Failed to register user", http.StatusInternalServerError)
		return
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
}

// GetUser godoc
// @Summary Get user by ID
// @Description Get user details by user ID
// @Tags users
// @Accept json
// @Produce json
// @Param id query string true "User ID"
// @Success 200 {object} map[string]interface{} "User details"
// @Failure 400 {string} string "User ID is required"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Failed to get user"
// @Router /users/get [get]
func GetUser(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from URL query parameter
	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Get user from database
	user, err := helpers.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, helpers.ErrUserNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get user", http.StatusInternalServerError)
		}
		return
	}

	// Return user data
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
// @Failure 500 {string} string "Failed to list users"
// @Router /users/list [get]
func ListUsers(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
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

	// Get users from database
	users, err := helpers.ListUsers(limit, offset)
	if err != nil {
		http.Error(w, "Failed to list users", http.StatusInternalServerError)
		return
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

	// Return user list
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users":  result,
		"limit":  limit,
		"offset": offset,
		"count":  len(users),
	})
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
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Failed to update user"
// @Router /users/update [put]
func UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Only allow PUT requests
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from URL query parameter
	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Get existing user
	user, err := helpers.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, helpers.ErrUserNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get user", http.StatusInternalServerError)
		}
		return
	}

	// Parse request body
	var req struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		AvatarURL string `json:"avatar_url"`
		AboutMe   string `json:"about_me"`
		BirthDate string `json:"birth_date"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update user data
	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.AvatarURL = req.AvatarURL
	user.AboutMe = req.FirstName
	user.BirthDate = req.BirthDate

	// Save updated user to database
	if err := helpers.UpdateUser(user); err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		return
	}

	// Return updated user
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
	})
}

// DeleteUser godoc
// @Summary Delete user
// @Description Delete a user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id query string true "User ID"
// @Success 200 {object} map[string]string "User deleted successfully"
// @Failure 400 {string} string "User ID is required"
// @Failure 404 {string} string "User not found"
// @Failure 500 {string} string "Failed to delete user"
// @Router /users/delete [delete]
func DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Only allow DELETE requests
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from URL query parameter
	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Delete user from database
	if err := helpers.DeleteUser(userID); err != nil {
		if errors.Is(err, helpers.ErrUserNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		}
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User deleted successfully",
	})
}

// OnlineUsers godoc
// @Summary Get online users
// @Description Get a list of currently online users
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {array} map[string]string "List of online users"
// @Failure 401 {string} string "Unauthorized"
// @Router /users/online [get]
func OnlineUsers(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from session
	currentUser, err := helpers.GetUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
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

	// Return the list
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"online_users": filteredUsers,
		"count":        len(filteredUsers),
	})
}
