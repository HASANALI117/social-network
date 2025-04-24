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

// No group ID specified
if len(parts) == 1 && parts[0] == "" {
switch r.Method {
case http.MethodPost: // POST /api/groups/ -> Create Group
return h.createGroup(w, r, currentUser)
case http.MethodGet: // GET /api/groups/ -> List Groups
return h.listGroups(w, r, currentUser)
default:
return httperr.NewMethodNotAllowed(nil, "Method not allowed for /api/groups/")
}
}

// Group ID specified: /api/groups/{groupID}...
if len(parts) >= 1 && parts[0] != "" {
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
case "members":
// /api/groups/{groupID}/members
if len(parts) == 2 {
switch r.Method {
case http.MethodPost: // POST /api/groups/{groupID}/members -> Add Member
return h.addMember(w, r, groupID, currentUser)
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

// addMember handles POST /api/groups/{id}/members
// @Summary Add a member to a group
// @Description Add a user to a group (requires admin privileges)
// @Tags groups
// @Accept json
// @Produce json
// @Param id path string true "Group ID"
// @Param member body object{user_id=string, role=string} true "User ID and optional role (defaults to 'member')"
// @Success 200 {object} map[string]string "Member added successfully"
// @Failure 400 {object} httperr.ErrorResponse "Invalid request body or missing user_id"
// @Failure 403 {object} httperr.ErrorResponse "Forbidden (not admin)"
// @Failure 404 {object} httperr.ErrorResponse "Group or target user not found"
// @Failure 409 {object} httperr.ErrorResponse "User already in group"
// @Failure 500 {object} httperr.ErrorResponse "Failed to add member"
// @Router /groups/{id}/members [post]
func (h *GroupHandler) addMember(w http.ResponseWriter, r *http.Request, groupID string, currentUser *services.UserResponse) error {
var req struct {
UserID string `json:"user_id"`
Role   string `json:"role"` // Optional, defaults to "member" in service
}
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
return httperr.NewBadRequest(err, "Invalid request body")
}
if req.UserID == "" {
return httperr.NewBadRequest(nil, "user_id is required")
}

err := h.groupService.AddMember(groupID, req.UserID, req.Role, currentUser.ID)
if err != nil {
if errors.Is(err, services.ErrGroupAdminRequired) {
return httperr.NewForbidden(err, "Only group admins can add members")
}
if errors.Is(err, repositories.ErrGroupNotFound) || errors.Is(err, repositories.ErrUserNotFound) || strings.Contains(err.Error(), "not found") {
// Treat user not found or group not found as 404
return httperr.NewNotFound(err, "Group or target user not found")
}
if errors.Is(err, repositories.ErrAlreadyGroupMember) {
return httperr.NewHTTPError(http.StatusConflict, "User is already a member of this group", err)
}
return httperr.NewInternalServerError(err, "Failed to add group member")
}

w.WriteHeader(http.StatusOK)
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]string{"message": "Member added successfully"})
return nil
}

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
if l, err := strconv.Atoi(limitStr); err == nil && l > 0 { limit = l }
}
offset := 0
if offsetStr != "" {
if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 { offset = o }
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
