package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration for the reconciliation job service
type Config struct {
	// Database
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string

	// Redis
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// Schedule
	CronSchedule string // e.g., "0 2 * * *" = 02:00 daily
	Timezone     string // e.g., "Asia/Jakarta"

	// Job Processing
	BatchSize            int
	Concurrency          int
	DiscrepancyThreshold float64

	// Notifications
	NotificationEnabled bool
	SlackWebhook        string
	EmailRecipients     string // Comma-separated

	// Logging
	LogLevel string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	redisDB, err := strconv.Atoi(getEnv("REDIS_DB", "0"))
	if err != nil {
		return nil, fmt.Errorf("invalid REDIS_DB: %w", err)
	}

	batchSize, err := strconv.Atoi(getEnv("BATCH_SIZE", "1000"))
	if err != nil {
		return nil, fmt.Errorf("invalid BATCH_SIZE: %w", err)
	}

	concurrency, err := strconv.Atoi(getEnv("CONCURRENCY", "5"))
	if err != nil {
		return nil, fmt.Errorf("invalid CONCURRENCY: %w", err)
	}

	threshold, err := strconv.ParseFloat(getEnv("DISCREPANCY_THRESHOLD", "0.01"), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid DISCREPANCY_THRESHOLD: %w", err)
	}

	notifEnabled, err := strconv.ParseBool(getEnv("NOTIFICATION_ENABLED", "false"))
	if err != nil {
		return nil, fmt.Errorf("invalid NOTIFICATION_ENABLED: %w", err)
	}

	return &Config{
		// Database
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBUser:     getEnv("DB_USER", "reconcile_user"),
		DBPassword: getEnv("DB_PASSWORD", "reconcile_password"),
		DBName:     getEnv("DB_NAME", "reconcile_db"),

		// Redis
		RedisAddr:     getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,

		// Schedule
		CronSchedule: getEnv("CRON_SCHEDULE", "0 2 * * *"),
		Timezone:     getEnv("TIMEZONE", "Asia/Jakarta"),

		// Job Processing
		BatchSize:            batchSize,
		Concurrency:          concurrency,
		DiscrepancyThreshold: threshold,

		// Notifications
		NotificationEnabled: notifEnabled,
		SlackWebhook:        getEnv("SLACK_WEBHOOK", ""),
		EmailRecipients:     getEnv("EMAIL_RECIPIENTS", ""),

		// Logging
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
