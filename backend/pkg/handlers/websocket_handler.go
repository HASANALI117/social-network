package handlers

import (
	"net/http"

	"github.com/HASANALI117/social-network/pkg/helpers"
	ws "github.com/HASANALI117/social-network/pkg/websocket"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var WebSocketHub *ws.Hub

func InitWebsocket() {
	WebSocketHub = ws.NewHub()
	go WebSocketHub.Run()
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// userID := r.URL.Query().Get("id")

	user, err := helpers.GetUserFromSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		return
	}

	client := ws.NewClient(WebSocketHub, conn, user.ID)

	WebSocketHub.Register <- client

	go client.WritePump()
	go client.ReadPump()
}
