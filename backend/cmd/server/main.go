package main

import (
	"log"
	"net/http"

	"github.com/HASANALI117/social-network/pkg/db"
	"github.com/HASANALI117/social-network/pkg/helpers"
	"github.com/HASANALI117/social-network/pkg/routes"
)

func main() {
	// Initialize database
	database, err := db.NewDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Initialize user db
	userDB := helpers.NewUserDB(database)

	// Setup HTTP routes
	handler := routes.Setup(userDB)

	// Start HTTP server
	addr := ":8080"
	log.Printf("Server running on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
