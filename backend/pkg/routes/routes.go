package routes

import (
	"net/http"

	"github.com/HASANALI117/social-network/docs"
	"github.com/HASANALI117/social-network/pkg/handlers"
	"github.com/HASANALI117/social-network/pkg/helpers"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Setup sets up all API routes
func Setup(userDB *helpers.UserDB) http.Handler {
	// This will ensure the swagger docs are registered
	docs.SwaggerInfo.BasePath = "/api"

	mux := http.NewServeMux()

	// Create handlers
	userHandler := handlers.NewUserHandler(userDB)
	authHandler := handlers.NewAuthHandler(userDB)

	// Swagger Documentation
	mux.HandleFunc("/swagger/", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	// Authentication routes
	mux.HandleFunc("/api/auth/signin", authHandler.SignIn)
	mux.HandleFunc("/api/auth/signout", authHandler.SignOut)

	// User routes
	mux.HandleFunc("/api/users/register", userHandler.Register)
	mux.HandleFunc("/api/users/get", userHandler.GetUser)
	mux.HandleFunc("/api/users/list", userHandler.ListUsers)
	mux.HandleFunc("/api/users/update", userHandler.UpdateUser)
	mux.HandleFunc("/api/users/delete", userHandler.DeleteUser)

	return mux
}
