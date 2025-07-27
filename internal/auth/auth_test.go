package auth

import (
	"go-chat-api/internal/models"
	"testing"
	"time"
)

func TestAuthService_HashPassword(t *testing.T) {
	authService := NewAuthService("test-secret", 24*time.Hour)

	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false, // bcrypt can hash empty strings
		},
		{
			name:     "long password",
			password: "very-long-password-with-special-characters-!@#$%^&*()",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := authService.HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && hash == "" {
				t.Error("HashPassword() returned empty hash")
			}
			if !tt.wantErr && hash == tt.password {
				t.Error("HashPassword() returned plaintext password")
			}
		})
	}
}

func TestAuthService_VerifyPassword(t *testing.T) {
	authService := NewAuthService("test-secret", 24*time.Hour)
	password := "test-password-123"

	// Hash the password first
	hashedPassword, err := authService.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	tests := []struct {
		name           string
		hashedPassword string
		password       string
		wantErr        bool
	}{
		{
			name:           "correct password",
			hashedPassword: hashedPassword,
			password:       password,
			wantErr:        false,
		},
		{
			name:           "incorrect password",
			hashedPassword: hashedPassword,
			password:       "wrong-password",
			wantErr:        true,
		},
		{
			name:           "empty password",
			hashedPassword: hashedPassword,
			password:       "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authService.VerifyPassword(tt.hashedPassword, tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("VerifyPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuthService_GenerateToken(t *testing.T) {
	authService := NewAuthService("test-secret", 24*time.Hour)

	user := models.User{
		ID:       "test-user-id",
		Username: "testuser",
		Email:    "test@example.com",
	}

	token, expiresAt, err := authService.GenerateToken(user)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	if token == "" {
		t.Error("GenerateToken() returned empty token")
	}

	if expiresAt <= time.Now().Unix() {
		t.Error("GenerateToken() returned expiry time in the past")
	}

	// Verify the token can be parsed
	claims, err := authService.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("Token UserID = %v, want %v", claims.UserID, user.ID)
	}

	if claims.Username != user.Username {
		t.Errorf("Token Username = %v, want %v", claims.Username, user.Username)
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	authService := NewAuthService("test-secret", 24*time.Hour)

	user := models.User{
		ID:       "test-user-id",
		Username: "testuser",
	}

	// Generate a valid token
	validToken, _, err := authService.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Generate an expired token
	expiredAuthService := NewAuthService("test-secret", -1*time.Hour)
	expiredToken, _, err := expiredAuthService.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate expired token: %v", err)
	}

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "valid token",
			token:   validToken,
			wantErr: false,
		},
		{
			name:    "expired token",
			token:   expiredToken,
			wantErr: true,
		},
		{
			name:    "invalid token format",
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
			claims, err := authService.ValidateToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if claims.UserID != user.ID {
					t.Errorf("ValidateToken() UserID = %v, want %v", claims.UserID, user.ID)
				}
				if claims.Username != user.Username {
					t.Errorf("ValidateToken() Username = %v, want %v", claims.Username, user.Username)
				}
			}
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	authService := NewAuthService("test-secret", 24*time.Hour)

	user := models.User{
		ID:       "test-user-id",
		Username: "testuser",
	}

	// Test with a token that's close to expiry (within refresh window)
	shortExpiryService := NewAuthService("test-secret", 10*time.Minute)
	refreshableToken, _, err := shortExpiryService.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate refreshable token: %v", err)
	}

	// Test with a token that's not close to expiry
	longExpiryToken, _, err := authService.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate long expiry token: %v", err)
	}

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "token eligible for refresh",
			token:   refreshableToken,
			wantErr: false,
		},
		{
			name:    "token not eligible for refresh",
			token:   longExpiryToken,
			wantErr: true,
		},
		{
			name:    "invalid token",
			token:   "invalid.token",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newToken, expiresAt, err := authService.RefreshToken(tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("RefreshToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if newToken == "" {
					t.Error("RefreshToken() returned empty token")
				}
				if expiresAt <= time.Now().Unix() {
					t.Error("RefreshToken() returned expiry time in the past")
				}
				// Verify the new token is valid
				_, err := authService.ValidateToken(newToken)
				if err != nil {
					t.Errorf("RefreshToken() generated invalid token: %v", err)
				}
			}
		})
	}
}

func TestAuthService_TokenSigning(t *testing.T) {
	// Test that tokens signed with different secrets are incompatible
	authService1 := NewAuthService("secret1", 24*time.Hour)
	authService2 := NewAuthService("secret2", 24*time.Hour)

	user := models.User{
		ID:       "test-user-id",
		Username: "testuser",
	}

	token1, _, err := authService1.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token with service1: %v", err)
	}

	// Try to validate token1 with service2 (different secret)
	_, err = authService2.ValidateToken(token1)
	if err == nil {
		t.Error("Expected error when validating token with different secret")
	}

	// Verify token1 is valid with service1
	_, err = authService1.ValidateToken(token1)
	if err != nil {
		t.Errorf("Token should be valid with original service: %v", err)
	}
}
