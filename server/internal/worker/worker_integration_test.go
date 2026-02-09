package worker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ziyad/cms-ai/server/internal/assets"
	"github.com/ziyad/cms-ai/server/internal/store"
	"github.com/ziyad/cms-ai/server/internal/store/memory"
)

func TestWorker_ProcessesExportJobsEndToEnd(t *testing.T) {
	ctx := context.Background()
	memStore := memory.New()
	renderer := assets.NewGoPPTXRenderer()
	storage := assets.LocalStorage{}

	worker := New(memStore, renderer, &storage, ai.NewAIService(memStore))

	// Create a template version first
	templateSpec := map[string]interface{}{
		"tokens": map[string]interface{}{
			"colors": map[string]interface{}{
				"primary":    "#2563eb",
				"secondary":  "#dc2626",
				"background": "#ffffff",
				"text":       "#1f2937",
			},
		},
		"layouts": []map[string]interface{}{
			{
				"name": "title-slide",
				"placeholders": []map[string]interface{}{
					{
						"id":      "title",
						"type":    "text",
						"content": "Test Export Presentation",
						"geometry": map[string]interface{}{
							"x": 1.0,
							"y": 1.5,
							"w": 8.0,
							"h": 1.5,
						},
					},
					{
						"id":      "subtitle",
						"type":    "text",
						"content": "Verifying worker export job processing",
						"geometry": map[string]interface{}{
							"x": 1.0,
							"y": 3.0,
							"w": 8.0,
							"h": 1.0,
						},
					},
				},
			},
			{
				"name": "content-slide",
				"placeholders": []map[string]interface{}{
					{
						"id":      "title",
						"type":    "text",
						"content": "Key Features",
						"geometry": map[string]interface{}{
							"x": 1.0,
							"y": 1.0,
							"w": 8.0,
							"h": 1.0,
						},
					},
					{
						"id":      "content",
						"type":    "text",
						"content": "• Worker picks up export jobs from queue\n• Executes Python PPTX renderer with AI\n• Updates job status during processing\n• Handles errors and retries appropriately",
						"geometry": map[string]interface{}{
							"x": 1.0,
							"y": 2.5,
							"w": 8.0,
							"h": 4.0,
						},
					},
				},
			},
		},
	}

	orgID := "test-org-worker"
	templateVersion := store.TemplateVersion{
		ID:        "worker-test-version",
		Template:  "worker-test-template",
		OrgID:     orgID,
		VersionNo: 1,
		SpecJSON:  templateSpec,
		CreatedBy: "worker-test-user",
		CreatedAt: time.Now(),
	}

	_, err := memStore.Templates().CreateVersion(ctx, templateVersion)
	require.NoError(t, err)

	// Test Case 1: Export Job Processing
	t.Run("ExportJobProcessing", func(t *testing.T) {
		exportJob := store.Job{
			ID:        "export-job-test",
			OrgID:     orgID,
			Type:      store.JobExport,
			Status:    store.JobQueued,
			InputRef:  templateVersion.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Enqueue the export job
		_, err := memStore.Jobs().Enqueue(ctx, exportJob)
		require.NoError(t, err)

		// Verify job is queued
		queuedJobs, err := memStore.Jobs().ListQueued(ctx)
		require.NoError(t, err)
		require.Len(t, queuedJobs, 1)
		assert.Equal(t, store.JobQueued, queuedJobs[0].Status)

		// Process jobs with worker
		worker.ProcessJobs()
		time.Sleep(100 * time.Millisecond) // Allow processing to complete

		// Verify job is completed
		completedJob, found, err := memStore.Jobs().Get(ctx, orgID, exportJob.ID)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, store.JobDone, completedJob.Status)
		assert.NotEmpty(t, completedJob.OutputRef)
		// OutputRef is now an Asset ID (UUID), so it should NOT contain .pptx
		assert.NotContains(t, completedJob.OutputRef, ".pptx")

		// Verify asset exists in store
		asset, found, err := memStore.Assets().Get(ctx, orgID, completedJob.OutputRef)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, store.AssetPPTX, asset.Type)
		assert.Contains(t, asset.Path, ".pptx")

		// Verify no more queued jobs
		queuedJobs, err = memStore.Jobs().ListQueued(ctx)
		require.NoError(t, err)
		assert.Len(t, queuedJobs, 0)
	})

	// Test Case 2: Render Job Processing
	t.Run("RenderJobProcessing", func(t *testing.T) {
		renderJob := store.Job{
			ID:        "render-job-test",
			OrgID:     orgID,
			Type:      store.JobRender,
			Status:    store.JobQueued,
			InputRef:  templateVersion.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Enqueue the render job
		_, err := memStore.Jobs().Enqueue(ctx, renderJob)
		require.NoError(t, err)

		// Process jobs with worker
		worker.ProcessJobs()
		time.Sleep(100 * time.Millisecond)

		// Verify job is completed
		completedJob, found, err := memStore.Jobs().Get(ctx, orgID, renderJob.ID)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, store.JobDone, completedJob.Status)
		assert.NotEmpty(t, completedJob.OutputRef)

		// Verify asset exists in store
		_, found, err = memStore.Assets().Get(ctx, orgID, completedJob.OutputRef)
		require.NoError(t, err)
		require.True(t, found)
	})

	// Test Case 3: Preview Job Processing
	t.Run("PreviewJobProcessing", func(t *testing.T) {
		previewJob := store.Job{
			ID:        "preview-job-test",
			OrgID:     orgID,
			Type:      store.JobPreview,
			Status:    store.JobQueued,
			InputRef:  templateVersion.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Enqueue the preview job
		_, err := memStore.Jobs().Enqueue(ctx, previewJob)
		require.NoError(t, err)

		// Process jobs with worker
		worker.ProcessJobs()
		time.Sleep(100 * time.Millisecond)

		// Verify job is completed
		completedJob, found, err := memStore.Jobs().Get(ctx, orgID, previewJob.ID)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, store.JobDone, completedJob.Status)
		assert.NotEmpty(t, completedJob.OutputRef)
		assert.Contains(t, completedJob.OutputRef, ".preview.png")
	})
}

func TestWorker_ErrorHandlingAndRetries(t *testing.T) {
	ctx := context.Background()
	memStore := memory.New()
	failingRenderer := &failingRenderer{}
	storage := assets.LocalStorage{}

	worker := New(memStore, failingRenderer, &storage, ai.NewAIService(memStore))

	// Create template version
	templateSpec := map[string]interface{}{
		"layouts": []map[string]interface{}{
			{
				"name": "test-slide",
				"placeholders": []map[string]interface{}{
					{
						"id":      "title",
						"type":    "text",
						"content": "Test Slide",
					},
				},
			},
		},
	}

	orgID := "test-org-retry"
	templateVersion := store.TemplateVersion{
		ID:        "retry-test-version",
		Template:  "retry-test-template",
		OrgID:     orgID,
		VersionNo: 1,
		SpecJSON:  templateSpec,
		CreatedBy: "retry-test-user",
		CreatedAt: time.Now(),
	}

	_, err := memStore.Templates().CreateVersion(ctx, templateVersion)
	require.NoError(t, err)

	t.Run("JobRetryOnFailure", func(t *testing.T) {
		retryJob := store.Job{
			ID:         "retry-job-test",
			OrgID:      orgID,
			Type:       store.JobRender,
			Status:     store.JobQueued,
			InputRef:   templateVersion.ID,
			MaxRetries: 2,
			RetryCount: 0,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		// Enqueue the job that will fail
		_, err := memStore.Jobs().Enqueue(ctx, retryJob)
		require.NoError(t, err)

		// First attempt - should retry
		worker.ProcessJobs()
		time.Sleep(50 * time.Millisecond)

		// Check job is in retry state
		job, found, err := memStore.Jobs().Get(ctx, orgID, retryJob.ID)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, store.JobRetry, job.Status)
		assert.Equal(t, 1, job.RetryCount)
		assert.Contains(t, job.Error, "simulated renderer failure")

		// Advance retry time to make it ready for retry
		pastTime := time.Now().UTC().Add(-10 * time.Second)
		job.LastRetryAt = &pastTime
		_, err = memStore.Jobs().Update(ctx, job)
		require.NoError(t, err)

		// Second attempt - should retry again
		worker.ProcessJobs()
		time.Sleep(50 * time.Millisecond)

		// Check job is still in retry state but retry count increased
		job, found, err = memStore.Jobs().Get(ctx, orgID, retryJob.ID)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, store.JobRetry, job.Status)
		assert.Equal(t, 2, job.RetryCount)

		// Advance retry time again
		job.LastRetryAt = &pastTime
		_, err = memStore.Jobs().Update(ctx, job)
		require.NoError(t, err)

		// Third attempt - should move to dead letter (max retries = 2)
		worker.ProcessJobs()
		time.Sleep(50 * time.Millisecond)

		// Check job is in dead letter queue
		job, found, err = memStore.Jobs().Get(ctx, orgID, retryJob.ID)
		require.NoError(t, err)
		require.True(t, found)
		assert.Equal(t, store.JobDeadLetter, job.Status)
		assert.Equal(t, 2, job.RetryCount)
		assert.Contains(t, job.Error, "Final retry")
	})
}

func TestWorker_WorkerServiceRunning(t *testing.T) {
	ctx := context.Background()
	memStore := memory.New()
	renderer := assets.NewGoPPTXRenderer()
	storage := assets.LocalStorage{}

	worker := New(memStore, renderer, &storage, ai.NewAIService(memStore))

	t.Run("WorkerPollingWithoutJobs", func(t *testing.T) {
		// Test that processJobs runs without error when no jobs are available
		worker.ProcessJobs()

		// Verify no jobs in any queue
		queuedJobs, err := memStore.Jobs().ListQueued(ctx)
		require.NoError(t, err)
		assert.Len(t, queuedJobs, 0)

		retryJobs, err := memStore.Jobs().ListRetry(ctx)
		require.NoError(t, err)
		assert.Len(t, retryJobs, 0)
	})

	t.Run("WorkerStartStop", func(t *testing.T) {
		// Test worker can be started and stopped cleanly
		worker.Start()

		// Let it run for a brief moment
		time.Sleep(100 * time.Millisecond)

		// Stop the worker
		worker.Stop()

		// Verify it stopped cleanly (no panic or hanging)
	})
}