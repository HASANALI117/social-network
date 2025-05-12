package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/HASANALI117/social-network/pkg/helpers" // Added for GetUserFromSession
	"github.com/HASANALI117/social-network/pkg/httperr"
	"github.com/HASANALI117/social-network/pkg/models"
	"github.com/HASANALI117/social-network/pkg/repositories" // For ErrUserNotFound comparison
	"github.com/HASANALI117/social-network/pkg/services"
)

// UserHandler handles HTTP requests for users
type UserHandler struct {
	userService     services.UserService
	authService     services.AuthService // Add AuthService
	followerHandler *FollowerHandler     // Added FollowerHandler
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(userService services.UserService, authService services.AuthService, followerHandler *FollowerHandler) *UserHandler {
	return &UserHandler{
		userService:     userService,
		authService:     authService,     // Store AuthService
		followerHandler: followerHandler, // Store FollowerHandler
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

	log.Printf("UserHandler: Method=%s, Path=%s, Parts=%v\n", r.Method, r.URL.Path, parts)

	// Route based on the number of path parts after /api/users/
	if len(parts) == 1 {
		id := parts[0]
		if id == "" { // Path is /api/users/
			switch r.Method {
			case http.MethodPost:
				return h.createUser(w, r)
			case http.MethodGet:
				return h.listUsers(w, r)
			default:
				return httperr.NewMethodNotAllowed(nil, "")
			}
		} else { // Path is /api/users/{id}
			switch r.Method {
			case http.MethodGet:
				return h.getUserByID(w, r, id)
			case http.MethodPut:
				return h.updateUser(w, r, id)
			case http.MethodDelete:
				return h.deleteUser(w, r, id)
			default:
				return httperr.NewMethodNotAllowed(nil, "")
			}
		}
	} else if len(parts) == 2 { // Path is /api/users/{id}/{action}
		userID := parts[0] // Extract userID here
		action := parts[1]

		if h.followerHandler == nil && action != "privacy" { // Only check followerHandler if needed
			log.Println("Error: FollowerHandler not initialized in UserHandler")
			return httperr.NewInternalServerError(nil, "Server configuration error")
		}

		switch action {
		case "follow":
			if r.Method == http.MethodPost {
				h.followerHandler.HandleFollowRequest(w, r)
				return nil // Follower handler writes response
			}
			return httperr.NewMethodNotAllowed(nil, "Method POST required for follow")
		case "unfollow":
			if r.Method == http.MethodDelete {
				h.followerHandler.HandleUnfollow(w, r)
				return nil
			}
			return httperr.NewMethodNotAllowed(nil, "Method DELETE required for unfollow")
		case "accept":
			if r.Method == http.MethodPost {
				h.followerHandler.HandleAcceptRequest(w, r)
				return nil
			}
			return httperr.NewMethodNotAllowed(nil, "Method POST required for accept")
		case "reject":
			if r.Method == http.MethodDelete {
				h.followerHandler.HandleRejectRequest(w, r)
				return nil
			}
			return httperr.NewMethodNotAllowed(nil, "Method DELETE required for reject")
		case "followers":
			if r.Method == http.MethodGet {
				h.followerHandler.HandleListFollowers(w, r)
				return nil
			}
			return httperr.NewMethodNotAllowed(nil, "Method GET required for followers")
		case "following":
			if r.Method == http.MethodGet {
				h.followerHandler.HandleListFollowing(w, r)
				return nil
			}
			return httperr.NewMethodNotAllowed(nil, "Method GET required for following")
		case "privacy":
			if r.Method == http.MethodPut {
				return h.updatePrivacy(w, r, userID)
			}
			return httperr.NewMethodNotAllowed(nil, "Method PUT required for privacy")
		case "cancel-follow-request":
			if r.Method == http.MethodDelete {
				h.followerHandler.HandleCancelFollowRequest(w, r)
				return nil // Follower handler writes response
			}
			return httperr.NewMethodNotAllowed(nil, "Method DELETE required for cancel-follow-request")
		default:
			// Action didn't match any known follower or privacy action
			return httperr.NewNotFound(nil, "Unknown user action: "+action)
		}
		// Code should not reach here because all cases in the switch return
	} else {
		// Any other path structure is not found
		return httperr.NewNotFound(nil, "Invalid user path")
	}
}

// createUser handles POST /api/users/
func (h *UserHandler) createUser(w http.ResponseWriter, r *http.Request) error {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		return httperr.NewBadRequest(err, "Invalid request body")
	}

	// Validation moved to service layer

	createdUserResponse, err := h.userService.Register(&user)
	if err != nil {
		// Check for specific duplicate user error from the repository/service layer
		if errors.Is(err, repositories.ErrUserAlreadyExists) {
			return httperr.NewConflict(err, "User already exists") // Return 409 Conflict
		}
		// Check for generic validation errors (assuming service returns fmt.Errorf for now)
		// TODO: Implement custom validation error types in service for better checking
		if err.Error() == "username is required" || err.Error() == "email is required" || err.Error() == "password is required" || err.Error() == "password must be at least 8 characters" {
			return httperr.NewBadRequest(err, err.Error())
		}
		// Handle other errors as internal server error
		return httperr.NewInternalServerError(err, "Failed to register user")
	}

	// Return created user response DTO from service
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdUserResponse)
	return nil
}

// getUserByID handles GET /api/users/{id} - Now fetches full profile
func (h *UserHandler) getUserByID(w http.ResponseWriter, r *http.Request, profileUserID string) error {
	// Get the ID of the user making the request (viewer), if logged in
	viewerID := ""
	currentUser, err := helpers.GetUserFromSession(r, h.authService)
	if err != nil && !errors.Is(err, helpers.ErrInvalidSession) {
		// Log unexpected errors, but proceed as anonymous viewer for session errors
		log.Printf("Error getting user from session: %v", err)
		// Allow anonymous viewing attempt
	}
	if currentUser != nil {
		viewerID = currentUser.ID
	}

	// Call the service to get the profile, passing viewer and profile IDs
	userProfileResponse, err := h.userService.GetUserProfile(viewerID, profileUserID)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return httperr.NewNotFound(err, "User profile not found")
		}
		if errors.Is(err, services.ErrForbidden) {
			// Return 403 Forbidden if the viewer is not allowed to see the profile
			return httperr.NewForbidden(err, "Access to this profile is restricted")
		}
		// Handle other potential errors
		return httperr.NewInternalServerError(err, "Failed to get user profile")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userProfileResponse) // Encode the full profile response
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

	// Service now returns sanitized UserResponse DTOs
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users":  users, // Encode the list of DTOs directly
		"limit":  limit,
		"offset": offset,
		"count":  len(users), // Note: This is count of returned users, not total count
	})
	return nil
}

// updateUser handles PUT /api/users/{id}
func (h *UserHandler) updateUser(w http.ResponseWriter, r *http.Request, id string) error {
	// TODO: Add authorization check - ensure the logged-in user can update this profile
	/*
	   currentUser, err := helpers.GetUserFromSession(r, h.authService) // Pass authService
	   if err != nil {
	       // Handle error, potentially map ErrInvalidSession to Unauthorized
	       if errors.Is(err, helpers.ErrInvalidSession) {
	           return httperr.NewUnauthorized(err, "Invalid session")
	       }
	       return httperr.NewInternalServerError(err, "Failed to get current user")
	   }
	   if currentUser.ID != id {
	       return httperr.NewForbidden(nil, "Not authorized to update this user") // Use 403 Forbidden
	   }
	*/

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
	json.NewEncoder(w).Encode(updatedUser) // Encode the response DTO directly
	return nil
}

// deleteUser handles DELETE /api/users/{id}
func (h *UserHandler) deleteUser(w http.ResponseWriter, r *http.Request, id string) error {
	// TODO: Add authorization check - ensure the logged-in user can delete this profile (or is admin)
	/*
	   currentUser, err := helpers.GetUserFromSession(r, h.authService) // Pass authService
	   if err != nil {
	       if errors.Is(err, helpers.ErrInvalidSession) {
	           return httperr.NewUnauthorized(err, "Invalid session")
	       }
	       return httperr.NewInternalServerError(err, "Failed to get current user")
	   }
	   // Add admin check or ensure currentUser.ID == id
	   isAdmin := false // Placeholder for actual admin check logic
	   if currentUser.ID != id && !isAdmin {
	        return httperr.NewForbidden(nil, "Not authorized to delete this user") // Use 403 Forbidden
	   }
	*/

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

// sanitizeUser function removed as sanitization is now handled by the service layer returning UserResponse DTOs

// updatePrivacy handles PUT /api/users/{id}/privacy
func (h *UserHandler) updatePrivacy(w http.ResponseWriter, r *http.Request, userID string) error {
	// TODO: Add authorization check - ensure the logged-in user can update this profile
	/*
	   currentUser, err := helpers.GetUserFromSession(r, h.authService)
	   if err != nil {
	       if errors.Is(err, helpers.ErrInvalidSession) {
	           return httperr.NewUnauthorized(err, "Invalid session")
	       }
	       return httperr.NewInternalServerError(err, "Failed to get current user")
	   }
	   if currentUser.ID != userID {
	       return httperr.NewForbidden(nil, "Not authorized to update this user's privacy")
	   }
	*/

	var payload struct {
		IsPrivate *bool `json:"is_private"` // Use pointer to distinguish between false and not provided
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return httperr.NewBadRequest(err, "Invalid request body")
	}

	if payload.IsPrivate == nil {
		return httperr.NewBadRequest(nil, "Missing 'is_private' field")
	}

	err := h.userService.UpdatePrivacy(userID, *payload.IsPrivate)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return httperr.NewNotFound(err, "User not found")
		}
		return httperr.NewInternalServerError(err, "Failed to update user privacy")
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Privacy setting updated successfully",
		"user_id":    userID,
		"is_private": *payload.IsPrivate,
	})
	return nil
}

//
// // OnlineUsers might need separate handling or integration with the new structure
// // if it relies on different mechanisms (like WebSockets) than the standard CRUD operations.
// // Keep helpers import if GetUserFromSession is needed here.
// // OnlineUsers godoc
// // @Summary Get online users
// // @Description Get a list of currently online users
// // @Tags users
// // @Accept json
// // @Produce json
// // @Success 200 {array} map[string]string "List of online users"
// // @Failure 401 {object} httperr.ErrorResponse "Unauthorized"
// // @Router /users/online [get]
// // NOTE: This function is standalone and needs access to AuthService.
// // Consider making it a method of UserHandler or passing AuthService differently.
// // For now, assuming it might be refactored or removed later, leaving the call broken.
// // If needed, it would require AuthService injection similar to other handlers.
// func OnlineUsers(w http.ResponseWriter, r *http.Request) error {
// if r.Method != http.MethodGet {
// return httperr.NewMethodNotAllowed(nil, "")
// }
//
// // This call is now broken as it needs AuthService.
// // currentUser, err := helpers.GetUserFromSession(r, /* Need AuthService instance here */)
// // Placeholder to avoid immediate compile error, but needs proper fix:
// currentUser, err := func() (*services.UserResponse, error) {
// return nil, errors.New("OnlineUsers needs AuthService injection")
// }() // Immediately invoked function expression as placeholder
//
// if err != nil {
// // Map error appropriately
// if errors.Is(err, helpers.ErrInvalidSession) || err.Error() == "OnlineUsers needs AuthService injection" {
// return httperr.NewUnauthorized(err, "Invalid session")
// }
// return httperr.NewInternalServerError(err, "Failed to get current user")
// }
//
// // Get online users from WebSocket hub (assuming WebSocketHub exists and has this method)
// // This part remains dependent on the WebSocket implementation details.
// onlineUsers := WebSocketHub.GetUsersWithStatus()
//
// // Filter out current user from the list
// filteredUsers := make([]map[string]string, 0)
// for _, user := range onlineUsers {
// if user["id"] != currentUser.ID {
// filteredUsers = append(filteredUsers, user)
// }
// }
//
// w.Header().Set("Content-Type", "application/json")
// json.NewEncoder(w).Encode(map[string]interface{}{
// "online_users": filteredUsers,
// "count":        len(filteredUsers),
// })
// return nil
// }
