package routes

import (
	"database/sql"
	"net/http"

	"github.com/HASANALI117/social-network/docs"
	"github.com/HASANALI117/social-network/pkg/handlers"
	"github.com/HASANALI117/social-network/pkg/httperr"
	"github.com/HASANALI117/social-network/pkg/repositories" // Import repositories for Init
	"github.com/HASANALI117/social-network/pkg/services"     // Import services for Init
	// "github.com/HASANALI117/social-network/pkg/websocket" // No longer directly needed here, hub is accessed via handlers
	httpSwagger "github.com/swaggo/http-swagger"
)

// Setup sets up all API routes
func Setup(dbConn *sql.DB) http.Handler {
	// Initialize Repositories
	repos := repositories.InitRepositories(dbConn) // Initialize all repositories

	// Initialize Websocket Hub first, as it's needed by NotificationService
	// handlers.InitWebsocket stores the hub in handlers.WebSocketHub
	handlers.InitWebsocket(repos.ChatMessage, nil) // GroupService might not be needed for Hub init if it's only for chat logic
	// If GroupService is truly needed for Hub's core (not just chat), this needs re-evaluation or a different Hub structure.
	// For now, assuming NotificationService needs the Hub (RealTimeNotifier) and GroupService is for chat features within Hub.
	// Let's assume for now that GroupService is not a direct dependency for the Hub's construction for notifications.
	// If it is, we might need to pass a subset of services or rethink the Hub's direct dependencies.
	// A better approach might be for the Hub to not depend on specific services like GroupService directly,
	// but rather on interfaces or have methods that services can call.
	// For now, to proceed, we'll pass nil for GroupService if it's not strictly for the Hub's core functioning
	// related to being a RealTimeNotifier. This might need adjustment if chat functionality breaks.
	// A safer bet, if GroupService is used by Hub's Run() for chat, is to initialize it partially first.
	// However, InitServices takes all repos.

	// Initialize Services, passing the WebSocketHub as the RealTimeNotifier
	// The Hub itself will be handlers.WebSocketHub
	// We need to ensure services.Group is available if handlers.InitWebsocket needs it.
	// This creates a potential ordering issue if Hub needs GroupService and GroupService needs Hub (via NotificationService).

	// Let's adjust the order:
	// 1. Init Repos
	// 2. Init a temporary GroupService if Hub needs it for construction.
	// 3. Init Hub
	// 4. Init all other Services (including full GroupService and NotificationService with Hub)

	// Simpler approach for now: Assume Hub can be constructed without GroupService, or GroupService is passed later if needed for specific hub functions.
	// The `services.InitServices` function now expects a `RealTimeNotifier`.
	// The `handlers.WebSocketHub` is of type `*ws.Hub` (aliased from `websocket.Hub`).
	// `websocket.Hub` implements `services.RealTimeNotifier`.

	// Initialize services, passing the hub.
	// The GroupService instance passed to InitWebsocket might need to be the one from the main services collection.
	// This suggests a potential refactor in Hub's dependencies or initialization sequence.

	// Let's try initializing Hub first, then services.
	// If Hub's NewHub needs GroupService, we have a circular dependency to resolve.
	// From hub.go: NewHub(chatMessageRepo repositories.ChatMessageRepository, groupService services.GroupService)
	// This means GroupService must be created before or alongside the Hub.

	// Revised order:
	// 1. Init Repositories
	// 2. Init GroupService (it doesn't depend on NotificationService or Hub directly for its own construction)
	// 3. Init Hub (passing the created GroupService)
	// 4. Init all Services (passing the Hub to NotificationService, and other services as needed)

	tempGroupService := services.NewGroupService(repos.Group, repos.User, repos.Post, repos.GroupEvent) // Temporary instance for Hub
	handlers.InitWebsocket(repos.ChatMessage, tempGroupService) // Pass the temporary GroupService

	// Now initialize all services, including the "final" GroupService and NotificationService
	allServices := services.InitServices(repos, handlers.WebSocketHub) // Pass the initialized Hub

	// This will ensure the swagger docs are registered
	docs.SwaggerInfo.BasePath = "/api"

	// --- Dependency Injection (Handlers) ---
	// Repositories and Services are already initialized above
	controllers := handlers.InitHandlers(allServices) // Initialize all handlers with all services
	// --- End Dependency Injection ---

	mux := http.NewServeMux()

	// Swagger Documentation
	mux.HandleFunc("/swagger/", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	// Websocket routes - Pass the AuthService instance
	mux.HandleFunc("/ws", handlers.HandleWebSocket(allServices.Auth)) // Pass AuthService from allServices

	// Authentication routes - Use methods from the initialized AuthHandler
	mux.HandleFunc("/api/auth/signin", httperr.ErrorHandler(controllers.Auth.SignIn))
	mux.HandleFunc("/api/auth/signout", httperr.ErrorHandler(controllers.Auth.SignOut))

	// User and Follower routes
	// Register handler for both with and without trailing slash to handle all user routes
	mux.Handle("/api/users", httperr.ErrorHandler(controllers.User.ServeHTTP))  // Handles /api/users/, /api/users/search, /api/users/{id}, and /api/users/{id}/{action}
	mux.Handle("/api/users/", httperr.ErrorHandler(controllers.User.ServeHTTP)) // Handles /api/users/{id} and /api/users/{id}/{action}
	// Specific route for the current user's pending follow requests
	mux.HandleFunc("/api/users/me/follow-requests", controllers.Follower.HandleListPending) // No ErrorHandler wrapper needed

	// Post routes - Use the PostHandler with prefix matching
	mux.Handle("/api/posts", httperr.ErrorHandler(controllers.Post.ServeHTTP)) // Note the trailing slash
	mux.Handle("/api/posts/", httperr.ErrorHandler(controllers.Post.ServeHTTP))

	// Message routes - Use the initialized MessageHandler
	// Message routes - register specific routes before general ones
	mux.HandleFunc("/api/messages/conversations", httperr.ErrorHandler(controllers.Message.GetChatConversations))
	mux.HandleFunc("/api/messages", httperr.ErrorHandler(controllers.Message.GetMessages))

	// Group routes - Use the consolidated GroupHandler with prefix matching
	mux.Handle("/api/groups", httperr.ErrorHandler(controllers.Group.ServeHTTP)) // Note the trailing slash
	mux.Handle("/api/groups/", httperr.ErrorHandler(controllers.Group.ServeHTTP))

	// Comment routes - Use the CommentHandler with prefix matching
	// Handles POST /api/posts/{postId}/comments and GET /api/posts/{postId}/comments via PostHandler's prefix
	// Handles DELETE /api/comments/{commentId}
	mux.Handle("/api/comments", httperr.ErrorHandler(controllers.Comment.ServeHTTP))  // Handles /api/comments/{commentId}
	mux.Handle("/api/comments/", httperr.ErrorHandler(controllers.Comment.ServeHTTP)) // Handles /api/comments/{commentId}

	// Notification routes
	// The NotificationHandler.ServeHTTP method itself returns an error, so it's compatible with httperr.ErrorHandler
	mux.Handle("/api/notifications", httperr.ErrorHandler(controllers.Notification.ServeHTTP))
	mux.Handle("/api/notifications/", httperr.ErrorHandler(controllers.Notification.ServeHTTP))

	return mux
}
