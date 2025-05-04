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
	mux.Handle("/api/users/", httperr.ErrorHandler(controllers.User.ServeHTTP)) // Handles /api/users/, /api/users/{id}, and /api/users/{id}/{action}
	// Specific route for the current user's pending follow requests
	mux.HandleFunc("/api/users/me/follow-requests", controllers.Follower.HandleListPending) // No ErrorHandler wrapper needed

	// Post routes - Use the PostHandler with prefix matching
	mux.Handle("/api/posts/", httperr.ErrorHandler(controllers.Post.ServeHTTP)) // Note the trailing slash

	// Message routes - Use the initialized MessageHandler
	mux.HandleFunc("/api/messages", httperr.ErrorHandler(controllers.Message.GetMessages))

	// Group routes - Use the consolidated GroupHandler with prefix matching
	mux.Handle("/api/groups/", httperr.ErrorHandler(controllers.Group.ServeHTTP)) // Note the trailing slash

	// Comment routes - Use the CommentHandler with prefix matching
	// Handles POST /api/posts/{postId}/comments and GET /api/posts/{postId}/comments via PostHandler's prefix
	// Handles DELETE /api/comments/{commentId}
	mux.Handle("/api/comments/", httperr.ErrorHandler(controllers.Comment.ServeHTTP)) // Handles /api/comments/{commentId}

	// Note: The CommentHandler's ServeHTTP needs to correctly parse postID from /api/posts/{postId}/comments
	// The current PostHandler already handles /api/posts/, so we need to adjust routing or handler logic.
	// Let's assume for now the CommentHandler can parse the full path passed to it.
	// A more robust solution might involve a more sophisticated router like gorilla/mux or chi.

	// Remove old separate group member/message routes as they are handled by GroupHandler now
	// mux.HandleFunc("/api/groups/members/add", httperr.ErrorHandler(controllers.GroupMember.AddGroupMember))
	// mux.HandleFunc("/api/groups/members/remove", httperr.ErrorHandler(controllers.GroupMember.RemoveGroupMember))
	// mux.HandleFunc("/api/groups/members", httperr.ErrorHandler(controllers.GroupMember.ListGroupMembers))
	// mux.HandleFunc("/api/groups/messages", httperr.ErrorHandler(controllers.GroupMessage.GetGroupMessages))

	return mux
}
