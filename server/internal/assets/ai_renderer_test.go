package assets

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test data helpers
func getHealthcareTestSpec() map[string]interface{} {
	return map[string]interface{}{
		"tokens": map[string]interface{}{
			"colors": map[string]interface{}{
				"primary":    "#48BB78",
				"background": "#ffffff",
				"text":       "#2D3748",
				"accent":     "#4299E1",
			},
			"company": map[string]interface{}{
				"name":        "HealthTech Corp",
				"industry":    "Healthcare Technology",
				"description": "AI-powered healthcare solutions",
			},
		},
		"constraints": map[string]interface{}{
			"safeMargin": 0.05,
		},
		"layouts": []map[string]interface{}{
			{
				"name": "Title Slide",
				"placeholders": []map[string]interface{}{
					{
						"id":      "title",
						"type":    "text",
						"content": "Healthcare Analytics Platform",
						"geometry": map[string]interface{}{
							"x": 0.1, "y": 0.3, "w": 0.8, "h": 0.15,
						},
					},
					{
						"id":      "subtitle",
						"type":    "text",
						"content": "Leveraging AI for Patient Care",
						"geometry": map[string]interface{}{
							"x": 0.1, "y": 0.5, "w": 0.8, "h": 0.1,
						},
					},
				},
			},
			{
				"name": "Content Slide",
				"placeholders": []map[string]interface{}{
					{
						"id":      "slide_title",
						"type":    "text",
						"content": "Key Features",
						"geometry": map[string]interface{}{
							"x": 0.1, "y": 0.1, "w": 0.8, "h": 0.1,
						},
					},
					{
						"id":      "content",
						"type":    "text",
						"content": "Real-time patient monitoring\nPredictive analytics\nHIPAA compliance\nEHR integration",
						"geometry": map[string]interface{}{
							"x": 0.1, "y": 0.25, "w": 0.8, "h": 0.5,
						},
					},
				},
			},
		},
	}
}

func getFinanceTestSpec() map[string]interface{} {
	return map[string]interface{}{
		"tokens": map[string]interface{}{
			"company": map[string]interface{}{
				"name":     "FinTech Solutions",
				"industry": "Financial Services",
			},
		},
		"layouts": []map[string]interface{}{
			{
				"name": "Title",
				"placeholders": []map[string]interface{}{
					{
						"id":      "title",
						"content": "Investment Portfolio Management",
					},
				},
			},
		},
	}
}

// Test PythonPPTXRenderer
func TestPythonPPTXRenderer(t *testing.T) {
	// Skip if Python renderer not available
	scriptPath := filepath.Join("..", "..", "tools", "renderer", "render_pptx.py")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Skip("Python renderer not available at", scriptPath)
	}

	renderer := &PythonPPTXRenderer{
		PythonPath: "python3",
		ScriptPath: scriptPath,
	}

	t.Run("RenderPPTX_Basic", func(t *testing.T) {
		spec := getHealthcareTestSpec()
		outputPath := filepath.Join(t.TempDir(), "test_output.pptx")

		err := renderer.RenderPPTX(context.Background(), spec, outputPath)
		require.NoError(t, err)

		// Verify file was created
		info, err := os.Stat(outputPath)
		require.NoError(t, err)
		assert.True(t, info.Size() > 1000, "PPTX file should have content")
	})

	t.Run("RenderPPTXBytes", func(t *testing.T) {
		spec := getHealthcareTestSpec()

		data, err := renderer.RenderPPTXBytes(context.Background(), spec)
		require.NoError(t, err)
		assert.True(t, len(data) > 1000, "PPTX data should have content")

		// Verify it's a valid PPTX (starts with PK like ZIP)
		assert.Equal(t, []byte("PK"), data[:2], "PPTX should be valid ZIP format")
	})

	t.Run("GenerateSlideThumbnails", func(t *testing.T) {
		spec := getHealthcareTestSpec()

		thumbnails, err := renderer.GenerateSlideThumbnails(context.Background(), spec)
		require.NoError(t, err)
		assert.Len(t, thumbnails, 2, "Should generate thumbnails for 2 slides")

		for i, thumb := range thumbnails {
			assert.True(t, len(thumb) > 100, "Thumbnail %d should have content", i)
			// Check PNG header
			assert.Equal(t, []byte{0x89, 0x50, 0x4E, 0x47}, thumb[:4], "Should be valid PNG")
		}
	})
}

// Test PythonPPTXRenderer with Company Context
func TestPythonPPTXRendererWithCompany(t *testing.T) {
	scriptPath := filepath.Join("..", "..", "tools", "renderer", "render_pptx.py")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Skip("Python renderer not available")
	}

	renderer := &PythonPPTXRenderer{
		PythonPath:        "python3",
		ScriptPath:        scriptPath,
		HuggingFaceAPIKey: os.Getenv("HUGGING_FACE_API_KEY"),
	}

	company := &CompanyContext{
		Name:        "MedTech Solutions",
		Industry:    "Healthcare",
		Personality: "AI-powered medical diagnostics improving patient outcomes through technology",
		Values:      []string{"Innovation", "Care", "Excellence"},
	}

	t.Run("RenderWithCompanyContext", func(t *testing.T) {
		spec := getHealthcareTestSpec()
		outputPath := filepath.Join(t.TempDir(), "test_with_company.pptx")

		err := renderer.RenderPPTXWithCompany(context.Background(), spec, outputPath, company)
		require.NoError(t, err)

		info, err := os.Stat(outputPath)
		require.NoError(t, err)
		assert.True(t, info.Size() > 1000, "PPTX file should have content")
	})

	t.Run("RenderWithoutCompanyFallsBack", func(t *testing.T) {
		spec := getHealthcareTestSpec()
		outputPath := filepath.Join(t.TempDir(), "test_no_company.pptx")

		// Should work even without company context
		err := renderer.RenderPPTX(context.Background(), spec, outputPath)
		require.NoError(t, err)
	})
}

// Test AIEnhancedRenderer
func TestAIEnhancedRenderer(t *testing.T) {
	scriptPath := filepath.Join("..", "..", "tools", "renderer", "render_pptx.py")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Skip("Python renderer not available")
	}

	// Create renderer without store (not needed for these tests)
	renderer := &AIEnhancedRenderer{
		pythonRenderer: &PythonPPTXRenderer{
			PythonPath: "python3",
			ScriptPath: scriptPath,
		},
		store: nil,
	}

	t.Run("ExtractCompanyContext_FromTokens", func(t *testing.T) {
		spec := getHealthcareTestSpec()
		company := renderer.extractCompanyContext(spec)

		require.NotNil(t, company)
		assert.Equal(t, "HealthTech Corp", company.Name)
		assert.Equal(t, "Healthcare Technology", company.Industry)
		assert.Equal(t, "AI-powered healthcare solutions", company.Personality)
	})

	t.Run("ExtractCompanyContext_FromJSON", func(t *testing.T) {
		spec := getHealthcareTestSpec()
		specJSON, err := json.Marshal(spec)
		require.NoError(t, err)

		company := renderer.extractCompanyContext(specJSON)
		require.NotNil(t, company)
		assert.Equal(t, "HealthTech Corp", company.Name)
	})

	t.Run("ExtractCompanyContext_FromRawMessage", func(t *testing.T) {
		spec := getHealthcareTestSpec()
		specJSON, err := json.Marshal(spec)
		require.NoError(t, err)

		var rawMsg json.RawMessage = specJSON
		company := renderer.extractCompanyContext(rawMsg)
		require.NotNil(t, company)
		assert.Equal(t, "HealthTech Corp", company.Name)
	})

	t.Run("InferIndustryFromContent_Healthcare", func(t *testing.T) {
		spec := map[string]interface{}{
			"layouts": []interface{}{
				map[string]interface{}{
					"placeholders": []interface{}{
						map[string]interface{}{
							"content": "Patient monitoring and medical diagnostics with HIPAA compliance",
						},
					},
				},
			},
		}

		company := renderer.inferCompanyFromContent(spec)
		assert.Equal(t, "Healthcare", company.Industry)
	})

	t.Run("InferIndustryFromContent_Finance", func(t *testing.T) {
		spec := map[string]interface{}{
			"layouts": []interface{}{
				map[string]interface{}{
					"placeholders": []interface{}{
						map[string]interface{}{
							"content": "Investment portfolio management with ROI tracking",
						},
					},
				},
			},
		}

		company := renderer.inferCompanyFromContent(spec)
		assert.Equal(t, "Finance", company.Industry)
	})

	t.Run("InferIndustryFromContent_Technology", func(t *testing.T) {
		spec := map[string]interface{}{
			"layouts": []interface{}{
				map[string]interface{}{
					"placeholders": []interface{}{
						map[string]interface{}{
							"content": "Cloud platform with API integration and machine learning",
						},
					},
				},
			},
		}

		company := renderer.inferCompanyFromContent(spec)
		assert.Equal(t, "Technology", company.Industry)
	})

	t.Run("InferIndustryFromContent_Education", func(t *testing.T) {
		spec := map[string]interface{}{
			"layouts": []interface{}{
				map[string]interface{}{
					"placeholders": []interface{}{
						map[string]interface{}{
							"content": "Student learning management and curriculum development",
						},
					},
				},
			},
		}

		company := renderer.inferCompanyFromContent(spec)
		assert.Equal(t, "Education", company.Industry)
	})

	t.Run("RenderPPTX_WithExtractedContext", func(t *testing.T) {
		spec := getHealthcareTestSpec()
		outputPath := filepath.Join(t.TempDir(), "test_ai_enhanced.pptx")

		err := renderer.RenderPPTX(context.Background(), spec, outputPath)
		require.NoError(t, err)

		info, err := os.Stat(outputPath)
		require.NoError(t, err)
		assert.True(t, info.Size() > 1000)
	})

	t.Run("RenderPPTX_NoCompanyContext", func(t *testing.T) {
		spec := map[string]interface{}{
			"layouts": []interface{}{
				map[string]interface{}{
					"placeholders": []interface{}{
						map[string]interface{}{
							"content": "Generic content",
						},
					},
				},
			},
		}
		outputPath := filepath.Join(t.TempDir(), "test_no_context.pptx")

		// Should still work without company context
		err := renderer.RenderPPTX(context.Background(), spec, outputPath)
		require.NoError(t, err)
	})
}

// Test helper functions
func TestHelperFunctions(t *testing.T) {
	t.Run("containsIgnoreCase", func(t *testing.T) {
		assert.True(t, containsIgnoreCase("Healthcare Platform", "health"))
		assert.True(t, containsIgnoreCase("HEALTHCARE", "health"))
		assert.True(t, containsIgnoreCase("patient health", "HEALTH"))
		assert.False(t, containsIgnoreCase("Finance", "health"))
		assert.False(t, containsIgnoreCase("", "health"))
		assert.False(t, containsIgnoreCase("test", ""))
	})

	t.Run("IndustryKeywordDetection", func(t *testing.T) {
		assert.True(t, containsHealthcareKeywords("Patient monitoring system"))
		assert.True(t, containsHealthcareKeywords("HIPAA compliant medical records"))
		assert.False(t, containsHealthcareKeywords("Financial portfolio"))

		assert.True(t, containsFinanceKeywords("Investment banking solutions"))
		assert.True(t, containsFinanceKeywords("ROI and capital management"))
		assert.False(t, containsFinanceKeywords("Patient care"))

		assert.True(t, containsTechKeywords("Cloud API platform"))
		assert.True(t, containsTechKeywords("Machine learning analytics"))
		assert.False(t, containsTechKeywords("Medical diagnosis"))

		assert.True(t, containsEducationKeywords("Student learning platform"))
		assert.True(t, containsEducationKeywords("University curriculum"))
		assert.False(t, containsEducationKeywords("Investment portfolio"))
	})
}

// Benchmark tests
func BenchmarkPythonRenderer(b *testing.B) {
	scriptPath := filepath.Join("..", "..", "tools", "renderer", "render_pptx.py")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		b.Skip("Python renderer not available")
	}

	renderer := &PythonPPTXRenderer{
		PythonPath: "python3",
		ScriptPath: scriptPath,
	}

	spec := getHealthcareTestSpec()
	tmpDir := b.TempDir()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		outputPath := filepath.Join(tmpDir, fmt.Sprintf("bench_%d.pptx", i))
		_ = renderer.RenderPPTX(context.Background(), spec, outputPath)
	}
}

func BenchmarkCompanyExtraction(b *testing.B) {
	renderer := &AIEnhancedRenderer{}
	spec := getHealthcareTestSpec()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = renderer.extractCompanyContext(spec)
	}
}

func BenchmarkIndustryInference(b *testing.B) {
	renderer := &AIEnhancedRenderer{}
	spec := map[string]interface{}{
		"layouts": []interface{}{
			map[string]interface{}{
				"placeholders": []interface{}{
					map[string]interface{}{
						"content": "Healthcare analytics platform with patient monitoring and HIPAA compliance",
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = renderer.inferCompanyFromContent(spec)
	}
}