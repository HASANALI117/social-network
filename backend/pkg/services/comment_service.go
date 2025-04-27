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
ID        string    `json:"id"`
PostID    string    `json:"post_id"`
UserID    string    `json:"user_id"`
Content   string    `json:"content"`
CreatedAt time.Time `json:"created_at"`
// TODO: Add user details (username, avatar) if needed
}

// CommentCreateRequest is the DTO for creating a new comment
type CommentCreateRequest struct {
UserID  string `json:"-"` // Set internally from authenticated user
PostID  string `json:"-"` // Set from URL parameter
Content string `json:"content" validate:"required,max=500"`
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
postService PostService // Use PostService to check post view permissions
// authService AuthService // Potentially needed if complex auth logic arises
}

// NewCommentService creates a new CommentService
func NewCommentService(commentRepo repositories.CommentRepository, postService PostService) CommentService {
return &commentService{
commentRepo: commentRepo,
postService: postService,
}
}

// mapCommentToResponse converts a model.Comment to a CommentResponse DTO
func mapCommentToResponse(comment *models.Comment) *CommentResponse {
if comment == nil {
return nil
}
return &CommentResponse{
ID:        comment.ID,
PostID:    comment.PostID,
UserID:    comment.UserID,
Content:   comment.Content,
CreatedAt: comment.CreatedAt,
}
}

// mapCommentsToResponse converts a slice of model.Comment to a slice of CommentResponse DTOs
func mapCommentsToResponse(comments []*models.Comment) []*CommentResponse {
responses := make([]*CommentResponse, len(comments))
for i, comment := range comments {
responses[i] = mapCommentToResponse(comment)
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
PostID:  request.PostID,
UserID:  request.UserID, // Assumes UserID is set correctly before calling
Content: request.Content,
}

// 4. Save the comment to the repository
err = s.commentRepo.Create(comment)
if err != nil {
// TODO: Handle specific DB errors like foreign key violation if post was deleted between check and create
log.Printf("Error creating comment in repository for post %s by user %s: %v", request.PostID, request.UserID, err)
return nil, fmt.Errorf("failed to save comment: %w", err)
}

// 5. Return the response DTO
return mapCommentToResponse(comment), nil
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
return mapCommentsToResponse(comments), nil
}

// DeleteComment handles the deletion of a comment, checking ownership
func (s *commentService) DeleteComment(commentID string, requestingUserID string) error {
// 1. Get the comment to check ownership
comment, err := s.commentRepo.GetByID(commentID)
if err != nil {
if errors.Is(err, ErrCommentNotFound) {
return ErrCommentNotFound // Propagate not found error
}
log.Printf("Error getting comment %s for delete check by user %s: %v", commentID, requestingUserID, err)
return fmt.Errorf("failed to retrieve comment for deletion: %w", err)
}

// 2. Authorization Check: Only the comment owner can delete
// TODO: Consider allowing post owner to delete comments as well? Requires fetching post owner.
if comment.UserID != requestingUserID {
return ErrCommentForbidden
}

// 3. Proceed with deletion
err = s.commentRepo.Delete(commentID)
if err != nil {
// Repository already returns ErrCommentNotFound if deletion failed due to not found
log.Printf("Error deleting comment %s in repository by user %s: %v", commentID, requestingUserID, err)
return fmt.Errorf("failed to delete comment: %w", err)
}

return nil
}
