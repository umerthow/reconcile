package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/umerthow/reconcile/recon-job/internal/utils"
)

type RedisService struct {
	client *redis.Client
}

func NewRedisService(addr, password string, db int) (*RedisService, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 2,
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

// AcquireLock attempts to acquire a distributed lock for reconciliation job
// Ensures only one instance runs at a time
func (s *RedisService) AcquireLock(ctx context.Context, lockKey string, ttl time.Duration) (bool, error) {
	utils.Info().
		Str("lock_key", lockKey).
		Dur("ttl", ttl).
		Msg("Attempting to acquire job lock")

	success, err := s.client.SetNX(ctx, lockKey, time.Now().Unix(), ttl).Result()
	if err != nil {
		return false, fmt.Errorf("failed to acquire lock: %w", err)
	}

	if success {
		utils.Info().Msg("Lock acquired successfully")
	} else {
		utils.Warn().Msg("Lock already held by another instance")
	}

	return success, nil
}

// ReleaseLock releases the distributed lock
func (s *RedisService) ReleaseLock(ctx context.Context, lockKey string) error {
	utils.Info().Str("lock_key", lockKey).Msg("Releasing job lock")

	err := s.client.Del(ctx, lockKey).Err()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	return nil
}

// CacheProviderData caches provider reconciliation data temporarily
// MOCK IMPLEMENTATION - Logs for POC
func (s *RedisService) CacheProviderData(ctx context.Context, provider, date string, data interface{}) error {
	utils.Info().
		Str("provider", provider).
		Str("date", date).
		Msg("[MOCK] Caching provider data")

	// PSEUDOCODE: Real implementation
	// key := fmt.Sprintf("recon:provider:%s:%s", provider, date)
	// jsonData, err := json.Marshal(data)
	// if err != nil {
	//     return err
	// }
	// return s.client.Set(ctx, key, jsonData, 24*time.Hour).Err()

	return nil
}

// GetProviderData retrieves cached provider data
// MOCK IMPLEMENTATION - Returns empty for POC
func (s *RedisService) GetProviderData(ctx context.Context, provider, date string, dest interface{}) (bool, error) {
	utils.Info().
		Str("provider", provider).
		Str("date", date).
		Msg("[MOCK] Fetching cached provider data")

	// PSEUDOCODE: Real implementation
	// key := fmt.Sprintf("recon:provider:%s:%s", provider, date)
	// val, err := s.client.Get(ctx, key).Result()
	// if err == redis.Nil {
	//     return false, nil
	// }
	// if err != nil {
	//     return false, err
	// }
	// return true, json.Unmarshal([]byte(val), dest)

	return false, nil
}

// GetRunStats retrieves statistics for previous runs
// MOCK IMPLEMENTATION - Returns sample stats for POC
func (s *RedisService) GetRunStats(ctx context.Context, date string) (map[string]interface{}, error) {
	utils.Info().Str("date", date).Msg("[MOCK] Fetching run statistics")

	// MOCK DATA: Return sample stats
	mockStats := map[string]interface{}{
		"total_processed":     1500,
		"total_discrepancies": 12,
		"avg_duration_ms":     45000,
		"last_run":            time.Now().Add(-24 * time.Hour),
	}

	// PSEUDOCODE: Real implementation
	// key := fmt.Sprintf("recon:stats:%s", date)
	// val, err := s.client.Get(ctx, key).Result()
	// if err == redis.Nil {
	//     return nil, nil
	// }
	// if err != nil {
	//     return nil, err
	// }
	// var stats map[string]interface{}
	// err = json.Unmarshal([]byte(val), &stats)
	// return stats, err

	return mockStats, nil
}

// StoreRunStats stores statistics for completed run
func (s *RedisService) StoreRunStats(ctx context.Context, date string, stats map[string]interface{}) error {
	key := fmt.Sprintf("recon:stats:%s", date)
	jsonData, err := json.Marshal(stats)
	if err != nil {
		return err
	}

	utils.Info().Str("date", date).Msg("Storing run statistics")
	return s.client.Set(ctx, key, jsonData, 30*24*time.Hour).Err() // Keep for 30 days
}
