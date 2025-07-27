package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds application configuration
type Config struct {
	Port            string
	Environment     string
	LogLevel        string
	JWTSecret       string
	JWTExpiry       time.Duration
	DatabaseURL     string
	DatabaseHost    string
	DatabasePort    string
	DatabaseName    string
	DatabaseUser    string
	DatabasePass    string
	DatabaseSSLMode string
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	jwtExpiryHours := getEnvAsInt("JWT_EXPIRY_HOURS", 24)

	return &Config{
		Port:            getEnv("PORT", "8080"),
		Environment:     getEnv("ENVIRONMENT", "development"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		JWTSecret:       getEnv("JWT_SECRET", "your-secret-key-change-this-in-production"),
		JWTExpiry:       time.Duration(jwtExpiryHours) * time.Hour,
		DatabaseURL:     getEnv("DATABASE_URL", ""),
		DatabaseHost:    getEnv("DB_HOST", "localhost"),
		DatabasePort:    getEnv("DB_PORT", "5432"),
		DatabaseName:    getEnv("DB_NAME", "chatapi"),
		DatabaseUser:    getEnv("DB_USER", "postgres"),
		DatabasePass:    getEnv("DB_PASSWORD", "postgres"),
		DatabaseSSLMode: getEnv("DB_SSLMODE", "disable"),
	}
}

// GetDatabaseConnectionString returns the database connection string
func (c *Config) GetDatabaseConnectionString() string {
	// If DATABASE_URL is provided, use it directly (common in cloud deployments)
	if c.DatabaseURL != "" {
		return c.DatabaseURL
	}

	// Otherwise, build connection string from individual components
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DatabaseHost, c.DatabasePort, c.DatabaseUser, c.DatabasePass, c.DatabaseName, c.DatabaseSSLMode)
}

// getEnv gets an environment variable with a fallback default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as integer with a fallback default value
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
