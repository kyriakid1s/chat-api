package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"go-chat-api/internal/auth"
	"go-chat-api/internal/models"
	"go-chat-api/internal/storage"
	"time"
)

// ChatService handles business logic for chat operations
type ChatService struct {
	messageStore storage.MessageStore
	userStore    storage.UserStore
	roomStore    storage.RoomStore
	authService  *auth.AuthService
}

// NewChatService creates a new chat service with injected dependencies
func NewChatService(messageStore storage.MessageStore, userStore storage.UserStore, roomStore storage.RoomStore, authService *auth.AuthService) *ChatService {
	return &ChatService{
		messageStore: messageStore,
		userStore:    userStore,
		roomStore:    roomStore,
		authService:  authService,
	}
}

// SendMessage handles sending a message
func (s *ChatService) SendMessage(req models.MessageRequest) (*models.Message, error) {
	// Generate unique ID for the message
	id, err := generateID()
	if err != nil {
		return nil, err
	}

	// Convert empty recipient to empty string for JSON, but handle NULL in database
	recipient := req.Recipient
	if recipient == "" {
		recipient = "" // Keep empty for JSON response, but database layer will handle NULL
	}

	message := models.Message{
		ID:        id,
		Sender:    req.Sender,
		Recipient: recipient,
		Content:   req.Content,
		RoomID:    req.RoomID,
		Timestamp: time.Now(),
	}

	err = s.messageStore.AddMessage(message)
	if err != nil {
		return nil, err
	}

	return &message, nil
}

// GetMessages retrieves all messages
func (s *ChatService) GetMessages() ([]models.Message, error) {
	return s.messageStore.GetMessages()
}

// GetMessagesByRoom retrieves messages for a specific room
func (s *ChatService) GetMessagesByRoom(roomID string) ([]models.Message, error) {
	return s.messageStore.GetMessagesByRoom(roomID)
}

// GetMessagesBetweenUsers retrieves messages between two users
func (s *ChatService) GetMessagesBetweenUsers(user1, user2 string) ([]models.Message, error) {
	return s.messageStore.GetMessagesBetweenUsers(user1, user2)
}

// CreateUser creates a new user
func (s *ChatService) CreateUser(username, email string) (*models.User, error) {
	id, err := generateID()
	if err != nil {
		return nil, err
	}

	user := models.User{
		ID:        id,
		Username:  username,
		Email:     email,
		IsOnline:  false,
		CreatedAt: time.Now(),
	}

	err = s.userStore.AddUser(user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// RegisterUser creates a new user with authentication
func (s *ChatService) RegisterUser(req models.RegisterRequest) (*models.User, error) {
	// Check if username already exists
	if existingUser, err := s.userStore.GetUserByUsername(req.Username); err != nil {
		return nil, err
	} else if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// Check if email already exists
	if existingUser, err := s.userStore.GetUserByEmail(req.Email); err != nil {
		return nil, err
	} else if existingUser != nil {
		return nil, errors.New("email already exists")
	}

	// Hash password
	hashedPassword, err := s.authService.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Generate user ID
	id, err := generateID()
	if err != nil {
		return nil, err
	}

	user := models.User{
		ID:           id,
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		IsOnline:     false,
		CreatedAt:    time.Now(),
	}

	err = s.userStore.AddUser(user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// AuthenticateUser authenticates a user and returns a token
func (s *ChatService) AuthenticateUser(req models.AuthRequest) (*models.AuthResponse, error) {
	// Find user by username
	user, err := s.userStore.GetUserByUsername(req.Username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	// Verify password
	err = s.authService.VerifyPassword(user.PasswordHash, req.Password)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Generate token
	token, expiresAt, err := s.authService.GenerateToken(*user)
	if err != nil {
		return nil, err
	}

	// Update user online status
	err = s.userStore.UpdateUserStatus(user.ID, true)
	if err != nil {
		return nil, err
	}

	// Get updated user data
	updatedUser, err := s.userStore.GetUser(user.ID)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		Token:     token,
		User:      *updatedUser,
		ExpiresAt: expiresAt,
	}, nil
}

// RefreshToken refreshes a user's authentication token
func (s *ChatService) RefreshToken(tokenString string) (*models.AuthResponse, error) {
	newToken, expiresAt, err := s.authService.RefreshToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Get user info from token
	claims, err := s.authService.ValidateToken(newToken)
	if err != nil {
		return nil, err
	}

	user, err := s.userStore.GetUser(claims.UserID)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		Token:     newToken,
		User:      *user,
		ExpiresAt: expiresAt,
	}, nil
}

// LogoutUser logs out a user by updating their online status
func (s *ChatService) LogoutUser(userID string) error {
	return s.userStore.UpdateUserStatus(userID, false)
}

// GetUser retrieves a user by ID
func (s *ChatService) GetUser(userID string) (*models.User, error) {
	return s.userStore.GetUser(userID)
}

// GetUserByUsername retrieves a user by username
func (s *ChatService) GetUserByUsername(username string) (*models.User, error) {
	return s.userStore.GetUserByUsername(username)
}

// UpdateUserStatus updates a user's online status
func (s *ChatService) UpdateUserStatus(userID string, isOnline bool) error {
	return s.userStore.UpdateUserStatus(userID, isOnline)
}

// GetAllUsers retrieves all users
func (s *ChatService) GetAllUsers() ([]models.User, error) {
	return s.userStore.GetAllUsers()
}

// CreateRoom creates a new chat room
func (s *ChatService) CreateRoom(req models.CreateRoomRequest) (*models.ChatRoom, error) {
	id, err := generateID()
	if err != nil {
		return nil, err
	}

	room := models.ChatRoom{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Members:     req.Members,
		CreatedAt:   time.Now(),
	}

	err = s.roomStore.CreateRoom(room)
	if err != nil {
		return nil, err
	}

	return &room, nil
}

// GetRoom retrieves a room by ID
func (s *ChatService) GetRoom(roomID string) (*models.ChatRoom, error) {
	return s.roomStore.GetRoom(roomID)
}

// GetRoomsByUser retrieves all rooms for a user
func (s *ChatService) GetRoomsByUser(userID string) ([]models.ChatRoom, error) {
	return s.roomStore.GetRoomsByUser(userID)
}

// AddUserToRoom adds a user to a room
func (s *ChatService) AddUserToRoom(roomID, userID string) error {
	return s.roomStore.AddUserToRoom(roomID, userID)
}

// RemoveUserFromRoom removes a user from a room
func (s *ChatService) RemoveUserFromRoom(roomID, userID string) error {
	return s.roomStore.RemoveUserFromRoom(roomID, userID)
}

// generateID generates a random hex ID
func generateID() (string, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
