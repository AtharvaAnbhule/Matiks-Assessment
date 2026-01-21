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

	CacheUserTTL         = 5 * time.Minute
	CacheLeaderboardTTL  = 2 * time.Minute
	CacheRankTTL         = 3 * time.Minute
	
	
	UserCacheKeyPrefix   = "user:"
	RankCacheKeyPrefix   = "rank:"
	LeaderboardCacheKey  = "leaderboard"
)


type CacheManager struct {
	client *redis.Client
}


func NewCacheManager(cfg *config.RedisConfig) (*CacheManager, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})


	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &CacheManager{client: client}, nil
}


func (cm *CacheManager) SetUser(ctx context.Context, user *models.User) error {
	key := fmt.Sprintf("%s%s", UserCacheKeyPrefix, user.ID)
	
	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	return cm.client.Set(ctx, key, data, CacheUserTTL).Err()
}


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


func (cm *CacheManager) InvalidateUser(ctx context.Context, userID string) error {
	key := fmt.Sprintf("%s%s", UserCacheKeyPrefix, userID)
	return cm.client.Del(ctx, key).Err()
}


func (cm *CacheManager) SetRank(ctx context.Context, userID string, rank int64) error {
	key := fmt.Sprintf("%s%s", RankCacheKeyPrefix, userID)
	return cm.client.Set(ctx, key, rank, CacheRankTTL).Err()
}


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


func (cm *CacheManager) InvalidateRank(ctx context.Context, userID string) error {
	key := fmt.Sprintf("%s%s", RankCacheKeyPrefix, userID)
	return cm.client.Del(ctx, key).Err()
}


func (cm *CacheManager) InvalidateLeaderboard(ctx context.Context) error {
	return cm.client.Del(ctx, LeaderboardCacheKey).Err()
}


func (cm *CacheManager) Close() error {
	return cm.client.Close()
}


func (cm *CacheManager) Flush(ctx context.Context) error {
	return cm.client.FlushDB(ctx).Err()
}
