package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/HASANALI117/social-network/pkg/helpers"
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
// @Router /messages [get]
func GetMessages(w http.ResponseWriter, r *http.Request) {
	senderID := r.URL.Query().Get("sender_id")
	receiverID := r.URL.Query().Get("receiver_id")

	if senderID == "" || receiverID == "" {
		http.Error(w, "Both sender_id and receiver_id are required", http.StatusBadRequest)
		return
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
		http.Error(w, "Failed to fetch messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}
