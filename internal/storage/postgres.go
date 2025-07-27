package storage

import (
	"database/sql"
	"fmt"
	"go-chat-api/internal/models"
	"strings"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

// PostgresDB wraps a database connection and implements all storage interfaces
type PostgresDB struct {
	db *sql.DB
}

// NewPostgresDB creates a new PostgreSQL database connection and returns storage interfaces
func NewPostgresDB(connectionString string) (*PostgresDB, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	pgDB := &PostgresDB{db: db}

	// Create tables if they don't exist
	if err := pgDB.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return pgDB, nil
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	return p.db.Close()
}

// createTables creates the necessary tables for the application
func (p *PostgresDB) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(255) PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			is_online BOOLEAN DEFAULT false,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS chat_rooms (
			id VARCHAR(255) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS room_members (
			room_id VARCHAR(255) REFERENCES chat_rooms(id) ON DELETE CASCADE,
			user_id VARCHAR(255) REFERENCES users(id) ON DELETE CASCADE,
			joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			PRIMARY KEY (room_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id VARCHAR(255) PRIMARY KEY,
			sender VARCHAR(255) NOT NULL REFERENCES users(username),
			recipient VARCHAR(255) REFERENCES users(username),
			content TEXT NOT NULL,
			timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			room_id VARCHAR(255) REFERENCES chat_rooms(id) ON DELETE SET NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_sender ON messages(sender)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_recipient ON messages(recipient)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_room_id ON messages(room_id)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
	}

	for _, query := range queries {
		if _, err := p.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query %s: %w", query, err)
		}
	}

	return nil
}

// MessageStore implementation

// AddMessage adds a new message to the database
func (p *PostgresDB) AddMessage(message models.Message) error {
	query := `
		INSERT INTO messages (id, sender, recipient, content, timestamp, room_id)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	var roomID interface{}
	if message.RoomID == "" {
		roomID = nil
	} else {
		roomID = message.RoomID
	}

	var recipient interface{}
	if message.Recipient == "" {
		recipient = nil
	} else {
		recipient = message.Recipient
	}

	_, err := p.db.Exec(query, message.ID, message.Sender, recipient,
		message.Content, message.Timestamp, roomID)
	if err != nil {
		return fmt.Errorf("failed to add message: %w", err)
	}
	return nil
} // GetMessages retrieves all messages from the database
func (p *PostgresDB) GetMessages() ([]models.Message, error) {
	query := `
		SELECT id, sender, COALESCE(recipient, '') as recipient, content, timestamp, COALESCE(room_id, '') as room_id
		FROM messages
		ORDER BY timestamp ASC
	`
	rows, err := p.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var message models.Message
		if err := rows.Scan(&message.ID, &message.Sender, &message.Recipient,
			&message.Content, &message.Timestamp, &message.RoomID); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}

	return messages, nil
}

// GetMessagesByRoom retrieves messages for a specific room
func (p *PostgresDB) GetMessagesByRoom(roomID string) ([]models.Message, error) {
	query := `
		SELECT id, sender, recipient, content, timestamp, COALESCE(room_id, '') as room_id
		FROM messages
		WHERE room_id = $1
		ORDER BY timestamp ASC
	`
	rows, err := p.db.Query(query, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by room: %w", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var message models.Message
		if err := rows.Scan(&message.ID, &message.Sender, &message.Recipient,
			&message.Content, &message.Timestamp, &message.RoomID); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}

	return messages, nil
}

// GetMessagesBetweenUsers retrieves messages between two users
func (p *PostgresDB) GetMessagesBetweenUsers(user1, user2 string) ([]models.Message, error) {
	query := `
		SELECT id, sender, recipient, content, timestamp, COALESCE(room_id, '') as room_id
		FROM messages
		WHERE (sender = $1 AND recipient = $2) OR (sender = $2 AND recipient = $1)
		ORDER BY timestamp ASC
	`
	rows, err := p.db.Query(query, user1, user2)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages between users: %w", err)
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var message models.Message
		if err := rows.Scan(&message.ID, &message.Sender, &message.Recipient,
			&message.Content, &message.Timestamp, &message.RoomID); err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, message)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}

	return messages, nil
}

// UserStore implementation

// AddUser adds a new user to the database
func (p *PostgresDB) AddUser(user models.User) error {
	query := `
		INSERT INTO users (id, username, email, password_hash, is_online, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := p.db.Exec(query, user.ID, user.Username, user.Email,
		user.PasswordHash, user.IsOnline, user.CreatedAt)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if strings.Contains(pqErr.Message, "username") {
					return fmt.Errorf("username already exists")
				}
				if strings.Contains(pqErr.Message, "email") {
					return fmt.Errorf("email already exists")
				}
			}
		}
		return fmt.Errorf("failed to add user: %w", err)
	}
	return nil
}

// GetUser retrieves a user by ID
func (p *PostgresDB) GetUser(userID string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, is_online, created_at
		FROM users
		WHERE id = $1
	`
	var user models.User
	err := p.db.QueryRow(query, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.IsOnline, &user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (p *PostgresDB) GetUserByUsername(username string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, is_online, created_at
		FROM users
		WHERE username = $1
	`
	var user models.User
	err := p.db.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.IsOnline, &user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (p *PostgresDB) GetUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, username, email, password_hash, is_online, created_at
		FROM users
		WHERE email = $1
	`
	var user models.User
	err := p.db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.IsOnline, &user.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &user, nil
}

// UpdateUserStatus updates a user's online status
func (p *PostgresDB) UpdateUserStatus(userID string, isOnline bool) error {
	query := `UPDATE users SET is_online = $1 WHERE id = $2`
	result, err := p.db.Exec(query, isOnline, userID)
	if err != nil {
		return fmt.Errorf("failed to update user status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// GetAllUsers retrieves all users
func (p *PostgresDB) GetAllUsers() ([]models.User, error) {
	query := `
		SELECT id, username, email, password_hash, is_online, created_at
		FROM users
		ORDER BY created_at ASC
	`
	rows, err := p.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email,
			&user.PasswordHash, &user.IsOnline, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// RoomStore implementation

// CreateRoom creates a new chat room
func (p *PostgresDB) CreateRoom(room models.ChatRoom) error {
	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create the room
	query := `
		INSERT INTO chat_rooms (id, name, description, created_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err = tx.Exec(query, room.ID, room.Name, room.Description, room.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create room: %w", err)
	}

	// Add members to the room
	if len(room.Members) > 0 {
		memberQuery := `INSERT INTO room_members (room_id, user_id) VALUES ($1, $2)`
		for _, memberID := range room.Members {
			_, err = tx.Exec(memberQuery, room.ID, memberID)
			if err != nil {
				return fmt.Errorf("failed to add member to room: %w", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetRoom retrieves a room by ID
func (p *PostgresDB) GetRoom(roomID string) (*models.ChatRoom, error) {
	// Get room details
	query := `
		SELECT id, name, description, created_at
		FROM chat_rooms
		WHERE id = $1
	`
	var room models.ChatRoom
	err := p.db.QueryRow(query, roomID).Scan(
		&room.ID, &room.Name, &room.Description, &room.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	// Get room members
	memberQuery := `
		SELECT user_id
		FROM room_members
		WHERE room_id = $1
	`
	rows, err := p.db.Query(memberQuery, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get room members: %w", err)
	}
	defer rows.Close()

	var members []string
	for rows.Next() {
		var memberID string
		if err := rows.Scan(&memberID); err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}
		members = append(members, memberID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating members: %w", err)
	}

	room.Members = members
	return &room, nil
}

// GetRoomsByUser retrieves rooms that a user is a member of
func (p *PostgresDB) GetRoomsByUser(userID string) ([]models.ChatRoom, error) {
	query := `
		SELECT r.id, r.name, r.description, r.created_at
		FROM chat_rooms r
		INNER JOIN room_members rm ON r.id = rm.room_id
		WHERE rm.user_id = $1
		ORDER BY r.created_at ASC
	`
	rows, err := p.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get rooms by user: %w", err)
	}
	defer rows.Close()

	var rooms []models.ChatRoom
	for rows.Next() {
		var room models.ChatRoom
		if err := rows.Scan(&room.ID, &room.Name, &room.Description, &room.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan room: %w", err)
		}

		// Get members for each room
		memberQuery := `
			SELECT user_id
			FROM room_members
			WHERE room_id = $1
		`
		memberRows, err := p.db.Query(memberQuery, room.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get room members: %w", err)
		}

		var members []string
		for memberRows.Next() {
			var memberID string
			if err := memberRows.Scan(&memberID); err != nil {
				memberRows.Close()
				return nil, fmt.Errorf("failed to scan member: %w", err)
			}
			members = append(members, memberID)
		}
		memberRows.Close()

		room.Members = members
		rooms = append(rooms, room)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rooms: %w", err)
	}

	return rooms, nil
}

// AddUserToRoom adds a user to a room
func (p *PostgresDB) AddUserToRoom(roomID, userID string) error {
	query := `
		INSERT INTO room_members (room_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (room_id, user_id) DO NOTHING
	`
	_, err := p.db.Exec(query, roomID, userID)
	if err != nil {
		return fmt.Errorf("failed to add user to room: %w", err)
	}
	return nil
}

// RemoveUserFromRoom removes a user from a room
func (p *PostgresDB) RemoveUserFromRoom(roomID, userID string) error {
	query := `DELETE FROM room_members WHERE room_id = $1 AND user_id = $2`
	result, err := p.db.Exec(query, roomID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove user from room: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found in room")
	}

	return nil
}
