package handlers

import (
	"errors" // Import errors
	"fmt"    // Import fmt
	"net/http"

	"github.com/HASANALI117/social-network/pkg/helpers"
	"github.com/HASANALI117/social-network/pkg/repositories" // Import repositories
	"github.com/HASANALI117/social-network/pkg/services"     // Import services
	ws "github.com/HASANALI117/social-network/pkg/websocket"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var WebSocketHub *ws.Hub

// InitWebsocket initializes the WebSocket Hub with necessary repository and service.
func InitWebsocket(chatMessageRepo repositories.ChatMessageRepository, groupRepo repositories.GroupRepository) { // Changed groupService to groupRepo
	WebSocketHub = ws.NewHub(chatMessageRepo, groupRepo) // Pass groupRepo to NewHub
	go WebSocketHub.Run()
}

// HandleWebSocket now accepts AuthService
func HandleWebSocket(authService services.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// userID := r.URL.Query().Get("id")

		// Pass authService to GetUserFromSession
		userResponse, err := helpers.GetUserFromSession(r, authService)
		if err != nil {
			// Check for specific session error
			if errors.Is(err, helpers.ErrInvalidSession) {
				http.Error(w, "Unauthorized: Invalid session", http.StatusUnauthorized)
			} else {
				// Log other errors for debugging
				fmt.Printf("WebSocket Auth Error: %v\n", err)
				http.Error(w, "Unauthorized", http.StatusUnauthorized) // Generic error to client
			}
			return
		}
		// Use userResponse which is *services.UserResponse

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
			return
		}

		// Use fields from userResponse
		client := ws.NewClient(WebSocketHub, conn, userResponse.ID, userResponse.Username, userResponse.AvatarURL)

		WebSocketHub.Register <- client

		go client.WritePump()
		go client.ReadPump()
	}
}
