package repositories

import (
"database/sql"
"errors"
"fmt"
"time"

"github.com/HASANALI117/social-network/pkg/models"
"github.com/google/uuid"
)

var (
// ErrCommentNotFound indicates that a comment with the given ID was not found.
ErrCommentNotFound = errors.New("comment not found")
)

// CommentRepository defines the interface for comment data access
type CommentRepository interface {
Create(comment *models.Comment) error
GetByID(id string) (*models.Comment, error)
GetByPostID(postID string, limit, offset int) ([]*models.Comment, error)
Delete(id string) error
// TODO: Consider adding an Update method if needed
}

// commentRepository implements CommentRepository interface
type commentRepository struct {
db *sql.DB
}

// NewCommentRepository creates a new CommentRepository
func NewCommentRepository(db *sql.DB) CommentRepository {
return &commentRepository{
db: db,
}
}

// Create inserts a new comment record into the database
func (r *commentRepository) Create(comment *models.Comment) error {
query := `
       INSERT INTO comments (id, post_id, user_id, content, image_url, created_at)
       VALUES (?, ?, ?, ?, ?, ?)
   `
comment.ID = uuid.New().String()
comment.CreatedAt = time.Now()

_, err := r.db.Exec(
	query,
	comment.ID,
	comment.PostID,
	comment.UserID,
	comment.Content,
	comment.ImageURL, // New parameter
	comment.CreatedAt,
)
if err != nil {
// TODO: Handle potential foreign key constraint errors (e.g., post_id doesn't exist)
return fmt.Errorf("failed to create comment: %w", err)
}
return nil
}

// GetByID retrieves a comment by its ID
func (r *commentRepository) GetByID(id string) (*models.Comment, error) {
query := `
       SELECT id, post_id, user_id, content, image_url, created_at
       FROM comments
       WHERE id = ?
   `
var comment models.Comment
var createdAt string // Scan as string first

err := r.db.QueryRow(query, id).Scan(
	&comment.ID,
	&comment.PostID,
	&comment.UserID,
	&comment.Content,
	&comment.ImageURL, // New field to scan
	&createdAt,
)
if err != nil {
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrCommentNotFound
	}
	return nil, fmt.Errorf("failed to get comment by ID: %w", err)
}

// Parse timestamp
// Custom layout for "YYYY-MM-DD HH:MM:SS.FFFFFFFFF+ZZ:ZZ"
const customTimeLayout = "2006-01-02 15:04:05.999999999Z07:00"
comment.CreatedAt, err = time.Parse(customTimeLayout, createdAt)
if err != nil {
	fmt.Printf("Warning: Failed to parse comment created_at timestamp '%s' with layout '%s': %v\n", createdAt, customTimeLayout, err)
	comment.CreatedAt = time.Time{}
}

return &comment, nil
}

// GetByPostID retrieves a paginated list of comments for a specific post
func (r *commentRepository) GetByPostID(postID string, limit, offset int) ([]*models.Comment, error) {
query := `
       SELECT id, post_id, user_id, content, image_url, created_at
       FROM comments
       WHERE post_id = ?
       ORDER BY created_at ASC -- Or DESC depending on desired order
       LIMIT ? OFFSET ?
   `
rows, err := r.db.Query(query, postID, limit, offset)
if err != nil {
return nil, fmt.Errorf("failed to get comments by post ID %s: %w", postID, err)
}
defer rows.Close()

comments := make([]*models.Comment, 0)
for rows.Next() {
var comment models.Comment
var createdAt string
err := rows.Scan(
	&comment.ID,
	&comment.PostID,
	&comment.UserID,
	&comment.Content,
	&comment.ImageURL, // New field to scan
	&createdAt,
)
if err != nil {
	return nil, fmt.Errorf("failed to scan comment for post ID %s: %w", postID, err)
}
// Parse timestamp
// Custom layout for "YYYY-MM-DD HH:MM:SS.FFFFFFFFF+ZZ:ZZ"
const customTimeLayout = "2006-01-02 15:04:05.999999999Z07:00"
comment.CreatedAt, err = time.Parse(customTimeLayout, createdAt)
if err != nil {
	fmt.Printf("Warning: Failed to parse comment created_at timestamp '%s' with layout '%s': %v\n", createdAt, customTimeLayout, err)
	comment.CreatedAt = time.Time{}
}
comments = append(comments, &comment)
}

if err := rows.Err(); err != nil {
return nil, fmt.Errorf("error iterating comment rows for post ID %s: %w", postID, err)
}

return comments, nil
}

// Delete removes a comment record by its ID
func (r *commentRepository) Delete(id string) error {
query := "DELETE FROM comments WHERE id = ?"
result, err := r.db.Exec(query, id)
if err != nil {
return fmt.Errorf("failed to delete comment: %w", err)
}

rowsAffected, err := result.RowsAffected()
if err != nil {
return fmt.Errorf("failed to get rows affected after deleting comment: %w", err)
}

if rowsAffected == 0 {
return ErrCommentNotFound // Return error if no rows were deleted
}

return nil
}
