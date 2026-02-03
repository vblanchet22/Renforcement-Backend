package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	JWT      JWTConfig
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// ServerConfig holds server settings
type ServerConfig struct {
	GRPCPort string
	HTTPPort string
}

// JWTConfig holds JWT settings
type JWTConfig struct {
	Secret             string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "coloc_user"),
			Password: getEnv("DB_PASSWORD", "coloc_password"),
			Name:     getEnv("DB_NAME", "coloc_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Server: ServerConfig{
			GRPCPort: getEnv("GRPC_PORT", "50051"),
			HTTPPort: getEnv("HTTP_PORT", "8080"),
		},
		JWT: JWTConfig{
			Secret:             getEnv("JWT_SECRET", "change-me-in-production"),
			AccessTokenExpiry:  getDurationEnv("JWT_EXPIRY", 24*time.Hour),
			RefreshTokenExpiry: getDurationEnv("REFRESH_TOKEN_EXPIRY", 168*time.Hour),
		},
	}
}

// DatabaseURL returns the PostgreSQL connection URL
func (c *DatabaseConfig) DatabaseURL() string {
	return "postgres://" + c.User + ":" + c.Password + "@" + c.Host + ":" + c.Port + "/" + c.Name + "?sslmode=" + c.SSLMode
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if hours, err := strconv.Atoi(value); err == nil {
			return time.Duration(hours) * time.Hour
		}
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
	}
	return defaultValue
}
