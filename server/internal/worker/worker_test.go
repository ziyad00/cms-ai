package worker

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ziyad/cms-ai/server/internal/ai"
	"github.com/ziyad/cms-ai/server/internal/assets"
	"github.com/ziyad/cms-ai/server/internal/store"
	"github.com/ziyad/cms-ai/server/internal/store/memory"
)

func TestWorker_ProcessJobs(t *testing.T) {
	// Setup test dependencies
	memStore := memory.New()
	renderer := assets.NewGoPPTXRenderer()
	storage, _ := assets.NewLocalStorage(assets.StorageConfig{Type: "local"})

	worker := New(memStore, renderer, storage, ai.NewAIService(memStore))

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
	storage, _ := assets.NewLocalStorage(assets.StorageConfig{Type: "local"})

	worker := New(memStore, renderer, storage, ai.NewAIService(memStore))

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
	storage, _ := assets.NewLocalStorage(assets.StorageConfig{Type: "local"})

	worker := New(memStore, renderer, storage, ai.NewAIService(memStore))

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
	storage, _ := assets.NewLocalStorage(assets.StorageConfig{Type: "local"})
	worker := New(memStore, renderer, storage, ai.NewAIService(memStore))

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
	storage, _ := assets.NewLocalStorage(assets.StorageConfig{Type: "local"})

	worker := New(memStore, renderer, storage, ai.NewAIService(memStore))

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
	storage, _ := assets.NewLocalStorage(assets.StorageConfig{Type: "local"})

	_ = New(memStore, renderer, storage, ai.NewAIService(memStore))

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

func TestWorker_GenerateJob_NilMetadata_ReturnsError(t *testing.T) {
	memStore := memory.New()
	renderer := assets.NewGoPPTXRenderer()
	storage, _ := assets.NewLocalStorage(assets.StorageConfig{Type: "local"})
	worker := New(memStore, renderer, storage, ai.NewAIService(memStore))

	ctx := context.Background()

	// Create a template so the job has a valid input ref
	tmpl := store.Template{ID: "tpl-gen-nil", OrgID: "org-1", Name: "Test", Status: store.TemplateDraft}
	_, err := memStore.Templates().CreateTemplate(ctx, tmpl)
	require.NoError(t, err)

	// Enqueue a generate job with NO metadata
	job := store.Job{
		ID:        "job-gen-nil-meta",
		OrgID:     "org-1",
		Type:      store.JobGenerate,
		Status:    store.JobQueued,
		InputRef:  "tpl-gen-nil",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err = memStore.Jobs().Enqueue(ctx, job)
	require.NoError(t, err)

	worker.processJobs()
	time.Sleep(100 * time.Millisecond)

	got, found, err := memStore.Jobs().Get(ctx, "org-1", job.ID)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, store.JobDeadLetter, got.Status, "generate job with nil metadata should dead-letter")
	assert.Contains(t, got.Error, "missing job metadata")
}

func TestWorker_BindJob_NilMetadata_ReturnsError(t *testing.T) {
	memStore := memory.New()
	renderer := assets.NewGoPPTXRenderer()
	storage, _ := assets.NewLocalStorage(assets.StorageConfig{Type: "local"})
	worker := New(memStore, renderer, storage, ai.NewAIService(memStore))

	ctx := context.Background()

	// Create a deck so the job has a valid input ref
	deck := store.Deck{ID: "deck-bind-nil", OrgID: "org-1", Name: "Test Deck"}
	_, err := memStore.Decks().CreateDeck(ctx, deck)
	require.NoError(t, err)

	// Enqueue a bind job with NO metadata
	job := store.Job{
		ID:        "job-bind-nil-meta",
		OrgID:     "org-1",
		Type:      store.JobBind,
		Status:    store.JobQueued,
		InputRef:  "deck-bind-nil",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err = memStore.Jobs().Enqueue(ctx, job)
	require.NoError(t, err)

	worker.processJobs()
	time.Sleep(100 * time.Millisecond)

	got, found, err := memStore.Jobs().Get(ctx, "org-1", job.ID)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, store.JobDeadLetter, got.Status, "bind job with nil metadata should dead-letter")
	assert.Contains(t, got.Error, "missing job metadata")
}

func TestWorker_ExportJob_WithMetadata_Roundtrips(t *testing.T) {
	memStore := memory.New()
	renderer := assets.NewGoPPTXRenderer()
	storage, _ := assets.NewLocalStorage(assets.StorageConfig{Type: "local"})
	worker := New(memStore, renderer, storage, ai.NewAIService(memStore))

	ctx := context.Background()
	orgID := "org-meta"

	// Create template version
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

	tv := store.TemplateVersion{
		ID:        "ver-meta-export",
		Template:  "tpl-meta-export",
		OrgID:     orgID,
		VersionNo: 1,
		SpecJSON:  templateSpec,
		CreatedBy: "user-1",
		CreatedAt: time.Now(),
	}
	_, err := memStore.Templates().CreateVersion(ctx, tv)
	require.NoError(t, err)

	// Enqueue export job with metadata — mimics production (router_v1.go:1030)
	metadata := store.JSONMap{
		"versionNo": "1",
		"filename":  "deck-export-v1-20260212.pptx",
	}
	job := store.Job{
		ID:        "job-export-meta",
		OrgID:     orgID,
		Type:      store.JobExport,
		Status:    store.JobQueued,
		InputRef:  "ver-meta-export",
		Metadata:  &metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err = memStore.Jobs().Enqueue(ctx, job)
	require.NoError(t, err)

	worker.processJobs()
	time.Sleep(100 * time.Millisecond)

	got, found, err := memStore.Jobs().Get(ctx, orgID, job.ID)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, store.JobDone, got.Status, "export job with metadata should succeed")
	assert.NotEmpty(t, got.OutputRef)
	// Metadata should be preserved after processing
	require.NotNil(t, got.Metadata, "metadata must survive job update")
	assert.Equal(t, "deck-export-v1-20260212.pptx", (*got.Metadata)["filename"])
	assert.Equal(t, "1", (*got.Metadata)["versionNo"])
}

func TestWorker_RenderJob_WithMetadata_Preserved(t *testing.T) {
	memStore := memory.New()
	renderer := assets.NewGoPPTXRenderer()
	storage, _ := assets.NewLocalStorage(assets.StorageConfig{Type: "local"})
	worker := New(memStore, renderer, storage, ai.NewAIService(memStore))

	ctx := context.Background()

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

	tv := store.TemplateVersion{
		ID:        "ver-render-meta",
		Template:  "tpl-render-meta",
		OrgID:     "org-1",
		VersionNo: 1,
		SpecJSON:  templateSpec,
		CreatedBy: "user-1",
		CreatedAt: time.Now(),
	}
	_, err := memStore.Templates().CreateVersion(ctx, tv)
	require.NoError(t, err)

	// Render job with metadata attached
	metadata := store.JSONMap{"source": "test", "requestId": "req-123"}
	job := store.Job{
		ID:        "job-render-meta",
		OrgID:     "org-1",
		Type:      store.JobRender,
		Status:    store.JobQueued,
		InputRef:  "ver-render-meta",
		Metadata:  &metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err = memStore.Jobs().Enqueue(ctx, job)
	require.NoError(t, err)

	worker.processJobs()
	time.Sleep(100 * time.Millisecond)

	got, found, err := memStore.Jobs().Get(ctx, "org-1", job.ID)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, store.JobDone, got.Status)
	require.NotNil(t, got.Metadata, "metadata must be preserved through job lifecycle")
	assert.Equal(t, "req-123", (*got.Metadata)["requestId"])
}

// TDD: processBindJob must handle string SpecJSON from pgx without double-encoding.
func TestWorker_BindJob_StringSpecJSON_NotDoubleEncoded(t *testing.T) {
	memStore := memory.New()
	renderer := assets.NewGoPPTXRenderer()
	storage, _ := assets.NewLocalStorage(assets.StorageConfig{Type: "local"})
	w := New(memStore, renderer, storage, ai.NewAIService(memStore))

	ctx := context.Background()
	orgID := "org-bind-str"

	// String spec — simulates what pgx returns when reading jsonb columns.
	specString := `{"layouts":[{"name":"title-slide","placeholders":[{"id":"title","type":"text","content":"Hello","geometry":{"x":0.1,"y":0.1,"w":0.8,"h":0.2}}]}]}`

	tv := store.TemplateVersion{
		ID:        "tv-bind-str",
		Template:  "tpl-bind-str",
		OrgID:     orgID,
		VersionNo: 1,
		SpecJSON:  specString, // Go string, not map — this is what pgx gives us
		CreatedBy: "user-1",
		CreatedAt: time.Now(),
	}
	_, err := memStore.Templates().CreateVersion(ctx, tv)
	require.NoError(t, err)

	deck := store.Deck{ID: "deck-bind-str", OrgID: orgID, Name: "Test Deck"}
	_, err = memStore.Decks().CreateDeck(ctx, deck)
	require.NoError(t, err)

	metadata := store.JSONMap{
		"sourceTemplateVersionId": "tv-bind-str",
		"content":                 "Some test content for the deck",
		"userId":                  "user-1",
	}
	job := store.Job{
		ID:        "job-bind-str",
		OrgID:     orgID,
		Type:      store.JobBind,
		Status:    store.JobQueued,
		InputRef:  "deck-bind-str",
		Metadata:  &metadata,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err = memStore.Jobs().Enqueue(ctx, job)
	require.NoError(t, err)

	w.processJobs()
	time.Sleep(100 * time.Millisecond)

	got, found, err := memStore.Jobs().Get(ctx, orgID, job.ID)
	require.NoError(t, err)
	require.True(t, found)

	// Job may fail at AI call (no HuggingFace key), but must NOT fail with
	// "invalid template spec" — that would mean double-encoding.
	if got.Status != store.JobDone {
		assert.NotContains(t, got.Error, "invalid template spec",
			"string SpecJSON from pgx must not cause double-encoding; got error: %s", got.Error)
	}
}

// TDD: Render with string SpecJSON (pgx returns string for jsonb).
func TestWorker_RenderJob_StringSpecJSON_Works(t *testing.T) {
	memStore := memory.New()
	renderer := assets.NewGoPPTXRenderer()
	storage, _ := assets.NewLocalStorage(assets.StorageConfig{Type: "local"})
	w := New(memStore, renderer, storage, ai.NewAIService(memStore))

	ctx := context.Background()
	orgID := "org-render-str"

	specString := `{"layouts":[{"name":"title-slide","placeholders":[{"id":"title","type":"text","geometry":{"x":0.1,"y":0.1,"w":0.8,"h":0.2}}]}]}`

	tv := store.TemplateVersion{
		ID:        "tv-render-str",
		Template:  "tpl-render-str",
		OrgID:     orgID,
		VersionNo: 1,
		SpecJSON:  specString,
		CreatedBy: "user-1",
		CreatedAt: time.Now(),
	}
	_, err := memStore.Templates().CreateVersion(ctx, tv)
	require.NoError(t, err)

	job := store.Job{
		ID:        "job-render-str",
		OrgID:     orgID,
		Type:      store.JobRender,
		Status:    store.JobQueued,
		InputRef:  "tv-render-str",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err = memStore.Jobs().Enqueue(ctx, job)
	require.NoError(t, err)

	w.processJobs()
	time.Sleep(100 * time.Millisecond)

	got, found, err := memStore.Jobs().Get(ctx, orgID, job.ID)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, store.JobDone, got.Status, "render job with string SpecJSON must succeed; error: %s", got.Error)
	assert.NotEmpty(t, got.OutputRef)
}

// Unit test for anyToJSONBytes — must handle all types without double-encoding.
func TestAnyToJSONBytes(t *testing.T) {
	jsonStr := `{"layouts":[{"name":"title"}]}`

	tests := []struct {
		name  string
		input any
		want  string
	}{
		{"string from pgx", jsonStr, jsonStr},
		{"[]byte", []byte(jsonStr), jsonStr},
		{"json.RawMessage", json.RawMessage(jsonStr), jsonStr},
		{"map", map[string]string{"key": "val"}, `{"key":"val"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := anyToJSONBytes(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(got))
			if tt.want[0] == '{' {
				assert.Equal(t, byte('{'), got[0], "must not double-encode")
			}
		})
	}
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

// TDD RED: anyToJSONBytes must handle base64-encoded strings from pgx.
// When GORM writes []byte to jsonb, json.Marshal([]byte) base64-encodes it.
// pgx reads it back as a Go string containing base64. anyToJSONBytes must
// detect and decode base64 to get the raw JSON.
func TestAnyToJSONBytes_base64_from_pgx(t *testing.T) {
	originalJSON := `{"layouts":[{"name":"title","placeholders":[{"id":"t","type":"text"}]}]}`

	// Simulate GORM write: json.Marshal([]byte) → base64 JSON string
	base64Encoded, err := json.Marshal([]byte(originalJSON))
	require.NoError(t, err)

	// Simulate pgx read: JSON string → Go string (strips outer quotes)
	var pgxValue string
	err = json.Unmarshal(base64Encoded, &pgxValue)
	require.NoError(t, err)
	// pgxValue is now a base64 string like "eyJsYXlvdXRzIj..."

	result, err := anyToJSONBytes(pgxValue)
	require.NoError(t, err)
	assert.Equal(t, byte('{'), result[0],
		"must decode base64 to JSON object, got: %s", string(result[:min(50, len(result))]))
	assert.JSONEq(t, originalJSON, string(result))
}

// TDD RED: Worker must enforce a timeout on job processing.
// Without timeout, a hanging renderer keeps the job in "Running" forever.
func TestWorker_ProcessJob_RespectsContextTimeout(t *testing.T) {
	memStore := memory.New()
	storage, _ := assets.NewLocalStorage(assets.StorageConfig{Type: "local"})

	// Use a slow renderer that blocks until context is cancelled
	slowRenderer := &slowRenderer{}
	w := New(memStore, slowRenderer, storage, ai.NewAIService(memStore))
	w.JobTimeout = 2 * time.Second // short timeout for tests

	ctx := context.Background()
	orgID := "org-timeout"

	specJSON := `{"layouts":[{"name":"test","placeholders":[{"id":"t","type":"text","geometry":{"x":0.1,"y":0.1,"w":0.8,"h":0.2}}]}]}`

	// Create a deck version
	deck := store.Deck{ID: "deck-timeout", OrgID: orgID, Name: "Timeout Test"}
	_, err := memStore.Decks().CreateDeck(ctx, deck)
	require.NoError(t, err)

	dv := store.DeckVersion{
		ID: "dv-timeout", Deck: "deck-timeout", OrgID: orgID,
		VersionNo: 1, SpecJSON: specJSON, CreatedBy: "user-1",
		CreatedAt: time.Now(),
	}
	_, err = memStore.Decks().CreateDeckVersion(ctx, dv)
	require.NoError(t, err)

	metadata := store.JSONMap{"versionNo": "1", "filename": "test.pptx"}
	job := store.Job{
		ID: "job-timeout", OrgID: orgID, Type: store.JobExport,
		Status: store.JobQueued, InputRef: "dv-timeout",
		Metadata: &metadata, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	_, err = memStore.Jobs().Enqueue(ctx, job)
	require.NoError(t, err)

	// Process jobs — with timeout, the slow renderer should be cancelled
	start := time.Now()
	w.processJobs()
	elapsed := time.Since(start)

	// Must not hang forever — should fail within the worker's timeout (2s + small overhead)
	assert.Less(t, elapsed, 10*time.Second,
		"job must not hang; worker should enforce timeout")

	got, found, err := memStore.Jobs().Get(ctx, orgID, job.ID)
	require.NoError(t, err)
	require.True(t, found)

	// Job should be failed (not still Running)
	assert.NotEqual(t, store.JobRunning, got.Status,
		"timed-out job must not remain in Running state")
}

// slowRenderer blocks forever in RenderPPTXBytes (until context cancelled).
type slowRenderer struct{}

func (s *slowRenderer) RenderPPTX(ctx context.Context, spec interface{}, outPath string) error {
	<-ctx.Done()
	return ctx.Err()
}

func (s *slowRenderer) RenderPPTXBytes(ctx context.Context, spec interface{}) ([]byte, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func (s *slowRenderer) GenerateSlideThumbnails(ctx context.Context, spec interface{}) ([][]byte, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}
