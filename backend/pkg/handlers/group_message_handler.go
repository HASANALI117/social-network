package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/HASANALI117/social-network/pkg/helpers"
)

// GetGroupMessages godoc
// @Summary Get group messages
// @Description Get messages from a group with pagination
// @Tags groups
// @Accept json
// @Produce json
// @Param id query string true "Group ID"
// @Param limit query int false "Number of messages to return (default 50)"
// @Param offset query int false "Number of messages to skip (default 0)"
// @Success 200 {object} map[string]interface{} "Group messages"
// @Failure 400 {string} string "Group ID is required"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Failed to get messages"
// @Router /groups/messages [get]
func GetGroupMessages(w http.ResponseWriter, r *http.Request) {
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

	// Check if user is a member of the group
	isMember, err := helpers.IsGroupMember(groupID, currentUser.ID)
	if err != nil || !isMember {
		http.Error(w, "Unauthorized - only members can view messages", http.StatusUnauthorized)
		return
	}

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // Default limit for messages
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

	messages, err := helpers.GetGroupMessages(groupID, limit, offset)
	if err != nil {
		http.Error(w, "Failed to get group messages", http.StatusInternalServerError)
		return
	}

	result := make([]map[string]interface{}, len(messages))
	for i, msg := range messages {
		result[i] = map[string]interface{}{
			"id":         msg.ID,
			"group_id":   msg.GroupID,
			"sender_id":  msg.SenderID,
			"content":    msg.Content,
			"created_at": msg.CreatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": result,
		"limit":    limit,
		"offset":   offset,
		"count":    len(messages),
	})
}
