package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/HASANALI117/social-network/pkg/helpers"
	"github.com/HASANALI117/social-network/pkg/httperr"
	"github.com/HASANALI117/social-network/pkg/models"
)

// CreateGroup godoc
// @Summary Create a new group
// @Description Create a new group chat
// @Tags groups
// @Accept json
// @Produce json
// @Param group body models.Group true "Group creation details"
// @Success 201 {object} map[string]interface{} "Group created successfully"
// @Failure 400 {string} string "Invalid request body"
// @Failure 400 {object} httperr.ErrorResponse "Invalid request body or Group name is required"
// @Failure 401 {object} httperr.ErrorResponse "Unauthorized"
// @Failure 405 {object} httperr.ErrorResponse "Method not allowed"
// @Failure 500 {object} httperr.ErrorResponse "Failed to create group"
// @Router /groups/create [post]
func CreateGroup(w http.ResponseWriter, r *http.Request) error { // Changed signature
	if r.Method != http.MethodPost {
		return httperr.NewMethodNotAllowed(nil, "")
	}

	// Get current user from session
	currentUser, err := helpers.GetUserFromSession(r)
	if err != nil {
		return httperr.NewUnauthorized(err, "")
	}

	var group models.Group
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		return httperr.NewBadRequest(err, "Invalid request body")
	}

	// Set creator ID to current user
	group.CreatorID = currentUser.ID

	// Basic validation
	if group.Name == "" {
		return httperr.NewBadRequest(nil, "Group name is required")
	}

	if err := helpers.CreateGroup(&group); err != nil {
		return httperr.NewInternalServerError(err, "Failed to create group")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          group.ID,
		"name":        group.Name,
		"description": group.Description,
		"creator_id":  group.CreatorID,
		"avatar_url":  group.AvatarURL,
		"created_at":  group.CreatedAt,
	})
	return nil // Return nil on success
}

// GetGroup godoc
// @Summary Get group by ID
// @Description Get group details by group ID
// @Tags groups
// @Accept json
// @Produce json
// @Param id query string true "Group ID"
// @Success 200 {object} map[string]interface{} "Group details"
// @Failure 400 {object} httperr.ErrorResponse "Group ID is required"
// @Failure 404 {object} httperr.ErrorResponse "Group not found"
// @Failure 405 {object} httperr.ErrorResponse "Method not allowed"
// @Failure 500 {object} httperr.ErrorResponse "Failed to get group"
// @Router /groups/get [get]
func GetGroup(w http.ResponseWriter, r *http.Request) error { // Changed signature
	if r.Method != http.MethodGet {
		return httperr.NewMethodNotAllowed(nil, "")
	}

	groupID := r.URL.Query().Get("id")
	if groupID == "" {
		return httperr.NewBadRequest(nil, "Group ID is required")
	}

	group, err := helpers.GetGroupByID(groupID)
	if err != nil {
		if errors.Is(err, helpers.ErrGroupNotFound) {
			return httperr.NewNotFound(err, "Group not found")
		}
		return httperr.NewInternalServerError(err, "Failed to get group")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          group.ID,
		"name":        group.Name,
		"description": group.Description,
		"creator_id":  group.CreatorID,
		"avatar_url":  group.AvatarURL,
		"created_at":  group.CreatedAt,
		"updated_at":  group.UpdatedAt,
	})
	return nil // Return nil on success
}

// ListGroups godoc
// @Summary List groups
// @Description Get a paginated list of groups
// @Tags groups
// @Accept json
// @Produce json
// @Param limit query int false "Number of groups to return (default 10)"
// @Param offset query int false "Number of groups to skip (default 0)"
// @Success 200 {object} map[string]interface{} "List of groups"
// @Failure 405 {object} httperr.ErrorResponse "Method not allowed"
// @Failure 500 {object} httperr.ErrorResponse "Failed to list groups"
// @Router /groups/list [get]
func ListGroups(w http.ResponseWriter, r *http.Request) error { // Changed signature
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

	groups, err := helpers.ListGroups(limit, offset)
	if err != nil {
		return httperr.NewInternalServerError(err, "Failed to list groups")
	}

	result := make([]map[string]interface{}, len(groups))
	for i, group := range groups {
		result[i] = map[string]interface{}{
			"id":          group.ID,
			"name":        group.Name,
			"description": group.Description,
			"creator_id":  group.CreatorID,
			"avatar_url":  group.AvatarURL,
			"created_at":  group.CreatedAt,
			"updated_at":  group.UpdatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"groups": result,
		"limit":  limit,
		"offset": offset,
		"count":  len(groups),
	})
	return nil // Return nil on success
}

// UpdateGroup godoc
// @Summary Update group
// @Description Update group details
// @Tags groups
// @Accept json
// @Produce json
// @Param id query string true "Group ID"
// @Param group body object true "Group update details"
// @Success 200 {object} map[string]interface{} "Updated group details"
// @Failure 400 {object} httperr.ErrorResponse "Invalid request or Group ID is required"
// @Failure 401 {object} httperr.ErrorResponse "Unauthorized"
// @Failure 404 {object} httperr.ErrorResponse "Group not found"
// @Failure 405 {object} httperr.ErrorResponse "Method not allowed"
// @Failure 500 {object} httperr.ErrorResponse "Failed to update group or Failed to get group"
// @Router /groups/update [put]
func UpdateGroup(w http.ResponseWriter, r *http.Request) error { // Changed signature
	if r.Method != http.MethodPut {
		return httperr.NewMethodNotAllowed(nil, "")
	}

	// Get current user from session
	currentUser, err := helpers.GetUserFromSession(r)
	if err != nil {
		return httperr.NewUnauthorized(err, "")
	}

	groupID := r.URL.Query().Get("id")
	if groupID == "" {
		return httperr.NewBadRequest(nil, "Group ID is required")
	}

	// Get existing group
	group, err := helpers.GetGroupByID(groupID)
	if err != nil {
		if errors.Is(err, helpers.ErrGroupNotFound) {
			return httperr.NewNotFound(err, "Group not found")
		}
		return httperr.NewInternalServerError(err, "Failed to get group")
	}

	// Check if user is admin
	isAdmin, err := helpers.IsGroupAdmin(groupID, currentUser.ID)
	if err != nil {
		// Log the underlying error but return Unauthorized
		return httperr.NewUnauthorized(err, "Unauthorized - only admins can update group")
	}
	if !isAdmin {
		return httperr.NewUnauthorized(nil, "Unauthorized - only admins can update group")
	}

	// Parse request body
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		AvatarURL   string `json:"avatar_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return httperr.NewBadRequest(err, "Invalid request body")
	}

	// Update group data
	group.Name = req.Name
	group.Description = req.Description
	group.AvatarURL = req.AvatarURL

	if err := helpers.UpdateGroup(group); err != nil {
		return httperr.NewInternalServerError(err, "Failed to update group")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          group.ID,
		"name":        group.Name,
		"description": group.Description,
		"creator_id":  group.CreatorID,
		"avatar_url":  group.AvatarURL,
		"created_at":  group.CreatedAt,
		"updated_at":  group.UpdatedAt,
	})
	return nil // Return nil on success
}

// DeleteGroup godoc
// @Summary Delete group
// @Description Delete a group by ID
// @Tags groups
// @Accept json
// @Produce json
// @Param id query string true "Group ID"
// @Success 200 {object} map[string]string "Group deleted successfully"
// @Failure 400 {object} httperr.ErrorResponse "Group ID is required"
// @Failure 401 {object} httperr.ErrorResponse "Unauthorized"
// @Failure 404 {object} httperr.ErrorResponse "Group not found"
// @Failure 405 {object} httperr.ErrorResponse "Method not allowed"
// @Failure 500 {object} httperr.ErrorResponse "Failed to delete group or Failed to get group"
// @Router /groups/delete [delete]
func DeleteGroup(w http.ResponseWriter, r *http.Request) error { // Changed signature
	if r.Method != http.MethodDelete {
		return httperr.NewMethodNotAllowed(nil, "")
	}

	// Get current user from session
	currentUser, err := helpers.GetUserFromSession(r)
	if err != nil {
		return httperr.NewUnauthorized(err, "")
	}

	groupID := r.URL.Query().Get("id")
	if groupID == "" {
		return httperr.NewBadRequest(nil, "Group ID is required")
	}

	// Get group to check if user is creator
	group, err := helpers.GetGroupByID(groupID)
	if err != nil {
		if errors.Is(err, helpers.ErrGroupNotFound) {
			return httperr.NewNotFound(err, "Group not found")
		}
		return httperr.NewInternalServerError(err, "Failed to get group")
	}

	// Only the creator can delete the group
	if group.CreatorID != currentUser.ID {
		return httperr.NewUnauthorized(nil, "Unauthorized - only the creator can delete the group")
	}

	if err := helpers.DeleteGroup(groupID); err != nil {
		if errors.Is(err, helpers.ErrGroupNotFound) {
			// This might be redundant if DeleteGroup already checked, but safe to keep
			return httperr.NewNotFound(err, "Group not found")
		}
		return httperr.NewInternalServerError(err, "Failed to delete group")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Group deleted successfully",
	})
	return nil // Return nil on success
}
