package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/HASANALI117/social-network/pkg/apperrors"
	"github.com/HASANALI117/social-network/pkg/helpers"
	"github.com/HASANALI117/social-network/pkg/models"
)

// --- DTOs ---

// AddGroupMemberRequest defines the expected body for adding a member to a group.
type AddGroupMemberRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"` // Optional, consider enum
}

// RemoveGroupMemberRequest defines the expected body for removing a member from a group.
type RemoveGroupMemberRequest struct {
	UserID string `json:"user_id"`
}

// GroupMemberResponse defines the success message for adding/removing a group member.
type GroupMemberResponse struct {
	Message string `json:"message"`
}

// ListGroupMembersResponse defines the structure for listing group members.
type ListGroupMembersResponse struct {
	Members []models.User `json:"members"` // Or a DTO if you want to sanitize
	Count   int           `json:"count"`
}

// AddGroupMember godoc
// @Summary Add a member to a group
// @Description Add a user to a group. Requires authentication and admin privileges.
// @Tags groups
// @Accept json
// @Produce json
// @Param id query string true "Group ID"
// @Param request body AddGroupMemberRequest true "User ID to add and role"
// @Success 200 {object} GroupMemberResponse "Member added successfully"
// @Failure 400 {object} map[string]string "Invalid request body or missing User ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - only admins can add members"
// @Failure 409 {object} map[string]string "User already in group"
// @Failure 500 {object} map[string]string "Failed to add member"
// @Router /groups/members/add [post]
func AddGroupMember(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return apperrors.ErrMethodNotAllowed("", nil)
	}

	// Get current user from session
	// TODO: Replace helpers.GetUserFromSession when auth middleware/service is added
	currentUser, err := helpers.GetUserFromSession(r)
	if err != nil {
		return apperrors.ErrUnauthorized("Unauthorized", err)
	}

	groupID := r.URL.Query().Get("id")
	if groupID == "" {
		return apperrors.ErrBadRequest("Group ID is required", nil)
	}

	// Check if current user is an admin
	isAdmin, err := helpers.IsGroupAdmin(groupID, currentUser.ID)
	if err != nil {
		return apperrors.ErrInternalServer("Failed to check admin status", err)
	}
	if !isAdmin {
		return apperrors.ErrForbidden("Unauthorized - only admins can add members", nil)
	}

	// Parse request body
	var req AddGroupMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apperrors.ErrBadRequest("Invalid request body", err)
	}

	if req.UserID == "" {
		return apperrors.ErrBadRequest("User ID is required", nil)
	}

	// Add member to group
	// TODO: Replace helpers.AddGroupMember when service layer is added
	if err := helpers.AddGroupMember(groupID, req.UserID, req.Role); err != nil {
		if errors.Is(err, helpers.ErrAlreadyGroupMember) {
			return apperrors.ErrConflict("User is already a member of this group", err)
		}
		log.Printf("Failed to add member %s to group %s: %v", req.UserID, groupID, err)
		return apperrors.ErrInternalServer("Failed to add member", err)
	}

	response := GroupMemberResponse{Message: "Member added successfully"}
	return helpers.RespondJSON(w, http.StatusOK, response)
}

// RemoveGroupMember godoc
// @Summary Remove a member from a group
// @Description Remove a user from a group. Requires authentication and admin privileges or self-removal.
// @Tags groups
// @Accept json
// @Produce json
// @Param id query string true "Group ID"
// @Param request body RemoveGroupMemberRequest true "User ID to remove"
// @Success 200 {object} GroupMemberResponse "Member removed successfully"
// @Failure 400 {object} map[string]string "Invalid request body or missing User ID"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - only admins can remove other members"
// @Failure 404 {object} map[string]string "User not in group"
// @Failure 500 {object} map[string]string "Failed to remove member"
// @Router /groups/members/remove [post]
func RemoveGroupMember(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return apperrors.ErrMethodNotAllowed("", nil)
	}

	// Get current user from session
	// TODO: Replace helpers.GetUserFromSession when auth middleware/service is added
	currentUser, err := helpers.GetUserFromSession(r)
	if err != nil {
		return apperrors.ErrUnauthorized("Unauthorized", err)
	}

	groupID := r.URL.Query().Get("id")
	if groupID == "" {
		return apperrors.ErrBadRequest("Group ID is required", nil)
	}

	// Parse request body
	var req RemoveGroupMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apperrors.ErrBadRequest("Invalid request body", err)
	}

	if req.UserID == "" {
		return apperrors.ErrBadRequest("User ID is required", nil)
	}

	// Check if current user is an admin or removing self
	isAdmin, err := helpers.IsGroupAdmin(groupID, currentUser.ID)
	if err != nil {
		return apperrors.ErrInternalServer("Failed to check admin status", err)
	}
	if !isAdmin && currentUser.ID != req.UserID {
		return apperrors.ErrForbidden("Unauthorized - only admins can remove other members", nil)
	}

	// Remove member from group
	// TODO: Replace helpers.RemoveGroupMember when service layer is added
	if err := helpers.RemoveGroupMember(groupID, req.UserID); err != nil {
		if errors.Is(err, helpers.ErrNotGroupMember) {
			return apperrors.ErrNotFound("User is not a member of this group", err)
		}
		log.Printf("Failed to remove member %s from group %s: %v", req.UserID, groupID, err)
		return apperrors.ErrInternalServer("Failed to remove member", err)
	}

	response := GroupMemberResponse{Message: "Member removed successfully"}
	return helpers.RespondJSON(w, http.StatusOK, response)
}

// ListGroupMembers godoc
// @Summary List group members
// @Description Get a list of members in a group. Requires authentication and membership in the group.
// @Tags groups
// @Accept json
// @Produce json
// @Param id query string true "Group ID"
// @Success 200 {object} ListGroupMembersResponse "List of group members"
// @Failure 400 {object} map[string]string "Group ID is required"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - only members can view the member list"
// @Failure 405 {object} map[string]string "Method not allowed"
// @Failure 500 {object} map[string]string "Failed to list members"
// @Router /groups/members [get]
func ListGroupMembers(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return apperrors.ErrMethodNotAllowed("", nil)
	}

	// Get current user from session
	currentUser, err := helpers.GetUserFromSession(r)
	if err != nil {
		return apperrors.ErrUnauthorized("Unauthorized", err)
	}

	groupID := r.URL.Query().Get("id")
	if groupID == "" {
		return apperrors.ErrBadRequest("Group ID is required", nil)
	}

	// Check if current user is a member
	isMember, err := helpers.IsGroupMember(groupID, currentUser.ID)
	if err != nil {
		return apperrors.ErrInternalServer("Failed to check group membership", err)
	}
	if !isMember {
		return apperrors.ErrForbidden("Unauthorized - only members can view the member list", nil)
	}

	members, err := helpers.ListGroupMembers(groupID)
	if err != nil {
		log.Printf("Failed to list members for group %s: %v", groupID, err)
		return apperrors.ErrInternalServer("Failed to list group members", err)
	}

	// Return member list
	response := ListGroupMembersResponse{
		Members: members,
		Count:   len(members),
	}
	return helpers.RespondJSON(w, http.StatusOK, response)
}
