package websockets

import (
	"fmt"

	"social-network/pkg/helpers"
	"social-network/pkg/models"
)

type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan *Message
	Register   chan *Client
	Unregister chan *Client
}

type Message struct {
	Type       string `json:"type"`
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"receiver_id"`
	Content    string `json:"content"`
	CreatedAt  string `json:"created_at"`
}

func NewHub() *Hub {
	return &Hub{
		Clients:    map[*Client]bool{},
		Broadcast:  make(chan *Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
			fmt.Printf("➡️  User %s connected to chat\n", client.UserID)
			fmt.Printf("📊 Active users: %d\n", len(h.Clients))

			// Log all connected users
			fmt.Println("🟢 Connected users:")
			for client := range h.Clients {
				fmt.Printf("   - User ID: %s\n", client.UserID)
			}

		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)

				fmt.Printf("❌ User %s disconnected from chat\n", client.UserID)
				fmt.Printf("📊 Remaining active users: %d\n", len(h.Clients))
			}

		case message := <-h.Broadcast:
			switch message.Type {
			case "direct":
				fmt.Printf("\n📨 New direct message received:\n")
				fmt.Printf("   From: User %s\n", message.SenderID)
				fmt.Printf("   To: User %s\n", message.ReceiverID)
				fmt.Printf("   Content: %s\n", message.Content)
				fmt.Printf("   Time: %s\n", message.CreatedAt)

				msg := &models.Message{
					SenderID:   message.SenderID,
					ReceiverID: message.ReceiverID,
					Content:    message.Content,
					CreatedAt:  message.CreatedAt,
				}

				err := helpers.SaveMessage(msg)
				if err != nil {
					fmt.Printf("❌ Error storing message: %v\n", err)
					continue
				}

				for client := range h.Clients {
					if client.UserID == message.ReceiverID || client.UserID == message.SenderID {
						select {
						case client.Send <- message:
							fmt.Printf("✅ Message delivered to User %s\n", client.UserID)

						default:
							fmt.Printf("⚠️ Failed to deliver message to User %s - connection closed\n", client.UserID)
							close(client.Send)
							delete(h.Clients, client)
						}
					}
				}

			case "group":
				// TODO: Implement group message handling
				fmt.Println("👥 Group messages not implemented yet")
			}
		}
	}
}

// func (h * Hub) sendDirectMessage() {

// }
