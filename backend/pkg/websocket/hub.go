package ws

import (
	"fmt"
	"time"

	"github.com/HASANALI117/social-network/pkg/helpers"
	"github.com/HASANALI117/social-network/pkg/models"
)

type Hub struct {
	Clients    map[string]*Client
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
		Clients:    make(map[string]*Client),
		Broadcast:  make(chan *Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client.UserID] = client
			fmt.Printf("➡️  User %s connected to chat\n", client.UserID)
			fmt.Printf("📊 Active users: %d\n", len(h.Clients))

			// Log all connected users
			fmt.Println("🟢 Connected users:")
			for userID := range h.Clients {
				fmt.Printf("   - User ID: %s\n", userID)
			}

			go h.broadcastUserStatusChange()

		case client := <-h.Unregister:
			if _, ok := h.Clients[client.UserID]; ok {
				delete(h.Clients, client.UserID)
				close(client.Send)

				fmt.Printf("❌ User %s disconnected from chat\n", client.UserID)
				fmt.Printf("📊 Remaining active users: %d\n", len(h.Clients))

				go h.broadcastUserStatusChange()
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

				if err := helpers.SaveMessage(msg); err != nil {
					fmt.Printf("❌ Error storing message: %v\n", err)
				}

				h.deliverMessage(message.SenderID, message)
				h.deliverMessage(message.ReceiverID, message)

			case "group":
				fmt.Printf("\n👥 New group message received:\n")
				fmt.Printf("   From: User %s\n", message.SenderID)
				fmt.Printf("   To Group: %s\n", message.ReceiverID)
				fmt.Printf("   Content: %s\n", message.Content)
				fmt.Printf("   Time: %s\n", message.CreatedAt)

				// Save the message to the database
				groupMsg := &models.GroupMessage{
					GroupID:   message.ReceiverID, // GroupID is in the ReceiverID field
					SenderID:  message.SenderID,
					Content:   message.Content,
					CreatedAt: message.CreatedAt,
				}

				if err := helpers.SaveGroupMessage(groupMsg); err != nil {
					fmt.Printf("❌ Error storing group message: %v\n", err)
				}

				// Check if sender is a member of the group
				isMember, err := helpers.IsGroupMember(message.ReceiverID, message.SenderID)
				if err != nil || !isMember {
					fmt.Printf("❌ User %s is not a member of group %s\n", message.SenderID, message.ReceiverID)
					continue
				}

				// Get all group members
				members, err := helpers.ListGroupMembers(message.ReceiverID)
				if err != nil {
					fmt.Printf("❌ Error getting group members: %v\n", err)
					continue
				}

				// Deliver the message to all online group members
				for _, member := range members {
					h.deliverMessage(member.ID, message)
				}
			}
		}
	}
}

func (h *Hub) deliverMessage(userID string, message *Message) {
	client, ok := h.Clients[userID]

	if ok {
		select {
		case client.Send <- message:
			fmt.Printf("✅ Message delivered to User %s\n", userID)

		default:
			fmt.Printf("⚠️ Failed to deliver message to User %s - connection closed\n", userID)
			delete(h.Clients, userID)
			close(client.Send)
		}
	}
}

func (h *Hub) broadcastUserStatusChange() {
	time.Sleep(500 * time.Millisecond)

	message := &Message{
		Type: "online_users",
	}

	for _, client := range h.Clients {
		h.deliverMessage(client.UserID, message)
	}
}
