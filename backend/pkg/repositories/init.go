package repositories

import (
	"database/sql"
)

// Repositories holds all repository instances.
type Repositories struct {
	User               UserRepository
	Post               PostRepository
	Group              GroupRepository
	Session            SessionRepository
	Follower           FollowerRepository
	Comment            CommentRepository
	GroupEvent         GroupEventRepository         // Added GroupEvent repository
	GroupEventResponse GroupEventResponseRepository // Added GroupEventResponse repository
	ChatMessage        ChatMessageRepository        // Added ChatMessage repository
}

// InitRepositories initializes all repositories.
func InitRepositories(db *sql.DB) *Repositories {
	userRepo := NewUserRepository(db)
	postRepo := NewPostRepository(db)                             // Initialize PostRepository
	groupRepo := NewGroupRepository(db)                           // Initialize GroupRepository
	sessionRepo := NewSessionRepository(db)                       // Initialize SessionRepository
	followerRepo := NewFollowerRepository(db)                     // Initialize FollowerRepository
	commentRepo := NewCommentRepository(db)                       // Initialize CommentRepository
	groupEventRepo := NewGroupEventRepository(db)                 // Initialize GroupEventRepository
	groupEventResponseRepo := NewGroupEventResponseRepository(db) // Initialize GroupEventResponseRepository
	chatMessageRepo := NewChatMessageRepository(db)               // Initialize ChatMessageRepository

	return &Repositories{
		User:               userRepo,
		Post:               postRepo,               // Assign initialized PostRepository
		Group:              groupRepo,              // Assign initialized GroupRepository
		Session:            sessionRepo,            // Assign initialized SessionRepository
		Follower:           followerRepo,           // Assign initialized FollowerRepository
		Comment:            commentRepo,            // Assign initialized CommentRepository
		GroupEvent:         groupEventRepo,         // Assign initialized GroupEventRepository
		GroupEventResponse: groupEventResponseRepo, // Assign initialized GroupEventResponseRepository
		ChatMessage:        chatMessageRepo,        // Assign initialized ChatMessageRepository
	}
}
