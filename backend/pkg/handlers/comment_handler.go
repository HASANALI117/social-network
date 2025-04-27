package handlers

import (
"encoding/json"
"errors"
"log"
"net/http"
"strconv"
"strings"

"github.com/HASANALI117/social-network/pkg/helpers"
"github.com/HASANALI117/social-network/pkg/httperr"
"github.com/HASANALI117/social-network/pkg/services"
// "github.com/gorilla/mux" // If using mux for path variables
)

// CommentHandler handles HTTP requests for comments
type CommentHandler struct {
commentService services.CommentService
authService    services.AuthService
}

// NewCommentHandler creates a new CommentHandler
func NewCommentHandler(commentService services.CommentService, authService services.AuthService) *CommentHandler {
return &CommentHandler{
commentService: commentService,
authService:    authService,
}
}

// ServeHTTP routes the request to the appropriate handler method based on path and method
// Assumes base path like /api/posts/{postId}/comments or /api/comments/{commentId}
func (h *CommentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
// Get current user for authorization
currentUser, err := helpers.GetUserFromSession(r, h.authService)
// Most comment actions require authentication
isGetRequest := r.Method == http.MethodGet

// Allow anonymous GET if post is public (checked in service), but require auth for POST/DELETE
if err != nil && !isGetRequest {
if errors.Is(err, helpers.ErrInvalidSession) {
return httperr.NewUnauthorized(err, "Invalid session")
}
return httperr.NewInternalServerError(err, "Failed to get current user")
}

// Extract IDs from path - this depends heavily on the router being used.
// Example using standard library path splitting (adjust if using mux or other router):
// /api/posts/{postId}/comments -> ["api", "posts", "{postId}", "comments"]
// /api/comments/{commentId} -> ["api", "comments", "{commentId}"]
parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
log.Printf("CommentHandler: Method=%s, Path=%s, Parts=%v\n", r.Method, r.URL.Path, parts)

// Routing Logic (Simplified - Adapt to your actual router)
if len(parts) >= 4 && parts[1] == "posts" && parts[3] == "comments" {
postID := parts[2]
switch r.Method {
case http.MethodPost: // POST /api/posts/{postId}/comments
if currentUser == nil {
return httperr.NewUnauthorized(nil, "Authentication required to create comment")
}
return h.handleCreateComment(w, r, postID, currentUser)
case http.MethodGet: // GET /api/posts/{postId}/comments
requestingUserID := ""
if currentUser != nil {
requestingUserID = currentUser.ID
}
return h.handleGetCommentsByPost(w, r, postID, requestingUserID)
default:
return httperr.NewMethodNotAllowed(nil, "")
}
} else if len(parts) >= 3 && parts[1] == "comments" {
commentID := parts[2]
switch r.Method {
case http.MethodDelete: // DELETE /api/comments/{commentId}
if currentUser == nil {
return httperr.NewUnauthorized(nil, "Authentication required to delete comment")
}
return h.handleDeleteComment(w, r, commentID, currentUser.ID)
default:
return httperr.NewMethodNotAllowed(nil, "")
}
}

return httperr.NewNotFound(nil, "Invalid path for comments")
}

// handleCreateComment handles POST /api/posts/{postId}/comments
func (h *CommentHandler) handleCreateComment(w http.ResponseWriter, r *http.Request, postID string, currentUser *services.UserResponse) error {
var req services.CommentCreateRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
return httperr.NewBadRequest(err, "Invalid request body")
}

// Set IDs from context/path
req.UserID = currentUser.ID
req.PostID = postID

// Call service to create comment
commentResponse, err := h.commentService.CreateComment(&req)
if err != nil {
if errors.Is(err, services.ErrPostNotFound) {
return httperr.NewNotFound(err, "Post not found or not accessible")
}
// TODO: Handle specific validation errors if service provides them
log.Printf("Error creating comment via handler: %v", err)
return httperr.NewInternalServerError(err, "Failed to create comment")
}

w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusCreated)
json.NewEncoder(w).Encode(commentResponse)
return nil
}

// handleGetCommentsByPost handles GET /api/posts/{postId}/comments
func (h *CommentHandler) handleGetCommentsByPost(w http.ResponseWriter, r *http.Request, postID string, requestingUserID string) error {
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

// Call service to get comments
commentsResponse, err := h.commentService.GetCommentsByPost(postID, requestingUserID, limit, offset)
if err != nil {
if errors.Is(err, services.ErrPostNotFound) {
return httperr.NewNotFound(err, "Post not found or not accessible")
}
log.Printf("Error getting comments via handler: %v", err)
return httperr.NewInternalServerError(err, "Failed to get comments")
}

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]interface{}{
"comments": commentsResponse,
"limit":    limit,
"offset":   offset,
"count":    len(commentsResponse),
})
return nil
}

// handleDeleteComment handles DELETE /api/comments/{commentId}
func (h *CommentHandler) handleDeleteComment(w http.ResponseWriter, r *http.Request, commentID string, requestingUserID string) error {
err := h.commentService.DeleteComment(commentID, requestingUserID)
if err != nil {
if errors.Is(err, services.ErrCommentNotFound) {
return httperr.NewNotFound(err, "Comment not found")
}
if errors.Is(err, services.ErrCommentForbidden) {
return httperr.NewForbidden(err, "You are not authorized to delete this comment")
}
log.Printf("Error deleting comment via handler: %v", err)
return httperr.NewInternalServerError(err, "Failed to delete comment")
}

w.WriteHeader(http.StatusOK)
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]string{
"message": "Comment deleted successfully",
})
return nil
}
