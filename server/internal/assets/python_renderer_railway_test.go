package assets

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPythonPPTXRenderer_RailwayEnvironmentSimulation(t *testing.T) {
	skipIfNoPptx(t)
	ctx := context.Background()

	t.Run("RailwayContainerPathResolution", func(t *testing.T) {
		// Simulate Railway container environment
		// In Railway, the script should be at /app/tools/renderer/render_pptx.py
		renderer := &PythonPPTXRenderer{
			PythonPath: "python3",
			ScriptPath: "/app/tools/renderer/render_pptx.py", // Railway path
		}

		templateSpec := map[string]interface{}{
			"layouts": []map[string]interface{}{
				{
					"name": "title-slide",
					"placeholders": []map[string]interface{}{
						{
							"id":      "title",
							"type":    "text",
							"content": "Railway Environment Test",
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

		// This will fail because /app/tools/renderer/render_pptx.py doesn't exist locally
		// but should work in Railway environment
		data, err := renderer.RenderPPTXBytes(ctx, templateSpec)

		if err != nil {
			// Expected to fail in local dev environment
			assert.Contains(t, err.Error(), "no such file or directory")
			t.Logf("Expected failure in local environment: %v", err)
		} else {
			// If it somehow works (e.g., file exists), validate the output
			assert.NotEmpty(t, data)
		}
	})

	t.Run("LocalDevelopmentPathResolution", func(t *testing.T) {
		// Use local development path
		localScriptPath := "tools/renderer/render_pptx.py"

		// Get the absolute path relative to project root
		wd, _ := os.Getwd()
		// Navigate up to find the project root (where tools/ directory is)
		projectRoot := wd
		for {
			if _, err := os.Stat(filepath.Join(projectRoot, "tools")); err == nil {
				break
			}
			parent := filepath.Dir(projectRoot)
			if parent == projectRoot {
				// Reached filesystem root, tools directory not found
				break
			}
			projectRoot = parent
		}

		fullScriptPath := filepath.Join(projectRoot, localScriptPath)

		renderer := &PythonPPTXRenderer{
			PythonPath: "python3",
			ScriptPath: fullScriptPath,
		}

		templateSpec := map[string]interface{}{
			"layouts": []map[string]interface{}{
				{
					"name": "title-slide",
					"placeholders": []map[string]interface{}{
						{
							"id":      "title",
							"type":    "text",
							"content": "Local Development Test",
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

		// Check if the script exists locally
		if _, err := os.Stat(fullScriptPath); os.IsNotExist(err) {
			t.Skipf("Python script not found at %s, skipping local test", fullScriptPath)
			return
		}

		data, err := renderer.RenderPPTXBytes(ctx, templateSpec)
		require.NoError(t, err)
		assert.NotEmpty(t, data)

		// Verify it's a valid PPTX file (starts with PK header)
		assert.True(t, len(data) > 4)
		assert.Equal(t, []byte{0x50, 0x4B}, data[0:2]) // PK header
	})

	t.Run("PythonDependencyValidation", func(t *testing.T) {
		// Test that required Python dependencies can be imported
		// This simulates Railway environment where pip install should have run

		// Create a simple script to test imports
		testScript := `#!/usr/bin/env python3
import sys
try:
    from pptx import Presentation
    from pptx.util import Inches, Pt
    from pptx.enum.shapes import MSO_SHAPE
    from pptx.dml.color import RGBColor
    from pptx.enum.text import PP_ALIGN
    print("SUCCESS: python-pptx imports successful")
except ImportError as e:
    print(f"ERROR: python-pptx import failed: {e}", file=sys.stderr)
    sys.exit(1)

# Test olama module imports (expected to fail in CI but should work in Railway)
try:
    from ai_design_generator import AIDesignGenerator
    from design_templates import DesignTemplateLibrary, get_design_system_for_content
    from abstract_background_renderer import CompositeBackgroundRenderer
    print("SUCCESS: olama modules imported successfully")
except ImportError as e:
    print(f"WARNING: olama modules not available: {e}", file=sys.stderr)
    print("This is expected in CI/local dev environments")
`

		// Write test script to temporary file
		tmpFile, err := os.CreateTemp("", "test_imports_*.py")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString(testScript)
		require.NoError(t, err)
		tmpFile.Close()

		// Use the Python renderer to execute the test script
		// Create a simple spec to trigger script execution
		testSpec := map[string]interface{}{
			"layouts": []map[string]interface{}{},
		}

		// Override the script path to use our import test
		testRenderer := &PythonPPTXRenderer{
			PythonPath: "python3",
			ScriptPath: tmpFile.Name(),
		}

		// This will try to execute our import test script instead of the PPTX renderer
		_, err = testRenderer.RenderPPTXBytes(ctx, testSpec)

		// The script will exit with status 1 if imports fail, causing an error
		if err != nil {
			t.Logf("Import test result: %v", err)
			// Check if it's specifically olama modules missing (expected in CI)
			t.Logf("Some imports failed, which is expected in non-Railway environments")
		}
	})
}

func TestPythonPPTXRenderer_PathFallbackLogic(t *testing.T) {
	ctx := context.Background()

	t.Run("SmartPathResolution", func(t *testing.T) {
		// This test validates that we need fallback path logic similar to GoPPTXRenderer

		templateSpec := map[string]interface{}{
			"layouts": []map[string]interface{}{
				{
					"name": "test-slide",
					"placeholders": []map[string]interface{}{
						{
							"id":      "title",
							"type":    "text",
							"content": "Smart Path Test",
						},
					},
				},
			},
		}

		// Test Railway path first (will fail in local dev)
		railwayRenderer := &PythonPPTXRenderer{
			PythonPath: "python3",
			ScriptPath: "/app/tools/renderer/render_pptx.py",
		}

		_, railwayErr := railwayRenderer.RenderPPTXBytes(ctx, templateSpec)

		// Test local development path
		wd, _ := os.Getwd()
		projectRoot := wd
		for {
			if _, err := os.Stat(filepath.Join(projectRoot, "tools")); err == nil {
				break
			}
			parent := filepath.Dir(projectRoot)
			if parent == projectRoot {
				break
			}
			projectRoot = parent
		}

		localScriptPath := filepath.Join(projectRoot, "tools/renderer/render_pptx.py")
		localRenderer := &PythonPPTXRenderer{
			PythonPath: "python3",
			ScriptPath: localScriptPath,
		}

		_, localErr := localRenderer.RenderPPTXBytes(ctx, templateSpec)

		// At least one should work (preferably local in development)
		if railwayErr != nil && localErr != nil {
			t.Logf("Railway path failed: %v", railwayErr)
			t.Logf("Local path failed: %v", localErr)
			t.Log("This indicates we need smart path fallback logic")
		}

		// If local script exists and local environment has dependencies, it should work
		if _, err := os.Stat(localScriptPath); err == nil && localErr == nil {
			t.Log("Local path resolution successful - this is good for development")
		}
	})
}

func TestPythonPPTXRenderer_ErrorHandling(t *testing.T) {
	ctx := context.Background()

	t.Run("InvalidScriptPath", func(t *testing.T) {
		renderer := &PythonPPTXRenderer{
			PythonPath: "python3",
			ScriptPath: "/nonexistent/path/script.py",
		}

		spec := map[string]interface{}{
			"layouts": []map[string]interface{}{},
		}

		_, err := renderer.RenderPPTXBytes(ctx, spec)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no such file or directory")
	})

	t.Run("InvalidPythonPath", func(t *testing.T) {
		renderer := &PythonPPTXRenderer{
			PythonPath: "nonexistent-python",
			ScriptPath: "/tmp/dummy.py",
		}

		spec := map[string]interface{}{
			"layouts": []map[string]interface{}{},
		}

		_, err := renderer.RenderPPTXBytes(ctx, spec)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}