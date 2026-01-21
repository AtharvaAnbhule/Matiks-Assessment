package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"leaderboard-system/config"
	"leaderboard-system/models"
)


func InitDB(cfg *config.DatabaseConfig, logLevel logger.LogLevel) (*gorm.DB, error) {
	dsn := cfg.GetDSN()
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}


	if err := createIndexes(db); err != nil {
		return nil, fmt.Errorf("failed to create indexes: %w", err)
	}

	return db, nil
}


func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(&models.User{})
}


func createIndexes(db *gorm.DB) error {

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_users_rating_username 
		ON users(rating DESC, username)
	`).Error; err != nil {
		return err
	}

	if err := db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_users_username_lower 
		ON users(LOWER(username))
	`).Error; err != nil {
		return err
	}

	return nil
}


func GetDB(cfg *config.DatabaseConfig) (*gorm.DB, error) {
	return InitDB(cfg, logger.Silent)
}
