package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
)

var (
	// Global logger instance
	Logger *slog.Logger
)

// LogLevel represents different log levels
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// Config holds logger configuration
type Config struct {
	Level  LogLevel `json:"level"`
	Format string   `json:"format"` // "json" or "text"
	Output io.Writer
}

// DefaultConfig returns default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:  LevelInfo,
		Format: "json",
		Output: os.Stdout,
	}
}

// Initialize sets up the global structured logger
func Initialize(config *Config) {
	if config == nil {
		config = DefaultConfig()
	}

	if config.Output == nil {
		config.Output = os.Stdout
	}

	// Convert string level to slog.Level
	var level slog.Level
	switch config.Level {
	case LevelDebug:
		level = slog.LevelDebug
	case LevelInfo:
		level = slog.LevelInfo
	case LevelWarn:
		level = slog.LevelWarn
	case LevelError:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	// Create handler based on format
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: level,
	}

	if config.Format == "json" {
		handler = slog.NewJSONHandler(config.Output, opts)
	} else {
		handler = slog.NewTextHandler(config.Output, opts)
	}

	Logger = slog.New(handler)
	slog.SetDefault(Logger)
}

func init() {
	// Ensure Logger is never nil by setting a default
	if Logger == nil {
		Initialize(nil)
	}
}

// Context keys for structured logging
type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	UserIDKey    contextKey = "user_id"
	OrgIDKey     contextKey = "org_id"
	TraceIDKey   contextKey = "trace_id"
)

// WithContext adds context values to logger
func WithContext(ctx context.Context) *slog.Logger {
	logger := Logger

	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		logger = logger.With("request_id", requestID)
	}
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		logger = logger.With("user_id", userID)
	}
	if orgID, ok := ctx.Value(OrgIDKey).(string); ok {
		logger = logger.With("org_id", orgID)
	}
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		logger = logger.With("trace_id", traceID)
	}

	return logger
}

// Component loggers with structured fields
func Database() *slog.Logger {
	return Logger.With("component", "database")
}

func API() *slog.Logger {
	return Logger.With("component", "api")
}

func Auth() *slog.Logger {
	return Logger.With("component", "auth")
}

func AI() *slog.Logger {
	return Logger.With("component", "ai")
}

func Jobs() *slog.Logger {
	return Logger.With("component", "jobs")
}

func Storage() *slog.Logger {
	return Logger.With("component", "storage")
}

// Helper functions for common log patterns
func LogHTTPRequest(ctx context.Context, method, path string, statusCode int, duration int64) {
	WithContext(ctx).Info("http_request",
		"method", method,
		"path", path,
		"status_code", statusCode,
		"duration_ms", duration,
	)
}

func LogDatabaseQuery(ctx context.Context, query string, duration int64, err error) {
	logger := Database().With(
		"query", query,
		"duration_ms", duration,
	)

	if err != nil {
		logger.Error("database_query_failed", "error", err)
	} else {
		logger.Debug("database_query_success")
	}
}

func LogAIRequest(ctx context.Context, model string, tokenUsage int, cost float64, duration int64) {
	AI().With(
		"model", model,
		"token_usage", tokenUsage,
		"cost", cost,
		"duration_ms", duration,
	).Info("ai_request_completed")
}

func LogJobProcessing(ctx context.Context, jobID, jobType, status string) {
	Jobs().With(
		"job_id", jobID,
		"job_type", jobType,
		"status", status,
	).Info("job_status_update")
}

func LogAuthEvent(ctx context.Context, event string, userID string, success bool) {
	logger := Auth().With(
		"event", event,
		"user_id", userID,
		"success", success,
	)

	if success {
		logger.Info("auth_event")
	} else {
		logger.Warn("auth_event_failed")
	}
}

func LogError(ctx context.Context, component string, operation string, err error, fields ...any) {
	logger := Logger.With(
		"component", component,
		"operation", operation,
		"error", err.Error(),
	)

	// Add additional fields
	if len(fields) > 0 {
		logger = logger.With(fields...)
	}

	logger.Error("operation_failed")
}

func LogBusinessEvent(ctx context.Context, event string, fields ...any) {
	logger := Logger.With("event", event)

	if len(fields) > 0 {
		logger = logger.With(fields...)
	}

	logger.Info("business_event")
}