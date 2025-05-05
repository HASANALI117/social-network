package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/HASANALI117/social-network/pkg/helpers"
	"github.com/HASANALI117/social-network/pkg/httperr"
	"github.com/HASANALI117/social-network/pkg/models" // Assuming Message model is here
	"github.com/HASANALI117/social-network/pkg/services"
)

// MessageHandler handles direct message requests
type MessageHandler struct {
	authService    services.AuthService
	messageService services.MessageService
}

// NewMessageHandler creates a new MessageHandler
func NewMessageHandler(messageService services.MessageService, authService services.AuthService) *MessageHandler {
	return &MessageHandler{
		authService:    authService,
		messageService: messageService,
	}
}

// MessagesResponse defines the structure for the messages response including pagination
type MessagesResponse struct {
	Messages   []models.Message `json:"messages"`
	TotalCount int64            `json:"total_count"`
	Limit      int              `json:"limit"`
	Offset     int              `json:"offset"`
}

// GetMessages godoc
// @Summary Get direct messages with a user
// @Description Get paginated direct messages between the authenticated user and another user
// @Tags messages
// @Accept json
// @Produce json
// @Param targetUserId query string true "Target User ID"
// @Param limit query int false "Number of messages to return (default 20)"
// @Param offset query int false "Number of messages to skip (default 0)"
// @Success 200 {object} handlers.MessagesResponse
// @Failure 400 {object} httperr.ErrorResponse "Invalid user ID format or missing parameters"
// @Failure 401 {object} httperr.ErrorResponse "Unauthorized"
// @Failure 404 {object} httperr.ErrorResponse "User not found" // Added possibility
// @Failure 405 {object} httperr.ErrorResponse "Method not allowed"
// @Failure 500 {object} httperr.ErrorResponse "Failed to fetch messages or session error"
// @Router /messages [get] // Use query param for target user
func (h *MessageHandler) GetMessages(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return httperr.NewMethodNotAllowed(nil, "")
	}

	// Get current user from session
	currentUser, err := helpers.GetUserFromSession(r, h.authService)
	if err != nil {
		if errors.Is(err, helpers.ErrInvalidSession) {
			return httperr.NewUnauthorized(err, "Invalid session")
		}
		return httperr.NewInternalServerError(err, "Failed to get current user")
	}

	// Get target user ID from query parameters
	targetUserIDStr := r.URL.Query().Get("targetUserId")
	if targetUserIDStr == "" {
		return httperr.NewBadRequest(nil, "targetUserId query parameter is required")
	}
	// Add any necessary validation for targetUserIDStr here if needed

	// Get limit and offset from query parameters
	limit := 20 // Default limit updated to 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 { // Ensure limit is positive
			limit = parsedLimit
		} else if err != nil {
			return httperr.NewBadRequest(err, "Invalid limit parameter format")
		}
	}

	offset := 0 // Default offset
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 { // Ensure offset is non-negative
			offset = parsedOffset
		} else if err != nil {
			return httperr.NewBadRequest(err, "Invalid offset parameter format")
		}
	}

	// Get messages using MessageService
	// Assuming a service method like GetDirectMessagesBetweenUsers exists or will be created
	messages, totalCount, err := h.messageService.GetDirectMessagesBetweenUsers(currentUser.ID, targetUserIDStr, limit, offset)
	if err != nil {
		// Handle specific service errors (e.g., user not found, authorization)
		// if errors.Is(err, services.ErrUserNotFound) { // Example
		// 	return httperr.NewNotFound(err, "Target user not found")
		// }
		return httperr.NewInternalServerError(err, "Failed to fetch messages")
	}

	// Prepare response
	response := MessagesResponse{
		Messages:   messages,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // Explicitly set status code
	json.NewEncoder(w).Encode(response)
	return nil
}
