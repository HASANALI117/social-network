package services

import (
	"fmt"
	"strconv"

	"github.com/HASANALI117/social-network/pkg/models"
	"github.com/HASANALI117/social-network/pkg/repositories"
)

// NotificationService defines the interface for notification business logic.
type NotificationService interface {
	// Core notification operations
	CreateNotification(notification *models.Notification) error
	GetNotificationsForUser(userID int, page, pageSize int) ([]models.Notification, error)
	MarkNotificationAsRead(notificationID int, userID int) error
	MarkAllNotificationsAsRead(userID int) error
	GetUnreadNotificationCount(userID int) (int, error)

	// Specific notification creation methods for different events
	NotifyFollowRequest(targetUserID, actorID int) error
	NotifyGroupInvite(targetUserID, actorID, groupID int) error
	NotifyGroupJoinRequest(groupCreatorID, actorID, groupID int) error
	NotifyNewGroupEvent(groupID, eventID int) error
}

// notificationService implements NotificationService.
type notificationService struct {
	notificationRepo repositories.NotificationRepository
	groupService     GroupService
}

// NewNotificationService creates a new instance of NotificationService.
func NewNotificationService(repo repositories.NotificationRepository, groupService GroupService) NotificationService {
	return &notificationService{
		notificationRepo: repo,
		groupService:     groupService,
	}
}

// CreateNotification creates a new notification.
func (s *notificationService) CreateNotification(notification *models.Notification) error {
	return s.notificationRepo.Create(notification)
}

// GetNotificationsForUser retrieves paginated notifications for a user.
func (s *notificationService) GetNotificationsForUser(userID int, page, pageSize int) ([]models.Notification, error) {
	offset := (page - 1) * pageSize
	return s.notificationRepo.GetByUserID(userID, pageSize, offset)
}

// MarkNotificationAsRead marks a single notification as read.
func (s *notificationService) MarkNotificationAsRead(notificationID int, userID int) error {
	return s.notificationRepo.MarkAsRead(notificationID, userID)
}

// MarkAllNotificationsAsRead marks all notifications for a user as read.
func (s *notificationService) MarkAllNotificationsAsRead(userID int) error {
	return s.notificationRepo.MarkAllAsRead(userID)
}

// GetUnreadNotificationCount gets the count of unread notifications.
func (s *notificationService) GetUnreadNotificationCount(userID int) (int, error) {
	return s.notificationRepo.GetUnreadCount(userID)
}

// NotifyFollowRequest creates a follow request notification.
func (s *notificationService) NotifyFollowRequest(targetUserID, actorID int) error {
	notification := &models.Notification{
		UserID:  targetUserID,
		ActorID: &actorID,
		Type:    models.NotificationTypeFollowRequest,
		// EntityID and EntityType not needed for follow requests
		IsRead: false,
	}
	return s.notificationRepo.Create(notification)
}

// NotifyGroupInvite creates a group invitation notification.
func (s *notificationService) NotifyGroupInvite(targetUserID, actorID, groupID int) error {
	entityType := models.EntityTypeGroup
	notification := &models.Notification{
		UserID:     targetUserID,
		ActorID:    &actorID,
		Type:       models.NotificationTypeGroupInvite,
		EntityID:   &groupID,
		EntityType: &entityType,
		IsRead:     false,
	}
	return s.notificationRepo.Create(notification)
}

// NotifyGroupJoinRequest creates a group join request notification.
func (s *notificationService) NotifyGroupJoinRequest(groupCreatorID, actorID, groupID int) error {
	entityType := models.EntityTypeGroup
	notification := &models.Notification{
		UserID:     groupCreatorID,
		ActorID:    &actorID,
		Type:       models.NotificationTypeGroupJoinRequest,
		EntityID:   &groupID,
		EntityType: &entityType,
		IsRead:     false,
	}
	return s.notificationRepo.Create(notification)
}

// NotifyNewGroupEvent creates notifications for all group members when a new event is created.
func (s *notificationService) NotifyNewGroupEvent(groupID, eventID int) error {
	if groupID <= 0 || eventID <= 0 {
		return fmt.Errorf("invalid group ID or event ID")
	}

	// Get the group profile to get member information
	groupProfile, err := s.groupService.GetGroupProfile(fmt.Sprintf("%d", groupID), "system") // Use system as requesting user to bypass restrictions
	if err != nil {
		return fmt.Errorf("failed to get group members: %w", err)
	}

	// Create notification for each member
	entityType := models.EntityTypeEvent
	for _, member := range groupProfile.Members {
		memberID, err := strconv.Atoi(member.ID)
		if err != nil {
			continue // Skip if ID conversion fails
		}

		notification := &models.Notification{
			UserID:     memberID,
			Type:       models.NotificationTypeNewGroupEvent,
			EntityID:   &eventID,
			EntityType: &entityType,
			IsRead:     false,
		}

		if err := s.CreateNotification(notification); err != nil {
			// Log error but continue with other members
			fmt.Printf("Failed to create event notification for member %d: %v\n", memberID, err)
		}
	}

	return nil
}
