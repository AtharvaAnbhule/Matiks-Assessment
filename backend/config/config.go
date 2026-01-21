package config

import (
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
)

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	// Neon database URL (takes precedence over individual params)
	DatabaseURL string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type ServerConfig struct {
	Port string
	Env  string
}

type Config struct {
	Database DatabaseConfig
	Redis    RedisConfig
	Server   ServerConfig
}

var (
	once     sync.Once
	instance *Config
)

// GetConfig returns singleton config instance
// This follows singleton pattern for thread-safe config access
func GetConfig() *Config {
	once.Do(func() {
		instance = loadConfig()
	})
	return instance
}

// LoadConfig loads configuration from environment variables
// Supports both Neon DATABASE_URL and individual connection parameters
// DATABASE_URL (Neon format) takes precedence if provided
// This allows 12-factor app compliance for cloud deployment
func loadConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			// Neon connection string (preferred)
			DatabaseURL: getEnv("DATABASE_URL", ""),
			// Fallback to individual parameters for local PostgreSQL
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "leaderboard"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		},
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Env:  getEnv("ENV", "development"),
		},
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

// GetDSN returns PostgreSQL connection string
// Uses DATABASE_URL if provided (Neon), otherwise constructs from individual params
func (c *DatabaseConfig) GetDSN() string {
	// If DATABASE_URL is provided (Neon), use it directly
	if c.DatabaseURL != "" {
		return c.DatabaseURL
	}
	// Otherwise construct DSN from individual parameters
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

// GetLogger returns configured logger instance
func GetLogger(env string) (*zap.Logger, error) {
	if env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}
