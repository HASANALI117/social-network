package handlers

import "github.com/HASANALI117/social-network/pkg/services"

// Handlers holds all handler instances.
type Handlers struct {
	User     *UserHandler
	Auth     *AuthHandler     // Add AuthHandler field
	Group    *GroupHandler    // Add GroupHandler field
	Post     *PostHandler     // Add PostHandler field
	Follower *FollowerHandler // Added FollowerHandler
	// Add other handlers here
}

// InitHandlers initializes all handlers.
func InitHandlers(svc *services.Services) *Handlers {
	authHandler := NewAuthHandler(svc.Auth)                            // Initialize AuthHandler using AuthService from services struct
	groupHandler := NewGroupHandler(svc.Group, svc.Auth)               // Initialize GroupHandler with GroupService and AuthService
	postHandler := NewPostHandler(svc.Post, svc.Auth)                  // Initialize PostHandler
	followerHandler := NewFollowerHandler(svc.Follower, svc.Auth)      // Initialize FollowerHandler with AuthService
	userHandler := NewUserHandler(svc.User, svc.Auth, followerHandler) // Pass FollowerHandler
	// Initialize other handlers...

	return &Handlers{
		User:     userHandler,
		Auth:     authHandler,     // Assign initialized AuthHandler
		Group:    groupHandler,    // Assign initialized GroupHandler
		Post:     postHandler,     // Assign initialized PostHandler
		Follower: followerHandler, // Assign initialized FollowerHandler
		// Assign other initialized handlers...
	}
}
