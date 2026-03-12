package services

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/umerthow/reconcile/recon-api/internal/config"
	"github.com/umerthow/reconcile/recon-api/internal/utils"
)

var redisLogger = utils.GetLogger("redis")

type RedisService struct {
	client *redis.Client
}

func NewRedisService(cfg *config.RedisConfig) (*RedisService, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	redisLogger.Info().Msg("Redis connection established")

	return &RedisService{client: client}, nil
}

func (s *RedisService) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

// CheckWebhookDuplicate checks if a webhook has already been processed
func (s *RedisService) CheckWebhookDuplicate(ctx context.Context, webhookID string) (bool, error) {
	// Check if key exists
	exists, err := s.client.Exists(ctx, webhookKey(webhookID)).Result()
	if err != nil {
		redisLogger.Error().Err(err).Str("webhook_id", webhookID).Msg("Failed to check webhook duplicate")
		return false, err
	}

	return exists > 0, nil
}

// MarkWebhookProcessed marks a webhook as processed with 48-hour TTL
func (s *RedisService) MarkWebhookProcessed(ctx context.Context, webhookID string) error {
	err := s.client.Set(ctx, webhookKey(webhookID), "1", 48*time.Hour).Err()
	if err != nil {
		redisLogger.Error().Err(err).Str("webhook_id", webhookID).Msg("Failed to mark webhook as processed")
		return err
	}

	redisLogger.Debug().Str("webhook_id", webhookID).Msg("Webhook marked as processed")
	return nil
}

// Health check
func (s *RedisService) Ping(ctx context.Context) error {
	return s.client.Ping(ctx).Err()
}

func webhookKey(webhookID string) string {
	return "webhook:" + webhookID
}
