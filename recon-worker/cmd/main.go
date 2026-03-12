package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/umerthow/reconcile/recon-worker/internal/config"
	"github.com/umerthow/reconcile/recon-worker/internal/consumers"
	"github.com/umerthow/reconcile/recon-worker/internal/polling"
	"github.com/umerthow/reconcile/recon-worker/internal/services"
	"github.com/umerthow/reconcile/recon-worker/internal/utils"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	utils.InitLogger(cfg.LogLevel)
	utils.Info().
		Str("environment", cfg.Environment).
		Msg("Starting recon-worker")

	// Initialize services
	db, err := services.NewDatabaseService(cfg.DBConnectionString())
	if err != nil {
		utils.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	redis, err := services.NewRedisService(cfg.RedisAddress())
	if err != nil {
		utils.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	defer redis.Close()

	kafka, err := services.NewKafkaService(cfg)
	if err != nil {
		utils.Fatal().Err(err).Msg("Failed to connect to Kafka")
	}
	defer kafka.Close()

	custodyAPI := services.NewCustodyAPIService(cfg)

	utils.Info().Msg("All services initialized successfully")

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// WaitGroup for graceful shutdown
	var wg sync.WaitGroup

	// Start Xendit consumer
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := startXenditConsumer(ctx, cfg, db, redis, kafka); err != nil {
			utils.Error().Err(err).Msg("Xendit consumer stopped with error")
		}
	}()

	// Start Custody consumer
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := startCustodyConsumer(ctx, cfg, db, redis, kafka); err != nil {
			utils.Error().Err(err).Msg("Custody consumer stopped with error")
		}
	}()

	// Start Custody status updates consumer
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := startCustodyStatusConsumer(ctx, cfg, db, redis, kafka); err != nil {
			utils.Error().Err(err).Msg("Custody status consumer stopped with error")
		}
	}()

	// Start DLQ consumer
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := startDLQConsumer(ctx, cfg, db, redis, kafka); err != nil {
			utils.Error().Err(err).Msg("DLQ consumer stopped with error")
		}
	}()

	// Start Custody API poller
	poller := polling.NewCustodyPoller(db, custodyAPI, kafka, cfg.PollingIntervalSec)
	wg.Add(1)
	go func() {
		defer wg.Done()
		poller.Start(ctx)
	}()

	utils.Info().Msg("All workers started successfully")

	// Wait for interrupt signal
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm

	utils.Info().Msg("Shutdown signal received, stopping workers...")

	// Stop poller first
	poller.Stop()

	// Cancel context to stop all consumers
	cancel()

	// Wait for all goroutines to finish with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		utils.Info().Msg("All workers stopped gracefully")
	case <-time.After(30 * time.Second):
		utils.Warn().Msg("Shutdown timeout exceeded, forcing exit")
	}

	utils.Info().Msg("recon-worker shutdown complete")
}

func startXenditConsumer(ctx context.Context, cfg *config.Config, db *services.DatabaseService, redis *services.RedisService, kafka *services.KafkaService) error {
	consumerGroup, err := kafka.CreateConsumerGroup(cfg.KafkaGroupID + "-xendit")
	if err != nil {
		return err
	}
	defer consumerGroup.Close()

	consumer := consumers.NewXenditConsumer(db, redis, kafka)
	topics := []string{"xendit-webhooks"}

	utils.Info().
		Strs("topics", topics).
		Str("group_id", cfg.KafkaGroupID+"-xendit").
		Msg("Starting Xendit consumer")

	for {
		if err := consumerGroup.Consume(ctx, topics, consumer); err != nil {
			utils.Error().Err(err).Msg("Error from Xendit consumer")
		}

		if ctx.Err() != nil {
			return nil
		}
	}
}

func startCustodyConsumer(ctx context.Context, cfg *config.Config, db *services.DatabaseService, redis *services.RedisService, kafka *services.KafkaService) error {
	consumerGroup, err := kafka.CreateConsumerGroup(cfg.KafkaGroupID + "-custody")
	if err != nil {
		return err
	}
	defer consumerGroup.Close()

	consumer := consumers.NewCustodyConsumer(db, redis, kafka)
	topics := []string{"custody-webhooks"}

	utils.Info().
		Strs("topics", topics).
		Str("group_id", cfg.KafkaGroupID+"-custody").
		Msg("Starting Custody consumer")

	for {
		if err := consumerGroup.Consume(ctx, topics, consumer); err != nil {
			utils.Error().Err(err).Msg("Error from Custody consumer")
		}

		if ctx.Err() != nil {
			return nil
		}
	}
}

func startCustodyStatusConsumer(ctx context.Context, cfg *config.Config, db *services.DatabaseService, redis *services.RedisService, kafka *services.KafkaService) error {
	consumerGroup, err := kafka.CreateConsumerGroup(cfg.KafkaGroupID + "-custody-status")
	if err != nil {
		return err
	}
	defer consumerGroup.Close()

	// Reuse custody consumer logic for status updates
	consumer := consumers.NewCustodyConsumer(db, redis, kafka)
	topics := []string{"custody-status-updates"}

	utils.Info().
		Strs("topics", topics).
		Str("group_id", cfg.KafkaGroupID+"-custody-status").
		Msg("Starting Custody status updates consumer")

	for {
		if err := consumerGroup.Consume(ctx, topics, consumer); err != nil {
			utils.Error().Err(err).Msg("Error from Custody status consumer")
		}

		if ctx.Err() != nil {
			return nil
		}
	}
}

func startDLQConsumer(ctx context.Context, cfg *config.Config, db *services.DatabaseService, redis *services.RedisService, kafka *services.KafkaService) error {
	consumerGroup, err := kafka.CreateConsumerGroup(cfg.KafkaGroupID + "-dlq")
	if err != nil {
		return err
	}
	defer consumerGroup.Close()

	consumer := consumers.NewDLQConsumer(db, redis)
	topics := []string{"dlq-webhooks"}

	utils.Info().
		Strs("topics", topics).
		Str("group_id", cfg.KafkaGroupID+"-dlq").
		Msg("Starting DLQ consumer")

	for {
		if err := consumerGroup.Consume(ctx, topics, consumer); err != nil {
			utils.Error().Err(err).Msg("Error from DLQ consumer")
		}

		if ctx.Err() != nil {
			return nil
		}
	}
}
