package models

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Message represents a chat message
type Message struct {
	ID        string    `json:"id"`
	Sender    string    `json:"sender"`
	Recipient string    `json:"recipient"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	RoomID    string    `json:"room_id,omitempty"`
}

// User represents a chat user
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Don't include in JSON responses
	IsOnline     bool      `json:"is_online"`
	CreatedAt    time.Time `json:"created_at"`
}

// ChatRoom represents a chat room
type ChatRoom struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Members     []string  `json:"members"`
	CreatedAt   time.Time `json:"created_at"`
}

// MessageRequest represents the request payload for sending a message
type MessageRequest struct {
	Sender    string `json:"sender" validate:"required"`
	Recipient string `json:"recipient"`
	Content   string `json:"content" validate:"required"`
	RoomID    string `json:"room_id,omitempty"`
}

// CreateRoomRequest represents the request payload for creating a room
type CreateRoomRequest struct {
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description"`
	Members     []string `json:"members"`
}

// AuthRequest represents authentication request
type AuthRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// RegisterRequest represents registration request
type RegisterRequest struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Token     string `json:"token"`
	User      User   `json:"user"`
	ExpiresAt int64  `json:"expires_at"`
}

// Claims represents JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}
