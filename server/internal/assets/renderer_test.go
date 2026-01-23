package assets

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoPPTXRenderer_RenderPPTXBytes(t *testing.T) {
	renderer := GoPPTXRenderer{}

	// Create a simple template spec
	templateSpec := map[string]interface{}{
		"tokens": map[string]interface{}{
			"colors": map[string]interface{}{
				"primary":    "#0078d4",
				"secondary":  "#107c10",
				"background": "#ffffff",
				"text":       "#323130",
			},
		},
		"layouts": []map[string]interface{}{
			{
				"name": "title-slide",
				"placeholders": []map[string]interface{}{
					{
						"id":   "title",
						"type": "text",
						"geometry": map[string]interface{}{
							"x": 0.1,
							"y": 0.1,
							"w": 0.8,
							"h": 0.2,
						},
					},
					{
						"id":   "subtitle",
						"type": "text",
						"geometry": map[string]interface{}{
							"x": 0.1,
							"y": 0.4,
							"w": 0.8,
							"h": 0.1,
						},
					},
				},
			},
		},
	}

	// Test RenderPPTXBytes
	data, err := renderer.RenderPPTXBytes(context.Background(), templateSpec)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Verify it's a valid PPTX by checking file signature
	assert.Equal(t, []byte{0x50, 0x4B, 0x03, 0x04}, data[:4]) // ZIP signature
}

func TestGoPPTXRenderer_RenderPPTXBytes_WithJSONBytes(t *testing.T) {
	renderer := GoPPTXRenderer{}

	// Same as other tests, but passed as raw JSON bytes.
	templateSpec := map[string]interface{}{
		"tokens": map[string]interface{}{
			"colors": map[string]interface{}{
				"primary":    "#0078d4",
				"secondary":  "#107c10",
				"background": "#ffffff",
				"text":       "#323130",
			},
		},
		"layouts": []map[string]interface{}{
			{
				"name": "title-slide",
				"placeholders": []map[string]interface{}{
					{
						"id":   "title",
						"type": "text",
						"geometry": map[string]interface{}{
							"x": 0.1,
							"y": 0.1,
							"w": 0.8,
							"h": 0.2,
						},
					},
				},
			},
		},
	}

	b, err := json.Marshal(templateSpec)
	require.NoError(t, err)

	data, err := renderer.RenderPPTXBytes(context.Background(), b)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(data), 4)
	assert.Equal(t, []byte{0x50, 0x4B, 0x03, 0x04}, data[:4])
}

func TestGoPPTXRenderer_RenderPPTX(t *testing.T) {
	renderer := GoPPTXRenderer{}

	// Create a simple template spec
	templateSpec := map[string]interface{}{
		"tokens": map[string]interface{}{
			"colors": map[string]interface{}{
				"primary":    "#0078d4",
				"secondary":  "#107c10",
				"background": "#ffffff",
				"text":       "#323130",
			},
		},
		"layouts": []map[string]interface{}{
			{
				"name": "title-slide",
				"placeholders": []map[string]interface{}{
					{
						"id":   "title",
						"type": "text",
						"geometry": map[string]interface{}{
							"x": 0.1,
							"y": 0.1,
							"w": 0.8,
							"h": 0.2,
						},
					},
				},
			},
		},
	}

	// Create temp output path
	tempFile, err := os.CreateTemp("", "test-*.pptx")
	require.NoError(t, err)
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	// Test RenderPPTX
	err = renderer.RenderPPTX(context.Background(), templateSpec, tempFile.Name())
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(tempFile.Name())
	assert.NoError(t, err)

	// Read and verify content
	data, err := os.ReadFile(tempFile.Name())
	require.NoError(t, err)
	assert.Equal(t, []byte{0x50, 0x4B, 0x03, 0x04}, data[:4]) // ZIP signature
}

func TestGoPPTXRenderer_EmptyLayouts(t *testing.T) {
	renderer := GoPPTXRenderer{}

	// Create a template spec with no layouts
	templateSpec := map[string]interface{}{
		"tokens": map[string]interface{}{
			"colors": map[string]interface{}{
				"primary": "#0078d4",
			},
		},
		"layouts": []map[string]interface{}{},
	}

	// Test that it returns an error
	_, err := renderer.RenderPPTXBytes(context.Background(), templateSpec)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no layouts found")
}

func TestGoPPTXRenderer_InvalidJSON(t *testing.T) {
	renderer := GoPPTXRenderer{}

	// Create an invalid spec that can't be marshaled
	invalidSpec := make(chan int) // channels can't be marshaled to JSON

	// Test that it returns an error
	_, err := renderer.RenderPPTXBytes(context.Background(), invalidSpec)
	assert.Error(t, err)
}

func TestGoPPTXRenderer_GenerateSlideThumbnails(t *testing.T) {
	renderer := GoPPTXRenderer{}

	// Create a template spec with multiple layouts
	templateSpec := map[string]interface{}{
		"layouts": []map[string]interface{}{
			{
				"name": "title-slide",
				"placeholders": []map[string]interface{}{
					{
						"id":   "title",
						"type": "text",
						"geometry": map[string]interface{}{
							"x": 0.1,
							"y": 0.1,
							"w": 0.8,
							"h": 0.2,
						},
					},
					{
						"id":   "subtitle",
						"type": "text",
						"geometry": map[string]interface{}{
							"x": 0.1,
							"y": 0.4,
							"w": 0.6,
							"h": 0.1,
						},
					},
				},
			},
			{
				"name": "content-slide",
				"placeholders": []map[string]interface{}{
					{
						"id":   "content",
						"type": "text",
						"geometry": map[string]interface{}{
							"x": 0.05,
							"y": 0.05,
							"w": 0.9,
							"h": 0.8,
						},
					},
				},
			},
		},
	}

	// Test thumbnail generation
	thumbnails, err := renderer.GenerateSlideThumbnails(context.Background(), templateSpec)
	require.NoError(t, err)
	assert.Len(t, thumbnails, 2) // Should have 2 thumbnails for 2 layouts

	// Verify each thumbnail is a valid PNG
	for i, thumbnail := range thumbnails {
		assert.NotEmpty(t, thumbnail, fmt.Sprintf("thumbnail %d should not be empty", i))

		// Check PNG signature
		if len(thumbnail) >= 8 {
			assert.Equal(t, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, thumbnail[:8])
		}
	}
}

func TestGoPPTXRenderer_GenerateSlideThumbnails_EmptyLayouts(t *testing.T) {
	renderer := GoPPTXRenderer{}

	// Create a template spec with no layouts
	templateSpec := map[string]interface{}{
		"layouts": []map[string]interface{}{},
	}

	// Test that it returns an error
	thumbnails, err := renderer.GenerateSlideThumbnails(context.Background(), templateSpec)
	assert.Error(t, err)
	assert.Nil(t, thumbnails)
	assert.Contains(t, err.Error(), "no layouts found")
}

func TestPythonPPTXRenderer_GenerateSlideThumbnails(t *testing.T) {
	renderer := PythonPPTXRenderer{}

	// Create a template spec with multiple layouts
	templateSpec := map[string]interface{}{
		"layouts": []map[string]interface{}{
			{"name": "title-slide"},
			{"name": "content-slide"},
			{"name": "summary-slide"},
		},
	}

	// Test thumbnail generation
	thumbnails, err := renderer.GenerateSlideThumbnails(context.Background(), templateSpec)
	require.NoError(t, err)
	assert.Len(t, thumbnails, 3) // Should have 3 thumbnails for 3 layouts

	// Verify each thumbnail is a valid PNG
	for i, thumbnail := range thumbnails {
		assert.NotEmpty(t, thumbnail, fmt.Sprintf("thumbnail %d should not be empty", i))

		// Check PNG signature
		if len(thumbnail) >= 8 {
			assert.Equal(t, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, thumbnail[:8])
		}
	}
}
