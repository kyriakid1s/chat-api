package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"go-chat-api/internal/auth"
	"go-chat-api/internal/models"
	"go-chat-api/internal/services"
	"go-chat-api/internal/storage"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func setupTestAuthHandler() (*AuthHandler, *services.ChatService) {
	// Create in-memory storage
	store := storage.NewInMemoryStorage()

	// Create auth service with test secret
	authService := auth.NewAuthService("test-secret", 24*time.Hour)

	// Create chat service
	chatService := services.NewChatService(store, store, store, authService)

	// Create auth handler
	authHandler := NewAuthHandler(chatService)

	return authHandler, chatService
}

func TestAuthHandler_Register(t *testing.T) {
	handler, _ := setupTestAuthHandler()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectUser     bool
	}{
		{
			name: "valid registration",
			requestBody: models.RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusCreated,
			expectUser:     true,
		},
		{
			name: "duplicate username",
			requestBody: models.RegisterRequest{
				Username: "testuser", // Same as above
				Email:    "different@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusConflict,
			expectUser:     false,
		},
		{
			name: "short password",
			requestBody: models.RegisterRequest{
				Username: "newuser",
				Email:    "new@example.com",
				Password: "123", // Too short
			},
			expectedStatus: http.StatusBadRequest,
			expectUser:     false,
		},
		{
			name: "missing username",
			requestBody: models.RegisterRequest{
				Email:    "missing@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectUser:     false,
		},
		{
			name: "missing email",
			requestBody: models.RegisterRequest{
				Username: "missingemail",
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectUser:     false,
		},
		{
			name: "missing password",
			requestBody: models.RegisterRequest{
				Username: "missingpass",
				Email:    "missingpass@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectUser:     false,
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectUser:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			if str, ok := tt.requestBody.(string); ok {
				body.WriteString(str)
			} else {
				json.NewEncoder(&body).Encode(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/auth/register", &body)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.Register(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Register() status = %v, want %v", rr.Code, tt.expectedStatus)
			}

			if tt.expectUser {
				var user models.User
				err := json.NewDecoder(rr.Body).Decode(&user)
				if err != nil {
					t.Errorf("Register() failed to decode response: %v", err)
					return
				}

				if user.ID == "" {
					t.Error("Register() returned user with empty ID")
				}

				if user.Username == "" {
					t.Error("Register() returned user with empty username")
				}

				if user.PasswordHash != "" {
					t.Error("Register() returned user with password hash (should be excluded)")
				}
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	handler, chatService := setupTestAuthHandler()

	// Register a user first
	registerReq := models.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	_, err := chatService.RegisterUser(registerReq)
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectToken    bool
	}{
		{
			name: "valid login",
			requestBody: models.AuthRequest{
				Username: "testuser",
				Password: "password123",
			},
			expectedStatus: http.StatusOK,
			expectToken:    true,
		},
		{
			name: "invalid username",
			requestBody: models.AuthRequest{
				Username: "nonexistent",
				Password: "password123",
			},
			expectedStatus: http.StatusUnauthorized,
			expectToken:    false,
		},
		{
			name: "invalid password",
			requestBody: models.AuthRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			expectToken:    false,
		},
		{
			name: "missing username",
			requestBody: models.AuthRequest{
				Password: "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectToken:    false,
		},
		{
			name: "missing password",
			requestBody: models.AuthRequest{
				Username: "testuser",
			},
			expectedStatus: http.StatusBadRequest,
			expectToken:    false,
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectToken:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body bytes.Buffer
			if str, ok := tt.requestBody.(string); ok {
				body.WriteString(str)
			} else {
				json.NewEncoder(&body).Encode(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/auth/login", &body)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.Login(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Login() status = %v, want %v", rr.Code, tt.expectedStatus)
			}

			if tt.expectToken {
				var authResp models.AuthResponse
				err := json.NewDecoder(rr.Body).Decode(&authResp)
				if err != nil {
					t.Errorf("Login() failed to decode response: %v", err)
					return
				}

				if authResp.Token == "" {
					t.Error("Login() returned empty token")
				}

				if authResp.User.Username != "testuser" {
					t.Errorf("Login() Username = %v, want testuser", authResp.User.Username)
				}

				if authResp.ExpiresAt <= time.Now().Unix() {
					t.Error("Login() ExpiresAt is in the past")
				}
			}
		})
	}
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	handler, chatService := setupTestAuthHandler()

	// Register and authenticate a user first
	registerReq := models.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	_, err := chatService.RegisterUser(registerReq)
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	authReq := models.AuthRequest{
		Username: "testuser",
		Password: "password123",
	}

	_, err = chatService.AuthenticateUser(authReq)
	if err != nil {
		t.Fatalf("Failed to authenticate test user: %v", err)
	}

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectToken    bool
	}{
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			expectToken:    false,
		},
		{
			name:           "invalid authorization format",
			authHeader:     "InvalidFormat token",
			expectedStatus: http.StatusUnauthorized,
			expectToken:    false,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid.token.format",
			expectedStatus: http.StatusUnauthorized,
			expectToken:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			handler.RefreshToken(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("RefreshToken() status = %v, want %v", rr.Code, tt.expectedStatus)
			}
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	handler, chatService := setupTestAuthHandler()

	// Register a user first
	registerReq := models.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	user, err := chatService.RegisterUser(registerReq)
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
	}{
		{
			name:           "valid logout",
			userID:         user.ID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing user context",
			userID:         "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)

			// Add user context if provided
			if tt.userID != "" {
				ctx := context.WithValue(req.Context(), "userID", tt.userID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			handler.Logout(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Logout() status = %v, want %v", rr.Code, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusOK {
				var response map[string]string
				err := json.NewDecoder(rr.Body).Decode(&response)
				if err != nil {
					t.Errorf("Logout() failed to decode response: %v", err)
					return
				}

				if response["message"] != "Logged out successfully" {
					t.Errorf("Logout() message = %v, want 'Logged out successfully'", response["message"])
				}
			}
		})
	}
}

func TestAuthHandler_GetProfile(t *testing.T) {
	handler, chatService := setupTestAuthHandler()

	// Register a user first
	registerReq := models.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	user, err := chatService.RegisterUser(registerReq)
	if err != nil {
		t.Fatalf("Failed to register test user: %v", err)
	}

	tests := []struct {
		name           string
		userID         string
		expectedStatus int
	}{
		{
			name:           "valid profile request",
			userID:         user.ID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing user context",
			userID:         "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "nonexistent user",
			userID:         "nonexistent-id",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/auth/profile", nil)

			// Add user context if provided
			if tt.userID != "" {
				ctx := context.WithValue(req.Context(), "userID", tt.userID)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			handler.GetProfile(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("GetProfile() status = %v, want %v", rr.Code, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusOK {
				var profile models.User
				err := json.NewDecoder(rr.Body).Decode(&profile)
				if err != nil {
					t.Errorf("GetProfile() failed to decode response: %v", err)
					return
				}

				if profile.ID != user.ID {
					t.Errorf("GetProfile() ID = %v, want %v", profile.ID, user.ID)
				}

				if profile.Username != user.Username {
					t.Errorf("GetProfile() Username = %v, want %v", profile.Username, user.Username)
				}

				if profile.PasswordHash != "" {
					t.Error("GetProfile() returned user with password hash (should be excluded)")
				}
			}
		})
	}
}

func TestAuthHandler_Integration_FullFlow(t *testing.T) {
	handler, _ := setupTestAuthHandler()

	// 1. Register user
	registerBody := models.RegisterRequest{
		Username: "integrationuser",
		Email:    "integration@example.com",
		Password: "password123",
	}

	var body bytes.Buffer
	json.NewEncoder(&body).Encode(registerBody)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", &body)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.Register(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("Integration test failed at registration: status = %v", rr.Code)
	}

	var user models.User
	err := json.NewDecoder(rr.Body).Decode(&user)
	if err != nil {
		t.Fatalf("Integration test failed to decode registration response: %v", err)
	}

	// 2. Login user
	loginBody := models.AuthRequest{
		Username: "integrationuser",
		Password: "password123",
	}

	body.Reset()
	json.NewEncoder(&body).Encode(loginBody)

	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", &body)
	req.Header.Set("Content-Type", "application/json")

	rr = httptest.NewRecorder()
	handler.Login(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Integration test failed at login: status = %v", rr.Code)
	}

	var authResp models.AuthResponse
	err = json.NewDecoder(rr.Body).Decode(&authResp)
	if err != nil {
		t.Fatalf("Integration test failed to decode login response: %v", err)
	}

	// 3. Get profile with user context
	req = httptest.NewRequest(http.MethodGet, "/api/auth/profile", nil)
	ctx := context.WithValue(req.Context(), "userID", user.ID)
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()
	handler.GetProfile(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Integration test failed at get profile: status = %v", rr.Code)
	}

	// 4. Logout user
	req = httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	ctx = context.WithValue(req.Context(), "userID", user.ID)
	req = req.WithContext(ctx)

	rr = httptest.NewRecorder()
	handler.Logout(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Integration test failed at logout: status = %v", rr.Code)
	}

	t.Log("Integration test completed successfully")
}
