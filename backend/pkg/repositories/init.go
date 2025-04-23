package repositories

import "database/sql"

// Repositories holds all repository instances.
type Repositories struct {
	User UserRepository
	// TODO: Add other repositories here (e.g., Post, Group, etc.)
}

// InitRepositories initializes all repositories.
func InitRepositories(dbConn *sql.DB) *Repositories {
	userRepo := NewUserRepository(dbConn)
	// Initialize other repositories...

	return &Repositories{
		User: userRepo,
		// Assign other initialized repositories...
	}
}
