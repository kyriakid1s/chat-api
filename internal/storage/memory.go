package storage

import (
	"errors"
	"go-chat-api/internal/models"
	"sort"
	"sync"
	"time"
)

// InMemoryStorage implements all storage interfaces using in-memory data structures
type InMemoryStorage struct {
	mu       sync.RWMutex
	messages []models.Message
	users    map[string]models.User
	rooms    map[string]models.ChatRoom
}

// NewInMemoryStorage creates a new in-memory storage instance
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		messages: make([]models.Message, 0),
		users:    make(map[string]models.User),
		rooms:    make(map[string]models.ChatRoom),
	}
}

// Message Store Implementation
func (s *InMemoryStorage) AddMessage(message models.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messages = append(s.messages, message)
	return nil
}

func (s *InMemoryStorage) GetMessages() ([]models.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy of messages sorted by timestamp
	messages := make([]models.Message, len(s.messages))
	copy(messages, s.messages)

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Timestamp.Before(messages[j].Timestamp)
	})

	return messages, nil
}

func (s *InMemoryStorage) GetMessagesByRoom(roomID string) ([]models.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var roomMessages []models.Message
	for _, msg := range s.messages {
		if msg.RoomID == roomID {
			roomMessages = append(roomMessages, msg)
		}
	}

	sort.Slice(roomMessages, func(i, j int) bool {
		return roomMessages[i].Timestamp.Before(roomMessages[j].Timestamp)
	})

	return roomMessages, nil
}

func (s *InMemoryStorage) GetMessagesBetweenUsers(user1, user2 string) ([]models.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var userMessages []models.Message
	for _, msg := range s.messages {
		if (msg.Sender == user1 && msg.Recipient == user2) ||
			(msg.Sender == user2 && msg.Recipient == user1) {
			userMessages = append(userMessages, msg)
		}
	}

	sort.Slice(userMessages, func(i, j int) bool {
		return userMessages[i].Timestamp.Before(userMessages[j].Timestamp)
	})

	return userMessages, nil
}

// User Store Implementation
func (s *InMemoryStorage) AddUser(user models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[user.ID]; exists {
		return errors.New("user already exists")
	}

	s.users[user.ID] = user
	return nil
}

func (s *InMemoryStorage) GetUser(userID string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[userID]
	if !exists {
		return nil, errors.New("user not found")
	}

	return &user, nil
}

func (s *InMemoryStorage) GetUserByUsername(username string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.Username == username {
			return &user, nil
		}
	}

	return nil, errors.New("user not found")
}

func (s *InMemoryStorage) GetUserByEmail(email string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.Email == email {
			return &user, nil
		}
	}

	return nil, errors.New("user not found")
}

func (s *InMemoryStorage) UpdateUserStatus(userID string, isOnline bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[userID]
	if !exists {
		return errors.New("user not found")
	}

	user.IsOnline = isOnline
	s.users[userID] = user
	return nil
}

func (s *InMemoryStorage) GetAllUsers() ([]models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]models.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}

	return users, nil
}

// Room Store Implementation
func (s *InMemoryStorage) CreateRoom(room models.ChatRoom) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.rooms[room.ID]; exists {
		return errors.New("room already exists")
	}

	room.CreatedAt = time.Now()
	s.rooms[room.ID] = room
	return nil
}

func (s *InMemoryStorage) GetRoom(roomID string) (*models.ChatRoom, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	room, exists := s.rooms[roomID]
	if !exists {
		return nil, errors.New("room not found")
	}

	return &room, nil
}

func (s *InMemoryStorage) GetRoomsByUser(userID string) ([]models.ChatRoom, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var userRooms []models.ChatRoom
	for _, room := range s.rooms {
		for _, member := range room.Members {
			if member == userID {
				userRooms = append(userRooms, room)
				break
			}
		}
	}

	return userRooms, nil
}

func (s *InMemoryStorage) AddUserToRoom(roomID, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, exists := s.rooms[roomID]
	if !exists {
		return errors.New("room not found")
	}

	// Check if user is already a member
	for _, member := range room.Members {
		if member == userID {
			return nil // User already in room
		}
	}

	room.Members = append(room.Members, userID)
	s.rooms[roomID] = room
	return nil
}

func (s *InMemoryStorage) RemoveUserFromRoom(roomID, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, exists := s.rooms[roomID]
	if !exists {
		return errors.New("room not found")
	}

	for i, member := range room.Members {
		if member == userID {
			room.Members = append(room.Members[:i], room.Members[i+1:]...)
			s.rooms[roomID] = room
			return nil
		}
	}

	return errors.New("user not found in room")
}
