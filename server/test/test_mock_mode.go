package test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ziyad/cms-ai/server/internal/ai"
	"github.com/ziyad/cms-ai/server/internal/spec"
)

func TestMockOrchestrator(t *testing.T) {
	// Enable mock mode
	os.Setenv("USE_MOCK_AI", "true")
	defer os.Unsetenv("USE_MOCK_AI")

	orchestrator := ai.NewOrchestrator()

	t.Run("Healthcare_Detection", func(t *testing.T) {
		req := ai.GenerationRequest{
			Prompt: "Create a healthcare presentation for patient monitoring system",
			ContentData: map[string]interface{}{
				"company":     "MedTech Solutions",
				"product":     "Patient Monitor Pro",
				"compliance":  "HIPAA compliant",
				"features":    []string{"Real-time monitoring", "Alert system", "EHR integration"},
			},
		}

		resp, err := orchestrator.GenerateTemplateSpec(context.Background(), req)
		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotNil(t, resp.Spec)

		// Check that it's a mock response (no cost)
		assert.Equal(t, 0.0, resp.Cost)
		assert.Equal(t, "mock", resp.Model)

		// Verify healthcare theme was detected
		tokens := resp.Spec.Tokens
		company, ok := tokens["company"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "Healthcare", company["industry"])

		// Verify healthcare colors
		colors, ok := tokens["colors"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "#48BB78", colors["primary"]) // Healthcare green
	})

	t.Run("Finance_Detection", func(t *testing.T) {
		req := ai.GenerationRequest{
			Prompt: "Quarterly financial report presentation",
			ContentData: map[string]interface{}{
				"company":  "FinanceCorr",
				"quarter":  "Q4 2024",
				"revenue":  "$25M",
				"profit":   "$5M",
				"growth":   "15%",
			},
		}

		resp, err := orchestrator.GenerateTemplateSpec(context.Background(), req)
		require.NoError(t, err)

		tokens := resp.Spec.Tokens
		company, _ := tokens["company"].(map[string]interface{})
		assert.Equal(t, "Finance", company["industry"])

		colors, _ := tokens["colors"].(map[string]interface{})
		assert.Equal(t, "#1B5E20", colors["primary"]) // Finance green
	})

	t.Run("Technology_Detection", func(t *testing.T) {
		req := ai.GenerationRequest{
			Prompt: "Cloud API platform presentation for software developers",
			ContentData: map[string]interface{}{
				"company": "TechCorp",
				"product": "API Gateway",
				"features": []string{
					"RESTful APIs",
					"GraphQL support",
					"Cloud deployment",
					"Machine learning",
				},
			},
		}

		resp, err := orchestrator.GenerateTemplateSpec(context.Background(), req)
		require.NoError(t, err)

		tokens := resp.Spec.Tokens
		company, _ := tokens["company"].(map[string]interface{})
		assert.Equal(t, "Technology", company["industry"])

		colors, _ := tokens["colors"].(map[string]interface{})
		assert.Equal(t, "#667EEA", colors["primary"]) // Tech purple
	})

	t.Run("Content_Population", func(t *testing.T) {
		req := ai.GenerationRequest{
			Prompt: "Sales presentation",
			ContentData: map[string]interface{}{
				"title":    "Sales Strategy 2025",
				"subtitle": "Growing Our Market Share",
				"features": []string{"New products", "Market expansion", "Team growth"},
				"revenue":  "$100M target",
			},
		}

		resp, err := orchestrator.GenerateTemplateSpec(context.Background(), req)
		require.NoError(t, err)

		// Check that content was populated
		assert.Greater(t, len(resp.Spec.Layouts), 1)

		// Verify title slide content
		titleSlide := resp.Spec.Layouts[0]
		assert.Equal(t, "Title Slide", titleSlide.Name)
		assert.Equal(t, "Sales Strategy 2025", titleSlide.Placeholders[0].Content)
		assert.Equal(t, "Growing Our Market Share", titleSlide.Placeholders[1].Content)

		// Should have features slide
		hasFeatures := false
		for _, layout := range resp.Spec.Layouts {
			if layout.Name == "Features" {
				hasFeatures = true
				break
			}
		}
		assert.True(t, hasFeatures)
	})

	t.Run("Custom_Mock_Response", func(t *testing.T) {
		mockOrch := ai.NewMockOrchestrator()

		// Set custom response
		customSpec := &spec.TemplateSpec{
			Tokens: map[string]interface{}{
				"test": "custom",
			},
			Layouts: []spec.Layout{
				{
					Name: "Custom Layout",
					Placeholders: []spec.Placeholder{
						{
							ID:       "custom",
							Content:  "Custom Content",
							Geometry: spec.Geometry{X: 0.5, Y: 0.5, W: 0.4, H: 0.4},
						},
					},
				},
			},
		}

		mockOrch.SetCustomResponse("test prompt", customSpec)

		resp, err := mockOrch.GenerateTemplateSpec(context.Background(), ai.GenerationRequest{
			Prompt: "test prompt",
		})

		require.NoError(t, err)
		assert.Equal(t, "custom", resp.Spec.Tokens["test"])
		assert.Equal(t, "Custom Layout", resp.Spec.Layouts[0].Name)
	})
}

func TestMockFallback(t *testing.T) {
	// Clear environment to test fallback behavior
	os.Unsetenv("USE_MOCK_AI")
	os.Unsetenv("HUGGINGFACE_API_KEY")

	// Without API key, should fall back to mock
	orchestrator := ai.NewOrchestrator()

	req := ai.GenerationRequest{
		Prompt: "Test presentation",
		ContentData: map[string]interface{}{
			"company": "TestCorp",
		},
	}

	resp, err := orchestrator.GenerateTemplateSpec(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 0.0, resp.Cost) // Mock has no cost
	assert.Equal(t, "mock", resp.Model)
}

func TestMockRepair(t *testing.T) {
	os.Setenv("USE_MOCK_AI", "true")
	defer os.Unsetenv("USE_MOCK_AI")

	mockOrch := ai.NewMockOrchestrator()

	t.Run("Repair_Invalid_Spec", func(t *testing.T) {
		// Invalid spec with no layouts
		invalidSpec := &spec.TemplateSpec{
			Layouts: []spec.Layout{},
		}

		repaired, err := mockOrch.RepairTemplateSpec(context.Background(), invalidSpec, nil)
		require.NoError(t, err)
		assert.NotNil(t, repaired)
		assert.Greater(t, len(repaired.Layouts), 0)
	})

	t.Run("Fix_Invalid_Geometry", func(t *testing.T) {
		invalidSpec := &spec.TemplateSpec{
			Layouts: []spec.Layout{
				{
					Name: "Test",
					Placeholders: []spec.Placeholder{
						{
							ID:       "", // Missing ID
							Geometry: spec.Geometry{X: -1, Y: 2, W: 0, H: -1}, // Invalid
						},
					},
				},
			},
		}

		repaired, err := mockOrch.RepairTemplateSpec(context.Background(), invalidSpec, nil)
		require.NoError(t, err)

		// Check repairs
		placeholder := repaired.Layouts[0].Placeholders[0]
		assert.NotEmpty(t, placeholder.ID)
		assert.GreaterOrEqual(t, placeholder.Geometry.X, 0.0)
		assert.LessOrEqual(t, placeholder.Geometry.X, 1.0)
		assert.Greater(t, placeholder.Geometry.W, 0.0)
		assert.Greater(t, placeholder.Geometry.H, 0.0)
	})
}

func BenchmarkMockGeneration(b *testing.B) {
	os.Setenv("USE_MOCK_AI", "true")
	defer os.Unsetenv("USE_MOCK_AI")

	orchestrator := ai.NewOrchestrator()
	req := ai.GenerationRequest{
		Prompt: "Create a presentation",
		ContentData: map[string]interface{}{
			"company": "BenchCorp",
			"data":    "test data",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = orchestrator.GenerateTemplateSpec(context.Background(), req)
	}
}

func TestMockConsistency(t *testing.T) {
	os.Setenv("USE_MOCK_AI", "true")
	defer os.Unsetenv("USE_MOCK_AI")

	orchestrator := ai.NewOrchestrator()

	// Same input should produce consistent industry detection
	req := ai.GenerationRequest{
		Prompt: "Healthcare patient monitoring system",
		ContentData: map[string]interface{}{
			"features": "HIPAA compliant medical records",
		},
	}

	// Run multiple times
	var results []string
	for i := 0; i < 5; i++ {
		resp, err := orchestrator.GenerateTemplateSpec(context.Background(), req)
		require.NoError(t, err)

		tokens := resp.Spec.Tokens
		company, _ := tokens["company"].(map[string]interface{})
		industry, _ := company["industry"].(string)
		results = append(results, industry)
	}

	// All should be Healthcare
	for _, industry := range results {
		assert.Equal(t, "Healthcare", industry)
	}
}

func TestMockJSONGeneration(t *testing.T) {
	os.Setenv("USE_MOCK_AI", "true")
	defer os.Unsetenv("USE_MOCK_AI")

	mockOrch := ai.NewMockOrchestrator()

	jsonStr, err := mockOrch.GenerateJSON(context.Background(), "test prompt")
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &result)
	require.NoError(t, err)

	assert.True(t, result["mock"].(bool))
	assert.Equal(t, "test prompt", result["prompt"])
}