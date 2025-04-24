package services

import (
"errors"
"fmt"
"time"

"github.com/HASANALI117/social-network/pkg/models"
"github.com/HASANALI117/social-network/pkg/repositories"
)

// GroupResponse is the DTO for group data sent to clients
type GroupResponse struct {
ID          string    `json:"id"`
CreatorID   string    `json:"creator_id"`
Name        string    `json:"name"`
Description string    `json:"description,omitempty"`
AvatarURL   string    `json:"avatar_url,omitempty"`
CreatedAt   time.Time `json:"created_at"`
UpdatedAt   time.Time `json:"updated_at,omitempty"`
// TODO: Add member count? Creator details?
}

// GroupCreateRequest is the DTO for creating a new group
type GroupCreateRequest struct {
CreatorID   string `json:"-"` // Set internally from authenticated user
Name        string `json:"name" validate:"required,max=50"`
Description string `json:"description" validate:"max=255"`
AvatarURL   string `json:"avatar_url" validate:"omitempty,url"`
}

// GroupUpdateRequest is the DTO for updating a group
type GroupUpdateRequest struct {
Name        string `json:"name" validate:"required,max=50"`
Description string `json:"description" validate:"max=255"`
AvatarURL   string `json:"avatar_url" validate:"omitempty,url"`
}

// GroupMemberResponse is the DTO for group member data
type GroupMemberResponse struct {
ID        string `json:"id"`
Username  string `json:"username"`
FirstName string `json:"first_name,omitempty"`
LastName  string `json:"last_name,omitempty"`
AvatarURL string `json:"avatar_url,omitempty"`
Role      string `json:"role"` // Added role
JoinedAt  time.Time `json:"joined_at"` // Added joined_at
}

// TODO: GroupMessageResponse DTO

var (
ErrGroupForbidden = errors.New("user not authorized to perform this action on the group")
ErrGroupAdminRequired = errors.New("admin privileges required for this group action")
ErrGroupMemberRequired = errors.New("group membership required for this action")
ErrGroupCreatorCannotBeRemoved = errors.New("group creator cannot be removed")
)


// GroupService defines the interface for group business logic
type GroupService interface {
Create(request *GroupCreateRequest) (*GroupResponse, error)
GetByID(groupID string, requestingUserID string) (*GroupResponse, error) // Check if user can view
List(limit, offset int, requestingUserID string) ([]*GroupResponse, error) // List groups user can see/join?
Update(groupID string, request *GroupUpdateRequest, requestingUserID string) (*GroupResponse, error)
Delete(groupID string, requestingUserID string) error

// Member Management
AddMember(groupID, targetUserID, role string, requestingUserID string) error
RemoveMember(groupID, targetUserID string, requestingUserID string) error
ListMembers(groupID string, requestingUserID string) ([]*UserResponse, error) // Use UserResponse for consistency?

// TODO: Message Management
// GetMessages(groupID string, limit, offset int, requestingUserID string) ([]*GroupMessageResponse, error)
// AddMessage(...)
}

// groupService implements GroupService interface
type groupService struct {
groupRepo repositories.GroupRepository
userRepo  repositories.UserRepository // Needed for checking user existence
}

// NewGroupService creates a new GroupService
func NewGroupService(groupRepo repositories.GroupRepository, userRepo repositories.UserRepository) GroupService {
return &groupService{
groupRepo: groupRepo,
userRepo:  userRepo,
}
}

// --- Helper Mappers ---

func mapGroupToResponse(group *models.Group) *GroupResponse {
if group == nil {
return nil
}
return &GroupResponse{
ID:          group.ID,
CreatorID:   group.CreatorID,
Name:        group.Name,
Description: group.Description,
AvatarURL:   group.AvatarURL,
CreatedAt:   group.CreatedAt,
UpdatedAt:   group.UpdatedAt,
}
}

func mapGroupsToResponse(groups []*models.Group) []*GroupResponse {
responses := make([]*GroupResponse, len(groups))
for i, group := range groups {
responses[i] = mapGroupToResponse(group)
}
return responses
}

// mapUsersToUserResponse converts []*models.User to []*UserResponse
// Note: This duplicates the one in user_service, consider moving to a shared place or using UserService?
func mapUsersToUserResponse(users []*models.User) []*UserResponse {
responses := make([]*UserResponse, len(users))
for i, user := range users {
if user != nil {
responses[i] = &UserResponse{ // Assuming UserResponse exists and has these fields
ID:        user.ID,
Username:  user.Username,
Email:     user.Email, // Consider if email should be exposed here
FirstName: user.FirstName,
LastName:  user.LastName,
AvatarURL: user.AvatarURL,
AboutMe:   user.AboutMe,
CreatedAt: user.CreatedAt,
}
}
}
return responses
}


// --- Group CRUD ---

func (s *groupService) Create(request *GroupCreateRequest) (*GroupResponse, error) {
// TODO: Validation
if request.Name == "" {
return nil, errors.New("group name is required")
}

group := &models.Group{
CreatorID:   request.CreatorID, // Assumes CreatorID is set correctly
Name:        request.Name,
Description: request.Description,
AvatarURL:   request.AvatarURL,
}

err := s.groupRepo.Create(group) // Repository handles adding creator as admin member
if err != nil {
return nil, fmt.Errorf("failed to create group in repository: %w", err)
}

return mapGroupToResponse(group), nil
}

func (s *groupService) GetByID(groupID string, requestingUserID string) (*GroupResponse, error) {
group, err := s.groupRepo.GetByID(groupID)
if err != nil {
if errors.Is(err, repositories.ErrGroupNotFound) {
return nil, err
}
return nil, fmt.Errorf("failed to get group from repository: %w", err)
}

// TODO: Authorization Check - Can the requesting user view this group?
// For now, assume any logged-in user can get any group by ID.
// A better approach would be to check membership or if the group is public.
// isMember, _ := s.groupRepo.IsMember(groupID, requestingUserID)
// if !isMember {
//     return nil, ErrGroupMemberRequired // Or ErrGroupForbidden
// }

return mapGroupToResponse(group), nil
}

func (s *groupService) List(limit, offset int, requestingUserID string) ([]*GroupResponse, error) {
// TODO: Implement filtering based on user's memberships or public groups
groups, err := s.groupRepo.List(limit, offset)
if err != nil {
return nil, fmt.Errorf("failed to list groups from repository: %w", err)
}
return mapGroupsToResponse(groups), nil // Return unfiltered for now
}

func (s *groupService) Update(groupID string, request *GroupUpdateRequest, requestingUserID string) (*GroupResponse, error) {
// TODO: Validation
if request.Name == "" {
return nil, errors.New("group name is required")
}

// Authorization: Check if user is admin
isAdmin, err := s.groupRepo.IsAdmin(groupID, requestingUserID)
if err != nil {
return nil, fmt.Errorf("failed to check admin status for update: %w", err)
}
if !isAdmin {
return nil, ErrGroupAdminRequired
}

// Get existing group to update
group, err := s.groupRepo.GetByID(groupID)
if err != nil {
if errors.Is(err, repositories.ErrGroupNotFound) {
return nil, err // Group not found
}
return nil, fmt.Errorf("failed to get group for update: %w", err)
}

// Apply updates
group.Name = request.Name
group.Description = request.Description
group.AvatarURL = request.AvatarURL

// Save updated group
err = s.groupRepo.Update(group)
if err != nil {
return nil, fmt.Errorf("failed to update group in repository: %w", err)
}

return mapGroupToResponse(group), nil
}

func (s *groupService) Delete(groupID string, requestingUserID string) error {
// Get group to check ownership (creator)
group, err := s.groupRepo.GetByID(groupID)
if err != nil {
if errors.Is(err, repositories.ErrGroupNotFound) {
return err // Group not found
}
return fmt.Errorf("failed to get group for delete check: %w", err)
}

// Authorization: Only creator can delete
if group.CreatorID != requestingUserID {
return ErrGroupForbidden // Specific error for non-creator delete attempt
}

// Proceed with deletion
err = s.groupRepo.Delete(groupID)
if err != nil {
// Repo returns ErrGroupNotFound if already deleted
return fmt.Errorf("failed to delete group in repository: %w", err)
}

return nil
}

// --- Member Management ---

func (s *groupService) AddMember(groupID, targetUserID, role string, requestingUserID string) error {
// Authorization: Check if requesting user is admin
isAdmin, err := s.groupRepo.IsAdmin(groupID, requestingUserID)
if err != nil {
return fmt.Errorf("failed to check admin status for adding member: %w", err)
}
if !isAdmin {
return ErrGroupAdminRequired
}

// Check if target user exists
_, err = s.userRepo.GetByID(targetUserID)
if err != nil {
if errors.Is(err, repositories.ErrUserNotFound) {
return fmt.Errorf("cannot add member: target user not found")
}
return fmt.Errorf("failed to check target user existence: %w", err)
}

// Attempt to add member
err = s.groupRepo.AddMember(groupID, targetUserID, role)
if err != nil {
// Repo handles ErrAlreadyGroupMember and FK errors
return fmt.Errorf("failed to add member in repository: %w", err)
}

return nil
}

func (s *groupService) RemoveMember(groupID, targetUserID string, requestingUserID string) error {
// Get group info (needed for creator check)
group, err := s.groupRepo.GetByID(groupID)
if err != nil {
// Handle group not found specifically? Repo does this on RemoveMember too.
return fmt.Errorf("failed to get group info for member removal: %w", err)
}

// Prevent removing the creator
if group.CreatorID == targetUserID {
return ErrGroupCreatorCannotBeRemoved
}

// Authorization: Check if requesting user is admin OR if user is removing self
isAdmin, err := s.groupRepo.IsAdmin(groupID, requestingUserID)
if err != nil {
return fmt.Errorf("failed to check admin status for removing member: %w", err)
}

isRemovingSelf := requestingUserID == targetUserID

if !isAdmin && !isRemovingSelf {
return ErrGroupForbidden // Not admin and not removing self
}

// Attempt to remove member
err = s.groupRepo.RemoveMember(groupID, targetUserID)
if err != nil {
// Repo handles ErrNotGroupMember
return fmt.Errorf("failed to remove member in repository: %w", err)
}

return nil
}

func (s *groupService) ListMembers(groupID string, requestingUserID string) ([]*UserResponse, error) {
// Authorization: Check if requesting user is a member
isMember, err := s.groupRepo.IsMember(groupID, requestingUserID)
if err != nil {
return nil, fmt.Errorf("failed to check membership status for listing members: %w", err)
}
if !isMember {
return nil, ErrGroupMemberRequired
}

// Get members from repository
members, err := s.groupRepo.ListMembers(groupID)
if err != nil {
return nil, fmt.Errorf("failed to list members from repository: %w", err)
}

// Map to response DTO
// TODO: Enhance mapping if role/joined_at needed in response
return mapUsersToUserResponse(members), nil
}

// TODO: Implement Group Message Service methods
