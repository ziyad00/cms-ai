package queue

import (
	"fmt"
	"testing"
	"time"

	"github.com/ziyad/cms-ai/server/internal/store"
)

func TestGetRetryPolicy(t *testing.T) {
	tests := []struct {
		name     string
		jobType  string
		expected RetryPolicy
	}{
		{
			name:    "render job policy",
			jobType: "render",
			expected: RetryPolicy{
				MaxRetries:    3,
				InitialDelay:  5 * time.Second,
				MaxDelay:      300 * time.Second,
				BackoffFactor: 2.0,
			},
		},
		{
			name:    "preview job policy",
			jobType: "preview",
			expected: RetryPolicy{
				MaxRetries:    2,
				InitialDelay:  3 * time.Second,
				MaxDelay:      60 * time.Second,
				BackoffFactor: 2.0,
			},
		},
		{
			name:    "export job policy",
			jobType: "export",
			expected: RetryPolicy{
				MaxRetries:    5,
				InitialDelay:  10 * time.Second,
				MaxDelay:      600 * time.Second,
				BackoffFactor: 1.5,
			},
		},
		{
			name:    "unknown job type defaults to render",
			jobType: "unknown",
			expected: RetryPolicy{
				MaxRetries:    3,
				InitialDelay:  5 * time.Second,
				MaxDelay:      300 * time.Second,
				BackoffFactor: 2.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := GetRetryPolicy(tt.jobType)
			if policy.MaxRetries != tt.expected.MaxRetries {
				t.Errorf("MaxRetries = %v, want %v", policy.MaxRetries, tt.expected.MaxRetries)
			}
			if policy.InitialDelay != tt.expected.InitialDelay {
				t.Errorf("InitialDelay = %v, want %v", policy.InitialDelay, tt.expected.InitialDelay)
			}
			if policy.MaxDelay != tt.expected.MaxDelay {
				t.Errorf("MaxDelay = %v, want %v", policy.MaxDelay, tt.expected.MaxDelay)
			}
			if policy.BackoffFactor != tt.expected.BackoffFactor {
				t.Errorf("BackoffFactor = %v, want %v", policy.BackoffFactor, tt.expected.BackoffFactor)
			}
		})
	}
}

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected ErrorType
	}{
		{
			name:     "nil error is transient",
			err:      nil,
			expected: ErrorTypeTransient,
		},
		{
			name:     "connection refused is transient",
			err:      &testError{"connection refused"},
			expected: ErrorTypeTransient,
		},
		{
			name:     "timeout is transient",
			err:      &testError{"timeout occurred"},
			expected: ErrorTypeTransient,
		},
		{
			name:     "network error is transient",
			err:      &testError{"network unreachable"},
			expected: ErrorTypeTransient,
		},
		{
			name:     "not found is permanent",
			err:      &testError{"template not found"},
			expected: ErrorTypePermanent,
		},
		{
			name:     "invalid is permanent",
			err:      &testError{"invalid input"},
			expected: ErrorTypePermanent,
		},
		{
			name:     "unauthorized is permanent",
			err:      &testError{"unauthorized access"},
			expected: ErrorTypePermanent,
		},
		{
			name:     "missing metadata is permanent",
			err:      &testError{"missing job metadata"},
			expected: ErrorTypePermanent,
		},
		{
			name:     "unknown error defaults to transient",
			err:      &testError{"some random error"},
			expected: ErrorTypeTransient,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ClassifyError(tt.err)
			if result != tt.expected {
				t.Errorf("ClassifyError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCalculateNextRetryDelay(t *testing.T) {
	policy := RetryPolicy{
		MaxRetries:    3,
		InitialDelay:  10 * time.Second,
		MaxDelay:      100 * time.Second,
		BackoffFactor: 2.0,
	}

	tests := []struct {
		name       string
		retryCount int
		expected   time.Duration
	}{
		{
			name:       "first retry",
			retryCount: 0,
			expected:   10 * time.Second,
		},
		{
			name:       "second retry",
			retryCount: 1,
			expected:   10 * time.Second,
		},
		{
			name:       "third retry",
			retryCount: 2,
			expected:   20 * time.Second,
		},
		{
			name:       "fourth retry",
			retryCount: 3,
			expected:   40 * time.Second,
		},
		{
			name:       "delay capped at max",
			retryCount: 10,
			expected:   100 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := CalculateNextRetryDelay(policy, tt.retryCount)
			if delay != tt.expected {
				t.Errorf("CalculateNextRetryDelay() = %v, want %v", delay, tt.expected)
			}
		})
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestJobDeduplicationID(t *testing.T) {
	tests := []struct {
		name          string
		jobType       store.JobType
		inputRef      string
		expectedDedup string
	}{
		{
			name:          "render job deduplication",
			jobType:       store.JobRender,
			inputRef:      "version-123",
			expectedDedup: "render-version-123",
		},
		{
			name:          "export job deduplication",
			jobType:       store.JobExport,
			inputRef:      "version-456",
			expectedDedup: "export-version-456",
		},
		{
			name:          "preview job deduplication",
			jobType:       store.JobPreview,
			inputRef:      "version-789",
			expectedDedup: "preview-version-789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dedupID := fmt.Sprintf("%s-%s", string(tt.jobType), tt.inputRef)
			if dedupID != tt.expectedDedup {
				t.Errorf("DeduplicationID = %v, want %v", dedupID, tt.expectedDedup)
			}
		})
	}
}
