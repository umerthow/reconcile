package utils

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Logger zerolog.Logger

func InitLogger(logLevel string) {
	// Set log level
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Pretty console output for development
	if os.Getenv("GO_ENV") == "development" {
		Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	} else {
		// JSON output for production
		Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}
}

func Info() *zerolog.Event {
	return Logger.Info()
}

func Error() *zerolog.Event {
	return Logger.Error()
}

func Debug() *zerolog.Event {
	return Logger.Debug()
}

func Warn() *zerolog.Event {
	return Logger.Warn()
}

func Fatal() *zerolog.Event {
	return Logger.Fatal()
}
