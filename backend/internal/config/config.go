package config

import (
	"fmt"
	"os"
)

// Config holds all application configuration
type Config struct {
	ServerPort    string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	NatsURL       string
	AIServiceURL  string
	MaxInstances  int
	LogLevel      string
}

// Load reads configuration from environment with defaults
func Load() *Config {
	return &Config{
		ServerPort:    envOrDefault("SERVER_PORT", "8080"),
		DBHost:        envOrDefault("DB_HOST", "localhost"),
		DBPort:        envOrDefault("DB_PORT", "5432"),
		DBUser:        envOrDefault("DB_USER", "defense"),
		DBPassword:    envOrDefault("DB_PASSWORD", "defense123"),
		DBName:        envOrDefault("DB_NAME", "defense_agent"),
		NatsURL:       envOrDefault("NATS_URL", "nats://localhost:4222"),
		AIServiceURL:  envOrDefault("AI_SERVICE_URL", "http://localhost:8100"),
		MaxInstances:  10,
		LogLevel:      envOrDefault("LOG_LEVEL", "info"),
	}
}

// DSN returns the PostgreSQL connection string
func (c *Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
