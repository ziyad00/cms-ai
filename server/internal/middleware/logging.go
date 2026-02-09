package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
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

// ValidationMiddleware provides comprehensive input validation and sanitization
func ValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Validate request method
		if !isValidHTTPMethod(r.Method) {
			logger.WithContext(ctx).Warn("invalid_http_method", "method", r.Method)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Validate URL path for suspicious patterns
		if hasSuspiciousPath(r.URL.Path) {
			logger.WithContext(ctx).Warn("suspicious_url_path", "path", r.URL.Path)
			http.Error(w, "Invalid request path", http.StatusBadRequest)
			return
		}

		// Validate headers for injection attempts
		if hasSuspiciousHeaders(r.Header) {
			logger.WithContext(ctx).Warn("suspicious_headers")
			http.Error(w, "Invalid request headers", http.StatusBadRequest)
			return
		}

		// Validate and sanitize JSON body for POST/PUT requests
		if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
			if err := validateJSONBody(r); err != nil {
				logger.WithContext(ctx).Warn("invalid_json_body", "error", err)
				http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
				return
			}
		}

		// Validate query parameters
		if err := validateQueryParams(r); err != nil {
			logger.WithContext(ctx).Warn("invalid_query_params", "error", err)
			http.Error(w, fmt.Sprintf("Invalid query parameters: %v", err), http.StatusBadRequest)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// isValidHTTPMethod checks if the HTTP method is allowed
func isValidHTTPMethod(method string) bool {
	validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"}
	for _, valid := range validMethods {
		if method == valid {
			return true
		}
	}
	return false
}

// hasSuspiciousPath detects path traversal and injection attempts
func hasSuspiciousPath(path string) bool {
	// Check for path traversal attempts
	suspiciousPatterns := []string{
		"../", "..\\", "%2e%2e", "%2E%2E",
		"<script", "</script", "javascript:",
		"data:", "vbscript:", "onload=",
		"eval(", "setTimeout(", "setInterval(",
	}

	lowerPath := strings.ToLower(path)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerPath, pattern) {
			return true
		}
	}
	return false
}

// hasSuspiciousHeaders detects header injection attempts
func hasSuspiciousHeaders(headers http.Header) bool {
	for name, values := range headers {
		for _, value := range values {
			// Check for CRLF injection
			if strings.Contains(value, "\r") || strings.Contains(value, "\n") {
				return true
			}
			// Check for suspicious scripts
			lowerValue := strings.ToLower(value)
			if strings.Contains(lowerValue, "<script") ||
			   strings.Contains(lowerValue, "javascript:") ||
			   strings.Contains(lowerValue, "data:text/html") {
				return true
			}
		}
		// Check for suspicious header names
		lowerName := strings.ToLower(name)
		if strings.Contains(lowerName, "..") || strings.Contains(lowerName, "/") {
			return true
		}
	}
	return false
}

// validateJSONBody validates and limits JSON request bodies
func validateJSONBody(r *http.Request) error {
	if r.Body == nil {
		return nil
	}

	// Limit body size to 10MB to prevent DoS
	const maxBodySize = 10 * 1024 * 1024
	r.Body = http.MaxBytesReader(nil, r.Body, maxBodySize)

	// Read and validate JSON
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	// Skip validation for empty bodies
	if len(body) == 0 {
		return nil
	}

	// Validate JSON syntax
	var jsonData interface{}
	if err := json.Unmarshal(body, &jsonData); err != nil {
		return fmt.Errorf("invalid JSON format: %w", err)
	}

	// Check for deeply nested JSON (DoS protection)
	if err := validateJSONDepth(jsonData, 0, 10); err != nil {
		return err
	}

	// Restore body for handlers to read
	r.Body = io.NopCloser(strings.NewReader(string(body)))
	return nil
}

// validateJSONDepth prevents deeply nested JSON DoS attacks
func validateJSONDepth(data interface{}, currentDepth, maxDepth int) error {
	if currentDepth > maxDepth {
		return fmt.Errorf("JSON nesting too deep (max: %d)", maxDepth)
	}

	switch v := data.(type) {
	case map[string]interface{}:
		for _, value := range v {
			if err := validateJSONDepth(value, currentDepth+1, maxDepth); err != nil {
				return err
			}
		}
	case []interface{}:
		for _, value := range v {
			if err := validateJSONDepth(value, currentDepth+1, maxDepth); err != nil {
				return err
			}
		}
	}
	return nil
}

// validateQueryParams validates URL query parameters
func validateQueryParams(r *http.Request) error {
	for key, values := range r.URL.Query() {
		// Validate parameter name
		if !isValidParamName(key) {
			return fmt.Errorf("invalid parameter name: %s", key)
		}

		// Validate parameter values
		for _, value := range values {
			if !isValidParamValue(value) {
				return fmt.Errorf("invalid parameter value for %s", key)
			}
		}
	}
	return nil
}

// isValidParamName validates query parameter names
func isValidParamName(name string) bool {
	// Only allow alphanumeric, underscore, hyphen
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_-]+$", name)
	return matched && len(name) <= 64 // Reasonable length limit
}

// isValidParamValue validates query parameter values
func isValidParamValue(value string) bool {
	// Check length limit
	if len(value) > 1000 {
		return false
	}

	// Check for suspicious patterns
	suspiciousPatterns := []string{
		"<script", "</script", "javascript:",
		"data:", "vbscript:", "onload=",
		"eval(", "setTimeout(", "../",
	}

	lowerValue := strings.ToLower(value)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(lowerValue, pattern) {
			return false
		}
	}

	return true
}