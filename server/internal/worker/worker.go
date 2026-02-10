package worker

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ziyad/cms-ai/server/internal/ai"
	"github.com/ziyad/cms-ai/server/internal/assets"
	"github.com/ziyad/cms-ai/server/internal/queue"
	"github.com/ziyad/cms-ai/server/internal/spec"
	"github.com/ziyad/cms-ai/server/internal/store"
)

type Worker struct {
	store     store.Store
	renderer  assets.Renderer
	storage   assets.ObjectStorage
	aiService ai.AIServiceInterface
	stop      chan struct{}
	wg        sync.WaitGroup
}

func New(store store.Store, renderer assets.Renderer, storage assets.ObjectStorage, aiService ai.AIServiceInterface) *Worker {
	return &Worker{
		store:     store,
		renderer:  renderer,
		storage:   storage,
		aiService: aiService,
		stop:      make(chan struct{}),
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

	log.Printf("Successfully completed job %s, output: %s", job.ID, outputRef)
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
		SpecJSON:  specJSON,
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
	specBytes, _ := json.Marshal(tv.SpecJSON)
	if err := json.Unmarshal(specBytes, &templateSpec); err != nil {
		return "", fmt.Errorf("invalid template spec")
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
		SpecJSON:  boundBytes,
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

	// Render PPTX
	data, err := w.renderer.RenderPPTXBytes(ctx, templateVersion.SpecJSON)
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

	// Render PPTX for deck version
	data, err := w.renderer.RenderPPTXBytes(ctx, deckVersion.SpecJSON)
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

// newID generates a UUID with the given prefix
func newID(prefix string) string {
	var b [16]byte
	_, err := rand.Read(b[:])
	if err != nil {
		return prefix + "-unknown"
	}
	return prefix + "-" + hex.EncodeToString(b[:])
}
