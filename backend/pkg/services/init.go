package services

import (
	"github.com/HASANALI117/social-network/pkg/repositories"
	"github.com/HASANALI117/social-network/pkg/types"
)

// Services holds all service instances.
type Services struct {
	Auth         AuthService
	User         UserService
	Post         PostService
	Group        GroupService
	Follower     FollowerService
	Comment      CommentService
	GroupEvent   GroupEventService   // Added GroupEvent service
	Message      MessageService      // Added Message service
	Notification NotificationService // Added Notification service
}

// InitServices initializes all services.
func InitServices(repos *repositories.Repositories, wsNotifier types.WebSocketNotifier) *Services {
	authService := NewAuthService(repos.User, repos.Session)
	postService := NewPostService(repos.Post, repos.Follower, repos.Group)

	// First create GroupService without notifications
	groupService := &groupService{
		groupRepo: repos.Group,
		userRepo:  repos.User,
	}

	// Create NotificationService with the base group service
	notificationService := NewNotificationService(repos.Notification, groupService, wsNotifier)

	// Now initialize the rest with NotificationService
	groupService.notificationSvc = notificationService // Inject notification service
	followerService := NewFollowerService(repos.Follower, repos.User, notificationService)
	commentService := NewCommentService(repos.Comment, postService, repos.Group)
	groupEventService := NewGroupEventService(repos.GroupEvent, repos.Group, repos.User, repos.GroupEventResponse)
	userService := NewUserService(repos.User, postService, followerService)
	messageService := NewMessageService(repos.ChatMessage, repos.Group)

	return &Services{
		Auth:         authService,
		User:         userService,
		Post:         postService,
		Group:        groupService,
		Follower:     followerService,
		Comment:      commentService,
		GroupEvent:   groupEventService,   // Assign initialized GroupEventService
		Message:      messageService,      // Assign initialized MessageService
		Notification: notificationService, // Assign initialized NotificationService
	}
}
