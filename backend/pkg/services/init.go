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
	Message    MessageService    // Added Message service
}

// InitServices initializes all services.
func InitServices(repos *repositories.Repositories) *Services {
	authService := NewAuthService(repos.User, repos.Session)
	postService := NewPostService(repos.Post, repos.Follower, repos.Group, repos.User)
	groupService := NewGroupService(repos.Group, repos.User)
	followerService := NewFollowerService(repos.Follower, repos.User)
	commentService := NewCommentService(repos.Comment, postService, repos.Group, repos.User)
	// Update NewGroupEventService to include GroupEventResponseRepository
	groupEventService := NewGroupEventService(repos.GroupEvent, repos.Group, repos.User, repos.GroupEventResponse)
	userService := NewUserService(repos.User, postService, followerService)
	messageService := NewMessageService(repos.ChatMessage, repos.Group) // Initialize MessageService

	return &Services{
		Auth:       authService,
		User:       userService,
		Post:       postService,
		Group:      groupService,
		Follower:   followerService,
		Comment:    commentService,
		GroupEvent: groupEventService, // Assign initialized GroupEventService
		Message:    messageService,    // Assign initialized MessageService
	}
}
