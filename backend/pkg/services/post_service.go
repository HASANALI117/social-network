package services

import (
	"database/sql" // Needed for sql.ErrNoRows check in follower lookup
	"errors"
	"fmt"
	"log" // For logging errors during auth checks
	"time"

	"github.com/HASANALI117/social-network/pkg/models"
	"github.com/HASANALI117/social-network/pkg/repositories"
)

// PostResponse is the DTO for post data sent to clients
type PostResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	ImageURL  string    `json:"image_url,omitempty"`
	Privacy   string    `json:"privacy"`
	CreatedAt time.Time `json:"created_at"`
	// TODO: Add user details (username, avatar) if needed
}

// PostCreateRequest is the DTO for creating a new post
type PostCreateRequest struct {
	UserID         string   `json:"-"` // Set internally from authenticated user
	Title          string   `json:"title" validate:"required,max=100"`
	Content        string   `json:"content" validate:"required"`
	ImageURL       string   `json:"image_url" validate:"omitempty,url"`
	Privacy        string   `json:"privacy" validate:"required,oneof=public almost_private private"` // Updated privacy options
	AllowedUserIDs []string `json:"allowed_user_ids,omitempty"`                                      // For 'private' posts
}

var (
	ErrPostForbidden = errors.New("user not authorized to perform this action on the post")
)

// PostService defines the interface for post business logic
type PostService interface {
	Create(request *PostCreateRequest) (*PostResponse, error)
	GetByID(postID string, requestingUserID string) (*PostResponse, error)                             // requestingUserID for auth check
	List(requestingUserID string, limit, offset int) ([]*PostResponse, error)                          // requestingUserID for auth check
	ListPostsByUser(targetUserID, requestingUserID string, limit, offset int) ([]*PostResponse, error) // requestingUserID for auth check
	// Update(...) // TODO: Implement Update
	Delete(postID string, requestingUserID string) error // requestingUserID for auth check
}

// postService implements PostService interface
type postService struct {
	postRepo     repositories.PostRepository
	followerRepo repositories.FollowerRepository // Added follower repository
	// authService AuthService // Potentially needed if complex auth logic arises
}

// NewPostService creates a new PostService
func NewPostService(postRepo repositories.PostRepository, followerRepo repositories.FollowerRepository) PostService {
	return &postService{
		postRepo:     postRepo,
		followerRepo: followerRepo,
	}
}

// mapPostToResponse converts a model.Post to a PostResponse DTO
func mapPostToResponse(post *models.Post) *PostResponse {
	if post == nil {
		return nil
	}
	return &PostResponse{
		ID:        post.ID,
		UserID:    post.UserID,
		Title:     post.Title,
		Content:   post.Content,
		ImageURL:  post.ImageURL,
		Privacy:   post.Privacy,
		CreatedAt: post.CreatedAt,
	}
}

// mapPostsToResponse converts a slice of model.Post to a slice of PostResponse DTOs
func mapPostsToResponse(posts []*models.Post) []*PostResponse {
	responses := make([]*PostResponse, len(posts))
	for i, post := range posts {
		responses[i] = mapPostToResponse(post)
	}
	return responses
}

// Create handles the creation of a new post
func (s *postService) Create(request *PostCreateRequest) (*PostResponse, error) {
	// TODO: Add validation using validator library
	if request.Title == "" || request.Content == "" {
		return nil, errors.New("title and content are required")
	}

	// Validate Privacy
	switch request.Privacy {
	case models.PrivacyPublic, models.PrivacyAlmostPrivate:
	// Valid
	case models.PrivacyPrivate:
		if len(request.AllowedUserIDs) == 0 {
			return nil, errors.New("allowed_user_ids are required for private posts")
		}
	// TODO: Optionally validate if AllowedUserIDs actually exist?
	default:
		return nil, errors.New("invalid privacy setting: must be public, almost_private, or private")
	}

	post := &models.Post{
		UserID:   request.UserID, // Assumes UserID is set correctly before calling
		Title:    request.Title,
		Content:  request.Content,
		ImageURL: request.ImageURL,
		Privacy:  request.Privacy,
		// AllowedUsers is handled below
	}

	// Create the post first
	err := s.postRepo.Create(post)
	if err != nil {
		return nil, fmt.Errorf("failed to create post in repository: %w", err)
	}

	// If private, add allowed users
	if post.Privacy == models.PrivacyPrivate {
		err = s.postRepo.AddAllowedUsers(post.ID, request.AllowedUserIDs)
		if err != nil {
			// Log the error, but should we delete the post? Or just return the error?
			// Returning error seems reasonable. The post exists but isn't configured correctly.
			log.Printf("Error adding allowed users for post %s: %v", post.ID, err)
			return nil, fmt.Errorf("failed to add allowed users for private post: %w", err)
		}
	}

	return mapPostToResponse(post), nil
}

// GetByID retrieves a single post, performing authorization checks based on requestingUserID
func (s *postService) GetByID(postID string, requestingUserID string) (*PostResponse, error) {
	post, err := s.postRepo.GetByID(postID)
	if err != nil {
		if errors.Is(err, repositories.ErrPostNotFound) {
			return nil, err // Propagate not found error
		}
		return nil, fmt.Errorf("failed to get post from repository: %w", err)
	}

	// Authorization Check
	canView := false
	isOwner := post.UserID == requestingUserID

	if isOwner {
		canView = true
	} else {
		switch post.Privacy {
		case models.PrivacyPublic:
			canView = true
		case models.PrivacyAlmostPrivate:
			// Check if requestingUser follows post.UserID
			if requestingUserID != "" { // Must be logged in to follow
				follow, err := s.followerRepo.FindFollow(requestingUserID, post.UserID)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					// Log error but treat as not following
					log.Printf("Error checking follow status from %s to %s for post %s: %v", requestingUserID, post.UserID, postID, err)
				} else if follow != nil && follow.Status == "accepted" { // Use string literal
					canView = true
				}
			}
		case models.PrivacyPrivate:
			// Check if requestingUser is in the allowed list
			if requestingUserID != "" { // Must be logged in to be allowed
				allowed, err := s.postRepo.IsUserAllowed(postID, requestingUserID)
				if err != nil {
					// Log error but treat as not allowed
					log.Printf("Error checking if user %s is allowed for post %s: %v", requestingUserID, postID, err)
				} else if allowed {
					canView = true
				}
			}
		default:
			// Unknown privacy setting, treat as private (only owner can view)
			log.Printf("Warning: Post %s has unknown privacy setting '%s'", postID, post.Privacy)
			// canView remains false
		}
	}

	if !canView {
		// Return NotFound to avoid revealing existence of non-public posts
		return nil, repositories.ErrPostNotFound
		// Or return ErrPostForbidden if revealing existence is acceptable:
		// return nil, ErrPostForbidden
	}

	// If private, fetch allowed users (optional, only if needed by frontend when owner views)
	// if isOwner && post.Privacy == models.PrivacyPrivate {
	// 	allowedIDs, err := s.postRepo.GetAllowedUsers(postID)
	// 	if err != nil {
	// 		log.Printf("Warning: Failed to get allowed users for owned private post %s: %v", postID, err)
	// 	} else {
	// 		post.AllowedUsers = allowedIDs // Attach to the model (won't be in JSON response due to tag)
	// 	}
	// }

	return mapPostToResponse(post), nil
}

// List retrieves a list of posts filtered by the repository based on the requesting user's permissions.
func (s *postService) List(requestingUserID string, limit, offset int) ([]*PostResponse, error) {
	posts, err := s.postRepo.List(requestingUserID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list posts from repository: %w", err)
	}
	return mapPostsToResponse(posts), nil
}

// ListPostsByUser retrieves posts for a specific user, filtered by the repository based on the requesting user's permissions.
func (s *postService) ListPostsByUser(targetUserID, requestingUserID string, limit, offset int) ([]*PostResponse, error) {
	posts, err := s.postRepo.ListByUser(targetUserID, requestingUserID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list posts by user from repository: %w", err)
	}
	return mapPostsToResponse(posts), nil
}

// Delete handles the deletion of a post, performing authorization checks
func (s *postService) Delete(postID string, requestingUserID string) error {
	// First, get the post to check ownership
	post, err := s.postRepo.GetByID(postID)
	if err != nil {
		if errors.Is(err, repositories.ErrPostNotFound) {
			return err // Propagate not found error
		}
		return fmt.Errorf("failed to get post for delete check: %w", err)
	}

	// Authorization Check: Only owner can delete
	if post.UserID != requestingUserID {
		return ErrPostForbidden
	}

	// Manually remove allowed users first, as migration 000003 lacks ON DELETE CASCADE
	if post.Privacy == models.PrivacyPrivate {
		allowedUserIDs, err := s.postRepo.GetAllowedUsers(postID)
		if err != nil {
			// Log error but proceed with deletion attempt
			log.Printf("Warning: Failed to get allowed users for post %s before deletion: %v", postID, err)
		} else if len(allowedUserIDs) > 0 {
			err = s.postRepo.RemoveAllowedUsers(postID, allowedUserIDs)
			if err != nil {
				// Log error but proceed with deletion attempt
				log.Printf("Warning: Failed to remove allowed users for post %s before deletion: %v", postID, err)
			}
		}
	}

	// Proceed with post deletion
	err = s.postRepo.Delete(postID)
	if err != nil {
		// Repository already returns ErrPostNotFound if deletion failed due to not found
		return fmt.Errorf("failed to delete post in repository: %w", err)
	}

	return nil
}
