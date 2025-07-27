package websocket

import (
	"encoding/json"
	"go-chat-api/internal/models"
	"log"
	"sync"
)

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from the clients
	broadcast chan []byte

	// Register requests from the clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// User ID to client mapping for direct messaging
	userClients map[string]*Client

	// Username to client mapping for direct messaging
	usernameClients map[string]*Client

	// Mutex for thread-safe access to userClients and usernameClients
	mutex sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:         make(map[*Client]bool),
		broadcast:       make(chan []byte),
		register:        make(chan *Client),
		unregister:      make(chan *Client),
		userClients:     make(map[string]*Client),
		usernameClients: make(map[string]*Client),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.mutex.Lock()
			h.userClients[client.UserID] = client
			h.usernameClients[client.Username] = client
			h.mutex.Unlock()

			log.Printf("WebSocket client connected: user %s (%s)", client.Username, client.UserID)

			// Send connection confirmation
			response := map[string]interface{}{
				"type":     "connection",
				"status":   "connected",
				"user_id":  client.Username,
				"username": client.Username,
			}
			if data, err := json.Marshal(response); err == nil {
				select {
				case client.send <- data:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				h.mutex.Lock()
				delete(h.userClients, client.UserID)
				delete(h.usernameClients, client.Username)
				h.mutex.Unlock()
				close(client.send)

				log.Printf("WebSocket client disconnected: user %s (%s)", client.Username, client.UserID)
			}

		case message := <-h.broadcast:
			// Broadcast message to all connected clients
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
					h.mutex.Lock()
					delete(h.userClients, client.UserID)
					h.mutex.Unlock()
				}
			}
		}
	}
}

// BroadcastMessage broadcasts a message to all connected clients
func (h *Hub) BroadcastMessage(message *models.Message) {
	data, err := json.Marshal(map[string]interface{}{
		"type":    "message",
		"message": message,
	})
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	select {
	case h.broadcast <- data:
	default:
		log.Println("Broadcast channel is full, dropping message")
	}
}

// SendToUser sends a message to a specific user by UserID
func (h *Hub) SendToUser(userID string, message *models.Message) bool {
	h.mutex.RLock()
	client, exists := h.userClients[userID]
	h.mutex.RUnlock()

	if !exists {
		return false
	}

	data, err := json.Marshal(map[string]interface{}{
		"type":    "direct_message",
		"message": message,
	})
	if err != nil {
		log.Printf("Error marshaling direct message: %v", err)
		return false
	}

	select {
	case client.send <- data:
		return true
	default:
		// Client's send channel is full, remove the client
		h.unregister <- client
		return false
	}
}

// SendToUsername sends a message to a specific user by username
func (h *Hub) SendToUsername(username string, message *models.Message) bool {
	h.mutex.RLock()
	client, exists := h.usernameClients[username]
	h.mutex.RUnlock()

	if !exists {
		return false
	}

	data, err := json.Marshal(map[string]interface{}{
		"type":    "direct_message",
		"message": message,
	})
	if err != nil {
		log.Printf("Error marshaling direct message: %v", err)
		return false
	}

	select {
	case client.send <- data:
		return true
	default:
		// Client's send channel is full, remove the client
		h.unregister <- client
		return false
	}
}

// SendToRoom sends a message to all users in a specific room
func (h *Hub) SendToRoom(roomID string, message *models.Message) {
	// For now, we'll broadcast to all clients
	// In a more advanced implementation, you'd track which users are in which rooms
	h.BroadcastMessage(message)
}

// GetConnectedUsers returns a list of currently connected usernames
func (h *Hub) GetConnectedUsers() []string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	users := make([]string, 0, len(h.usernameClients))
	for username := range h.usernameClients {
		users = append(users, username)
	}
	return users
}

// IsUserOnline checks if a user is currently connected by username
func (h *Hub) IsUserOnline(username string) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	_, exists := h.usernameClients[username]
	return exists
}
