package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/HASANALI117/social-network/pkg/helpers"
	"github.com/HASANALI117/social-network/pkg/httperr"
	"github.com/HASANALI117/social-network/pkg/models"
	"github.com/HASANALI117/social-network/pkg/services"
)

type MessageHandler struct {
	messageService services.MessageService
	authService   services.AuthService
}

func NewMessageHandler(messageService services.MessageService, authService services.AuthService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
		authService:   authService,
	}
}

// GetChatConversations godoc
// @Summary Get chat conversations for current user
// @Description Get a list of users with whom the current user has exchanged messages, sorted by most recent message
// @Tags messages
// @Accept json
// @Produce json
// @Success 200 {array} models.ChatPartner
// @Failure 401 {object} httperr.ErrorResponse "Unauthorized"
// @Failure 500 {object} httperr.ErrorResponse "Failed to fetch conversations"
// @Router /messages/conversations [get]
func (h *MessageHandler) GetChatConversations(w http.ResponseWriter, r *http.Request) error {
	// Get current user from session
	currentUser, err := helpers.GetUserFromSession(r, h.authService)
	if err != nil {
		if errors.Is(err, helpers.ErrInvalidSession) {
			return httperr.NewUnauthorized(err, "Invalid session")
		}
		return httperr.NewInternalServerError(err, "Failed to get current user")
	}

	// Get chat partners from service
	partners, err := h.messageService.GetChatPartners(currentUser.ID)
	if err != nil {
		return httperr.NewInternalServerError(err, "Failed to fetch chat conversations")
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(partners)
}

// GetMessages godoc
// @Summary Get direct messages between users
// @Description Get paginated direct messages between the current user and a target user
// @Tags messages
// @Accept json
// @Produce json
// @Param targetUserId query string true "Target user ID"
// @Param limit query int false "Number of messages to return (default 20)"
// @Param offset query int false "Number of messages to skip (default 0)"
// @Success 200 {object} map[string]interface{} "Messages response with pagination"
// @Failure 400 {object} httperr.ErrorResponse "Missing targetUserId"
// @Failure 401 {object} httperr.ErrorResponse "Unauthorized"
// @Failure 500 {object} httperr.ErrorResponse "Failed to fetch messages"
// @Router /messages [get]
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
    targetUserID := r.URL.Query().Get("targetUserId")
    if targetUserID == "" {
        return httperr.NewBadRequest(nil, "targetUserId query parameter is required")
    }

    // Get pagination parameters
    limit := 20 // Default limit
    offset := 0 // Default offset
    
    if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
        if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
            limit = parsedLimit
        }
    }
    
    if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
        if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
            offset = parsedOffset
        }
    }

    // Get messages from service
    messages, totalCount, err := h.messageService.GetDirectMessagesBetweenUsers(currentUser.ID, targetUserID, limit, offset)
    if err != nil {
        return httperr.NewInternalServerError(err, "Failed to fetch messages")
    }

    // Prepare response
    response := struct {
        Messages []models.Message `json:"messages"`
        TotalCount int64 `json:"total_count"`
        Limit int `json:"limit"`
        Offset int `json:"offset"`
    }{
        Messages: messages,
        TotalCount: totalCount,
        Limit: limit,
        Offset: offset,
    }

    // Write response
    w.Header().Set("Content-Type", "application/json")
    return json.NewEncoder(w).Encode(response)
}
