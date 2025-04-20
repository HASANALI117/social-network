package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/HASANALI117/social-network/pkg/apperrors"
	"github.com/HASANALI117/social-network/pkg/helpers"
	"github.com/HASANALI117/social-network/pkg/models"
)

// --- DTOs ---

// GroupMessageResponse defines the structure for a group message returned by the API.
type GroupMessageResponse struct {
	ID        string    `json:"id"`
	GroupID   string    `json:"group_id"`
	SenderID  string    `json:"sender_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

// ListGroupMessagesResponse defines the structure for listing group messages.
type ListGroupMessagesResponse struct {
	Messages []GroupMessageResponse `json:"messages"`
	Limit    int                  `json:"limit"`
	Offset   int                  `json:"offset"`
	Count    int                  `json:"count"`
}

// mapModelGroupMessageToGroupMessageResponse converts a models.GroupMessage to a GroupMessageResponse DTO.
func mapModelGroupMessageToGroupMessageResponse(msg *models.GroupMessage) GroupMessageResponse {
	return GroupMessageResponse{
		ID:        msg.ID,
		GroupID:   msg.GroupID,
		SenderID:  msg.SenderID,
		Content:   msg.Content,
		CreatedAt: msg.CreatedAt,
	}
}

// GetGroupMessages godoc
// @Summary Get group messages
// @Description Get messages from a group with pagination. Requires authentication and group membership.
// @Tags groups
// @Accept json
// @Produce json
// @Param id query string true "Group ID"
// @Param limit query int false "Number of messages to return (default 50)" minimum(1) maximum(200)
// @Param offset query int false "Number of messages to skip (default 0)" minimum(0)
// @Success 200 {object} ListGroupMessagesResponse "Group messages"
// @Failure 400 {object} map[string]string "Group ID is required or invalid pagination parameters"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - only members can view messages"
// @Failure 405 {object} map[string]string "Method not allowed"
// @Failure 500 {object} map[string]string "Failed to get messages"
// @Router /groups/messages [get]
func GetGroupMessages(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
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

	// Check if user is a member of the group
	isMember, err := helpers.IsGroupMember(groupID, currentUser.ID)
	if err != nil {
		return apperrors.ErrInternalServer("Failed to check group membership", err)
	}
	if !isMember {
		return apperrors.ErrForbidden("Unauthorized - only members can view messages", nil)
	}

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // Default limit for messages
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit <= 0 || parsedLimit > 200 {
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

	// Get messages from database
	// TODO: Replace helpers.GetGroupMessages when service layer is added
	messages, err := helpers.GetGroupMessages(groupID, limit, offset)
	if err != nil {
		log.Printf("Failed to get group messages for group %s: %v", groupID, err)
		return apperrors.ErrInternalServer("Failed to get group messages", err)
	}

	// Map models to response DTOs
	messageResponses := make([]GroupMessageResponse, len(messages))
	for i, msg := range messages {
		messageResponses[i] = mapModelGroupMessageToGroupMessageResponse(msg)
	}

	// Return message list
	response := ListGroupMessagesResponse{
		Messages: messageResponses,
		Limit:    limit,
		Offset:   offset,
		Count:    len(messageResponses),
	}
	return helpers.RespondJSON(w, http.StatusOK, response)
}
