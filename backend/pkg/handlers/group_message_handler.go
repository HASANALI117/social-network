package handlers

import (
	"encoding/json"
	"errors" // Import errors
	"net/http"
	"strconv"

	"github.com/HASANALI117/social-network/pkg/helpers"
	"github.com/HASANALI117/social-network/pkg/httperr"
	"github.com/HASANALI117/social-network/pkg/repositories" // Added import for repository errors
	"github.com/HASANALI117/social-network/pkg/services"     // Import services
)

// GroupMessageHandler handles group message requests
type GroupMessageHandler struct {
	authService    services.AuthService
	messageService services.MessageService // Inject MessageService
	groupService   services.GroupService   // Inject GroupService
}

// NewGroupMessageHandler creates a new GroupMessageHandler
func NewGroupMessageHandler(messageService services.MessageService, groupService services.GroupService, authService services.AuthService) *GroupMessageHandler {
	return &GroupMessageHandler{
		authService:    authService,
		messageService: messageService, // Assign MessageService
		groupService:   groupService,   // Assign GroupService
	}
}

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
// @Failure 400 {object} httperr.ErrorResponse "Group ID is required"
// @Failure 401 {object} httperr.ErrorResponse "Unauthorized"
// @Failure 405 {object} httperr.ErrorResponse "Method not allowed"
// @Failure 500 {object} httperr.ErrorResponse "Failed to get messages or session error"
// @Router /groups/messages [get]
func (h *GroupMessageHandler) GetGroupMessages(w http.ResponseWriter, r *http.Request) error {
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

	// Check if user is a member of the group using GroupService
	isMember, err := h.groupService.IsMember(groupID, currentUser.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrGroupNotFound) { // Assuming GroupService propagates repo errors
			return httperr.NewNotFound(err, "Group not found")
		}
		return httperr.NewInternalServerError(err, "Failed to check member status")
	}
	if !isMember {
		return httperr.NewUnauthorized(nil, "Only members can view messages")
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

	// Get messages using MessageService
	// Pass currentUser.ID as requestingUserID for authorization check within the service
	messages, err := h.messageService.GetGroupMessages(groupID, limit, offset, currentUser.ID)
	if err != nil {
		// Handle specific service errors
		if errors.Is(err, services.ErrGroupMemberRequired) {
			// This shouldn't happen if the IsMember check above passed, but handle defensively
			return httperr.NewUnauthorized(err, "Membership required to view messages")
		}
		if errors.Is(err, repositories.ErrGroupNotFound) { // Assuming MessageService propagates repo errors via GroupRepo check
			return httperr.NewNotFound(err, "Group not found")
		}
		// Handle other potential errors
		return httperr.NewInternalServerError(err, "Failed to get group messages")
	}

	// The service returns []*models.GroupMessage, which should be suitable for JSON encoding
	// TODO: Consider mapping to a specific response DTO if needed
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": messages, // Use the response directly from the service
		"limit":    limit,
		"offset":   offset,
		"count":    len(messages),
	})
	return nil
}
