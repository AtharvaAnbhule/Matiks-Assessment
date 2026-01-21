package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"leaderboard-system/models"
)

 
type UserRepository struct {
	db *gorm.DB
}

 
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

 
func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

 
func (r *UserRepository) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

 
func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).
		Where("LOWER(username) = LOWER(?)", username).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return &user, nil
}

 
func (r *UserRepository) UpdateUserRating(ctx context.Context, userID string, newRating int32) error {
	if err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("rating", newRating).Error; err != nil {
		return fmt.Errorf("failed to update rating: %w", err)
	}
	return nil
}

 
func (r *UserRepository) GetLeaderboard(ctx context.Context, offset, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

 
	if err := r.db.WithContext(ctx).Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

 
	if err := r.db.WithContext(ctx).
		Order("rating DESC, username ASC").
		Offset(offset).
		Limit(limit).
		Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get leaderboard: %w", err)
	}

	return users, total, nil
}

 
func (r *UserRepository) CalculateRank(ctx context.Context, userID string) (int64, error) {
	var rank int64
 
	var targetRating int32
	if err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Select("rating").
		Scan(&targetRating).Error; err != nil {
		return 0, fmt.Errorf("failed to get user rating: %w", err)
	}

 
	if err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("rating > ?", targetRating).
		Count(&rank).Error; err != nil {
		return 0, fmt.Errorf("failed to calculate rank: %w", err)
	}

	return rank + 1, nil
}
 
func (r *UserRepository) GetUsersByRating(ctx context.Context, rating int32) ([]models.User, error) {
	var users []models.User
	if err := r.db.WithContext(ctx).
		Where("rating = ?", rating).
		Order("username ASC").
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get users by rating: %w", err)
	}
	return users, nil
}

 
func (r *UserRepository) SearchUserByUsername(ctx context.Context, username string, limit int) ([]models.User, error) {
	var users []models.User
	if err := r.db.WithContext(ctx).
		Where("LOWER(username) LIKE LOWER(?)", username+"%").
		Limit(limit).
		Order("username ASC").
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}
	return users, nil
}

 
func (r *UserRepository) GetAllUsers(ctx context.Context) ([]models.User, error) {
	var users []models.User
	if err := r.db.WithContext(ctx).
		Order("rating DESC, username ASC").
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	return users, nil
}

 
func (r *UserRepository) DeleteUser(ctx context.Context, userID string) error {
	if err := r.db.WithContext(ctx).Delete(&models.User{}, "id = ?", userID).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

 
func (r *UserRepository) BulkCreateUsers(ctx context.Context, users []models.User) error {
	if err := r.db.WithContext(ctx).CreateInBatches(users, 100).Error; err != nil {
		return fmt.Errorf("failed to bulk create users: %w", err)
	}
	return nil
}

 
func (r *UserRepository) GetUserCount(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}
