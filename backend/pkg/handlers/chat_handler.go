package handlers

import (
	"log"
	"net/http"

	"social-network/pkg/websockets"

	"social-network/pkg/helpers"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var hub = websockets.NewHub()

func init() {
	go hub.Run() // Start the hub in a goroutine
}

func Chat(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID, err := helpers.GetSession(cookie.Value)
	if err != nil {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}
	client := websockets.NewClient(hub, conn, userID)
	hub.Register <- client
	go client.WritePump()
	client.ReadPump() // Runs in the main goroutine for this connection
}
