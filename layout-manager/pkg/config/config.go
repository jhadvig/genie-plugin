package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Database    DatabaseConfig
	Server      ServerConfig
	MCP         MCPConfig
	Environment string
}

// DatabaseConfig holds database connection configuration
type DatabaseConfig struct {
	Host     string
	User     string
	Password string
	Name     string
	Port     string
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host string
	Port string
}

// MCPConfig holds MCP server configuration
type MCPConfig struct {
	Host string
	Port string
}

// Load configuration from environment variables with sensible defaults
func Load() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or couldn't be loaded: %v", err)
		log.Println("Falling back to system environment variables")
	} else {
		log.Println("Loaded configuration from .env file")
	}

	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "password"),
			Name:     getEnv("DB_NAME", "layout_manager"),
			Port:     getEnv("DB_PORT", "5433"),
		},
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnv("SERVER_PORT", "9080"),
		},
		MCP: MCPConfig{
			Host: getEnv("MCP_HOST", "0.0.0.0"),
			Port: getEnv("MCP_PORT", "9081"),
		},
		Environment: getEnv("ENV", "development"),
	}
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}