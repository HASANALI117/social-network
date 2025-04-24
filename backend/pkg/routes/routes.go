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
	// Initialize Websocket Hub
	handlers.InitWebsocket()

	// This will ensure the swagger docs are registered
	docs.SwaggerInfo.BasePath = "/api"

	// --- Dependency Injection ---
	repos := repositories.InitRepositories(dbConn) // Initialize all repositories
	services := services.InitServices(repos)       // Initialize all services using the repositories
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

mux.Handle("/api/users/", httperr.ErrorHandler(controllers.User.ServeHTTP)) // Note the trailing slash for prefix matching
	// Keep OnlineUsers separate for now
mux.HandleFunc("/api/users/online", httperr.ErrorHandler(handlers.OnlineUsers))

// Post routes - Use the PostHandler with prefix matching
mux.Handle("/api/posts/", httperr.ErrorHandler(controllers.Post.ServeHTTP)) // Note the trailing slash

// Message routes
mux.HandleFunc("/api/messages", httperr.ErrorHandler(handlers.GetMessages))

// Group routes - Use the consolidated GroupHandler with prefix matching
mux.Handle("/api/groups/", httperr.ErrorHandler(controllers.Group.ServeHTTP)) // Note the trailing slash

// Remove old separate group member/message routes as they are handled by GroupHandler now
// mux.HandleFunc("/api/groups/members/add", httperr.ErrorHandler(controllers.GroupMember.AddGroupMember))
// mux.HandleFunc("/api/groups/members/remove", httperr.ErrorHandler(controllers.GroupMember.RemoveGroupMember))
// mux.HandleFunc("/api/groups/members", httperr.ErrorHandler(controllers.GroupMember.ListGroupMembers))
// mux.HandleFunc("/api/groups/messages", httperr.ErrorHandler(controllers.GroupMessage.GetGroupMessages))

return mux
}
