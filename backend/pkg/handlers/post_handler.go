package handlers

	"time" // Added for DTO

	"github.com/HASANALI117/social-network/pkg/apperrors" // Added
	"github.com/HASANALI117/social-network/pkg/helpers"
	"github.com/HASANALI117/social-network/pkg/models"
)

// --- DTOs ---

// CreatePostRequest defines the expected body for creating a post.
type CreatePostRequest struct {
	UserID   string `json:"user_id"` // Should likely be inferred from session/auth token
	Title    string `json:"title"`
	Content  string `json:"content"`
	ImageURL string `json:"image_url"` // Optional?
	Privacy  string `json:"privacy"`   // Consider enum or validation
}

// PostResponse defines the post data returned by the API.
type PostResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	ImageURL  string    `json:"image_url"`
	Privacy   string    `json:"privacy"`
	CreatedAt time.Time `json:"created_at"`
	// Add UpdatedAt if applicable
}

// ListPostsResponse defines the structure for listing posts.
type ListPostsResponse struct {
	Posts  []PostResponse `json:"posts"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
	Count  int            `json:"count"`
}

// DeletePostResponse defines the success message for post deletion.
type DeletePostResponse struct {
	Message string `json:"message"`
}

// --- Helper Function ---

// mapModelPostToPostResponse converts a models.Post to a PostResponse DTO.
func mapModelPostToPostResponse(post *models.Post) PostResponse {
	return PostResponse{
		ID:        post.ID,
		UserID:    post.UserID,
		Title:     post.Title,
		Content:   post.Content,
		ImageURL:  post.ImageURL,
		Privacy:   post.Privacy,
		CreatedAt: post.CreatedAt,
	}
}

// --- Handlers ---

// PostHandler handles HTTP requests for posts (Placeholder for future DI)
// type PostHandler struct {
// 	postDB *helpers.PostDB
// }

// NewPostHandler creates a new PostHandler
// func NewPostHandler() *PostHandler {
// 	return &PostHandler{}
// }

// CreatePost godoc
// @Summary Create a new post
// @Description Create a new post. Requires authentication. UserID should be inferred from session.
// @Tags posts
// @Accept json
// @Produce json
// @Param post body CreatePostRequest true "Post creation details (UserID ignored, taken from session)"
// @Success 201 {object} PostResponse "Post created successfully"
// @Failure 400 {object} map[string]string "Invalid request body or missing fields (e.g., content)"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 405 {object} map[string]string "Method not allowed"
// @Failure 500 {object} map[string]string "Failed to create post"
// @Router /posts/create [post]
func CreatePost(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return apperrors.ErrMethodNotAllowed("", nil)
	}

	// Get current user from session
	// TODO: Replace helpers.GetUserFromSession when auth middleware/service is added
	currentUser, err := helpers.GetUserFromSession(r)
	if err != nil {
		return apperrors.ErrUnauthorized("Unauthorized", err)
	}

	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return apperrors.ErrBadRequest("Invalid request body", err)
	}
	defer r.Body.Close()

	// Basic validation
	if req.Content == "" { // Title might be optional?
		return apperrors.ErrBadRequest("Post content is required", nil)
	}
	// TODO: Validate Privacy field against allowed values ("public", "private", "friends")?

	// Map request DTO to internal model, using authenticated user's ID
	post := models.Post{
		UserID:   currentUser.ID, // Use ID from session, ignore req.UserID
		Title:    req.Title,
		Content:  req.Content,
		ImageURL: req.ImageURL,
		Privacy:  req.Privacy,
	}

	// Create post in database
	// TODO: Replace helpers.CreatePost when service layer is added
	if err := helpers.CreatePost(&post); err != nil {
		log.Printf("Failed to create post for user %s: %v", currentUser.ID, err)
		return apperrors.ErrInternalServer("Failed to create post", err)
	}

	// Return created post DTO
	response := mapModelPostToPostResponse(&post)
	return helpers.RespondJSON(w, http.StatusCreated, response)
}

// GetPost godoc
// @Summary Get post by ID
// @Description Get post details by post ID. Access control based on post privacy might be needed.
// @Tags posts
// @Accept json
// @Produce json
// @Param id query string true "Post ID"
// @Success 200 {object} PostResponse "Post details"
// @Failure 400 {object} map[string]string "Post ID is required or invalid"
// @Failure 404 {object} map[string]string "Post not found"
// @Failure 405 {object} map[string]string "Method not allowed"
// @Failure 500 {object} map[string]string "Failed to get post"
// @Router /posts/get [get]
func GetPost(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return apperrors.ErrMethodNotAllowed("", nil)
	}

	postID := r.URL.Query().Get("id")
	if postID == "" {
		return apperrors.ErrBadRequest("Post ID is required", nil)
	}
	// Optional: Validate postID format

	// Get post from database
	// TODO: Replace helpers.GetPostByID when service layer is added
	post, err := helpers.GetPostByID(postID)
	if err != nil {
		if errors.Is(err, helpers.ErrPostNotFound) { // Assuming helpers.ErrPostNotFound exists
			return apperrors.ErrNotFound("Post not found", err)
		}
		return apperrors.ErrInternalServer("Failed to get post", err)
	}

	// TODO: Implement privacy check here.
	// Get current user (if authenticated) and check if they have access based on post.Privacy
	// currentUser, authErr := helpers.GetUserFromSession(r)
	// if post.Privacy == "private" && (authErr != nil || currentUser.ID != post.UserID) {
	//     return apperrors.ErrForbidden("You do not have permission to view this post", nil)
	// }
	// Add checks for "friends" privacy level if implemented.

	// Return post DTO
	response := mapModelPostToPostResponse(post)
	return helpers.RespondJSON(w, http.StatusOK, response)
}

// ListPosts godoc
// @Summary List posts
// @Description Get a paginated list of public posts (or posts visible to the user).
// @Tags posts
// @Accept json
// @Produce json
// @Param limit query int false "Number of posts to return (default 10)" minimum(1) maximum(100)
// @Param offset query int false "Number of posts to skip (default 0)" minimum(0)
// @Success 200 {object} ListPostsResponse "List of posts"
// @Failure 400 {object} map[string]string "Invalid limit or offset parameter"
// @Failure 405 {object} map[string]string "Method not allowed"
// @Failure 500 {object} map[string]string "Failed to list posts"
// @Router /posts/list [get]
func ListPosts(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return apperrors.ErrMethodNotAllowed("", nil)
	}

	// Parse pagination parameters
	limit, offset, err := helpers.ParsePagination(r.URL.Query(), 10, 100) // Assuming ParsePagination helper exists/is created
	if err != nil {
		return apperrors.ErrBadRequest(err.Error(), err)
	}

	// TODO: Implement logic to fetch posts based on visibility (public, friends, own)
	// This might involve getting the current user ID and passing it to the ListPosts helper/service.
	// currentUser, _ := helpers.GetUserFromSession(r) // Ignore error for now, handle anonymous users

	// Get posts from database
	// TODO: Replace helpers.ListPosts when service layer is added. It should handle privacy.
	posts, err := helpers.ListPosts(limit, offset) // This likely needs modification for privacy
	if err != nil {
		return apperrors.ErrInternalServer("Failed to list posts", err)
	}

	// Map models to response DTOs
	postResponses := make([]PostResponse, len(posts))
	for i, post := range posts {
		postResponses[i] = mapModelPostToPostResponse(post)
	}

	// Return post list DTO
	response := ListPostsResponse{
		Posts:  postResponses,
		Limit:  limit,
		Offset: offset,
		Count:  len(postResponses), // Note: This might not be the total count in DB
	}
	return helpers.RespondJSON(w, http.StatusOK, response)
}

// ListUserPosts godoc
// @Summary List posts by user
// @Description Get a paginated list of posts created by a specific user. Considers privacy relative to the requesting user.
// @Tags posts
// @Accept json
// @Produce json
// @Param id query string true "User ID whose posts are being requested"
// @Param limit query int false "Number of posts to return (default 10)" minimum(1) maximum(100)
// @Param offset query int false "Number of posts to skip (default 0)" minimum(0)
// @Success 200 {object} ListPostsResponse "List of user's posts"
// @Failure 400 {object} map[string]string "User ID is required or invalid pagination"
// @Failure 404 {object} map[string]string "Target user not found" // Added
// @Failure 405 {object} map[string]string "Method not allowed"
// @Failure 500 {object} map[string]string "Failed to list posts"
// @Router /posts/user [get]
func ListUserPosts(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodGet {
		return apperrors.ErrMethodNotAllowed("", nil)
	}

	targetUserID := r.URL.Query().Get("id")
	if targetUserID == "" {
		return apperrors.ErrBadRequest("Target User ID is required", nil)
	}

	// Parse pagination parameters
	limit, offset, err := helpers.ParsePagination(r.URL.Query(), 10, 100) // Assuming ParsePagination helper
	if err != nil {
		return apperrors.ErrBadRequest(err.Error(), err)
	}

	// TODO: Check if targetUserID exists? Optional optimization.
	// _, userErr := helpers.GetUserByID(targetUserID)
	// if userErr != nil {
	//     if errors.Is(userErr, helpers.ErrUserNotFound) {
	//         return apperrors.ErrNotFound("Target user not found", userErr)
	//     }
	//     return apperrors.ErrInternalServer("Failed to check target user", userErr)
	// }


	// Get current user (if any) to check privacy rules
	// TODO: Replace helpers.GetUserFromSession when auth middleware/service is added
	// currentUser, _ := helpers.GetUserFromSession(r) // Ignore error for anonymous access check

	// Get posts from database
	// TODO: Replace helpers.ListPostsByUser when service layer is added.
	// This helper/service MUST implement privacy checks based on the requesting user (currentUser.ID)
	// and the target user (targetUserID).
	posts, err := helpers.ListPostsByUser(targetUserID, limit, offset /*, currentUser.ID */)
	if err != nil {
		log.Printf("Failed to list posts for user %s: %v", targetUserID, err)
		// Check if the error indicates the target user doesn't exist, if ListPostsByUser provides that info
		// if errors.Is(err, helpers.ErrUserNotFound) {
		//     return apperrors.ErrNotFound("Target user not found", err)
		// }
		return apperrors.ErrInternalServer("Failed to list user posts", err)
	}

	// Map models to response DTOs
	postResponses := make([]PostResponse, len(posts))
	for i, post := range posts {
		postResponses[i] = mapModelPostToPostResponse(post)
	}

	// Return post list DTO
	response := ListPostsResponse{
		Posts:  postResponses,
		Limit:  limit,
		Offset: offset,
		Count:  len(postResponses),
	}
	return helpers.RespondJSON(w, http.StatusOK, response)
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
// func UpdatePost(w http.ResponseWriter, r *http.Request) {
// 	if r.Method != http.MethodPut {
// 		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
// 		return
// 	}

// 	postID := r.URL.Query().Get("id")
// 	if postID == "" {
// 		http.Error(w, "Post ID is required", http.StatusBadRequest)
// 		return
// 	}

// 	post, err := helpers.GetByID(postID)
// 	if err != nil {
// 		if errors.Is(err, helpers.ErrPostNotFound) {
// 			http.Error(w, "Post not found", http.StatusNotFound)
// 		} else {
// 			http.Error(w, "Failed to get post", http.StatusInternalServerError)
// 		}
// 		return
// 	}

// 	var req struct {
// 		Title    string `json:"title"`
// 		Content  string `json:"content"`
// 		ImageURL string `json:"image_url"`
// 		Privacy  string `json:"privacy"`
// 	}

// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		http.Error(w, "Invalid request body", http.StatusBadRequest)
// 		return
// 	}

// 	post.Title = req.Title
// 	post.Content = req.Content
// 	post.ImageURL = req.ImageURL
// 	post.Privacy = req.Privacy

// 	if err := helpers.Update(post); err != nil {
// 		http.Error(w, "Failed to update post", http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(map[string]interface{}{
// 		"id":         post.ID,
// 		"user_id":    post.UserID,
// 		"title":      post.Title,
// 		"content":    post.Content,
// 		"image_url":  post.ImageURL,
// 		"privacy":    post.Privacy,
// 		"created_at": post.CreatedAt,
// 	})
// }

// DeletePost godoc
// @Summary Delete post
// @Description Delete a post by ID. Requires authentication and ownership.
// @Tags posts
// @Accept json
// @Produce json
// @Param id query string true "Post ID"
// @Success 200 {object} DeletePostResponse "Post deleted successfully"
// @Failure 400 {object} map[string]string "Post ID is required or invalid"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden (user does not own post)"
// @Failure 404 {object} map[string]string "Post not found"
// @Failure 405 {object} map[string]string "Method not allowed"
// @Failure 500 {object} map[string]string "Failed to delete post"
// @Router /posts/delete [delete]
func DeletePost(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodDelete {
		return apperrors.ErrMethodNotAllowed("", nil)
	}

	postID := r.URL.Query().Get("id")
	if postID == "" {
		return apperrors.ErrBadRequest("Post ID is required", nil)
	}

	// Get current user from session for authorization
	// TODO: Replace helpers.GetUserFromSession when auth middleware/service is added
	currentUser, err := helpers.GetUserFromSession(r)
	if err != nil {
		return apperrors.ErrUnauthorized("Unauthorized", err)
	}

	// --- Authorization Check ---
	// Get the post first to check ownership
	// TODO: Replace helpers.GetPostByID when service layer is added
	post, err := helpers.GetPostByID(postID)
	if err != nil {
		if errors.Is(err, helpers.ErrPostNotFound) {
			return apperrors.ErrNotFound("Post not found", err)
		}
		return apperrors.ErrInternalServer("Failed to retrieve post for deletion check", err)
	}

	// Check if the current user owns the post
	if post.UserID != currentUser.ID {
		log.Printf("WARN: User %s attempted to delete post %s owned by user %s", currentUser.ID, postID, post.UserID)
		return apperrors.ErrForbidden("You do not have permission to delete this post", nil)
	}
	// --- End Authorization Check ---

	// Delete post from database
	// TODO: Replace helpers.DeletePost when service layer is added
	if err := helpers.DeletePost(postID); err != nil {
		// We already checked for not found, so this is likely a real DB error
		return apperrors.ErrInternalServer("Failed to delete post", err)
	}

	// Return success response
	response := DeletePostResponse{Message: "Post deleted successfully"}
	return helpers.RespondJSON(w, http.StatusOK, response)
}
