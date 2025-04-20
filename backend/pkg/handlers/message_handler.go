package handlers

import (
	"encoding/json"
	"log" // Added
	"net/http"
	// "strconv" // Will be handled by ParsePagination later
	"time" // Added for DTO

	"github.com/HASANALI117/social-network/pkg/apperrors" // Added
	"github.com/HASANALI117/social-network/pkg/helpers"
	"github.com/HASANALI117/social-network/pkg/models" // Added
)

// --- DTOs ---

// MessageResponse defines the structure for a message returned by the API.
type MessageResponse struct {
	ID         string    `json:"id"`
	SenderID   string    `json:"sender_id"`
	ReceiverID string    `json:"receiver_id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	// Add UpdatedAt if applicable
}

// ListMessagesResponse defines the structure for the list messages endpoint.
type ListMessagesResponse struct {
	Messages []MessageResponse `json:"messages"`
	Limit    int               `json:"limit"`
	Offset   int               `json:"offset"`
	Count    int               `json:"count"`
}

// --- Helper Function ---

// mapModelMessageToMessageResponse converts a models.Message to a MessageResponse DTO.
func mapModelMessageToMessageResponse(msg *models.Message) MessageResponse {
	return MessageResponse{
		ID:         msg.ID,
		SenderID:   msg.SenderID,
		ReceiverID: msg.ReceiverID,
		Content:    msg.Content,
		CreatedAt:  msg.CreatedAt,
	}
}

// --- Handler ---

// GetMessages godoc
// @Summary Get direct messages between two users
// @Description Get a paginated list of direct messages between a sender and receiver. Requires authentication, and the requester must be one of the participants.
// @Tags messages
// @Accept json
// @Produce json
// @Param sender_id query string true "Sender User ID"
// @Param receiver_id query string true "Receiver User ID"
// @Param limit query int false "Number of messages to return (default 50)" minimum(1) maximum(200)
// @Param offset query int false "Number of messages to skip (default 0)" minimum(0)
// @Success 200 {object} ListMessagesResponse "List of messages"
// @Failure 400 {object} map[string]string "Missing sender_id or receiver_id, or invalid pagination"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden (requester is not part of the conversation)"
// @Failure 405 {object} map[string]string "Method not allowed"
// @Failure 500 {object} map[string]string "Failed to fetch messages"
// @Router /messages [get]
func GetMessages(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return apperrors.ErrMethodNotAllowed("", nil)
	}

	// Get current user from session for authorization
	// TODO: Replace helpers.GetUserFromSession when auth middleware/service is added
	currentUser, err := helpers.GetUserFromSession(r)
	if err != nil {
		return apperrors.ErrUnauthorized("Unauthorized", err)
	}

	senderID := r.URL.Query().Get("sender_id")
	receiverID := r.URL.Query().Get("receiver_id")

	if senderID == "" || receiverID == "" {
		return apperrors.ErrBadRequest("Both sender_id and receiver_id query parameters are required", nil)
	}

	// Authorization check: Ensure the current user is either the sender or receiver
	if currentUser.ID != senderID && currentUser.ID != receiverID {
		log.Printf("WARN: User %s attempted to access messages between %s and %s", currentUser.ID, senderID, receiverID)
		return apperrors.ErrForbidden("You are not authorized to view these messages", nil)
	}

	// Parse pagination parameters
	// TODO: Create and use helpers.ParsePagination
	limit := 50 // Default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit <= 0 || parsedLimit > 200 { // Add max limit
			return apperrors.ErrBadRequest("Invalid limit parameter", err)
		}
		limit = parsedLimit
	}
	offset := 0 // Default offset
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err != nil || parsedOffset < 0 {
			return apperrors.ErrBadRequest("Invalid offset parameter", err)
		}
		offset = parsedOffset
	}
	// --- End TODO for ParsePagination ---


	// Get messages from database
	// TODO: Replace helpers.GetUserMessages when service layer is added
	messages, err := helpers.GetUserMessages(senderID, receiverID, limit, offset)
	if err != nil {
		log.Printf("Failed to fetch messages between %s and %s: %v", senderID, receiverID, err)
		// TODO: Check for specific errors from GetUserMessages if available (e.g., user not found)
		return apperrors.ErrInternalServer("Failed to fetch messages", err)
	}

	// Map models to response DTOs
	messageResponses := make([]MessageResponse, len(messages))
	for i, msg := range messages {
		messageResponses[i] = mapModelMessageToMessageResponse(msg)
	}

	// Return message list DTO
	response := ListMessagesResponse{
		Messages: messageResponses,
		Limit:    limit,
		Offset:   offset,
		Count:    len(messageResponses),
	}
	return helpers.RespondJSON(w, http.StatusOK, response)
}
