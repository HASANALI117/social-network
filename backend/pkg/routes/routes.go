package routes

import (
	"database/sql"
	"net/http"

	"github.com/HASANALI117/social-network/docs"
	"github.com/HASANALI117/social-network/pkg/handlers"
	"github.com/HASANALI117/social-network/pkg/httperr"
	"github.com/HASANALI117/social-network/pkg/repositories" // Import repositories for Init
	"github.com/HASANALI117/social-network/pkg/services"     // Import services for Init
	httpSwagger "github.com/swaggo/http-swagger"
)

// Setup sets up all API routes
func Setup(dbConn *sql.DB) http.Handler {
	// Initialize Repositories and Services first
	repos := repositories.InitRepositories(dbConn) // Initialize all repositories
	services := services.InitServices(repos)       // Initialize all services using the repositories

	// Initialize Websocket Hub with required repository and service
	handlers.InitWebsocket(repos.ChatMessage, services.Group) // Pass GroupService

	// This will ensure the swagger docs are registered
	docs.SwaggerInfo.BasePath = "/api"

	// --- Dependency Injection (Handlers) ---
	// Repositories and Services are already initialized above
	controllers := handlers.InitHandlers(services) // Initialize all handlers
	// --- End Dependency Injection ---

	mux := http.NewServeMux()

	// Swagger Documentation
	mux.HandleFunc("/swagger/", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	// Websocket routes - Pass the AuthService instance
	mux.HandleFunc("/ws", handlers.HandleWebSocket(services.Auth))

	// Authentication routes - Use methods from the initialized AuthHandler
	mux.HandleFunc("/api/auth/signin", httperr.ErrorHandler(controllers.Auth.SignIn))
	mux.HandleFunc("/api/auth/signout", httperr.ErrorHandler(controllers.Auth.SignOut))

	// User and Follower routes
	// Register handler for both with and without trailing slash to handle all user routes
	mux.Handle("/api/users", httperr.ErrorHandler(controllers.User.ServeHTTP))  // Handles /api/users/, /api/users/{id}, and /api/users/{id}/{action}
	mux.Handle("/api/users/", httperr.ErrorHandler(controllers.User.ServeHTTP)) // Handles /api/users/{id} and /api/users/{id}/{action}
	// Specific route for the current user's pending follow requests
	mux.HandleFunc("/api/users/me/follow-requests", controllers.Follower.HandleListPending) // No ErrorHandler wrapper needed

	// Post routes - Use the PostHandler with prefix matching
	mux.Handle("/api/posts", httperr.ErrorHandler(controllers.Post.ServeHTTP)) // Note the trailing slash
	mux.Handle("/api/posts/", httperr.ErrorHandler(controllers.Post.ServeHTTP))

	// Message routes - Use the initialized MessageHandler
	mux.HandleFunc("/api/messages", httperr.ErrorHandler(controllers.Message.GetMessages))
	mux.HandleFunc("/api/messages/", httperr.ErrorHandler(controllers.Message.GetMessages)) // Handles /api/messages/{id}

	// Group routes - Use the consolidated GroupHandler with prefix matching
	mux.Handle("/api/groups", httperr.ErrorHandler(controllers.Group.ServeHTTP)) // Note the trailing slash
	mux.Handle("/api/groups/", httperr.ErrorHandler(controllers.Group.ServeHTTP))

	// Comment routes - Use the CommentHandler with prefix matching
	// Handles POST /api/posts/{postId}/comments and GET /api/posts/{postId}/comments via PostHandler's prefix
	// Handles DELETE /api/comments/{commentId}
	mux.Handle("/api/comments", httperr.ErrorHandler(controllers.Comment.ServeHTTP))  // Handles /api/comments/{commentId}
	mux.Handle("/api/comments/", httperr.ErrorHandler(controllers.Comment.ServeHTTP)) // Handles /api/comments/{commentId}

	return mux
}
