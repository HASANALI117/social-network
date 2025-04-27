package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings" // For path parsing

	"github.com/HASANALI117/social-network/pkg/helpers"
	"github.com/HASANALI117/social-network/pkg/httperr"
	"github.com/HASANALI117/social-network/pkg/models"
	"github.com/HASANALI117/social-network/pkg/services"
)

// FollowerHandler handles HTTP requests related to followers
type FollowerHandler struct {
	service     services.FollowerService
	authService services.AuthService
}

// NewFollowerHandler creates a new FollowerHandler
func NewFollowerHandler(s services.FollowerService, as services.AuthService) *FollowerHandler {
	return &FollowerHandler{
		service:     s,
		authService: as,
	}
}

// Helper function to write JSON response
func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("Error encoding JSON response: %v", err)
		}
	}
}

// Helper function to write error response
func writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	log.Printf("HTTP Error %d: %s", statusCode, message)
	writeJSONResponse(w, statusCode, map[string]string{"error": message})
}

// getAuthenticatedUserID retrieves the user ID from the session cookie
func (h *FollowerHandler) getAuthenticatedUserID(r *http.Request) (string, error) {
	currentUser, err := helpers.GetUserFromSession(r, h.authService)
	if err != nil {
		return "", httperr.NewUnauthorized(err, "Invalid session")
	}
	return currentUser.ID, nil
}

// HandleFollowRequest sends a follow request to another user
// POST /users/{target_id}/follow
func (h *FollowerHandler) HandleFollowRequest(w http.ResponseWriter, r *http.Request) {
	requesterID, err := h.getAuthenticatedUserID(r)
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) != 4 || pathParts[0] != "api" || pathParts[1] != "users" || pathParts[3] != "follow" {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid URL path format for follow request")
		return
	}
	targetID := pathParts[2]
	if targetID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Target user ID cannot be empty")
		return
	}

	err = h.service.RequestFollow(requesterID, targetID)
	if err != nil {
		log.Printf("Error in HandleFollowRequest service call: %v", err)
		if err.Error() == "already following this user" || err.Error() == "follow request already pending" || err.Error() == "cannot follow yourself" {
			writeErrorResponse(w, http.StatusBadRequest, err.Error())
		} else {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to send follow request")
		}
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{"message": "Follow request sent successfully"})
}

// HandleAcceptRequest accepts a pending follow request
// POST /users/{requester_id}/accept
func (h *FollowerHandler) HandleAcceptRequest(w http.ResponseWriter, r *http.Request) {
	accepterID, err := h.getAuthenticatedUserID(r)
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) != 4 || pathParts[0] != "api" || pathParts[1] != "users" || pathParts[3] != "accept" {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid URL path format for accept request")
		return
	}
	requesterID := pathParts[2]
	if requesterID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Requester user ID cannot be empty")
		return
	}

	err = h.service.AcceptFollow(accepterID, requesterID)
	if err != nil {
		log.Printf("Error in HandleAcceptRequest service call: %v", err)
		if err.Error() == "no pending follow request found from this user" {
			writeErrorResponse(w, http.StatusNotFound, err.Error())
		} else {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to accept follow request")
		}
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{"message": "Follow request accepted"})
}

// HandleRejectRequest rejects or deletes a follow request/relationship
// DELETE /users/{requester_id}/reject
func (h *FollowerHandler) HandleRejectRequest(w http.ResponseWriter, r *http.Request) {
	rejecterID, err := h.getAuthenticatedUserID(r)
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) != 4 || pathParts[0] != "api" || pathParts[1] != "users" || pathParts[3] != "reject" {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid URL path format for reject request")
		return
	}
	requesterID := pathParts[2]
	if requesterID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Requester user ID cannot be empty")
		return
	}

	err = h.service.RejectFollow(rejecterID, requesterID)
	if err != nil {
		log.Printf("Error in HandleRejectRequest service call: %v", err)
		if err.Error() == "no follow request found from this user to reject" {
			writeErrorResponse(w, http.StatusNotFound, err.Error())
		} else {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to reject follow request")
		}
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{"message": "Follow request rejected"})
}

// HandleUnfollow stops following a user
// DELETE /users/{target_id}/unfollow
func (h *FollowerHandler) HandleUnfollow(w http.ResponseWriter, r *http.Request) {
	unfollowerID, err := h.getAuthenticatedUserID(r)
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}

	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) != 4 || pathParts[0] != "api" || pathParts[1] != "users" || pathParts[3] != "unfollow" {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid URL path format for unfollow request")
		return
	}
	targetID := pathParts[2]
	if targetID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Target user ID cannot be empty")
		return
	}

	err = h.service.Unfollow(unfollowerID, targetID)
	if err != nil {
		log.Printf("Error in HandleUnfollow service call: %v", err)
		if err.Error() == "not following this user" {
			writeErrorResponse(w, http.StatusNotFound, err.Error())
		} else {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to unfollow user")
		}
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]string{"message": "Successfully unfollowed user"})
}

// HandleListFollowers lists users following a given user
// GET /users/{user_id}/followers?limit=10&offset=0
func (h *FollowerHandler) HandleListFollowers(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) != 4 || pathParts[0] != "api" || pathParts[1] != "users" || pathParts[3] != "followers" {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid URL path format for list followers")
		return
	}
	userID := pathParts[2]
	if userID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "User ID cannot be empty")
		return
	}

	// Parse pagination parameters
	limit, offset := helpers.GetPaginationParams(r)

	// TODO: Update service call signature to accept limit, offset
	followers, err := h.service.ListFollowers(userID, limit, offset)
	if err != nil {
		log.Printf("Error in HandleListFollowers service call: %v", err)
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve followers")
		return
	}

	// Return paginated response
	response := map[string]interface{}{
		"followers": followers,
		"limit":     limit,
		"offset":    offset,
		"count":     len(followers), // Note: This is count of returned items, not total
	}
	writeJSONResponse(w, http.StatusOK, response)
}

// HandleListFollowing lists users a given user is following
// GET /users/{user_id}/following?limit=10&offset=0
func (h *FollowerHandler) HandleListFollowing(w http.ResponseWriter, r *http.Request) {
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) != 4 || pathParts[0] != "api" || pathParts[1] != "users" || pathParts[3] != "following" {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid URL path format for list following")
		return
	}
	userID := pathParts[2]
	if userID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "User ID cannot be empty")
		return
	}

	// Parse pagination parameters
	limit, offset := helpers.GetPaginationParams(r)

	// TODO: Update service call signature to accept limit, offset
	following, err := h.service.ListFollowing(userID, limit, offset)
	if err != nil {
		log.Printf("Error in HandleListFollowing service call: %v", err)
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve following list")
		return
	}

	// Return paginated response
	response := map[string]interface{}{
		"following": following,
		"limit":     limit,
		"offset":    offset,
		"count":     len(following), // Note: This is count of returned items, not total
	}
	writeJSONResponse(w, http.StatusOK, response)
}

// HandleListPending lists pending follow requests for the authenticated user
// GET /users/me/follow-requests
func (h *FollowerHandler) HandleListPending(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getAuthenticatedUserID(r)
	if err != nil {
		writeErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Service now returns a map {"received": [...], "sent": [...]}
	pendingRequestsMap, err := h.service.ListPendingRequests(userID)
	if err != nil {
		log.Printf("Error in HandleListPending service call: %v", err)
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve pending requests")
		return
	}

	// Ensure the map keys exist even if the lists are empty
	if _, ok := pendingRequestsMap["received"]; !ok {
		pendingRequestsMap["received"] = []models.User{}
	}
	if _, ok := pendingRequestsMap["sent"]; !ok {
		pendingRequestsMap["sent"] = []models.User{}
	}

	writeJSONResponse(w, http.StatusOK, pendingRequestsMap)
}
