package routes

import (
	"go-chat-api/internal/auth"
	"go-chat-api/internal/handlers"
	"go-chat-api/internal/middleware"

	"github.com/gorilla/mux"
)

// SetupRoutes configures all API routes
func SetupRoutes(chatHandler *handlers.ChatHandler, authHandler *handlers.AuthHandler, authService *auth.AuthService) *mux.Router {
	router := mux.NewRouter()

	// API prefix
	api := router.PathPrefix("/api").Subrouter()

	// Public auth routes (no authentication required)
	auth := api.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/register", authHandler.Register).Methods("POST")
	auth.HandleFunc("/login", authHandler.Login).Methods("POST")
	auth.HandleFunc("/refresh", authHandler.RefreshToken).Methods("POST")

	// Protected auth routes (authentication required)
	authProtected := api.PathPrefix("/auth").Subrouter()
	authProtected.Use(middleware.AuthMiddleware(authService))
	authProtected.HandleFunc("/logout", authHandler.Logout).Methods("POST")
	authProtected.HandleFunc("/profile", authHandler.GetProfile).Methods("GET")

	// Protected message routes (authentication required)
	messages := api.PathPrefix("/messages").Subrouter()
	messages.Use(middleware.AuthMiddleware(authService))
	messages.HandleFunc("", chatHandler.SendMessage).Methods("POST")
	messages.HandleFunc("", chatHandler.GetMessages).Methods("GET")
	messages.HandleFunc("/between/{user1}/{user2}", chatHandler.GetMessagesBetweenUsers).Methods("GET")

	// Protected user routes (authentication required)
	users := api.PathPrefix("/users").Subrouter()
	users.Use(middleware.AuthMiddleware(authService))
	users.HandleFunc("", chatHandler.GetAllUsers).Methods("GET")
	users.HandleFunc("/{userId}", chatHandler.GetUser).Methods("GET")
	users.HandleFunc("/{userId}/rooms", chatHandler.GetRoomsByUser).Methods("GET")

	// Protected room routes (authentication required)
	rooms := api.PathPrefix("/rooms").Subrouter()
	rooms.Use(middleware.AuthMiddleware(authService))
	rooms.HandleFunc("", chatHandler.CreateRoom).Methods("POST")
	rooms.HandleFunc("/{roomId}", chatHandler.GetRoom).Methods("GET")
	rooms.HandleFunc("/{roomId}/messages", chatHandler.GetMessagesByRoom).Methods("GET")
	rooms.HandleFunc("/{roomId}/members/{userId}", chatHandler.AddUserToRoom).Methods("POST")
	rooms.HandleFunc("/{roomId}/members/{userId}", chatHandler.RemoveUserFromRoom).Methods("DELETE")

	return router
}
