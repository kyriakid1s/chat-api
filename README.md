# Go## Features

- 🔐 **JWT Authentication** - Secure user registration and login with cookie support
- 🍪 **Cookie Authentication** - Automatic token handling for easy API testing
- 💬 **Send and receive messages** - Real-time messaging capabilities
- 👥 **User management** - User profiles and online status
- 🏠 **Chat rooms** - Create and manage chat rooms
- 🔒 **Protected endpoints** - Role-based access control
- 🐘 **PostgreSQL database** - Persistent data storage with clean schema
- 🏗️ **Clean architecture** - Dependency injection patterns
- 🌐 **CORS support** - Cross-origin resource sharing with credential support
- 📝 **Logging middleware** - Request/response loggingA RESTful chat application API built with Go, featuring JWT authentication, clean architecture principles and dependency injection patterns.

## Features

- 🔐 **JWT Authentication** - Secure user registration and login
- 💬 **Send and receive messages** - Real-time messaging capabilities
- 👥 **User management** - User profiles and online status
- 🏠 **Chat rooms** - Create and manage chat rooms
- 🔒 **Protected endpoints** - Role-based access control
- � **PostgreSQL database** - Persistent data storage with clean schema
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
│       ├── memory.go          # In-memory storage implementation (for testing)
│       └── postgres.go        # PostgreSQL storage implementation (default)
├── go.mod
└── README.md
```

## API Endpoints

### Authentication (Public)
- `POST /api/auth/register` - Register a new user
- `POST /api/auth/login` - Login user and get JWT token
- `POST /api/auth/refresh` - Refresh JWT token (accepts header or cookie auth, refreshes within 15min of expiry)

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
- Docker and Docker Compose (for PostgreSQL database)

### Quick Start

1. Clone the repository or copy the files to your project directory

2. Run the setup script (recommended):
```bash
./scripts/setup.sh
```

This script will:
- Start a PostgreSQL database using Docker Compose
- Create the .env configuration file
- Install Go dependencies
- Build the application

3. Or set up manually:

   a. Start PostgreSQL database:
   ```bash
   docker-compose up -d postgres
   ```

   b. Copy environment configuration:
   ```bash
   cp .env.example .env
   ```

   c. Install dependencies:
   ```bash
   go mod tidy
   ```

4. Start the application:
```bash
go run cmd/main.go
```

The server will start on port 8080 by default.

### Configuration

The application uses PostgreSQL as the default database. Configure it using environment variables:

**Database Configuration:**
- `DATABASE_URL` - Full PostgreSQL connection string (takes precedence if set)
- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_NAME` - Database name (default: chatapi)
- `DB_USER` - Database user (default: postgres)
- `DB_PASSWORD` - Database password (default: postgres)
- `DB_SSLMODE` - SSL mode (default: disable)

**Server Configuration:**
- `PORT` - Server port (default: 8080)
- `ENVIRONMENT` - Environment (default: development)
- `LOG_LEVEL` - Log level (default: info)

**JWT Configuration:**
- `JWT_SECRET` - JWT signing secret (default: "your-secret-key-change-this-in-production")
- `JWT_EXPIRY_HOURS` - JWT token expiry in hours (default: 24)

Copy `.env.example` to `.env` and customize as needed.

## Example Usage

### Authentication Methods

The API supports **two authentication methods**:

1. **Bearer Token** (in Authorization header)
2. **HTTP-only Cookie** (automatic for browsers/tools like Postman)

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

### 2. Login and get JWT token + cookie
```bash
# Option A: Save cookie to file for easy reuse
curl -c cookies.txt -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "password": "password123"
  }'

# Option B: Regular login (extract token from JSON response)
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

**🍪 Cookie Details:**
- **Name**: `jwt_token`
- **Security**: `HttpOnly`, `SameSite=Lax`
- **Expiry**: 24 hours (configurable)
- **Path**: `/` (all endpoints)

### 3. Access protected endpoints

**Method A: Using saved cookie (easiest for testing)**
```bash
# Send message using cookie authentication
curl -b cookies.txt -X POST http://localhost:8080/api/messages \
  -H "Content-Type: application/json" \
  -d '{
    "sender": "john_doe",
    "recipient": "jane_doe",
    "content": "Hello!"
  }'

# Get user profile using cookie
curl -b cookies.txt http://localhost:8080/api/auth/profile
```

**Method B: Using Bearer token**
```bash
# Send message using Authorization header
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

**Using cookie:**
```bash
curl -b cookies.txt http://localhost:8080/api/auth/profile
```

**Using Bearer token:**
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/auth/profile
```

### 5. Logout (clears cookie and invalidates session)

**Using cookie:**
```bash
curl -b cookies.txt -c cookies.txt -X POST http://localhost:8080/api/auth/logout
```

**Using Bearer token:**
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -X POST http://localhost:8080/api/auth/logout
```

**Note:** After logout, the `jwt_token` cookie is cleared automatically.

### 6. Get all messages (requires authentication)

**Using cookie:**
```bash
curl -b cookies.txt http://localhost:8080/api/messages
```

**Using Bearer token:**
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  http://localhost:8080/api/messages
```

## 🔄 Token Refresh Guide

### When to Use Refresh Token

The refresh endpoint (`POST /api/auth/refresh`) should be used to extend your session without requiring the user to login again.

**Refresh Window**: Tokens can only be refreshed within the **last 15 minutes** before expiry (security feature).

### Refresh Token Examples

**Method A: Cookie-based refresh (Recommended)**
```bash
# Refresh using saved cookie - token automatically updates
curl -b cookies.txt -c cookies.txt -X POST http://localhost:8080/api/auth/refresh
```

**Method B: Header-based refresh**
```bash
# Extract current token and refresh it
TOKEN=$(curl -b cookies.txt http://localhost:8080/api/auth/profile 2>/dev/null | jq -r '.token // empty')
curl -H "Authorization: Bearer $TOKEN" -X POST http://localhost:8080/api/auth/refresh
```

### Frontend Integration Examples

**JavaScript (Cookie-based - Easiest)**
```javascript
// Automatic refresh - no token management needed
async function refreshToken() {
  try {
    const response = await fetch('/api/auth/refresh', {
      method: 'POST',
      credentials: 'include' // Include cookies automatically
    });
    
    if (response.ok) {
      console.log('Token refreshed successfully');
      return true;
    } else {
      console.log('Refresh failed:', await response.text());
      return false;
    }
  } catch (error) {
    console.error('Refresh error:', error);
    return false;
  }
}

// Auto-refresh before API calls
async function makeApiCall(endpoint, options = {}) {
  // First, try the API call
  let response = await fetch(endpoint, {
    ...options,
    credentials: 'include'
  });
  
  // If unauthorized, try to refresh and retry
  if (response.status === 401) {
    const refreshed = await refreshToken();
    if (refreshed) {
      // Retry the original request
      response = await fetch(endpoint, {
        ...options,
        credentials: 'include'
      });
    } else {
      // Redirect to login
      window.location.href = '/login';
      return;
    }
  }
  
  return response;
}
```

**JavaScript (Header-based)**
```javascript
// Manual token management
class ApiClient {
  constructor() {
    this.token = localStorage.getItem('jwt_token');
    this.tokenExpiry = localStorage.getItem('token_expiry');
  }
  
  async refreshToken() {
    if (!this.token) return false;
    
    try {
      const response = await fetch('/api/auth/refresh', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${this.token}`
        }
      });
      
      if (response.ok) {
        const data = await response.json();
        this.token = data.token;
        this.tokenExpiry = data.expires_at;
        
        localStorage.setItem('jwt_token', this.token);
        localStorage.setItem('token_expiry', this.tokenExpiry);
        return true;
      }
    } catch (error) {
      console.error('Refresh failed:', error);
    }
    
    return false;
  }
  
  async apiCall(endpoint, options = {}) {
    // Check if token needs refresh (within 5 minutes of expiry)
    const now = Date.now() / 1000;
    if (this.tokenExpiry && (this.tokenExpiry - now < 300)) {
      await this.refreshToken();
    }
    
    return fetch(endpoint, {
      ...options,
      headers: {
        ...options.headers,
        'Authorization': `Bearer ${this.token}`
      }
    });
  }
}
```

### Best Practices

1. **Automatic Refresh**: Check token expiry before API calls
2. **Graceful Degradation**: Handle refresh failures by redirecting to login
3. **Cookie Method**: Easier to implement and more secure
4. **Timing**: Refresh tokens when they have < 15 minutes remaining
5. **Error Handling**: Always handle 401 responses with refresh retry

### Common Refresh Scenarios

**Scenario 1: Long-running web application**
```javascript
// Check every 5 minutes if refresh is needed
setInterval(async () => {
  const tokenExpiry = localStorage.getItem('token_expiry');
  const now = Date.now() / 1000;
  
  // Refresh if expiring in next 10 minutes
  if (tokenExpiry && (tokenExpiry - now < 600)) {
    await refreshToken();
  }
}, 5 * 60 * 1000); // Every 5 minutes
```

**Scenario 2: Before critical operations**
```javascript
async function sendImportantMessage(messageData) {
  // Ensure fresh token before important operations
  await refreshToken();
  
  return fetch('/api/messages', {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(messageData)
  });
}
```

**Scenario 3: Mobile app background refresh**
```javascript
// React Native example
import { AppState } from 'react-native';

AppState.addEventListener('change', async (nextAppState) => {
  if (nextAppState === 'active') {
    // App came to foreground, check if token needs refresh
    await refreshToken();
  }
});
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

### Database Schema

The application automatically creates the following PostgreSQL tables:

- `users` - User accounts with authentication data
- `chat_rooms` - Chat room definitions
- `room_members` - Many-to-many relationship between users and rooms
- `messages` - Chat messages with sender, recipient, and room information

The schema is automatically created when the application starts.

### Switching to Different Databases

The storage layer uses interfaces, making it easy to implement different databases:

1. Implement the storage interfaces (`MessageStore`, `UserStore`, `RoomStore`) for your database
2. Replace the PostgreSQL storage in `main.go` with your implementation
3. No other code changes needed!

### Adding In-Memory Storage (for testing)

The project includes an in-memory storage implementation. To use it:

```go
// In main.go, replace the PostgreSQL storage with:
storage := storage.NewInMemoryStorage()
chatService := services.NewChatService(storage, storage, storage, authService)
```

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
- `github.com/lib/pq` - PostgreSQL driver
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

You can test the complete authentication flow using the curl examples above, or use tools like Postman, Insomnia, or any HTTP client.

### Testing with Postman/Insomnia

**Cookie Authentication (Recommended):**
1. Register a new user via `POST /api/auth/register`
2. Login via `POST /api/auth/login` 
3. 🍪 **Cookies are automatically handled** - no manual token copying needed!
4. Make requests to protected endpoints - authentication works seamlessly

**Bearer Token Authentication:**
1. Register a new user first
2. Login to get a JWT token from the JSON response
3. Copy the `token` field and add it to Authorization header: `Bearer YOUR_TOKEN`
4. Include the token in all protected endpoint requests
5. Handle token expiry by refreshing or re-authenticating

### Testing with curl

**Option 1: Cookie-based (saves cookies to file)**
```bash
# Save cookies during login
curl -c cookies.txt -X POST http://localhost:8080/api/auth/login -d '{"username":"user","password":"pass"}'

# Use saved cookies for subsequent requests
curl -b cookies.txt http://localhost:8080/api/auth/profile

# Test refresh token (only works when token expires in <15 minutes)
curl -b cookies.txt -c cookies.txt -X POST http://localhost:8080/api/auth/refresh
```

**Option 2: Token-based (manual token handling)**
```bash
# Extract token from login response and use in header
TOKEN=$(curl -X POST http://localhost:8080/api/auth/login -d '{"username":"user","password":"pass"}' | jq -r '.token')
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/auth/profile

# Test refresh token
NEW_TOKEN=$(curl -H "Authorization: Bearer $TOKEN" -X POST http://localhost:8080/api/auth/refresh | jq -r '.token')
```

### Testing Refresh Token Functionality

**⚠️ Note**: Refresh only works when tokens are within 15 minutes of expiry.

**For Development Testing**, you can temporarily modify the refresh window:

1. Edit `internal/auth/auth.go` line 93:
```go
// Change from 15 minutes to 23 hours for testing
if time.Until(claims.ExpiresAt.Time) > 23*time.Hour {
```

2. Restart server and test:
```bash
# Login and save cookie
curl -c cookies.txt -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"password123"}'

# Test refresh (should work with modified time window)
curl -b cookies.txt -c cookies.txt -X POST http://localhost:8080/api/auth/refresh

# Verify cookie was updated
cat cookies.txt
```

3. **Remember to revert** the change for production!

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

## Database Management

### Starting the Database

```bash
# Start PostgreSQL with Docker Compose
docker-compose up -d postgres

# View database logs
docker-compose logs postgres

# Stop the database
docker-compose down
```

### Connecting to the Database

```bash
# Connect using psql
docker-compose exec postgres psql -U postgres -d chatapi

# Or connect from your host machine
psql -h localhost -p 5432 -U postgres -d chatapi
```

### Database Operations

The application automatically creates all necessary tables and indexes when it starts. The database schema includes:

- Proper foreign key constraints
- Indexes for optimal query performance  
- Timestamps with timezone support
- Unique constraints for usernames and emails

### Backup and Restore

```bash
# Create a backup
docker-compose exec postgres pg_dump -U postgres chatapi > backup.sql

# Restore from backup
docker-compose exec -T postgres psql -U postgres chatapi < backup.sql
```
