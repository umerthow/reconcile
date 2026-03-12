package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/umerthow/reconcile/recon-job/internal/config"
	"github.com/umerthow/reconcile/recon-job/internal/reconciliation"
	"github.com/umerthow/reconcile/recon-job/internal/scheduler"
	"github.com/umerthow/reconcile/recon-job/internal/services"
	"github.com/umerthow/reconcile/recon-job/internal/utils"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	utils.InitLogger(cfg.LogLevel)
	utils.Info().Msg("Starting Reconciliation Job Service")

	// Print configuration
	utils.Info().
		Str("cron_schedule", cfg.CronSchedule).
		Str("timezone", cfg.Timezone).
		Int("batch_size", cfg.BatchSize).
		Int("concurrency", cfg.Concurrency).
		Msg("Configuration loaded")

	// Initialize services
	db, redis, notification, err := initializeServices(cfg)
	if err != nil {
		utils.Fatal().Err(err).Msg("Failed to initialize services")
	}
	defer db.Close()
	defer redis.Close()

	// Initialize reconciliation engine
	engine := reconciliation.NewEngine(
		db,
		redis,
		notification,
		cfg.BatchSize,
		cfg.Concurrency,
		cfg.DiscrepancyThreshold,
	)

	// Initialize scheduler
	sched, err := scheduler.NewScheduler(engine, redis, cfg.CronSchedule, cfg.Timezone)
	if err != nil {
		utils.Fatal().Err(err).Msg("Failed to initialize scheduler")
	}

	// Start scheduler
	ctx := context.Background()
	if err := sched.Start(ctx); err != nil {
		utils.Fatal().Err(err).Msg("Failed to start scheduler")
	}

	utils.Info().Msg("Reconciliation Job Service started successfully")
	utils.Info().Msg("Waiting for scheduled execution or shutdown signal...")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	utils.Info().Msg("Shutdown signal received")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	utils.Info().Msg("Shutting down gracefully...")
	if err := sched.Stop(shutdownCtx); err != nil {
		utils.Error().Err(err).Msg("Error during scheduler shutdown")
	}

	utils.Info().Msg("Reconciliation Job Service stopped")
}

// initializeServices sets up database, redis, and notification services
func initializeServices(cfg *config.Config) (*services.DatabaseService, *services.RedisService, *services.NotificationService, error) {
	// Database connection
	dbConnStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	db, err := services.NewDatabaseService(dbConnStr)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("database initialization failed: %w", err)
	}

	// Redis connection
	redis, err := services.NewRedisService(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		db.Close()
		return nil, nil, nil, fmt.Errorf("redis initialization failed: %w", err)
	}

	// Notification service
	emailRecipients := []string{}
	if cfg.EmailRecipients != "" {
		// Split comma-separated emails
		// For POC: simplified parsing
		emailRecipients = append(emailRecipients, cfg.EmailRecipients)
	}

	notification := services.NewNotificationService(
		cfg.NotificationEnabled,
		cfg.SlackWebhook,
		emailRecipients,
	)

	return db, redis, notification, nil
}
