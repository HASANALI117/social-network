package handlers

import (
	"encoding/json"
	"net/http"

	"social-network/pkg/helpers"
)

// ListNotifications godoc
// @Summary List user notifications
// @Description Get a list of notifications for a user
// @Tags notifications
// @Accept json
// @Produce json
// @Param user_id query string true "User ID"
// @Success 200 {object} []models.Notification
// @Failure 400 {string} string "User ID required"
// @Failure 500 {string} string "Failed to list notifications"
// @Router /notifications/list [get]
func ListNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	notifications, err := helpers.ListNotifications(userID)
	if err != nil {
		http.Error(w, "Failed to list notifications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
}
