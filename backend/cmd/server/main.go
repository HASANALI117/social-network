// cmd/server/cmain.go

// @title Social Network API
// @version 1.0
// @description API server for Social Network application
// @host localhost:8080
// @BasePath /api

package main

import (
	"log"
	"net/http"

	"github.com/HASANALI117/social-network/pkg/db"
	"github.com/HASANALI117/social-network/pkg/routes"
)

func main() {
	// Initialize database
	database, err := db.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Setup HTTP routes
	handler := routes.Setup(database)

	// Start HTTP server
	addr := ":8080"
	log.Printf("Server running on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
