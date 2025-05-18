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
	CreatorName    string `json:"creator_name,omitempty"`
	GroupName      string `json:"group_name,omitempty"`
	GroupAvatarURL string `json:"group_avatar_url,omitempty"`
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

// --- DTOs for Event Responses ---

// GroupEventResponseRequest is the DTO for submitting a response to an event
type GroupEventResponseRequest struct {
	Response string `json:"response" validate:"required,oneof=going not_going"`
}

// GroupEventResponseDetails is the DTO for listing event responses with user info
type GroupEventResponseDetails struct {
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	AvatarURL string    `json:"avatar_url"`
	Response  string    `json:"response"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GroupEventResponseCounts is the DTO for returning the counts of responses
type GroupEventResponseCounts struct {
	Going    int `json:"going"`
	NotGoing int `json:"not_going"`
}

// GroupEventDetailsResponse is the DTO for returning full event details including responses and counts
type GroupEventDetailsResponse struct {
	*GroupEventResponse                              // Embed basic event details
	Responses           []*GroupEventResponseDetails `json:"responses"`       // List of user responses
	ResponseCounts      *GroupEventResponseCounts    `json:"response_counts"` // Counts of responses
}

// GroupEventService defines the interface for group event business logic
type GroupEventService interface {
	Create(request *GroupEventCreateRequest) (*GroupEventResponse, error)
	GetByID(eventID string, requestingUserID string) (*GroupEventDetailsResponse, error) // Changed return type
	ListByGroupID(groupID string, limit, offset int, requestingUserID string) ([]*GroupEventResponse, error)
	Update(eventID string, request *GroupEventUpdateRequest, requestingUserID string) (*GroupEventResponse, error)
	Delete(eventID string, requestingUserID string) error

	// Event Response Methods
	RespondToEvent(eventID, userID string, request *GroupEventResponseRequest) error
	ListEventResponses(eventID, requestingUserID string) ([]*GroupEventResponseDetails, error)
	GetEventResponseCounts(eventID, requestingUserID string) (*GroupEventResponseCounts, error)
}

// groupEventService implements GroupEventService interface
type groupEventService struct {
	groupEventRepo         repositories.GroupEventRepository
	groupRepo              repositories.GroupRepository
	userRepo               repositories.UserRepository
	groupEventResponseRepo repositories.GroupEventResponseRepository // Added
	notificationService    NotificationService
}

// NewGroupEventService creates a new GroupEventService
func NewGroupEventService(
	groupEventRepo repositories.GroupEventRepository,
	groupRepo repositories.GroupRepository,
	userRepo repositories.UserRepository,
	groupEventResponseRepo repositories.GroupEventResponseRepository, // Added
	notificationService NotificationService,
) GroupEventService {
	return &groupEventService{
		groupEventRepo:         groupEventRepo,
		groupRepo:              groupRepo,
		userRepo:               userRepo,
		groupEventResponseRepo: groupEventResponseRepo, // Added
		notificationService:    notificationService,
	}
}

// --- Helper Mappers ---

// mapGroupEventToResponse converts models.GroupEvent to GroupEventResponse
func mapGroupEventToResponse(event *models.GroupEvent) *GroupEventResponse {
	if event == nil {
		return nil
	}
	response := &GroupEventResponse{
		ID:          event.ID,
		GroupID:     event.GroupID,
		CreatorID:   event.CreatorID,
		Title:       event.Title,
		Description: event.Description,
		EventTime:   event.EventTime,
		CreatedAt:   event.CreatedAt,
		UpdatedAt:   event.UpdatedAt,
	}
	if event.GroupAvatarURL.Valid {
		response.GroupAvatarURL = event.GroupAvatarURL.String
	}
	return response
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

	// Send notifications to group members
	if s.notificationService != nil {
		groupMembers, err := s.groupRepo.GetMembersByGroupID(request.GroupID)
		if err != nil {
			fmt.Printf("Warning: Failed to get group members for event notification (event %s, group %s): %v\n", event.ID, request.GroupID, err)
		} else {
			group, groupErr := s.groupRepo.GetByID(request.GroupID)
			if groupErr != nil {
				fmt.Printf("Warning: Failed to get group details for event notification (event %s, group %s): %v\n", event.ID, request.GroupID, groupErr)
			} else {
				for _, member := range groupMembers {
					// Don't notify the event creator
					if member.UserID == request.CreatorID {
						continue
					}
					message := fmt.Sprintf("A new event '%s' has been created in %s.", event.Title, group.Name)
					_, errNotif := s.notificationService.CreateNotification(
						nil, // Context
						member.UserID,
						models.GroupEventCreatedNotification,
						models.EventEntityType,
						message,
						event.ID,
					)
					if errNotif != nil {
						fmt.Printf("Warning: Failed to create group event created notification for member %s (event %s): %v\n", member.UserID, event.ID, errNotif)
					}
				}
			}
		}
	}

	return mapGroupEventToResponse(event), nil
}

// GetByID retrieves detailed group event information including responses and counts
func (s *groupEventService) GetByID(eventID string, requestingUserID string) (*GroupEventDetailsResponse, error) {
	// 1. Get event with enriched responses from repository
	eventAPI, err := s.groupEventRepo.GetEventWithResponsesByID(eventID)
	if err != nil {
		if errors.Is(err, repositories.ErrEventNotFound) {
			return nil, repositories.ErrEventNotFound
		}
		return nil, fmt.Errorf("failed to get event with responses: %w", err)
	}

	// Check if requesting user is a member of the group
	isMember, err := s.groupRepo.IsMember(eventAPI.GroupID, requestingUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check group membership: %w", err)
	}
	if !isMember {
		return nil, ErrGroupMemberRequired
	}

	// 2. Create base response DTO from eventAPI.GroupEvent
	baseResponse := mapGroupEventToResponse(&eventAPI.GroupEvent)

	// Add creator name if possible
	creator, err := s.userRepo.GetByID(eventAPI.CreatorID)
	if err == nil {
		baseResponse.CreatorName = creator.FirstName + " " + creator.LastName
	} else {
		fmt.Printf("Warning: Failed to get creator details for event %s: %v\n", eventID, err)
	}

	// Add group name if possible
	group, err := s.groupRepo.GetByID(eventAPI.GroupID)
	if err == nil {
		baseResponse.GroupName = group.Name
	} else {
		fmt.Printf("Warning: Failed to get group details for event %s: %v\n", eventID, err)
	}

	// 3. Map []models.EventResponseAPI to []*GroupEventResponseDetails
	mappedResponses := make([]*GroupEventResponseDetails, len(eventAPI.Responses))
	for i, repoResp := range eventAPI.Responses {
		var updatedAt time.Time
		// Attempt to parse the UpdatedAt string. Similar to repository logic for event_time.
		// Assuming RFC3339 or a common SQLite format.
		parsedTime, errParse := time.Parse(time.RFC3339, repoResp.UpdatedAt)
		if errParse != nil {
			parsedTimeFallback, errFallback := time.Parse("2006-01-02 15:04:05", repoResp.UpdatedAt) // Common SQLite format
			if errFallback != nil {
				// Further fallback for "YYYY-MM-DD HH:MM:SSZ" or other timezone variants if necessary
				parsedTimeTZ, errTZ := time.Parse("2006-01-02 15:04:05Z07:00", repoResp.UpdatedAt)
				if errTZ != nil {
					fmt.Printf("Warning: Failed to parse UpdatedAt timestamp '%s' for response by user %s (event %s): %v, %v, %v\n", repoResp.UpdatedAt, repoResp.UserID, eventID, errParse, errFallback, errTZ)
					updatedAt = time.Time{} // Default to zero time if parsing fails
				} else {
					updatedAt = parsedTimeTZ
				}
			} else {
				updatedAt = parsedTimeFallback
			}
		} else {
			updatedAt = parsedTime
		}

		mappedResponses[i] = &GroupEventResponseDetails{
			UserID:    repoResp.UserID,
			Username:  repoResp.Username,
			FirstName: repoResp.FirstName,
			LastName:  repoResp.LastName,
			AvatarURL: repoResp.AvatarURL,
			Response:  repoResp.Response,
			UpdatedAt: updatedAt,
		}
	}

	// 4. Fetch response counts
	counts, err := s.GetEventResponseCounts(eventID, requestingUserID)
	if err != nil {
		// Log the error but don't fail the whole request, return nil counts
		fmt.Printf("Warning: Failed to get event response counts for event %s: %v\n", eventID, err)
		counts = nil // Return nil counts if fetching failed
	}

	// 6. Assemble the detailed response
	detailedResponse := &GroupEventDetailsResponse{
		GroupEventResponse: baseResponse,
		Responses:          mappedResponses,
		ResponseCounts:     counts,
	}

	return detailedResponse, nil
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

// --- Event Response Method Implementations ---

// RespondToEvent handles a user responding to an event
func (s *groupEventService) RespondToEvent(eventID, userID string, request *GroupEventResponseRequest) error {
	// 1. Get the event to find the group ID
	event, err := s.groupEventRepo.GetByID(eventID)
	if err != nil {
		if errors.Is(err, repositories.ErrEventNotFound) {
			return repositories.ErrEventNotFound // Or a more specific "event not found" error
		}
		return fmt.Errorf("failed to get event details: %w", err)
	}

	// 2. Check if the responding user is a member of the group
	isMember, err := s.groupRepo.IsMember(event.GroupID, userID)
	if err != nil {
		return fmt.Errorf("failed to check group membership for response: %w", err)
	}
	if !isMember {
		return ErrGroupMemberRequired // User must be a member to respond
	}

	// 3. Create the response model
	responseModel := &models.GroupEventResponse{
		EventID:  eventID,
		UserID:   userID,
		Response: request.Response,
		// ID, CreatedAt, UpdatedAt are handled by the repository/DB
	}

	// 4. Call repository to create or update the response
	err = s.groupEventResponseRepo.CreateOrUpdate(responseModel)
	if err != nil {
		// Specific errors (like constraint violations) are already handled in the repo
		return fmt.Errorf("failed to save event response: %w", err)
	}

	return nil // Success
}

// ListEventResponses lists all responses for a given event, including usernames
func (s *groupEventService) ListEventResponses(eventID, requestingUserID string) ([]*GroupEventResponseDetails, error) {
	// 1. Get the event to find the group ID
	event, err := s.groupEventRepo.GetByID(eventID)
	if err != nil {
		if errors.Is(err, repositories.ErrEventNotFound) {
			return nil, repositories.ErrEventNotFound
		}
		return nil, fmt.Errorf("failed to get event details: %w", err)
	}

	// 2. Check if the requesting user is a member of the group
	isMember, err := s.groupRepo.IsMember(event.GroupID, requestingUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check group membership for listing responses: %w", err)
	}
	if !isMember {
		return nil, ErrGroupMemberRequired // User must be a member to see responses
	}

	// 3. Fetch all responses for the event
	rawResponses, err := s.groupEventResponseRepo.GetByEventID(eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event responses from repository: %w", err)
	}

	// 4. & 5. Fetch user details and map to DTOs
	responseDetails := make([]*GroupEventResponseDetails, 0, len(rawResponses))
	for _, resp := range rawResponses {
		user, err := s.userRepo.GetByID(resp.UserID)
		username := "Unknown User" // Default if user fetch fails
		if err == nil {
			username = user.Username // Using Username for now
		} else {
			// Log the error, but continue - maybe the user was deleted
			fmt.Printf("Warning: Failed to get user details for user ID %s: %v\n", resp.UserID, err)
		}

		details := &GroupEventResponseDetails{
			UserID:    resp.UserID,
			Username:  username,
			Response:  resp.Response,
			UpdatedAt: resp.UpdatedAt,
		}
		responseDetails = append(responseDetails, details)
	}

	return responseDetails, nil
}

// GetEventResponseCounts gets the counts of 'going' and 'not_going' responses for an event
func (s *groupEventService) GetEventResponseCounts(eventID, requestingUserID string) (*GroupEventResponseCounts, error) {
	// 1. Get the event to find the group ID
	event, err := s.groupEventRepo.GetByID(eventID)
	if err != nil {
		if errors.Is(err, repositories.ErrEventNotFound) {
			return nil, repositories.ErrEventNotFound
		}
		return nil, fmt.Errorf("failed to get event details: %w", err)
	}

	// 2. Check if the requesting user is a member of the group
	isMember, err := s.groupRepo.IsMember(event.GroupID, requestingUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check group membership for getting counts: %w", err)
	}
	if !isMember {
		return nil, ErrGroupMemberRequired // User must be a member to see counts
	}

	// 3. Fetch counts from the repository
	goingCount, notGoingCount, err := s.groupEventResponseRepo.GetCountsByEventID(eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event response counts from repository: %w", err)
	}

	// 4. Map to DTO
	counts := &GroupEventResponseCounts{
		Going:    goingCount,
		NotGoing: notGoingCount,
	}

	return counts, nil
}
