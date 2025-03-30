package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/HASANALI117/social-network/pkg/helpers"
)

// AddGroupMember godoc
// @Summary Add a member to a group
// @Description Add a user to a group
// @Tags groups
// @Accept json
// @Produce json
// @Param group_id query string true "Group ID"
// @Param user_id body string true "User ID to add"
// @Param role body string false "Role (default: member)"
// @Success 200 {object} map[string]string "Member added successfully"
// @Failure 400 {string} string "Invalid request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 409 {string} string "User already in group"
// @Failure 500 {string} string "Failed to add member"
// @Router /groups/members/add [post]
func AddGroupMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from session
	currentUser, err := helpers.GetUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupID := r.URL.Query().Get("id")
	if groupID == "" {
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	// Check if current user is an admin
	isAdmin, err := helpers.IsGroupAdmin(groupID, currentUser.ID)
	if err != nil || !isAdmin {
		http.Error(w, "Unauthorized - only admins can add members", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Add member to group
	if err := helpers.AddGroupMember(groupID, req.UserID, req.Role); err != nil {
		if errors.Is(err, helpers.ErrAlreadyGroupMember) {
			http.Error(w, "User is already a member of this group", http.StatusConflict)
		} else {
			http.Error(w, "Failed to add member", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Member added successfully",
	})
}

// RemoveGroupMember godoc
// @Summary Remove a member from a group
// @Description Remove a user from a group
// @Tags groups
// @Accept json
// @Produce json
// @Param group_id query string true "Group ID"
// @Param user_id body string true "User ID to remove"
// @Success 200 {object} map[string]string "Member removed successfully"
// @Failure 400 {string} string "Invalid request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "User not in group"
// @Failure 500 {string} string "Failed to remove member"
// @Router /groups/members/remove [post]
func RemoveGroupMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get current user from session
	currentUser, err := helpers.GetUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	groupID := r.URL.Query().Get("id")
	if groupID == "" {
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req struct {
		UserID string `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Check if current user is an admin or removing self
	isAdmin, err := helpers.IsGroupAdmin(groupID, currentUser.ID)
	if err != nil || (!isAdmin && currentUser.ID != req.UserID) {
		http.Error(w, "Unauthorized - only admins can remove other members", http.StatusUnauthorized)
		return
	}

	// Remove member from group
	if err := helpers.RemoveGroupMember(groupID, req.UserID); err != nil {
		if errors.Is(err, helpers.ErrNotGroupMember) {
			http.Error(w, "User is not a member of this group", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to remove member", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Member removed successfully",
	})
}

// ListGroupMembers godoc
// @Summary List group members
// @Description Get a list of members in a group
// @Tags groups
// @Accept json
// @Produce json
// @Param id query string true "Group ID"
// @Success 200 {object} map[string]interface{} "List of group members"
// @Failure 400 {string} string "Group ID is required"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Failed to list members"
// @Router /groups/members [get]
func ListGroupMembers(w http.ResponseWriter, r *http.Request) {
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

	groupID := r.URL.Query().Get("id")
	if groupID == "" {
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	// Check if current user is a member
	isMember, err := helpers.IsGroupMember(groupID, currentUser.ID)
	if err != nil || !isMember {
		http.Error(w, "Unauthorized - only members can view the member list", http.StatusUnauthorized)
		return
	}

	members, err := helpers.ListGroupMembers(groupID)
	if err != nil {
		http.Error(w, "Failed to list group members", http.StatusInternalServerError)
		return
	}

	// Sanitize user data (remove password hash)
	result := make([]map[string]interface{}, len(members))
	for i, user := range members {
		result[i] = map[string]interface{}{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"avatar_url": user.AvatarURL,
			"about_me":   user.AboutMe,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"members": result,
		"count":   len(members),
	})
}
