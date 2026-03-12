package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Kafka    KafkaConfig
	JWT      JWTConfig
	Logger   LoggerConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
	// Connection pool settings
	MaxOpenConns int
	MaxIdleConns int
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type KafkaConfig struct {
	Brokers   []string
	SSLEnable bool
	Username  string
	Password  string
	Topics    KafkaTopics
}

type KafkaTopics struct {
	XenditWebhooks       string
	CustodyWebhooks      string
	CustodyStatusUpdates string
	DLQ                  string
}

type JWTConfig struct {
	Secret string
}

type LoggerConfig struct {
	Level string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "3000"),
			Env:  getEnv("GO_ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnv("DB_PORT", "3306"),
			Name:         getEnv("DB_NAME", "reconciliation"),
			User:         getEnv("DB_USER", "recon_user"),
			Password:     getEnv("DB_PASSWORD", "recon_pass"),
			MaxOpenConns: getEnvAsInt("DB_MAX_OPEN_CONNS", 20),
			MaxIdleConns: getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Kafka: KafkaConfig{
			Brokers:   strings.Split(getEnv("KAFKA_BROKERS", "localhost:29092"), ","),
			SSLEnable: getEnvAsBool("KAFKA_SSL_ENABLE", false),
			Username:  getEnv("KAFKA_USERNAME", ""),
			Password:  getEnv("KAFKA_PASSWORD", ""),
			Topics: KafkaTopics{
				XenditWebhooks:       "xendit-webhooks",
				CustodyWebhooks:      "custody-webhooks",
				CustodyStatusUpdates: "custody-status-updates",
				DLQ:                  "dlq-webhooks",
			},
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		},
		Logger: LoggerConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
	}
}

// Helper functions to get environment variables with defaults
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}

// DSN returns the MySQL data source name
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.Name,
	)
}

// Address returns the Redis address
func (r *RedisConfig) Address() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

// IsDevelopment checks if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Server.Env == "development"
}

// IsProduction checks if running in production mode
func (c *Config) IsProduction() bool {
	return c.Server.Env == "production"
}
