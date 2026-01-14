package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ziyad/cms-ai/server/internal/store"
	"github.com/ziyad/cms-ai/server/internal/store/memory"
)

func TestServer_ListDeadLetterJobs(t *testing.T) {
	server := NewServer()
	memStore := server.Store.(*memory.MemoryStore)
	ctx := context.Background()

	// Create test jobs
	jobs := []store.Job{
		{
			ID:        "job-1",
			OrgID:     "org-1",
			Type:      store.JobRender,
			Status:    store.JobDeadLetter,
			InputRef:  "version-1",
			Error:     "Render failed permanently",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "job-2",
			OrgID:     "org-2", // Different org
			Type:      store.JobExport,
			Status:    store.JobDeadLetter,
			InputRef:  "version-2",
			Error:     "Export failed",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, job := range jobs {
		_, err := memStore.Jobs().Enqueue(ctx, job)
		require.NoError(t, err)
	}

	tests := []struct {
		name           string
		role           string
		expectedStatus int
		expectedJobs   int
	}{
		{
			name:           "admin can see DLQ jobs",
			role:           "Admin",
			expectedStatus: http.StatusOK,
			expectedJobs:   1, // Only job-1 from org-1
		},
		{
			name:           "owner can see DLQ jobs",
			role:           "Owner",
			expectedStatus: http.StatusOK,
			expectedJobs:   1, // Only job-1 from org-1
		},
		{
			name:           "editor cannot see DLQ jobs",
			role:           "Editor",
			expectedStatus: http.StatusForbidden,
			expectedJobs:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/v1/admin/jobs/dead-letter", nil)
			req.Header.Set("X-User-Id", "user-1")
			req.Header.Set("X-Org-Id", "org-1")
			req.Header.Set("X-Role", tt.role)

			w := httptest.NewRecorder()
			server.Handler().ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				jobs, ok := response["jobs"].([]interface{})
				require.True(t, ok)
				assert.Len(t, jobs, tt.expectedJobs)
			}
		})
	}
}

func TestServer_RetryDeadLetterJob(t *testing.T) {
	tests := []struct {
		name           string
		role           string
		expectedStatus int
		expectRetry    bool
	}{
		{
			name:           "admin can retry DLQ job",
			role:           "Admin",
			expectedStatus: http.StatusOK,
			expectRetry:    true,
		},
		{
			name:           "owner can retry DLQ job",
			role:           "Owner",
			expectedStatus: http.StatusOK,
			expectRetry:    true,
		},
		{
			name:           "editor cannot retry DLQ job",
			role:           "Editor",
			expectedStatus: http.StatusForbidden,
			expectRetry:    false,
		},
		{
			name:           "job not found",
			role:           "Admin",
			expectedStatus: http.StatusNotFound,
			expectRetry:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer()
			memStore := server.Store.(*memory.MemoryStore)
			ctx := context.Background()

			var jobID string
			if tt.name != "job not found" {
				job := store.Job{
					ID:         "job-retry-test",
					OrgID:      "org-1",
					Type:       store.JobRender,
					Status:     store.JobDeadLetter,
					InputRef:   "version-1",
					RetryCount: 2,
					MaxRetries: 3,
					Error:      "Failed after retries",
					CreatedAt:  time.Now(),
					UpdatedAt:  time.Now(),
				}
				created, err := memStore.Jobs().Enqueue(ctx, job)
				require.NoError(t, err)
				jobID = created.ID
			} else {
				jobID = "nonexistent"
			}

			req := httptest.NewRequest("POST", "/v1/admin/jobs/"+jobID+"/retry", nil)
			req.Header.Set("X-User-Id", "user-1")
			req.Header.Set("X-Org-Id", "org-1")
			req.Header.Set("X-Role", tt.role)

			w := httptest.NewRecorder()
			server.Handler().ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectRetry {
				jobAfter, found, err := memStore.Jobs().Get(ctx, "org-1", jobID)
				require.NoError(t, err)
				require.True(t, found)
				assert.Equal(t, store.JobQueued, jobAfter.Status)
				assert.Equal(t, 0, jobAfter.RetryCount)
				assert.Empty(t, jobAfter.Error)
			}
		})
	}
}
