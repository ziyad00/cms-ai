package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ziyad/cms-ai/server/internal/store"
)

// TestPostgresJobStore_SchemaAlignment verifies that the PostgreSQL schema
// actually supports all the Types and Statuses defined in the Go models.
func TestPostgresJobStore_SchemaAlignment(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("Skipping postgres integration test: TEST_DATABASE_URL not set")
	}

	ctx := context.Background()
	s, err := New(dsn)
	require.NoError(t, err)
	defer s.Close()

	jobStore := s.Jobs()

	// Test all Job Types
	jobTypes := []store.JobType{
		store.JobRender,
		store.JobPreview,
		store.JobExport,
		store.JobGenerate,
		store.JobBind,
	}

	// Test all Job Statuses
	jobStatuses := []store.JobStatus{
		store.JobQueued,
		store.JobRunning,
		store.JobDone,
		store.JobFailed,
		store.JobRetry,
		store.JobDeadLetter,
	}

	for _, jt := range jobTypes {
		for _, js := range jobStatuses {
			t.Run(string(jt)+"-"+string(js), func(t *testing.T) {
				job := store.Job{
					OrgID:    "00000000-0000-0000-0000-000000000000", // Use valid UUID string
					Type:     jt,
					Status:   js,
					InputRef: "test-ref",
				}
				
				created, err := jobStore.Enqueue(ctx, job)
				if err != nil {
					t.Errorf("DB rejected JobType '%s' or JobStatus '%s': %v. Ensure migrations 006 is applied.", jt, js, err)
				} else {
					// Clean up
					_, _ = s.db.Exec("DELETE FROM jobs WHERE id = $1", created.ID)
				}
			})
		}
	}
}
