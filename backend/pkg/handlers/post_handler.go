package handlers

import (
	"encoding/json"
	"errors"
	"log" // Import log
	"net/http"
	"strconv"
	"strings" // Import strings

	"github.com/HASANALI117/social-network/pkg/helpers"
	"github.com/HASANALI117/social-network/pkg/httperr"

	// "github.com/HASANALI117/social-network/pkg/models" // Model used via service request/response
	"github.com/HASANALI117/social-network/pkg/repositories" // For error checking
	"github.com/HASANALI117/social-network/pkg/services"
)

// PostHandler handles HTTP requests for posts and delegates comment routes
type PostHandler struct {
postService    services.PostService
authService    services.AuthService
commentHandler *CommentHandler // Added CommentHandler
}

// NewPostHandler creates a new PostHandler
func NewPostHandler(postService services.PostService, authService services.AuthService, commentHandler *CommentHandler) *PostHandler {
return &PostHandler{
postService:    postService,
authService:    authService,
commentHandler: commentHandler, // Store CommentHandler
}
}

// ServeHTTP routes the request to the appropriate handler method
func (h *PostHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	// Trim prefix and split path
	// Example: /api/posts/123 -> ["", "123"]
	// Example: /api/posts/ -> [""]
	// Example: /api/posts/user/456 -> ["user", "456"]
	path := strings.TrimPrefix(r.URL.Path, "/api/posts")
	path = strings.TrimPrefix(path, "/") // Remove leading slash if present
	parts := strings.Split(path, "/")

	log.Printf("PostHandler: Method=%s, Path=%s, Parts=%v\n", r.Method, r.URL.Path, parts)

	// Get current user for authorization (required for most actions)
	currentUser, err := helpers.GetUserFromSession(r, h.authService)
	if err != nil && !(r.Method == http.MethodGet && len(parts) == 1 && parts[0] != "") {
		// Allow anonymous GET /api/posts/{id} for public posts (checked in service)
		// Allow anonymous GET /api/posts/ for public posts (checked in service)
		// Allow anonymous GET /api/posts/user/{userID} for public posts (checked in service)
		// For other methods (POST, DELETE) or paths, require authentication.
		if errors.Is(err, helpers.ErrInvalidSession) {
			return httperr.NewUnauthorized(err, "Invalid session")
		}
		return httperr.NewInternalServerError(err, "Failed to get current user")
	}
// If currentUser is nil here, it means it's an anonymous GET request

// Check if this is a comment-related route nested under posts
// e.g., /api/posts/{postId}/comments -> parts = ["{postId}", "comments"]
if len(parts) >= 2 && parts[1] == "comments" {
// Delegate to CommentHandler. ServeHTTP needs to handle this specific path structure.
// We might need to adjust CommentHandler's ServeHTTP if it relies on a different path format.
log.Printf("PostHandler delegating to CommentHandler for path: %s", r.URL.Path)
// Re-route by calling the CommentHandler's ServeHTTP directly
// Note: The CommentHandler's ServeHTTP will re-parse the path, which might be inefficient
// or require adjustments in CommentHandler.
return h.commentHandler.ServeHTTP(w, r)
}

// --- Original Post Routing Logic ---
switch r.Method {
case http.MethodPost:
		// POST /api/posts/ -> Create Post
		if len(parts) == 1 && parts[0] == "" {
			if currentUser == nil { // Must be logged in to create
				return httperr.NewUnauthorized(nil, "Authentication required to create post")
			}
			return h.createPost(w, r, currentUser)
		}
		return httperr.NewNotFound(nil, "Invalid path for POST")

	case http.MethodGet:
		// GET /api/posts/{id} -> Get Post by ID
		if len(parts) == 1 && parts[0] != "" {
			postID := parts[0]
			requestingUserID := ""
			if currentUser != nil {
				requestingUserID = currentUser.ID
			}
			return h.getPostByID(w, r, postID, requestingUserID)
		}
		// GET /api/posts/ -> List Posts (all public/friends)
		if len(parts) == 1 && parts[0] == "" {
			requestingUserID := ""
			if currentUser != nil {
				requestingUserID = currentUser.ID
			}
			return h.listPosts(w, r, requestingUserID)
		}
		// GET /api/posts/user/{userID} -> List Posts by User
		if len(parts) == 2 && parts[0] == "user" && parts[1] != "" {
			targetUserID := parts[1]
			requestingUserID := ""
			if currentUser != nil {
				requestingUserID = currentUser.ID
			}
			return h.listUserPosts(w, r, targetUserID, requestingUserID)
		}
		return httperr.NewNotFound(nil, "Invalid path for GET")

	case http.MethodPut:
		// PUT /api/posts/{id} -> Update Post (Not Implemented Yet)
		return httperr.NewMethodNotAllowed(nil, "PUT method not implemented for posts")

	case http.MethodDelete:
		// DELETE /api/posts/{id} -> Delete Post
		if len(parts) == 1 && parts[0] != "" {
			if currentUser == nil { // Must be logged in to delete
				return httperr.NewUnauthorized(nil, "Authentication required to delete post")
			}
			postID := parts[0]
			return h.deletePost(w, r, postID, currentUser.ID)
		}
		return httperr.NewNotFound(nil, "Invalid path for DELETE")

	default:
		return httperr.NewMethodNotAllowed(nil, "")
	}
}

// createPost handles POST /api/posts/
// @Summary Create a new post
// @Description Create a new post in the system
// @Tags posts
// @Accept json
// @Produce json
// @Param post body models.Post true "Post creation details"
// @Success 201 {object} map[string]interface{} "Post created successfully"
// @Failure 400 {object} httperr.ErrorResponse "Invalid request body"
// @Failure 405 {object} httperr.ErrorResponse "Method not allowed"
// @Failure 500 {object} httperr.ErrorResponse "Failed to create post or validation error"
// @Router /posts [post]
func (h *PostHandler) createPost(w http.ResponseWriter, r *http.Request, currentUser *services.UserResponse) error {
	var req services.PostCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return httperr.NewBadRequest(err, "Invalid request body")
	}

	// Set UserID from authenticated user
	req.UserID = currentUser.ID

	// Call service to create post
	postResponse, err := h.postService.Create(&req)
	if err != nil {
		// TODO: Handle specific validation errors from service if implemented
		return httperr.NewInternalServerError(err, "Failed to create post")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(postResponse)
	return nil
}

// getPostByID handles GET /api/posts/{id}
// @Summary Get post by ID
// @Description Get post details by post ID
// @Tags posts
// @Accept json
// @Produce json
// @Param id query string true "Post ID"
// @Success 200 {object} map[string]interface{} "Post details"
// @Failure 400 {object} httperr.ErrorResponse "Post ID is required"
// @Failure 404 {object} httperr.ErrorResponse "Post not found"
// @Failure 405 {object} httperr.ErrorResponse "Method not allowed"
// @Failure 403 {object} httperr.ErrorResponse "Forbidden (not authorized to view)"
// @Failure 500 {object} httperr.ErrorResponse "Failed to get post"
// @Router /posts/{id} [get]
func (h *PostHandler) getPostByID(w http.ResponseWriter, r *http.Request, postID string, requestingUserID string) error {
	postResponse, err := h.postService.GetByID(postID, requestingUserID)
	if err != nil {
		if errors.Is(err, repositories.ErrPostNotFound) {
			return httperr.NewNotFound(err, "Post not found")
		}
		if errors.Is(err, services.ErrPostForbidden) {
			// Return 404 instead of 403 to not reveal existence of private posts
			return httperr.NewNotFound(err, "Post not found")
			// Or return 403 if revealing existence is acceptable:
			// return httperr.NewForbidden(err, "You are not authorized to view this post")
		}
		return httperr.NewInternalServerError(err, "Failed to get post")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(postResponse)
	return nil
}

// listPosts handles GET /api/posts/
// @Summary List posts
// @Description Get a paginated list of posts
// @Tags posts
// @Accept json
// @Produce json
// @Param limit query int false "Number of posts to return (default 10)"
// @Param offset query int false "Number of posts to skip (default 0)"
// @Success 200 {object} map[string]interface{} "List of posts"
// @Failure 405 {object} httperr.ErrorResponse "Method not allowed"
// @Failure 500 {object} httperr.ErrorResponse "Failed to list posts"
// @Router /posts [get]
func (h *PostHandler) listPosts(w http.ResponseWriter, r *http.Request, requestingUserID string) error {
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

	// Call service to list posts (service handles filtering logic)
	postsResponse, err := h.postService.List(requestingUserID, limit, offset) // Pass requestingUserID first
	if err != nil {
		return httperr.NewInternalServerError(err, "Failed to list posts")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"posts":  postsResponse,
		"limit":  limit,
		"offset": offset,
		"count":  len(postsResponse), // Count of returned posts
	})
	return nil
}

// listUserPosts handles GET /api/posts/user/{userID}
// @Summary List posts by user
// @Description Get a paginated list of posts created by a specific user
// @Tags posts
// @Accept json
// @Produce json
// @Param user_id query string true "User ID"
// @Param limit query int false "Number of posts to return (default 10)"
// @Param offset query int false "Number of posts to skip (default 0)"
// @Success 200 {object} map[string]interface{} "List of user's posts"
// @Failure 400 {object} httperr.ErrorResponse "User ID is required"
// @Failure 405 {object} httperr.ErrorResponse "Method not allowed"
// @Failure 500 {object} httperr.ErrorResponse "Failed to list user posts"
// @Router /posts/user/{userID} [get]
func (h *PostHandler) listUserPosts(w http.ResponseWriter, r *http.Request, targetUserID string, requestingUserID string) error {
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

	// Call service to list posts by user (service handles filtering logic)
	postsResponse, err := h.postService.ListPostsByUser(targetUserID, requestingUserID, limit, offset) // Pass requestingUserID before limit/offset
	if err != nil {
		return httperr.NewInternalServerError(err, "Failed to list user posts")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"posts":  postsResponse,
		"limit":  limit,
		"offset": offset,
		"count":  len(postsResponse), // Count of returned posts
	})
	return nil
}

// deletePost handles DELETE /api/posts/{id}
// @Summary Delete post
// @Description Delete a post by ID
// @Tags posts
// @Accept json
// @Produce json
// @Param id query string true "Post ID"
// @Success 200 {object} map[string]string "Post deleted successfully"
// @Failure 400 {object} httperr.ErrorResponse "Post ID is required"
// @Failure 404 {object} httperr.ErrorResponse "Post not found"
// @Failure 405 {object} httperr.ErrorResponse "Method not allowed"
// @Failure 403 {object} httperr.ErrorResponse "Forbidden (not authorized to delete)"
// @Failure 500 {object} httperr.ErrorResponse "Failed to delete post"
// @Router /posts/{id} [delete]
func (h *PostHandler) deletePost(w http.ResponseWriter, r *http.Request, postID string, requestingUserID string) error {
	err := h.postService.Delete(postID, requestingUserID)
	if err != nil {
		if errors.Is(err, repositories.ErrPostNotFound) {
			return httperr.NewNotFound(err, "Post not found")
		}
		if errors.Is(err, services.ErrPostForbidden) {
			return httperr.NewForbidden(err, "You are not authorized to delete this post")
		}
		return httperr.NewInternalServerError(err, "Failed to delete post")
	}

	w.WriteHeader(http.StatusOK) // Use 200 OK for successful delete
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Post deleted successfully",
	})
	return nil
}
