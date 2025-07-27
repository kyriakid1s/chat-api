# Go Chat API

A RESTful chat application API built with Go, featuring JWT authentication, clean architecture principles and dependency injection patterns.

## Features

- 🔐 **JWT Authentication** - Secure user registration and login
- 💬 **Send and receive messages** - Real-time messaging capabilities
- 👥 **User management** - User profiles and online status
- 🏠 **Chat rooms** - Create and manage chat rooms
- 🔒 **Protected endpoints** - Role-based access control
- 💾 **In-memory storage** - Easily replaceable with database
- 🏗️ **Clean architecture** - Dependency injection patterns
- 🌐 **CORS support** - Cross-origin resource sharing
- 📝 **Logging middleware** - Request/response logging

## Project Structure

```
go-chat-api/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── auth/
│   │   └── auth.go            # JWT authentication service
│   ├── config/
│   │   └── config.go          # Configuration management
│   ├── handlers/
│   │   ├── auth_handler.go    # Authentication HTTP handlers
│   │   └── chat_handler.go    # Chat HTTP handlers
│   ├── middleware/
│   │   └── middleware.go      # HTTP middleware (auth, CORS, logging)
│   ├── models/
│   │   └── models.go          # Data models and DTOs
│   ├── routes/
│   │   └── routes.go          # Route definitions with auth protection
│   ├── services/
│   │   └── chat_service.go    # Business logic layer
│   └── storage/
│       ├── interfaces.go      # Storage abstractions
│       └── memory.go          # In-memory storage implementation
├── go.mod
└── README.md
```

## API Endpoints

### Authentication (Public)
- `POST /api/auth/register` - Register a new user
- `POST /api/auth/login` - Login user and get JWT token
- `POST /api/auth/refresh` - Refresh JWT token

### Authentication (Protected)
- `POST /api/auth/logout` - Logout user (requires authentication)
- `GET /api/auth/profile` - Get current user profile (requires authentication)

### Messages (Protected - requires JWT token)
- `POST /api/messages` - Send a message
- `GET /api/messages` - Get all messages
- `GET /api/messages/between/{user1}/{user2}` - Get messages between two users

### Users (Protected - requires JWT token)
- `GET /api/users` - Get all users
- `GET /api/users/{userId}` - Get user by ID
- `GET /api/users/{userId}/rooms` - Get user's rooms

### Rooms (Protected - requires JWT token)
- `POST /api/rooms` - Create a room
- `GET /api/rooms/{roomId}` - Get room by ID
- `GET /api/rooms/{roomId}/messages` - Get room messages
- `POST /api/rooms/{roomId}/members/{userId}` - Add user to room
- `DELETE /api/rooms/{roomId}/members/{userId}` - Remove user from room

## Getting Started

### Prerequisites
- Go 1.21 or higher

### Installation

1. Clone the repository or copy the files to your project directory

2. Install dependencies:
```bash
go mod tidy
```

3. Run the application:
```bash
go run cmd/main.go
```

The server will start on port 8080 by default.

### Configuration

Set environment variables to configure the application:

- `PORT` - Server port (default: 8080)
- `ENVIRONMENT` - Environment (default: development)
- `LOG_LEVEL` - Log level (default: info)
- `JWT_SECRET` - JWT signing secret (default: "your-secret-key-change-this-in-production")
- `JWT_EXPIRY_HOURS` - JWT token expiry in hours (default: 24)

## Example Usage

### 1. Register a new user
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "email": "john@example.com",
    "password": "password123"
  }'
```

### 2. Login and get JWT token
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "password": "password123"
  }'
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "user_id",
    "username": "john_doe",
    "email": "john@example.com",
    "is_online": true,
    "created_at": "2025-07-27T17:00:00Z"
  },
  "expires_at": 1690490400
}
```

### 3. Send a message (requires authentication)
```bash
curl -X POST http://localhost:8080/api/messages \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "sender": "john_doe",
    "recipient": "jane_doe",
    "content": "Hello!"
  }'
```

### 4. Get user profile (requires authentication)
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/auth/profile
```

### 5. Get all messages (requires authentication)
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/messages
```

## Architecture

This project follows clean architecture principles:

- **Models**: Define data structures
- **Storage**: Abstract storage interfaces with in-memory implementation
- **Services**: Business logic layer
- **Handlers**: HTTP request handling
- **Routes**: API route definitions
- **Middleware**: Cross-cutting concerns (logging, CORS)

The dependency injection pattern makes it easy to:
- Test individual components
- Swap implementations (e.g., replace in-memory storage with database)
- Maintain loose coupling between layers

## Extending the Application

### Adding Database Support

1. Implement the storage interfaces (`MessageStore`, `UserStore`, `RoomStore`) for your database
2. Replace the in-memory storage in `main.go` with your database implementation
3. No other code changes needed!

### Adding WebSocket Support

1. Create a WebSocket handler in the handlers package
2. Add WebSocket routes
3. Use the existing services for business logic

### Adding Authentication

1. Create an auth service
2. Add auth middleware
3. Inject user context into handlers

## Dependencies

- `github.com/gorilla/mux` - HTTP router and URL matcher
- `github.com/gorilla/websocket` - WebSocket support (ready for real-time features)
- `github.com/rs/cors` - CORS middleware
- `github.com/golang-jwt/jwt/v5` - JWT token handling
- `golang.org/x/crypto` - Password hashing with bcrypt

## Security Features

- **Password Hashing**: Uses bcrypt for secure password storage
- **JWT Authentication**: Stateless authentication with configurable expiry
- **Protected Routes**: Middleware-based route protection
- **Token Refresh**: Secure token refresh mechanism
- **Input Validation**: Request payload validation
- **CORS Protection**: Configurable cross-origin request handling

## Authentication Flow

1. **Register**: User creates account with username, email, and password
2. **Login**: User authenticates and receives JWT token
3. **Access Protected Routes**: Include JWT token in Authorization header
4. **Token Refresh**: Refresh token before expiry to maintain session
5. **Logout**: Update user status and invalidate session

## Testing the Authentication System

You can test the complete authentication flow using the curl examples above, or use tools like Postman, Insomnia, or any HTTP client. Make sure to:

1. Register a new user first
2. Login to get a JWT token
3. Include the token in the Authorization header for protected endpoints
4. Handle token expiry by refreshing or re-authenticating

### Running Tests

The project includes comprehensive tests for the authentication system:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Run specific package tests
go test ./internal/auth -v
go test ./internal/services -v
go test ./internal/handlers -v
go test ./internal/middleware -v
```

### Test Coverage

- **Auth Service**: 86.7% coverage - Tests password hashing, JWT operations, token validation
- **Middleware**: 100% coverage - Tests authentication, CORS, logging middleware
- **Services**: 51.8% coverage - Tests business logic, user registration, authentication flow
- **Handlers**: 36.5% coverage - Tests HTTP endpoints, request validation, response handling

### Test Structure

- **Unit Tests**: Each component is tested in isolation
- **Integration Tests**: Full authentication flow tests
- **Mock Dependencies**: In-memory storage for isolated testing
- **Edge Cases**: Invalid inputs, expired tokens, missing data
- **Security Tests**: Password verification, token validation, unauthorized access
