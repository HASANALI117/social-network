package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/HASANALI117/social-network/pkg/helpers"
	"github.com/HASANALI117/social-network/pkg/httperr"
	"github.com/HASANALI117/social-network/pkg/models"
	"github.com/HASANALI117/social-network/pkg/services"
)

// NotificationHandler handles notification-related requests
type NotificationHandler struct {
	authService         services.AuthService
	notificationService services.NotificationService
}

// NewNotificationHandler creates a new NotificationHandler
func NewNotificationHandler(notificationService services.NotificationService, authService services.AuthService) *NotificationHandler {
	return &NotificationHandler{
		authService:         authService,
		notificationService: notificationService,
	}
}

// NotificationsResponse defines the structure for the notifications response including pagination
type NotificationsResponse struct {
	Notifications []models.Notification `json:"notifications"`
	Page          int                   `json:"page"`
	PageSize      int                   `json:"page_size"`
	UnreadCount   int                   `json:"unread_count"`
}

// ServeHTTP handles all notification-related requests
func (h *NotificationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	// Get current user from session for all endpoints
	currentUser, err := helpers.GetUserFromSession(r, h.authService)
	if err != nil {
		if errors.Is(err, helpers.ErrInvalidSession) {
			return httperr.NewUnauthorized(err, "Invalid session")
		}
		return httperr.NewInternalServerError(err, "Failed to get current user")
	}

	// Parse the path to determine the endpoint
	path := strings.TrimPrefix(r.URL.Path, "/api/notifications")
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")

	// Route to appropriate handler based on path and method
	switch {
	case path == "": // GET /api/notifications
		if r.Method != http.MethodGet {
			return httperr.NewMethodNotAllowed(nil, "Method not allowed for /api/notifications")
		}
		return h.getNotifications(w, r, currentUser)

	case len(parts) == 2 && parts[1] == "read": // PATCH /api/notifications/{id}/read
		if r.Method != http.MethodPatch {
			return httperr.NewMethodNotAllowed(nil, "Method not allowed for /api/notifications/{id}/read")
		}
		notificationID, err := strconv.Atoi(parts[0])
		if err != nil {
			return httperr.NewBadRequest(err, "Invalid notification ID format")
		}
		return h.markNotificationAsRead(w, r, notificationID, currentUser)

	case path == "read-all": // POST /api/notifications/read-all
		if r.Method != http.MethodPost {
			return httperr.NewMethodNotAllowed(nil, "Method not allowed for /api/notifications/read-all")
		}
		return h.markAllNotificationsAsRead(w, r, currentUser)

	default:
		return httperr.NewNotFound(nil, "Invalid notifications endpoint")
	}
}

// getNotifications handles GET /api/notifications
func (h *NotificationHandler) getNotifications(w http.ResponseWriter, r *http.Request, currentUser *services.UserResponse) error {
	// Parse pagination parameters
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
			page = parsedPage
		} else if err != nil {
			return httperr.NewBadRequest(err, "Invalid page parameter format")
		}
	}

	pageSize := 20
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if parsedPageSize, err := strconv.Atoi(pageSizeStr); err == nil && parsedPageSize > 0 {
			pageSize = parsedPageSize
		} else if err != nil {
			return httperr.NewBadRequest(err, "Invalid page_size parameter format")
		}
	}

	// Convert user ID from string to int
	userID, err := strconv.Atoi(currentUser.ID)
	if err != nil {
		return httperr.NewInternalServerError(err, "Failed to parse user ID")
	}

	// Get notifications using NotificationService
	notifications, err := h.notificationService.GetNotificationsForUser(userID, page, pageSize)
	if err != nil {
		return httperr.NewInternalServerError(err, "Failed to fetch notifications")
	}

	// Get unread count
	unreadCount, err := h.notificationService.GetUnreadNotificationCount(userID)
	if err != nil {
		return httperr.NewInternalServerError(err, "Failed to get unread notification count")
	}

	// Prepare response
	response := NotificationsResponse{
		Notifications: notifications,
		Page:          page,
		PageSize:      pageSize,
		UnreadCount:   unreadCount,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	return nil
}

// markNotificationAsRead handles PATCH /api/notifications/{id}/read
func (h *NotificationHandler) markNotificationAsRead(w http.ResponseWriter, r *http.Request, notificationID int, currentUser *services.UserResponse) error {
	userID, err := strconv.Atoi(currentUser.ID)
	if err != nil {
		return httperr.NewInternalServerError(err, "Failed to parse user ID")
	}

	if err := h.notificationService.MarkNotificationAsRead(notificationID, userID); err != nil {
		return httperr.NewInternalServerError(err, "Failed to mark notification as read")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Notification marked as read"})
	return nil
}

// markAllNotificationsAsRead handles POST /api/notifications/read-all
func (h *NotificationHandler) markAllNotificationsAsRead(w http.ResponseWriter, r *http.Request, currentUser *services.UserResponse) error {
	userID, err := strconv.Atoi(currentUser.ID)
	if err != nil {
		return httperr.NewInternalServerError(err, "Failed to parse user ID")
	}

	if err := h.notificationService.MarkAllNotificationsAsRead(userID); err != nil {
		return httperr.NewInternalServerError(err, "Failed to mark all notifications as read")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "All notifications marked as read"})
	return nil
}
