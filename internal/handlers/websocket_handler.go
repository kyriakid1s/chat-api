package handlers

import (
	"encoding/json"
	"go-chat-api/internal/services"
	"go-chat-api/internal/websocket"
	"net/http"
)

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	hub         *websocket.Hub
	chatService *services.ChatService
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(hub *websocket.Hub, chatService *services.ChatService) *WebSocketHandler {
	return &WebSocketHandler{
		hub:         hub,
		chatService: chatService,
	}
}

// HandleWebSocket handles WebSocket connection requests
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get user info from context (set by auth middleware)
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	username, ok := r.Context().Value("username").(string)
	if !ok {
		http.Error(w, "Username not found", http.StatusUnauthorized)
		return
	}

	// Upgrade the HTTP connection to WebSocket
	websocket.ServeWS(h.hub, h.chatService, w, r, userID, username)
}

// GetConnectedUsers returns currently connected users
func (h *WebSocketHandler) GetConnectedUsers(w http.ResponseWriter, r *http.Request) {
	users := h.hub.GetConnectedUsers()

	response := map[string]interface{}{
		"connected_users": users,
		"count":           len(users),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
