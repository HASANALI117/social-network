package routes

import (
	"net/http"

	"github.com/HASANALI117/social-network/docs"
	"github.com/HASANALI117/social-network/pkg/handlers"
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
	mux.HandleFunc("/api/auth/signin", handlers.SignIn)
	mux.HandleFunc("/api/auth/signout", handlers.SignOut)

	// User routes
	mux.HandleFunc("/api/users/register", handlers.RegisterUser)
	mux.HandleFunc("/api/users/get", handlers.GetUser)
	mux.HandleFunc("/api/users/list", handlers.ListUsers)
	mux.HandleFunc("/api/users/update", handlers.UpdateUser)
	mux.HandleFunc("/api/users/delete", handlers.DeleteUser)

	// Post routes
	mux.HandleFunc("/api/posts/create", handlers.CreatePost)
	mux.HandleFunc("/api/posts/get", handlers.GetPost)
	mux.HandleFunc("/api/posts/list", handlers.ListPosts)
	mux.HandleFunc("/api/posts/user", handlers.ListUserPosts)
	// mux.HandleFunc("/api/posts/update", handlers.u)
	mux.HandleFunc("/api/posts/delete", handlers.DeletePost)

	return mux
}
