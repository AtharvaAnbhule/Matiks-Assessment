package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"leaderboard-system/config"
	"leaderboard-system/models"
)

const (
	// TTL for cached data
	CacheUserTTL         = 5 * time.Minute
	CacheLeaderboardTTL  = 2 * time.Minute
	CacheRankTTL         = 3 * time.Minute
	
	// Cache keys
	UserCacheKeyPrefix   = "user:"
	RankCacheKeyPrefix   = "rank:"
	LeaderboardCacheKey  = "leaderboard"
)

// CacheManager handles all caching operations
// Uses Redis sorted set for efficient ranking calculations
// Pattern: Cache-aside (lazy loading) for user data
// Pattern: TTL-based invalidation for leaderboard
type CacheManager struct {
	client *redis.Client
}

// NewCacheManager creates a new cache manager instance
func NewCacheManager(cfg *config.RedisConfig) (*CacheManager, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &CacheManager{client: client}, nil
}

// SetUser caches user data with TTL
// Uses hash structure for efficient storage
func (cm *CacheManager) SetUser(ctx context.Context, user *models.User) error {
	key := fmt.Sprintf("%s%s", UserCacheKeyPrefix, user.ID)
	
	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	return cm.client.Set(ctx, key, data, CacheUserTTL).Err()
}

// GetUser retrieves user from cache
// Returns nil if not found or expired
func (cm *CacheManager) GetUser(ctx context.Context, userID string) (*models.User, error) {
	key := fmt.Sprintf("%s%s", UserCacheKeyPrefix, userID)
	
	val, err := cm.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := json.Unmarshal([]byte(val), &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// InvalidateUser removes user from cache
func (cm *CacheManager) InvalidateUser(ctx context.Context, userID string) error {
	key := fmt.Sprintf("%s%s", UserCacheKeyPrefix, userID)
	return cm.client.Del(ctx, key).Err()
}

// SetRank caches the rank of a user
// Rank is calculated once and cached to avoid repeated DB queries
func (cm *CacheManager) SetRank(ctx context.Context, userID string, rank int64) error {
	key := fmt.Sprintf("%s%s", RankCacheKeyPrefix, userID)
	return cm.client.Set(ctx, key, rank, CacheRankTTL).Err()
}

// GetRank retrieves cached rank
func (cm *CacheManager) GetRank(ctx context.Context, userID string) (int64, error) {
	key := fmt.Sprintf("%s%s", RankCacheKeyPrefix, userID)
	
	val, err := cm.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	var rank int64
	if _, err := fmt.Sscanf(val, "%d", &rank); err != nil {
		return 0, err
	}

	return rank, nil
}

// InvalidateRank removes rank from cache
func (cm *CacheManager) InvalidateRank(ctx context.Context, userID string) error {
	key := fmt.Sprintf("%s%s", RankCacheKeyPrefix, userID)
	return cm.client.Del(ctx, key).Err()
}

// InvalidateLeaderboard clears leaderboard cache
// Called when rankings change
func (cm *CacheManager) InvalidateLeaderboard(ctx context.Context) error {
	return cm.client.Del(ctx, LeaderboardCacheKey).Err()
}

// Close closes the Redis connection
func (cm *CacheManager) Close() error {
	return cm.client.Close()
}

// Flush clears all cache keys (use with caution)
func (cm *CacheManager) Flush(ctx context.Context) error {
	return cm.client.FlushDB(ctx).Err()
}
