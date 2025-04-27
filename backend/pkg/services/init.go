package services

import "github.com/HASANALI117/social-network/pkg/repositories"

// Services holds all service instances.
type Services struct {
	User     UserService
	Auth     AuthService     // Add AuthService field
	Post     PostService     // Add PostService field
	Group    GroupService    // Add GroupService field
	Follower FollowerService // Added FollowerService
	// Add other services here
}

// InitServices initializes all services.
func InitServices(repos *repositories.Repositories) *Services {
	// Initialize services that don't depend on other services first
	authService := NewAuthService(repos.User, repos.Session)          // Initialize AuthService, passing User and Session repos
	postService := NewPostService(repos.Post, repos.Follower)         // Initialize PostService, passing PostRepo and FollowerRepo
	groupService := NewGroupService(repos.Group, repos.User)          // Initialize GroupService
	followerService := NewFollowerService(repos.Follower, repos.User) // Initialize FollowerService

	// Now initialize services that depend on other services
	userService := NewUserService(repos.User, postService, followerService) // Pass dependencies
	// Initialize other services, passing required repositories...

	return &Services{
		User:     userService,
		Auth:     authService,     // Assign initialized AuthService
		Post:     postService,     // Assign initialized PostService
		Group:    groupService,    // Assign initialized GroupService
		Follower: followerService, // Assign initialized FollowerService
		// Assign other initialized services...
	}
}
