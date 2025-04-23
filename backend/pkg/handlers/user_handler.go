package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/HASANALI117/social-network/pkg/helpers" // Keep for GetUserFromSession if needed for auth/online users
	"github.com/HASANALI117/social-network/pkg/httperr"
	"github.com/HASANALI117/social-network/pkg/models"
	"github.com/HASANALI117/social-network/pkg/repositories" // For ErrUserNotFound comparison
	"github.com/HASANALI117/social-network/pkg/services"
)

// UserHandler handles HTTP requests for users
type UserHandler struct {
	userService services.UserService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// ServeHTTP routes the request to the appropriate handler method based on method and path
func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	// Trim prefix and split path
	// Example: /api/users/123 -> ["", "123"]
	// Example: /api/users/ -> [""]
	// Example: /api/users/online -> ["online"]
	path := strings.TrimPrefix(r.URL.Path, "/api/users")
	path = strings.TrimPrefix(path, "/") // Remove leading slash if present
	parts := strings.Split(path, "/")
	id := ""
	// action := "" // Removed unused variable

	if len(parts) == 1 && parts[0] != "" {
		// Could be an ID or a specific action like "online"
		// Check if it looks like a UUID or numeric ID, otherwise treat as action
		// For simplicity now, assume it's an ID if not empty. Refine if needed.
		// TODO: Add better routing/muxing if actions like 'online' need to coexist cleanly
		// with ID-based routes under /api/users/
		id = parts[0]
	} else if len(parts) > 1 {
		// Handle potential sub-routes if any, e.g., /api/users/123/posts
		// For now, assume only /api/users/ and /api/users/{id}
		return httperr.NewNotFound(nil, "Invalid user path")
	}

	switch r.Method {
	case http.MethodPost:
		// POST /api/users/ -> Create User
		if id != "" {
			return httperr.NewMethodNotAllowed(nil, "Cannot POST to a specific user ID")
		}
		return h.createUser(w, r)
	case http.MethodGet:
		// GET /api/users/{id} -> Get User by ID
		if id != "" {
			return h.getUserByID(w, r, id)
		}
		// GET /api/users/ -> List Users (with query params)
		return h.listUsers(w, r)
	case http.MethodPut:
		// PUT /api/users/{id} -> Update User
		if id == "" {
			return httperr.NewBadRequest(nil, "User ID is required for update")
		}
		return h.updateUser(w, r, id)
	case http.MethodDelete:
		// DELETE /api/users/{id} -> Delete User
		if id == "" {
			return httperr.NewBadRequest(nil, "User ID is required for delete")
		}
		return h.deleteUser(w, r, id)
	default:
		return httperr.NewMethodNotAllowed(nil, "")
	}
}

// createUser handles POST /api/users/
func (h *UserHandler) createUser(w http.ResponseWriter, r *http.Request) error {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		return httperr.NewBadRequest(err, "Invalid request body")
	}

	// Basic validation (can be expanded in service layer if needed)
	if user.Email == "" || user.Password == "" || user.Username == "" {
		return httperr.NewBadRequest(nil, "Username, email, and password are required")
	}

	if err := h.userService.Register(&user); err != nil {
		// Check for specific duplicate user error from the repository/service layer
		if errors.Is(err, repositories.ErrUserAlreadyExists) {
			// Could check the original error string if we need to differentiate between email/username
			// For now, assume email based on the last error message.
			return httperr.NewConflict(err, "user already exists") // Return 409 Conflict
		}
		// Handle other errors as internal server error
		return httperr.NewInternalServerError(err, "Failed to register user")
	}

	// Return created user (excluding password)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(sanitizeUser(&user)) // Use sanitize helper
	return nil
}

// getUserByID handles GET /api/users/{id}
func (h *UserHandler) getUserByID(w http.ResponseWriter, r *http.Request, id string) error {
	user, err := h.userService.GetByID(id)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return httperr.NewNotFound(err, "User not found")
		}
		return httperr.NewInternalServerError(err, "Failed to get user")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sanitizeUser(user))
	return nil
}

// listUsers handles GET /api/users/
func (h *UserHandler) listUsers(w http.ResponseWriter, r *http.Request) error {
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

	users, err := h.userService.List(limit, offset)
	if err != nil {
		return httperr.NewInternalServerError(err, "Failed to list users")
	}

	// Sanitize user data
	sanitizedUsers := make([]map[string]interface{}, len(users))
	for i, user := range users {
		sanitizedUsers[i] = sanitizeUser(user)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users":  sanitizedUsers,
		"limit":  limit,
		"offset": offset,
		"count":  len(users), // Note: This is count of returned users, not total count
	})
	return nil
}

// updateUser handles PUT /api/users/{id}
func (h *UserHandler) updateUser(w http.ResponseWriter, r *http.Request, id string) error {
	// TODO: Add authorization check - ensure the logged-in user can update this profile
	// currentUser, err := helpers.GetUserFromSession(r)
	// if err != nil || currentUser.ID != id {
	//     return httperr.NewUnauthorized(err, "Not authorized to update this user")
	// }

	var updateData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		return httperr.NewBadRequest(err, "Invalid request body")
	}

	// Remove potentially harmful fields or fields not allowed to be updated directly
	delete(updateData, "id")
	delete(updateData, "created_at")
	delete(updateData, "updated_at")

	updatedUser, err := h.userService.Update(id, updateData)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return httperr.NewNotFound(err, "User not found")
		}
		// Handle other potential errors like validation errors if service adds them
		return httperr.NewInternalServerError(err, "Failed to update user")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sanitizeUser(updatedUser))
	return nil
}

// deleteUser handles DELETE /api/users/{id}
func (h *UserHandler) deleteUser(w http.ResponseWriter, r *http.Request, id string) error {
	// TODO: Add authorization check - ensure the logged-in user can delete this profile (or is admin)

	if err := h.userService.Delete(id); err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
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

// sanitizeUser removes sensitive fields like password hash before sending response
func sanitizeUser(user *models.User) map[string]interface{} {
	if user == nil {
		return nil
	}
	return map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email, // Consider if email should always be public
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"avatar_url": user.AvatarURL,
		"about_me":   user.AboutMe,
		"birth_date": user.BirthDate,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}
}

// OnlineUsers might need separate handling or integration with the new structure
// if it relies on different mechanisms (like WebSockets) than the standard CRUD operations.
// Keep helpers import if GetUserFromSession is needed here.
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

	currentUser, err := helpers.GetUserFromSession(r) // Assumes session/auth helper exists
	if err != nil {
		return httperr.NewUnauthorized(err, "")
	}

	// Get online users from WebSocket hub (assuming WebSocketHub exists and has this method)
	// This part remains dependent on the WebSocket implementation details.
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
