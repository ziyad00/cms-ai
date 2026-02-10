package test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ziyad/cms-ai/server/internal/ai"
	"github.com/ziyad/cms-ai/server/internal/assets"
	"github.com/ziyad/cms-ai/server/internal/store"
	"github.com/ziyad/cms-ai/server/internal/store/memory"
	"github.com/ziyad/cms-ai/server/internal/worker"
)

// TestRegression_WorkerOutputRef_IsAssetID verifies fix for bug where worker set job.OutputRef to file path
// The fix ensures worker sets job.OutputRef to the Asset ID (UUID), and the file path is stored in the Asset record
func TestRegression_WorkerOutputRef_IsAssetID(t *testing.T) {
	ctx := context.Background()
	memStore := memory.New()
	renderer := assets.NewGoPPTXRenderer()
	storage, _ := assets.NewLocalStorage(assets.StorageConfig{Type: "local"})

	worker := worker.New(memStore, renderer, storage, ai.NewAIService(memStore))

	// Create a template version
	templateVersion := store.TemplateVersion{
		ID:        "reg-version-1",
		Template:  "reg-template-1",
		OrgID:     "reg-org",
		VersionNo: 1,
		SpecJSON: map[string]interface{}{
			"layouts": []map[string]interface{}{
				{
					"name": "default",
					"placeholders": []map[string]interface{}{
						{
							"id":   "title",
							"type": "text",
						},
					},
				},
			},
		},
		CreatedAt: time.Now(),
	}
	_, err := memStore.Templates().CreateVersion(ctx, templateVersion)
	require.NoError(t, err)

	// Create a render job
	job := store.Job{
		ID:        "reg-job-1",
		OrgID:     "reg-org",
		Type:      store.JobRender,
		Status:    store.JobQueued,
		InputRef:  "reg-version-1",
		CreatedAt: time.Now(),
	}
	_, err = memStore.Jobs().Enqueue(ctx, job)
	require.NoError(t, err)

	// Process job
	worker.ProcessJobs()
	time.Sleep(50 * time.Millisecond)

	// Verify job output
	processedJob, found, err := memStore.Jobs().Get(ctx, "reg-org", "reg-job-1")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, store.JobDone, processedJob.Status)

	outputRef := processedJob.OutputRef
	t.Logf("Job OutputRef: %s", outputRef)

	// Assertions for FIX
	assert.NotEmpty(t, outputRef, "OutputRef should not be empty")
	assert.NotContains(t, outputRef, "/", "OutputRef should NOT be a path (should not contain slashes)")
	assert.NotContains(t, outputRef, ".pptx", "OutputRef should NOT contain file extension (should be UUID)")
	assert.False(t, strings.HasPrefix(outputRef, "assets/"), "OutputRef should NOT start with assets/ path")

	// Verify it links to a valid asset
	asset, found, err := memStore.Assets().Get(ctx, "reg-org", outputRef)
	require.NoError(t, err)
	require.True(t, found, "OutputRef should be a valid Asset ID")
	assert.Contains(t, asset.Path, ".pptx", "Linked Asset record should have the file path")

	t.Log("✅ REGRESSION TEST PASSED: Worker correctly sets OutputRef to Asset ID")
}