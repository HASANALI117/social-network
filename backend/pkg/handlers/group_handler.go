package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/HASANALI117/social-network/pkg/apperrors"
	"github.com/HASANALI117/social-network/pkg/helpers"
	"github.com/HASANALI117/social-network/pkg/models"
)

// --- DTOs ---

// CreateGroupRequest defines the expected body for creating a group.
type CreateGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	AvatarURL   string `json:"avatar_url"`
}

// GroupResponse defines the group data returned by the API.
type GroupResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatorID   string    `json:"creator_id"`
	AvatarURL   string    `json:"avatar_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListGroupsResponse defines the structure for listing groups.
type ListGroupsResponse struct {
	Groups []GroupResponse `json:"groups"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
	Count  int             `json:"count"`
}

// UpdateGroupRequest defines the expected body for updating a group.
type UpdateGroupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	AvatarURL   string `json:"avatar_url"`
}

// DeleteGroupResponse defines the success message for group deletion.
type DeleteGroupResponse struct {
	Message string `json:"message"`
}

// --- Helper Function ---

// mapModelGroupToGroupResponse converts a models.Group to a GroupResponse DTO.
func mapModelGroupToGroupResponse(group *models.Group) GroupResponse {
	return GroupResponse{
		ID:          group.ID,
		Name:        group.Name,
		Description: group.Description,
		CreatorID:   group.CreatorID,
		AvatarURL:   group.AvatarURL,
		CreatedAt:   group.CreatedAt,
		UpdatedAt:   group.UpdatedAt,
	}
}

// CreateGroup godoc
// @Summary Create a new group
// @Description Create a new group chat. Requires authentication.
// @Tags groups
// @Accept json
// @Produce json
// @Param group body CreateGroupRequest true "Group creation details"
// @Success 201 {object} GroupResponse "Group created successfully"
// @Failure 400 {object} map[string]string "Invalid request body or missing group name"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 405 {object} map[string]string "Method not allowed"
// @Failure 500 {object} map[string]string "Failed to create group"
// @Router /groups/create [post]
func CreateGroup(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return apperrors.ErrMethodNotAllowed("", nil)
	}

	// Get current user from session
	// TODO: Replace helpers.GetUserFromSession when auth middleware/service is added
	currentUser, err := helpers.GetUserFromSession(r)
	if err != nil {
		return apperrors.ErrUnauthorized("Unauthorized", err)
	}

	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apperrors.ErrBadRequest("Invalid request body", err)
	}

	// Basic validation
	if req.Name == "" {
		return apperrors.ErrBadRequest("Group name is required", nil)
	}

	// Map request DTO to internal model
	group := models.Group{
		Name:        req.Name,
		Description: req.Description,
		AvatarURL:   req.AvatarURL,
		CreatorID:   currentUser.ID, // Set creator ID to current user
	}

	// Create group in database
	// TODO: Replace helpers.CreateGroup when service layer is added
	if err := helpers.CreateGroup(&group); err != nil {
		log.Printf("Failed to create group: %v", err)
		return apperrors.ErrInternalServer("Failed to create group", err)
	}

	// Return created group DTO
	response := mapModelGroupToGroupResponse(&group)
	return helpers.RespondJSON(w, http.StatusCreated, response)
}

// GetGroup godoc
// @Summary Get group by ID
// @Description Get group details by group ID
// @Tags groups
// @Accept json
// @Produce json
// @Param id query string true "Group ID"
// @Success 200 {object} GroupResponse "Group details"
// @Failure 400 {object} map[string]string "Group ID is required"
// @Failure 404 {object} map[string]string "Group not found"
// @Failure 405 {object} map[string]string "Method not allowed"
// @Failure 500 {object} map[string]string "Failed to get group"
// @Router /groups/get [get]
func GetGroup(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return apperrors.ErrMethodNotAllowed("", nil)
	}

	groupID := r.URL.Query().Get("id")
	if groupID == "" {
		return apperrors.ErrBadRequest("Group ID is required", nil)
	}

	group, err := helpers.GetGroupByID(groupID)
	if err != nil {
		if errors.Is(err, helpers.ErrGroupNotFound) {
			return apperrors.ErrNotFound("Group not found", err)
		}
		return apperrors.ErrInternalServer("Failed to get group", err)
	}

	response := mapModelGroupToGroupResponse(group)
	return helpers.RespondJSON(w, http.StatusOK, response)
}

// ListGroups godoc
// @Summary List groups
// @Description Get a paginated list of groups
// @Tags groups
// @Accept json
// @Produce json
// @Param limit query int false "Number of groups to return (default 10)" minimum(1) maximum(100)
// @Param offset query int false "Number of groups to skip (default 0)" minimum(0)
// @Success 200 {object} ListGroupsResponse "List of groups"
// @Failure 400 {object} map[string]string "Invalid limit or offset parameter"
// @Failure 405 {object} map[string]string "Method not allowed"
// @Failure 500 {object} map[string]string "Failed to list groups"
// @Router /groups/list [get]
func ListGroups(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return apperrors.ErrMethodNotAllowed("", nil)
	}

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10 // Default limit
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit <= 0 || parsedLimit > 100 {
			return apperrors.ErrBadRequest("Invalid limit parameter", err)
		}
		limit = parsedLimit
	}

	offset := 0 // Default offset
	if offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err != nil || parsedOffset < 0 {
			return apperrors.ErrBadRequest("Invalid offset parameter", err)
		}
		offset = parsedOffset
	}

	groups, err := helpers.ListGroups(limit, offset)
	if err != nil {
		return apperrors.ErrInternalServer("Failed to list groups", err)
	}

	groupResponses := make([]GroupResponse, len(groups))
	for i, group := range groups {
		groupResponses[i] = mapModelGroupToGroupResponse(group)
	}

	response := ListGroupsResponse{
		Groups: groupResponses,
		Limit:  limit,
		Offset: offset,
		Count:  len(groupResponses),
	}
	return helpers.RespondJSON(w, http.StatusOK, response)
}

// UpdateGroup godoc
// @Summary Update group
// @Description Update group details. Requires authentication and admin privileges.
// @Tags groups
// @Accept json
// @Produce json
// @Param id path string true "Group ID"
// @Param group body UpdateGroupRequest true "Group update details"
// @Success 200 {object} GroupResponse "Updated group details"
// @Failure 400 {object} map[string]string "Invalid request body or missing group name"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - only admins can update group"
// @Failure 404 {object} map[string]string "Group not found"
// @Failure 405 {object} map[string]string "Method not allowed"
// @Failure 500 {object} map[string]string "Failed to update group"
// @Router /groups/update/{id} [put]
func UpdateGroup(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPut {
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

	// Get existing group
	group, err := helpers.GetGroupByID(groupID)
	if err != nil {
		if errors.Is(err, helpers.ErrGroupNotFound) {
			return apperrors.ErrNotFound("Group not found", err)
		}
		return apperrors.ErrInternalServer("Failed to get group", err)
	}

	// Check if user is admin
	isAdmin, err := helpers.IsGroupAdmin(groupID, currentUser.ID)
	if err != nil {
		return apperrors.ErrInternalServer("Failed to check admin status", err)
	}
	if !isAdmin {
		return apperrors.ErrForbidden("Unauthorized - only admins can update group", nil)
	}

	// Parse request body
	var req UpdateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apperrors.ErrBadRequest("Invalid request body", err)
	}

	// Update group data
	group.Name = req.Name
	group.Description = req.Description
	group.AvatarURL = req.AvatarURL

	if err := helpers.UpdateGroup(group); err != nil {
		return apperrors.ErrInternalServer("Failed to update group", err)
	}

	response := mapModelGroupToGroupResponse(group)
	return helpers.RespondJSON(w, http.StatusOK, response)
}

// DeleteGroup godoc
// @Summary Delete group
// @Description Delete a group by ID. Requires authentication and group creator privileges.
// @Tags groups
// @Accept json
// @Produce json
// @Param id path string true "Group ID"
// @Success 200 {object} DeleteGroupResponse "Group deleted successfully"
// @Failure 400 {object} map[string]string "Group ID is required"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - only the creator can delete the group"
// @Failure 404 {object} map[string]string "Group not found"
// @Failure 405 {object} map[string]string "Method not allowed"
// @Failure 500 {object} map[string]string "Failed to delete group"
// @Router /groups/delete/{id} [delete]
func DeleteGroup(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodDelete {
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

	// Get group to check if user is creator
	group, err := helpers.GetGroupByID(groupID)
	if err != nil {
		if errors.Is(err, helpers.ErrGroupNotFound) {
			return apperrors.ErrNotFound("Group not found", err)
		}
		return apperrors.ErrInternalServer("Failed to get group", err)
	}

	// Only the creator can delete the group
	if group.CreatorID != currentUser.ID {
		return apperrors.ErrForbidden("Unauthorized - only the creator can delete the group", nil)
	}

	if err := helpers.DeleteGroup(groupID); err != nil {
		return apperrors.ErrInternalServer("Failed to delete group", err)
	}

	response := DeleteGroupResponse{Message: "Group deleted successfully"}
	return helpers.RespondJSON(w, http.StatusOK, response)
}
