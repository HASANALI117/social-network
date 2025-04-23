package handlers

import (
"encoding/json"
"errors"
"net/http"

"github.com/HASANALI117/social-network/pkg/helpers"
"github.com/HASANALI117/social-network/pkg/httperr"
"github.com/HASANALI117/social-network/pkg/services" // Import services
)

// GroupMemberHandler handles group membership requests
type GroupMemberHandler struct {
authService services.AuthService
// TODO: Inject GroupService/GroupMemberService when created
}

// NewGroupMemberHandler creates a new GroupMemberHandler
func NewGroupMemberHandler(authService services.AuthService) *GroupMemberHandler {
return &GroupMemberHandler{
authService: authService,
}
}

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
// @Failure 400 {object} httperr.ErrorResponse "Invalid request"
// @Failure 401 {object} httperr.ErrorResponse "Unauthorized"
// @Failure 405 {object} httperr.ErrorResponse "Method not allowed"
// @Failure 409 {object} httperr.ErrorResponse "User already in group"
// @Failure 500 {object} httperr.ErrorResponse "Failed to add member or session error"
// @Router /groups/members/add [post]
func (h *GroupMemberHandler) AddGroupMember(w http.ResponseWriter, r *http.Request) error {
if r.Method != http.MethodPost {
return httperr.NewMethodNotAllowed(nil, "")
}

// Get current user from session using AuthService
currentUser, err := helpers.GetUserFromSession(r, h.authService)
if err != nil {
if errors.Is(err, helpers.ErrInvalidSession) {
return httperr.NewUnauthorized(err, "Invalid session")
}
return httperr.NewInternalServerError(err, "Failed to get current user")
}

groupID := r.URL.Query().Get("id")
	if groupID == "" {
		return httperr.NewBadRequest(nil, "Group ID is required")
	}

	// Check if current user is an admin
	isAdmin, err := helpers.IsGroupAdmin(groupID, currentUser.ID)
	if err != nil {
		return httperr.NewInternalServerError(err, "Failed to check admin status")
	}
	if !isAdmin {
		return httperr.NewUnauthorized(nil, "Only admins can add members")
	}

	// Parse request body
	var req struct {
		UserID string `json:"user_id"`
		Role   string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return httperr.NewBadRequest(err, "Invalid request body")
	}

	if req.UserID == "" {
		return httperr.NewBadRequest(nil, "User ID is required")
	}

	// Add member to group
	if err := helpers.AddGroupMember(groupID, req.UserID, req.Role); err != nil {
		if errors.Is(err, helpers.ErrAlreadyGroupMember) {
			return httperr.NewHTTPError(http.StatusConflict, "User is already a member of this group", err)
		}
		return httperr.NewInternalServerError(err, "Failed to add member")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Member added successfully",
	})
	return nil
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
// @Failure 400 {object} httperr.ErrorResponse "Invalid request"
// @Failure 401 {object} httperr.ErrorResponse "Unauthorized"
// @Failure 404 {object} httperr.ErrorResponse "User not in group"
// @Failure 405 {object} httperr.ErrorResponse "Method not allowed"
// @Failure 500 {object} httperr.ErrorResponse "Failed to remove member or session error"
// @Router /groups/members/remove [post]
func (h *GroupMemberHandler) RemoveGroupMember(w http.ResponseWriter, r *http.Request) error {
if r.Method != http.MethodPost {
return httperr.NewMethodNotAllowed(nil, "")
}

// Get current user from session using AuthService
currentUser, err := helpers.GetUserFromSession(r, h.authService)
if err != nil {
if errors.Is(err, helpers.ErrInvalidSession) {
return httperr.NewUnauthorized(err, "Invalid session")
}
return httperr.NewInternalServerError(err, "Failed to get current user")
}

groupID := r.URL.Query().Get("id")
	if groupID == "" {
		return httperr.NewBadRequest(nil, "Group ID is required")
	}

	// Parse request body
	var req struct {
		UserID string `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return httperr.NewBadRequest(err, "Invalid request body")
	}

	if req.UserID == "" {
		return httperr.NewBadRequest(nil, "User ID is required")
	}

	// Check if current user is an admin or removing self
	isAdmin, err := helpers.IsGroupAdmin(groupID, currentUser.ID)
	if err != nil {
		return httperr.NewInternalServerError(err, "Failed to check admin status")
	}
	if !isAdmin && currentUser.ID != req.UserID {
		return httperr.NewUnauthorized(nil, "Only admins can remove other members")
	}

	// Remove member from group
	if err := helpers.RemoveGroupMember(groupID, req.UserID); err != nil {
		if errors.Is(err, helpers.ErrNotGroupMember) {
			return httperr.NewNotFound(err, "User is not a member of this group")
		}
		return httperr.NewInternalServerError(err, "Failed to remove member")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Member removed successfully",
	})
	return nil
}

// ListGroupMembers godoc
// @Summary List group members
// @Description Get a list of members in a group
// @Tags groups
// @Accept json
// @Produce json
// @Param id query string true "Group ID"
// @Success 200 {object} map[string]interface{} "List of group members"
// @Failure 400 {object} httperr.ErrorResponse "Group ID is required"
// @Failure 401 {object} httperr.ErrorResponse "Unauthorized"
// @Failure 405 {object} httperr.ErrorResponse "Method not allowed"
// @Failure 500 {object} httperr.ErrorResponse "Failed to list members or session error"
// @Router /groups/members [get]
func (h *GroupMemberHandler) ListGroupMembers(w http.ResponseWriter, r *http.Request) error {
if r.Method != http.MethodGet {
return httperr.NewMethodNotAllowed(nil, "")
}

// Get current user from session using AuthService
currentUser, err := helpers.GetUserFromSession(r, h.authService)
if err != nil {
if errors.Is(err, helpers.ErrInvalidSession) {
return httperr.NewUnauthorized(err, "Invalid session")
}
return httperr.NewInternalServerError(err, "Failed to get current user")
}

groupID := r.URL.Query().Get("id")
	if groupID == "" {
		return httperr.NewBadRequest(nil, "Group ID is required")
	}

	// Check if current user is a member
	isMember, err := helpers.IsGroupMember(groupID, currentUser.ID)
	if err != nil {
		return httperr.NewInternalServerError(err, "Failed to check member status")
	}
	if !isMember {
		return httperr.NewUnauthorized(nil, "Only members can view the member list")
	}

	members, err := helpers.ListGroupMembers(groupID)
	if err != nil {
		return httperr.NewInternalServerError(err, "Failed to list group members")
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
	return nil
}
