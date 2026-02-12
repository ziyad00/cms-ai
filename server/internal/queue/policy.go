package queue

import (
	"time"
)

type RetryPolicy struct {
	MaxRetries    int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	BackoffFactor float64
}

type ErrorType string

const (
	ErrorTypeTransient ErrorType = "transient"
	ErrorTypePermanent ErrorType = "permanent"
)

var DefaultRetryPolicies = map[string]RetryPolicy{
	"render": {
		MaxRetries:    3,
		InitialDelay:  5 * time.Second,
		MaxDelay:      300 * time.Second,
		BackoffFactor: 2.0,
	},
	"preview": {
		MaxRetries:    2,
		InitialDelay:  3 * time.Second,
		MaxDelay:      60 * time.Second,
		BackoffFactor: 2.0,
	},
	"export": {
		MaxRetries:    5,
		InitialDelay:  10 * time.Second,
		MaxDelay:      600 * time.Second,
		BackoffFactor: 1.5,
	},
}

func GetRetryPolicy(jobType string) RetryPolicy {
	if policy, exists := DefaultRetryPolicies[jobType]; exists {
		return policy
	}
	return DefaultRetryPolicies["render"]
}

func ClassifyError(err error) ErrorType {
	if err == nil {
		return ErrorTypeTransient
	}

	errStr := err.Error()

	transientPatterns := []string{
		"connection refused",
		"timeout",
		"temporary",
		"network",
		"deadline exceeded",
		"context deadline exceeded",
		"resource temporarily unavailable",
	}

	for _, pattern := range transientPatterns {
		if contains(errStr, pattern) {
			return ErrorTypeTransient
		}
	}

	permanentPatterns := []string{
		"not found",
		"invalid",
		"unauthorized",
		"forbidden",
		"bad request",
		"malformed",
		"unsupported",
		"missing",
	}

	for _, pattern := range permanentPatterns {
		if contains(errStr, pattern) {
			return ErrorTypePermanent
		}
	}

	return ErrorTypeTransient
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func CalculateNextRetryDelay(policy RetryPolicy, retryCount int) time.Duration {
	if retryCount <= 0 {
		return policy.InitialDelay
	}

	delay := float64(policy.InitialDelay) *
		pow(policy.BackoffFactor, float64(retryCount-1))

	if delay > float64(policy.MaxDelay) {
		delay = float64(policy.MaxDelay)
	}

	return time.Duration(delay)
}

func pow(base, exp float64) float64 {
	if exp == 0 {
		return 1
	}
	result := base
	for i := 1; i < int(exp); i++ {
		result *= base
	}
	return result
}
