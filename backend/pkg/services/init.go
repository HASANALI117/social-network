package services

import "github.com/HASANALI117/social-network/pkg/repositories"

// Services holds all service instances.
type Services struct {
	Auth               AuthService
	User               UserService
	Post               PostService
	Group              GroupService
	Follower           FollowerService
	Comment            CommentService
	GroupEvent         GroupEventService // Added GroupEvent service
	Message            MessageService    // Added Message service
	Notification       NotificationService // Added Notification service
}

// InitServices initializes all services.
// It now requires a RealTimeNotifier (e.g., the websocket.Hub) for the NotificationService.
func InitServices(repos *repositories.Repositories, notifier RealTimeNotifier) *Services {
	authService := NewAuthService(repos.User, repos.Session)
	postService := NewPostService(repos.Post, repos.Follower, repos.Group, repos.User)
	// Initialize NotificationService first as other services might depend on it
	notificationService := NewNotificationService(repos.Notification, notifier) // Initialize NotificationService
	groupService := NewGroupService(repos.Group, repos.User, repos.Post, repos.GroupEvent, notificationService)
	// NotificationService needs to be initialized before services that depend on it.
	// It's already initialized further down, so we can use it here.
	// followerService := NewFollowerService(repos.Follower, repos.User) // Old call
	commentService := NewCommentService(repos.Comment, postService, repos.Group, repos.User)
	// Update NewGroupEventService to include GroupEventResponseRepository
	groupEventService := NewGroupEventService(repos.GroupEvent, repos.Group, repos.User, repos.GroupEventResponse, notificationService)
	
	// Now initialize services that might depend on NotificationService
	followerService := NewFollowerService(repos.Follower, repos.User, notificationService) // Pass NotificationService
	userService := NewUserService(repos.User, postService, followerService, repos.Group) // Pass GroupRepository
	messageService := NewMessageService(repos.ChatMessage, repos.Group) // Initialize MessageService


	return &Services{
		Auth:               authService,
		User:               userService,
		Post:               postService,
		Group:              groupService,
		Follower:           followerService,
		Comment:            commentService,
		GroupEvent:         groupEventService, // Assign initialized GroupEventService
		Message:            messageService,    // Assign initialized MessageService
		Notification:       notificationService, // Assign initialized NotificationService
	}
}
