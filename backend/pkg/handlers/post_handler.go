package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"social-network/pkg/db"
	"social-network/pkg/helpers"
	"social-network/pkg/models"
)

// PostHandler handles HTTP requests for posts
// type PostHandler struct {
// 	postDB *helpers.PostDB
// }

// NewPostHandler creates a new PostHandler
// func NewPostHandler() *PostHandler {
// 	return &PostHandler{}
// }

// CreatePost godoc
// @Summary Create a new post
// @Description Create a new post in the system
// @Tags posts
// @Accept json
// @Produce json
// @Param post body models.Post true "Post creation details"
// @Success 201 {object} map[string]interface{} "Post created successfully"
// @Failure 400 {string} string "Invalid request body"
// @Failure 500 {string} string "Failed to create post"
// @Router /posts/create [post]
func CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var post models.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := helpers.CreatePost(&post); err != nil {
		log.Printf("Failed to create post: %v", err) // Log the error
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
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
}

// GetPost godoc
// @Summary Get post by ID
// @Description Get post details by post ID
// @Tags posts
// @Accept json
// @Produce json
// @Param id query string true "Post ID"
// @Success 200 {object} map[string]interface{} "Post details"
// @Failure 400 {string} string "Post ID is required"
// @Failure 404 {string} string "Post not found"
// @Failure 500 {string} string "Failed to get post"
// @Router /posts/get [get]
func GetPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	postID := r.URL.Query().Get("id")
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	post, err := helpers.GetPostByID(postID)
	if err != nil {
		if errors.Is(err, helpers.ErrPostNotFound) {
			http.Error(w, "Post not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get post", http.StatusInternalServerError)
		}
		return
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
// @Failure 500 {string} string "Failed to list posts"
// @Router /posts/list [get]
func ListPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
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
		http.Error(w, "Failed to list posts", http.StatusInternalServerError)
		return
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
// @Failure 400 {string} string "User ID is required"
// @Failure 500 {string} string "Failed to list posts"
// @Router /posts/user [get]
func ListUserPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
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
		log.Printf("Failed to list user posts: %v", err)
		http.Error(w, "Failed to list posts", http.StatusInternalServerError)
		return
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
}

// UpdatePost godoc
// @Summary Update post
// @Description Update post details
// @Tags posts
// @Accept json
// @Produce json
// @Param id query string true "Post ID"
// @Param post body object true "Post update details"
// @Success 200 {object} map[string]interface{} "Updated post details"
// @Failure 400 {string} string "Invalid request"
// @Failure 404 {string} string "Post not found"
// @Failure 500 {string} string "Failed to update post"
// @Router /posts/update [put]
func UpdatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	postID := r.URL.Query().Get("id")
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	post, err := helpers.GetPostByID(postID)
	if err != nil {
		if errors.Is(err, helpers.ErrPostNotFound) {
			http.Error(w, "Post not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get post", http.StatusInternalServerError)
		}
		return
	}

	var req struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		ImageURL string `json:"image_url"`
		Privacy  string `json:"privacy"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	post.Title = req.Title
	post.Content = req.Content
	post.ImageURL = req.ImageURL
	post.Privacy = req.Privacy

	query := `
        UPDATE posts
        SET title = ?, content = ?, image_url = ?, privacy = ?
        WHERE id = ?
    `
	_, err = db.GlobalDB.Exec(query, post.Title, post.Content, post.ImageURL, post.Privacy, post.ID)
	if err != nil {
		http.Error(w, "Failed to update post", http.StatusInternalServerError)
		return
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
}

// DeletePost godoc
// @Summary Delete post
// @Description Delete a post by ID
// @Tags posts
// @Accept json
// @Produce json
// @Param id query string true "Post ID"
// @Success 200 {object} map[string]string "Post deleted successfully"
// @Failure 400 {string} string "Post ID is required"
// @Failure 404 {string} string "Post not found"
// @Failure 500 {string} string "Failed to delete post"
// @Router /posts/delete [delete]
func DeletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	postID := r.URL.Query().Get("id")
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	if err := helpers.DeletePost(postID); err != nil {
		if errors.Is(err, helpers.ErrPostNotFound) {
			http.Error(w, "Post not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete post", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Post deleted successfully",
	})
}
