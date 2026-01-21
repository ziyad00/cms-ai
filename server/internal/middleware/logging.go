package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/ziyad/cms-ai/server/internal/auth"
	"github.com/ziyad/cms-ai/server/internal/logger"
)

// GenerateRequestID creates a random request ID
func GenerateRequestID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// LoggingMiddleware adds request logging with structured logging
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := GenerateRequestID()

		// Add request ID to context
		ctx := context.WithValue(r.Context(), logger.RequestIDKey, requestID)

		// Add authentication context if available
		if identity, ok := auth.GetIdentity(r.Context()); ok {
			ctx = context.WithValue(ctx, logger.UserIDKey, identity.UserID)
			ctx = context.WithValue(ctx, logger.OrgIDKey, identity.OrgID)
		}

		r = r.WithContext(ctx)

		// Create response writer wrapper to capture status code
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     200,
		}

		// Add request ID header to response
		wrapped.Header().Set("X-Request-ID", requestID)

		// Log request start
		logger.WithContext(ctx).Info("request_start",
			"method", r.Method,
			"path", r.URL.Path,
			"query", r.URL.RawQuery,
			"user_agent", r.UserAgent(),
			"remote_addr", r.RemoteAddr,
		)

		// Process request
		next.ServeHTTP(wrapped, r)

		// Log request completion
		duration := time.Since(start).Milliseconds()
		logger.LogHTTPRequest(ctx, r.Method, r.URL.Path, wrapped.statusCode, duration)

		// Log slow requests
		if duration > 1000 { // > 1 second
			logger.WithContext(ctx).Warn("slow_request",
				"method", r.Method,
				"path", r.URL.Path,
				"duration_ms", duration,
			)
		}
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RecoveryMiddleware handles panics with structured logging
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				logger.WithContext(r.Context()).Error("panic_recovered",
					"error", err,
					"method", r.Method,
					"path", r.URL.Path,
				)

				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}