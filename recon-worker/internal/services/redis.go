package services

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/umerthow/reconcile/recon-worker/internal/utils"
)

type RedisService struct {
	client *redis.Client
}

func NewRedisService(addr string) (*RedisService, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     "", // No password for local dev
		DB:           0,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	utils.Info().Msg("Redis connection established")

	return &RedisService{client: client}, nil
}

func (s *RedisService) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

// CheckDeduplication checks if webhook has been processed recently
func (s *RedisService) CheckDeduplication(ctx context.Context, provider, webhookID string) (bool, error) {
	key := fmt.Sprintf("webhook:%s:%s", provider, webhookID)

	exists, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check deduplication: %w", err)
	}

	return exists > 0, nil
}

// SetDeduplication marks webhook as processed with TTL
func (s *RedisService) SetDeduplication(ctx context.Context, provider, webhookID string, ttl time.Duration) error {
	key := fmt.Sprintf("webhook:%s:%s", provider, webhookID)

	err := s.client.Set(ctx, key, "processed", ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set deduplication: %w", err)
	}

	return nil
}

// Health check for Redis connection
func (s *RedisService) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}
