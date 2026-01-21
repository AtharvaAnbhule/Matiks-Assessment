package service

import (
	"context"
	"errors"
	"fmt"
	"sync"


	"go.uber.org/zap"
	"leaderboard-system/cache"
	"leaderboard-system/models"
	"leaderboard-system/repository"
)

// UserService provides business logic for user operations
// Handles:
// - Cache management (reduces DB load)
// - Rank calculations (with tie-awareness)
// - Input validation
// - Concurrent update safety via goroutine-per-request pattern
// - Non-blocking operations using channels
type UserService struct {
	repo     *repository.UserRepository
	cache    *cache.CacheManager
	logger   *zap.Logger
	mu       sync.RWMutex // Protects concurrent rank updates
	rankMu   map[string]*sync.Mutex // Per-user rank calculation lock
}

// NewUserService creates a new user service instance
func NewUserService(repo *repository.UserRepository, cache *cache.CacheManager, logger *zap.Logger) *UserService {
	return &UserService{
		repo:   repo,
		cache:  cache,
		logger: logger,
		rankMu: make(map[string]*sync.Mutex),
	}
}

// getRankMutex gets or creates a mutex for a specific user
// Ensures thread-safe rank calculations for concurrent updates
func (s *UserService) getRankMutex(userID string) *sync.Mutex {
	s.mu.Lock()
	defer s.mu.Unlock()

	if mu, exists := s.rankMu[userID]; exists {
		return mu
	}

	mu := &sync.Mutex{}
	s.rankMu[userID] = mu
	return mu
}

// CreateUser creates a new user with validation
// Returns error if validation fails or user exists
func (s *UserService) CreateUser(ctx context.Context, userID, username string, initialRating int32) (*models.User, error) {
	// Validate input
	if err := ValidateUsername(username); err != nil {
		s.logger.Warn("Invalid username", zap.String("username", username), zap.Error(err))
		return nil, fmt.Errorf("invalid username: %w", err)
	}

	if err := ValidateRating(initialRating); err != nil {
		s.logger.Warn("Invalid rating", zap.Int32("rating", initialRating), zap.Error(err))
		return nil, fmt.Errorf("invalid rating: %w", err)
	}

	// Check if user already exists
	existingUser, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		s.logger.Error("Failed to check user existence", zap.Error(err))
		return nil, err
	}

	if existingUser != nil {
		return nil, errors.New("user already exists")
	}

	// Create user
	user := &models.User{
		ID:       userID,
		Username: username,
		Rating:   initialRating,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, err
	}

	// Cache the new user
	if err := s.cache.SetUser(ctx, user); err != nil {
		s.logger.Warn("Failed to cache user", zap.Error(err))
		// Not critical, continue
	}

	s.logger.Info("User created", zap.String("user_id", userID), zap.String("username", username))
	return user, nil
}

// GetUserByID retrieves user with rank information
// Uses cache-aside pattern:
// 1. Check cache
// 2. If miss, fetch from DB and cache
func (s *UserService) GetUserByID(ctx context.Context, userID string) (*models.UserDTO, int64, error) {
	// Try cache first
	user, err := s.cache.GetUser(ctx, userID)
	if err != nil {
		s.logger.Warn("Cache error", zap.Error(err))
	}

	// Cache miss or error, fetch from DB
	if user == nil {
		dbUser, err := s.repo.GetUserByID(ctx, userID)
		if err != nil {
			return nil, 0, err
		}
		if dbUser == nil {
			return nil, 0, errors.New("user not found")
		}
		user = dbUser

		// Cache the user (fire and forget)
		go func() {
			if err := s.cache.SetUser(context.Background(), user); err != nil {
				s.logger.Warn("Failed to cache user", zap.Error(err))
			}
		}()
	}

	// Calculate rank
	rank, err := s.GetUserRank(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to calculate rank", zap.Error(err))
		return nil, 0, err
	}

	return &models.UserDTO{
		ID:       user.ID,
		Username: user.Username,
		Rating:   user.Rating,
	}, rank, nil
}

// GetUserRank calculates user's rank with caching
// Implements tie-aware ranking:
// Users with same rating have same rank
// Uses sorted set logic (COUNT WHERE rating > user_rating + 1)
func (s *UserService) GetUserRank(ctx context.Context, userID string) (int64, error) {
	// Try cache first
	cachedRank, err := s.cache.GetRank(ctx, userID)
	if err != nil {
		s.logger.Warn("Cache error for rank", zap.Error(err))
	}

	if cachedRank > 0 {
		return cachedRank, nil
	}

	// Acquire per-user lock to prevent concurrent rank calculations
	rankMu := s.getRankMutex(userID)
	rankMu.Lock()
	defer rankMu.Unlock()

	// Double-check cache after acquiring lock
	cachedRank, _ = s.cache.GetRank(ctx, userID)
	if cachedRank > 0 {
		return cachedRank, nil
	}

	// Calculate rank from database
	rank, err := s.repo.CalculateRank(ctx, userID)
	if err != nil {
		return 0, err
	}

	// Cache the rank (fire and forget)
	go func() {
		if err := s.cache.SetRank(context.Background(), userID, rank); err != nil {
			s.logger.Warn("Failed to cache rank", zap.Error(err))
		}
	}()

	return rank, nil
}

// UpdateUserRating updates user's rating and invalidates rank cache
// Non-blocking: cache invalidation happens asynchronously
// This ensures API response is fast
func (s *UserService) UpdateUserRating(ctx context.Context, userID string, newRating int32) (*models.UserDTO, int64, error) {
	// Validate rating
	if err := ValidateRating(newRating); err != nil {
		return nil, 0, fmt.Errorf("invalid rating: %w", err)
	}

	// Get current user
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	if user == nil {
		return nil, 0, errors.New("user not found")
	}

	// Update rating
	if err := s.repo.UpdateUserRating(ctx, userID, newRating); err != nil {
		return nil, 0, err
	}

	// Update user object
	user.Rating = newRating

	// Invalidate caches asynchronously (fire and forget)
	// This prevents blocking the API response
	go func() {
		ctx := context.Background()
		if err := s.cache.InvalidateUser(ctx, userID); err != nil {
			s.logger.Warn("Failed to invalidate user cache", zap.Error(err))
		}
		if err := s.cache.InvalidateRank(ctx, userID); err != nil {
			s.logger.Warn("Failed to invalidate rank cache", zap.Error(err))
		}
		if err := s.cache.InvalidateLeaderboard(ctx); err != nil {
			s.logger.Warn("Failed to invalidate leaderboard cache", zap.Error(err))
		}
	}()

	// Calculate new rank
	rank, err := s.GetUserRank(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to calculate new rank", zap.Error(err))
		// Still return user, but with error logged
	}

	s.logger.Info("User rating updated",
		zap.String("user_id", userID),
		zap.Int32("old_rating", user.Rating),
		zap.Int32("new_rating", newRating),
	)

	return &models.UserDTO{
		ID:       user.ID,
		Username: user.Username,
		Rating:   newRating,
	}, rank, nil
}

// SearchUserByUsername searches for user by username
// Returns user with rank if found
// Implements case-insensitive search for better UX
func (s *UserService) SearchUserByUsername(ctx context.Context, username string) (*models.UserDTO, int64, error) {
	// Validate input
	if err := ValidateUsername(username); err != nil {
		return nil, 0, fmt.Errorf("invalid username: %w", err)
	}

	// Search database (uses indexed column)
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, 0, err
	}

	if user == nil {
		return nil, 0, nil
	}

	// Get rank
	rank, err := s.GetUserRank(ctx, user.ID)
	if err != nil {
		s.logger.Error("Failed to get rank", zap.Error(err))
		return nil, 0, err
	}

	s.logger.Info("User search",
		zap.String("search_term", username),
		zap.String("found_user", user.Username),
	)

	return &models.UserDTO{
		ID:       user.ID,
		Username: user.Username,
		Rating:   user.Rating,
	}, rank, nil
}

// GetLeaderboard retrieves paginated leaderboard with ranks
// Non-blocking pagination using offset-limit
// For 100M+ users, consider keyset pagination
func (s *UserService) GetLeaderboard(ctx context.Context, page, pageSize int) (*models.LeaderboardResponse, error) {
	// Validate pagination params
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100 // Default page size
	}

	offset := (page - 1) * pageSize

	// Fetch from database
	users, total, err := s.repo.GetLeaderboard(ctx, offset, pageSize)
	if err != nil {
		return nil, err
	}

	// Convert to leaderboard entries with ranks
	entries := make([]models.LeaderboardEntry, 0, len(users))
	var currentRank int64 = 1
	var previousRating int32 = -1

	for i, user := range users {
		// Implement tie-aware ranking
		// When rating changes, rank increments by count of users at previous rating
		if user.Rating != previousRating {
			currentRank = int64(offset + i + 1)
			previousRating = user.Rating
		}

		entries = append(entries, models.LeaderboardEntry{
			Rank:     currentRank,
			Username: user.Username,
			Rating:   user.Rating,
		})
	}

	hasMore := offset+int(int64(pageSize)) < int(total)

	s.logger.Info("Leaderboard fetched",
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.Int64("total", total),
	)

	return &models.LeaderboardResponse{
		Entries:  entries,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		HasMore:  hasMore,
	}, nil
}

// GetLeaderboardAroundUser gets leaderboard with user's position
// Shows ranking context: users before and after the target user
func (s *UserService) GetLeaderboardAroundUser(ctx context.Context, userID string, contextSize int) (*models.LeaderboardResponse, error) {
	// Get user's rank
	rank, err := s.GetUserRank(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Calculate page: user's rank determines page number
	page := 1
	pageSize := contextSize * 2
	if rank > int64(contextSize) {
		page = int((rank - int64(contextSize)) / int64(pageSize))
		if page < 1 {
			page = 1
		}
	}

	return s.GetLeaderboard(ctx, page, pageSize)
}

// IsHealthy checks service health
func (s *UserService) IsHealthy(ctx context.Context) bool {
	// Check cache connection
	_, err := s.cache.GetUser(ctx, "health-check")
	return err == nil
}
