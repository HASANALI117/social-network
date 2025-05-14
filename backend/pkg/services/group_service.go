package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/HASANALI117/social-network/pkg/models"
	"github.com/HASANALI117/social-network/pkg/repositories"
	"github.com/HASANALI117/social-network/pkg/types" // Added import
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
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	FirstName string    `json:"first_name,omitempty"`
	LastName  string    `json:"last_name,omitempty"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	Role      string    `json:"role"`      // Added role
	JoinedAt  time.Time `json:"joined_at"` // Added joined_at
}

// GroupInvitationResponse is the DTO for group invitation data
type GroupInvitationResponse struct {
	ID        string         `json:"id"`
	GroupID   string         `json:"group_id"`
	GroupName string         `json:"group_name,omitempty"` // Optional: Include group name
	InviterID string         `json:"inviter_id"`
	Inviter   *UserResponse  `json:"inviter,omitempty"` // Optional: Include inviter details
	InviteeID string         `json:"invitee_id"`
	Invitee   *UserResponse  `json:"invitee,omitempty"` // Optional: Include invitee details
	Status    string         `json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	Group     *GroupResponse `json:"group,omitempty"` // Optional: Include full group details
}

// GroupJoinRequestResponse is the DTO for group join request data
type GroupJoinRequestResponse struct {
	ID          string         `json:"id"`
	GroupID     string         `json:"group_id"`
	GroupName   string         `json:"group_name,omitempty"` // Optional: Include group name
	RequesterID string         `json:"requester_id"`
	Requester   *UserResponse  `json:"requester,omitempty"` // Optional: Include requester details
	Status      string         `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	Group       *GroupResponse `json:"group,omitempty"` // Optional: Include full group details
}

// GroupProfileResponse is the comprehensive DTO for a group's profile page
type GroupProfileResponse struct {
	ID             string          `json:"id"`
	Creator        *UserResponse   `json:"creator"` // Include creator details
	Name           string          `json:"name"`
	Description    string          `json:"description,omitempty"`
	AvatarURL      string          `json:"avatar_url,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at,omitempty"`
	Members        []*UserResponse `json:"members"`          // List of members with basic details
	MemberCount    int             `json:"member_count"`     // Total number of members
	ViewerIsMember bool            `json:"viewer_is_member"` // Is the requesting user a member?
	ViewerIsAdmin  bool            `json:"viewer_is_admin"`  // Is the requesting user an admin?
	// TODO: Add pending request/invitation counts if needed
}

// TODO: GroupMessageResponse DTO

var (
	ErrGroupForbidden              = errors.New("user not authorized to perform this action on the group")
	ErrGroupAdminRequired          = errors.New("admin privileges required for this group action")
	ErrGroupMemberRequired         = errors.New("group membership required for this action")
	ErrGroupCreatorCannotBeRemoved = errors.New("group creator cannot be removed")
	ErrInvalidInvitationStatus     = errors.New("invalid status transition for invitation")
	ErrInvalidJoinRequestStatus    = errors.New("invalid status transition for join request")
	ErrCannotInviteSelf            = errors.New("cannot invite yourself to a group")
	ErrCannotRequestToJoinOwnGroup = errors.New("cannot request to join a group you created")
	ErrNotInvited                  = errors.New("user was not invited to this group")
	ErrNoJoinRequestFound          = errors.New("no pending join request found for this user in this group")
	ErrNotGroupCreator             = errors.New("only the group creator can perform this action")
)

// GroupService defines the interface for group business logic
type GroupService interface {
	Create(request *GroupCreateRequest) (*GroupResponse, error)
	GetByID(groupID string, requestingUserID string) (*types.GroupDetailResponse, error)                       // Updated return type
	GetGroupProfile(groupID string, requestingUserID string) (*GroupProfileResponse, error)                    // Detailed profile view
	List(limit, offset int, searchQuery string, requestingUserID string) ([]*types.GroupDetailResponse, error) // Updated return type
	Update(groupID string, request *GroupUpdateRequest, requestingUserID string) (*GroupResponse, error)
	Delete(groupID string, requestingUserID string) error

	// Member Management (Revised)
	AddMember(groupID, targetUserID, role string, requestingUserID string) error  // Added back for direct admin addition
	RemoveMember(groupID, targetUserID string, requestingUserID string) error     // Kick member (admin) or leave group (self)
	ListMembers(groupID string, requestingUserID string) ([]*UserResponse, error) // Check membership before listing
	IsAdmin(groupID, userID string) (bool, error)                                 // Added
	IsMember(groupID, userID string) (bool, error)                                // Added

	// Invitation Management
	InviteUser(groupID, inviteeID string, inviterID string) (*GroupInvitationResponse, error)
	AcceptInvitation(invitationID string, userID string) error                // User accepting their own invite
	RejectInvitation(invitationID string, userID string) error                // User rejecting their own invite
	ListPendingInvitations(userID string) ([]*GroupInvitationResponse, error) // List invites received by the user
	// TODO: Maybe add CancelInvitation(invitationID, inviterID)?

	// Join Request Management
	RequestToJoin(groupID string, requesterID string) (*GroupJoinRequestResponse, error)
	AcceptJoinRequest(requestID string, adminUserID string) error                                    // Admin accepting a request
	RejectJoinRequest(requestID string, adminUserID string) error                                    // Admin rejecting a request
	ListPendingJoinRequests(groupID string, adminUserID string) ([]*GroupJoinRequestResponse, error) // List requests for admins
	// TODO: Maybe add CancelJoinRequest(requestID, requesterID)?

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

// mapInvitationToResponse converts models.GroupInvitation to GroupInvitationResponse
func mapInvitationToResponse(inv *models.GroupInvitation, includeDetails bool, userRepo repositories.UserRepository, groupRepo repositories.GroupRepository) *GroupInvitationResponse {
	if inv == nil {
		return nil
	}
	resp := &GroupInvitationResponse{
		ID:        inv.ID,
		GroupID:   inv.GroupID,
		InviterID: inv.InviterID,
		InviteeID: inv.InviteeID,
		Status:    inv.Status,
		CreatedAt: inv.CreatedAt,
		UpdatedAt: inv.UpdatedAt,
	}

	if includeDetails {
		// Fetch optional details - handle errors gracefully (e.g., log but don't fail)
		if group, err := groupRepo.GetByID(inv.GroupID); err == nil {
			resp.Group = mapGroupToResponse(group)
			resp.GroupName = group.Name
		} else {
			fmt.Printf("Warning: Failed to get group details for invitation %s: %v\n", inv.ID, err)
		}
		if inviter, err := userRepo.GetByID(inv.InviterID); err == nil {
			resp.Inviter = mapUserToResponse(inviter) // Assuming mapUserToResponse exists
		} else {
			fmt.Printf("Warning: Failed to get inviter details for invitation %s: %v\n", inv.ID, err)
		}
		if invitee, err := userRepo.GetByID(inv.InviteeID); err == nil {
			resp.Invitee = mapUserToResponse(invitee) // Assuming mapUserToResponse exists
		} else {
			fmt.Printf("Warning: Failed to get invitee details for invitation %s: %v\n", inv.ID, err)
		}
	}
	return resp
}

// mapInvitationsToResponse converts a slice of invitations
func mapInvitationsToResponse(invs []*models.GroupInvitation, includeDetails bool, userRepo repositories.UserRepository, groupRepo repositories.GroupRepository) []*GroupInvitationResponse {
	responses := make([]*GroupInvitationResponse, len(invs))
	for i, inv := range invs {
		responses[i] = mapInvitationToResponse(inv, includeDetails, userRepo, groupRepo)
	}
	return responses
}

// mapJoinRequestToResponse converts models.GroupJoinRequest to GroupJoinRequestResponse
func mapJoinRequestToResponse(req *models.GroupJoinRequest, includeDetails bool, userRepo repositories.UserRepository, groupRepo repositories.GroupRepository) *GroupJoinRequestResponse {
	if req == nil {
		return nil
	}
	resp := &GroupJoinRequestResponse{
		ID:          req.ID,
		GroupID:     req.GroupID,
		RequesterID: req.RequesterID,
		Status:      req.Status,
		CreatedAt:   req.CreatedAt,
		UpdatedAt:   req.UpdatedAt,
	}

	if includeDetails {
		// Fetch optional details - handle errors gracefully
		if group, err := groupRepo.GetByID(req.GroupID); err == nil {
			resp.Group = mapGroupToResponse(group)
			resp.GroupName = group.Name
		} else {
			fmt.Printf("Warning: Failed to get group details for join request %s: %v\n", req.ID, err)
		}
		if requester, err := userRepo.GetByID(req.RequesterID); err == nil {
			resp.Requester = mapUserToResponse(requester) // Assuming mapUserToResponse exists
		} else {
			fmt.Printf("Warning: Failed to get requester details for join request %s: %v\n", req.ID, err)
		}
	}
	return resp
}

// mapJoinRequestsToResponse converts a slice of join requests
func mapJoinRequestsToResponse(reqs []*models.GroupJoinRequest, includeDetails bool, userRepo repositories.UserRepository, groupRepo repositories.GroupRepository) []*GroupJoinRequestResponse {
	responses := make([]*GroupJoinRequestResponse, len(reqs))
	for i, req := range reqs {
		responses[i] = mapJoinRequestToResponse(req, includeDetails, userRepo, groupRepo)
	}
	return responses
}

// mapUserToResponse converts models.User to UserResponse (ensure this exists or is imported)
func mapUserToResponse(user *models.User) *UserResponse {
	if user == nil {
		return nil
	}
	return &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		AvatarURL: user.AvatarURL,
		AboutMe:   user.AboutMe,
		CreatedAt: user.CreatedAt,
		// Add other fields as needed, ensure consistency with UserResponse definition
	}
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

func (s *groupService) GetByID(groupID string, requestingUserID string) (*types.GroupDetailResponse, error) {
	groupDetail, err := s.groupRepo.GetGroupDetailsByID(groupID) // Use new method for detailed response
	if err != nil {
		if errors.Is(err, repositories.ErrGroupNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get group details from repository: %w", err)
	}

	// Populate CreatorInfo
	if groupDetail.CreatorInfo.UserID != "" {
		creator, err := s.userRepo.GetByID(groupDetail.CreatorInfo.UserID)
		if err != nil {
			fmt.Printf("Warning: Failed to get creator (ID: %s) details for group %s: %v\n", groupDetail.CreatorInfo.UserID, groupDetail.ID, err)
			groupDetail.CreatorInfo = types.UserBasicInfo{} // Clear if not found
		} else if creator != nil {
			groupDetail.CreatorInfo.FirstName = creator.FirstName
			groupDetail.CreatorInfo.LastName = creator.LastName
			groupDetail.CreatorInfo.Username = creator.Username
			groupDetail.CreatorInfo.AvatarURL = creator.AvatarURL
		}
	}

	isMember, err := s.IsMember(groupID, requestingUserID)
	if err != nil {
		// Log error but proceed, as non-members can still view basic info
		fmt.Printf("Warning: Failed to check membership for group %s, user %s: %v\n", groupID, requestingUserID, err)
		// Treat as non-member if error occurs during check, or decide if this should be a hard error
		isMember = false
	}

	if isMember {
		// User is a member, return full details
		// Potentially, more details could be added here if GroupDetailResponse had member-specific fields
		// that are not fetched by default by groupRepo.GetByID
		return groupDetail, nil
	} else {
		// User is NOT a member, return limited information
		// The groupDetail already contains the necessary counts and basic info from the repository
		// We just need to ensure no sensitive member-only data is accidentally included if it were part of GroupDetailResponse
		// For now, GroupDetailResponse is structured to be suitable for both, with service controlling population.
		// If GroupDetailResponse had fields like "DetailedMemberActivity", we would explicitly nullify them here.
		// The current structure of GroupDetailResponse (ID, Name, Description, ImageURL, CreatorInfo, Counts, Timestamps)
		// is generally safe for non-members.
		return &types.GroupDetailResponse{
			ID:           groupDetail.ID,
			Name:         groupDetail.Name,
			Description:  groupDetail.Description,
			AvatarURL:    groupDetail.AvatarURL,
			CreatorInfo:  groupDetail.CreatorInfo, // Already populated
			MembersCount: groupDetail.MembersCount,
			PostsCount:   groupDetail.PostsCount,
			EventsCount:  groupDetail.EventsCount,
			CreatedAt:    groupDetail.CreatedAt,
			UpdatedAt:    groupDetail.UpdatedAt,
			// Any fields specific to members would be omitted here or set to nil/empty
		}, nil
	}
}

func (s *groupService) List(limit, offset int, searchQuery string, requestingUserID string) ([]*types.GroupDetailResponse, error) {
	// TODO: Implement filtering based on user's memberships or public groups
	groupDetails, err := s.groupRepo.List(limit, offset, searchQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to list groups from repository: %w", err)
	}

	// Populate CreatorInfo for each group
	for _, groupDetail := range groupDetails {
		if groupDetail.CreatorInfo.UserID != "" {
			creator, err := s.userRepo.GetByID(groupDetail.CreatorInfo.UserID)
			if err != nil {
				// Log error but don't fail the entire list if a creator isn't found
				// This could happen if a user account was deleted but groups remain
				fmt.Printf("Warning: Failed to get creator (ID: %s) details for group %s: %v\n", groupDetail.CreatorInfo.UserID, groupDetail.ID, err)
				// Optionally, clear or set a default for CreatorInfo
				groupDetail.CreatorInfo = types.UserBasicInfo{} // Clear if not found
			} else if creator != nil {
				groupDetail.CreatorInfo.FirstName = creator.FirstName
				groupDetail.CreatorInfo.LastName = creator.LastName
				groupDetail.CreatorInfo.Username = creator.Username
				groupDetail.CreatorInfo.AvatarURL = creator.AvatarURL
			}
		}
	}

	return groupDetails, nil
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

func (s *groupService) GetGroupProfile(groupID string, requestingUserID string) (*GroupProfileResponse, error) {
	// 1. Get basic group info
	group, err := s.groupRepo.GetByID(groupID)
	if err != nil {
		if errors.Is(err, repositories.ErrGroupNotFound) {
			return nil, err // Propagate not found error
		}
		return nil, fmt.Errorf("failed to get group for profile: %w", err)
	}

	// 2. Authorization: Check if the requesting user is a member (required to view profile)
	//    Alternatively, allow public viewing if group is public (not implemented yet)
	isMember, err := s.groupRepo.IsMember(groupID, requestingUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership for profile view: %w", err)
	}
	if !isMember {
		// If group was public, we might allow viewing here. For now, require membership.
		return nil, ErrGroupMemberRequired
	}

	// 3. Get Creator Details
	creator, err := s.userRepo.GetByID(group.CreatorID)
	if err != nil {
		// Log error but don't fail the whole request if creator not found (might be deleted user)
		fmt.Printf("Warning: Failed to get creator details for group profile %s: %v\n", groupID, err)
		// Optionally return a placeholder or nil creator
	}

	// 4. Get Members List
	members, err := s.groupRepo.ListMembers(groupID) // This returns []*models.User
	if err != nil {
		return nil, fmt.Errorf("failed to list members for profile: %w", err)
	}
	memberResponses := mapUsersToUserResponse(members) // Convert to []*UserResponse

	// 5. Check if viewer is admin
	isAdmin, err := s.groupRepo.IsAdmin(groupID, requestingUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check admin status for profile view: %w", err)
	}

	// 6. Construct the response DTO
	profileResponse := &GroupProfileResponse{
		ID:             group.ID,
		Creator:        mapUserToResponse(creator), // Map creator model to response DTO
		Name:           group.Name,
		Description:    group.Description,
		AvatarURL:      group.AvatarURL,
		CreatedAt:      group.CreatedAt,
		UpdatedAt:      group.UpdatedAt,
		Members:        memberResponses,
		MemberCount:    len(memberResponses),
		ViewerIsMember: isMember, // We already checked this
		ViewerIsAdmin:  isAdmin,
	}

	return profileResponse, nil
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

// IsAdmin checks if a user is an admin of a specific group.
func (s *groupService) IsAdmin(groupID, userID string) (bool, error) {
	isAdmin, err := s.groupRepo.IsAdmin(groupID, userID)
	if err != nil {
		// The repo method already wraps errors, so just return it.
		return false, fmt.Errorf("failed to check group admin status: %w", err)
	}
	return isAdmin, nil
}

// IsMember checks if a user is a member of a specific group.
func (s *groupService) IsMember(groupID, userID string) (bool, error) {
	isMember, err := s.groupRepo.IsMember(groupID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to check group membership status: %w", err)
	}
	return isMember, nil
}

// TODO: Implement Group Message Service methods

// --- Invitation Management ---

func (s *groupService) InviteUser(groupID, inviteeID string, inviterID string) (*GroupInvitationResponse, error) {
	// 1. Authorization: Check if inviter is a member (or admin?)
	isMember, err := s.groupRepo.IsMember(groupID, inviterID)
	if err != nil {
		return nil, fmt.Errorf("failed to check inviter membership: %w", err)
	}
	if !isMember {
		return nil, ErrGroupMemberRequired // Or ErrGroupAdminRequired if only admins can invite
	}

	// 2. Validation: Check if invitee exists
	_, err = s.userRepo.GetByID(inviteeID)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, fmt.Errorf("cannot invite: target user not found")
		}
		return nil, fmt.Errorf("failed to check target user existence: %w", err)
	}

	// 3. Validation: Cannot invite self
	if inviteeID == inviterID {
		return nil, ErrCannotInviteSelf
	}

	// 4. Validation: Check if invitee is already a member
	isAlreadyMember, err := s.groupRepo.IsMember(groupID, inviteeID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if invitee is already member: %w", err)
	}
	if isAlreadyMember {
		return nil, repositories.ErrAlreadyGroupMember
	}

	// 5. Create Invitation
	invitation := &models.GroupInvitation{
		GroupID:   groupID,
		InviterID: inviterID,
		InviteeID: inviteeID,
		// Status, CreatedAt, UpdatedAt set by repo
	}

	err = s.groupRepo.CreateInvitation(invitation)
	if err != nil {
		// Repo handles ErrAlreadyInvited and FK errors
		return nil, fmt.Errorf("failed to create invitation in repository: %w", err)
	}

	// 6. Map and return response (without details initially)
	return mapInvitationToResponse(invitation, false, s.userRepo, s.groupRepo), nil
}

func (s *groupService) AcceptInvitation(invitationID string, userID string) error {
	// 1. Get Invitation
	inv, err := s.groupRepo.GetInvitationByID(invitationID)
	if err != nil {
		return fmt.Errorf("failed to get invitation: %w", err) // Handles ErrInvitationNotFound
	}

	// 2. Authorization: Check if the user accepting is the invitee
	if inv.InviteeID != userID {
		return ErrGroupForbidden // Not the intended recipient
	}

	// 3. Validation: Check if invitation is pending
	if inv.Status != "pending" {
		return ErrInvalidInvitationStatus // Already accepted/rejected
	}

	// 4. Update invitation status
	err = s.groupRepo.UpdateInvitationStatus(invitationID, "accepted")
	if err != nil {
		// If update fails, return error
		return fmt.Errorf("failed to update invitation status: %w", err)
	}

	// 5. Add user as member (Best effort after status update)
	err = s.groupRepo.AddMember(inv.GroupID, inv.InviteeID, "member")
	if err != nil {
		// If adding member fails, log it but don't necessarily revert status.
		// The user might already be a member due to a race condition, which is acceptable.
		// Or another error occurred.
		fmt.Printf("Warning: Failed to add member %s to group %s after accepting invite %s: %v\n", inv.InviteeID, inv.GroupID, invitationID, err)
		// If it's specifically ErrAlreadyGroupMember, it's fine.
		if !errors.Is(err, repositories.ErrAlreadyGroupMember) {
			// For other errors, we might consider trying to revert the status, but for now, just log.
			// return fmt.Errorf("failed to add member after accepting invitation: %w", err) // Optionally return error
		}
	}

	// TODO: Send notification?
	return nil
}

func (s *groupService) RejectInvitation(invitationID string, userID string) error {
	// 1. Get Invitation
	inv, err := s.groupRepo.GetInvitationByID(invitationID)
	if err != nil {
		return fmt.Errorf("failed to get invitation: %w", err) // Handles ErrInvitationNotFound
	}

	// 2. Authorization: Check if the user rejecting is the invitee
	if inv.InviteeID != userID {
		return ErrGroupForbidden // Not the intended recipient
	}

	// 3. Validation: Check if invitation is pending
	if inv.Status != "pending" {
		return ErrInvalidInvitationStatus // Already accepted/rejected
	}

	// 4. Update invitation status
	err = s.groupRepo.UpdateInvitationStatus(invitationID, "rejected")
	if err != nil {
		return fmt.Errorf("failed to update invitation status: %w", err)
	}

	// TODO: Send notification?
	return nil
}

func (s *groupService) ListPendingInvitations(userID string) ([]*GroupInvitationResponse, error) {
	invitations, err := s.groupRepo.ListPendingInvitationsForUser(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending invitations from repository: %w", err)
	}

	// Map to response DTO, include details like Group and Inviter
	return mapInvitationsToResponse(invitations, true, s.userRepo, s.groupRepo), nil
}

// --- Join Request Management ---

func (s *groupService) RequestToJoin(groupID string, requesterID string) (*GroupJoinRequestResponse, error) {
	// 1. Validation: Check if group exists
	group, err := s.groupRepo.GetByID(groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group for join request: %w", err) // Handles ErrGroupNotFound
	}

	// 2. Validation: Cannot request to join own group (creator is already admin)
	if group.CreatorID == requesterID {
		return nil, ErrCannotRequestToJoinOwnGroup
	}

	// 3. Validation: Check if requester exists (redundant if requesterID comes from auth)
	_, err = s.userRepo.GetByID(requesterID)
	if err != nil {
		if errors.Is(err, repositories.ErrUserNotFound) {
			return nil, fmt.Errorf("cannot request join: requester user not found")
		}
		return nil, fmt.Errorf("failed to check requester user existence: %w", err)
	}

	// 4. Validation: Check if requester is already a member
	isAlreadyMember, err := s.groupRepo.IsMember(groupID, requesterID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if requester is already member: %w", err)
	}
	if isAlreadyMember {
		return nil, repositories.ErrAlreadyGroupMember
	}

	// 5. Create Join Request
	request := &models.GroupJoinRequest{
		GroupID:     groupID,
		RequesterID: requesterID,
		// Status, CreatedAt, UpdatedAt set by repo
	}

	err = s.groupRepo.CreateJoinRequest(request)
	if err != nil {
		// Repo handles ErrAlreadyRequested and FK errors
		return nil, fmt.Errorf("failed to create join request in repository: %w", err)
	}

	// 6. Map and return response (without details initially)
	return mapJoinRequestToResponse(request, false, s.userRepo, s.groupRepo), nil
}

func (s *groupService) AcceptJoinRequest(requestID string, adminUserID string) error {
	// 1. Get Join Request
	req, err := s.groupRepo.GetJoinRequestByID(requestID)
	if err != nil {
		return fmt.Errorf("failed to get join request: %w", err) // Handles ErrJoinRequestNotFound
	}

	// 2. Authorization: Check if user accepting is an admin of the group
	isAdmin, err := s.groupRepo.IsAdmin(req.GroupID, adminUserID)
	if err != nil {
		return fmt.Errorf("failed to check admin status for accepting request: %w", err)
	}
	if !isAdmin {
		return ErrGroupAdminRequired
	}

	// 3. Validation: Check if request is pending
	if req.Status != "pending" {
		return ErrInvalidJoinRequestStatus // Already accepted/rejected
	}

	// 4. Update request status
	err = s.groupRepo.UpdateJoinRequestStatus(requestID, "accepted")
	if err != nil {
		return fmt.Errorf("failed to update join request status: %w", err)
	}

	// 5. Add user as member (Best effort)
	err = s.groupRepo.AddMember(req.GroupID, req.RequesterID, "member")
	if err != nil {
		fmt.Printf("Warning: Failed to add member %s to group %s after accepting join request %s: %v\n", req.RequesterID, req.GroupID, requestID, err)
		if !errors.Is(err, repositories.ErrAlreadyGroupMember) {
			// Log or handle other errors if necessary
		}
	}

	// TODO: Send notification?
	return nil
}

func (s *groupService) RejectJoinRequest(requestID string, adminUserID string) error {
	// 1. Get Join Request
	req, err := s.groupRepo.GetJoinRequestByID(requestID)
	if err != nil {
		return fmt.Errorf("failed to get join request: %w", err) // Handles ErrJoinRequestNotFound
	}

	// 2. Authorization: Check if user rejecting is an admin of the group
	isAdmin, err := s.groupRepo.IsAdmin(req.GroupID, adminUserID)
	if err != nil {
		return fmt.Errorf("failed to check admin status for rejecting request: %w", err)
	}
	if !isAdmin {
		return ErrGroupAdminRequired
	}

	// 3. Validation: Check if request is pending
	if req.Status != "pending" {
		return ErrInvalidJoinRequestStatus // Already accepted/rejected
	}

	// 4. Update request status
	err = s.groupRepo.UpdateJoinRequestStatus(requestID, "rejected")
	if err != nil {
		return fmt.Errorf("failed to update join request status: %w", err)
	}

	// TODO: Send notification?
	return nil
}

func (s *groupService) ListPendingJoinRequests(groupID string, adminUserID string) ([]*GroupJoinRequestResponse, error) {
	// 1. Authorization: Check if user listing is an admin of the group
	isAdmin, err := s.groupRepo.IsAdmin(groupID, adminUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to check admin status for listing requests: %w", err)
	}
	if !isAdmin {
		return nil, ErrGroupAdminRequired
	}

	// 2. Get pending requests from repository
	requests, err := s.groupRepo.ListPendingJoinRequestsForGroup(groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending join requests from repository: %w", err)
	}

	// 3. Map to response DTO, include details like Requester
	return mapJoinRequestsToResponse(requests, true, s.userRepo, s.groupRepo), nil
}
