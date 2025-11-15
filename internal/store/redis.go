package store

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/yourusername/reco-engine/internal/util/config"
	"github.com/yourusername/reco-engine/internal/util/logger"
	"go.uber.org/zap"
)

// RedisStore handles Redis operations
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore creates a new Redis store
func NewRedisStore(cfg config.RedisConfig) (*RedisStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		MaxRetries:   cfg.MaxRetries,
		PoolSize:     cfg.PoolSize,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Connected to Redis", zap.String("addr", cfg.Addr))
	return &RedisStore{client: client}, nil
}

// Client returns the Redis client
func (r *RedisStore) Client() *redis.Client {
	return r.client
}

// Close closes the Redis connection
func (r *RedisStore) Close() error {
	return r.client.Close()
}

// AddRecentItem adds an item to user's recent items list
func (r *RedisStore) AddRecentItem(ctx context.Context, userID, itemID int64, limit int) error {
	key := fmt.Sprintf("user:recent:%d", userID)
	pipe := r.client.Pipeline()
	pipe.LPush(ctx, key, itemID)
	pipe.LTrim(ctx, key, 0, int64(limit-1))
	pipe.Expire(ctx, key, 24*time.Hour)
	_, err := pipe.Exec(ctx)
	return err
}

// GetRecentItems gets user's recent items
func (r *RedisStore) GetRecentItems(ctx context.Context, userID int64, count int) ([]string, error) {
	key := fmt.Sprintf("user:recent:%d", userID)
	return r.client.LRange(ctx, key, 0, int64(count-1)).Result()
}

// IncrPopularity increments item popularity score
func (r *RedisStore) IncrPopularity(ctx context.Context, itemID int64, weight float64) error {
	return r.client.ZIncrBy(ctx, "item:popularity", weight, fmt.Sprintf("%d", itemID)).Err()
}

// GetPopularItems gets top popular items
func (r *RedisStore) GetPopularItems(ctx context.Context, count int) ([]redis.Z, error) {
	return r.client.ZRevRangeWithScores(ctx, "item:popularity", 0, int64(count-1)).Result()
}

// IncrCoView increments co-view count between two items
func (r *RedisStore) IncrCoView(ctx context.Context, itemID1, itemID2 int64) error {
	key := fmt.Sprintf("co_view:%d", itemID1)
	pipe := r.client.Pipeline()
	pipe.ZIncrBy(ctx, key, 1, fmt.Sprintf("%d", itemID2))
	pipe.Expire(ctx, key, 7*24*time.Hour)
	_, err := pipe.Exec(ctx)
	return err
}

// GetCoViewItems gets items co-viewed with given item
func (r *RedisStore) GetCoViewItems(ctx context.Context, itemID int64, count int) ([]redis.Z, error) {
	key := fmt.Sprintf("co_view:%d", itemID)
	return r.client.ZRevRangeWithScores(ctx, key, 0, int64(count-1)).Result()
}

// SetItemKNN stores precomputed k-nearest neighbors for an item
func (r *RedisStore) SetItemKNN(ctx context.Context, itemID int64, neighbors []int64) error {
	key := fmt.Sprintf("item:knn:%d", itemID)
	values := make([]interface{}, len(neighbors))
	for i, n := range neighbors {
		values[i] = n
	}

	pipe := r.client.Pipeline()
	pipe.Del(ctx, key)
	if len(values) > 0 {
		pipe.RPush(ctx, key, values...)
	}
	pipe.Expire(ctx, key, 7*24*time.Hour)
	_, err := pipe.Exec(ctx)
	return err
}

// GetItemKNN gets precomputed k-nearest neighbors
func (r *RedisStore) GetItemKNN(ctx context.Context, itemID int64, count int) ([]string, error) {
	key := fmt.Sprintf("item:knn:%d", itemID)
	return r.client.LRange(ctx, key, 0, int64(count-1)).Result()
}

// CacheRecommendations caches recommendations for a user
func (r *RedisStore) CacheRecommendations(ctx context.Context, userID int64, data string, ttl time.Duration) error {
	key := fmt.Sprintf("cache:reco:%d", userID)
	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedRecommendations gets cached recommendations
func (r *RedisStore) GetCachedRecommendations(ctx context.Context, userID int64) (string, error) {
	key := fmt.Sprintf("cache:reco:%d", userID)
	return r.client.Get(ctx, key).Result()
}
