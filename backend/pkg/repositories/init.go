package repositories

import "database/sql"

// Repositories holds all repository instances.
type Repositories struct {
	User     UserRepository
	Session  SessionRepository  // Add SessionRepository field
	Post     PostRepository     // Add PostRepository field
	Group    GroupRepository    // Add GroupRepository field
	Follower FollowerRepository // Added FollowerRepository
	// Add other repository fields here
}

// InitRepositories initializes all repositories.
func InitRepositories(db *sql.DB) *Repositories {
	userRepo := NewUserRepository(db)
	postRepo := NewPostRepository(db)         // Initialize PostRepository
	groupRepo := NewGroupRepository(db)       // Initialize GroupRepository
	sessionRepo := NewSessionRepository(db)   // Initialize SessionRepository
	followerRepo := NewFollowerRepository(db) // Initialize FollowerRepository
	// Initialize other repositories...

	return &Repositories{
		User:     userRepo,
		Post:     postRepo,     // Assign initialized PostRepository
		Group:    groupRepo,    // Assign initialized GroupRepository
		Session:  sessionRepo,  // Assign initialized SessionRepository
		Follower: followerRepo, // Assign initialized FollowerRepository
		// Assign other initialized repositories...
	}
} // Added missing closing brace
