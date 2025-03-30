package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/HASANALI117/social-network/pkg/helpers"
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
// @Failure 500 {string} string "Failed to create group"
// @Router /groups/create [post]
func CreateGroup(w http.ResponseWriter, r *http.Request) {
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

	var group models.Group
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set creator ID to current user
	group.CreatorID = currentUser.ID

	// Basic validation
	if group.Name == "" {
		http.Error(w, "Group name is required", http.StatusBadRequest)
		return
	}

	if err := helpers.CreateGroup(&group); err != nil {
		http.Error(w, "Failed to create group", http.StatusInternalServerError)
		return
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
}

// GetGroup godoc
// @Summary Get group by ID
// @Description Get group details by group ID
// @Tags groups
// @Accept json
// @Produce json
// @Param id query string true "Group ID"
// @Success 200 {object} map[string]interface{} "Group details"
// @Failure 400 {string} string "Group ID is required"
// @Failure 404 {string} string "Group not found"
// @Failure 500 {string} string "Failed to get group"
// @Router /groups/get [get]
func GetGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	groupID := r.URL.Query().Get("id")
	if groupID == "" {
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	group, err := helpers.GetGroupByID(groupID)
	if err != nil {
		if errors.Is(err, helpers.ErrGroupNotFound) {
			http.Error(w, "Group not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get group", http.StatusInternalServerError)
		}
		return
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
// @Failure 500 {string} string "Failed to list groups"
// @Router /groups/list [get]
func ListGroups(w http.ResponseWriter, r *http.Request) {
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

	groups, err := helpers.ListGroups(limit, offset)
	if err != nil {
		http.Error(w, "Failed to list groups", http.StatusInternalServerError)
		return
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
// @Failure 400 {string} string "Invalid request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Group not found"
// @Failure 500 {string} string "Failed to update group"
// @Router /groups/update [put]
func UpdateGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
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

	// Get existing group
	group, err := helpers.GetGroupByID(groupID)
	if err != nil {
		if errors.Is(err, helpers.ErrGroupNotFound) {
			http.Error(w, "Group not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get group", http.StatusInternalServerError)
		}
		return
	}

	// Check if user is admin
	isAdmin, err := helpers.IsGroupAdmin(groupID, currentUser.ID)
	if err != nil || !isAdmin {
		http.Error(w, "Unauthorized - only admins can update group", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		AvatarURL   string `json:"avatar_url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update group data
	group.Name = req.Name
	group.Description = req.Description
	group.AvatarURL = req.AvatarURL

	if err := helpers.UpdateGroup(group); err != nil {
		http.Error(w, "Failed to update group", http.StatusInternalServerError)
		return
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
}

// DeleteGroup godoc
// @Summary Delete group
// @Description Delete a group by ID
// @Tags groups
// @Accept json
// @Produce json
// @Param id query string true "Group ID"
// @Success 200 {object} map[string]string "Group deleted successfully"
// @Failure 400 {string} string "Group ID is required"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "Group not found"
// @Failure 500 {string} string "Failed to delete group"
// @Router /groups/delete [delete]
func DeleteGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
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

	// Get group to check if user is creator
	group, err := helpers.GetGroupByID(groupID)
	if err != nil {
		if errors.Is(err, helpers.ErrGroupNotFound) {
			http.Error(w, "Group not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get group", http.StatusInternalServerError)
		}
		return
	}

	// Only the creator can delete the group
	if group.CreatorID != currentUser.ID {
		http.Error(w, "Unauthorized - only the creator can delete the group", http.StatusUnauthorized)
		return
	}

	if err := helpers.DeleteGroup(groupID); err != nil {
		if errors.Is(err, helpers.ErrGroupNotFound) {
			http.Error(w, "Group not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete group", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Group deleted successfully",
	})
}
