package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/HASANALI117/social-network/pkg/helpers"
	"github.com/HASANALI117/social-network/pkg/httperr"
)

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
// @Failure 405 {object} httperr.ErrorResponse "Method not allowed"
// @Failure 500 {object} httperr.ErrorResponse "Failed to fetch messages"
// @Router /messages [get]
func GetMessages(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return httperr.NewMethodNotAllowed(nil, "")
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

	messages, err := helpers.GetUserMessages(senderID, receiverID, limit, offset)
	if err != nil {
		return httperr.NewInternalServerError(err, "Failed to fetch messages")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
	return nil
}
