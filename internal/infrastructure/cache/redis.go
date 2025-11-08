package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/eneskaya/insider-messaging/pkg/config"
	"github.com/eneskaya/insider-messaging/pkg/logger"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCache(cfg *config.RedisConfig) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Get().Info("connected to Redis cache",
		zap.String("address", cfg.Address()),
		zap.Int("db", cfg.DB),
	)

	return &RedisCache{
		client: client,
		ttl:    cfg.CacheTTL,
	}, nil
}

func (r *RedisCache) Close() error {
	if r.client != nil {
		logger.Get().Info("closing Redis connection")
		return r.client.Close()
	}
	return nil
}

func (r *RedisCache) HealthCheck(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *RedisCache) Set(ctx context.Context, key string, value interface{}) error {
	return r.client.Set(ctx, key, value, r.ttl).Err()
}

func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}
