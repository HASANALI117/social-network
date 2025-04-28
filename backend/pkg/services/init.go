package services

import "github.com/HASANALI117/social-network/pkg/repositories"

// Services holds all service instances.
type Services struct {
	User     UserService
	Auth     AuthService     // Add AuthService field
	Post     PostService     // Add PostService field
	Group    GroupService    // Add GroupService field
	Follower FollowerService // Added FollowerService
	Comment  CommentService  // Added CommentService
	// Add other services here
}

// InitServices initializes all services.
func InitServices(repos *repositories.Repositories) *Services {
	// Initialize services that don't depend on other services first
	authService := NewAuthService(repos.User, repos.Session) // Initialize AuthService, passing User and Session repos
	// Pass GroupRepo to PostService constructor
	postService := NewPostService(repos.Post, repos.Follower, repos.Group)
	groupService := NewGroupService(repos.Group, repos.User)          // Initialize GroupService
	followerService := NewFollowerService(repos.Follower, repos.User) // Initialize FollowerService
	// Pass GroupRepo to CommentService constructor
	commentService := NewCommentService(repos.Comment, postService, repos.Group)

	// Now initialize services that depend on other services
	userService := NewUserService(repos.User, postService, followerService) // Pass dependencies
	// Initialize other services, passing required repositories...

	return &Services{
		User:     userService,
		Auth:     authService,     // Assign initialized AuthService
		Post:     postService,     // Assign initialized PostService
		Group:    groupService,    // Assign initialized GroupService
		Follower: followerService, // Assign initialized FollowerService
		Comment:  commentService,  // Assign initialized CommentService
		// Assign other initialized services...
	}
}
