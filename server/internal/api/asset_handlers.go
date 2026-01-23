package api

import (
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/ziyad/cms-ai/server/internal/auth"
	"github.com/ziyad/cms-ai/server/internal/store"
)

// handleAssetDownload handles GET /v1/assets/{id}
func (s *Server) handleAssetDownload(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.GetIdentity(r.Context())
	assetID := r.PathValue("id")
	asset, ok, err := s.Store.Assets().Get(r.Context(), id.OrgID, assetID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed")
		return
	}
	if !ok {
		writeError(w, r, http.StatusNotFound, "not found")
		return
	}

	// Try to get signed URL first.
	// If the storage returns a relative URL (local storage), don't redirect because
	// the API server is not serving that path; instead stream the bytes directly.
	signedURL, err := s.ObjectStorage.GetURL(r.Context(), asset.Path, 15*time.Minute)
	if err == nil {
		if strings.HasPrefix(signedURL, "http://") || strings.HasPrefix(signedURL, "https://") {
			// Redirect to a real signed URL (S3, etc.)
			http.Redirect(w, r, signedURL, http.StatusTemporaryRedirect)
			return
		}
	}

	// Fallback: direct download
	data, err := s.ObjectStorage.Download(r.Context(), asset.Path)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to download asset")
		return
	}

	// Determine appropriate filename based on asset type
	filename := assetID
	switch asset.Type {
	case store.AssetPPTX:
		filename += ".pptx"
	case store.AssetPNG:
		filename += ".png"
	default:
		// For generic files, try to extract from path or use generic extension
		if ext := filepath.Ext(asset.Path); ext != "" {
			filename += ext
		} else {
			filename += ".bin"
		}
	}

	w.Header().Set("Content-Type", asset.Mime)
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Write(data)
}

// handleJobAssetDownload handles GET /v1/jobs/{jobId}/assets/{filename}
// This provides an alternative way to download assets using the job ID and filename
func (s *Server) handleJobAssetDownload(w http.ResponseWriter, r *http.Request) {
	id, _ := auth.GetIdentity(r.Context())
	jobID := r.PathValue("jobId")
	filename := r.PathValue("filename")

	// Get the job to find the associated asset
	job, ok, err := s.Store.Jobs().Get(r.Context(), id.OrgID, jobID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed")
		return
	}
	if !ok {
		writeError(w, r, http.StatusNotFound, "job not found")
		return
	}

	// Check if job is complete and has output
	if job.Status != store.JobDone || job.OutputRef == "" {
		writeError(w, r, http.StatusNotFound, "asset not ready")
		return
	}

	// Get the asset
	asset, ok, err := s.Store.Assets().Get(r.Context(), id.OrgID, job.OutputRef)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed")
		return
	}
	if !ok {
		writeError(w, r, http.StatusNotFound, "asset not found")
		return
	}

	// Verify filename matches (optional security check)
	expectedFilename := job.OutputRef
	switch asset.Type {
	case store.AssetPPTX:
		expectedFilename += ".pptx"
	case store.AssetPNG:
		expectedFilename += ".png"
	}

	// Allow partial matches or exact matches
	if !strings.HasPrefix(filename, expectedFilename[:8]) { // Match by asset ID prefix
		writeError(w, r, http.StatusNotFound, "filename mismatch")
		return
	}

	// Try to get signed URL first.
	// If the storage returns a relative URL (local storage), don't redirect because
	// the API server is not serving that path; instead stream the bytes directly.
	signedURL, err := s.ObjectStorage.GetURL(r.Context(), asset.Path, 15*time.Minute)
	if err == nil {
		if strings.HasPrefix(signedURL, "http://") || strings.HasPrefix(signedURL, "https://") {
			// Redirect to a real signed URL (S3, etc.)
			http.Redirect(w, r, signedURL, http.StatusTemporaryRedirect)
			return
		}
	}

	// Fallback: direct download
	data, err := s.ObjectStorage.Download(r.Context(), asset.Path)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to download asset")
		return
	}

	w.Header().Set("Content-Type", asset.Mime)
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Write(data)
}
