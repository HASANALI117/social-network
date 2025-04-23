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
// ErrPostNotFound indicates that a post with the given ID was not found.
ErrPostNotFound = errors.New("post not found")
)

// PostRepository defines the interface for post data access
type PostRepository interface {
Create(post *models.Post) error
GetByID(id string) (*models.Post, error)
List(limit, offset int) ([]*models.Post, error)
ListByUser(userID string, limit, offset int) ([]*models.Post, error)
// Update(post *models.Post) error // TODO: Implement Update
Delete(id string) error
}

// postRepository implements PostRepository interface
type postRepository struct {
db *sql.DB
}

// NewPostRepository creates a new PostRepository
func NewPostRepository(db *sql.DB) PostRepository {
return &postRepository{
db: db,
}
}

// Create inserts a new post record into the database
func (r *postRepository) Create(post *models.Post) error {
query := `
        INSERT INTO posts (id, user_id, title, content, image_url, privacy, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `
post.ID = uuid.New().String()
post.CreatedAt = time.Now()

_, err := r.db.Exec(
query,
post.ID,
post.UserID,
post.Title,
post.Content,
post.ImageURL,
post.Privacy,
post.CreatedAt,
)
if err != nil {
return fmt.Errorf("failed to create post: %w", err)
}
return nil
}

// GetByID retrieves a post by its ID
func (r *postRepository) GetByID(id string) (*models.Post, error) {
query := `
        SELECT id, user_id, title, content, image_url, privacy, created_at
        FROM posts
        WHERE id = ?
    `
var post models.Post
var createdAt string // Scan as string first

err := r.db.QueryRow(query, id).Scan(
&post.ID,
&post.UserID,
&post.Title,
&post.Content,
&post.ImageURL,
&post.Privacy,
&createdAt, // Scan into string
)
if err != nil {
if errors.Is(err, sql.ErrNoRows) {
return nil, ErrPostNotFound
}
return nil, fmt.Errorf("failed to get post by ID: %w", err)
}

// Parse timestamp
post.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
if err != nil {
// Log parsing error but return the post anyway? Or return error?
fmt.Printf("Warning: Failed to parse post created_at timestamp '%s': %v\n", createdAt, err)
// Decide on error handling strategy. For now, return post with zero time.
post.CreatedAt = time.Time{}
}

return &post, nil
}

// List retrieves a paginated list of all posts (consider privacy later)
func (r *postRepository) List(limit, offset int) ([]*models.Post, error) {
query := `
        SELECT id, user_id, title, content, image_url, privacy, created_at
        FROM posts
        ORDER BY created_at DESC
        LIMIT ? OFFSET ?
    `
rows, err := r.db.Query(query, limit, offset)
if err != nil {
return nil, fmt.Errorf("failed to list posts: %w", err)
}
defer rows.Close()

posts := make([]*models.Post, 0)
for rows.Next() {
var post models.Post
var createdAt string
err := rows.Scan(
&post.ID,
&post.UserID,
&post.Title,
&post.Content,
&post.ImageURL,
&post.Privacy,
&createdAt,
)
if err != nil {
return nil, fmt.Errorf("failed to scan post during list: %w", err)
}
// Parse timestamp
post.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
if err != nil {
fmt.Printf("Warning: Failed to parse post created_at timestamp '%s': %v\n", createdAt, err)
post.CreatedAt = time.Time{}
}
posts = append(posts, &post)
}

if err := rows.Err(); err != nil {
return nil, fmt.Errorf("error iterating post list rows: %w", err)
}

return posts, nil
}

// ListByUser retrieves a paginated list of posts for a specific user
func (r *postRepository) ListByUser(userID string, limit, offset int) ([]*models.Post, error) {
query := `
        SELECT id, user_id, title, content, image_url, privacy, created_at
        FROM posts
        WHERE user_id = ?
        ORDER BY created_at DESC
        LIMIT ? OFFSET ?
    `
rows, err := r.db.Query(query, userID, limit, offset)
if err != nil {
return nil, fmt.Errorf("failed to list posts by user: %w", err)
}
defer rows.Close()

posts := make([]*models.Post, 0)
for rows.Next() {
var post models.Post
var createdAt string
err := rows.Scan(
&post.ID,
&post.UserID,
&post.Title,
&post.Content,
&post.ImageURL,
&post.Privacy,
&createdAt,
)
if err != nil {
return nil, fmt.Errorf("failed to scan post during list by user: %w", err)
}
// Parse timestamp
post.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
if err != nil {
fmt.Printf("Warning: Failed to parse post created_at timestamp '%s': %v\n", createdAt, err)
post.CreatedAt = time.Time{}
}
posts = append(posts, &post)
}

if err := rows.Err(); err != nil {
return nil, fmt.Errorf("error iterating post list by user rows: %w", err)
}

return posts, nil
}

// Delete removes a post record by its ID
func (r *postRepository) Delete(id string) error {
query := "DELETE FROM posts WHERE id = ?"
result, err := r.db.Exec(query, id)
if err != nil {
return fmt.Errorf("failed to delete post: %w", err)
}

rowsAffected, err := result.RowsAffected()
if err != nil {
return fmt.Errorf("failed to get rows affected after deleting post: %w", err)
}

if rowsAffected == 0 {
return ErrPostNotFound // Return error if no rows were deleted
}

return nil
}
