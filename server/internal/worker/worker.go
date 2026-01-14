package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ziyad/cms-ai/server/internal/assets"
	"github.com/ziyad/cms-ai/server/internal/queue"
	"github.com/ziyad/cms-ai/server/internal/store"
)

type Worker struct {
	store    store.Store
	renderer assets.Renderer
	storage  assets.Storage
	stop     chan struct{}
	wg       sync.WaitGroup
}

func New(store store.Store, renderer assets.Renderer, storage assets.Storage) *Worker {
	return &Worker{
		store:    store,
		renderer: renderer,
		storage:  storage,
		stop:     make(chan struct{}),
	}
}

func (w *Worker) Start() {
	w.wg.Add(1)
	go w.run()
}

func (w *Worker) Stop() {
	close(w.stop)
	w.wg.Wait()
}

func (w *Worker) run() {
	defer w.wg.Done()
	ticker := time.NewTicker(5 * time.Second) // poll every 5s
	defer ticker.Stop()

	for {
		select {
		case <-w.stop:
			return
		case <-ticker.C:
			w.processJobs()
		}
	}
}

func (w *Worker) processJobs() {
	ctx := context.Background()

	// Get all queued jobs and jobs ready for retry
	queuedJobs, err := w.store.Jobs().ListQueued(ctx)
	if err != nil {
		log.Printf("Error listing queued jobs: %v", err)
		return
	}

	retryJobs, err := w.store.Jobs().ListRetry(ctx)
	if err != nil {
		log.Printf("Error listing retry jobs: %v", err)
		return
	}

	// Filter retry jobs that are ready to be retried based on their policy
	readyRetryJobs := w.filterReadyRetryJobs(ctx, retryJobs)

	allJobs := append(queuedJobs, readyRetryJobs...)

	if len(allJobs) == 0 {
		log.Println("Worker polling... no jobs to process")
		return
	}

	log.Printf("Worker processing %d jobs (%d queued, %d retry)", len(allJobs), len(queuedJobs), len(readyRetryJobs))

	for _, job := range allJobs {
		if err := w.processJob(ctx, job); err != nil {
			log.Printf("Error processing job %s: %v", job.ID, err)
		}
	}
}

func (w *Worker) filterReadyRetryJobs(ctx context.Context, jobs []store.Job) []store.Job {
	var readyJobs []store.Job
	now := time.Now().UTC()

	for _, job := range jobs {
		policy := queue.GetRetryPolicy(string(job.Type))
		nextRetryDelay := queue.CalculateNextRetryDelay(policy, job.RetryCount)

		if job.LastRetryAt == nil || job.LastRetryAt.Add(nextRetryDelay).Before(now) {
			readyJobs = append(readyJobs, job)
		}
	}

	return readyJobs
}

// ProcessJobs is a public wrapper for testing
func (w *Worker) ProcessJobs() {
	w.processJobs()
}

func (w *Worker) processJob(ctx context.Context, job store.Job) error {
	// Update job status to Running
	job.Status = store.JobRunning
	if _, err := w.store.Jobs().Update(ctx, job); err != nil {
		return fmt.Errorf("failed to update job status to running: %w", err)
	}

	// Get template version from inputRef
	templateVersion, ok, err := w.store.Templates().GetVersion(ctx, job.OrgID, job.InputRef)
	if err != nil {
		return w.handleJobFailure(ctx, job, fmt.Errorf("failed to get template version: %w", err))
	}
	if !ok {
		return w.handleJobFailure(ctx, job, fmt.Errorf("template version not found"))
	}

	var outputRef string
	var processErr error

	switch job.Type {
	case store.JobRender, store.JobExport:
		outputRef, processErr = w.processRenderJob(ctx, job, templateVersion)
	case store.JobPreview:
		outputRef, processErr = w.processPreviewJob(ctx, job, templateVersion)
	default:
		return w.handleJobFailure(ctx, job, fmt.Errorf("unsupported job type: %s", job.Type))
	}

	if processErr != nil {
		return w.handleJobFailure(ctx, job, processErr)
	}

	// Mark job as completed
	job.Status = store.JobDone
	job.OutputRef = outputRef
	if _, err := w.store.Jobs().Update(ctx, job); err != nil {
		return fmt.Errorf("failed to update job status to done: %w", err)
	}

	log.Printf("Successfully completed job %s, output: %s", job.ID, outputRef)
	return nil
}

func (w *Worker) processRenderJob(ctx context.Context, job store.Job, templateVersion store.TemplateVersion) (string, error) {
	// Render PPTX
	data, err := w.renderer.RenderPPTXBytes(ctx, templateVersion.SpecJSON)
	if err != nil {
		return "", fmt.Errorf("failed to render PPTX: %w", err)
	}

	// Generate asset ID
	assetID := fmt.Sprintf("%s-%d.pptx", job.ID, time.Now().Unix())

	// Store asset
	asset := store.Asset{
		ID:    assetID,
		OrgID: job.OrgID,
		Type:  store.AssetPPTX,
		Mime:  "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	}
	if _, err := w.store.Assets().Create(ctx, asset); err != nil {
		return "", fmt.Errorf("failed to create asset record: %w", err)
	}

	// Store file
	path, err := w.store.Assets().Store(ctx, job.OrgID, assetID, data)
	if err != nil {
		return "", fmt.Errorf("failed to store asset data: %w", err)
	}

	return path, nil
}

func (w *Worker) processPreviewJob(ctx context.Context, job store.Job, templateVersion store.TemplateVersion) (string, error) {
	// Generate thumbnails for each slide
	thumbnails, err := w.renderer.GenerateSlideThumbnails(ctx, templateVersion.SpecJSON)
	if err != nil {
		return "", fmt.Errorf("failed to generate slide thumbnails: %w", err)
	}

	if len(thumbnails) == 0 {
		return "", fmt.Errorf("no thumbnails generated")
	}

	var assetPaths []string

	// Store each thumbnail as a separate asset
	for i, thumbnailData := range thumbnails {
		// Generate asset ID for this thumbnail
		assetID := fmt.Sprintf("%s-%d-slide-%d.preview.png", job.ID, time.Now().Unix(), i+1)

		// Store preview asset record
		asset := store.Asset{
			ID:    assetID,
			OrgID: job.OrgID,
			Type:  store.AssetPNG,
			Mime:  "image/png",
		}
		if _, err := w.store.Assets().Create(ctx, asset); err != nil {
			return "", fmt.Errorf("failed to create preview asset record for slide %d: %w", i+1, err)
		}

		// Store the thumbnail data
		path, err := w.store.Assets().Store(ctx, job.OrgID, assetID, thumbnailData)
		if err != nil {
			return "", fmt.Errorf("failed to store preview data for slide %d: %w", i+1, err)
		}

		assetPaths = append(assetPaths, path)
	}

	// Return the first thumbnail as the primary preview, with metadata about all thumbnails
	// In a more complete implementation, we might want to return JSON metadata instead
	return assetPaths[0], nil
}

func (w *Worker) handleJobFailure(ctx context.Context, job store.Job, processErr error) error {
	errorMsg := processErr.Error()
	errorType := queue.ClassifyError(processErr)
	policy := queue.GetRetryPolicy(string(job.Type))

	// Use job's MaxRetries if set, otherwise use policy default
	maxRetries := job.MaxRetries
	if maxRetries == 0 {
		maxRetries = policy.MaxRetries
		job.MaxRetries = maxRetries
	}

	log.Printf("Job %s failed with %s error: %s", job.ID, errorType, errorMsg)

	if errorType == queue.ErrorTypePermanent || job.RetryCount >= maxRetries {
		// Move to dead letter queue
		job.Status = store.JobDeadLetter
		job.Error = fmt.Sprintf("%s (Error type: %s, Final retry: %d/%d)", errorMsg, errorType, job.RetryCount, maxRetries)
		if _, err := w.store.Jobs().Update(ctx, job); err != nil {
			return fmt.Errorf("failed to update job status to dead letter: %w", err)
		}
		log.Printf("Moved job %s to dead letter queue after %d retries", job.ID, job.RetryCount)
		return fmt.Errorf("job moved to dead letter: %s", errorMsg)
	}

	// Schedule for retry
	job.Status = store.JobRetry
	job.RetryCount++
	job.Error = errorMsg
	now := time.Now().UTC()
	job.LastRetryAt = &now

	if _, err := w.store.Jobs().Update(ctx, job); err != nil {
		return fmt.Errorf("failed to update job for retry: %w", err)
	}

	nextRetryDelay := queue.CalculateNextRetryDelay(policy, job.RetryCount)
	log.Printf("Scheduled job %s for retry %d/%d in %v", job.ID, job.RetryCount, maxRetries, nextRetryDelay)
	return fmt.Errorf("job scheduled for retry: %s", errorMsg)
}

func (w *Worker) failJob(ctx context.Context, job store.Job, errorMsg string) error {
	return w.handleJobFailure(ctx, job, fmt.Errorf("%s", errorMsg))
}
