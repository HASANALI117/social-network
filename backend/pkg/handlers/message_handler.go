package handlers

import (
	"encoding/json"
	"errors" // Added import
	"net/http"
	"strconv"

	"github.com/HASANALI117/social-network/pkg/helpers"
	"github.com/HASANALI117/social-network/pkg/httperr"
	"github.com/HASANALI117/social-network/pkg/services" // Added import
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

// GetMessages godoc
// @Summary Get user messages
// @Description Get messages for a specific user
// @Tags messages
// @Accept json
// @Produce json
// @Param user_id query string true "User ID"
// @Param limit query int false "Number of messages to return (default 50)"
// @Param offset query int false "Number of messages to skip (default 0)"
// @Success 200 {object} []models.Message
// @Failure 400 {object} httperr.ErrorResponse "Missing required parameters"
// @Failure 401 {object} httperr.ErrorResponse "Unauthorized"
// @Failure 405 {object} httperr.ErrorResponse "Method not allowed"
// @Failure 500 {object} httperr.ErrorResponse "Failed to fetch messages or session error"
// @Router /messages [get]
func (h *MessageHandler) GetMessages(w http.ResponseWriter, r *http.Request) error {
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

	senderID := r.URL.Query().Get("sender_id")
	receiverID := r.URL.Query().Get("receiver_id")

	if senderID == "" || receiverID == "" {
		return httperr.NewBadRequest(nil, "Both sender_id and receiver_id are required")
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil {
			offset = parsedOffset
		}
	}

	// Get messages using MessageService, passing current user ID for authorization
	messages, err := h.messageService.GetUserMessages(senderID, receiverID, limit, offset, currentUser.ID)
	if err != nil {
		// Handle specific service errors (e.g., authorization error)
		// Assuming MessageService returns a specific error for authorization failure
		// if errors.Is(err, services.ErrUnauthorized) { // Example error check
		// 	return httperr.NewUnauthorized(err, "Not authorized to view these messages")
		// }
		// Handle other potential errors
		return httperr.NewInternalServerError(err, "Failed to fetch messages")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
	return nil
}
