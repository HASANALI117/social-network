package websockets

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	Hub    *Hub
	Conn   *websocket.Conn
	Send   chan *Message
	UserID string
}

func NewClient(hub *Hub, conn *websocket.Conn, userID string) *Client {
	return &Client{
		Hub:    hub,
		Conn:   conn,
		Send:   make(chan *Message, 256),
		UserID: userID,
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		// _, msgBytes, err := c.Conn.ReadMessage()
		// if err != nil {
		// 	if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		// 		log.Printf("error: %v", err)
		// 	}
		// 	break
		// }

		var message Message
		// if err := json.Unmarshal(msgBytes, &message); err != nil {
		// 	log.Printf("error decoding message: %v", err)
		// 	continue
		// }

		err := c.Conn.ReadJSON(&message)
		if err != nil {
			log.Println("read error:", err)
			break
		}

		message.SenderID = c.UserID
		if message.CreatedAt == "" {
			message.CreatedAt = time.Now().Format(time.RFC3339)
		}

		switch message.Type {
		case "direct":
			c.Hub.Broadcast <- &message

		case "group":
			c.Hub.Broadcast <- &message

		default:
			log.Printf("unknown message type: %s", message.Type)
		}
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()

	for message := range c.Send {
		err := c.Conn.WriteJSON(&message)
		if err != nil {
			log.Println("write error:", err)
			c.Conn.Close()
			return
		}

		// w, err := c.conn.NextWriter(websocket.TextMessage)
		// if err != nil {
		// 	return
		// }

		// messageBytes, err := json.Marshal(message)
		// if err != nil {
		// 	log.Printf("error encoding message: %v", err)
		// 	return
		// }

		// w.Write(messageBytes)

		// if err := w.Close(); err != nil {
		// 	return
		// }
	}
}
