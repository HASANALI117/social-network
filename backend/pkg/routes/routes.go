package routes

import (
	"net/http"

	"github.com/HASANALI117/social-network/pkg/handlers"
	"github.com/HASANALI117/social-network/pkg/helpers"
)

// Setup sets up all API routes
func Setup(userDB *helpers.UserDB) http.Handler {
	mux := http.NewServeMux()

	// Create handlers
	userHandler := handlers.NewUserHandler(userDB)
	authHandler := handlers.NewAuthHandler(userDB)

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
