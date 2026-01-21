package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"leaderboard-system/config"
	"leaderboard-system/models"
)

// InitDB initializes and returns a PostgreSQL database connection
// Creates necessary indexes for optimal query performance
// Handles migrations and schema creation
func InitDB(cfg *config.DatabaseConfig, logLevel logger.LogLevel) (*gorm.DB, error) {
	dsn := cfg.GetDSN()
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run migrations
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create indexes for optimal performance
	if err := createIndexes(db); err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	return db, nil
}

// runMigrations creates the necessary database tables
func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(&models.User{})
}

// createIndexes creates database indexes for frequently queried columns
// This ensures O(log n) complexity for searches and ranking queries
// Indexes:
// 1. username - for fast user search (UNIQUE to prevent duplicates)
// 2. rating - for ranking queries and range scans
// 3. (rating, username) - composite for efficient rank calculation
func createIndexes(db *gorm.DB) error {
	// Composite index on rating DESC and username for rank calculation
	// This allows efficient range queries without sorting in application
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_users_rating_username 
		ON users(rating DESC, username)
	`).Error; err != nil {
		return err
	}

	// Index for search queries with prefix matching
	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_users_username_lower 
		ON users(LOWER(username))
	`).Error; err != nil {
		return err
	}

	return nil
}

// GetDB returns the database connection
// Useful for dependency injection
func GetDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	return InitDB(cfg, logger.Silent)
}
