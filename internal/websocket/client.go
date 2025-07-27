package websocket

import (
	"bytes"
	"encoding/json"
	"go-chat-api/internal/models"
	"go-chat-api/internal/services"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin
		// In production, you should be more restrictive
		return true
	},
}

// Client is a middleman between the websocket connection and the hub
type Client struct {
	// The websocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan []byte

	// The hub
	hub *Hub

	// User ID of the connected user
	UserID string

	// Username of the connected user
	Username string

	// Chat service for handling messages
	chatService *services.ChatService
}

// IncomingMessage represents a message received from the client
type IncomingMessage struct {
	Type      string `json:"type"`
	Content   string `json:"content"`
	Recipient string `json:"recipient,omitempty"`
	RoomID    string `json:"room_id,omitempty"`
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		messageBytes = bytes.TrimSpace(bytes.Replace(messageBytes, newline, space, -1))

		// Parse the incoming message
		var incomingMsg IncomingMessage
		if err := json.Unmarshal(messageBytes, &incomingMsg); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		// Handle different message types
		switch incomingMsg.Type {
		case "message":
			c.handleMessage(incomingMsg)
		case "ping":
			c.handlePing()
		default:
			log.Printf("Unknown message type: %s", incomingMsg.Type)
		}
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming chat messages
func (c *Client) handleMessage(msg IncomingMessage) {
	// Create message request
	messageReq := models.MessageRequest{
		Sender:    c.Username,
		Content:   msg.Content,
		Recipient: msg.Recipient,
		RoomID:    msg.RoomID,
	}

	// Save message using chat service
	savedMessage, err := c.chatService.SendMessage(messageReq)
	if err != nil {
		log.Printf("Error saving message: %v", err)
		// Send error response to client
		errorResponse := map[string]interface{}{
			"type":  "error",
			"error": "Failed to save message",
		}
		if data, err := json.Marshal(errorResponse); err == nil {
			select {
			case c.send <- data:
			default:
				close(c.send)
			}
		}
		return
	}

	// Broadcast the message based on type
	if msg.RoomID != "" {
		// Room message
		c.hub.SendToRoom(msg.RoomID, savedMessage)
	} else if msg.Recipient != "" {
		// Direct message - send to recipient by username
		c.hub.SendToUsername(msg.Recipient, savedMessage)
		// Also send to sender for confirmation (by username)
		c.hub.SendToUsername(c.Username, savedMessage)
	} else {
		// Global message
		c.hub.BroadcastMessage(savedMessage)
	}
}

// handlePing responds to ping messages
func (c *Client) handlePing() {
	response := map[string]interface{}{
		"type":   "pong",
		"status": "ok",
	}
	if data, err := json.Marshal(response); err == nil {
		select {
		case c.send <- data:
		default:
			close(c.send)
		}
	}
}

// ServeWS handles websocket requests from the peer
func ServeWS(hub *Hub, chatService *services.ChatService, w http.ResponseWriter, r *http.Request, userID, username string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		conn:        conn,
		send:        make(chan []byte, 256),
		hub:         hub,
		UserID:      userID,
		Username:    username,
		chatService: chatService,
	}

	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in new goroutines
	go client.writePump()
	go client.readPump()
}
