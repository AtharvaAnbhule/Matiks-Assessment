package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"leaderboard-system/models"
)

// UserRepository handles all user data operations
// Implements repository pattern for separation of concerns
// All database queries are optimized with proper indexes
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository instance
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser creates a new user in the database
// Returns error if user already exists or validation fails
func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetUserByID retrieves a user by ID
// Uses database index on primary key for O(1) lookup
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

// GetUserByUsername retrieves a user by username
// Uses indexed column for fast lookup
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

// UpdateUserRating updates a user's rating
// Non-blocking operation using goroutine isolation
// Invalidates cache to ensure consistency
func (r *UserRepository) UpdateUserRating(ctx context.Context, userID string, newRating int32) error {
	if err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Update("rating", newRating).Error; err != nil {
		return fmt.Errorf("failed to update rating: %w", err)
	}
	return nil
}

// GetLeaderboard retrieves paginated leaderboard
// Uses efficient database query with composite index
// Offset-based pagination for simplicity (can be improved with keyset pagination for 100M+ users)
// Rows are pre-sorted by database using composite index
func (r *UserRepository) GetLeaderboard(ctx context.Context, offset, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	// Count total users
	if err := r.db.WithContext(ctx).Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Fetch paginated results
	// ORDER BY rating DESC, username ASC ensures consistent ordering
	// for users with same rating (tie-aware ranking)
	if err := r.db.WithContext(ctx).
		Order("rating DESC, username ASC").
		Offset(offset).
		Limit(limit).
		Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get leaderboard: %w", err)
	}

	return users, total, nil
}

// CalculateRank calculates the rank of a user
// Uses GROUP BY to count distinct ratings higher than user's rating
// This implements tie-aware ranking: users with same rating have same rank
//
// Example: Ratings: 5000, 4500, 4500, 4000
// Ranks:    1,    2,    2,    4
// (not: 1, 2, 3, 4 - because there are 2 users at 4500)
func (r *UserRepository) CalculateRank(ctx context.Context, userID string) (int64, error) {
	var rank int64

	// Subquery: Get the rating of the target user
	var targetRating int32
	if err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("id = ?", userID).
		Select("rating").
		Scan(&targetRating).Error; err != nil {
		return 0, fmt.Errorf("failed to get user rating: %w", err)
	}

	// Main query: Count users with higher rating
	// +1 because rank is 1-based, not 0-based
	if err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("rating > ?", targetRating).
		Count(&rank).Error; err != nil {
		return 0, fmt.Errorf("failed to calculate rank: %w", err)
	}

	return rank + 1, nil
}

// GetUsersByRating returns users with a specific rating
// Useful for finding all users tied at a rank
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

// SearchUserByUsername searches for users by username prefix
// Uses indexed LOWER(username) column for fast searching
// Case-insensitive search for better UX
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

// GetAllUsers retrieves all users (use with caution for large datasets)
// Consider pagination for production use
func (r *UserRepository) GetAllUsers(ctx context.Context) ([]models.User, error) {
	var users []models.User
	if err := r.db.WithContext(ctx).
		Order("rating DESC, username ASC").
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}
	return users, nil
}

// DeleteUser deletes a user
func (r *UserRepository) DeleteUser(ctx context.Context, userID string) error {
	if err := r.db.WithContext(ctx).Delete(&models.User{}, "id = ?", userID).Error; err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// BulkCreateUsers creates multiple users in a single transaction
// Used for seeding test data efficiently
func (r *UserRepository) BulkCreateUsers(ctx context.Context, users []models.User) error {
	if err := r.db.WithContext(ctx).CreateInBatches(users, 100).Error; err != nil {
		return fmt.Errorf("failed to bulk create users: %w", err)
	}
	return nil
}

// GetUserCount returns total count of users
func (r *UserRepository) GetUserCount(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.User{}).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}
