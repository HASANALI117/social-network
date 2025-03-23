package routes

import (
	"net/http"

	"social-network/docs"
	"social-network/pkg/handlers"
	"social-network/pkg/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

func Setup() http.Handler {
	docs.SwaggerInfo.BasePath = "/api"
	mux := http.NewServeMux()

	// Swagger Documentation
	mux.HandleFunc("/swagger/", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	// comments routes
	mux.HandleFunc("/api/comments/create", middleware.Authenticate(handlers.CreateComment))
	mux.HandleFunc("/api/comments/list", handlers.ListComments)
	mux.HandleFunc("/api/comments/delete", middleware.Authenticate(handlers.DeleteComment))

	// Authentication routes
	mux.HandleFunc("/api/auth/signin", handlers.SignIn)
	mux.HandleFunc("/api/auth/signout", handlers.SignOut)

	// User routes
	mux.HandleFunc("/api/users/register", handlers.RegisterUser)
	mux.HandleFunc("/api/users/get", handlers.GetUser)
	mux.HandleFunc("/api/users/list", handlers.ListUsers)
	mux.HandleFunc("/api/users/update", middleware.Authenticate(handlers.UpdateUser))
	mux.HandleFunc("/api/users/delete", middleware.Authenticate(handlers.DeleteUser))

	// Post routes
	mux.HandleFunc("/api/posts/create", middleware.Authenticate(handlers.CreatePost))
	mux.HandleFunc("/api/posts/get", handlers.GetPost)
	mux.HandleFunc("/api/posts/list", handlers.ListPosts)
	mux.HandleFunc("/api/posts/user", handlers.ListUserPosts)
	mux.HandleFunc("/api/posts/delete", middleware.Authenticate(handlers.DeletePost))

	// Group routes
	mux.HandleFunc("/api/groups/create", middleware.Authenticate(handlers.CreateGroup))
	mux.HandleFunc("/api/groups/invite", middleware.Authenticate(handlers.AddGroupMember))

	// Notification routes
	mux.HandleFunc("/api/notifications/list", middleware.Authenticate(handlers.ListNotifications))

	// Chat routes
	mux.HandleFunc("/api/chat", middleware.Authenticate(handlers.Chat))

	// pkg/routes/routes.go
	mux.HandleFunc("/api/posts/update", middleware.Authenticate(handlers.UpdatePost))

	return mux
}
