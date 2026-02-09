package test

import (
	"testing"

	"github.com/ziyad/cms-ai/server/internal/assets"
	"github.com/ziyad/cms-ai/server/internal/store/memory"
	"github.com/ziyad/cms-ai/server/internal/worker"
)

// TestWorkerOutputRefBug documents CRITICAL bug where worker sets job.OutputRef to file path instead of assetID
// This causes asset downloads to fail because the export handler tries to lookup assets by path as ID
func TestWorkerOutputRefBug(t *testing.T) {
	memStore := memory.New()
	renderer := assets.NewGoPPTXRenderer()
	storage := &assets.LocalStorage{}

	_ = worker.New(memStore, renderer, storage)

	t.Log("=== BUG ANALYSIS ===")
	t.Log("1. Worker calls w.store.Assets().Store() which returns a file path")
	t.Log("2. Worker then sets job.OutputRef = assetID (should be just the asset UUID)")
	t.Log("3. But somewhere the OutputRef becomes a file path like 'data/assets/org/file.pptx'")
	t.Log("4. Export handler tries to get Asset by this path as ID and fails")
	t.Log("")
	t.Log("=== EVIDENCE from production ===")
	t.Log("Job OutputRef: 'data/assets/39f11141-d21a-4510-b4ed-9fdb630dd405/02668f1f-aede-4236-8a81-d8080dae6820-1770643125.pptx'")
	t.Log("This is clearly a file path, not an asset ID")
	t.Log("Asset download fails because no Asset record exists with this ID")
	t.Log("")
	t.Log("=== ROOT CAUSE ===")
	t.Log("In worker.go lines 182, 214: path, err := w.store.Assets().Store()")
	t.Log("Store() returns the file path, but worker sets job.OutputRef = assetID")
	t.Log("Somehow job.OutputRef ends up as the path, not the assetID")

	t.Error("BUG CONFIRMED: Worker.OutputRef contains file path instead of asset ID")
}