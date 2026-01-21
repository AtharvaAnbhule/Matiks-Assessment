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

 
type UserService struct {
	repo     *repository.UserRepository
	cache    *cache.CacheManager
	logger   *zap.Logger
	mu       sync.RWMutex 
	rankMu   map[string]*sync.Mutex 
}


func NewUserService(repo *repository.UserRepository, cache *cache.CacheManager, logger *zap.Logger) *UserService {
	return &UserService{
		repo:   repo,
		cache:  cache,
		logger: logger,
		rankMu: make(map[string]*sync.Mutex),
	}
}


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


func (s *UserService) CreateUser(ctx context.Context, userID, username string, initialRating int32) (*models.User, error) {

	if err := ValidateUsername(username); err != nil {
		s.logger.Warn("Invalid username", zap.String("username", username), zap.Error(err))
		return nil, fmt.Errorf("invalid username: %w", err)
	}

	if err := ValidateRating(initialRating); err != nil {
		s.logger.Warn("Invalid rating", zap.Int32("rating", initialRating), zap.Error(err))
		return nil, fmt.Errorf("invalid rating: %w", err)
	}


	existingUser, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		s.logger.Error("Failed to check user existence", zap.Error(err))
		return nil, err
	}

	if existingUser != nil {
		return nil, errors.New("user already exists")
	}


	user := &models.User{
		ID:       userID,
		Username: username,
		Rating:   initialRating,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, err
	}


	if err := s.cache.SetUser(ctx, user); err != nil {
		s.logger.Warn("Failed to cache user", zap.Error(err))
		
	}

	s.logger.Info("User created", zap.String("user_id", userID), zap.String("username", username))
	return user, nil
}


func (s *UserService) GetUserByID(ctx context.Context, userID string) (*models.UserDTO, int64, error) {

	user, err := s.cache.GetUser(ctx, userID)
	if err != nil {
		s.logger.Warn("Cache error", zap.Error(err))
	}


	if user == nil {
		dbUser, err := s.repo.GetUserByID(ctx, userID)
		if err != nil {
			return nil, 0, err
		}
		if dbUser == nil {
			return nil, 0, errors.New("user not found")
		}
		user = dbUser

		
		go func() {
			if err := s.cache.SetUser(context.Background(), user); err != nil {
				s.logger.Warn("Failed to cache user", zap.Error(err))
			}
		}()
	}

	
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


func (s *UserService) GetUserRank(ctx context.Context, userID string) (int64, error) {

	cachedRank, err := s.cache.GetRank(ctx, userID)
	if err != nil {
		s.logger.Warn("Cache error for rank", zap.Error(err))
	}

	if cachedRank > 0 {
		return cachedRank, nil
	}

	 
	rankMu := s.getRankMutex(userID)
	rankMu.Lock()
	defer rankMu.Unlock()

 
	cachedRank, _ = s.cache.GetRank(ctx, userID)
	if cachedRank > 0 {
		return cachedRank, nil
	}

 
	rank, err := s.repo.CalculateRank(ctx, userID)
	if err != nil {
		return 0, err
	}

 
	go func() {
		if err := s.cache.SetRank(context.Background(), userID, rank); err != nil {
			s.logger.Warn("Failed to cache rank", zap.Error(err))
		}
	}()

	return rank, nil
}

 
func (s *UserService) UpdateUserRating(ctx context.Context, userID string, newRating int32) (*models.UserDTO, int64, error) {
	 
	if err := ValidateRating(newRating); err != nil {
		return nil, 0, fmt.Errorf("invalid rating: %w", err)
	}

 
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, 0, err
	}

	if user == nil {
		return nil, 0, errors.New("user not found")
	}

 
	if err := s.repo.UpdateUserRating(ctx, userID, newRating); err != nil {
		return nil, 0, err
	}

 
	user.Rating = newRating

 
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

	 
	rank, err := s.GetUserRank(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to calculate new rank", zap.Error(err))
		 
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

 
func (s *UserService) SearchUserByUsername(ctx context.Context, username string) (*models.UserDTO, int64, error) {
 
	if err := ValidateUsername(username); err != nil {
		return nil, 0, fmt.Errorf("invalid username: %w", err)
	}

	 
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, 0, err
	}

	if user == nil {
		return nil, 0, nil
	}

 
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


func (s *UserService) GetLeaderboard(ctx context.Context, page, pageSize int) (*models.LeaderboardResponse, error) {

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 1000 {
		pageSize = 100 
	}

	offset := (page - 1) * pageSize

	
	users, total, err := s.repo.GetLeaderboard(ctx, offset, pageSize)
	if err != nil {
		return nil, err
	}

	
	entries := make([]models.LeaderboardEntry, 0, len(users))
	var currentRank int64 = 1
	var previousRating int32 = -1

	for i, user := range users {
		
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


func (s *UserService) GetLeaderboardAroundUser(ctx context.Context, userID string, contextSize int) (*models.LeaderboardResponse, error) {
	
	rank, err := s.GetUserRank(ctx, userID)
	if err != nil {
		return nil, err
	}

	
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


func (s *UserService) IsHealthy(ctx context.Context) bool {
	
	_, err := s.cache.GetUser(ctx, "health-check")
	return err == nil
}
