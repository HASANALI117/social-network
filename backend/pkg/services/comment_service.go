package services

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/HASANALI117/social-network/pkg/models"
	"github.com/HASANALI117/social-network/pkg/repositories"
)

// CommentResponse is the DTO for comment data sent to clients
type CommentResponse struct {
	ID            string    `json:"id"`
	PostID        string    `json:"post_id"`
	UserID        string    `json:"user_id"`
	Content       string    `json:"content"`
	ImageURL      string    `json:"image_url,omitempty"` // New field
	CreatedAt     time.Time `json:"created_at"`
	UserFirstName string    `json:"user_first_name,omitempty"`
	UserLastName  string    `json:"user_last_name,omitempty"`
	UserAvatarURL string    `json:"user_avatar_url,omitempty"`
	Username      string    `json:"username,omitempty"`
}

// CommentCreateRequest is the DTO for creating a new comment
type CommentCreateRequest struct {
	UserID   string `json:"-"` // Set internally from authenticated user
	PostID   string `json:"-"` // Set from URL parameter
	Content  string `json:"content" validate:"required,max=500"`
	ImageURL string `json:"image_url" validate:"omitempty,url"` // New field
}

var (
	ErrCommentForbidden = errors.New("user not authorized to perform this action on the comment")
	ErrCommentNotFound  = repositories.ErrCommentNotFound // Alias for convenience
	ErrPostNotFound     = repositories.ErrPostNotFound    // Alias for convenience
)

// CommentService defines the interface for comment business logic
type CommentService interface {
	CreateComment(request *CommentCreateRequest) (*CommentResponse, error)
	GetCommentsByPost(postID string, requestingUserID string, limit, offset int) ([]*CommentResponse, error)
	DeleteComment(commentID string, requestingUserID string) error
}

// commentService implements CommentService interface
type commentService struct {
	commentRepo repositories.CommentRepository
	postService PostService                  // Use PostService to check post view permissions
	groupRepo   repositories.GroupRepository // Needed for group admin check on delete
	userRepo    repositories.UserRepository  // New dependency
	// authService AuthService // Potentially needed if complex auth logic arises
}

// NewCommentService creates a new CommentService
func NewCommentService(commentRepo repositories.CommentRepository, postService PostService, groupRepo repositories.GroupRepository, userRepo repositories.UserRepository) CommentService {
	return &commentService{
		commentRepo: commentRepo,
		postService: postService,
		groupRepo:   groupRepo,
		userRepo:    userRepo, // Store userRepo
	}
}

// mapCommentToResponse converts a model.Comment to a CommentResponse DTO, enriching with user details
func (s *commentService) mapCommentToResponse(comment *models.Comment) *CommentResponse {
	if comment == nil {
		return nil
	}
	response := &CommentResponse{
		ID:        comment.ID,
		PostID:    comment.PostID,
		UserID:    comment.UserID,
		Content:   comment.Content,
		ImageURL:  comment.ImageURL, // Map ImageURL
		CreatedAt: comment.CreatedAt,
	}

	// Fetch and populate user details
	user, err := s.userRepo.GetByID(comment.UserID)
	if err == nil && user != nil {
		response.UserFirstName = user.FirstName
		response.UserLastName = user.LastName
		response.UserAvatarURL = user.AvatarURL
		response.Username = user.Username // Assuming User model has Username
	} else if err != nil {
		// Log error if user not found, but don't fail the whole comment mapping
		log.Printf("Error fetching user details for comment %s (user %s): %v", comment.ID, comment.UserID, err)
	}
	return response
}

// mapCommentsToResponse converts a slice of model.Comment to a slice of CommentResponse DTOs
func (s *commentService) mapCommentsToResponse(comments []*models.Comment) []*CommentResponse {
	responses := make([]*CommentResponse, len(comments))
	for i, comment := range comments {
		responses[i] = s.mapCommentToResponse(comment) // Call the service's map method
	}
	return responses
}

// CreateComment handles the creation of a new comment
func (s *commentService) CreateComment(request *CommentCreateRequest) (*CommentResponse, error) {
	// 1. Validate input
	if request.Content == "" {
		return nil, errors.New("comment content cannot be empty")
	}
	if len(request.Content) > 500 { // Example length limit
		return nil, errors.New("comment content exceeds maximum length")
	}

	// 2. Check if the user can view the post (implies they can comment)
	// We use GetByID from PostService which includes authorization checks.
	_, err := s.postService.GetByID(request.PostID, request.UserID)
	if err != nil {
		if errors.Is(err, ErrPostNotFound) {
			// If post not found or user cannot view it, they cannot comment
			return nil, ErrPostNotFound // Return NotFound to avoid revealing post existence
		}
		// Handle other potential errors from GetByID
		log.Printf("Error checking post view permission before commenting on post %s by user %s: %v", request.PostID, request.UserID, err)
		return nil, fmt.Errorf("failed to verify post access: %w", err)
	}

	// 3. Create the comment model
	comment := &models.Comment{
		PostID:   request.PostID,
		UserID:   request.UserID, // Assumes UserID is set correctly before calling
		Content:  request.Content,
		ImageURL: request.ImageURL, // Map ImageURL from request
	}

	// 4. Save the comment to the repository
	err = s.commentRepo.Create(comment)
	if err != nil {
		// TODO: Handle specific DB errors like foreign key violation if post was deleted between check and create
		log.Printf("Error creating comment in repository for post %s by user %s: %v", request.PostID, request.UserID, err)
		return nil, fmt.Errorf("failed to save comment: %w", err)
	}

	// 5. Return the response DTO
	return s.mapCommentToResponse(comment), nil
}

// GetCommentsByPost retrieves comments for a post, checking view permissions first
func (s *commentService) GetCommentsByPost(postID string, requestingUserID string, limit, offset int) ([]*CommentResponse, error) {
	// 1. Check if the user can view the post
	_, err := s.postService.GetByID(postID, requestingUserID)
	if err != nil {
		if errors.Is(err, ErrPostNotFound) {
			// If post not found or user cannot view it, they cannot see comments
			return nil, ErrPostNotFound // Return NotFound
		}
		// Handle other potential errors from GetByID
		log.Printf("Error checking post view permission before getting comments for post %s by user %s: %v", postID, requestingUserID, err)
		return nil, fmt.Errorf("failed to verify post access: %w", err)
	}

	// 2. Fetch comments from the repository
	comments, err := s.commentRepo.GetByPostID(postID, limit, offset)
	if err != nil {
		log.Printf("Error getting comments from repository for post %s: %v", postID, err)
		return nil, fmt.Errorf("failed to retrieve comments: %w", err)
	}

	// 3. Map to response DTOs
	return s.mapCommentsToResponse(comments), nil
}

// DeleteComment handles the deletion of a comment, checking ownership or group admin status
func (s *commentService) DeleteComment(commentID string, requestingUserID string) error {
	// 1. Get the comment
	comment, err := s.commentRepo.GetByID(commentID)
	if err != nil {
		if errors.Is(err, ErrCommentNotFound) {
			return ErrCommentNotFound // Propagate not found error
		}
		log.Printf("Error getting comment %s for delete check by user %s: %v", commentID, requestingUserID, err)
		return fmt.Errorf("failed to retrieve comment for deletion: %w", err)
	}

	// 2. Get the parent post to check if it's a group post
	// Use GetByID which includes auth check (user must be able to view post to delete comment)
	post, err := s.postService.GetByID(comment.PostID, requestingUserID)
	if err != nil {
		if errors.Is(err, ErrPostNotFound) {
			// If user can't view the post, they can't delete comments on it
			return ErrPostNotFound // Or ErrCommentForbidden? ErrPostNotFound hides info.
		}
		log.Printf("Error getting parent post %s for comment %s delete check by user %s: %v", comment.PostID, commentID, requestingUserID, err)
		return fmt.Errorf("failed to verify post access for comment deletion: %w", err)
	}

	// 3. Authorization Check
	isAuthorized := false
	isCommentOwner := comment.UserID == requestingUserID

	if post.GroupID != nil && *post.GroupID != "" {
		// --- Group Post Comment Deletion Auth ---
		// Allow comment owner OR group admin
		if isCommentOwner {
			isAuthorized = true
		} else {
			isAdmin, err := s.groupRepo.IsAdmin(*post.GroupID, requestingUserID)
			if err != nil {
				log.Printf("Error checking group admin status for user %s in group %s for comment %s deletion: %v", requestingUserID, *post.GroupID, commentID, err)
				// Treat error as not admin for safety
			} else if isAdmin {
				isAuthorized = true
			}
		}
	} else {
		// --- Non-Group Post Comment Deletion Auth ---
		// Only comment owner can delete
		if isCommentOwner {
			isAuthorized = true
		}
		// TODO: Optionally allow post owner to delete comments?
		// Requires fetching post owner info if not already available in postService.GetByID response.
	}

	if !isAuthorized {
		return ErrCommentForbidden
	}

	// 4. Proceed with deletion
	err = s.commentRepo.Delete(commentID)
	if err != nil {
		// Repository already returns ErrCommentNotFound if deletion failed due to not found
		log.Printf("Error deleting comment %s in repository by user %s: %v", commentID, requestingUserID, err)
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	return nil
}
