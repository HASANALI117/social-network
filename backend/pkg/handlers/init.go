package handlers

import "github.com/HASANALI117/social-network/pkg/services"

// Handlers holds all handler instances.
type Handlers struct {
	User *UserHandler
	// Add other handlers here (e.g., Post, Group, Auth, etc.)
}

// InitHandlers initializes all handlers.
func InitHandlers(svc *services.Services) *Handlers {
	userHandler := NewUserHandler(svc.User)
	// Initialize other handlers...

	return &Handlers{
		User: userHandler,
		// Assign other initialized handlers...
	}
}
