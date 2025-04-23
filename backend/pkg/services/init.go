package services

import "github.com/HASANALI117/social-network/pkg/repositories"

// Services holds all service instances.
type Services struct {
User  UserService
Auth  AuthService // Add AuthService field
Post  PostService // Add PostService field
Group GroupService // Add GroupService field
// Add other services here
}

// InitServices initializes all services.
func InitServices(repos *repositories.Repositories) *Services {
userService := NewUserService(repos.User)
authService := NewAuthService(repos.User, repos.Session) // Initialize AuthService, passing User and Session repos
postService := NewPostService(repos.Post /*, repos.User */) // Initialize PostService, passing PostRepo (and UserRepo when needed)
groupService := NewGroupService(repos.Group, repos.User)    // Initialize GroupService
// Initialize other services, passing required repositories...

return &Services{
User:  userService,
Auth:  authService, // Assign initialized AuthService
Post:  postService, // Assign initialized PostService
Group: groupService, // Assign initialized GroupService
// Assign other initialized services...
}
}
