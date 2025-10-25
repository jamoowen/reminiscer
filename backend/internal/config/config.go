package config

import (
	"os"
	"time"
)

// Config holds all configuration settings
type Config struct {
	JWT JWTConfig
}

// JWTConfig holds JWT-specific configuration
type JWTConfig struct {
	Secret          string
	ExpirationHours int
}

// DefaultJWTConfig returns the default JWT configuration
func DefaultJWTConfig() JWTConfig {
	return JWTConfig{
		Secret:          getEnvOrDefault("JWT_SECRET", "your-default-secret-key-change-in-production"),
		ExpirationHours: 24, // 24 hours by default
	}
}

// GetConfig returns the application configuration
func GetConfig() *Config {
	return &Config{
		JWT: DefaultJWTConfig(),
	}
}

// TokenExpiration returns the token expiration duration
func (c *JWTConfig) TokenExpiration() time.Duration {
	return time.Duration(c.ExpirationHours) * time.Hour
}

// getEnvOrDefault returns the environment variable value or the default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
