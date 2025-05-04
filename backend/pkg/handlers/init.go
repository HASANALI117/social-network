package handlers

import "github.com/HASANALI117/social-network/pkg/services"

// Handlers holds all handler instances.
type Handlers struct {
	User         *UserHandler
	Auth         *AuthHandler         // Add AuthHandler field
	Group        *GroupHandler        // Add GroupHandler field
	Post         *PostHandler         // Add PostHandler field
	Follower     *FollowerHandler     // Added FollowerHandler
	Comment      *CommentHandler      // Added CommentHandler
	Message      *MessageHandler      // Added MessageHandler
	GroupMessage *GroupMessageHandler // Added GroupMessageHandler
	GroupMember  *GroupMemberHandler  // Added GroupMemberHandler
	// TODO: Add handlers for GroupInvite, GroupJoinReq, GroupEvent, GroupEventRes later
}

// InitHandlers initializes all handlers.
func InitHandlers(svc *services.Services) *Handlers {
	authHandler := NewAuthHandler(svc.Auth) // Initialize AuthHandler using AuthService from services struct
	// Pass PostService to GroupHandler constructor
	groupHandler := NewGroupHandler(svc.Group, svc.Post, svc.Auth, svc.GroupEvent, svc.Message) // Pass MessageService
	followerHandler := NewFollowerHandler(svc.Follower, svc.Auth)                               // Initialize FollowerHandler with AuthService
	commentHandler := NewCommentHandler(svc.Comment, svc.Auth)                                  // Initialize CommentHandler
	postHandler := NewPostHandler(svc.Post, svc.Auth, commentHandler)                           // Initialize PostHandler, passing CommentHandler
	userHandler := NewUserHandler(svc.User, svc.Auth, followerHandler)                          // Pass FollowerHandler
	messageHandler := NewMessageHandler(svc.Message, svc.Auth)                                  // Initialize MessageHandler
	groupMessageHandler := NewGroupMessageHandler(svc.Message, svc.Group, svc.Auth)             // Initialize GroupMessageHandler
	groupMemberHandler := NewGroupMemberHandler(svc.Group, svc.Auth)                            // Initialize GroupMemberHandler

	return &Handlers{
		User:         userHandler,
		Auth:         authHandler,         // Assign initialized AuthHandler
		Group:        groupHandler,        // Assign initialized GroupHandler
		Post:         postHandler,         // Assign initialized PostHandler
		Follower:     followerHandler,     // Assign initialized FollowerHandler
		Comment:      commentHandler,      // Assign initialized CommentHandler
		Message:      messageHandler,      // Assign initialized MessageHandler
		GroupMessage: groupMessageHandler, // Assign initialized GroupMessageHandler
		GroupMember:  groupMemberHandler,  // Assign initialized GroupMemberHandler
	}
}
