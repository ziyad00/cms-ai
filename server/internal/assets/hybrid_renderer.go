package assets

import (
	"context"
	"strings"
)

// HybridRenderer chooses between Go and Python renderers based on requirements
type HybridRenderer struct {
	goRenderer     *GoPPTXRenderer
	pythonRenderer *PythonPPTXRenderer
}

func NewHybridRenderer() *HybridRenderer {
	return &HybridRenderer{
		goRenderer: NewGoPPTXRenderer(),
		pythonRenderer: &PythonPPTXRenderer{
			PythonPath: "/usr/bin/python3", // System Python with python-pptx
			ScriptPath: "tools/renderer/render_pptx.py",
		},
	}
}

// RenderPPTX chooses the best renderer for the presentation requirements
func (h *HybridRenderer) RenderPPTX(ctx context.Context, spec any, outPath string) error {
	renderer := h.selectRenderer(spec)
	return renderer.RenderPPTX(ctx, spec, outPath)
}

// RenderPPTXBytes chooses the best renderer and returns bytes
func (h *HybridRenderer) RenderPPTXBytes(ctx context.Context, spec any) ([]byte, error) {
	renderer := h.selectRenderer(spec)
	return renderer.RenderPPTXBytes(ctx, spec)
}

// GenerateSlideThumbnails generates thumbnails using the appropriate renderer
func (h *HybridRenderer) GenerateSlideThumbnails(ctx context.Context, spec any) ([][]byte, error) {
	renderer := h.selectRenderer(spec)
	return renderer.GenerateSlideThumbnails(ctx, spec)
}

// selectRenderer determines which renderer to use based on presentation requirements
func (h *HybridRenderer) selectRenderer(spec any) Renderer {
	if h.requiresRichVisuals(spec) {
		return h.pythonRenderer
	}
	return h.goRenderer
}

// requiresRichVisuals determines if presentation needs rich visual features
func (h *HybridRenderer) requiresRichVisuals(spec any) bool {
	specBytes, err := specToJSONBytes(spec)
	if err != nil {
		return false // Default to Go if can't parse
	}

	specStr := strings.ToLower(string(specBytes))

	// Use Python renderer for:
	// 1. Industry-specific themes that benefit from rich backgrounds
	// 2. Presentations with multiple slides (more visual impact needed)
	// 3. Content that mentions visual/design elements
	// 4. Specific industries known to require professional visuals

	industryKeywords := []string{
		"healthcare", "medical", "finance", "financial", "banking",
		"technology", "tech", "software", "security", "cyber",
		"education", "learning", "corporate", "professional",
	}

	visualKeywords := []string{
		"presentation", "visual", "design", "theme", "background",
		"professional", "corporate", "showcase", "innovation",
	}

	// Check for industry-specific content
	for _, keyword := range industryKeywords {
		if strings.Contains(specStr, keyword) {
			return true
		}
	}

	// Check for visual-focused content
	for _, keyword := range visualKeywords {
		if strings.Contains(specStr, keyword) {
			return true
		}
	}

	// Check number of layouts - use Python for multi-slide presentations
	layoutCount := strings.Count(specStr, `"name"`)
	if layoutCount > 2 {
		return true
	}

	// Default to Go renderer for simple presentations
	return false
}

// ForceGoRenderer returns the Go renderer for cases where Python isn't available
func (h *HybridRenderer) ForceGoRenderer() Renderer {
	return h.goRenderer
}

// ForcePythonRenderer returns the Python renderer for cases requiring rich visuals
func (h *HybridRenderer) ForcePythonRenderer() Renderer {
	return h.pythonRenderer
}