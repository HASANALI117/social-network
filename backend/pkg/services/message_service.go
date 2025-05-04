package services

import (
	"fmt"

	"github.com/HASANALI117/social-network/pkg/models"
	"github.com/HASANALI117/social-network/pkg/repositories"
)

// TODO: Define MessageResponse DTOs if needed for transformation

// MessageService defines the interface for message-related business logic.
type MessageService interface {
	GetUserMessages(senderID, receiverID string, limit, offset int, requestingUserID string) ([]*models.Message, error)
	GetGroupMessages(groupID string, limit, offset int, requestingUserID string) ([]*models.GroupMessage, error)
	// TODO: Add methods for sending messages if needed (currently handled by websocket)
}

// messageService implements MessageService interface.
type messageService struct {
	messageRepo repositories.ChatMessageRepository
	groupRepo   repositories.GroupRepository // Needed for authorization (e.g., checking group membership)
	// Add other dependencies like userRepo if needed
}

// NewMessageService creates a new MessageService.
func NewMessageService(messageRepo repositories.ChatMessageRepository, groupRepo repositories.GroupRepository) MessageService {
	return &messageService{
		messageRepo: messageRepo,
		groupRepo:   groupRepo,
	}
}

// GetUserMessages retrieves direct messages between two users.
// Authorization: Ensure the requesting user is one of the participants.
func (s *messageService) GetUserMessages(senderID, receiverID string, limit, offset int, requestingUserID string) ([]*models.Message, error) {
	// Authorization check
	if requestingUserID != senderID && requestingUserID != receiverID {
		// Using a standard error type might be better
		return nil, fmt.Errorf("user %s not authorized to view messages between %s and %s", requestingUserID, senderID, receiverID)
	}

	messages, err := s.messageRepo.GetUserMessages(senderID, receiverID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user messages from repository: %w", err)
	}

	// TODO: Map to response DTOs if needed
	return messages, nil
}

// GetGroupMessages retrieves messages for a specific group.
// Authorization: Ensure the requesting user is a member of the group.
func (s *messageService) GetGroupMessages(groupID string, limit, offset int, requestingUserID string) ([]*models.GroupMessage, error) {
	// Authorization check: Is user a member of the group?
	isMember, err := s.groupRepo.IsMember(groupID, requestingUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check group membership for user %s in group %s: %w", requestingUserID, groupID, err)
	}
	if !isMember {
		// Using the error from group service for consistency
		return nil, ErrGroupMemberRequired
	}

	messages, err := s.messageRepo.GetGroupMessages(groupID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get group messages from repository: %w", err)
	}

	// TODO: Map to response DTOs if needed
	return messages, nil
}
