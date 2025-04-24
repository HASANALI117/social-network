package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/HASANALI117/social-network/pkg/helpers"
	"github.com/HASANALI117/social-network/pkg/httperr"
	"github.com/HASANALI117/social-network/pkg/models"
)

// CreatePost godoc
// @Summary Create a new post
// @Description Create a new post in the system
// @Tags posts
// @Accept json
// @Produce json
// @Param post body models.Post true "Post creation details"
// @Success 201 {object} map[string]interface{} "Post created successfully"
// @Failure 400 {object} httperr.ErrorResponse "Invalid request body"
// @Failure 405 {object} httperr.ErrorResponse "Method not allowed"
// @Failure 500 {object} httperr.ErrorResponse "Failed to create post"
// @Router /posts/create [post]
func CreatePost(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return httperr.NewMethodNotAllowed(nil, "")
	}

	var post models.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		return httperr.NewBadRequest(err, "Invalid request body")
	}

	if err := helpers.CreatePost(&post); err != nil {
		return httperr.NewInternalServerError(err, "Failed to create post")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         post.ID,
		"user_id":    post.UserID,
		"title":      post.Title,
		"content":    post.Content,
		"image_url":  post.ImageURL,
		"privacy":    post.Privacy,
		"created_at": post.CreatedAt,
	})
	return nil
}

// GetPost godoc
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
// @Failure 500 {object} httperr.ErrorResponse "Failed to get post"
// @Router /posts/get [get]
func GetPost(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return httperr.NewMethodNotAllowed(nil, "")
	}

	postID := r.URL.Query().Get("id")
	if postID == "" {
		return httperr.NewBadRequest(nil, "Post ID is required")
	}

	post, err := helpers.GetPostByID(postID)
	if err != nil {
		if errors.Is(err, helpers.ErrPostNotFound) {
			return httperr.NewNotFound(err, "Post not found")
		}
		return httperr.NewInternalServerError(err, "Failed to get post")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":         post.ID,
		"user_id":    post.UserID,
		"title":      post.Title,
		"content":    post.Content,
		"image_url":  post.ImageURL,
		"privacy":    post.Privacy,
		"created_at": post.CreatedAt,
	})
	return nil
}

// ListPosts godoc
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
// @Router /posts/list [get]
func ListPosts(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return httperr.NewMethodNotAllowed(nil, "")
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	posts, err := helpers.ListPosts(limit, offset)
	if err != nil {
		return httperr.NewInternalServerError(err, "Failed to list posts")
	}

	result := make([]map[string]interface{}, len(posts))
	for i, post := range posts {
		result[i] = map[string]interface{}{
			"id":         post.ID,
			"user_id":    post.UserID,
			"title":      post.Title,
			"content":    post.Content,
			"image_url":  post.ImageURL,
			"privacy":    post.Privacy,
			"created_at": post.CreatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"posts":  result,
		"limit":  limit,
		"offset": offset,
		"count":  len(posts),
	})
	return nil
}

// ListUserPosts godoc
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
// @Failure 500 {object} httperr.ErrorResponse "Failed to list posts"
// @Router /posts/user [get]
func ListUserPosts(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return httperr.NewMethodNotAllowed(nil, "")
	}

	userID := r.URL.Query().Get("id")
	if userID == "" {
		return httperr.NewBadRequest(nil, "User ID is required")
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 10
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	posts, err := helpers.ListPostsByUser(userID, limit, offset)
	if err != nil {
		return httperr.NewInternalServerError(err, "Failed to list posts")
	}

	result := make([]map[string]interface{}, len(posts))
	for i, post := range posts {
		result[i] = map[string]interface{}{
			"id":         post.ID,
			"user_id":    post.UserID,
			"title":      post.Title,
			"content":    post.Content,
			"image_url":  post.ImageURL,
			"privacy":    post.Privacy,
			"created_at": post.CreatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"posts":  result,
		"limit":  limit,
		"offset": offset,
		"count":  len(posts),
	})
	return nil
}

// DeletePost godoc
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
// @Failure 500 {object} httperr.ErrorResponse "Failed to delete post"
// @Router /posts/delete [delete]
func DeletePost(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodDelete {
		return httperr.NewMethodNotAllowed(nil, "")
	}

	postID := r.URL.Query().Get("id")
	if postID == "" {
		return httperr.NewBadRequest(nil, "Post ID is required")
	}

	if err := helpers.DeletePost(postID); err != nil {
		if errors.Is(err, helpers.ErrPostNotFound) {
			return httperr.NewNotFound(err, "Post not found")
		}
		return httperr.NewInternalServerError(err, "Failed to delete post")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Post deleted successfully",
	})
	return nil
}
