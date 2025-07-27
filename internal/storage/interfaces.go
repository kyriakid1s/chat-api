package storage

import "go-chat-api/internal/models"

// MessageStore defines the interface for message storage operations
type MessageStore interface {
	AddMessage(message models.Message) error
	GetMessages() ([]models.Message, error)
	GetMessagesByRoom(roomID string) ([]models.Message, error)
	GetMessagesBetweenUsers(user1, user2 string) ([]models.Message, error)
}

// UserStore defines the interface for user storage operations
type UserStore interface {
	AddUser(user models.User) error
	GetUser(userID string) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	UpdateUserStatus(userID string, isOnline bool) error
	GetAllUsers() ([]models.User, error)
}

// RoomStore defines the interface for chat room storage operations
type RoomStore interface {
	CreateRoom(room models.ChatRoom) error
	GetRoom(roomID string) (*models.ChatRoom, error)
	GetRoomsByUser(userID string) ([]models.ChatRoom, error)
	AddUserToRoom(roomID, userID string) error
	RemoveUserFromRoom(roomID, userID string) error
}
