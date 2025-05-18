package handlers

import (
	// "context" // No longer directly needed as r.Context() is used
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/HASANALI117/social-network/pkg/helpers"
	"github.com/HASANALI117/social-network/pkg/httperr"
	"github.com/HASANALI117/social-network/pkg/models" // Added import for models.Notification
	"github.com/HASANALI117/social-network/pkg/services"
)

type NotificationHandler struct {
	service     services.NotificationService
	authService services.AuthService
}

func NewNotificationHandler(service services.NotificationService, authService services.AuthService) *NotificationHandler {
	return &NotificationHandler{service: service, authService: authService}
}

// ServeHTTP handles incoming HTTP requests for notifications.
// It routes requests based on the path and method.
// GET /api/notifications - List notifications
// POST /api/notifications/{notificationId}/read - Mark a notification as read
// POST /api/notifications/read-all - Mark all notifications as read
func (h *NotificationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	user, err := helpers.GetUserFromSession(r, h.authService)
	if err != nil {
		// If GetUserFromSession returns an HTTPError, use it, otherwise wrap.
		if httpErr, ok := err.(*httperr.HTTPError); ok {
			return httpErr
		}
		return httperr.NewUnauthorized(err, "Authentication required.")
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/notifications")
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")

	// Log request details for debugging
	// log.Printf("NotificationHandler: Method=%s, Path=%s, Parts=%v", r.Method, r.URL.Path, parts)

	switch r.Method {
	case http.MethodGet:
		if len(parts) == 1 && parts[0] == "" { // Path is /api/notifications
			return h.listNotifications(w, r, user.ID)
		} else {
			return httperr.NewNotFound(nil, "Notification endpoint not found.")
		}
	case http.MethodPost:
		if len(parts) == 1 && parts[0] == "read-all" { // Path is /api/notifications/read-all
			return h.markAllAsRead(w, r, user.ID)
		} else if len(parts) == 2 && parts[1] == "read" { // Path is /api/notifications/{notificationId}/read
			notificationID := parts[0]
			return h.markAsRead(w, r, user.ID, notificationID)
		} else {
			return httperr.NewNotFound(nil, "Notification action not found.")
		}
	default:
		return httperr.NewMethodNotAllowed(nil, "")
	}
	// return nil // Should be unreachable if all paths return
}

func (h *NotificationHandler) listNotifications(w http.ResponseWriter, r *http.Request, userID string) error {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20 // Default limit
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0 // Default offset
	}

	notifications, err := h.service.GetUserNotifications(r.Context(), userID, limit, offset)
	if err != nil {
		log.Printf("Error getting notifications for user %s: %v", userID, err)
		return httperr.NewInternalServerError(err, "Failed to retrieve notifications.")
	}

	unreadCount, err := h.service.GetUnreadNotificationCount(r.Context(), userID)
	if err != nil {
		log.Printf("Error getting unread notification count for user %s: %v", userID, err)
		// Continue without unread count if it fails, or handle error differently
		// For now, we'll return an error if this crucial part fails.
		return httperr.NewInternalServerError(err, "Failed to retrieve unread notification count.")
	}

	response := struct {
		Notifications []*models.Notification `json:"notifications"` // Changed to models.Notification
		UnreadCount   int                    `json:"unread_count"`
		Limit         int                    `json:"limit"`
		Offset        int                    `json:"offset"`
		HasMore       bool                   `json:"has_more"`
	}{
		Notifications: notifications,
		UnreadCount:   unreadCount,
		Limit:         limit,
		Offset:        offset,
		HasMore:       len(notifications) == limit,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding notifications response: %v", err)
		return httperr.NewInternalServerError(err, "Failed to encode response.")
	}
	return nil
}

func (h *NotificationHandler) markAsRead(w http.ResponseWriter, r *http.Request, userID string, notificationID string) error {
	if notificationID == "" {
		return httperr.NewBadRequest(nil, "Notification ID is required.")
	}

	err := h.service.MarkNotificationAsRead(r.Context(), notificationID, userID)
	if err != nil {
		log.Printf("Error marking notification %s as read for user %s: %v", notificationID, userID, err)
		// Consider specific errors, e.g., if notification not found or not owned by user
		return httperr.NewInternalServerError(err, "Failed to mark notification as read.")
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *NotificationHandler) markAllAsRead(w http.ResponseWriter, r *http.Request, userID string) error {
	err := h.service.MarkAllUserNotificationsAsRead(r.Context(), userID)
	if err != nil {
		log.Printf("Error marking all notifications as read for user %s: %v", userID, err)
		return httperr.NewInternalServerError(err, "Failed to mark all notifications as read.")
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}