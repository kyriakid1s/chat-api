package services

import (
	"go-chat-api/internal/auth"
	"go-chat-api/internal/models"
	"go-chat-api/internal/storage"
	"testing"
	"time"
)

func setupTestChatService() *ChatService {
	// Create in-memory storage
	store := storage.NewInMemoryStorage()

	// Create auth service with test secret
	authService := auth.NewAuthService("test-secret", 24*time.Hour)

	// Create chat service
	return NewChatService(store, store, store, authService)
}

func TestChatService_RegisterUser(t *testing.T) {
	service := setupTestChatService()

	tests := []struct {
		name    string
		req     models.RegisterRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid registration",
			req: models.RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "duplicate username",
			req: models.RegisterRequest{
				Username: "testuser", // Same as above
				Email:    "different@example.com",
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "username already exists",
		},
		{
			name: "duplicate email",
			req: models.RegisterRequest{
				Username: "differentuser",
				Email:    "test@example.com", // Same as first test
				Password: "password123",
			},
			wantErr: true,
			errMsg:  "email already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.RegisterUser(tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("RegisterUser() expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("RegisterUser() error = %v, want %v", err.Error(), tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("RegisterUser() unexpected error = %v", err)
				return
			}

			if user == nil {
				t.Error("RegisterUser() returned nil user")
				return
			}

			if user.Username != tt.req.Username {
				t.Errorf("RegisterUser() Username = %v, want %v", user.Username, tt.req.Username)
			}

			if user.Email != tt.req.Email {
				t.Errorf("RegisterUser() Email = %v, want %v", user.Email, tt.req.Email)
			}

			if user.PasswordHash == "" {
				t.Error("RegisterUser() PasswordHash is empty")
			}

			if user.PasswordHash == tt.req.Password {
				t.Error("RegisterUser() PasswordHash should not be plaintext password")
			}

			if user.ID == "" {
				t.Error("RegisterUser() ID is empty")
			}

			if user.IsOnline {
				t.Error("RegisterUser() user should not be online initially")
			}
		})
	}
}

func TestChatService_AuthenticateUser(t *testing.T) {
	service := setupTestChatService()

	// First register a user
	registerReq := models.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	_, err := service.RegisterUser(registerReq)
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	tests := []struct {
		name    string
		req     models.AuthRequest
		wantErr bool
	}{
		{
			name: "valid credentials",
			req: models.AuthRequest{
				Username: "testuser",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "invalid username",
			req: models.AuthRequest{
				Username: "nonexistent",
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "invalid password",
			req: models.AuthRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			wantErr: true,
		},
		{
			name: "empty username",
			req: models.AuthRequest{
				Username: "",
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "empty password",
			req: models.AuthRequest{
				Username: "testuser",
				Password: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authResp, err := service.AuthenticateUser(tt.req)

			if tt.wantErr {
				if err == nil {
					t.Errorf("AuthenticateUser() expected error but got none")
				}
				if err != nil && err.Error() != "invalid credentials" {
					t.Errorf("AuthenticateUser() error = %v, want 'invalid credentials'", err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("AuthenticateUser() unexpected error = %v", err)
				return
			}

			if authResp == nil {
				t.Error("AuthenticateUser() returned nil response")
				return
			}

			if authResp.Token == "" {
				t.Error("AuthenticateUser() Token is empty")
			}

			if authResp.User.Username != tt.req.Username {
				t.Errorf("AuthenticateUser() Username = %v, want %v", authResp.User.Username, tt.req.Username)
			}

			if authResp.ExpiresAt <= time.Now().Unix() {
				t.Error("AuthenticateUser() ExpiresAt is in the past")
			}

			if !authResp.User.IsOnline {
				t.Error("AuthenticateUser() user should be online after authentication")
			}
		})
	}
}

func TestChatService_RefreshToken(t *testing.T) {
	service := setupTestChatService()

	// Register and authenticate a user first
	registerReq := models.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	_, err := service.RegisterUser(registerReq)
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	authReq := models.AuthRequest{
		Username: "testuser",
		Password: "password123",
	}

	authResp, err := service.AuthenticateUser(authReq)
	if err != nil {
		t.Fatalf("Failed to authenticate test user: %v", err)
	}

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "invalid token",
			token:   "invalid.token.format",
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refreshResp, err := service.RefreshToken(tt.token)

			if tt.wantErr {
				if err == nil {
					t.Errorf("RefreshToken() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("RefreshToken() unexpected error = %v", err)
				return
			}

			if refreshResp == nil {
				t.Error("RefreshToken() returned nil response")
				return
			}

			if refreshResp.Token == "" {
				t.Error("RefreshToken() Token is empty")
			}

			if refreshResp.Token == authResp.Token {
				t.Error("RefreshToken() should return a new token")
			}
		})
	}
}

func TestChatService_LogoutUser(t *testing.T) {
	service := setupTestChatService()

	// Register a user first
	registerReq := models.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	user, err := service.RegisterUser(registerReq)
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	// Set user online first
	err = service.UpdateUserStatus(user.ID, true)
	if err != nil {
		t.Fatalf("Failed to set user online: %v", err)
	}

	// Test logout
	err = service.LogoutUser(user.ID)
	if err != nil {
		t.Errorf("LogoutUser() unexpected error = %v", err)
	}

	// Verify user is offline
	updatedUser, err := service.GetUser(user.ID)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}

	if updatedUser.IsOnline {
		t.Error("LogoutUser() user should be offline after logout")
	}

	// Test logout with invalid user ID
	err = service.LogoutUser("nonexistent-id")
	if err == nil {
		t.Error("LogoutUser() should return error for nonexistent user")
	}
}

func TestChatService_SendMessage_WithAuth(t *testing.T) {
	service := setupTestChatService()

	// Register a user first
	registerReq := models.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	user, err := service.RegisterUser(registerReq)
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	// Test sending a message
	messageReq := models.MessageRequest{
		Sender:    user.Username,
		Recipient: "recipient",
		Content:   "Hello, World!",
	}

	message, err := service.SendMessage(messageReq)
	if err != nil {
		t.Errorf("SendMessage() unexpected error = %v", err)
		return
	}

	if message == nil {
		t.Error("SendMessage() returned nil message")
		return
	}

	if message.Sender != messageReq.Sender {
		t.Errorf("SendMessage() Sender = %v, want %v", message.Sender, messageReq.Sender)
	}

	if message.Content != messageReq.Content {
		t.Errorf("SendMessage() Content = %v, want %v", message.Content, messageReq.Content)
	}

	if message.ID == "" {
		t.Error("SendMessage() ID is empty")
	}
}

func TestChatService_Integration_FullAuthFlow(t *testing.T) {
	service := setupTestChatService()

	// 1. Register user
	registerReq := models.RegisterRequest{
		Username: "integrationuser",
		Email:    "integration@example.com",
		Password: "password123",
	}

	user, err := service.RegisterUser(registerReq)
	if err != nil {
		t.Fatalf("Integration test failed at registration: %v", err)
	}

	if user.IsOnline {
		t.Error("User should not be online after registration")
	}

	// 2. Authenticate user
	authReq := models.AuthRequest{
		Username: "integrationuser",
		Password: "password123",
	}

	authResp, err := service.AuthenticateUser(authReq)
	if err != nil {
		t.Fatalf("Integration test failed at authentication: %v", err)
	}

	if !authResp.User.IsOnline {
		t.Error("User should be online after authentication")
	}

	// 3. Send a message
	messageReq := models.MessageRequest{
		Sender:    user.Username,
		Recipient: "someone",
		Content:   "Integration test message",
	}

	message, err := service.SendMessage(messageReq)
	if err != nil {
		t.Fatalf("Integration test failed at sending message: %v", err)
	}

	// 4. Get messages
	messages, err := service.GetMessages()
	if err != nil {
		t.Fatalf("Integration test failed at getting messages: %v", err)
	}

	found := false
	for _, msg := range messages {
		if msg.ID == message.ID {
			found = true
			break
		}
	}

	if !found {
		t.Error("Integration test: sent message not found in messages list")
	}

	// 5. Logout user
	err = service.LogoutUser(user.ID)
	if err != nil {
		t.Fatalf("Integration test failed at logout: %v", err)
	}

	// 6. Verify user is offline
	updatedUser, err := service.GetUser(user.ID)
	if err != nil {
		t.Fatalf("Integration test failed at getting updated user: %v", err)
	}

	if updatedUser.IsOnline {
		t.Error("User should be offline after logout")
	}
}
