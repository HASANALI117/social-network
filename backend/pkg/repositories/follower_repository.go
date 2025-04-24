package repositories

import (
	"database/sql"
	"log"

	"github.com/HASANALI117/social-network/pkg/models"
)

// FollowerRepository defines the interface for follower data operations
type FollowerRepository interface {
	CreateFollowRequest(followerID, followingID string) error
	UpdateFollowStatus(followerID, followingID, status string) error
	DeleteFollow(followerID, followingID string) error
	GetFollowers(userID string) ([]models.User, error)
	GetFollowing(userID string) ([]models.User, error)
	GetPendingRequests(userID string) ([]models.User, error)
	FindFollow(followerID, followingID string) (*models.Follower, error)
}

// followerRepository implements FollowerRepository
type followerRepository struct {
	db *sql.DB
}

// NewFollowerRepository creates a new instance of FollowerRepository
func NewFollowerRepository(db *sql.DB) FollowerRepository {
	return &followerRepository{db: db}
}

// CreateFollowRequest inserts a new follow request into the database
func (r *followerRepository) CreateFollowRequest(followerID, followingID string) error {
	query := `INSERT INTO followers (follower_id, following_id, status) VALUES (?, ?, 'pending')`
	_, err := r.db.Exec(query, followerID, followingID)
	if err != nil {
		log.Printf("Error creating follow request: %v", err)
		return err
	}
	return nil
}

// UpdateFollowStatus updates the status of a follow relationship
func (r *followerRepository) UpdateFollowStatus(followerID, followingID, status string) error {
	query := `UPDATE followers SET status = ? WHERE follower_id = ? AND following_id = ?`
	_, err := r.db.Exec(query, status, followerID, followingID)
	if err != nil {
		log.Printf("Error updating follow status: %v", err)
		return err
	}
	return nil
}

// DeleteFollow removes a follow relationship or request
func (r *followerRepository) DeleteFollow(followerID, followingID string) error {
	query := `DELETE FROM followers WHERE follower_id = ? AND following_id = ?`
	_, err := r.db.Exec(query, followerID, followingID)
	if err != nil {
		log.Printf("Error deleting follow: %v", err)
		return err
	}
	return nil
}

// GetFollowers retrieves users who follow the given userID
func (r *followerRepository) GetFollowers(userID string) ([]models.User, error) {
	query := `
        SELECT u.id, u.username, u.email, u.first_name, u.last_name, u.avatar_url, u.about_me, u.birth_date, u.created_at, u.updated_at
        FROM users u
        JOIN followers f ON u.id = f.follower_id
        WHERE f.following_id = ? AND f.status = 'accepted'
    `
	rows, err := r.db.Query(query, userID)
	if err != nil {
		log.Printf("Error getting followers: %v", err)
		return nil, err
	}
	defer rows.Close()

	var followers []models.User
	for rows.Next() {
		var user models.User
		// Note: Ensure Scan order matches SELECT columns exactly
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName, &user.AvatarURL, &user.AboutMe, &user.BirthDate, &user.CreatedAt, &user.UpdatedAt); err != nil {
			log.Printf("Error scanning follower row: %v", err)
			return nil, err
		}
		followers = append(followers, user)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating follower rows: %v", err)
		return nil, err
	}

	return followers, nil
}

// GetFollowing retrieves users whom the given userID follows
func (r *followerRepository) GetFollowing(userID string) ([]models.User, error) {
	query := `
        SELECT u.id, u.username, u.email, u.first_name, u.last_name, u.avatar_url, u.about_me, u.birth_date, u.created_at, u.updated_at
        FROM users u
        JOIN followers f ON u.id = f.following_id
        WHERE f.follower_id = ? AND f.status = 'accepted'
    `
	rows, err := r.db.Query(query, userID)
	if err != nil {
		log.Printf("Error getting following: %v", err)
		return nil, err
	}
	defer rows.Close()

	var following []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName, &user.AvatarURL, &user.AboutMe, &user.BirthDate, &user.CreatedAt, &user.UpdatedAt); err != nil {
			log.Printf("Error scanning following row: %v", err)
			return nil, err
		}
		following = append(following, user)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating following rows: %v", err)
		return nil, err
	}

	return following, nil
}

// GetPendingRequests retrieves users who have sent a follow request to the given userID
func (r *followerRepository) GetPendingRequests(userID string) ([]models.User, error) {
	query := `
        SELECT u.id, u.username, u.email, u.first_name, u.last_name, u.avatar_url, u.about_me, u.birth_date, u.created_at, u.updated_at
        FROM users u
        JOIN followers f ON u.id = f.follower_id
        WHERE f.following_id = ? AND f.status = 'pending'
    `
	rows, err := r.db.Query(query, userID)
	if err != nil {
		log.Printf("Error getting pending requests: %v", err)
		return nil, err
	}
	defer rows.Close()

	var requests []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName, &user.AvatarURL, &user.AboutMe, &user.BirthDate, &user.CreatedAt, &user.UpdatedAt); err != nil {
			log.Printf("Error scanning pending request row: %v", err)
			return nil, err
		}
		requests = append(requests, user)
	}
	if err = rows.Err(); err != nil {
		log.Printf("Error iterating pending request rows: %v", err)
		return nil, err
	}

	return requests, nil
}

// FindFollow retrieves a specific follow relationship or request
func (r *followerRepository) FindFollow(followerID, followingID string) (*models.Follower, error) {
	query := `SELECT follower_id, following_id, status, created_at FROM followers WHERE follower_id = ? AND following_id = ?`
	row := r.db.QueryRow(query, followerID, followingID)

	var follow models.Follower
	err := row.Scan(&follow.FollowerID, &follow.FollowingID, &follow.Status, &follow.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found is not an error in this context
		}
		log.Printf("Error finding follow: %v", err)
		return nil, err
	}
	return &follow, nil
}
