package repositories

import "database/sql"

// Repositories holds all repository instances.
type Repositories struct {
User    UserRepository
Session SessionRepository // Add SessionRepository field
Post    PostRepository    // Add PostRepository field
Group   GroupRepository   // Add GroupRepository field
// TODO: Add other repositories here
}

// InitRepositories initializes all repositories.
func InitRepositories(dbConn *sql.DB) *Repositories {
userRepo := NewUserRepository(dbConn)
sessionRepo := NewSessionRepository(dbConn) // Initialize SessionRepository
postRepo := NewPostRepository(dbConn)       // Initialize PostRepository
groupRepo := NewGroupRepository(dbConn)     // Initialize GroupRepository
// Initialize other repositories...

return &Repositories{
User:    userRepo,
Session: sessionRepo, // Assign initialized SessionRepository
Post:    postRepo,    // Assign initialized PostRepository
Group:   groupRepo,   // Assign initialized GroupRepository
// Assign other initialized repositories...
}
}
