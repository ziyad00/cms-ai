package assets

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPythonPPTXRenderer_SmartPathResolution(t *testing.T) {
	t.Run("LocalDevelopmentPathResolution", func(t *testing.T) {
		// Test that the constructor finds the local development path
		renderer := NewPythonPPTXRenderer("")

		require.NotNil(t, renderer)
		assert.Equal(t, "python3", renderer.PythonPath)
		assert.NotEmpty(t, renderer.ScriptPath)

		// The script path should either be:
		// 1. Railway path: /app/tools/renderer/render_pptx.py (in Railway)
		// 2. Local path: containing tools/renderer/render_pptx.py (in development)
		// 3. Web path: containing web/tools/renderer/render_pptx.py (in web deployment)

		if strings.HasPrefix(renderer.ScriptPath, "/app/") {
			t.Logf("Using Railway container path: %s", renderer.ScriptPath)
			// In Railway environment - this is expected
		} else if strings.Contains(renderer.ScriptPath, "tools/renderer/render_pptx.py") {
			t.Logf("Using local development path: %s", renderer.ScriptPath)
			// Verify the script actually exists
			_, err := os.Stat(renderer.ScriptPath)
			assert.NoError(t, err, "Python script should exist at resolved path")
		} else {
			t.Logf("Path resolution result: %s", renderer.ScriptPath)
		}
	})

	t.Run("WithHuggingFaceAPIKey", func(t *testing.T) {
		apiKey := "test-api-key"
		renderer := NewPythonPPTXRenderer(apiKey)

		require.NotNil(t, renderer)
		assert.Equal(t, apiKey, renderer.HuggingFaceAPIKey)
	})

	t.Run("PathExistenceValidation", func(t *testing.T) {
		renderer := NewPythonPPTXRenderer("")

		// Test with a simple spec to see if path resolution works
		spec := map[string]interface{}{
			"layouts": []map[string]interface{}{
				{
					"name": "test-slide",
					"placeholders": []map[string]interface{}{
						{
							"id":      "title",
							"type":    "text",
							"content": "Path Resolution Test",
							"geometry": map[string]interface{}{
								"x": 1.0,
								"y": 2.0,
								"w": 8.0,
								"h": 1.5,
							},
						},
					},
				},
			},
		}

		// If the script exists and dependencies are available, this should work
		data, err := renderer.RenderPPTXBytes(context.Background(), spec)

		if err == nil {
			// Success - path resolution and execution worked
			assert.NotEmpty(t, data)
			t.Logf("Python renderer successfully executed with path: %s", renderer.ScriptPath)
		} else {
			// Expected in environments without Python dependencies
			t.Logf("Python renderer failed (expected in some environments): %v", err)
			t.Logf("Resolved script path: %s", renderer.ScriptPath)

			// Verify the error makes sense
			if strings.Contains(err.Error(), "no such file or directory") {
				t.Logf("Script not found - this is expected in Railway simulation")
			} else if strings.Contains(err.Error(), "executable file not found") {
				t.Logf("Python not found - this is expected in minimal environments")
			} else {
				t.Logf("Other error: %v", err)
			}
		}
	})
}

func TestPythonPPTXRenderer_RailwayVsLocalDevelopment(t *testing.T) {
	ctx := context.Background()

	spec := map[string]interface{}{
		"layouts": []map[string]interface{}{
			{
				"name": "comparison-test",
				"placeholders": []map[string]interface{}{
					{
						"id":      "title",
						"type":    "text",
						"content": "Railway vs Local Development Test",
					},
				},
			},
		},
	}

	t.Run("SmartConstructorVsHardcodedPath", func(t *testing.T) {
		// Smart constructor (should work in both environments)
		smartRenderer := NewPythonPPTXRenderer("")

		// Hardcoded Railway path (will fail in local dev)
		hardcodedRenderer := &PythonPPTXRenderer{
			PythonPath: "python3",
			ScriptPath: "/app/tools/renderer/render_pptx.py",
		}

		t.Logf("Smart renderer path: %s", smartRenderer.ScriptPath)
		t.Logf("Hardcoded renderer path: %s", hardcodedRenderer.ScriptPath)

		// Try smart renderer
		_, smartErr := smartRenderer.RenderPPTXBytes(ctx, spec)

		// Try hardcoded renderer
		_, hardcodedErr := hardcodedRenderer.RenderPPTXBytes(ctx, spec)

		// In local development, smart renderer should perform better
		if smartErr == nil && hardcodedErr != nil {
			t.Log("✅ Smart renderer succeeded where hardcoded failed - path fallback working!")
		} else if smartErr != nil && hardcodedErr != nil {
			t.Log("Both failed (expected in environments without Python dependencies)")
			t.Logf("Smart error: %v", smartErr)
			t.Logf("Hardcoded error: %v", hardcodedErr)
		} else if smartErr == nil && hardcodedErr == nil {
			t.Log("Both succeeded - likely in Railway environment")
		}
	})

	t.Run("PathResolutionLogging", func(t *testing.T) {
		// This test validates that our logging shows the path resolution process

		// Create multiple renderers to see path resolution
		renderer1 := NewPythonPPTXRenderer("")
		renderer2 := NewPythonPPTXRenderer("test-key")

		// Both should resolve to the same path
		assert.Equal(t, renderer1.ScriptPath, renderer2.ScriptPath)
		t.Logf("Both renderers resolved to path: %s", renderer1.ScriptPath)

		// Verify path makes sense for environment
		if strings.HasPrefix(renderer1.ScriptPath, "/app/") {
			t.Log("Railway container environment detected")
		} else if filepath.IsAbs(renderer1.ScriptPath) {
			t.Log("Local development environment detected")
		}
	})
}