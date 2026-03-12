package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Environment string
	LogLevel    string

	// Database
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string

	// Redis
	RedisHost string
	RedisPort string

	// Kafka
	KafkaBrokers   []string
	KafkaGroupID   string
	KafkaSSLEnable bool
	KafkaUsername  string
	KafkaPassword  string

	// Custody API
	CustodyAPIURL string
	CustodyAPIKey string

	// Polling
	PollingIntervalSec int

	// Worker settings
	WorkerConcurrency int
	MaxRetries        int
	RetryBackoff      time.Duration
}

func Load() *Config {
	return &Config{
		Environment: getEnv("GO_ENV", "development"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "3306"),
		DBName:     getEnv("DB_NAME", "reconciliation"),
		DBUser:     getEnv("DB_USER", "recon_user"),
		DBPassword: getEnv("DB_PASSWORD", "recon_pass"),

		RedisHost: getEnv("REDIS_HOST", "localhost"),
		RedisPort: getEnv("REDIS_PORT", "6379"),

		KafkaBrokers:   parseKafkaBrokers(getEnv("KAFKA_BROKERS", "localhost:29092")),
		KafkaGroupID:   getEnv("KAFKA_GROUP_ID", "recon-worker-group"),
		KafkaSSLEnable: getEnvBool("KAFKA_SSL_ENABLE", false),
		KafkaUsername:  getEnv("KAFKA_USERNAME", ""),
		KafkaPassword:  getEnv("KAFKA_PASSWORD", ""),

		CustodyAPIURL: getEnv("CUSTODY_API_URL", "https://api.custody-provider.com"),
		CustodyAPIKey: getEnv("CUSTODY_API_KEY", "mock-custody-api-key"),

		PollingIntervalSec: getEnvInt("POLLING_INTERVAL_SEC", 300), // 5 minutes

		WorkerConcurrency: getEnvInt("WORKER_CONCURRENCY", 5),
		MaxRetries:        getEnvInt("MAX_RETRIES", 3),
		RetryBackoff:      time.Duration(getEnvInt("RETRY_BACKOFF_SEC", 2)) * time.Second,
	}
}

func (c *Config) DBConnectionString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=UTC",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
}

func (c *Config) RedisAddress() string {
	return fmt.Sprintf("%s:%s", c.RedisHost, c.RedisPort)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func parseKafkaBrokers(brokers string) []string {
	if brokers == "" {
		return []string{"localhost:29092"}
	}
	// Simple split by comma
	result := []string{}
	current := ""
	for _, char := range brokers {
		if char == ',' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}
