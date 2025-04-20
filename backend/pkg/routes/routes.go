package routes

import (
	"net/http"

	"github.com/HASANALI117/social-network/docs"
	"github.com/HASANALI117/social-network/pkg/handlers"
	"github.com/HASANALI117/social-network/pkg/helpers" // Added import
	httpSwagger "github.com/swaggo/http-swagger"
)

// Setup sets up all API routes
func Setup() http.Handler {
	// Initialize Websocket Hub
	handlers.InitWebsocket()

	// This will ensure the swagger docs are registered
	docs.SwaggerInfo.BasePath = "/api"

	mux := http.NewServeMux()

	// Swagger Documentation
	mux.HandleFunc("/swagger/", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	// Websocket routes
	mux.HandleFunc("/ws", handlers.HandleWebSocket) // Assuming HandleWebSocket doesn't need error wrapping yet

	// Authentication routes (using MakeHandler)
	mux.HandleFunc("/api/auth/signin", helpers.MakeHandler(handlers.SignIn))
	mux.HandleFunc("/api/auth/signout", helpers.MakeHandler(handlers.SignOut))

	// User routes (using MakeHandler)
	mux.HandleFunc("/api/users/register", helpers.MakeHandler(handlers.RegisterUser))
	mux.HandleFunc("/api/users/get", helpers.MakeHandler(handlers.GetUser))
	mux.HandleFunc("/api/users/list", helpers.MakeHandler(handlers.ListUsers))
	mux.HandleFunc("/api/users/update", helpers.MakeHandler(handlers.UpdateUser)) // Assumes UpdateUser is refactored
	mux.HandleFunc("/api/users/delete", helpers.MakeHandler(handlers.DeleteUser))
	mux.HandleFunc("/api/users/online", helpers.MakeHandler(handlers.OnlineUsers))

	// Post routes (using MakeHandler)
	mux.HandleFunc("/api/posts/create", helpers.MakeHandler(handlers.CreatePost))
	mux.HandleFunc("/api/posts/get", helpers.MakeHandler(handlers.GetPost))
	mux.HandleFunc("/api/posts/list", helpers.MakeHandler(handlers.ListPosts))
	mux.HandleFunc("/api/posts/user", helpers.MakeHandler(handlers.ListUserPosts))
	// mux.HandleFunc("/api/posts/update", helpers.MakeHandler(handlers.UpdatePost)) // UpdatePost is commented out in handler
	mux.HandleFunc("/api/posts/delete", helpers.MakeHandler(handlers.DeletePost))

	// Message routes (using MakeHandler)
	mux.HandleFunc("/api/messages", helpers.MakeHandler(handlers.GetMessages))

	// Group routes (using MakeHandler)
	mux.HandleFunc("/api/groups/create", helpers.MakeHandler(handlers.CreateGroup))
	mux.HandleFunc("/api/groups/get", helpers.MakeHandler(handlers.GetGroup))
	mux.HandleFunc("/api/groups/list", helpers.MakeHandler(handlers.ListGroups))
	mux.HandleFunc("/api/groups/update", helpers.MakeHandler(handlers.UpdateGroup))
	mux.HandleFunc("/api/groups/delete", helpers.MakeHandler(handlers.DeleteGroup))

	// Group membership routes (using MakeHandler)
	mux.HandleFunc("/api/groups/members/add", helpers.MakeHandler(handlers.AddGroupMember))
	mux.HandleFunc("/api/groups/members/remove", helpers.MakeHandler(handlers.RemoveGroupMember))
	mux.HandleFunc("/api/groups/members", helpers.MakeHandler(handlers.ListGroupMembers))

	// Group messages routes (using MakeHandler)
	mux.HandleFunc("/api/groups/messages", helpers.MakeHandler(handlers.GetGroupMessages))

	return mux
}
