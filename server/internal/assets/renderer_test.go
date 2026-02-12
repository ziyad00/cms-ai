package assets

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoPPTXRenderer_RenderPPTXBytes(t *testing.T) {
	renderer := NewGoPPTXRenderer()

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
	renderer := NewGoPPTXRenderer()

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

func TestSpecToJSONBytes_string_from_pgx(t *testing.T) {
	// CRITICAL: When GORM reads a jsonb column via pgx, SpecJSON (type any)
	// comes back as a Go string. specToJSONBytes must handle this without
	// double-encoding. json.Marshal(string) wraps it in quotes, producing
	// `"{\"layouts\":...}"` which is NOT a JSON object — it's a JSON string.
	// This causes: AttributeError: 'str' object has no attribute 'get'
	jsonStr := `{"layouts":[{"name":"title"}]}`

	result, err := specToJSONBytes(jsonStr)
	require.NoError(t, err)

	// Must get raw JSON bytes back, NOT a quoted string
	assert.Equal(t, jsonStr, string(result), "string spec must pass through as-is, not double-encoded")
	assert.Equal(t, byte('{'), result[0], "must start with { not quote")
}

// TDD RED: Simulates the ACTUAL production bug.
// When DeckVersion.SpecJSON is []byte, GORM writes to jsonb via json.Marshal([]byte)
// which base64-encodes it. So jsonb stores a JSON string: "eyJ0b2...".
// When pgx reads it back, SpecJSON (type any) is Go string: "eyJ0b2..." (base64).
// specToJSONBytes must detect base64 and decode it to raw JSON.
func TestSpecToJSONBytes_base64_from_pgx_roundtrip(t *testing.T) {
	// 1. Original spec as []byte (what processBindJob creates)
	originalJSON := `{"layouts":[{"name":"title","placeholders":[{"id":"t","type":"text","content":"Hello"}]}]}`

	// 2. GORM writes []byte to jsonb → json.Marshal([]byte) → base64
	base64Encoded, err := json.Marshal([]byte(originalJSON))
	require.NoError(t, err)
	// base64Encoded is: "eyJsYXlvdXRzIj..."  (a JSON string containing base64)

	// 3. pgx reads jsonb string → Go string (unwraps JSON string quotes)
	var pgxValue string
	err = json.Unmarshal(base64Encoded, &pgxValue)
	require.NoError(t, err)
	// pgxValue is: eyJsYXlvdXRzIj...  (base64 without quotes)

	// 4. specToJSONBytes must handle this and return raw JSON
	result, err := specToJSONBytes(pgxValue)
	require.NoError(t, err)
	assert.Equal(t, byte('{'), result[0], "must decode base64 to JSON object, got: %s", string(result[:50]))
	assert.JSONEq(t, originalJSON, string(result))
}

// Same bug but the string arrives WITH JSON quotes (raw jsonb text)
func TestSpecToJSONBytes_quoted_base64_string(t *testing.T) {
	originalJSON := `{"layouts":[{"name":"title"}]}`
	base64Str, _ := json.Marshal([]byte(originalJSON))
	// base64Str is: "eyJsYXlvdXRz..." (with quotes — raw jsonb text)

	result, err := specToJSONBytes(string(base64Str))
	require.NoError(t, err)
	assert.Equal(t, byte('{'), result[0], "must unwrap quoted base64 to JSON object")
}

// Base64 without padding (= stripped) should also be decoded.
func TestSpecToJSONBytes_base64_no_padding(t *testing.T) {
	originalJSON := `{"layouts":[{"name":"title"}]}`
	b64 := base64.StdEncoding.EncodeToString([]byte(originalJSON))
	// Strip padding
	b64NoPad := strings.TrimRight(b64, "=")

	result, err := specToJSONBytes(b64NoPad)
	require.NoError(t, err)
	assert.Equal(t, byte('{'), result[0], "must decode base64 without padding, got: %s", string(result[:min(50, len(result))]))
	assert.JSONEq(t, originalJSON, string(result))
}

func TestGoPPTXRenderer_RenderPPTXBytes_WithString(t *testing.T) {
	// Simulates what happens when GORM reads SpecJSON from PostgreSQL jsonb:
	// the value comes back as a Go string, not []byte or map.
	renderer := NewGoPPTXRenderer()

	specJSON := `{"layouts":[{"name":"title-slide","placeholders":[{"id":"title","type":"text","geometry":{"x":0.1,"y":0.1,"w":0.8,"h":0.2}}]}]}`

	data, err := renderer.RenderPPTXBytes(context.Background(), specJSON)
	require.NoError(t, err, "renderer must handle string spec from pgx")
	require.GreaterOrEqual(t, len(data), 4)
	assert.Equal(t, []byte{0x50, 0x4B, 0x03, 0x04}, data[:4], "must produce valid PPTX (ZIP magic)")
}

func TestGoPPTXRenderer_RenderPPTX(t *testing.T) {
	renderer := NewGoPPTXRenderer()

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
	renderer := NewGoPPTXRenderer()

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
	renderer := NewGoPPTXRenderer()

	// Create an invalid spec that can't be marshaled
	invalidSpec := make(chan int) // channels can't be marshaled to JSON

	// Test that it returns an error
	_, err := renderer.RenderPPTXBytes(context.Background(), invalidSpec)
	assert.Error(t, err)
}

func TestGoPPTXRenderer_GenerateSlideThumbnails(t *testing.T) {
	renderer := NewGoPPTXRenderer()

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
	renderer := NewGoPPTXRenderer()

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
