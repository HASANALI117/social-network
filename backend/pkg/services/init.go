package services

import "github.com/HASANALI117/social-network/pkg/repositories"

// Services holds all service instances.
type Services struct {
	User UserService
	// Add other services here (e.g., Post, Group, Auth, etc.)
}

// InitServices initializes all services.
func InitServices(repos *repositories.Repositories) *Services {
	userService := NewUserService(repos.User)
	// Initialize other services, passing required repositories...

	return &Services{
		User: userService,
		// Assign other initialized services...
	}
}
