package main

import (
	"go-chat-api/internal/auth"
	"go-chat-api/internal/config"
	"go-chat-api/internal/handlers"
	"go-chat-api/internal/middleware"
	"go-chat-api/internal/routes"
	"go-chat-api/internal/services"
	"go-chat-api/internal/storage"
	"log"
	"net/http"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize storage (dependency injection)
	storage := storage.NewInMemoryStorage()

	// Initialize auth service
	authService := auth.NewAuthService(cfg.JWTSecret, cfg.JWTExpiry)

	// Initialize services with dependency injection
	chatService := services.NewChatService(storage, storage, storage, authService)

	// Initialize handlers with dependency injection
	chatHandler := handlers.NewChatHandler(chatService)
	authHandler := handlers.NewAuthHandler(chatService)

	// Setup routes
	router := routes.SetupRoutes(chatHandler, authHandler, authService)

	// Add middleware
	handler := middleware.LoggingMiddleware(middleware.CORSMiddleware(router))

	// Start server
	log.Printf("Starting chat API server on port %s", cfg.Port)
	log.Printf("Environment: %s", cfg.Environment)

	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
