package handlers

import "github.com/HASANALI117/social-network/pkg/services"

// Handlers holds all handler instances.
type Handlers struct {
User         *UserHandler
Auth         *AuthHandler // Add AuthHandler field
Group        *GroupHandler // Add GroupHandler field
GroupMember  *GroupMemberHandler // Add GroupMemberHandler field
GroupMessage *GroupMessageHandler // Add GroupMessageHandler field
Post         *PostHandler         // Add PostHandler field
// Add other handlers here
}

// InitHandlers initializes all handlers.
func InitHandlers(svc *services.Services) *Handlers {
userHandler := NewUserHandler(svc.User, svc.Auth) // Pass both UserService and AuthService
authHandler := NewAuthHandler(svc.Auth) // Initialize AuthHandler using AuthService from services struct
groupHandler := NewGroupHandler(svc.Group, svc.Auth) // Initialize GroupHandler, passing GroupService and AuthService
groupMemberHandler := NewGroupMemberHandler(svc.Auth) // Initialize GroupMemberHandler - TODO: Remove this handler?
groupMessageHandler := NewGroupMessageHandler(svc.Auth) // Initialize GroupMessageHandler - TODO: Remove this handler?
postHandler := NewPostHandler(svc.Post, svc.Auth)       // Initialize PostHandler
// Initialize other handlers...

return &Handlers{
User:         userHandler,
Auth:         authHandler, // Assign initialized AuthHandler
Group:        groupHandler, // Assign initialized GroupHandler
GroupMember:  groupMemberHandler, // Assign initialized GroupMemberHandler - TODO: Remove this handler?
GroupMessage: groupMessageHandler, // Assign initialized GroupMessageHandler - TODO: Remove this handler?
Post:         postHandler,         // Assign initialized PostHandler
// Assign other initialized handlers...
}
}
