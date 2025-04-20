package main

import (
"log"
"net/http"
"os"
"os/signal"
"syscall"

"github.com/HASANALI117/social-network/pkg/routes"
"github.com/HASANALI117/social-network/pkg/services"
)

func main() {
// Initialize all services
if err := services.InitServices(); err != nil {
log.Fatalf("Failed to initialize services: %v", err)
}
defer func() {
if err := services.CleanupServices(); err != nil {
log.Printf("Error during service cleanup: %v", err)
}
}()

// Setup routes
handler := routes.Setup()

// Create server
server := &http.Server{
Addr:    ":8080", // TODO: Make configurable
Handler: handler,
}

// Channel to listen for errors coming from the listener.
serverErrors := make(chan error, 1)

// Start the server in a goroutine
go func() {
log.Printf("Server listening on %s", server.Addr)
serverErrors <- server.ListenAndServe()
}()

// Channel to listen for an interrupt or terminate signal from the OS.
shutdown := make(chan os.Signal, 1)
signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

// Blocking main and waiting for shutdown.
select {
case err := <-serverErrors:
log.Fatalf("Error starting server: %v", err)

case <-shutdown:
log.Println("Starting shutdown...")
// TODO: Implement graceful shutdown
// - Close websocket connections
// - Wait for ongoing requests to complete
// - etc.
}
}
