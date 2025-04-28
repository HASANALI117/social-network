package handlers

import (
	"encoding/json"
	"errors"
	"fmt" // Import fmt
	"log" // Import log
	"net/http"
	"strconv"
	"strings" // Import strings

	"github.com/HASANALI117/social-network/pkg/helpers"
	"github.com/HASANALI117/social-network/pkg/httperr"

	// "github.com/HASANALI117/social-network/pkg/models" // Removed unused import
	"github.com/HASANALI117/social-network/pkg/repositories" // For error checking
	"github.com/HASANALI117/social-network/pkg/services"
)

// InviteUserRequest defines the structure for inviting a user
type InviteUserRequest struct {
	InviteeID string `json:"invitee_id" validate:"required"`
}

// RequestToJoinRequest defines the structure for requesting to join (empty body, groupID in path)

// RespondToInvitationRequest defines the structure for accepting/rejecting an invitation
type RespondToInvitationRequest struct {
	// No body needed, action determined by endpoint, invitationID in path
}

// RespondToJoinRequest defines the structure for accepting/rejecting a join request
type RespondToJoinRequest struct {
	// No body needed, action determined by endpoint, requestID in path
}

// GroupHandler handles group and group membership/message related HTTP requests
type GroupHandler struct {
	groupService services.GroupService
	authService  services.AuthService
}

// NewGroupHandler creates a new GroupHandler
func NewGroupHandler(groupService services.GroupService, authService services.AuthService) *GroupHandler {
	return &GroupHandler{
		groupService: groupService,
		authService:  authService,
	}
}

// ServeHTTP routes the request to the appropriate handler method based on path and method
func (h *GroupHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	// Get current user for authorization
	currentUser, err := helpers.GetUserFromSession(r, h.authService)
	// Most group actions require authentication. GET requests might be allowed for public info later.
	if err != nil {
		if errors.Is(err, helpers.ErrInvalidSession) {
			return httperr.NewUnauthorized(err, "Invalid session")
		}
		return httperr.NewInternalServerError(err, "Failed to get current user")
	}
	if currentUser == nil { // Should not happen if GetUserFromSession is correct, but double-check
		return httperr.NewUnauthorized(nil, "Authentication required")
	}

	// Path Routing Logic: /api/groups/{groupID}/{subResource}/{subID}
	path := strings.TrimPrefix(r.URL.Path, "/api/groups")
	path = strings.TrimPrefix(path, "/")
	parts := strings.Split(path, "/")

	log.Printf("GroupHandler: Method=%s, Path=%s, Parts=%v\n", r.Method, r.URL.Path, parts)

	// No group ID specified OR special top-level routes
	if len(parts) == 1 && parts[0] == "" { // Matches "/api/groups" or "/api/groups/"
		switch r.Method {
		case http.MethodPost: // POST /api/groups -> Create Group
			return h.createGroup(w, r, currentUser)
		case http.MethodGet: // GET /api/groups -> List Groups
			return h.listGroups(w, r, currentUser)
		default:
			return httperr.NewMethodNotAllowed(nil, "Method not allowed for /api/groups")
		}
	}

	// Handle /api/groups/invitations/... and /api/groups/requests/... separately
	if len(parts) >= 2 && parts[0] == "invitations" {
		// /api/groups/invitations/pending
		if len(parts) == 2 && parts[1] == "pending" && r.Method == http.MethodGet {
			return h.listPendingInvitations(w, r, currentUser)
		}
		// /api/groups/invitations/{invitationID}/accept
		if len(parts) == 3 && parts[2] == "accept" && r.Method == http.MethodPost {
			invitationID := parts[1]
			return h.acceptGroupInvitation(w, r, invitationID, currentUser)
		}
		// /api/groups/invitations/{invitationID}/reject
		if len(parts) == 3 && parts[2] == "reject" && r.Method == http.MethodPost {
			invitationID := parts[1]
			return h.rejectGroupInvitation(w, r, invitationID, currentUser)
		}
		return httperr.NewNotFound(nil, "Invalid invitation path")
	}

	if len(parts) >= 2 && parts[0] == "requests" {
		// /api/groups/requests/{requestID}/accept
		if len(parts) == 3 && parts[2] == "accept" && r.Method == http.MethodPost {
			requestID := parts[1]
			return h.acceptGroupJoinRequest(w, r, requestID, currentUser)
		}
		// /api/groups/requests/{requestID}/reject
		if len(parts) == 3 && parts[2] == "reject" && r.Method == http.MethodPost {
			requestID := parts[1]
			return h.rejectGroupJoinRequest(w, r, requestID, currentUser)
		}
		return httperr.NewNotFound(nil, "Invalid join request path")
	}

	// Group ID specified: /api/groups/{groupID}... (Must be after top-level checks)
	if len(parts) >= 1 && parts[0] != "" && parts[0] != "invitations" && parts[0] != "requests" {
		groupID := parts[0]

		// No sub-resource: /api/groups/{groupID}
		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet: // GET /api/groups/{groupID} -> Get Group Details
				return h.getGroupByID(w, r, groupID, currentUser)
			case http.MethodPut: // PUT /api/groups/{groupID} -> Update Group Details
				return h.updateGroup(w, r, groupID, currentUser)
			case http.MethodDelete: // DELETE /api/groups/{groupID} -> Delete Group
				return h.deleteGroup(w, r, groupID, currentUser)
			default:
				return httperr.NewMethodNotAllowed(nil, fmt.Sprintf("Method not allowed for /api/groups/%s", groupID))
			}
		}

		// Sub-resource specified: /api/groups/{groupID}/{subResource}...
		if len(parts) >= 2 {
			subResource := parts[1]

			switch subResource {
			case "invitations":
				// POST /api/groups/{groupID}/invitations -> Invite User
				if len(parts) == 2 && r.Method == http.MethodPost {
					return h.inviteUserToGroup(w, r, groupID, currentUser)
				}
				return httperr.NewMethodNotAllowed(nil, fmt.Sprintf("Method not allowed for /api/groups/%s/invitations", groupID))

			case "requests":
				// POST /api/groups/{groupID}/requests -> Request to Join
				if len(parts) == 2 && r.Method == http.MethodPost {
					return h.requestToJoinGroup(w, r, groupID, currentUser)
				}
				// GET /api/groups/{groupID}/requests/pending -> List Pending Join Requests for Group
				if len(parts) == 3 && parts[2] == "pending" && r.Method == http.MethodGet {
					return h.listPendingJoinRequests(w, r, groupID, currentUser)
				}
				return httperr.NewMethodNotAllowed(nil, fmt.Sprintf("Method not allowed for /api/groups/%s/requests", groupID))

			case "members":
				// /api/groups/{groupID}/members
				if len(parts) == 2 {
					switch r.Method {
					// case http.MethodPost: // POST /api/groups/{groupID}/members -> Add Member (Replaced by Invite/Request)
					// return h.addMember(w, r, groupID, currentUser)
					case http.MethodGet: // GET /api/groups/{groupID}/members -> List Members
						return h.listMembers(w, r, groupID, currentUser)
					default:
						return httperr.NewMethodNotAllowed(nil, fmt.Sprintf("Method not allowed for /api/groups/%s/members", groupID))
					}
				}
				// /api/groups/{groupID}/members/{userID}
				if len(parts) == 3 && parts[2] != "" {
					targetUserID := parts[2]
					switch r.Method {
					case http.MethodDelete: // DELETE /api/groups/{groupID}/members/{userID} -> Remove Member
						return h.removeMember(w, r, groupID, targetUserID, currentUser)
					default:
						return httperr.NewMethodNotAllowed(nil, fmt.Sprintf("Method not allowed for /api/groups/%s/members/%s", groupID, targetUserID))
					}
				}

			case "messages":
				// /api/groups/{groupID}/messages
				if len(parts) == 2 {
					switch r.Method {
					case http.MethodGet: // GET /api/groups/{groupID}/messages -> List Messages
						return h.getGroupMessages(w, r, groupID, currentUser) // Stubbed
					// case http.MethodPost: // POST /api/groups/{groupID}/messages -> Send Message (TODO)
					default:
						return httperr.NewMethodNotAllowed(nil, fmt.Sprintf("Method not allowed for /api/groups/%s/messages", groupID))
					}
				}

			default:
				return httperr.NewNotFound(nil, "Invalid group sub-resource")
			}
		}
	}

	return httperr.NewNotFound(nil, "Invalid group API path")
}

// --- Handler Methods ---

// createGroup handles POST /api/groups/
// @Summary Create a new group
// @Description Create a new group chat
// @Tags groups
// @Accept json
// @Produce json
// @Param group body models.Group true "Group creation details"
// @Success 201 {object} map[string]interface{} "Group created successfully"
// @Failure 400 {string} string "Invalid request body"
// @Failure 400 {object} httperr.ErrorResponse "Invalid request body or Group name is required"
// @Failure 401 {object} httperr.ErrorResponse "Unauthorized"
// @Failure 405 {object} httperr.ErrorResponse "Method not allowed"
// @Failure 500 {object} httperr.ErrorResponse "Failed to create group"
// @Router /groups [post]
func (h *GroupHandler) createGroup(w http.ResponseWriter, r *http.Request, currentUser *services.UserResponse) error {
	var req services.GroupCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return httperr.NewBadRequest(err, "Invalid request body")
	}

	req.CreatorID = currentUser.ID // Set creator from authenticated user

	groupResponse, err := h.groupService.Create(&req)
	if err != nil {
		// TODO: Handle specific validation errors
		return httperr.NewInternalServerError(err, "Failed to create group")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(groupResponse)
	return nil
}

// getGroupByID handles GET /api/groups/{id}
// @Summary Get group by ID
// @Description Get group details by group ID
// @Tags groups
// @Accept json
// @Produce json
// @Param id query string true "Group ID"
// @Success 200 {object} map[string]interface{} "Group details"
// @Failure 400 {object} httperr.ErrorResponse "Group ID is required"
// @Failure 404 {object} httperr.ErrorResponse "Group not found"
// @Failure 405 {object} httperr.ErrorResponse "Method not allowed"
// @Failure 403 {object} httperr.ErrorResponse "Forbidden"
// @Failure 500 {object} httperr.ErrorResponse "Failed to get group"
// @Router /groups/{id} [get]
func (h *GroupHandler) getGroupByID(w http.ResponseWriter, r *http.Request, groupID string, currentUser *services.UserResponse) error {
	groupResponse, err := h.groupService.GetByID(groupID, currentUser.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrGroupNotFound) {
			return httperr.NewNotFound(err, "Group not found")
		}
		if errors.Is(err, services.ErrGroupMemberRequired) || errors.Is(err, services.ErrGroupForbidden) {
			// Hide existence if not a member/allowed
			return httperr.NewNotFound(err, "Group not found")
			// Or return 403: return httperr.NewForbidden(err, "Access denied to this group")
		}
		return httperr.NewInternalServerError(err, "Failed to get group details")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groupResponse)
	return nil
}

// listGroups handles GET /api/groups/
// @Summary List groups
// @Description Get a paginated list of groups
// @Tags groups
// @Accept json
// @Produce json
// @Param limit query int false "Number of groups to return (default 10)"
// @Param offset query int false "Number of groups to skip (default 0)"
// @Success 200 {object} map[string]interface{} "List of groups"
// @Failure 405 {object} httperr.ErrorResponse "Method not allowed"
// @Failure 500 {object} httperr.ErrorResponse "Failed to list groups"
// @Router /groups [get]
func (h *GroupHandler) listGroups(w http.ResponseWriter, r *http.Request, currentUser *services.UserResponse) error {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20 // Default limit
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0 // Default offset
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	groupsResponse, err := h.groupService.List(limit, offset, currentUser.ID)
	if err != nil {
		return httperr.NewInternalServerError(err, "Failed to list groups")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"groups": groupsResponse,
		"limit":  limit,
		"offset": offset,
		"count":  len(groupsResponse),
	})
	return nil
}

// updateGroup handles PUT /api/groups/{id}
// @Summary Update group
// @Description Update group details
// @Tags groups
// @Accept json
// @Produce json
// @Param id query string true "Group ID"
// @Param group body object true "Group update details"
// @Success 200 {object} map[string]interface{} "Updated group details"
// @Failure 400 {object} httperr.ErrorResponse "Invalid request or Group ID is required"
// @Failure 401 {object} httperr.ErrorResponse "Unauthorized"
// @Failure 404 {object} httperr.ErrorResponse "Group not found"
// @Failure 405 {object} httperr.ErrorResponse "Method not allowed"
// @Failure 403 {object} httperr.ErrorResponse "Forbidden (not admin)"
// @Failure 500 {object} httperr.ErrorResponse "Failed to update group"
// @Router /groups/{id} [put]
func (h *GroupHandler) updateGroup(w http.ResponseWriter, r *http.Request, groupID string, currentUser *services.UserResponse) error {
	var req services.GroupUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return httperr.NewBadRequest(err, "Invalid request body")
	}

	groupResponse, err := h.groupService.Update(groupID, &req, currentUser.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrGroupNotFound) {
			return httperr.NewNotFound(err, "Group not found")
		}
		if errors.Is(err, services.ErrGroupAdminRequired) {
			return httperr.NewForbidden(err, "Only group admins can update the group")
		}
		// TODO: Handle validation errors
		return httperr.NewInternalServerError(err, "Failed to update group")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groupResponse)
	return nil
}

// deleteGroup handles DELETE /api/groups/{id}
// @Summary Delete group
// @Description Delete a group by ID
// @Tags groups
// @Accept json
// @Produce json
// @Param id query string true "Group ID"
// @Success 200 {object} map[string]string "Group deleted successfully"
// @Failure 400 {object} httperr.ErrorResponse "Group ID is required"
// @Failure 401 {object} httperr.ErrorResponse "Unauthorized"
// @Failure 404 {object} httperr.ErrorResponse "Group not found"
// @Failure 405 {object} httperr.ErrorResponse "Method not allowed"
// @Failure 403 {object} httperr.ErrorResponse "Forbidden (not creator)"
// @Failure 500 {object} httperr.ErrorResponse "Failed to delete group"
// @Router /groups/{id} [delete]
func (h *GroupHandler) deleteGroup(w http.ResponseWriter, r *http.Request, groupID string, currentUser *services.UserResponse) error {
	err := h.groupService.Delete(groupID, currentUser.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrGroupNotFound) {
			return httperr.NewNotFound(err, "Group not found")
		}
		if errors.Is(err, services.ErrGroupForbidden) {
			return httperr.NewForbidden(err, "Only the group creator can delete the group")
		}
		return httperr.NewInternalServerError(err, "Failed to delete group")
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Group deleted successfully"})
	return nil
}

// --- Member Handlers (Merged) ---

// removeMember handles DELETE /api/groups/{id}/members/{userID}
// @Summary Remove a member from a group
// @Description Remove a user from a group (admin can remove anyone except creator, user can remove self)
// @Tags groups
// @Accept json
// @Produce json
// @Param id path string true "Group ID"
// @Param userID path string true "User ID to remove"
// @Success 200 {object} map[string]string "Member removed successfully"
// @Failure 403 {object} httperr.ErrorResponse "Forbidden (cannot remove creator or not authorized)"
// @Failure 404 {object} httperr.ErrorResponse "Group or user not found / User not in group"
// @Failure 500 {object} httperr.ErrorResponse "Failed to remove member"
// @Router /groups/{id}/members/{userID} [delete]
func (h *GroupHandler) removeMember(w http.ResponseWriter, r *http.Request, groupID, targetUserID string, currentUser *services.UserResponse) error {
	err := h.groupService.RemoveMember(groupID, targetUserID, currentUser.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrGroupNotFound) || errors.Is(err, repositories.ErrUserNotFound) {
			return httperr.NewNotFound(err, "Group or user not found")
		}
		if errors.Is(err, repositories.ErrNotGroupMember) {
			return httperr.NewNotFound(err, "User is not a member of this group")
		}
		if errors.Is(err, services.ErrGroupCreatorCannotBeRemoved) {
			return httperr.NewForbidden(err, "The group creator cannot be removed")
		}
		if errors.Is(err, services.ErrGroupForbidden) {
			return httperr.NewForbidden(err, "Only admins can remove other members (or you can remove yourself)")
		}
		return httperr.NewInternalServerError(err, "Failed to remove group member")
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Member removed successfully"})
	return nil
}

// listMembers handles GET /api/groups/{id}/members
// @Summary List group members
// @Description Get a list of members in a group (requires membership)
// @Tags groups
// @Accept json
// @Produce json
// @Param id path string true "Group ID"
// @Success 200 {object} map[string]interface{} "List of group members"
// @Failure 403 {object} httperr.ErrorResponse "Forbidden (not a member)"
// @Failure 404 {object} httperr.ErrorResponse "Group not found"
// @Failure 500 {object} httperr.ErrorResponse "Failed to list members"
// @Router /groups/{id}/members [get]
func (h *GroupHandler) listMembers(w http.ResponseWriter, r *http.Request, groupID string, currentUser *services.UserResponse) error {
	membersResponse, err := h.groupService.ListMembers(groupID, currentUser.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrGroupNotFound) {
			return httperr.NewNotFound(err, "Group not found")
		}
		if errors.Is(err, services.ErrGroupMemberRequired) {
			return httperr.NewForbidden(err, "Only group members can view the member list")
		}
		return httperr.NewInternalServerError(err, "Failed to list group members")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"members": membersResponse,
		"count":   len(membersResponse),
	})
	return nil
}

// --- Message Handlers (Merged - Stubbed) ---

// getGroupMessages handles GET /api/groups/{id}/messages
// @Summary Get group messages
// @Description Get messages from a group with pagination (requires membership)
// @Tags groups
// @Accept json
// @Produce json
// @Param id path string true "Group ID"
// @Param limit query int false "Number of messages to return (default 50)"
// @Param offset query int false "Number of messages to skip (default 0)"
// @Success 200 {object} map[string]interface{} "Group messages"
// @Failure 403 {object} httperr.ErrorResponse "Forbidden (not a member)"
// @Failure 404 {object} httperr.ErrorResponse "Group not found"
// @Failure 500 {object} httperr.ErrorResponse "Failed to get messages (or Not Implemented)"
// @Router /groups/{id}/messages [get]
func (h *GroupHandler) getGroupMessages(w http.ResponseWriter, r *http.Request, groupID string, currentUser *services.UserResponse) error {
	// TODO: Implement fully when GroupMessage service/repo methods exist
	log.Printf("getGroupMessages called for group %s (Not Implemented)", groupID)

	// Basic authorization check (is member?) - can be moved to service later
	isMember, err := h.groupService.ListMembers(groupID, currentUser.ID) // Re-use ListMembers auth check for now
	if err != nil {
		if errors.Is(err, repositories.ErrGroupNotFound) {
			return httperr.NewNotFound(err, "Group not found")
		}
		if errors.Is(err, services.ErrGroupMemberRequired) {
			return httperr.NewForbidden(err, "Only group members can view messages")
		}
		return httperr.NewInternalServerError(err, "Failed to check membership for messages")
	}
	_ = isMember // Avoid unused variable error

	// Placeholder response
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Group message endpoint not fully implemented",
		"messages": []string{},
		"limit":    limit,
		"offset":   offset,
		"count":    0,
	})
	return nil
	// return httperr.NewInternalServerError(errors.New("not implemented"), "Group message retrieval not implemented")
}

// --- Invitation Handlers ---

// inviteUserToGroup handles POST /api/groups/{groupID}/invitations
// @Summary Invite a user to a group
// @Description Invite a user to join a specific group (requires membership/admin privileges)
// @Tags groups-invitations
// @Accept json
// @Produce json
// @Param groupID path string true "Group ID"
// @Param invite body InviteUserRequest true "User ID to invite"
// @Success 201 {object} services.GroupInvitationResponse "Invitation created successfully"
// @Failure 400 {object} httperr.ErrorResponse "Invalid request body or missing invitee_id"
// @Failure 401 {object} httperr.ErrorResponse "Unauthorized"
// @Failure 403 {object} httperr.ErrorResponse "Forbidden (not member/admin)"
// @Failure 404 {object} httperr.ErrorResponse "Group or invitee user not found"
// @Failure 409 {object} httperr.ErrorResponse "User already member or already invited"
// @Failure 500 {object} httperr.ErrorResponse "Failed to invite user"
// @Router /groups/{groupID}/invitations [post]
func (h *GroupHandler) inviteUserToGroup(w http.ResponseWriter, r *http.Request, groupID string, currentUser *services.UserResponse) error {
	var req InviteUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return httperr.NewBadRequest(err, "Invalid request body")
	}
	if req.InviteeID == "" {
		return httperr.NewBadRequest(nil, "invitee_id is required")
	}

	invitationResponse, err := h.groupService.InviteUser(groupID, req.InviteeID, currentUser.ID)
	if err != nil {
		if errors.Is(err, services.ErrGroupMemberRequired) || errors.Is(err, services.ErrGroupAdminRequired) {
			return httperr.NewForbidden(err, "Only group members (or admins) can invite users")
		}
		if errors.Is(err, repositories.ErrUserNotFound) {
			return httperr.NewNotFound(err, "Invitee user not found")
		}
		if errors.Is(err, repositories.ErrGroupNotFound) {
			return httperr.NewNotFound(err, "Group not found")
		}
		if errors.Is(err, services.ErrCannotInviteSelf) {
			return httperr.NewBadRequest(err, "Cannot invite yourself")
		}
		if errors.Is(err, repositories.ErrAlreadyGroupMember) {
			return httperr.NewHTTPError(http.StatusConflict, "User is already a member", err)
		}
		if errors.Is(err, repositories.ErrAlreadyInvited) {
			return httperr.NewHTTPError(http.StatusConflict, "User has already been invited", err)
		}
		return httperr.NewInternalServerError(err, "Failed to invite user")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created for new resource
	json.NewEncoder(w).Encode(invitationResponse)
	return nil
}

// listPendingInvitations handles GET /api/groups/invitations/pending
// @Summary List pending group invitations
// @Description Get a list of pending group invitations for the current user
// @Tags groups-invitations
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "List of pending invitations"
// @Failure 401 {object} httperr.ErrorResponse "Unauthorized"
// @Failure 500 {object} httperr.ErrorResponse "Failed to list pending invitations"
// @Router /groups/invitations/pending [get]
func (h *GroupHandler) listPendingInvitations(w http.ResponseWriter, r *http.Request, currentUser *services.UserResponse) error {
	invitations, err := h.groupService.ListPendingInvitations(currentUser.ID)
	if err != nil {
		// No specific errors expected here other than internal
		return httperr.NewInternalServerError(err, "Failed to list pending invitations")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"invitations": invitations,
		"count":       len(invitations),
	})
	return nil
}

// acceptGroupInvitation handles POST /api/groups/invitations/{invitationID}/accept
// @Summary Accept a group invitation
// @Description Accept a pending group invitation
// @Tags groups-invitations
// @Accept json
// @Produce json
// @Param invitationID path string true "Invitation ID"
// @Success 200 {object} map[string]string "Invitation accepted successfully"
// @Failure 401 {object} httperr.ErrorResponse "Unauthorized"
// @Failure 403 {object} httperr.ErrorResponse "Forbidden (not the invitee)"
// @Failure 404 {object} httperr.ErrorResponse "Invitation not found"
// @Failure 409 {object} httperr.ErrorResponse "Invitation is not pending"
// @Failure 500 {object} httperr.ErrorResponse "Failed to accept invitation"
// @Router /groups/invitations/{invitationID}/accept [post]
func (h *GroupHandler) acceptGroupInvitation(w http.ResponseWriter, r *http.Request, invitationID string, currentUser *services.UserResponse) error {
	err := h.groupService.AcceptInvitation(invitationID, currentUser.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrInvitationNotFound) {
			return httperr.NewNotFound(err, "Invitation not found")
		}
		if errors.Is(err, services.ErrGroupForbidden) {
			return httperr.NewForbidden(err, "You are not authorized to accept this invitation")
		}
		if errors.Is(err, services.ErrInvalidInvitationStatus) {
			return httperr.NewHTTPError(http.StatusConflict, "Invitation is not pending", err)
		}
		// Handle potential downstream errors like group not found during member add
		if errors.Is(err, repositories.ErrGroupNotFound) {
			log.Printf("Warning: Group not found when trying to add member after accepting invite %s: %v", invitationID, err)
			// Still return OK because the invitation status was updated.
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"message": "Invitation accepted, but failed to add to group (group may no longer exist)"})
			return nil
		}
		return httperr.NewInternalServerError(err, "Failed to accept invitation")
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Invitation accepted successfully"})
	return nil
}

// rejectGroupInvitation handles POST /api/groups/invitations/{invitationID}/reject
// @Summary Reject a group invitation
// @Description Reject a pending group invitation
// @Tags groups-invitations
// @Accept json
// @Produce json
// @Param invitationID path string true "Invitation ID"
// @Success 200 {object} map[string]string "Invitation rejected successfully"
// @Failure 401 {object} httperr.ErrorResponse "Unauthorized"
// @Failure 403 {object} httperr.ErrorResponse "Forbidden (not the invitee)"
// @Failure 404 {object} httperr.ErrorResponse "Invitation not found"
// @Failure 409 {object} httperr.ErrorResponse "Invitation is not pending"
// @Failure 500 {object} httperr.ErrorResponse "Failed to reject invitation"
// @Router /groups/invitations/{invitationID}/reject [post]
func (h *GroupHandler) rejectGroupInvitation(w http.ResponseWriter, r *http.Request, invitationID string, currentUser *services.UserResponse) error {
	err := h.groupService.RejectInvitation(invitationID, currentUser.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrInvitationNotFound) {
			return httperr.NewNotFound(err, "Invitation not found")
		}
		if errors.Is(err, services.ErrGroupForbidden) {
			return httperr.NewForbidden(err, "You are not authorized to reject this invitation")
		}
		if errors.Is(err, services.ErrInvalidInvitationStatus) {
			return httperr.NewHTTPError(http.StatusConflict, "Invitation is not pending", err)
		}
		return httperr.NewInternalServerError(err, "Failed to reject invitation")
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Invitation rejected successfully"})
	return nil
}

// --- Join Request Handlers ---

// requestToJoinGroup handles POST /api/groups/{groupID}/requests
// @Summary Request to join a group
// @Description Send a request to join a specific group
// @Tags groups-requests
// @Accept json
// @Produce json
// @Param groupID path string true "Group ID"
// @Success 201 {object} services.GroupJoinRequestResponse "Join request created successfully"
// @Failure 400 {object} httperr.ErrorResponse "Cannot request to join own group"
// @Failure 401 {object} httperr.ErrorResponse "Unauthorized"
// @Failure 404 {object} httperr.ErrorResponse "Group not found"
// @Failure 409 {object} httperr.ErrorResponse "Already member or already requested"
// @Failure 500 {object} httperr.ErrorResponse "Failed to create join request"
// @Router /groups/{groupID}/requests [post]
func (h *GroupHandler) requestToJoinGroup(w http.ResponseWriter, r *http.Request, groupID string, currentUser *services.UserResponse) error {
	// No request body needed for this one

	requestResponse, err := h.groupService.RequestToJoin(groupID, currentUser.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrGroupNotFound) {
			return httperr.NewNotFound(err, "Group not found")
		}
		if errors.Is(err, services.ErrCannotRequestToJoinOwnGroup) {
			return httperr.NewBadRequest(err, "Cannot request to join a group you created")
		}
		if errors.Is(err, repositories.ErrAlreadyGroupMember) {
			return httperr.NewHTTPError(http.StatusConflict, "You are already a member of this group", err)
		}
		if errors.Is(err, repositories.ErrAlreadyRequested) {
			return httperr.NewHTTPError(http.StatusConflict, "You have already requested to join this group", err)
		}
		// ErrUserNotFound should ideally not happen if currentUser.ID is valid
		return httperr.NewInternalServerError(err, "Failed to create join request")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // 201 Created for new resource
	json.NewEncoder(w).Encode(requestResponse)
	return nil
}

// listPendingJoinRequests handles GET /api/groups/{groupID}/requests/pending
// @Summary List pending join requests for a group
// @Description Get a list of pending join requests for a specific group (requires admin privileges)
// @Tags groups-requests
// @Accept json
// @Produce json
// @Param groupID path string true "Group ID"
// @Success 200 {object} map[string]interface{} "List of pending join requests"
// @Failure 401 {object} httperr.ErrorResponse "Unauthorized"
// @Failure 403 {object} httperr.ErrorResponse "Forbidden (not admin)"
// @Failure 404 {object} httperr.ErrorResponse "Group not found"
// @Failure 500 {object} httperr.ErrorResponse "Failed to list pending join requests"
// @Router /groups/{groupID}/requests/pending [get]
func (h *GroupHandler) listPendingJoinRequests(w http.ResponseWriter, r *http.Request, groupID string, currentUser *services.UserResponse) error {
	requests, err := h.groupService.ListPendingJoinRequests(groupID, currentUser.ID)
	if err != nil {
		if errors.Is(err, services.ErrGroupAdminRequired) {
			return httperr.NewForbidden(err, "Only group admins can view pending join requests")
		}
		if errors.Is(err, repositories.ErrGroupNotFound) {
			// Check if the error is group not found before checking admin status
			return httperr.NewNotFound(err, "Group not found")
		}
		return httperr.NewInternalServerError(err, "Failed to list pending join requests")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"requests": requests,
		"count":    len(requests),
	})
	return nil
}

// acceptGroupJoinRequest handles POST /api/groups/requests/{requestID}/accept
// @Summary Accept a group join request
// @Description Accept a pending group join request (requires admin privileges)
// @Tags groups-requests
// @Accept json
// @Produce json
// @Param requestID path string true "Join Request ID"
// @Success 200 {object} map[string]string "Join request accepted successfully"
// @Failure 401 {object} httperr.ErrorResponse "Unauthorized"
// @Failure 403 {object} httperr.ErrorResponse "Forbidden (not admin)"
// @Failure 404 {object} httperr.ErrorResponse "Join request not found"
// @Failure 409 {object} httperr.ErrorResponse "Join request is not pending"
// @Failure 500 {object} httperr.ErrorResponse "Failed to accept join request"
// @Router /groups/requests/{requestID}/accept [post]
func (h *GroupHandler) acceptGroupJoinRequest(w http.ResponseWriter, r *http.Request, requestID string, currentUser *services.UserResponse) error {
	err := h.groupService.AcceptJoinRequest(requestID, currentUser.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrJoinRequestNotFound) {
			return httperr.NewNotFound(err, "Join request not found")
		}
		if errors.Is(err, services.ErrGroupAdminRequired) {
			return httperr.NewForbidden(err, "Only group admins can accept join requests")
		}
		if errors.Is(err, services.ErrInvalidJoinRequestStatus) {
			return httperr.NewHTTPError(http.StatusConflict, "Join request is not pending", err)
		}
		// Handle potential downstream errors like group not found during member add
		if errors.Is(err, repositories.ErrGroupNotFound) {
			log.Printf("Warning: Group not found when trying to add member after accepting join request %s: %v", requestID, err)
			// Still return OK because the request status was updated.
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"message": "Join request accepted, but failed to add member (group may no longer exist)"})
			return nil
		}
		return httperr.NewInternalServerError(err, "Failed to accept join request")
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Join request accepted successfully"})
	return nil
}

// rejectGroupJoinRequest handles POST /api/groups/requests/{requestID}/reject
// @Summary Reject a group join request
// @Description Reject a pending group join request (requires admin privileges)
// @Tags groups-requests
// @Accept json
// @Produce json
// @Param requestID path string true "Join Request ID"
// @Success 200 {object} map[string]string "Join request rejected successfully"
// @Failure 401 {object} httperr.ErrorResponse "Unauthorized"
// @Failure 403 {object} httperr.ErrorResponse "Forbidden (not admin)"
// @Failure 404 {object} httperr.ErrorResponse "Join request not found"
// @Failure 409 {object} httperr.ErrorResponse "Join request is not pending"
// @Failure 500 {object} httperr.ErrorResponse "Failed to reject join request"
// @Router /groups/requests/{requestID}/reject [post]
func (h *GroupHandler) rejectGroupJoinRequest(w http.ResponseWriter, r *http.Request, requestID string, currentUser *services.UserResponse) error {
	err := h.groupService.RejectJoinRequest(requestID, currentUser.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrJoinRequestNotFound) {
			return httperr.NewNotFound(err, "Join request not found")
		}
		if errors.Is(err, services.ErrGroupAdminRequired) {
			return httperr.NewForbidden(err, "Only group admins can reject join requests")
		}
		if errors.Is(err, services.ErrInvalidJoinRequestStatus) {
			return httperr.NewHTTPError(http.StatusConflict, "Join request is not pending", err)
		}
		return httperr.NewInternalServerError(err, "Failed to reject join request")
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Join request rejected successfully"})
	return nil
}
