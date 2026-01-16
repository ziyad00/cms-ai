package assets

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ziyad/cms-ai/server/internal/spec"
)

// Renderer interface for rendering templates to PPTX
type Renderer interface {
	// RenderPPTX renders a template spec to a PPTX file at the given path
	RenderPPTX(ctx context.Context, spec *spec.TemplateSpec, outputPath string) error
	
	// RenderPPTXBytes renders a template spec to PPTX and returns the bytes
	RenderPPTXBytes(ctx context.Context, spec *spec.TemplateSpec) ([]byte, error)
	
	// GenerateSlideThumbnails generates thumbnail images for each slide
	GenerateSlideThumbnails(ctx context.Context, spec *spec.TemplateSpec) ([][]byte, error)
}

// GoPPTXRenderer uses the Python renderer script to generate PPTX files
type GoPPTXRenderer struct{}

// RenderPPTX renders a template spec to a PPTX file
func (r GoPPTXRenderer) RenderPPTX(ctx context.Context, spec *spec.TemplateSpec, outputPath string) error {
	// Use Python renderer script
	rendererPath := filepath.Join("/app/tools/renderer/render_pptx.py")
	
	// Create a temporary JSON file with the spec
	// For now, we'll pass it via stdin or use a temp file
	// This is a stub implementation - the actual Python script should handle the rendering
	cmd := exec.CommandContext(ctx, "python3", rendererPath, outputPath)
	
	// TODO: Pass spec JSON to the Python script
	// For now, this is a placeholder
	return cmd.Run()
}

// RenderPPTXBytes renders to memory
func (r GoPPTXRenderer) RenderPPTXBytes(ctx context.Context, spec *spec.TemplateSpec) ([]byte, error) {
	// Create temp file, render, read, delete
	tempPath := filepath.Join("/tmp", "render-temp.pptx")
	if err := r.RenderPPTX(ctx, spec, tempPath); err != nil {
		return nil, err
	}
	
	// Read the file
	data, err := os.ReadFile(tempPath)
	if err != nil {
		return nil, err
	}
	
	// Clean up
	os.Remove(tempPath)
	return data, nil
}

// GenerateSlideThumbnails generates thumbnails for slides
func (r GoPPTXRenderer) GenerateSlideThumbnails(ctx context.Context, spec *spec.TemplateSpec) ([][]byte, error) {
	// TODO: Implement thumbnail generation
	// For now, return empty slice
	return [][]byte{}, nil
}
