package services

import "github.com/HASANALI117/social-network/pkg/repositories"

// Services holds all service instances.
type Services struct {
	Auth       AuthService
	User       UserService
	Post       PostService
	Group      GroupService
	Follower   FollowerService
	Comment    CommentService
	GroupEvent GroupEventService // Added GroupEvent service
}

// InitServices initializes all services.
func InitServices(repos *repositories.Repositories) *Services {
	authService := NewAuthService(repos.User, repos.Session)
	postService := NewPostService(repos.Post, repos.Follower, repos.Group)
	groupService := NewGroupService(repos.Group, repos.User)
	followerService := NewFollowerService(repos.Follower, repos.User)
	commentService := NewCommentService(repos.Comment, postService, repos.Group)
	groupEventService := NewGroupEventService(repos.GroupEvent, repos.Group, repos.User) // Initialize GroupEventService
	userService := NewUserService(repos.User, postService, followerService)

	return &Services{
		Auth:       authService,
		User:       userService,
		Post:       postService,
		Group:      groupService,
		Follower:   followerService,
		Comment:    commentService,
		GroupEvent: groupEventService, // Assign initialized GroupEventService
	}
}
