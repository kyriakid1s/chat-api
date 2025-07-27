package middleware

import (
	"go-chat-api/internal/auth"
	"go-chat-api/internal/models"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Define a custom type for context keys to avoid collisions
type contextKey string

const userIDKey contextKey = "userID"
const usernameKey contextKey = "username"

func TestAuthMiddleware(t *testing.T) {
	authService := auth.NewAuthService("test-secret", 24*time.Hour)

	// Create a test user and generate a token
	user := models.User{
		ID:       "test-user-id",
		Username: "testuser",
	}

	validToken, _, err := authService.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	// Create a test handler that checks for user context
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID")
		username := r.Context().Value("username")

		if userID == nil || username == nil {
			t.Error("AuthMiddleware did not set user context")
		}

		if userID.(string) != user.ID {
			t.Errorf("AuthMiddleware userID = %v, want %v", userID, user.ID)
		}

		if username.(string) != user.Username {
			t.Errorf("AuthMiddleware username = %v, want %v", username, user.Username)
		}

		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(authService)
	protectedHandler := middleware(testHandler)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid authorization format",
			authHeader:     "InvalidFormat " + validToken,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "missing bearer prefix",
			authHeader:     validToken,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid.token.format",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "expired token",
			authHeader:     "Bearer " + generateExpiredToken(authService, user),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			protectedHandler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("AuthMiddleware status = %v, want %v", rr.Code, tt.expectedStatus)
			}
		})
	}
}

func TestOptionalAuthMiddleware(t *testing.T) {
	authService := auth.NewAuthService("test-secret", 24*time.Hour)

	// Create a test user and generate a token
	user := models.User{
		ID:       "test-user-id",
		Username: "testuser",
	}

	validToken, _, err := authService.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate test token: %v", err)
	}

	// Create a test handler that checks for user context
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID")
		username := r.Context().Value("username")

		// For optional auth, we always proceed but may or may not have user context
		if userID != nil && username != nil {
			// User is authenticated
			if userID.(string) != user.ID {
				t.Errorf("OptionalAuthMiddleware userID = %v, want %v", userID, user.ID)
			}

			if username.(string) != user.Username {
				t.Errorf("OptionalAuthMiddleware username = %v, want %v", username, user.Username)
			}
		}

		w.WriteHeader(http.StatusOK)
	})

	middleware := OptionalAuthMiddleware(authService)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectUser     bool
	}{
		{
			name:           "valid token",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
			expectUser:     true,
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedStatus: http.StatusOK,
			expectUser:     false,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid.token.format",
			expectedStatus: http.StatusOK,
			expectUser:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/optional", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Create a handler that checks if user context is set correctly
			contextCheckHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userID := r.Context().Value("userID")

				if tt.expectUser && userID == nil {
					t.Error("OptionalAuthMiddleware should have set user context")
				}

				if !tt.expectUser && userID != nil {
					t.Error("OptionalAuthMiddleware should not have set user context for invalid token")
				}

				testHandler.ServeHTTP(w, r)
			})

			contextCheckMiddleware := middleware(contextCheckHandler)

			rr := httptest.NewRecorder()
			contextCheckMiddleware.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("OptionalAuthMiddleware status = %v, want %v", rr.Code, tt.expectedStatus)
			}
		})
	}
}

func TestCORSMiddleware(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	corsHandler := CORSMiddleware(testHandler)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		checkHeaders   bool
	}{
		{
			name:           "OPTIONS request",
			method:         http.MethodOptions,
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
		{
			name:           "GET request",
			method:         http.MethodGet,
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
		{
			name:           "POST request",
			method:         http.MethodPost,
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			rr := httptest.NewRecorder()

			corsHandler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("CORSMiddleware status = %v, want %v", rr.Code, tt.expectedStatus)
			}

			if tt.checkHeaders {
				expectedHeaders := map[string]string{
					"Access-Control-Allow-Origin":  "*",
					"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
					"Access-Control-Allow-Headers": "Content-Type, Authorization",
				}

				for header, expectedValue := range expectedHeaders {
					actualValue := rr.Header().Get(header)
					if actualValue != expectedValue {
						t.Errorf("CORSMiddleware %s = %v, want %v", header, actualValue, expectedValue)
					}
				}
			}
		})
	}
}

func TestLoggingMiddleware(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	loggingHandler := LoggingMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	// The logging middleware should not affect the response
	loggingHandler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("LoggingMiddleware status = %v, want %v", rr.Code, http.StatusOK)
	}
}

// Helper function to generate an expired token for testing
func generateExpiredToken(authService *auth.AuthService, user models.User) string {
	// Create an auth service with negative expiry to generate expired token
	expiredAuthService := auth.NewAuthService("test-secret", -1*time.Hour)
	token, _, err := expiredAuthService.GenerateToken(user)
	if err != nil {
		return "invalid.expired.token"
	}
	return token
}
