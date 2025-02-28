package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/HASANALI117/social-network/pkg/helpers"
	"github.com/HASANALI117/social-network/pkg/models"
)

// UserHandler handles HTTP requests for users
type UserHandler struct {
	userDB *helpers.UserDB
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userDB *helpers.UserDB) *UserHandler {
	return &UserHandler{userDB: userDB}
}

// Register handles user registration
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
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
	if err := h.userDB.Create(&user); err != nil {
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

// GetUser handles requests to get a user by ID
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
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
	user, err := h.userDB.GetByID(userID)
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

// ListUsers handles requests to list users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
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
	users, err := h.userDB.ListUsers(limit, offset)
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

// UpdateUser handles requests to update a user
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
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
	user, err := h.userDB.GetByID(userID)
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
	if err := h.userDB.Update(user); err != nil {
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

// DeleteUser handles requests to delete a user
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
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
	if err := h.userDB.Delete(userID); err != nil {
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
