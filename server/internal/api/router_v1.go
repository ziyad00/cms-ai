package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ziyad/cms-ai/server/internal/ai"
	"github.com/ziyad/cms-ai/server/internal/auth"
	"github.com/ziyad/cms-ai/server/internal/spec"
	"github.com/ziyad/cms-ai/server/internal/store"
)

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	// Auth endpoints (no auth middleware for signup/signin)
	mux.HandleFunc("POST /v1/auth/signup", s.handleSignup)
	mux.HandleFunc("GET /v1/auth/signup", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("DEBUG: GET request to /v1/auth/signup - this should be POST")
		writeError(w, r, http.StatusMethodNotAllowed, "method not allowed - use POST")
	})
	mux.HandleFunc("/v1/auth/signup", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("DEBUG: %s request to /v1/auth/signup - unsupported method", r.Method)
		writeError(w, r, http.StatusMethodNotAllowed, "only POST supported")
	})
	mux.HandleFunc("POST /v1/auth/signin", s.handleSignin)
	mux.HandleFunc("POST /v1/auth/user", s.handleGetOrCreateUser) // Legacy endpoint

	// Protected auth endpoint (requires auth)
	mux.HandleFunc("GET /v1/auth/me", s.handleGetMe) // Get current user from JWT

	mux.HandleFunc("POST /v1/templates/validate", s.handleValidateTemplateSpec)
	mux.HandleFunc("POST /v1/templates/analyze", s.handleAnalyzeTemplate)
	mux.HandleFunc("POST /v1/templates", s.handleCreateTemplate)
	mux.HandleFunc("POST /v1/templates/generate", s.handleGenerateTemplate)
	mux.HandleFunc("GET /v1/templates", s.handleListTemplates)
	mux.HandleFunc("GET /v1/templates/{id}", s.handleGetTemplate)
	mux.HandleFunc("POST /v1/templates/{id}/versions", s.handleCreateVersion)
	mux.HandleFunc("GET /v1/templates/{id}/versions", s.handleListVersions)

	mux.HandleFunc("POST /v1/decks", s.handleCreateDeck)
	mux.HandleFunc("GET /v1/decks", s.handleListDecks)
	mux.HandleFunc("GET /v1/decks/{id}", s.handleGetDeck)
	mux.HandleFunc("POST /v1/decks/{id}/versions", s.handleCreateDeckVersion)
	mux.HandleFunc("GET /v1/decks/{id}/versions", s.handleListDeckVersions)
	mux.HandleFunc("POST /v1/deck-versions/{versionId}/export", s.handleExportDeckVersion)
	mux.HandleFunc("PATCH /v1/versions/{versionId}", s.handlePatchVersion)
	mux.HandleFunc("POST /v1/versions/{versionId}/render", s.handleRenderVersion)
	mux.HandleFunc("POST /v1/versions/{versionId}/export", s.handleExportVersion)
	mux.HandleFunc("GET /v1/assets/{id}/download-url", s.handleDownloadURL)
	mux.HandleFunc("GET /v1/assets/{id}", s.handleAssetDownload)
	mux.HandleFunc("GET /v1/jobs/{jobId}", s.handleGetJob)
	mux.HandleFunc("GET /v1/jobs/{jobId}/assets/{filename}", s.handleJobAssetDownload)
	mux.HandleFunc("GET /v1/admin/jobs/dead-letter", s.handleListDeadLetterJobs)
	mux.HandleFunc("POST /v1/admin/jobs/{jobId}/retry", s.handleRetryDeadLetterJob)
	mux.HandleFunc("POST /v1/brand-kits", s.handleCreateBrandKit)
	mux.HandleFunc("GET /v1/brand-kits", s.handleListBrandKits)
	mux.HandleFunc("GET /v1/usage", s.handleUsage)

	// Database diagnostics endpoints
	mux.HandleFunc("GET /v1/admin/db/diagnostics", s.handleDatabaseDiagnostics)
	mux.HandleFunc("GET /v1/admin/db/query", s.handleDatabaseQuery)

	h := http.Handler(mux)
	h = requireJSON(h)
	h = withRequestID(h)

	// Re-enable auth middleware with skip paths for public endpoints
	skipPaths := []string{
		"/v1/auth/signup",
		"/v1/auth/signin",
		"/v1/auth/user", // Legacy endpoint
		"/healthz",
	}
	// Use the server's configured authenticator (JWT in prod, header-based in dev/tests)
	authMiddleware := withAuth(s.Authenticator)
	h = skipAuthForPaths(h, skipPaths, authMiddleware)

	h = withRecovery(h)
	h = withLogging(h)

	// Wrap with catch-all handler that returns 404 for unmatched routes
	// This prevents auth middleware from returning unauthorized for non-API routes
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If path doesn't match any route, return 404 without auth
		if !strings.HasPrefix(r.URL.Path, "/v1/") && r.URL.Path != "/healthz" {
			writeError(w, r, http.StatusNotFound, "not found")
			return
		}

		// DEBUG: Log all /v1/* requests to help debug routing issues
		urlParts := strings.Split(r.URL.Path, "/")
		log.Printf("DEBUG: %s %s - urlParts: %v", r.Method, r.URL.Path, urlParts)

		// Otherwise, use the main handler (which includes auth for /v1/*)
		h.ServeHTTP(w, r)
	})
}

func (s *Server) handleValidateTemplateSpec(w http.ResponseWriter, r *http.Request) {
	var ts spec.TemplateSpec
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	if err := dec.Decode(&ts); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid JSON body")
		return
	}

	errList := s.Validator.Validate(ts)
	if len(errList) > 0 {
		writeJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": errList})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *Server) handleAnalyzeTemplate(w http.ResponseWriter, r *http.Request) {
	var req AnalyzeTemplateRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if strings.TrimSpace(req.Prompt) == "" {
		writeError(w, r, http.StatusBadRequest, "prompt is required")
		return
	}

	// Analyze the prompt to determine template type and required fields
	analysis := analyzeTemplatePrompt(req.Prompt)

	writeJSON(w, http.StatusOK, analysis)
}

func analyzeTemplatePrompt(prompt string) AnalyzeTemplateResponse {
	prompt = strings.ToLower(prompt)

	// Simple pattern matching for template types
	if strings.Contains(prompt, "sales") || strings.Contains(prompt, "revenue") {
		return AnalyzeTemplateResponse{
			TemplateType:    "sales-report",
			SuggestedName:   "Sales Report",
			EstimatedSlides: 8,
			Description:     "A comprehensive sales performance report with metrics, trends, and insights",
			RequiredFields: []RequiredField{
				{Key: "period", Label: "Reporting Period", Type: "text", Required: true, Example: "Q4 2024", Description: "The time period this report covers"},
				{Key: "revenue", Label: "Total Revenue", Type: "currency", Required: true, Example: "$2.5M", Description: "Total revenue for the period"},
				{Key: "growth", Label: "Growth Rate", Type: "percentage", Required: false, Example: "15%", Description: "Revenue growth compared to previous period"},
				{Key: "deals", Label: "Number of Deals", Type: "number", Required: false, Example: "47", Description: "Total deals closed"},
				{Key: "topProducts", Label: "Top Products", Type: "list", Required: false, Example: "Product A, Product B", Description: "Best performing products"},
				{Key: "teamSize", Label: "Team Size", Type: "number", Required: false, Example: "25", Description: "Sales team size"},
				{Key: "goals", Label: "Goals Met", Type: "percentage", Required: false, Example: "105%", Description: "Percentage of goals achieved"},
			},
		}
	}

	if strings.Contains(prompt, "meeting") || strings.Contains(prompt, "agenda") {
		return AnalyzeTemplateResponse{
			TemplateType:    "meeting-notes",
			SuggestedName:   "Meeting Agenda",
			EstimatedSlides: 5,
			Description:     "Meeting agenda and notes template for team meetings",
			RequiredFields: []RequiredField{
				{Key: "title", Label: "Meeting Title", Type: "text", Required: true, Example: "Weekly Team Sync", Description: "Title of the meeting"},
				{Key: "date", Label: "Meeting Date", Type: "date", Required: true, Example: "2024-01-19", Description: "Date and time of meeting"},
				{Key: "attendees", Label: "Attendees", Type: "list", Required: false, Example: "John, Jane, Mike", Description: "List of attendees"},
				{Key: "agenda", Label: "Agenda Items", Type: "list", Required: true, Example: "Project updates, Budget review", Description: "Main topics to discuss"},
				{Key: "duration", Label: "Duration", Type: "text", Required: false, Example: "60 minutes", Description: "Expected meeting duration"},
			},
		}
	}

	if strings.Contains(prompt, "product") || strings.Contains(prompt, "demo") {
		return AnalyzeTemplateResponse{
			TemplateType:    "product-demo",
			SuggestedName:   "Product Demo",
			EstimatedSlides: 12,
			Description:     "Product demonstration and feature showcase presentation",
			RequiredFields: []RequiredField{
				{Key: "productName", Label: "Product Name", Type: "text", Required: true, Example: "My Product", Description: "Name of the product being presented"},
				{Key: "version", Label: "Version", Type: "text", Required: false, Example: "v2.1", Description: "Product version"},
				{Key: "keyFeatures", Label: "Key Features", Type: "list", Required: true, Example: "Feature A, Feature B", Description: "Main features to highlight"},
				{Key: "benefits", Label: "Benefits", Type: "list", Required: false, Example: "Saves time, Increases efficiency", Description: "Key benefits for users"},
				{Key: "audience", Label: "Target Audience", Type: "text", Required: false, Example: "Enterprise customers", Description: "Who this demo is for"},
			},
		}
	}

	// Default/generic template
	return AnalyzeTemplateResponse{
		TemplateType:    "generic",
		SuggestedName:   "Custom Presentation",
		EstimatedSlides: 6,
		Description:     "A general-purpose presentation template",
		RequiredFields: []RequiredField{
			{Key: "title", Label: "Presentation Title", Type: "text", Required: true, Example: "My Presentation", Description: "Main title for the presentation"},
			{Key: "subtitle", Label: "Subtitle", Type: "text", Required: false, Example: "Subtitle here", Description: "Optional subtitle"},
			{Key: "mainContent", Label: "Main Content", Type: "text", Required: false, Example: "Key points to present", Description: "Main content or talking points"},
		},
	}
}

func (s *Server) handleCreateTemplate(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.GetIdentity(r.Context())
	if !auth.RequireRole(id, auth.RoleEditor) {
		writeError(w, r, http.StatusForbidden, "forbidden")
		return
	}

	var req CreateTemplateRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid JSON body")
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		writeError(w, r, http.StatusBadRequest, "name is required")
		return
	}

	template := store.Template{
		OrgID:       id.OrgID,
		OwnerUserID: id.UserID,
		Name:        name,
		Status:      store.TemplateDraft,
	}

	created, err := s.Store.Templates().CreateTemplate(r.Context(), template)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to create template")
		return
	}

	_, _ = s.Store.Audit().Append(r.Context(), store.AuditLog{ID: newID("aud"), OrgID: id.OrgID, ActorID: id.UserID, Action: "template.create", TargetRef: created.ID, Metadata: map[string]any{"name": created.Name}})

	writeJSON(w, http.StatusOK, map[string]any{"template": created})
}

func (s *Server) handleGenerateTemplate(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.GetIdentity(r.Context())
	log.Printf("DEBUG: handleGenerateTemplate - UserID: %s, OrgID: %s", id.UserID, id.OrgID)

	var req GenerateTemplateRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		log.Printf("ERROR: Failed to decode request body: %v", err)
		writeError(w, r, http.StatusBadRequest, "invalid JSON body")
		return
	}
	log.Printf("DEBUG: Request - Name: '%s', Prompt: '%s'", req.Name, req.Prompt)
	if strings.TrimSpace(req.Prompt) == "" {
		writeError(w, r, http.StatusBadRequest, "prompt is required")
		return
	}
	if isBlocked, usage := s.enforceQuota(r); isBlocked {
		writeJSON(w, http.StatusPaymentRequired, usage)
		return
	}

	template := store.Template{
		OrgID:       id.OrgID,
		OwnerUserID: id.UserID,
		Name:        req.Name,
		Status:      store.TemplateDraft,
	}
	if template.Name == "" {
		template.Name = "Untitled"
	}

	created, err := s.Store.Templates().CreateTemplate(r.Context(), template)
	if err != nil {
		log.Printf("ERROR: Failed to create template: %v", err)
		writeError(w, r, http.StatusInternalServerError, "failed to create template")
		return
	}

	// Generate template spec using AI with user content
	aiReq := ai.GenerationRequest{
		Prompt:      req.Prompt,
		Language:    req.Language,
		Tone:        req.Tone,
		RTL:         req.RTL,
		ContentData: req.ContentData, // Pass user content to AI
	}

	templateSpec, aiResp, err := s.AIService.GenerateTemplateForRequest(r.Context(), id.OrgID, id.UserID, aiReq, req.BrandKitID)
	if err != nil {
		log.Printf("ERROR: AI template generation failed: %v", err)
		writeError(w, r, http.StatusInternalServerError, "failed to generate template with AI")
		return
	}

	// Convert template spec to JSON for storage
	specJSON, err := json.Marshal(templateSpec)
	if err != nil {
		log.Printf("ERROR: Failed to marshal template spec: %v", err)
		writeError(w, r, http.StatusInternalServerError, "failed to create template")
		return
	}

	version := store.TemplateVersion{
		Template:  created.ID,
		OrgID:     id.OrgID,
		VersionNo: 1,
		SpecJSON:  specJSON,
		CreatedBy: id.UserID,
	}
	createdVer, err := s.Store.Templates().CreateVersion(r.Context(), version)
	if err != nil {
		log.Printf("ERROR: Failed to create version: %v", err)
		writeError(w, r, http.StatusInternalServerError, "failed to create version")
		return
	}
	created.CurrentVersion = &createdVer.ID
	created.LatestVersionNo = 1
	created, _ = s.Store.Templates().UpdateTemplate(r.Context(), created)

	_, _ = s.Store.Metering().Record(r.Context(), store.MeteringEvent{ID: newID("met"), OrgID: id.OrgID, UserID: id.UserID, Type: "generate", Quantity: 1})

	metadata := map[string]any{"prompt": req.Prompt}
	if aiResp != nil {
		metadata["aiModel"] = aiResp.Model
		metadata["aiTokenUsage"] = aiResp.TokenUsage
		metadata["aiCost"] = aiResp.Cost
	}
	_, _ = s.Store.Audit().Append(r.Context(), store.AuditLog{ID: newID("aud"), OrgID: id.OrgID, ActorID: id.UserID, Action: "template.generate", TargetRef: created.ID, Metadata: metadata})

	response := map[string]any{"template": created, "version": createdVer}
	if aiResp != nil {
		response["aiResponse"] = map[string]any{
			"model":      aiResp.Model,
			"tokenUsage": aiResp.TokenUsage,
			"cost":       aiResp.Cost,
			"timestamp":  aiResp.Timestamp,
		}
	}

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleListTemplates(w http.ResponseWriter, r *http.Request) {
	id, ok := auth.GetIdentity(r.Context())
	log.Printf("DEBUG: handleListTemplates - Auth OK: %v, UserID: %s, OrgID: %s", ok, id.UserID, id.OrgID)

	log.Printf("DEBUG: About to call ListTemplates for OrgID: %s", id.OrgID)
	tpls, err := s.Store.Templates().ListTemplates(r.Context(), id.OrgID)
	if err != nil {
		log.Printf("ERROR: ListTemplates failed for OrgID %s: %v", id.OrgID, err)
		writeError(w, r, http.StatusInternalServerError, "failed to list templates")
		return
	}
	log.Printf("DEBUG: ListTemplates success for OrgID %s, found %d templates", id.OrgID, len(tpls))
	writeJSON(w, http.StatusOK, map[string]any{"templates": tpls})
}

func (s *Server) handleGetTemplate(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.GetIdentity(r.Context())

	tplID := r.PathValue("id")
	tpl, ok, err := s.Store.Templates().GetTemplate(r.Context(), id.OrgID, tplID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to get template")
		return
	}
	if !ok {
		writeError(w, r, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"template": tpl})
}

func (s *Server) handleListVersions(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.GetIdentity(r.Context())
	pl := r.PathValue("id")

	vs, err := s.Store.Templates().ListVersions(r.Context(), id.OrgID, pl)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to list versions")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"versions": vs})
}

func (s *Server) handleCreateVersion(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.GetIdentity(r.Context())

	tplID := r.PathValue("id")
	tpl, ok, err := s.Store.Templates().GetTemplate(r.Context(), id.OrgID, tplID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to get template")
		return
	}
	if !ok {
		writeError(w, r, http.StatusNotFound, "not found")
		return
	}

	if !auth.RequireRole(id, auth.RoleEditor) {
		writeError(w, r, http.StatusForbidden, "forbidden")
		return
	}

	var req CreateVersionRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid JSON body")
		return
	}

	specJSON := req.Spec
	if specJSON == nil {
		specJSON = stubTemplateSpec()
	}

	newNo := tpl.LatestVersionNo + 1
	// Convert spec to JSON for storage
	specJSONBytes, err := json.Marshal(specJSON)
	if err != nil {
		log.Printf("ERROR: Failed to marshal spec JSON: %v", err)
		writeError(w, r, http.StatusInternalServerError, "failed to create version")
		return
	}

	ver := store.TemplateVersion{Template: tpl.ID, OrgID: tpl.OrgID, VersionNo: newNo, SpecJSON: specJSONBytes, CreatedBy: id.UserID}
	created, err := s.Store.Templates().CreateVersion(r.Context(), ver)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to create version")
		return
	}
	tpl.LatestVersionNo = newNo
	createdTpl, _ := s.Store.Templates().UpdateTemplate(r.Context(), tpl)

	_, _ = s.Store.Audit().Append(r.Context(), store.AuditLog{ID: newID("aud"), OrgID: id.OrgID, ActorID: id.UserID, Action: "template.version.create", TargetRef: created.ID, Metadata: map[string]any{"templateId": tpl.ID}})

	writeJSON(w, http.StatusOK, map[string]any{"template": createdTpl, "version": created})
}

func (s *Server) handlePatchVersion(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.GetIdentity(r.Context())
	versionID := r.PathValue("versionId")
	v, ok, err := s.Store.Templates().GetVersion(r.Context(), id.OrgID, versionID)

	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed")
		return
	}
	if !ok {
		writeError(w, r, http.StatusNotFound, "not found")
		return
	}
	if !auth.RequireRole(id, auth.RoleEditor) {
		writeError(w, r, http.StatusForbidden, "forbidden")
		return
	}

	var req PatchVersionRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Spec == nil {
		writeError(w, r, http.StatusBadRequest, "spec is required")
		return
	}

	// Immutable versions strategy: create a new version with incremented version number.
	tpl, ok2, err := s.Store.Templates().GetTemplate(r.Context(), id.OrgID, v.Template)
	if err != nil || !ok2 {
		writeError(w, r, http.StatusInternalServerError, "failed to load template")
		return
	}
	newNo := tpl.LatestVersionNo + 1
	// Convert spec to JSON for storage
	specJSONBytes, err := json.Marshal(req.Spec)
	if err != nil {
		log.Printf("ERROR: Failed to marshal spec JSON: %v", err)
		writeError(w, r, http.StatusInternalServerError, "failed to create version")
		return
	}

	newV := store.TemplateVersion{Template: tpl.ID, OrgID: tpl.OrgID, VersionNo: newNo, SpecJSON: specJSONBytes, CreatedBy: id.UserID}
	created, err := s.Store.Templates().CreateVersion(r.Context(), newV)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to create version")
		return
	}
	tpl.LatestVersionNo = newNo
	_, _ = s.Store.Templates().UpdateTemplate(r.Context(), tpl)

	_, _ = s.Store.Audit().Append(r.Context(), store.AuditLog{ID: newID("aud"), OrgID: id.OrgID, ActorID: id.UserID, Action: "template.version.patch", TargetRef: created.ID, Metadata: map[string]any{"fromVersionId": v.ID}})

	writeJSON(w, http.StatusOK, map[string]any{"version": created})
}

func (s *Server) handleRenderVersion(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.GetIdentity(r.Context())
	versionID := r.PathValue("versionId")
	_, ok, err := s.Store.Templates().GetVersion(r.Context(), id.OrgID, versionID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed")
		return
	}
	if !ok {
		writeError(w, r, http.StatusNotFound, "not found")
		return
	}
	if !auth.RequireRole(id, auth.RoleEditor) {
		writeError(w, r, http.StatusForbidden, "forbidden")
		return
	}

	job := store.Job{
		ID:              newID("job"),
		OrgID:           id.OrgID,
		Type:            store.JobRender,
		Status:          store.JobQueued,
		InputRef:        versionID,
		DeduplicationID: fmt.Sprintf("%s-%s", string(store.JobRender), versionID),
	}
	created, wasDuplicate, err := s.Store.Jobs().EnqueueWithDeduplication(r.Context(), job)
	if err != nil {
		log.Printf("ERROR: Failed to enqueue render job: %v", err)
		writeError(w, r, http.StatusInternalServerError, "failed to enqueue job")
		return
	}
	if wasDuplicate {
		writeJSON(w, http.StatusAccepted, map[string]any{"job": created, "duplicate": true})
		return
	}
	_, _ = s.Store.Audit().Append(r.Context(), store.AuditLog{ID: newID("aud"), OrgID: id.OrgID, ActorID: id.UserID, Action: "version.render.request", TargetRef: versionID, Metadata: map[string]any{"jobId": created.ID}})
	writeJSON(w, http.StatusAccepted, map[string]any{"job": created})
}

func (s *Server) handleGetJob(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.GetIdentity(r.Context())
	jobID := r.PathValue("jobId")

	job, ok, err := s.Store.Jobs().Get(r.Context(), id.OrgID, jobID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to get job")
		return
	}
	if !ok {
		writeError(w, r, http.StatusNotFound, "job not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"job": job})
}

func (s *Server) handleCreateDeck(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.GetIdentity(r.Context())
	if !auth.RequireRole(id, auth.RoleEditor) {
		writeError(w, r, http.StatusForbidden, "forbidden")
		return
	}

	var req CreateDeckRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		writeError(w, r, http.StatusBadRequest, "name is required")
		return
	}
	if strings.TrimSpace(req.SourceTemplateVersion) == "" {
		writeError(w, r, http.StatusBadRequest, "sourceTemplateVersionId is required")
		return
	}
	if strings.TrimSpace(req.Content) == "" {
		writeError(w, r, http.StatusBadRequest, "content is required")
		return
	}

	// Load template version spec (the "template")
	tv, ok, err := s.Store.Templates().GetVersion(r.Context(), id.OrgID, req.SourceTemplateVersion)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to load template version")
		return
	}
	if !ok {
		writeError(w, r, http.StatusNotFound, "template version not found")
		return
	}

	var templateSpec spec.TemplateSpec
	specBytes, err := assetsSpecBytes(tv.SpecJSON)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to read template spec")
		return
	}
	if err := json.Unmarshal(specBytes, &templateSpec); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid stored template spec")
		return
	}

	boundSpec, aiResp, err := s.AIService.BindDeckSpec(r.Context(), id.OrgID, id.UserID, &templateSpec, req.Content)
	if err != nil {
		writeError(w, r, http.StatusBadGateway, "failed to bind deck with AI")
		return
	}

	boundBytes, err := json.Marshal(boundSpec)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to marshal bound spec")
		return
	}

	// Create deck + version 1
	deck := store.Deck{
		OrgID:                 id.OrgID,
		OwnerUserID:           id.UserID,
		Name:                  req.Name,
		SourceTemplateVersion: req.SourceTemplateVersion,
		Content:               req.Content,
	}

	createdDeck, err := s.Store.Decks().CreateDeck(r.Context(), deck)
	if err != nil {
		requestID, _ := r.Context().Value(ctxKeyRequestID{}).(string)
		log.Printf("ERROR: Failed to create deck: request_id=%s err=%v", requestID, err)
		writeError(w, r, http.StatusInternalServerError, "failed to create deck")
		return
	}

	ver := store.DeckVersion{Deck: createdDeck.ID, OrgID: id.OrgID, VersionNo: 1, SpecJSON: boundBytes, CreatedBy: id.UserID}
	createdVer, err := s.Store.Decks().CreateDeckVersion(r.Context(), ver)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to create deck version")
		return
	}
	createdDeck.CurrentVersion = &createdVer.ID
	createdDeck.LatestVersionNo = 1
	createdDeck, _ = s.Store.Decks().UpdateDeck(r.Context(), createdDeck)

	resp := map[string]any{"deck": createdDeck, "version": createdVer}
	if aiResp != nil {
		resp["aiResponse"] = map[string]any{"model": aiResp.Model, "tokenUsage": aiResp.TokenUsage, "cost": aiResp.Cost, "timestamp": aiResp.Timestamp}
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleListDecks(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.GetIdentity(r.Context())
	ds, err := s.Store.Decks().ListDecks(r.Context(), id.OrgID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to list decks")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"decks": ds})
}

func (s *Server) handleGetDeck(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.GetIdentity(r.Context())
	deckID := r.PathValue("id")
	d, ok, err := s.Store.Decks().GetDeck(r.Context(), id.OrgID, deckID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to get deck")
		return
	}
	if !ok {
		writeError(w, r, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deck": d})
}

func (s *Server) handleListDeckVersions(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.GetIdentity(r.Context())
	deckID := r.PathValue("id")
	vs, err := s.Store.Decks().ListDeckVersions(r.Context(), id.OrgID, deckID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to list versions")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"versions": vs})
}

func (s *Server) handleCreateDeckVersion(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.GetIdentity(r.Context())
	if !auth.RequireRole(id, auth.RoleEditor) {
		writeError(w, r, http.StatusForbidden, "forbidden")
		return
	}

	deckID := r.PathValue("id")
	d, ok, err := s.Store.Decks().GetDeck(r.Context(), id.OrgID, deckID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to get deck")
		return
	}
	if !ok {
		writeError(w, r, http.StatusNotFound, "not found")
		return
	}

	var req CreateDeckVersionRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Spec == nil {
		writeError(w, r, http.StatusBadRequest, "spec is required")
		return
	}

	newNo := d.LatestVersionNo + 1
	specBytes, err := json.Marshal(req.Spec)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to marshal spec")
		return
	}

	ver := store.DeckVersion{Deck: d.ID, OrgID: id.OrgID, VersionNo: newNo, SpecJSON: specBytes, CreatedBy: id.UserID}
	created, err := s.Store.Decks().CreateDeckVersion(r.Context(), ver)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to create version")
		return
	}
	d.LatestVersionNo = newNo
	d.CurrentVersion = &created.ID
	updated, _ := s.Store.Decks().UpdateDeck(r.Context(), d)

	writeJSON(w, http.StatusOK, map[string]any{"deck": updated, "version": created})
}

func (s *Server) handleExportDeckVersion(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.GetIdentity(r.Context())
	versionID := r.PathValue("versionId")
	ver, ok, err := s.Store.Decks().GetDeckVersion(r.Context(), id.OrgID, versionID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed")
		return
	}
	if !ok {
		writeError(w, r, http.StatusNotFound, "not found")
		return
	}
	if isBlocked, usage := s.enforceExportQuota(r); isBlocked {
		writeJSON(w, http.StatusPaymentRequired, usage)
		return
	}

	// Synchronous export (same pattern as template export)
	objectKey := newID("asset") + ".pptx"
	tempPath := filepath.Join(os.TempDir(), objectKey)
	if err := s.Renderer.RenderPPTX(r.Context(), ver.SpecJSON, tempPath); err != nil {
		writeError(w, r, http.StatusInternalServerError, "render failed")
		return
	}
	defer os.Remove(tempPath)

	data, err := os.ReadFile(tempPath)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to read rendered file")
		return
	}
	_, err = s.ObjectStorage.Upload(r.Context(), objectKey, data, "application/vnd.openxmlformats-officedocument.presentationml.presentation")
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to upload asset")
		return
	}

	asset := store.Asset{OrgID: id.OrgID, Type: store.AssetPPTX, Path: objectKey, Mime: "application/vnd.openxmlformats-officedocument.presentationml.presentation"}
	createdAsset, err := s.Store.Assets().Create(r.Context(), asset)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to create asset")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"asset": createdAsset, "downloadUrl": "/v1/assets/" + createdAsset.ID})
}

func (s *Server) handleExportVersion(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.GetIdentity(r.Context())
	versionID := r.PathValue("versionId")
	ver, ok, err := s.Store.Templates().GetVersion(r.Context(), id.OrgID, versionID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed")
		return
	}
	if !ok {
		writeError(w, r, http.StatusNotFound, "not found")
		return
	}
	if isBlocked, usage := s.enforceExportQuota(r); isBlocked {
		writeJSON(w, http.StatusPaymentRequired, usage)
		return
	}

	job := store.Job{
		ID:              newID("job"),
		OrgID:           id.OrgID,
		Type:            store.JobExport,
		Status:          store.JobQueued,
		InputRef:        versionID,
		DeduplicationID: fmt.Sprintf("%s-%s", string(store.JobExport), versionID),
	}
	createdJob, wasDuplicate, err := s.Store.Jobs().EnqueueWithDeduplication(r.Context(), job)
	if err != nil {
		log.Printf("ERROR: Failed to enqueue export job: %v", err)
		writeError(w, r, http.StatusInternalServerError, "failed to enqueue job")
		return
	}
	if wasDuplicate {
		writeJSON(w, http.StatusAccepted, map[string]any{"job": createdJob, "duplicate": true})
		return
	}

	// Use a random filename for the stored object; the DB asset ID will be a UUID.
	objectKey := newID("asset") + ".pptx"

	// Render to temporary file first
	tempPath := filepath.Join(os.TempDir(), objectKey)
	if err := s.Renderer.RenderPPTX(r.Context(), ver.SpecJSON, tempPath); err != nil {
		writeError(w, r, http.StatusInternalServerError, "render failed")
		return
	}
	defer os.Remove(tempPath)

	// Read the rendered file and upload to object storage
	data, err := os.ReadFile(tempPath)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to read rendered file")
		return
	}

	_, err = s.ObjectStorage.Upload(r.Context(), objectKey, data, "application/vnd.openxmlformats-officedocument.presentationml.presentation")
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to upload asset")
		return
	}

	asset := store.Asset{OrgID: id.OrgID, Type: store.AssetPPTX, Path: objectKey, Mime: "application/vnd.openxmlformats-officedocument.presentationml.presentation"}
	createdAsset, err := s.Store.Assets().Create(r.Context(), asset)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to create asset")
		return
	}

	createdJob.Status = store.JobDone
	createdJob.OutputRef = createdAsset.ID
	if _, err := s.Store.Jobs().Update(r.Context(), createdJob); err != nil {
		requestID, _ := r.Context().Value(ctxKeyRequestID{}).(string)
		log.Printf("ERROR: Failed to update export job status: request_id=%s job_id=%s err=%v", requestID, createdJob.ID, err)
		writeError(w, r, http.StatusInternalServerError, "failed to update job")
		return
	}
	_, _ = s.Store.Metering().Record(r.Context(), store.MeteringEvent{ID: newID("met"), OrgID: id.OrgID, UserID: id.UserID, Type: "export", Quantity: 1})
	_, _ = s.Store.Audit().Append(r.Context(), store.AuditLog{ID: newID("aud"), OrgID: id.OrgID, ActorID: id.UserID, Action: "version.export", TargetRef: versionID, Metadata: map[string]any{"jobId": createdJob.ID, "assetId": createdAsset.ID}})

	writeJSON(w, http.StatusOK, map[string]any{"job": createdJob, "asset": createdAsset, "downloadUrl": "/v1/assets/" + createdAsset.ID})
}

func (s *Server) handleDownloadURL(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.GetIdentity(r.Context())
	assetID := r.PathValue("id")

	// Get the asset
	asset, ok, err := s.Store.Assets().Get(r.Context(), id.OrgID, assetID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to get asset")
		return
	}
	if !ok {
		writeError(w, r, http.StatusNotFound, "asset not found")
		return
	}

	// Generate signed URL
	signedURL, err := s.ObjectStorage.GetURL(r.Context(), asset.Path, 15*time.Minute)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to generate download URL")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"assetId": assetID, "downloadUrl": signedURL})
}

func (s *Server) handleCreateBrandKit(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.GetIdentity(r.Context())
	if !auth.RequireRole(id, auth.RoleEditor) {
		writeError(w, r, http.StatusForbidden, "forbidden")
		return
	}

	var payload struct {
		Name   string `json:"name"`
		Tokens any    `json:"tokens"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&payload); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if strings.TrimSpace(payload.Name) == "" {
		writeError(w, r, http.StatusBadRequest, "name is required")
		return
	}

	bk := store.BrandKit{ID: newID("bk"), OrgID: id.OrgID, Name: payload.Name, Tokens: payload.Tokens}
	created, err := s.Store.BrandKits().Create(r.Context(), bk)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed")
		return
	}
	_, _ = s.Store.Audit().Append(r.Context(), store.AuditLog{ID: newID("aud"), OrgID: id.OrgID, ActorID: id.UserID, Action: "brandkit.create", TargetRef: created.ID})
	writeJSON(w, http.StatusOK, map[string]any{"brandKit": created})
}

func (s *Server) handleListBrandKits(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.GetIdentity(r.Context())
	bks, err := s.Store.BrandKits().List(r.Context(), id.OrgID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"brandKits": bks})
}

func (s *Server) handleUsage(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.GetIdentity(r.Context())

	gen, _ := s.Store.Metering().SumByType(r.Context(), id.OrgID, "generate")
	exp, _ := s.Store.Metering().SumByType(r.Context(), id.OrgID, "export")

	limits := map[string]int{"generate": s.Config.GenerateLimitPerMonth, "export": s.Config.ExportLimitPerMonth}
	used := map[string]int{"generate": gen, "export": exp}
	blocked := gen >= limits["generate"] || exp >= limits["export"]

	writeJSON(w, http.StatusOK, UsageResponse{OrgID: id.OrgID, Limits: limits, Used: used, Blocked: blocked})
}

func (s *Server) enforceQuota(r *http.Request) (bool, UsageResponse) {
	id, _ := auth.GetIdentity(r.Context())
	gen, _ := s.Store.Metering().SumByType(r.Context(), id.OrgID, "generate")
	limits := map[string]int{"generate": s.Config.GenerateLimitPerMonth, "export": s.Config.ExportLimitPerMonth}
	used := map[string]int{"generate": gen}
	blocked := gen >= limits["generate"]
	return blocked, UsageResponse{OrgID: id.OrgID, Limits: limits, Used: used, Blocked: blocked}
}

func (s *Server) enforceExportQuota(r *http.Request) (bool, UsageResponse) {
	id, _ := auth.GetIdentity(r.Context())
	exp, _ := s.Store.Metering().SumByType(r.Context(), id.OrgID, "export")
	limits := map[string]int{"generate": s.Config.GenerateLimitPerMonth, "export": s.Config.ExportLimitPerMonth}
	used := map[string]int{"export": exp}
	blocked := exp >= limits["export"]
	return blocked, UsageResponse{OrgID: id.OrgID, Limits: limits, Used: used, Blocked: blocked}
}

func (s *Server) handleGetOrCreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"userId"`
		Email  string `json:"email"`
		Name   string `json:"name"`
	}

	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.UserID == "" || req.Email == "" {
		writeError(w, r, http.StatusBadRequest, "userId and email are required")
		return
	}

	// Try to get existing user
	user, ok, err := s.Store.Users().GetUser(r.Context(), req.UserID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to lookup user")
		return
	}

	if ok {
		// Get user's org membership
		memberships, err := s.Store.Users().ListUserOrgs(r.Context(), req.UserID)
		if err != nil {
			writeError(w, r, http.StatusInternalServerError, "failed to lookup user orgs")
			return
		}

		var org store.Organization
		var role auth.Role
		if len(memberships) > 0 {
			membership := memberships[0]
			org, err = s.Store.Organizations().GetOrganization(r.Context(), membership.OrgID)
			if err != nil {
				writeError(w, r, http.StatusInternalServerError, "failed to lookup organization")
				return
			}
			role = membership.Role
		}

		responseUser := map[string]any{
			"userId": user.ID,
			"email":  user.Email,
			"name":   user.Name,
			"orgId":  org.ID,
			"role":   role,
		}
		writeJSON(w, http.StatusOK, map[string]any{"user": responseUser})
		return
	}

	// User not found - return error so frontend can call signup
	writeJSON(w, http.StatusNotFound, map[string]any{"error": "user not found"})
}

func (s *Server) handleSignup(w http.ResponseWriter, r *http.Request) {
	log.Printf("DEBUG: handleSignup called - Method: %s, Path: %s", r.Method, r.URL.Path)

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}

	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Email == "" || req.Password == "" {
		writeError(w, r, http.StatusBadRequest, "email and password are required")
		return
	}

	// Check if user already exists
	_, exists, err := s.Store.Users().GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to check user")
		return
	}
	if exists {
		writeError(w, r, http.StatusConflict, "user already exists")
		return
	}

	// Generate user ID
	userID := newID("user")

	// Create user
	user := store.User{
		ID:    userID,
		Email: req.Email,
		Name:  req.Name,
	}

	// Create organization
	org := store.Organization{
		ID:   newID("org"),
		Name: req.Name + "'s Organization",
	}

	// Create user-org membership
	membership := store.UserOrg{
		UserID: user.ID,
		OrgID:  org.ID,
		Role:   auth.RoleOwner,
	}

	// Create all records
	if err := s.Store.Users().CreateUser(r.Context(), &user); err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to create user")
		return
	}

	if err := s.Store.Organizations().CreateOrganization(r.Context(), &org); err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to create organization")
		return
	}

	// Update membership with the actual UUIDs returned from database
	membership.UserID = user.ID
	membership.OrgID = org.ID

	if err := s.Store.Users().CreateUserOrg(r.Context(), membership); err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to create user membership")
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, org.ID, membership.Role)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to generate token")
		return
	}

	// Return user info and token
	responseUser := map[string]any{
		"userId": user.ID,
		"email":  user.Email,
		"name":   user.Name,
		"orgId":  org.ID,
		"role":   membership.Role,
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user":  responseUser,
		"token": token,
	})
}

func (s *Server) handleSignin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Email == "" {
		writeError(w, r, http.StatusBadRequest, "email is required")
		return
	}

	// Find user by email
	foundUser, ok, err := s.Store.Users().GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to lookup user")
		return
	}
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "invalid email or password")
		return
	}

	// TODO: Verify password hash (for now, we skip password check)
	// In production, you'd hash passwords with bcrypt and verify here

	// Get user's org membership
	log.Printf("DEBUG: Looking up memberships for user ID: %s", foundUser.ID)
	memberships, err := s.Store.Users().ListUserOrgs(r.Context(), foundUser.ID)
	if err != nil {
		log.Printf("ERROR: Failed to list user orgs: %v", err)
		writeError(w, r, http.StatusInternalServerError, "failed to lookup user orgs")
		return
	}
	if len(memberships) == 0 {
		log.Printf("ERROR: No memberships found for user ID: %s", foundUser.ID)
		writeError(w, r, http.StatusInternalServerError, "failed to lookup user orgs")
		return
	}

	membership := memberships[0]
	log.Printf("DEBUG: Found membership - OrgID: %s, Role: %s", membership.OrgID, membership.Role)
	org, err := s.Store.Organizations().GetOrganization(r.Context(), membership.OrgID)
	if err != nil {
		log.Printf("ERROR: Failed to get organization for OrgID %s: %v", membership.OrgID, err)
		writeError(w, r, http.StatusInternalServerError, "failed to lookup organization")
		return
	}
	log.Printf("DEBUG: Found organization: %s", org.Name)

	// Generate JWT token
	token, err := auth.GenerateToken(foundUser.ID, org.ID, membership.Role)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to generate token")
		return
	}

	responseUser := map[string]any{
		"userId": foundUser.ID,
		"email":  foundUser.Email,
		"name":   foundUser.Name,
		"orgId":  org.ID,
		"role":   membership.Role,
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user":  responseUser,
		"token": token,
	})
}

func (s *Server) handleGetMe(w http.ResponseWriter, r *http.Request) {
	// Get identity from context (set by auth middleware)
	id, ok := auth.GetIdentity(r.Context())
	if !ok {
		writeError(w, r, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Get user details
	user, ok, err := s.Store.Users().GetUser(r.Context(), id.UserID)
	if err != nil || !ok {
		writeError(w, r, http.StatusInternalServerError, "failed to get user")
		return
	}

	// Get organization
	org, err := s.Store.Organizations().GetOrganization(r.Context(), id.OrgID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to get organization")
		return
	}

	responseUser := map[string]any{
		"userId": user.ID,
		"email":  user.Email,
		"name":   user.Name,
		"orgId":  org.ID,
		"role":   id.Role,
	}

	writeJSON(w, http.StatusOK, map[string]any{"user": responseUser})
}

func (s *Server) handleListDeadLetterJobs(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.GetIdentity(r.Context())

	// Only allow admin/owner to view DLQ
	if !auth.RequireRole(id, auth.RoleAdmin) && !auth.RequireRole(id, auth.RoleOwner) {
		writeError(w, r, http.StatusForbidden, "insufficient permissions")
		return
	}

	jobs, err := s.Store.Jobs().ListDeadLetter(r.Context())
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to list dead letter jobs")
		return
	}

	// Filter jobs by organization
	var orgJobs []store.Job
	for _, job := range jobs {
		if job.OrgID == id.OrgID {
			orgJobs = append(orgJobs, job)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{"jobs": orgJobs})
}

func (s *Server) handleRetryDeadLetterJob(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.GetIdentity(r.Context())

	// Only allow admin/owner to retry DLQ jobs
	if !auth.RequireRole(id, auth.RoleAdmin) && !auth.RequireRole(id, auth.RoleOwner) {
		writeError(w, r, http.StatusForbidden, "insufficient permissions")
		return
	}

	jobID := r.PathValue("jobId")
	if jobID == "" {
		writeError(w, r, http.StatusBadRequest, "job ID is required")
		return
	}

	// Get the job to verify it exists and belongs to user's org
	job, ok, err := s.Store.Jobs().Get(r.Context(), id.OrgID, jobID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to get job")
		return
	}
	if !ok {
		writeError(w, r, http.StatusNotFound, "job not found")
		return
	}

	if job.Status != store.JobDeadLetter {
		writeError(w, r, http.StatusBadRequest, "job is not in dead letter queue")
		return
	}

	// Retry the job
	if err := s.Store.Jobs().RetryDeadLetterJob(r.Context(), jobID); err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to retry job")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"message": "job queued for retry"})
}
