package services

import (
	"errors"
	"fmt"
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
	UserID   string `json:"-"` // Set internally from authenticated user
	Title    string `json:"title" validate:"required,max=100"`
	Content  string `json:"content" validate:"required"`
	ImageURL string `json:"image_url" validate:"omitempty,url"`
	Privacy  string `json:"privacy" validate:"required,oneof=public friends private"`
}

var (
	ErrPostForbidden = errors.New("user not authorized to perform this action on the post")
)

// PostService defines the interface for post business logic
type PostService interface {
	Create(request *PostCreateRequest) (*PostResponse, error)
	GetByID(postID string, userID string) (*PostResponse, error)                                              // userID for auth check
	List(limit, offset int, userID string) ([]*PostResponse, error)                                           // userID for auth check (e.g., friends posts)
	ListPostsByUser(targetUserID string, limit, offset int, requestingUserID string) ([]*PostResponse, error) // Renamed, requestingUserID for auth check
	// Update(...) // TODO: Implement Update
	Delete(postID string, userID string) error // userID for auth check
}

// postService implements PostService interface
type postService struct {
	postRepo repositories.PostRepository
	// userRepo repositories.UserRepository // Needed for relationship checks (friends)
	// authService AuthService // Potentially needed if complex auth logic arises
}

// NewPostService creates a new PostService
func NewPostService(postRepo repositories.PostRepository /*, userRepo repositories.UserRepository*/) PostService {
	return &postService{
		postRepo: postRepo,
		// userRepo: userRepo,
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
	if request.Title == "" || request.Content == "" || request.Privacy == "" {
		return nil, errors.New("title, content, and privacy are required")
	}
	if request.Privacy != "public" && request.Privacy != "friends" && request.Privacy != "private" {
		return nil, errors.New("invalid privacy setting")
	}

	post := &models.Post{
		UserID:   request.UserID, // Assumes UserID is set correctly before calling
		Title:    request.Title,
		Content:  request.Content,
		ImageURL: request.ImageURL,
		Privacy:  request.Privacy,
	}

	err := s.postRepo.Create(post)
	if err != nil {
		return nil, fmt.Errorf("failed to create post in repository: %w", err)
	}

	return mapPostToResponse(post), nil
}

// GetByID retrieves a single post, performing authorization checks
func (s *postService) GetByID(postID string, userID string) (*PostResponse, error) {
	post, err := s.postRepo.GetByID(postID)
	if err != nil {
		if errors.Is(err, repositories.ErrPostNotFound) {
			return nil, err // Propagate not found error
		}
		return nil, fmt.Errorf("failed to get post from repository: %w", err)
	}

	// Authorization Check
	canView := false
	switch post.Privacy {
	case "public":
		canView = true
	case "private":
		if post.UserID == userID {
			canView = true
		}
	case "friends":
		// TODO: Implement friend check using UserRepository
		// isFriend, _ := s.userRepo.AreFriends(post.UserID, userID)
		//
		//	if post.UserID == userID || isFriend {
		//	    canView = true
		//	}
		//
		// Placeholder: Allow owner only for now
		if post.UserID == userID {
			canView = true
		}
	default:
		// Unknown privacy setting, treat as private? Log error?
		if post.UserID == userID {
			canView = true
		}
	}

	if !canView {
		return nil, ErrPostForbidden // Or return ErrPostNotFound to hide existence
	}

	return mapPostToResponse(post), nil
}

// List retrieves a list of posts, potentially filtered by privacy/friendship
func (s *postService) List(limit, offset int, userID string) ([]*PostResponse, error) {
	// TODO: Implement privacy filtering (e.g., only public and friends' posts)
	// This currently fetches all posts regardless of privacy or relationship.
	posts, err := s.postRepo.List(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list posts from repository: %w", err)
	}

	// Filter based on privacy and relationship (Example - needs refinement)
	// visiblePosts := make([]*models.Post, 0)
	// for _, post := range posts {
	//     if post.Privacy == "public" {
	//         visiblePosts = append(visiblePosts, post)
	//     } else if post.UserID == userID {
	//          visiblePosts = append(visiblePosts, post)
	//     } else if post.Privacy == "friends" {
	//         // TODO: Check friendship
	//         // isFriend, _ := s.userRepo.AreFriends(post.UserID, userID)
	//         // if isFriend {
	//         //     visiblePosts = append(visiblePosts, post)
	//         // }
	//     }
	// }
	// return mapPostsToResponse(visiblePosts), nil

	return mapPostsToResponse(posts), nil // Return unfiltered for now
}

// ListPostsByUser retrieves posts for a specific user, checking privacy against the requesting user
func (s *postService) ListPostsByUser(targetUserID string, limit, offset int, requestingUserID string) ([]*PostResponse, error) {
	// TODO: Need ListByUser in PostRepository interface and implementation
	posts, err := s.postRepo.ListByUser(targetUserID, limit, offset) // Assuming repo method exists
	if err != nil {
		return nil, fmt.Errorf("failed to list posts by user from repository: %w", err)
	}

	// If not viewing own posts, filter based on privacy
	if targetUserID != requestingUserID {
		visiblePosts := make([]*models.Post, 0)
		for _, post := range posts {
			if post.Privacy == "public" {
				visiblePosts = append(visiblePosts, post)
			} else if post.Privacy == "friends" {
				// TODO: Check friendship between targetUserID and requestingUserID
				// isFriend, _ := s.userRepo.AreFriends(targetUserID, requestingUserID)
				//
				//	if isFriend {
				//	    visiblePosts = append(visiblePosts, post)
				//	}
			}
			// Private posts are implicitly excluded
		}
		return mapPostsToResponse(visiblePosts), nil
	}

	// Return all posts if viewing own profile
	return mapPostsToResponse(posts), nil
}

// Delete handles the deletion of a post, performing authorization checks
func (s *postService) Delete(postID string, userID string) error {
	// First, get the post to check ownership
	post, err := s.postRepo.GetByID(postID)
	if err != nil {
		if errors.Is(err, repositories.ErrPostNotFound) {
			return err // Propagate not found error
		}
		return fmt.Errorf("failed to get post for delete check: %w", err)
	}

	// Authorization Check: Only owner can delete
	if post.UserID != userID {
		return ErrPostForbidden
	}

	// Proceed with deletion
	err = s.postRepo.Delete(postID)
	if err != nil {
		// Repository already returns ErrPostNotFound if deletion failed due to not found
		return fmt.Errorf("failed to delete post in repository: %w", err)
	}

	return nil
}
