package utils

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// InitLogger initializes the global logger
func InitLogger(level string) {
	// Set time format
	zerolog.TimeFieldFormat = time.RFC3339

	// Parse log level
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(logLevel)

	// Pretty logging for development
	if os.Getenv("GO_ENV") == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		})
	}
}

// GetLogger returns a logger with context
func GetLogger(component string) zerolog.Logger {
	return log.With().Str("component", component).Logger()
}
