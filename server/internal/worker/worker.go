package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ziyad/cms-ai/server/internal/ai"
	"github.com/ziyad/cms-ai/server/internal/assets"
	"github.com/ziyad/cms-ai/server/internal/logger"
	"github.com/ziyad/cms-ai/server/internal/queue"
	"github.com/ziyad/cms-ai/server/internal/spec"
	"github.com/ziyad/cms-ai/server/internal/store"
)

type Worker struct {
	store      store.Store
	renderer   assets.Renderer
	storage    assets.ObjectStorage
	aiService  ai.AIServiceInterface
	stop       chan struct{}
	wg         sync.WaitGroup
	JobTimeout time.Duration // max time per job; 0 = default (2 min)
}

func New(store store.Store, renderer assets.Renderer, storage assets.ObjectStorage, aiService ai.AIServiceInterface) *Worker {
	return &Worker{
		store:      store,
		renderer:   renderer,
		storage:    storage,
		aiService:  aiService,
		stop:       make(chan struct{}),
		JobTimeout: 2 * time.Minute,
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
		logger.LogError(ctx, "worker", "list_queued_jobs", err)
		return
	}

	retryJobs, err := w.store.Jobs().ListRetry(ctx)
	if err != nil {
		logger.LogError(ctx, "worker", "list_retry_jobs", err)
		return
	}

	// Filter retry jobs that are ready to be retried based on their policy
	readyRetryJobs := w.filterReadyRetryJobs(ctx, retryJobs)

	allJobs := append(queuedJobs, readyRetryJobs...)

	if len(allJobs) == 0 {
		logger.Jobs().Debug("worker_polling_no_jobs")
		return
	}

	logger.Jobs().Info("worker_processing_jobs", "total", len(allJobs), "queued", len(queuedJobs), "retry", len(readyRetryJobs))

	for _, job := range allJobs {
		if err := w.processJob(ctx, job); err != nil {
			logger.LogError(ctx, "worker", "process_job", err, "job_id", job.ID)
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
	// Enforce a timeout so jobs don't hang forever (e.g., if Python renderer hangs).
	timeout := w.JobTimeout
	if timeout == 0 {
		timeout = 2 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Update job status to Running
	job.Status = store.JobRunning
	if _, err := w.store.Jobs().Update(ctx, job); err != nil {
		return fmt.Errorf("failed to update job status to running: %w", err)
	}

	var outputRef string
	var processErr error

	switch job.Type {
	case store.JobGenerate:
		outputRef, processErr = w.processGenerateJob(ctx, job)
	case store.JobBind:
		outputRef, processErr = w.processBindJob(ctx, job)
	case store.JobRender, store.JobExport:
		// Check if it's a deck export (deck version ID) or template export
		if deckVersion, ok, err := w.store.Decks().GetDeckVersion(ctx, job.OrgID, job.InputRef); err == nil && ok {
			outputRef, processErr = w.processDeckRenderJob(ctx, job, deckVersion)
		} else {
			// Fall back to template version
			templateVersion, ok, err := w.store.Templates().GetVersion(ctx, job.OrgID, job.InputRef)
			if err != nil {
				return w.handleJobFailure(ctx, job, fmt.Errorf("failed to get template version: %w", err))
			}
			if !ok {
				return w.handleJobFailure(ctx, job, fmt.Errorf("template version not found"))
			}
			outputRef, processErr = w.processRenderJob(ctx, job, templateVersion)
		}
	case store.JobPreview:
		// Preview only works for templates
		templateVersion, ok, err := w.store.Templates().GetVersion(ctx, job.OrgID, job.InputRef)
		if err != nil {
			return w.handleJobFailure(ctx, job, fmt.Errorf("failed to get template version: %w", err))
		}
		if !ok {
			return w.handleJobFailure(ctx, job, fmt.Errorf("template version not found"))
		}
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

	logger.Jobs().Info("job_completed_successfully", "job_id", job.ID, "output_ref", outputRef)
	return nil
}

func (w *Worker) processGenerateJob(ctx context.Context, job store.Job) (string, error) {
	if job.Metadata == nil {
		return "", fmt.Errorf("missing job metadata")
	}
	m := *job.Metadata
	prompt := m["prompt"]
	language := m["language"]
	tone := m["tone"]
	rtl := m["rtl"] == "true"
	brandKitID := m["brandKitId"]
	userID := m["userId"]

	w.updateProgress(ctx, &job, "Analyzing prompt with AI", 20)

	aiReq := ai.GenerationRequest{
		Prompt:   prompt,
		Language: language,
		Tone:     tone,
		RTL:      rtl,
	}

	templateSpec, _, err := w.aiService.GenerateTemplateForRequest(ctx, job.OrgID, userID, aiReq, brandKitID)
	if err != nil {
		return "", fmt.Errorf("AI template generation failed: %w", err)
	}

	w.updateProgress(ctx, &job, "Finalizing design tokens", 70)

	specJSON, err := json.Marshal(templateSpec)
	if err != nil {
		return "", fmt.Errorf("failed to marshal template spec: %w", err)
	}

	version := store.TemplateVersion{
		ID:        newID("tv"),
		Template:  job.InputRef,
		OrgID:     job.OrgID,
		VersionNo: 1,
		SpecJSON:  json.RawMessage(specJSON),
		CreatedBy: userID,
	}
	createdVer, err := w.store.Templates().CreateVersion(ctx, version)
	if err != nil {
		return "", fmt.Errorf("failed to create template version: %w", err)
	}

	// Update template with current version
	template, ok, err := w.store.Templates().GetTemplate(ctx, job.OrgID, job.InputRef)
	if err == nil && ok {
		template.CurrentVersion = &createdVer.ID
		template.LatestVersionNo = 1
		_, _ = w.store.Templates().UpdateTemplate(ctx, template)
	}

	return createdVer.ID, nil
}

func (w *Worker) processBindJob(ctx context.Context, job store.Job) (string, error) {
	if job.Metadata == nil {
		return "", fmt.Errorf("missing job metadata")
	}
	m := *job.Metadata
	templateVersionID := m["sourceTemplateVersionId"]
	content := m["content"]
	userID := m["userId"]
	deckID := job.InputRef

	w.updateProgress(ctx, &job, "Summarizing content with AI", 20)

	// Load template version
	tv, ok, err := w.store.Templates().GetVersion(ctx, job.OrgID, templateVersionID)
	if err != nil || !ok {
		return "", fmt.Errorf("failed to load template version")
	}

	var templateSpec spec.TemplateSpec
	// tv.SpecJSON is type `any`. From pgx it arrives as Go string (not []byte).
	// json.Marshal(string) double-encodes → "\"...\"" which breaks Unmarshal.
	specBytes, err := anyToJSONBytes(tv.SpecJSON)
	if err != nil {
		return "", fmt.Errorf("invalid template spec: %w", err)
	}
	if err := json.Unmarshal(specBytes, &templateSpec); err != nil {
		return "", fmt.Errorf("invalid template spec: %w", err)
	}

	boundSpec, _, err := w.aiService.BindDeckSpec(ctx, job.OrgID, userID, &templateSpec, content)
	if err != nil {
		return "", fmt.Errorf("AI binding failed: %w", err)
	}

	w.updateProgress(ctx, &job, "Assembling slides", 70)

	boundBytes, err := json.Marshal(boundSpec)
	if err != nil {
		return "", fmt.Errorf("failed to marshal bound spec: %w", err)
	}

	version := store.DeckVersion{
		ID:        newID("dv"),
		Deck:      deckID,
		OrgID:     job.OrgID,
		VersionNo: 1,
		SpecJSON:  json.RawMessage(boundBytes),
		CreatedBy: userID,
	}
	createdVer, err := w.store.Decks().CreateDeckVersion(ctx, version)
	if err != nil {
		return "", fmt.Errorf("failed to create deck version: %w", err)
	}

	// Update deck with current version
	deck, ok, err := w.store.Decks().GetDeck(ctx, job.OrgID, deckID)
	if err == nil && ok {
		deck.CurrentVersion = &createdVer.ID
		deck.LatestVersionNo = 1
		_, _ = w.store.Decks().UpdateDeck(ctx, deck)
	}

	return createdVer.ID, nil
}

func (w *Worker) updateProgress(ctx context.Context, job *store.Job, step string, pct int) {
	job.ProgressStep = step
	job.ProgressPct = pct
	_, _ = w.store.Jobs().Update(ctx, *job)
}

func (w *Worker) processRenderJob(ctx context.Context, job store.Job, templateVersion store.TemplateVersion) (string, error) {
	w.updateProgress(ctx, &job, "Generating PowerPoint slides", 20)

	// Normalize spec — pgx returns jsonb as Go string, possibly base64-encoded.
	normalizedSpec, err := anyToJSONBytes(templateVersion.SpecJSON)
	if err != nil {
		return "", fmt.Errorf("failed to normalize template spec: %w", err)
	}

	// Render PPTX
	data, err := w.renderer.RenderPPTXBytes(ctx, json.RawMessage(normalizedSpec))
	if err != nil {
		return "", fmt.Errorf("failed to render PPTX: %w", err)
	}

	w.updateProgress(ctx, &job, "Applying Olama AI themes", 60)

	// Generate proper UUID asset ID
	assetID := newID("asset")
	storageKey := assetID + ".pptx"

	// Upload to object storage
	metadata, err := w.storage.Upload(ctx, storageKey, data, "application/vnd.openxmlformats-officedocument.presentationml.presentation")
	if err != nil {
		return "", fmt.Errorf("failed to upload asset to storage: %w", err)
	}

	w.updateProgress(ctx, &job, "Saving to database", 90)

	// Create asset record with storage key
	asset := store.Asset{
		ID:    assetID,
		OrgID: job.OrgID,
		Type:  store.AssetPPTX,
		Path:  metadata.Key,
		Mime:  metadata.ContentType,
	}
	if _, err := w.store.Assets().Create(ctx, asset); err != nil {
		return "", fmt.Errorf("failed to create asset record: %w", err)
	}

	return assetID, nil
}

func (w *Worker) processDeckRenderJob(ctx context.Context, job store.Job, deckVersion store.DeckVersion) (string, error) {
	w.updateProgress(ctx, &job, "Generating deck visuals", 20)

	// CRITICAL: Normalize the spec BEFORE passing to renderer.
	// pgx returns jsonb as Go string. If GORM wrote []byte, the string is base64.
	// The renderer's specToJSONBytes should handle this, but we normalize here as
	// a belt-and-suspenders approach to prevent the Python script from receiving
	// a base64 string instead of a JSON object.
	normalizedSpec, err := anyToJSONBytes(deckVersion.SpecJSON)
	if err != nil {
		return "", fmt.Errorf("failed to normalize deck spec: %w", err)
	}
	logger.Jobs().Info("deck_export_spec_normalized",
		"job_id", job.ID,
		"input_type", fmt.Sprintf("%T", deckVersion.SpecJSON),
		"output_len", len(normalizedSpec),
		"first50", string(normalizedSpec[:min(50, len(normalizedSpec))]))

	// Render PPTX for deck version — pass normalized JSON bytes
	data, err := w.renderer.RenderPPTXBytes(ctx, json.RawMessage(normalizedSpec))
	if err != nil {
		return "", fmt.Errorf("failed to render deck PPTX: %w", err)
	}

	w.updateProgress(ctx, &job, "Enhancing with AI themes", 60)

	// Generate proper UUID asset ID
	assetID := newID("asset")
	storageKey := assetID + ".pptx"

	// Upload to object storage
	metadata, err := w.storage.Upload(ctx, storageKey, data, "application/vnd.openxmlformats-officedocument.presentationml.presentation")
	if err != nil {
		return "", fmt.Errorf("failed to upload deck asset to storage: %w", err)
	}

	w.updateProgress(ctx, &job, "Finalizing export", 90)

	// Create asset record with storage key
	asset := store.Asset{
		ID:    assetID,
		OrgID: job.OrgID,
		Type:  store.AssetPPTX,
		Path:  metadata.Key,
		Mime:  metadata.ContentType,
	}
	if _, err := w.store.Assets().Create(ctx, asset); err != nil {
		return "", fmt.Errorf("failed to create deck asset record: %w", err)
	}

	return assetID, nil
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

	var firstAssetURL string

	// Store each thumbnail as a separate asset
	for i, thumbnailData := range thumbnails {
		// Generate asset ID for this thumbnail
		assetID := fmt.Sprintf("%s-%d-slide-%d.preview.png", job.ID, time.Now().Unix(), i+1)

		// Upload to storage
		metadata, err := w.storage.Upload(ctx, assetID, thumbnailData, "image/png")
		if err != nil {
			return "", fmt.Errorf("failed to upload preview data for slide %d: %w", i+1, err)
		}

		// Create preview asset record
		asset := store.Asset{
			ID:    assetID,
			OrgID: job.OrgID,
			Type:  store.AssetPNG,
			Path:  metadata.Key,
			Mime:  "image/png",
		}
		if _, err := w.store.Assets().Create(ctx, asset); err != nil {
			return "", fmt.Errorf("failed to create preview asset record for slide %d: %w", i+1, err)
		}

		if i == 0 {
			firstAssetURL = metadata.URL
		}
	}

	// Return the first thumbnail URL or ID as the primary preview
	return firstAssetURL, nil
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

	logger.Jobs().Warn("job_execution_failed", "job_id", job.ID, "error_type", errorType, "error", errorMsg, "retry_count", job.RetryCount, "max_retries", maxRetries)

	if errorType == queue.ErrorTypePermanent || job.RetryCount >= maxRetries {
		// Move to dead letter queue
		job.Status = store.JobDeadLetter
		job.Error = fmt.Sprintf("%s (Error type: %s, Final retry: %d/%d)", errorMsg, errorType, job.RetryCount, maxRetries)
		if _, err := w.store.Jobs().Update(ctx, job); err != nil {
			return fmt.Errorf("failed to update job status to dead letter: %w", err)
		}
		logger.Jobs().Error("job_moved_to_dead_letter", "job_id", job.ID, "retries", job.RetryCount)
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
	logger.Jobs().Info("job_scheduled_for_retry", "job_id", job.ID, "retry_no", job.RetryCount, "max_retries", maxRetries, "delay_seconds", nextRetryDelay.Seconds())
	return fmt.Errorf("job scheduled for retry: %s", errorMsg)
}

func (w *Worker) failJob(ctx context.Context, job store.Job, errorMsg string) error {
	return w.handleJobFailure(ctx, job, fmt.Errorf("%s", errorMsg))
}

// anyToJSONBytes converts an `any` value to JSON bytes safely.
// Handles the pgx quirk where jsonb columns return as Go string.
// json.Marshal(string) would double-encode, so we must handle it explicitly.
func anyToJSONBytes(v any) ([]byte, error) {
	switch val := v.(type) {
	case []byte:
		return assets.NormalizeJSONBytes(val), nil
	case json.RawMessage:
		return assets.NormalizeJSONBytes([]byte(val)), nil
	case string:
		return assets.NormalizeJSONBytes([]byte(val)), nil
	default:
		return json.Marshal(v)
	}
}

// newID generates a proper UUID (compatible with PostgreSQL uuid columns).
func newID(prefix string) string {
	return uuid.New().String()
}
