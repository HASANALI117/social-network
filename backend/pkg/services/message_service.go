package services

import (
	"fmt"

	"github.com/HASANALI117/social-network/pkg/models" // Ensure models.Message is imported
	"github.com/HASANALI117/social-network/pkg/repositories"
)

// TODO: Define MessageResponse DTOs if needed for transformation

// MessageService defines the interface for message-related business logic.
type MessageService interface {
	// GetDirectMessagesBetweenUsers retrieves paginated messages between two specified users.
	// It returns the list of messages, the total count of messages between them, and an error if any.
	GetDirectMessagesBetweenUsers(user1ID, user2ID string, limit, offset int) ([]models.Message, int64, error)
	GetGroupMessages(groupID string, limit, offset int, requestingUserID string) ([]*models.GroupMessage, error)
	// GetChatPartners retrieves a list of users with whom the current user has a chat history.
	GetChatPartners(currentUserID string) ([]models.ChatPartner, error)
	// TODO: Add methods for sending messages if needed (currently handled by websocket)
}

// messageService implements MessageService interface.
type messageService struct {
	messageRepo repositories.MessageRepository // Updated to use the new MessageRepository
	groupRepo   repositories.GroupRepository // Needed for authorization (e.g., checking group membership)
	// Add other dependencies like userRepo if needed
}

// NewMessageService creates a new MessageService.
func NewMessageService(messageRepo repositories.MessageRepository, groupRepo repositories.GroupRepository) MessageService { // Updated parameter type
	return &messageService{
		messageRepo: messageRepo,
		groupRepo:   groupRepo,
	}
}

// GetDirectMessagesBetweenUsers retrieves paginated direct messages between two users.
// The repository layer handles fetching messages where the pair are sender/receiver in either order.
// Authorization is implicitly handled by the handler ensuring the requestor is one of the users.
func (s *messageService) GetDirectMessagesBetweenUsers(user1ID, user2ID string, limit, offset int) ([]models.Message, int64, error) {
	// Call the repository method that fetches messages and total count
	messages, totalCount, err := s.messageRepo.GetDirectMessagesBetweenUsers(user1ID, user2ID, limit, offset)
	if err != nil {
		// It's often better to return the specific repository error or wrap it
		return nil, 0, fmt.Errorf("failed to get direct messages from repository: %w", err)
	}

	// TODO: Map to response DTOs if needed. Assuming models.Message is suitable for now.
	return messages, totalCount, nil
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
		return nil, ErrGroupMemberRequired // Assuming ErrGroupMemberRequired is defined elsewhere
	}

	messages, err := s.messageRepo.GetGroupMessages(groupID, limit, offset, requestingUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group messages from repository: %w", err)
	}

	// TODO: Map to response DTOs if needed
	return messages, nil
}

// Note: ErrGroupMemberRequired is likely defined in group_service.go or a shared errors package

// GetChatPartners retrieves a list of users with whom the current user has a chat history,
// sorted by the most recent message.
func (s *messageService) GetChatPartners(currentUserID string) ([]models.ChatPartner, error) {
	partners, err := s.messageRepo.GetChatPartners(currentUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat partners from repository: %w", err)
	}
	// TODO: Further transformations or checks if needed
	return partners, nil
}
