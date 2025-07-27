package auth

import (
	"errors"
	"go-chat-api/internal/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication operations
type AuthService struct {
	jwtSecret []byte
	jwtExpiry time.Duration
}

// NewAuthService creates a new authentication service
func NewAuthService(jwtSecret string, jwtExpiry time.Duration) *AuthService {
	return &AuthService{
		jwtSecret: []byte(jwtSecret),
		jwtExpiry: jwtExpiry,
	}
}

// HashPassword hashes a password using bcrypt
func (s *AuthService) HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// VerifyPassword verifies a password against its hash
func (s *AuthService) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// GenerateToken generates a JWT token for a user
func (s *AuthService) GenerateToken(user models.User) (string, int64, error) {
	expirationTime := time.Now().Add(s.jwtExpiry)

	claims := &models.Claims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", 0, err
	}

	return tokenString, expirationTime.Unix(), nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *AuthService) ValidateToken(tokenString string) (*models.Claims, error) {
	claims := &models.Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// RefreshToken generates a new token from an existing valid token
func (s *AuthService) RefreshToken(tokenString string) (string, int64, error) {
	claims, err := s.ValidateToken(tokenString)
	if err != nil {
		return "", 0, err
	}

	// Check if token is close to expiry (within 15 minutes)
	if time.Until(claims.ExpiresAt.Time) > 15*time.Minute {
		return "", 0, errors.New("token not eligible for refresh")
	}

	// Create new token with same user info
	user := models.User{
		ID:       claims.UserID,
		Username: claims.Username,
	}

	return s.GenerateToken(user)
}
