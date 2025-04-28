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
	GroupID   *string   `json:"group_id,omitempty"` // Use pointer for optional field
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	ImageURL  string    `json:"image_url,omitempty"`
	Privacy   string    `json:"privacy"` // Note: For group posts, this might always be 'public' conceptually
	CreatedAt time.Time `json:"created_at"`
	// TODO: Add user details (username, avatar) if needed
}

// PostCreateRequest is the DTO for creating a new post
type PostCreateRequest struct {
	UserID         string   `json:"-"`                  // Set internally from authenticated user
	GroupID        *string  `json:"group_id,omitempty"` // Optional: ID of the group to post in
	Title          string   `json:"title" validate:"required,max=100"`
	Content        string   `json:"content" validate:"required"`
	ImageURL       string   `json:"image_url" validate:"omitempty,url"`
	Privacy        string   `json:"privacy" validate:"required_without=GroupID,omitempty,oneof=public almost_private private"` // Required if not a group post
	AllowedUserIDs []string `json:"allowed_user_ids,omitempty"`                                                                // For 'private' non-group posts
}

var (
	ErrPostForbidden     = errors.New("user not authorized to perform this action on the post")
	ErrGroupAccessDenied = errors.New("user is not a member of the group")
)

// PostService defines the interface for post business logic
type PostService interface {
	Create(request *PostCreateRequest) (*PostResponse, error)
	GetByID(postID string, requestingUserID string) (*PostResponse, error)                              // requestingUserID for auth check
	List(requestingUserID string, limit, offset int) ([]*PostResponse, error)                           // General feed (non-group)
	ListPostsByUser(targetUserID, requestingUserID string, limit, offset int) ([]*PostResponse, error)  // User profile (non-group)
	ListGroupPosts(groupID string, requestingUserID string, limit, offset int) ([]*PostResponse, error) // Group posts
	// Update(...) // TODO: Implement Update
	Delete(postID string, requestingUserID string) error // requestingUserID for auth check
}

// postService implements PostService interface
type postService struct {
	postRepo     repositories.PostRepository
	followerRepo repositories.FollowerRepository // Needed for non-group privacy checks
	groupRepo    repositories.GroupRepository    // Needed for group membership/admin checks
	// authService AuthService // Potentially needed if complex auth logic arises
}

// NewPostService creates a new PostService
func NewPostService(postRepo repositories.PostRepository, followerRepo repositories.FollowerRepository, groupRepo repositories.GroupRepository) PostService {
	return &postService{
		postRepo:     postRepo,
		followerRepo: followerRepo,
		groupRepo:    groupRepo,
	}
}

// mapPostToResponse converts a model.Post to a PostResponse DTO
func mapPostToResponse(post *models.Post) *PostResponse {
	if post == nil {
		return nil
	}
	var groupID *string
	if post.GroupID.Valid {
		groupID = &post.GroupID.String
	}
	return &PostResponse{
		ID:        post.ID,
		UserID:    post.UserID,
		GroupID:   groupID, // Map the GroupID
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

// Create handles the creation of a new post (either user or group post)
func (s *postService) Create(request *PostCreateRequest) (*PostResponse, error) {
	// Basic validation
	if request.Title == "" || request.Content == "" {
		return nil, errors.New("title and content are required")
	}

	post := &models.Post{
		UserID:   request.UserID, // Assumes UserID is set correctly before calling
		Title:    request.Title,
		Content:  request.Content,
		ImageURL: request.ImageURL,
		// GroupID and Privacy are set below
	}

	// Handle Group Post vs User Post
	if request.GroupID != nil && *request.GroupID != "" {
		// --- Group Post ---
		groupID := *request.GroupID
		// Verify user is a member of the group
		isMember, err := s.groupRepo.IsMember(groupID, request.UserID)
		if err != nil {
			log.Printf("Error checking group membership for user %s in group %s: %v", request.UserID, groupID, err)
			return nil, fmt.Errorf("failed to verify group membership: %w", err)
		}
		if !isMember {
			return nil, ErrGroupAccessDenied
		}
		post.GroupID = sql.NullString{String: groupID, Valid: true}
		post.Privacy = models.PrivacyPublic // Group posts are public within the group context
	} else {
		// --- User Post ---
		post.GroupID = sql.NullString{Valid: false} // Ensure GroupID is NULL
		// Validate Privacy for non-group posts
		switch request.Privacy {
		case models.PrivacyPublic, models.PrivacyAlmostPrivate:
			post.Privacy = request.Privacy
		case models.PrivacyPrivate:
			if len(request.AllowedUserIDs) == 0 {
				return nil, errors.New("allowed_user_ids are required for private non-group posts")
			}
			post.Privacy = request.Privacy
		// TODO: Optionally validate if AllowedUserIDs actually exist?
		default:
			return nil, errors.New("invalid privacy setting for non-group post: must be public, almost_private, or private")
		}
	}

	// Create the post in the repository
	err := s.postRepo.Create(post)
	if err != nil {
		return nil, fmt.Errorf("failed to create post in repository: %w", err)
	}

	// If it's a private *user* post, add allowed users
	if !post.GroupID.Valid && post.Privacy == models.PrivacyPrivate {
		err = s.postRepo.AddAllowedUsers(post.ID, request.AllowedUserIDs)
		if err != nil {
			// Log the error, but should we delete the post? Or just return the error?
			// Returning error seems reasonable. The post exists but isn't configured correctly.
			log.Printf("Error adding allowed users for private post %s: %v", post.ID, err)
			// Consider a cleanup mechanism or transaction if this is critical
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

	if post.GroupID.Valid {
		// --- Group Post Authorization ---
		// Check if requestingUser is a member of the group
		if requestingUserID != "" {
			isMember, err := s.groupRepo.IsMember(post.GroupID.String, requestingUserID)
			if err != nil {
				log.Printf("Error checking group membership for user %s in group %s for post %s: %v", requestingUserID, post.GroupID.String, postID, err)
				// Treat error as not being a member for safety
			} else if isMember {
				canView = true
			}
		}
	} else {
		// --- Non-Group Post Authorization ---
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
						log.Printf("Error checking follow status from %s to %s for post %s: %v", requestingUserID, post.UserID, postID, err)
					} else if follow != nil && follow.Status == "accepted" {
						canView = true
					}
				}
			case models.PrivacyPrivate:
				// Check if requestingUser is in the allowed list
				if requestingUserID != "" { // Must be logged in to be allowed
					allowed, err := s.postRepo.IsUserAllowed(postID, requestingUserID)
					if err != nil {
						log.Printf("Error checking if user %s is allowed for post %s: %v", requestingUserID, postID, err)
					} else if allowed {
						canView = true
					}
				}
			default:
				log.Printf("Warning: Post %s has unknown privacy setting '%s'", postID, post.Privacy)
				// canView remains false
			}
		}
	}

	if !canView {
		// Return NotFound to avoid revealing existence of non-public/non-group posts
		return nil, repositories.ErrPostNotFound
	}

	return mapPostToResponse(post), nil
}

// List retrieves a list of non-group posts for the general feed, filtered by the repository.
func (s *postService) List(requestingUserID string, limit, offset int) ([]*PostResponse, error) {
	posts, err := s.postRepo.List(requestingUserID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list non-group posts from repository: %w", err)
	}
	return mapPostsToResponse(posts), nil
}

// ListPostsByUser retrieves non-group posts for a specific user's profile, filtered by the repository.
func (s *postService) ListPostsByUser(targetUserID, requestingUserID string, limit, offset int) ([]*PostResponse, error) {
	posts, err := s.postRepo.ListByUser(targetUserID, requestingUserID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list non-group posts by user from repository: %w", err)
	}
	return mapPostsToResponse(posts), nil
}

// ListGroupPosts retrieves posts belonging to a specific group, checking membership first.
func (s *postService) ListGroupPosts(groupID string, requestingUserID string, limit, offset int) ([]*PostResponse, error) {
	// 1. Check if requesting user is a member of the group
	isMember, err := s.groupRepo.IsMember(groupID, requestingUserID)
	if err != nil {
		log.Printf("Error checking group membership for user %s in group %s: %v", requestingUserID, groupID, err)
		return nil, fmt.Errorf("failed to verify group membership: %w", err)
	}
	if !isMember {
		// Return empty list or specific error? Empty list might be better to avoid revealing group existence.
		// Or return ErrGroupAccessDenied if revealing group existence is okay.
		return []*PostResponse{}, nil // Return empty list if not a member
		// return nil, ErrGroupAccessDenied
	}

	// 2. Fetch posts from the repository
	posts, err := s.postRepo.ListByGroupID(groupID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list posts by group ID %s from repository: %w", groupID, err)
	}

	// 3. Map to response DTOs
	return mapPostsToResponse(posts), nil
}

// Delete handles the deletion of a post, performing authorization checks
func (s *postService) Delete(postID string, requestingUserID string) error {
	// 1. Get the post to check ownership/group admin status
	post, err := s.postRepo.GetByID(postID)
	if err != nil {
		if errors.Is(err, repositories.ErrPostNotFound) {
			return err // Propagate not found error
		}
		return fmt.Errorf("failed to get post for delete check: %w", err)
	}

	// 2. Authorization Check
	isAuthorized := false
	isOwner := post.UserID == requestingUserID

	if post.GroupID.Valid {
		// --- Group Post Deletion Auth ---
		// Check if user is owner OR group admin
		if isOwner {
			isAuthorized = true
		} else {
			isAdmin, err := s.groupRepo.IsAdmin(post.GroupID.String, requestingUserID)
			if err != nil {
				log.Printf("Error checking group admin status for user %s in group %s for post %s deletion: %v", requestingUserID, post.GroupID.String, postID, err)
				// Treat error as not admin for safety
			} else if isAdmin {
				isAuthorized = true
			}
		}
	} else {
		// --- Non-Group Post Deletion Auth ---
		// Only owner can delete
		if isOwner {
			isAuthorized = true
		}
		// Manually remove allowed users first if it's a private non-group post
		// (CASCADE DELETE might not be set up for post_allowed_users)
		if isAuthorized && post.Privacy == models.PrivacyPrivate {
			allowedUserIDs, err := s.postRepo.GetAllowedUsers(postID)
			if err != nil {
				log.Printf("Warning: Failed to get allowed users for private post %s before deletion: %v", postID, err)
			} else if len(allowedUserIDs) > 0 {
				err = s.postRepo.RemoveAllowedUsers(postID, allowedUserIDs)
				if err != nil {
					log.Printf("Warning: Failed to remove allowed users for private post %s before deletion: %v", postID, err)
				}
			}
		}
	}

	if !isAuthorized {
		return ErrPostForbidden
	}

	// 3. Proceed with post deletion
	err = s.postRepo.Delete(postID)
	if err != nil {
		// Repository already returns ErrPostNotFound if deletion failed due to not found
		return fmt.Errorf("failed to delete post in repository: %w", err)
	}

	return nil
}
