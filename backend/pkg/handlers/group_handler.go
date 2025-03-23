package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"social-network/pkg/helpers"
	"social-network/pkg/models"

	"github.com/google/uuid"
)

// CreateGroup godoc
// @Summary Create a new group
// @Description Create a new group in the system
// @Tags groups
// @Accept json
// @Produce json
// @Param group body object true "Group creation details"
// @Success 201 {object} map[string]interface{} "Group created"
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Failed to create group"
// @Router /groups/create [post]
func CreateGroup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		CreatorID   string `json:"creator_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	group := &models.Group{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		CreatorID:   req.CreatorID,
		CreatedAt:   time.Now(),
	}
	if err := helpers.CreateGroup(group); err != nil {
		http.Error(w, "Failed to create group", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":          group.ID,
		"title":       group.Title,
		"description": group.Description,
		"creator_id":  group.CreatorID,
		"created_at":  group.CreatedAt,
	})
}

// AddGroupMember godoc
// @Summary Invite user to group
// @Description Send a group invitation
// @Tags groups
// @Accept json
// @Produce json
// @Param group_id query string true "Group ID"
// @Param user_id query string true "User ID"
// @Success 200 {string} string "Invitation sent"
// @Failure 400 {string} string "Invalid request"
// @Failure 500 {string} string "Failed to send invitation"
// @Router /groups/invite [post]
func AddGroupMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	groupID := r.URL.Query().Get("group_id")
	userID := r.URL.Query().Get("user_id")
	if groupID == "" || userID == "" {
		http.Error(w, "Group ID and User ID are required", http.StatusBadRequest)
		return
	}

	if err := helpers.AddGroupMember(groupID, userID, "pending"); err != nil {
		http.Error(w, "Failed to send invitation", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Invitation sent"})
}
