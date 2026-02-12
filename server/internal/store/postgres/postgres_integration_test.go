package postgres

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ziyad/cms-ai/server/internal/store"
)

func TestPostgresJobStore_MetadataSerialization(t *testing.T) {
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("Skipping postgres integration test: TEST_DATABASE_URL not set")
	}

	ctx := context.Background()
	s, err := New(dsn)
	require.NoError(t, err)
	defer s.Close()

	// Clear jobs for test
	err = s.db.Exec("DELETE FROM jobs").Error
	require.NoError(t, err)

	jobStore := s.Jobs()

	metadata := store.JSONMap{
		"test-key": "test-value",
		"version":  "v1.0",
	}

	job := store.Job{
		OrgID:    "test-org",
		Type:     store.JobExport,
		Status:   store.JobQueued,
		InputRef: "test-ref",
		Metadata: &metadata,
	}

	// 1. Test Enqueue
	created, err := jobStore.Enqueue(ctx, job)
	require.NoError(t, err, "Should not fail to enqueue job with metadata")
	assert.NotEmpty(t, created.ID)

	// 2. Test Get
	got, found, err := jobStore.Get(ctx, "test-org", created.ID)
	require.NoError(t, err)
	assert.True(t, found)
	require.NotNil(t, got.Metadata)
	assert.Equal(t, "test-value", (*got.Metadata)["test-key"])
	assert.Equal(t, "v1.0", (*got.Metadata)["version"])

	// 3. Test Update
	(*got.Metadata)["new-key"] = "new-value"
	got.Status = store.JobRunning
	updated, err := jobStore.Update(ctx, got)
	require.NoError(t, err)
	assert.Equal(t, store.JobRunning, updated.Status)

	// 4. Verify Update persisted
	got2, found, err := jobStore.Get(ctx, "test-org", created.ID)
	require.NoError(t, err)
	assert.True(t, found)
	require.NotNil(t, got2.Metadata)
	assert.Equal(t, "new-value", (*got2.Metadata)["new-key"])
}
