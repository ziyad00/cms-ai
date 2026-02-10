package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ziyad/cms-ai/server/internal/store"
)

func TestTemplateStore(t *testing.T) {
	s := New()
	ctx := context.Background()

	// Create
	tmpl := store.Template{
		ID:    "tmpl-1",
		OrgID: "org-1",
		Name:  "Test Template",
	}
	created, err := s.Templates().CreateTemplate(ctx, tmpl)
	require.NoError(t, err)
	assert.Equal(t, tmpl.ID, created.ID)
	assert.False(t, created.CreatedAt.IsZero())

	// Get
	got, found, err := s.Templates().GetTemplate(ctx, "org-1", "tmpl-1")
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "Test Template", got.Name)

	// List
	list, err := s.Templates().ListTemplates(ctx, "org-1")
	require.NoError(t, err)
	assert.Len(t, list, 1)

	// Update
	created.Name = "Updated Name"
	updated, err := s.Templates().UpdateTemplate(ctx, created)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)

	// Get Updated
	gotUpdated, found, err := s.Templates().GetTemplate(ctx, "org-1", "tmpl-1")
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "Updated Name", gotUpdated.Name)
}

func TestJobStore(t *testing.T) {
	s := New()
	ctx := context.Background()

	// Enqueue
	job := store.Job{
		ID:    "job-1",
		OrgID: "org-1",
		Type:  store.JobRender,
		Status: store.JobQueued,
	}
	enqueued, err := s.Jobs().Enqueue(ctx, job)
	require.NoError(t, err)
	assert.Equal(t, job.ID, enqueued.ID)
	assert.Equal(t, store.JobQueued, enqueued.Status)

	// Get
	got, found, err := s.Jobs().Get(ctx, "org-1", "job-1")
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "job-1", got.ID)

	// Update
	enqueued.Status = store.JobRunning
	updated, err := s.Jobs().Update(ctx, enqueued)
	require.NoError(t, err)
	assert.Equal(t, store.JobRunning, updated.Status)

	// List Queued
	job2 := store.Job{
		ID:    "job-2",
		OrgID: "org-1",
		Type:  store.JobRender,
		Status: store.JobQueued,
		Metadata: &map[string]string{"test-key": "test-value"},
	}
	_, err = s.Jobs().Enqueue(ctx, job2)
	require.NoError(t, err)

	queued, err := s.Jobs().ListQueued(ctx)
	require.NoError(t, err)
	assert.Len(t, queued, 1)
	assert.Equal(t, "job-2", queued[0].ID)
	assert.NotNil(t, queued[0].Metadata)
	assert.Equal(t, "test-value", (*queued[0].Metadata)["test-key"])
}

func TestJobDeduplication(t *testing.T) {
	s := New()
	ctx := context.Background()

	// 1. Enqueue first job
	job1 := store.Job{
		ID:              "job-1",
		OrgID:           "org-1",
		Type:            store.JobExport,
		Status:          store.JobQueued,
		DeduplicationID: "dedup-1",
	}
	created1, isDup, err := s.Jobs().EnqueueWithDeduplication(ctx, job1)
	require.NoError(t, err)
	assert.False(t, isDup)
	assert.Equal(t, "job-1", created1.ID)

	// 2. Enqueue duplicate job (should return existing)
	job2 := store.Job{
		ID:              "job-2",
		OrgID:           "org-1",
		Type:            store.JobExport,
		Status:          store.JobQueued,
		DeduplicationID: "dedup-1",
	}
	created2, isDup, err := s.Jobs().EnqueueWithDeduplication(ctx, job2)
	require.NoError(t, err)
	assert.True(t, isDup)
	assert.Equal(t, "job-1", created2.ID) // Returns original ID

	// 3. Complete the job
	created1.Status = store.JobDone
	_, err = s.Jobs().Update(ctx, created1)
	require.NoError(t, err)

	// 4. Enqueue duplicate job again (should return completed)
	created3, isDup, err := s.Jobs().EnqueueWithDeduplication(ctx, job2)
	require.NoError(t, err)
	assert.True(t, isDup)
	assert.Equal(t, "job-1", created3.ID)
	assert.Equal(t, store.JobDone, created3.Status)

	// 5. Fail the job
	created1.Status = store.JobFailed
	_, err = s.Jobs().Update(ctx, created1)
	require.NoError(t, err)

	// 6. Enqueue duplicate job (should create NEW one because previous failed)
	job3 := store.Job{
		ID:              "job-3",
		OrgID:           "org-1",
		Type:            store.JobExport,
		Status:          store.JobQueued,
		DeduplicationID: "dedup-1",
	}
	created4, isDup, err := s.Jobs().EnqueueWithDeduplication(ctx, job3)
	require.NoError(t, err)
	assert.False(t, isDup)
	assert.Equal(t, "job-3", created4.ID)
}
