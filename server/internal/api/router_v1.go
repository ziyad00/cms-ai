package api

import (
	"encoding/json"
	"fmt"
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
	mux.HandleFunc("POST /v1/auth/signin", s.handleSignin)
	mux.HandleFunc("POST /v1/auth/user", s.handleGetOrCreateUser) // Legacy endpoint
	
	// Protected auth endpoint (requires auth)
	mux.HandleFunc("GET /v1/auth/me", s.handleGetMe) // Get current user from JWT

	mux.HandleFunc("POST /v1/templates/validate", s.handleValidateTemplateSpec)
	mux.HandleFunc("POST /v1/templates/generate", s.handleGenerateTemplate)
	mux.HandleFunc("GET /v1/templates", s.handleListTemplates)
	mux.HandleFunc("GET /v1/templates/{id}", s.handleGetTemplate)
	mux.HandleFunc("POST /v1/templates/{id}/versions", s.handleCreateVersion)
	mux.HandleFunc("GET /v1/templates/{id}/versions", s.handleListVersions)
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

	h := http.Handler(mux)
	h = requireJSON(h)
	h = withRequestID(h)
	// NOTE: Auth middleware temporarily disabled to allow all endpoints without JWT/headers.
	// TODO: Re-enable withAuth + skipAuthForPaths when auth is fully wired through frontend.
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

func (s *Server) handleGenerateTemplate(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.GetIdentity(r.Context())

	var req GenerateTemplateRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if strings.TrimSpace(req.Prompt) == "" {
		writeError(w, r, http.StatusBadRequest, "prompt is required")
		return
	}
	if isBlocked, usage := s.enforceQuota(r); isBlocked {
		writeJSON(w, http.StatusPaymentRequired, usage)
		return
	}

	template := store.Template{
		ID:          newID("tpl"),
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
		writeError(w, r, http.StatusInternalServerError, "failed to create template")
		return
	}

	// Generate template spec using AI
	aiReq := ai.GenerationRequest{
		Prompt:   req.Prompt,
		Language: req.Language,
		Tone:     req.Tone,
		RTL:      req.RTL,
	}

	templateSpec, aiResp, err := s.AIService.GenerateTemplateForRequest(r.Context(), id.OrgID, id.UserID, aiReq, req.BrandKitID)
	if err != nil {
		// Fall back to stub spec if AI generation fails
		templateSpec = &spec.TemplateSpec{
			Tokens: map[string]any{
				"colors": map[string]any{
					"primary":    "#3366FF",
					"background": "#FFFFFF",
					"text":       "#111111",
				},
			},
			Constraints: spec.Constraints{SafeMargin: 0.05},
			Layouts: []spec.Layout{
				{
					Name: "Title Slide",
					Placeholders: []spec.Placeholder{
						{ID: "title", Type: "text", Geometry: spec.Geometry{X: 0.1, Y: 0.3, W: 0.8, H: 0.15}},
						{ID: "subtitle", Type: "text", Geometry: spec.Geometry{X: 0.1, Y: 0.5, W: 0.8, H: 0.1}},
					},
				},
			},
		}
	}

	version := store.TemplateVersion{
		ID:        newID("ver"),
		Template:  created.ID,
		OrgID:     id.OrgID,
		VersionNo: 1,
		SpecJSON:  templateSpec,
		CreatedBy: id.UserID,
	}
	createdVer, err := s.Store.Templates().CreateVersion(r.Context(), version)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to create version")
		return
	}
	created.CurrentVersion = createdVer.ID
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
	id, _ := auth.GetIdentity(r.Context())

	tpls, err := s.Store.Templates().ListTemplates(r.Context(), id.OrgID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to list templates")
		return
	}
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
	ver := store.TemplateVersion{ID: newID("ver"), Template: tpl.ID, OrgID: tpl.OrgID, VersionNo: newNo, SpecJSON: specJSON, CreatedBy: id.UserID}
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
	newV := store.TemplateVersion{ID: newID("ver"), Template: tpl.ID, OrgID: tpl.OrgID, VersionNo: newNo, SpecJSON: req.Spec, CreatedBy: id.UserID}
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
	created, wasDuplicate, _ := s.Store.Jobs().EnqueueWithDeduplication(r.Context(), job)
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
	createdJob, wasDuplicate, _ := s.Store.Jobs().EnqueueWithDeduplication(r.Context(), job)
	if wasDuplicate {
		writeJSON(w, http.StatusAccepted, map[string]any{"job": createdJob, "duplicate": true})
		return
	}

	assetID := newID("asset")
	assetPath := assetID + ".pptx"

	// Render to temporary file first
	tempPath := filepath.Join(os.TempDir(), assetID+".pptx")
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

	_, err = s.ObjectStorage.Upload(r.Context(), assetPath, data, "application/vnd.openxmlformats-officedocument.presentationml.presentation")
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to upload asset")
		return
	}

	asset := store.Asset{ID: assetID, OrgID: id.OrgID, Type: store.AssetPPTX, Path: assetPath, Mime: "application/vnd.openxmlformats-officedocument.presentationml.presentation"}
	createdAsset, _ := s.Store.Assets().Create(r.Context(), asset)

	createdJob.Status = store.JobDone
	createdJob.OutputRef = createdAsset.ID
	s.Store.Jobs().Update(r.Context(), createdJob)
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
	if err := s.Store.Users().CreateUser(r.Context(), user); err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to create user")
		return
	}

	if err := s.Store.Organizations().CreateOrganization(r.Context(), org); err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to create organization")
		return
	}

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
	memberships, err := s.Store.Users().ListUserOrgs(r.Context(), foundUser.ID)
	if err != nil || len(memberships) == 0 {
		writeError(w, r, http.StatusInternalServerError, "failed to lookup user orgs")
		return
	}

	membership := memberships[0]
	org, err := s.Store.Organizations().GetOrganization(r.Context(), membership.OrgID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to lookup organization")
		return
	}

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
