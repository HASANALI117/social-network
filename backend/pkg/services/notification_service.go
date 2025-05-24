package services

import (
	"context"
	"log"

	"time" // Added for time.RFC3339

	"github.com/HASANALI117/social-network/pkg/models"
	"github.com/HASANALI117/social-network/pkg/repositories"
	// "github.com/HASANALI117/social-network/pkg/websocket" // Removed to break import cycle
)

// RealTimeNotifier defines an interface for sending real-time messages.
// This helps to decouple the NotificationService from the concrete WebSocket Hub implementation.
type RealTimeNotifier interface {
	NotifyUser(userID string, payload interface{}) error
}

type NotificationService interface {
	CreateNotification(ctx context.Context, userID string, notificationType models.NotificationType, entityType models.EntityType, message string, entityID string) (*models.Notification, error)
	GetUserNotifications(ctx context.Context, userID string, limit, offset int) ([]*models.Notification, error)
	MarkNotificationAsRead(ctx context.Context, notificationID string, userID string) error
	MarkAllUserNotificationsAsRead(ctx context.Context, userID string) error
	GetUnreadNotificationCount(ctx context.Context, userID string) (int, error)
	SendNotificationToUser(userID string, notification *models.Notification) error
}

type notificationService struct {
	repo     repositories.NotificationRepository
	notifier RealTimeNotifier // Use the interface
}

func NewNotificationService(repo repositories.NotificationRepository, notifier RealTimeNotifier) NotificationService {
	return &notificationService{repo: repo, notifier: notifier}
}

func (s *notificationService) CreateNotification(ctx context.Context, userID string, notificationType models.NotificationType, entityType models.EntityType, message string, entityID string) (*models.Notification, error) {
	notification := &models.Notification{
		UserID:     userID,
		Type:       notificationType,
		EntityType: entityType,
		Message:    message,
		EntityID:   entityID,
	}

	err := s.repo.Create(ctx, notification) // Pass the context here
	if err != nil {
		log.Printf("Error creating notification in service: %v", err)
		return nil, err
	}

	// After successful creation, send real-time notification
	if s.notifier != nil { // Check if notifier is available
		go func() { // Send in a goroutine to not block the main flow
			// Construct the payload as defined in the design document for SendNotificationToUser
			payload := map[string]interface{}{
				"type": "new_notification",
				"payload": map[string]interface{}{
					"id":          notification.ID,
					"user_id":     notification.UserID,
					"type":        notification.Type,
					"entity_type": notification.EntityType,
					"message":     notification.Message,
					"entity_id":   notification.EntityID,
					"is_read":     notification.IsRead,
					"created_at":  notification.CreatedAt.Format(time.RFC3339),
				},
			}
			err := s.notifier.NotifyUser(userID, payload) // Changed from recipientID
			if err != nil {
				log.Printf("Error sending real-time notification to user %s: %v", userID, err)
			}
		}()
	} else {
		log.Printf("RealTimeNotifier is not initialized, cannot send real-time notification for user %s", userID)
	}

	return notification, nil
}

func (s *notificationService) GetUserNotifications(ctx context.Context, userID string, limit, offset int) ([]*models.Notification, error) {
	return s.repo.GetByUserID(ctx, userID, limit, offset)
}

func (s *notificationService) MarkNotificationAsRead(ctx context.Context, notificationID string, userID string) error {
	return s.repo.MarkAsRead(ctx, notificationID, userID)
}

func (s *notificationService) MarkAllUserNotificationsAsRead(ctx context.Context, userID string) error {
	return s.repo.MarkAllAsRead(ctx, userID)
}

func (s *notificationService) GetUnreadNotificationCount(ctx context.Context, userID string) (int, error) {
	return s.repo.GetUnreadCount(ctx, userID)
}

// SendNotificationToUser is now effectively handled by the notifier.NotifyUser call within CreateNotification.
// If direct access to send arbitrary notification payloads is still needed on the service,
// this method could be adapted to use s.notifier.NotifyUser(userID, somePayload).
// For now, the primary path is via CreateNotification.
// We keep the method signature in the interface as per the design doc,
// but its direct implementation here might become simpler or be primarily for testing/specific scenarios.
func (s *notificationService) SendNotificationToUser(userID string, notification *models.Notification) error {
	if s.notifier == nil {
		log.Printf("RealTimeNotifier is not initialized. Cannot send notification to user %s.", userID)
		return nil // Or an error
	}
	payload := map[string]interface{}{
		"type": "new_notification", // This structure should match what the frontend expects
		"payload": map[string]interface{}{
			"id":          notification.ID,
			"user_id":     notification.UserID,
			"type":        notification.Type,
			"entity_type": notification.EntityType,
			"message":     notification.Message,
			"entity_id":   notification.EntityID,
			"is_read":     notification.IsRead,
			"created_at":  notification.CreatedAt.Format(time.RFC3339),
		},
	}
	log.Printf("Attempting to send explicit notification to user %s via Notifier: %+v", userID, payload)
	return s.notifier.NotifyUser(userID, payload)
}