package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/HASANALI117/social-network/pkg/models"
	"github.com/HASANALI117/social-network/pkg/repositories"
	"github.com/google/uuid" // Added for UUID generation
)

// GroupEventResponse is the DTO for group event data sent to clients
type GroupEventResponse struct {
	ID          string    `json:"id"`       // Changed from int to string
	GroupID     string    `json:"group_id"` // Changed from int to string
	CreatorID   string    `json:"creator_id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	EventTime   time.Time `json:"event_time"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Optional related data
	CreatorName string `json:"creator_name,omitempty"`
	GroupName   string `json:"group_name,omitempty"`
}

// GroupEventCreateRequest is the DTO for creating a new group event
type GroupEventCreateRequest struct {
	GroupID     string    `json:"group_id" validate:"required"` // Changed from int to string
	CreatorID   string    `json:"-"`                            // Set internally from authenticated user
	Title       string    `json:"title" validate:"required"`
	Description string    `json:"description"`
	EventTime   time.Time `json:"event_time" validate:"required"`
}

// GroupEventUpdateRequest is the DTO for updating a group event
type GroupEventUpdateRequest struct {
	Title       string    `json:"title" validate:"required"`
	Description string    `json:"description"`
	EventTime   time.Time `json:"event_time" validate:"required"`
}

// GroupEventService defines the interface for group event business logic
type GroupEventService interface {
	Create(request *GroupEventCreateRequest) (*GroupEventResponse, error)
	GetByID(eventID string, requestingUserID string) (*GroupEventResponse, error)
	ListByGroupID(groupID string, limit, offset int, requestingUserID string) ([]*GroupEventResponse, error)
	Update(eventID string, request *GroupEventUpdateRequest, requestingUserID string) (*GroupEventResponse, error)
	Delete(eventID string, requestingUserID string) error
}

// groupEventService implements GroupEventService interface
type groupEventService struct {
	groupEventRepo repositories.GroupEventRepository
	groupRepo      repositories.GroupRepository
	userRepo       repositories.UserRepository
}

// NewGroupEventService creates a new GroupEventService
func NewGroupEventService(
	groupEventRepo repositories.GroupEventRepository,
	groupRepo repositories.GroupRepository,
	userRepo repositories.UserRepository,
) GroupEventService {
	return &groupEventService{
		groupEventRepo: groupEventRepo,
		groupRepo:      groupRepo,
		userRepo:       userRepo,
	}
}

// --- Helper Mappers ---

// mapGroupEventToResponse converts models.GroupEvent to GroupEventResponse
func mapGroupEventToResponse(event *models.GroupEvent) *GroupEventResponse {
	if event == nil {
		return nil
	}
	return &GroupEventResponse{
		ID:          event.ID,
		GroupID:     event.GroupID,
		CreatorID:   event.CreatorID,
		Title:       event.Title,
		Description: event.Description,
		EventTime:   event.EventTime,
		CreatedAt:   event.CreatedAt,
		UpdatedAt:   event.UpdatedAt,
	}
}

// mapGroupEventsToResponse converts a slice of models.GroupEvent to a slice of GroupEventResponse
func mapGroupEventsToResponse(events []*models.GroupEvent) []*GroupEventResponse {
	responses := make([]*GroupEventResponse, len(events))
	for i, event := range events {
		responses[i] = mapGroupEventToResponse(event)
	}
	return responses
}

// --- Business Logic Implementation ---

// Create handles creating a new group event
func (s *groupEventService) Create(request *GroupEventCreateRequest) (*GroupEventResponse, error) {
	// Validate request fields
	if request.Title == "" {
		return nil, errors.New("event title is required")
	}
	if request.EventTime.IsZero() {
		return nil, errors.New("event time is required")
	}

	// Check if user is a member of the group
	isMember, err := s.groupRepo.IsMember(request.GroupID, request.CreatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to check group membership: %w", err)
	}
	if !isMember {
		return nil, ErrGroupMemberRequired
	}

	// Create event model
	event := &models.GroupEvent{
		GroupID:     request.GroupID,
		CreatorID:   request.CreatorID,
		Title:       request.Title,
		Description: request.Description,
		EventTime:   request.EventTime,
	}

	// Generate UUID for the new event
	event.ID = uuid.New().String()

	// Create through repository
	err = s.groupEventRepo.Create(event)
	if err != nil {
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	return mapGroupEventToResponse(event), nil
}

// GetByID retrieves a group event by its ID
func (s *groupEventService) GetByID(eventID string, requestingUserID string) (*GroupEventResponse, error) {
	// Get event from repository
	event, err := s.groupEventRepo.GetByID(eventID)
	if err != nil {
		if errors.Is(err, repositories.ErrEventNotFound) {
			return nil, repositories.ErrEventNotFound
		}
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	// Check if requesting user is a member of the group
	isMember, err := s.groupRepo.IsMember(event.GroupID, requestingUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check group membership: %w", err)
	}
	if !isMember {
		return nil, ErrGroupMemberRequired
	}

	// Create response with additional details
	response := mapGroupEventToResponse(event)

	// Add creator name if possible
	creator, err := s.userRepo.GetByID(event.CreatorID)
	if err == nil {
		response.CreatorName = creator.FirstName + " " + creator.LastName
	}

	// Add group name if possible
	group, err := s.groupRepo.GetByID(event.GroupID)
	if err == nil {
		response.GroupName = group.Name
	}

	return response, nil
}

// ListByGroupID lists all events for a group
func (s *groupEventService) ListByGroupID(groupID string, limit, offset int, requestingUserID string) ([]*GroupEventResponse, error) {
	// Check if requesting user is a member of the group
	isMember, err := s.groupRepo.IsMember(groupID, requestingUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check group membership: %w", err)
	}
	if !isMember {
		return nil, ErrGroupMemberRequired
	}

	// Apply pagination defaults if needed
	if limit <= 0 {
		limit = 20 // Default limit
	}
	if offset < 0 {
		offset = 0 // Default offset
	}

	// Get events from repository
	events, err := s.groupEventRepo.ListByGroupID(groupID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	// Create response
	responses := mapGroupEventsToResponse(events)

	// Optionally add creator names (could be optimized with a batch query)
	for _, resp := range responses {
		creator, err := s.userRepo.GetByID(resp.CreatorID)
		if err == nil {
			resp.CreatorName = creator.FirstName + " " + creator.LastName
		}
	}

	return responses, nil
}

// Update updates an existing group event
func (s *groupEventService) Update(eventID string, request *GroupEventUpdateRequest, requestingUserID string) (*GroupEventResponse, error) {
	// Get existing event to verify ownership
	event, err := s.groupEventRepo.GetByID(eventID)
	if err != nil {
		if errors.Is(err, repositories.ErrEventNotFound) {
			return nil, repositories.ErrEventNotFound
		}
		return nil, fmt.Errorf("failed to get event: %w", err)
	}

	// Verify the requesting user is the creator
	if event.CreatorID != requestingUserID {
		return nil, repositories.ErrEventCreatorRequired
	}

	// Check user is still a member of the group
	isMember, err := s.groupRepo.IsMember(event.GroupID, requestingUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check group membership: %w", err)
	}
	if !isMember {
		return nil, ErrGroupMemberRequired
	}

	// Update fields
	event.Title = request.Title
	event.Description = request.Description
	event.EventTime = request.EventTime

	// Save changes
	err = s.groupEventRepo.Update(event)
	if err != nil {
		return nil, fmt.Errorf("failed to update event: %w", err)
	}

	return mapGroupEventToResponse(event), nil
}

// Delete removes a group event
func (s *groupEventService) Delete(eventID string, requestingUserID string) error {
	// Get existing event to verify ownership
	event, err := s.groupEventRepo.GetByID(eventID)
	if err != nil {
		if errors.Is(err, repositories.ErrEventNotFound) {
			return repositories.ErrEventNotFound
		}
		return fmt.Errorf("failed to get event: %w", err)
	}

	// Check if requesting user is the creator or an admin
	if event.CreatorID != requestingUserID {
		// Check if user is a group admin
		isAdmin, err := s.groupRepo.IsAdmin(event.GroupID, requestingUserID)
		if err != nil {
			return fmt.Errorf("failed to check admin status: %w", err)
		}
		if !isAdmin {
			return repositories.ErrEventCreatorRequired
		}
	}

	// Delete the event
	err = s.groupEventRepo.Delete(eventID)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	return nil
}
