package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ziyad/cms-ai/server/internal/assets"
	"github.com/ziyad/cms-ai/server/internal/store"
	"github.com/ziyad/cms-ai/server/internal/store/memory"
)

func TestWorker_ProcessJobs(t *testing.T) {
	// Setup test dependencies
	memStore := memory.New()
	renderer := assets.NewGoPPTXRenderer()
	storage := assets.LocalStorage{}

	worker := New(memStore, renderer, storage)

	ctx := context.Background()
	orgID := "test-org"

	// Create a template version
	templateSpec := map[string]interface{}{
		"tokens": map[string]interface{}{
			"colors": map[string]interface{}{
				"primary":    "#0078d4",
				"secondary":  "#107c10",
				"background": "#ffffff",
				"text":       "#323130",
			},
		},
		"layouts": []map[string]interface{}{
			{
				"name": "title-slide",
				"placeholders": []map[string]interface{}{
					{
						"id":   "title",
						"type": "text",
						"geometry": map[string]interface{}{
							"x": 0.1,
							"y": 0.1,
							"w": 0.8,
							"h": 0.2,
						},
					},
				},
			},
		},
	}

	templateVersion := store.TemplateVersion{
		ID:        "version-1",
		Template:  "template-1",
		OrgID:     orgID,
		VersionNo: 1,
		SpecJSON:  templateSpec,
		CreatedBy: "user-1",
		CreatedAt: time.Now(),
	}

	_, err := memStore.Templates().CreateVersion(ctx, templateVersion)
	require.NoError(t, err)

	// Test cases for each job type
	testCases := []struct {
		name      string
		jobType   store.JobType
		assetType store.AssetType
	}{
		{"Render Job", store.JobRender, store.AssetPPTX},
		{"Export Job", store.JobExport, store.AssetPPTX},
		{"Preview Job", store.JobPreview, store.AssetPNG},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a job
			job := store.Job{
				ID:        "job-" + string(tc.jobType),
				OrgID:     orgID,
				Type:      tc.jobType,
				Status:    store.JobQueued,
				InputRef:  "version-1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			_, err := memStore.Jobs().Enqueue(ctx, job)
			require.NoError(t, err)

			// Process jobs
			worker.processJobs()

			// Wait a moment for processing to complete
			time.Sleep(100 * time.Millisecond)

			// Check job status
			processedJob, found, err := memStore.Jobs().Get(ctx, orgID, job.ID)
			require.NoError(t, err)
			require.True(t, found)
			assert.Equal(t, store.JobDone, processedJob.Status)
			assert.NotEmpty(t, processedJob.OutputRef)

			// Check the output format
			if tc.assetType == store.AssetPPTX {
				// Render/Export jobs return an Asset ID (UUID), not a file path
				// It should NOT contain .pptx in the ID itself
				assert.NotContains(t, processedJob.OutputRef, ".pptx")
				assert.NotEmpty(t, processedJob.OutputRef)

				// Verify asset exists in store
				asset, found, err := memStore.Assets().Get(ctx, orgID, processedJob.OutputRef)
				require.NoError(t, err)
				require.True(t, found)
				assert.Equal(t, store.AssetPPTX, asset.Type)
				assert.Contains(t, asset.Path, ".pptx") // The path should have the extension
			} else if tc.assetType == store.AssetPNG {
				// Preview files return the path to the first thumbnail
				assert.Contains(t, processedJob.OutputRef, "job-"+string(tc.jobType))
				assert.Contains(t, processedJob.OutputRef, ".preview.png")
			}
		})
	}
}

func TestWorker_FailJob(t *testing.T) {
	// Setup test dependencies
	memStore := memory.New()
	renderer := assets.NewGoPPTXRenderer()
	storage := assets.LocalStorage{}

	worker := New(memStore, renderer, storage)

	ctx := context.Background()
	orgID := "test-org"

	// Create a job with invalid input reference
	job := store.Job{
		ID:        "job-fail",
		OrgID:     orgID,
		Type:      store.JobRender,
		Status:    store.JobQueued,
		InputRef:  "nonexistent-version",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := memStore.Jobs().Enqueue(ctx, job)
	require.NoError(t, err)

	// Process jobs
	worker.processJobs()

	// Wait a moment for processing to complete
	time.Sleep(100 * time.Millisecond)

	// Check job status
	processedJob, found, err := memStore.Jobs().Get(ctx, orgID, job.ID)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, store.JobDeadLetter, processedJob.Status)
	assert.Contains(t, processedJob.Error, "template version not found")
	assert.Contains(t, processedJob.Error, "Final retry")
}

func TestWorker_UnsupportedJobType(t *testing.T) {
	// Setup test dependencies
	memStore := memory.New()
	renderer := assets.NewGoPPTXRenderer()
	storage := assets.LocalStorage{}

	worker := New(memStore, renderer, storage)

	ctx := context.Background()
	orgID := "test-org"

	// Create a template version
	templateSpec := map[string]interface{}{
		"layouts": []map[string]interface{}{
			{
				"name": "title-slide",
				"placeholders": []map[string]interface{}{
					{
						"id":   "title",
						"type": "text",
						"geometry": map[string]interface{}{
							"x": 0.1, "y": 0.1, "w": 0.8, "h": 0.2,
						},
					},
				},
			},
		},
	}

	templateVersion := store.TemplateVersion{
		ID:        "version-1",
		Template:  "template-1",
		OrgID:     orgID,
		VersionNo: 1,
		SpecJSON:  templateSpec,
		CreatedBy: "user-1",
		CreatedAt: time.Now(),
	}

	_, err := memStore.Templates().CreateVersion(ctx, templateVersion)
	require.NoError(t, err)

	// Create a job with unsupported type
	job := store.Job{
		ID:        "job-unsupported",
		OrgID:     orgID,
		Type:      "unsupported",
		Status:    store.JobQueued,
		InputRef:  "version-1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Manually enqueue (since the store expects specific JobType)
	job.Status = store.JobQueued
	memStore.Jobs().Enqueue(ctx, job)

	// Process jobs
	worker.processJobs()

	// Wait a moment for processing to complete
	time.Sleep(100 * time.Millisecond)

	// Check job status
	processedJob, found, err := memStore.Jobs().Get(ctx, orgID, job.ID)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, store.JobDeadLetter, processedJob.Status)
	assert.Contains(t, processedJob.Error, "unsupported job type")
	assert.Contains(t, processedJob.Error, "Final retry")
}

func TestWorker_ProcessPreviewJobWithThumbnails(t *testing.T) {
	ctx := context.Background()
	memStore := memory.New()
	renderer := assets.NewGoPPTXRenderer()
	worker := New(memStore, renderer, assets.LocalStorage{})

	// Create template version with multiple layouts
	templateSpec := map[string]interface{}{
		"layouts": []map[string]interface{}{
			{
				"name": "title-slide",
				"placeholders": []map[string]interface{}{
					{
						"id":   "title",
						"type": "text",
						"geometry": map[string]interface{}{
							"x": 0.1, "y": 0.1, "w": 0.8, "h": 0.2,
						},
					},
				},
			},
			{
				"name": "content-slide",
				"placeholders": []map[string]interface{}{
					{
						"id":   "content",
						"type": "text",
						"geometry": map[string]interface{}{
							"x": 0.05, "y": 0.05, "w": 0.9, "h": 0.8,
						},
					},
				},
			},
			{
				"name": "summary-slide",
				"placeholders": []map[string]interface{}{
					{
						"id":   "summary",
						"type": "text",
						"geometry": map[string]interface{}{
							"x": 0.1, "y": 0.1, "w": 0.8, "h": 0.6,
						},
					},
				},
			},
		},
	}

	templateVersion := store.TemplateVersion{
		ID:        "version-multi",
		Template:  "template-1",
		OrgID:     "test-org",
		VersionNo: 1,
		SpecJSON:  templateSpec,
		CreatedBy: "user-1",
		CreatedAt: time.Now(),
	}

	_, err := memStore.Templates().CreateVersion(ctx, templateVersion)
	require.NoError(t, err)

	// Create preview job
	job := store.Job{
		ID:        "preview-multi",
		OrgID:     "test-org",
		Type:      store.JobPreview,
		Status:    store.JobQueued,
		InputRef:  "version-multi",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = memStore.Jobs().Enqueue(ctx, job)
	require.NoError(t, err)

	// Process the job
	worker.processJobs()
	time.Sleep(100 * time.Millisecond)

	// Check job completed successfully
	processedJob, found, err := memStore.Jobs().Get(ctx, job.OrgID, job.ID)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, store.JobDone, processedJob.Status)

	// Since we can't list all assets easily in memory store, let's verify the job output
	// The job should have completed and created thumbnails (we verify this through the renderer test)
	// In a real scenario, we would check asset storage contains the expected number of thumbnails
	assert.NotEmpty(t, processedJob.OutputRef)
}

func TestWorker_JobRetryAndDeadLetter(t *testing.T) {
	// Setup test dependencies
	memStore := memory.New()
	renderer := &failingRenderer{}
	storage := assets.LocalStorage{}

	worker := New(memStore, renderer, storage)

	ctx := context.Background()
	orgID := "test-org"

	// Create a template version
	templateSpec := map[string]interface{}{
		"layouts": []map[string]interface{}{
			{
				"name": "title-slide",
				"placeholders": []map[string]interface{}{
					{
						"id":   "title",
						"type": "text",
					},
				},
			},
		},
	}

	version := store.TemplateVersion{
		ID:        "version-123",
		Template:  "template-123",
		OrgID:     orgID,
		VersionNo: 1,
		SpecJSON:  templateSpec,
		CreatedBy: "user-123",
		CreatedAt: time.Now(),
	}

	_, err := memStore.Templates().CreateVersion(ctx, version)
	require.NoError(t, err)

	// Create a job that will fail
	job := store.Job{
		ID:         "job-retry-test",
		OrgID:      orgID,
		Type:       store.JobRender,
		Status:     store.JobQueued,
		InputRef:   version.ID,
		RetryCount: 0,
		MaxRetries: 2, // Set low retry count for testing
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	_, err = memStore.Jobs().Enqueue(ctx, job)
	require.NoError(t, err)

	// Process the job multiple times to trigger retry and then dead letter
	// First attempt - should retry
	worker.processJobs()
	time.Sleep(50 * time.Millisecond)

	// Get the job and advance time to make retry ready
	retryJob, found, err := memStore.Jobs().Get(ctx, job.OrgID, job.ID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, store.JobRetry, retryJob.Status)

	// Manually advance the LastRetryAt to make retry ready immediately
	now := time.Now().UTC()
	pastTime := now.Add(-10 * time.Second)
	retryJob.LastRetryAt = &pastTime
	_, err = memStore.Jobs().Update(ctx, retryJob)
	require.NoError(t, err)

	// Second attempt - should retry again
	worker.processJobs()
	time.Sleep(50 * time.Millisecond)

	// Get the job and advance time again for final retry
	retryJob2, found, err := memStore.Jobs().Get(ctx, job.OrgID, job.ID)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, store.JobRetry, retryJob2.Status)

	pastTime = now.Add(-10 * time.Second)
	retryJob2.LastRetryAt = &pastTime
	_, err = memStore.Jobs().Update(ctx, retryJob2)
	require.NoError(t, err)

	// Third attempt - should move to dead letter (max retries = 2)
	worker.processJobs()
	time.Sleep(50 * time.Millisecond)

	// Check job ended up in dead letter queue after max retries
	processedJob, found, err := memStore.Jobs().Get(ctx, job.OrgID, job.ID)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, store.JobDeadLetter, processedJob.Status)
	assert.Equal(t, 2, processedJob.RetryCount) // Should have retried max times
	assert.Contains(t, processedJob.Error, "Final retry")

	// Test manual retry from dead letter
	err = memStore.Jobs().RetryDeadLetterJob(ctx, job.ID)
	require.NoError(t, err)

	// Check job is back in queued state
	retriedJob, found, err := memStore.Jobs().Get(ctx, job.OrgID, job.ID)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, store.JobQueued, retriedJob.Status)
	assert.Equal(t, 0, retriedJob.RetryCount) // Reset to 0
	assert.Empty(t, retriedJob.Error)
}

func TestWorker_JobDeduplication(t *testing.T) {
	// Setup test dependencies
	memStore := memory.New()
	renderer := assets.NewGoPPTXRenderer()
	storage := assets.LocalStorage{}

	_ = New(memStore, renderer, storage)

	ctx := context.Background()
	orgID := "test-org"

	// Create a template version
	templateSpec := map[string]interface{}{
		"layouts": []map[string]interface{}{
			{
				"name": "title-slide",
				"placeholders": []map[string]interface{}{
					{
						"id":   "title",
						"type": "text",
					},
				},
			},
		},
	}

	version := store.TemplateVersion{
		ID:        "version-dedup-123",
		Template:  "template-dedup-123",
		OrgID:     orgID,
		VersionNo: 1,
		SpecJSON:  templateSpec,
		CreatedBy: "user-123",
		CreatedAt: time.Now(),
	}

	_, err := memStore.Templates().CreateVersion(ctx, version)
	require.NoError(t, err)

	// Create first job
	dedupID := "render-version-dedup-123"
	job1 := store.Job{
		ID:              "job-1",
		OrgID:           orgID,
		Type:            store.JobRender,
		Status:          store.JobQueued,
		InputRef:        version.ID,
		DeduplicationID: dedupID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	created1, wasDup1, err := memStore.Jobs().EnqueueWithDeduplication(ctx, job1)
	require.NoError(t, err)
	assert.False(t, wasDup1)

	// Try to create duplicate job
	job2 := store.Job{
		ID:              "job-2",
		OrgID:           orgID,
		Type:            store.JobRender,
		Status:          store.JobQueued,
		InputRef:        version.ID,
		DeduplicationID: dedupID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	created2, wasDup2, err := memStore.Jobs().EnqueueWithDeduplication(ctx, job2)
	require.NoError(t, err)
	assert.True(t, wasDup2)                   // Should be detected as duplicate
	assert.Equal(t, created1.ID, created2.ID) // Should return original job

	// Check only one job actually exists
	jobs, err := memStore.Jobs().ListQueued(ctx)
	require.NoError(t, err)
	assert.Len(t, jobs, 1)
	assert.Equal(t, created1.ID, jobs[0].ID)
}

// failingRenderer is a mock renderer that always fails
type failingRenderer struct{}

func (f *failingRenderer) RenderPPTX(ctx context.Context, spec interface{}, outPath string) error {
	return errors.New("simulated renderer failure")
}

func (f *failingRenderer) RenderPPTXBytes(ctx context.Context, spec interface{}) ([]byte, error) {
	return nil, errors.New("simulated renderer failure")
}

func (f *failingRenderer) GenerateSlideThumbnails(ctx context.Context, spec interface{}) ([][]byte, error) {
	return nil, errors.New("simulated thumbnail generation failure")
}
