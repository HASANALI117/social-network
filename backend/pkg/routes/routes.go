package routes

import (
	"net/http"

	"github.com/HASANALI117/social-network/docs"
	"github.com/HASANALI117/social-network/pkg/handlers"
	"github.com/HASANALI117/social-network/pkg/httperr"
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
	mux.HandleFunc("/ws", handlers.HandleWebSocket)

	// Authentication routes
	mux.HandleFunc("/api/auth/signin", httperr.ErrorHandler(handlers.SignIn))
	mux.HandleFunc("/api/auth/signout", httperr.ErrorHandler(handlers.SignOut))

	// User routes
	mux.HandleFunc("/api/users/register", httperr.ErrorHandler(handlers.RegisterUser))
	mux.HandleFunc("/api/users/get", httperr.ErrorHandler(handlers.GetUser))
	mux.HandleFunc("/api/users/list", httperr.ErrorHandler(handlers.ListUsers))
	mux.HandleFunc("/api/users/update", httperr.ErrorHandler(handlers.UpdateUser))
	mux.HandleFunc("/api/users/delete", httperr.ErrorHandler(handlers.DeleteUser))
	mux.HandleFunc("/api/users/online", httperr.ErrorHandler(handlers.OnlineUsers))

	// Post routes
	mux.HandleFunc("/api/posts/create", httperr.ErrorHandler(handlers.CreatePost))
	mux.HandleFunc("/api/posts/get", httperr.ErrorHandler(handlers.GetPost))
	mux.HandleFunc("/api/posts/list", httperr.ErrorHandler(handlers.ListPosts))
	mux.HandleFunc("/api/posts/user", httperr.ErrorHandler(handlers.ListUserPosts))
	// mux.HandleFunc("/api/posts/update", handlers.u)
	mux.HandleFunc("/api/posts/delete", httperr.ErrorHandler(handlers.DeletePost))

	// Message routes
	mux.HandleFunc("/api/messages", httperr.ErrorHandler(handlers.GetMessages))

	// Group routes
	mux.HandleFunc("/api/groups/create", httperr.ErrorHandler(handlers.CreateGroup))
	mux.HandleFunc("/api/groups/get", httperr.ErrorHandler(handlers.GetGroup))
	mux.HandleFunc("/api/groups/list", httperr.ErrorHandler(handlers.ListGroups))
	mux.HandleFunc("/api/groups/update", httperr.ErrorHandler(handlers.UpdateGroup))
	mux.HandleFunc("/api/groups/delete", httperr.ErrorHandler(handlers.DeleteGroup))

	// Group membership routes
	mux.HandleFunc("/api/groups/members/add", httperr.ErrorHandler(handlers.AddGroupMember))
	mux.HandleFunc("/api/groups/members/remove", httperr.ErrorHandler(handlers.RemoveGroupMember))
	mux.HandleFunc("/api/groups/members", httperr.ErrorHandler(handlers.ListGroupMembers))

	// Group messages routes
	mux.HandleFunc("/api/groups/messages", httperr.ErrorHandler(handlers.GetGroupMessages))

	return mux
}
