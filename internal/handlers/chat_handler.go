package handlers

import (
	"encoding/json"
	"go-chat-api/internal/models"
	"go-chat-api/internal/services"
	"go-chat-api/internal/websocket"
	"net/http"

	"github.com/gorilla/mux"
)

// ChatHandler handles HTTP requests for chat operations
type ChatHandler struct {
	chatService *services.ChatService
	hub         *websocket.Hub // WebSocket hub for live messaging
}

// NewChatHandler creates a new chat handler with injected dependencies
func NewChatHandler(chatService *services.ChatService, hub *websocket.Hub) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		hub:         hub,
	}
}

// SendMessage handles POST /api/messages
func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	var req models.MessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	message, err := h.chatService.SendMessage(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Broadcast the message to WebSocket clients
	if h.hub != nil {
		if req.RoomID != "" {
			// Room message
			h.hub.SendToRoom(req.RoomID, message)
		} else if req.Recipient != "" {
			// Direct message - send to recipient and sender
			h.hub.SendToUser(req.Recipient, message)
			h.hub.SendToUser(req.Sender, message)
		} else {
			// Global message
			h.hub.BroadcastMessage(message)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(message)
}

// GetMessages handles GET /api/messages
func (h *ChatHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	messages, err := h.chatService.GetMessages()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// GetMessagesByRoom handles GET /api/rooms/{roomId}/messages
func (h *ChatHandler) GetMessagesByRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]

	messages, err := h.chatService.GetMessagesByRoom(roomID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// GetMessagesBetweenUsers handles GET /api/messages/between/{user1}/{user2}
func (h *ChatHandler) GetMessagesBetweenUsers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	user1 := vars["user1"]
	user2 := vars["user2"]

	messages, err := h.chatService.GetMessagesBetweenUsers(user1, user2)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// CreateUser handles POST /api/users
func (h *ChatHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.chatService.CreateUser(req.Username, req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// GetUser handles GET /api/users/{userId}
func (h *ChatHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	user, err := h.chatService.GetUser(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// GetAllUsers handles GET /api/users
func (h *ChatHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.chatService.GetAllUsers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// CreateRoom handles POST /api/rooms
func (h *ChatHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var req models.CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	room, err := h.chatService.CreateRoom(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(room)
}

// GetRoom handles GET /api/rooms/{roomId}
func (h *ChatHandler) GetRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]

	room, err := h.chatService.GetRoom(roomID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(room)
}

// GetRoomsByUser handles GET /api/users/{userId}/rooms
func (h *ChatHandler) GetRoomsByUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userId"]

	rooms, err := h.chatService.GetRoomsByUser(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rooms)
}

// AddUserToRoom handles POST /api/rooms/{roomId}/members/{userId}
func (h *ChatHandler) AddUserToRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	userID := vars["userId"]

	err := h.chatService.AddUserToRoom(roomID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

// RemoveUserFromRoom handles DELETE /api/rooms/{roomId}/members/{userId}
func (h *ChatHandler) RemoveUserFromRoom(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	roomID := vars["roomId"]
	userID := vars["userId"]

	err := h.chatService.RemoveUserFromRoom(roomID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
