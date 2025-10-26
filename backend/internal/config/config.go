package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	JWT      JWTConfig
	Database DatabaseConfig
	Security SecurityConfig
}

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port string
	Env  string
}

// JWTConfig holds JWT-specific configuration
type JWTConfig struct {
	Secret          string
	ExpirationHours int
}

// DatabaseConfig holds database-specific configuration
type DatabaseConfig struct {
	Path string
}

// SecurityConfig holds security-specific configuration
type SecurityConfig struct {
	RateLimitRequests int
	RateLimitDuration int
	AllowedOrigins    string
}

// LoadConfig loads the application configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	// Server configuration
	port := getEnvOrDefault("PORT", "8080")
	env := getEnvOrDefault("ENV", "development")

	// JWT configuration
	jwtSecret := getEnvOrDefault("JWT_SECRET", "your-super-secret-key-change-this-in-production")
	jwtExpHours, _ := strconv.Atoi(getEnvOrDefault("JWT_EXPIRATION_HOURS", "24"))

	// Database configuration
	dbPath := getEnvOrDefault("DB_PATH", filepath.Join(".", "data", "reminiscer.db"))

	// Security configuration
	rateLimitReq, _ := strconv.Atoi(getEnvOrDefault("RATE_LIMIT_REQUESTS", "100"))
	rateLimitDur, _ := strconv.Atoi(getEnvOrDefault("RATE_LIMIT_DURATION", "60"))
	allowedOrigins := getEnvOrDefault("ALLOWED_ORIGINS", "http://localhost:3000")

	// Ensure database directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	return &Config{
		Server: ServerConfig{
			Port: port,
			Env:  env,
		},
		JWT: JWTConfig{
			Secret:          jwtSecret,
			ExpirationHours: jwtExpHours,
		},
		Database: DatabaseConfig{
			Path: dbPath,
		},
		Security: SecurityConfig{
			RateLimitRequests: rateLimitReq,
			RateLimitDuration: rateLimitDur,
			AllowedOrigins:    allowedOrigins,
		},
	}, nil
}

// IsDevelopment returns true if the application is running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Server.Env == "development"
}

// IsProduction returns true if the application is running in production mode
func (c *Config) IsProduction() bool {
	return c.Server.Env == "production"
}

// getEnvOrDefault returns the environment variable value or the default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
