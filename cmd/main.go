package main

import (
	"go-chat-api/internal/auth"
	"go-chat-api/internal/config"
	"go-chat-api/internal/handlers"
	"go-chat-api/internal/middleware"
	"go-chat-api/internal/routes"
	"go-chat-api/internal/services"
	"go-chat-api/internal/storage"
	"go-chat-api/internal/websocket"
	"log"
	"net/http"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize PostgreSQL storage
	db, err := storage.NewPostgresDB(cfg.GetDatabaseConnectionString())
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize WebSocket hub
	hub := websocket.NewHub()
	go hub.Run() // Start the hub in a goroutine

	// Initialize auth service
	authService := auth.NewAuthService(cfg.JWTSecret, cfg.JWTExpiry)

	// Initialize services with dependency injection
	chatService := services.NewChatService(db, db, db, authService)

	// Initialize handlers with dependency injection
	chatHandler := handlers.NewChatHandler(chatService, hub)
	authHandler := handlers.NewAuthHandler(chatService)
	wsHandler := handlers.NewWebSocketHandler(hub, chatService)

	// Setup routes
	router := routes.SetupRoutes(chatHandler, authHandler, wsHandler, authService)

	// Add middleware
	handler := middleware.LoggingMiddleware(middleware.CORSMiddleware(router))

	// Start server
	log.Printf("Starting chat API server on port %s", cfg.Port)
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("Database: Connected to PostgreSQL")
	log.Printf("WebSocket: Hub initialized and running")

	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
