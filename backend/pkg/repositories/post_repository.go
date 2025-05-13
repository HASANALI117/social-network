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
	List(requestingUserID string, limit, offset int) ([]*models.Post, error)                     // General feed (non-group posts)
	ListByUser(targetUserID, requestingUserID string, limit, offset int) ([]*models.Post, error) // User profile posts (non-group)
	ListByGroupID(groupID string, limit, offset int) ([]*models.Post, error)                     // Group-specific posts
	ListPublic(limit, offset int) ([]*models.Post, error)                                        // For "Explore" feed
	ListFollowedByUser(requestingUserID string, limit, offset int) ([]*models.Post, error)
	// Update(post *models.Post) error // TODO: Implement Update
	Delete(id string) error

	// Methods for managing allowed users for private posts (Only applicable if post.GroupID is NULL)
	AddAllowedUsers(postID string, userIDs []string) error
	RemoveAllowedUsers(postID string, userIDs []string) error // Optional: For editing allowed list
	IsUserAllowed(postID, userID string) (bool, error)
	GetAllowedUsers(postID string) ([]string, error) // Optional: For editing/display
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
        INSERT INTO posts (id, user_id, title, content, image_url, privacy, group_id, created_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    `
	post.ID = uuid.New().String()
	post.CreatedAt = time.Now()

	// If it's a group post, privacy is implicitly handled by group membership, set to public for simplicity within the group context.
	// If it's not a group post, use the specified privacy.
	privacy := post.Privacy
	if post.GroupID.Valid {
		privacy = models.PrivacyPublic // Group posts are 'public' within the group
	}

	_, err := r.db.Exec(
		query,
		post.ID,
		post.UserID,
		post.Title,
		post.Content,
		post.ImageURL,
		privacy,      // Use determined privacy
		post.GroupID, // Can be NULL
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
        SELECT id, user_id, title, content, image_url, privacy, group_id, created_at
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
		&post.GroupID, // Scan GroupID
		&createdAt,    // Scan into string
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

// List retrieves a paginated list of non-group posts for the general feed,
// filtered by privacy rules based on the requesting user.
func (r *postRepository) List(requestingUserID string, limit, offset int) ([]*models.Post, error) {
	// Base query selects posts based on privacy rules, EXCLUDING group posts
	// 1. Public posts (non-group)
	// 2. User's own posts (non-group)
	// 3. Posts from users the requesting user follows (almost_private, non-group)
	// 4. Private posts where the requesting user is specifically allowed (non-group)
	query := `
SELECT DISTINCT p.id, p.user_id, p.title, p.content, p.image_url, p.privacy, p.group_id, p.created_at
FROM posts p
LEFT JOIN followers f ON p.user_id = f.following_id AND f.follower_id = ? AND f.status = 'accepted' -- requestingUserID for almost_private check
LEFT JOIN post_allowed_users pau ON p.id = pau.post_id AND pau.user_id = ? -- requestingUserID for private check
WHERE
    p.group_id IS NULL -- Exclude group posts
AND (
    p.privacy = ? -- models.PrivacyPublic
    OR p.user_id = ? -- requestingUserID (own posts)
    OR (p.privacy = ? AND f.follower_id IS NOT NULL) -- models.PrivacyAlmostPrivate and follower relationship exists
    OR (p.privacy = ? AND pau.user_id IS NOT NULL) -- models.PrivacyPrivate and user is allowed
)
ORDER BY p.created_at DESC
LIMIT ? OFFSET ?;
`

	rows, err := r.db.Query(query,
		requestingUserID, // For follower check
		requestingUserID, // For allowed user check
		models.PrivacyPublic,
		requestingUserID, // For own post check
		models.PrivacyAlmostPrivate,
		models.PrivacyPrivate,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list non-group posts with privacy filter: %w", err)
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
			&post.GroupID, // Scan GroupID
			&createdAt,
		)
		if err != nil {
			// Log or return error? Return for now.
			return nil, fmt.Errorf("failed to scan post during filtered list: %w", err)
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
		return nil, fmt.Errorf("error iterating filtered post list rows: %w", err)
	}

	return posts, nil
}

// ListByUser retrieves a paginated list of non-group posts for a specific user's profile,
// filtered by privacy rules based on the requesting user.
func (r *postRepository) ListByUser(targetUserID, requestingUserID string, limit, offset int) ([]*models.Post, error) {
	// Similar logic to List, but initially filtered by targetUserID and excludes group posts
	query := `
SELECT DISTINCT p.id, p.user_id, p.title, p.content, p.image_url, p.privacy, p.group_id, p.created_at
FROM posts p
LEFT JOIN followers f ON p.user_id = f.following_id AND f.follower_id = ? AND f.status = 'accepted' -- requestingUserID for almost_private check
LEFT JOIN post_allowed_users pau ON p.id = pau.post_id AND pau.user_id = ? -- requestingUserID for private check
WHERE
    p.user_id = ? -- targetUserID
AND p.group_id IS NULL -- Exclude group posts
AND ( -- Privacy conditions based on requestingUserID
    p.privacy = ? -- models.PrivacyPublic
    OR p.user_id = ? -- requestingUserID (viewing own profile, though filtered by targetUserID already)
    OR (p.privacy = ? AND f.follower_id IS NOT NULL) -- models.PrivacyAlmostPrivate and follower relationship exists
    OR (p.privacy = ? AND pau.user_id IS NOT NULL) -- models.PrivacyPrivate and user is allowed
)
ORDER BY p.created_at DESC
LIMIT ? OFFSET ?;
`

	rows, err := r.db.Query(query,
		requestingUserID, // For follower check
		requestingUserID, // For allowed user check
		targetUserID,     // Filter by post owner
		models.PrivacyPublic,
		requestingUserID, // For own post check (redundant here but harmless)
		models.PrivacyAlmostPrivate,
		models.PrivacyPrivate,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list non-group posts by user with privacy filter: %w", err)
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
			&post.GroupID, // Scan GroupID
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post during filtered list by user: %w", err)
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
		return nil, fmt.Errorf("error iterating filtered post list by user rows: %w", err)
	}

	return posts, nil
}

// ListByGroupID retrieves a paginated list of posts belonging to a specific group.
// Assumes authorization (checking if requesting user is a member) is done in the service layer.
func (r *postRepository) ListByGroupID(groupID string, limit, offset int) ([]*models.Post, error) {
	query := `
        SELECT id, user_id, title, content, image_url, privacy, group_id, created_at
        FROM posts
        WHERE group_id = ?
        ORDER BY created_at DESC
        LIMIT ? OFFSET ?
    `
	rows, err := r.db.Query(query, groupID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list posts by group ID %s: %w", groupID, err)
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
			&post.GroupID, // Scan GroupID
			&createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post during list by group ID %s: %w", groupID, err)
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
		return nil, fmt.Errorf("error iterating post list by group ID %s rows: %w", groupID, err)
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

// --- Methods for post_allowed_users ---

// AddAllowedUsers inserts multiple user IDs allowed to view a specific private post.
// It assumes the post's privacy is already set to 'private'.
func (r *postRepository) AddAllowedUsers(postID string, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil // Nothing to add
	}

	// Use transaction for multiple inserts
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction for adding allowed users: %w", err)
	}
	defer tx.Rollback() // Rollback if anything fails

	stmt, err := tx.Prepare("INSERT OR IGNORE INTO post_allowed_users (post_id, user_id) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement for adding allowed users: %w", err)
	}
	defer stmt.Close()

	for _, userID := range userIDs {
		_, err := stmt.Exec(postID, userID)
		if err != nil {
			// Consider logging the specific user ID that failed
			return fmt.Errorf("failed to insert allowed user %s for post %s: %w", userID, postID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction for adding allowed users: %w", err)
	}

	return nil
}

// RemoveAllowedUsers removes specified user IDs from the allowed list for a post.
func (r *postRepository) RemoveAllowedUsers(postID string, userIDs []string) error {
	if len(userIDs) == 0 {
		return nil // Nothing to remove
	}

	// Use transaction for multiple deletes
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction for removing allowed users: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("DELETE FROM post_allowed_users WHERE post_id = ? AND user_id = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare statement for removing allowed users: %w", err)
	}
	defer stmt.Close()

	for _, userID := range userIDs {
		_, err := stmt.Exec(postID, userID)
		if err != nil {
			// Log or handle error - e.g., user wasn't in the list anyway
			fmt.Printf("Warning: Failed to remove allowed user %s for post %s (may not have existed): %v\n", userID, postID, err)
			// Continue trying to remove others
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction for removing allowed users: %w", err)
	}

	return nil
}

// IsUserAllowed checks if a specific user is in the allowed list for a private post.
func (r *postRepository) IsUserAllowed(postID, userID string) (bool, error) {
	query := "SELECT 1 FROM post_allowed_users WHERE post_id = ? AND user_id = ? LIMIT 1"
	var exists int
	err := r.db.QueryRow(query, postID, userID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil // User is not in the list
		}
		return false, fmt.Errorf("failed to check if user %s is allowed for post %s: %w", userID, postID, err)
	}
	return true, nil // User is in the list
}

// GetAllowedUsers retrieves all user IDs allowed to view a specific private post.
func (r *postRepository) GetAllowedUsers(postID string) ([]string, error) {
	query := "SELECT user_id FROM post_allowed_users WHERE post_id = ?"
	rows, err := r.db.Query(query, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to query allowed users for post %s: %w", postID, err)
	}
	defer rows.Close()

	allowedUserIDs := make([]string, 0)
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan allowed user ID for post %s: %w", postID, err)
		}
		allowedUserIDs = append(allowedUserIDs, userID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating allowed users rows for post %s: %w", postID, err)
	}

	return allowedUserIDs, nil
}

// ListPublic retrieves a paginated list of public, non-group posts.
func (r *postRepository) ListPublic(limit, offset int) ([]*models.Post, error) {
	query := `
		SELECT id, user_id, title, content, image_url, privacy, group_id, created_at
		FROM posts
		WHERE privacy = ? AND group_id IS NULL
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?;
	`
	rows, err := r.db.Query(query, models.PrivacyPublic, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list public posts: %w", err)
	}
	defer rows.Close()

	posts := make([]*models.Post, 0)
	for rows.Next() {
		var post models.Post
		var createdAtStr string
		var groupID sql.NullString // Ensure GroupID is scanned as sql.NullString

		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.ImageURL, // This is string, can be empty
			&post.Privacy,
			&groupID, // Scan into sql.NullString
			&createdAtStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan public post row: %w", err)
		}

		post.GroupID = groupID // Assign scanned NullString

		parsedTime, timeErr := time.Parse(time.RFC3339, createdAtStr)
		if timeErr != nil {
			// Log error and/or decide on fallback. For now, set to zero time.
			fmt.Printf("Warning: Failed to parse post created_at timestamp '%s': %v\n", createdAtStr, timeErr)
			post.CreatedAt = time.Time{}
		} else {
			post.CreatedAt = parsedTime
		}
		posts = append(posts, &post)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating public post rows: %w", err)
	}
	return posts, nil
}

// ListFollowedByUser retrieves posts from users that the requestingUserID follows.
// It includes 'public', 'semi-private' (almost_private), and 'private' posts (if the user is allowed) and excludes group posts.
func (r *postRepository) ListFollowedByUser(requestingUserID string, limit, offset int) ([]*models.Post, error) {
	query := `
		SELECT DISTINCT p.id, p.user_id, p.title, p.content, p.image_url, p.privacy, p.group_id, p.created_at
		FROM posts p
		JOIN followers f ON p.user_id = f.following_id AND f.follower_id = ? AND f.status = 'accepted'
		LEFT JOIN post_allowed_users pau ON p.id = pau.post_id AND pau.user_id = ? -- For checking private post access
		WHERE
		    f.follower_id = ? -- Ensures we are only getting posts from followed users
		  AND f.status = 'accepted'
		  AND p.group_id IS NULL
		  AND (
		    p.privacy = ? -- Public posts from followed user
		    OR p.privacy = ? -- Semi-private posts from followed user
		    OR (p.privacy = ? AND pau.user_id IS NOT NULL) -- Private posts from followed user where requestingUser is allowed
		  )
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?;
	`
	// Parameters for the query:
	// 1. requestingUserID (for JOIN followers f ON p.user_id = f.following_id AND f.follower_id = ?)
	// 2. requestingUserID (for LEFT JOIN post_allowed_users pau ON p.id = pau.post_id AND pau.user_id = ?)
	// 3. requestingUserID (for WHERE f.follower_id = ?)
	// 4. models.PrivacyPublic
	// 5. models.PrivacyAlmostPrivate
	// 6. models.PrivacyPrivate
	// 7. limit
	// 8. offset
	rows, err := r.db.Query(query,
		requestingUserID,
		requestingUserID,
		requestingUserID,
		models.PrivacyPublic,
		models.PrivacyAlmostPrivate,
		models.PrivacyPrivate,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list posts followed by user %s: %w", requestingUserID, err)
	}
	defer rows.Close()

	posts := make([]*models.Post, 0)
	for rows.Next() {
		var post models.Post
		var createdAtStr string
		var groupID sql.NullString // Handles potential NULL group_id

		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.ImageURL,
			&post.Privacy,
			&groupID,
			&createdAtStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post row for followed user feed: %w", err)
		}
		post.GroupID = groupID // Assign scanned NullString
		parsedTime, timeErr := time.Parse(time.RFC3339, createdAtStr)
		if timeErr != nil {
			// Use log package if available, otherwise fmt.Printf
			// log.Printf("Warning: Failed to parse post created_at timestamp '%s' in ListFollowedByUser: %v\n", createdAtStr, timeErr)
			fmt.Printf("Warning: Failed to parse post created_at timestamp '%s' in ListFollowedByUser: %v\n", createdAtStr, timeErr)
			post.CreatedAt = time.Time{}
		} else {
			post.CreatedAt = parsedTime
		}
		posts = append(posts, &post)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating post rows for followed user feed: %w", err)
	}
	return posts, nil
}
