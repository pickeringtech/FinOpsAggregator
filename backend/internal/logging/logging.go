package logging

import (
	"os"
	"strings"
	"time"

	"github.com/pickeringtech/FinOpsAggregator/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init initializes the global logger with the provided configuration
func Init(cfg config.LoggingConfig) {
	// Set log level
	level := parseLogLevel(cfg.Level)
	zerolog.SetGlobalLevel(level)

	// Configure console writer for better readability in development
	if isDevMode() {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.RFC3339,
		})
	}

	// Add caller information in debug mode
	if level == zerolog.DebugLevel {
		log.Logger = log.With().Caller().Logger()
	}

	log.Info().
		Str("level", level.String()).
		Msg("Logger initialized")
}

// parseLogLevel converts string log level to zerolog.Level
func parseLogLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

// isDevMode checks if we're running in development mode
func isDevMode() bool {
	env := os.Getenv("FINOPS_ENV")
	return env == "" || env == "development" || env == "dev"
}

// GetLogger returns a logger with the given component name
func GetLogger(component string) zerolog.Logger {
	return log.With().Str("component", component).Logger()
}

// WithRequestID adds a request ID to the logger context
func WithRequestID(logger zerolog.Logger, requestID string) zerolog.Logger {
	return logger.With().Str("request_id", requestID).Logger()
}

// WithUserID adds a user ID to the logger context
func WithUserID(logger zerolog.Logger, userID string) zerolog.Logger {
	return logger.With().Str("user_id", userID).Logger()
}

// WithNodeID adds a node ID to the logger context
func WithNodeID(logger zerolog.Logger, nodeID string) zerolog.Logger {
	return logger.With().Str("node_id", nodeID).Logger()
}

// WithRunID adds a computation run ID to the logger context
func WithRunID(logger zerolog.Logger, runID string) zerolog.Logger {
	return logger.With().Str("run_id", runID).Logger()
}
