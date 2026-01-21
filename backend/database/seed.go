package database

import (
	"context"
	"fmt"
	"time"

	"leaderboard-system/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SeedData seeds the database with initial users for development/testing
func SeedData(db *gorm.DB) error {
	ctx := context.Background()
	numUsers := 500
	users := make([]models.User, 0, numUsers)
	now := time.Now()

	for i := 1; i <= numUsers; i++ {
		user := models.User{
			ID:        uuid.NewString(),
			Username:  fmt.Sprintf("user%03d", i),
			Rating:    int32(100 + (i*37)%4901), // pseudo-random rating between 100-5000
			CreatedAt: now,
			UpdatedAt: now,
		}
		users = append(users, user)
	}

	for _, user := range users {
		var existing models.User
		err := db.WithContext(ctx).Where("username = ?", user.Username).First(&existing).Error
		if err == gorm.ErrRecordNotFound {
			if err := db.WithContext(ctx).Create(&user).Error; err != nil {
				return fmt.Errorf("failed to seed user %s: %w", user.Username, err)
			}
		}
	}
	return nil
}
