package middleware

import (
	"bufio"
	"context"
	"go-chat-api/internal/auth"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom ResponseWriter to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		log.Printf("%s %s %d %v", r.Method, r.URL.Path, wrapped.statusCode, time.Since(start))
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Hijack implements the http.Hijacker interface for WebSocket support
func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

// CORSMiddleware handles CORS headers
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow multiple origins for development
		origin := r.Header.Get("Origin")
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:8080",
			"http://127.0.0.1:8080",
		}

		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}

		// If no origin header (same-origin requests), allow it
		if origin == "" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true") // Allow cookies
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// AuthMiddleware validates JWT tokens and adds user context
func AuthMiddleware(authService *auth.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tokenString string

			// First, try to get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
				if tokenString == authHeader {
					tokenString = "" // Invalid format
				}
			}

			// If no valid token in header, try to get from cookie
			if tokenString == "" {
				if cookie, err := r.Cookie("jwt_token"); err == nil {
					tokenString = cookie.Value
				}
			}

			// If still no token found, return unauthorized
			if tokenString == "" {
				http.Error(w, "Authorization required (provide Bearer token or jwt_token cookie)", http.StatusUnauthorized)
				return
			}

			claims, err := authService.ValidateToken(tokenString)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Add user information to request context
			ctx := context.WithValue(r.Context(), "userID", claims.UserID)
			ctx = context.WithValue(ctx, "username", claims.Username)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// OptionalAuthMiddleware validates JWT tokens but doesn't require them
func OptionalAuthMiddleware(authService *auth.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tokenString string

			// First, try to get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
				if tokenString == authHeader {
					tokenString = "" // Invalid format
				}
			}

			// If no valid token in header, try to get from cookie
			if tokenString == "" {
				if cookie, err := r.Cookie("jwt_token"); err == nil {
					tokenString = cookie.Value
				}
			}

			// If token found, validate it and add to context
			if tokenString != "" {
				claims, err := authService.ValidateToken(tokenString)
				if err == nil {
					// Add user information to request context
					ctx := context.WithValue(r.Context(), "userID", claims.UserID)
					ctx = context.WithValue(ctx, "username", claims.Username)
					r = r.WithContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
