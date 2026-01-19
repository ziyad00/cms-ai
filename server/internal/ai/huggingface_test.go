package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ziyad/cms-ai/server/internal/spec"
)

func TestHuggingFaceClient_NewClient(t *testing.T) {
	client := NewHuggingFaceClient("test-key", "test-model")

	assert.Equal(t, "test-key", client.apiKey)
	assert.Equal(t, "test-model", client.model)
	assert.Contains(t, client.baseURL, "test-model")
	assert.NotNil(t, client.httpClient)
}

func TestHuggingFaceClient_NewClientDefaults(t *testing.T) {
	client := NewHuggingFaceClient("", "")

	assert.Equal(t, "hf_default", client.apiKey)
	assert.Equal(t, "mistralai/Mixtral-8x7B-Instruct-v0.1", client.model)
}

func TestHuggingFaceClient_BuildSystemPrompt(t *testing.T) {
	client := NewHuggingFaceClient("test-key", "test-model")

	req := GenerationRequest{
		Prompt:   "Create a business presentation",
		Language: "English",
		Tone:     "Professional",
		RTL:      false,
	}

	prompt := client.buildSystemPrompt(req)

	assert.Contains(t, prompt, "TemplateSpec")
	assert.Contains(t, prompt, "English")
	assert.Contains(t, prompt, "Professional")
	assert.Contains(t, prompt, "Examples")
}

func TestHuggingFaceClient_ParseTemplateSpec(t *testing.T) {
	client := NewHuggingFaceClient("test-key", "test-model")

	// Valid JSON response
	validResponse := `{"tokens": {"colors": {"primary": "#FF0000"}}, "constraints": {"safeMargin": 0.05}, "layouts": []}`
	spec, err := client.parseTemplateSpec(validResponse)

	require.NoError(t, err)
	assert.NotNil(t, spec)
	assert.Equal(t, "#FF0000", spec.Tokens["colors"].(map[string]any)["primary"])
}

func TestHuggingFaceClient_ParseTemplateSpec_Invalid(t *testing.T) {
	client := NewHuggingFaceClient("test-key", "test-model")

	// No JSON in response
	invalidResponse := "This is just text without JSON"
	_, err := client.parseTemplateSpec(invalidResponse)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no valid JSON found")
}

func TestHuggingFaceClient_ValidateTemplateSpec(t *testing.T) {
	client := NewHuggingFaceClient("test-key", "test-model")

	// Valid spec
	validSpec := &spec.TemplateSpec{
		Tokens:      map[string]any{},
		Constraints: spec.Constraints{SafeMargin: 0.05},
		Layouts: []spec.Layout{
			{
				Name: "Test Layout",
				Placeholders: []spec.Placeholder{
					{
						ID:       "title",
						Type:     "text",
						Geometry: spec.Geometry{X: 0.1, Y: 0.1, W: 0.8, H: 0.2},
					},
				},
			},
		},
	}

	err := client.validateTemplateSpec(validSpec)
	assert.NoError(t, err)

	// Invalid spec - no layouts
	invalidSpec := &spec.TemplateSpec{
		Tokens:      map[string]any{},
		Constraints: spec.Constraints{SafeMargin: 0.05},
		Layouts:     []spec.Layout{},
	}

	err = client.validateTemplateSpec(invalidSpec)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one layout is required")
}

func TestHuggingFaceClient_ValidateTemplateSpec_Geometry(t *testing.T) {
	client := NewHuggingFaceClient("test-key", "test-model")

	// Invalid geometry
	invalidSpec := &spec.TemplateSpec{
		Tokens:      map[string]any{},
		Constraints: spec.Constraints{SafeMargin: 0.05},
		Layouts: []spec.Layout{
			{
				Name: "Test Layout",
				Placeholders: []spec.Placeholder{
					{
						ID:       "title",
						Type:     "text",
						Geometry: spec.Geometry{X: 1.5, Y: 0.1, W: 0.8, H: 0.2}, // X > 1.0
					},
				},
			},
		},
	}

	err := client.validateTemplateSpec(invalidSpec)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid geometry")
}

func TestHuggingFaceClient_EstimateTokenUsage(t *testing.T) {
	client := NewHuggingFaceClient("test-key", "test-model")

	prompt := "This is a test prompt"
	response := "This is a test response"

	usage := client.estimateTokenUsage(prompt, response)

	assert.Greater(t, usage, 0)
	// Should be approximately (len(prompt) + len(response)) / 4
	expected := (len(prompt) + len(response)) / 4
	assert.InDelta(t, expected, usage, 2)
}

func TestHuggingFaceClient_CalculateCost(t *testing.T) {
	client := NewHuggingFaceClient("test-key", "test-model")

	tokens := 1000
	cost := client.calculateCost(tokens)

	assert.Greater(t, cost, 0.0)
	// Should be approximately (300 * 0.50 / 1M) + (700 * 1.50 / 1M)
	expected := float64(300)*0.50/1000000 + float64(700)*1.50/1000000
	assert.InDelta(t, expected, cost, 0.000001)
}

func TestOrchestrator_NewOrchestrator(t *testing.T) {
	orch := NewOrchestrator()

	assert.NotNil(t, orch)
}

func TestOrchestrator_BuildRepairPrompt(t *testing.T) {
	orch := NewOrchestrator().(*orchestrator)

	invalidSpec := &spec.TemplateSpec{
		Tokens:      map[string]any{},
		Constraints: spec.Constraints{SafeMargin: 0.05},
		Layouts:     []spec.Layout{}, // Invalid - no layouts
	}

	errors := []spec.ValidationError{
		{Path: "layouts", Message: "at least one layout is required"},
	}

	prompt := orch.buildRepairPrompt(invalidSpec, errors)

	assert.Contains(t, prompt, "validation errors")
	assert.Contains(t, prompt, "at least one layout is required")
	assert.Contains(t, prompt, "TemplateSpec")
}

// Integration test with mocked HTTP client
func TestHuggingFaceClient_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This test would require a real API key and network access
	// For now, we'll skip it but keep the structure for future testing
	t.Skip("Requires Hugging Face API key")

	client := NewHuggingFaceClient("real-api-key", "mistralai/Mixtral-8x7B-Instruct-v0.1")

	req := GenerationRequest{
		Prompt: "Create a simple title slide template",
		RTL:    false,
	}

	resp, err := client.GenerateTemplateSpec(context.Background(), req)

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, resp.Spec)
	assert.Greater(t, resp.TokenUsage, 0)
	assert.Greater(t, resp.Cost, 0.0)
}
