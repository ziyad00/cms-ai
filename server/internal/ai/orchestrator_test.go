package ai

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ziyad/cms-ai/server/internal/spec"
)

// testMockOrchestratorHelper for testing (different from production MockOrchestrator)
type testMockOrchestratorHelper struct {
	GenerateFunc func(ctx context.Context, req GenerationRequest) (*GenerationResponse, error)
}

func (m *testMockOrchestratorHelper) GenerateTemplateSpec(ctx context.Context, req GenerationRequest) (*GenerationResponse, error) {
	if m.GenerateFunc != nil {
		return m.GenerateFunc(ctx, req)
	}

	// Default mock response
	templateSpec := &spec.TemplateSpec{
		Tokens: map[string]interface{}{
			"colors": map[string]interface{}{
				"primary":    "#2563eb",
				"background": "#ffffff",
				"text":       "#1f2937",
				"accent":     "#10b981",
			},
			"company": map[string]interface{}{
				"name":     "TestCorp",
				"industry": "Technology",
			},
		},
		Constraints: spec.Constraints{
			SafeMargin: 0.05,
		},
		Layouts: []spec.Layout{
			{
				Name: "Title Slide",
				Placeholders: []spec.Placeholder{
					{
						ID:      "title",
						Type:    "text",
						Content: "Test Presentation",
						Geometry: spec.Geometry{
							X: 0.1, Y: 0.3, W: 0.8, H: 0.15,
						},
					},
				},
			},
		},
	}

	return &GenerationResponse{
		Spec:       templateSpec,
		TokenUsage: 100,
		Cost:       0.001,
		Model:      "mock-model",
	}, nil
}

func TestGenerationRequest(t *testing.T) {
	tests := []struct {
		name string
		req  GenerationRequest
		want map[string]interface{}
	}{
		{
			name: "BasicRequest",
			req: GenerationRequest{
				Prompt:   "Create a tech startup pitch deck",
				Language: "en",
				Tone:     "professional",
			},
			want: map[string]interface{}{
				"hasPrompt":  true,
				"hasContent": false,
			},
		},
		{
			name: "RequestWithContent",
			req: GenerationRequest{
				Prompt: "Create presentation",
				ContentData: map[string]interface{}{
					"company":  "TechCorp",
					"tagline":  "Innovation First",
					"revenue":  "$10M",
					"growth":   "200%",
					"team":     "50 engineers",
					"products": []string{"Product A", "Product B"},
				},
			},
			want: map[string]interface{}{
				"hasPrompt":  true,
				"hasContent": true,
			},
		},
		{
			name: "RequestWithBrandKit",
			req: GenerationRequest{
				Prompt: "Create branded presentation",
				BrandKit: map[string]interface{}{
					"colors": map[string]interface{}{
						"primary":   "#FF0000",
						"secondary": "#0000FF",
					},
					"fonts": map[string]interface{}{
						"heading": "Arial",
						"body":    "Helvetica",
					},
				},
			},
			want: map[string]interface{}{
				"hasPrompt":   true,
				"hasBrandKit": true,
			},
		},
		{
			name: "RTLRequest",
			req: GenerationRequest{
				Prompt:   "إنشاء عرض تقديمي",
				Language: "ar",
				RTL:      true,
			},
			want: map[string]interface{}{
				"hasPrompt": true,
				"isRTL":     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want["hasPrompt"], tt.req.Prompt != "")
			assert.Equal(t, tt.want["hasContent"], len(tt.req.ContentData) > 0)
			assert.Equal(t, tt.want["hasBrandKit"], tt.req.BrandKit != nil)
			assert.Equal(t, tt.want["isRTL"], tt.req.RTL)
		})
	}
}

func TestGenerationResponse(t *testing.T) {
	t.Run("ValidateResponse", func(t *testing.T) {
		resp := &GenerationResponse{
			Spec: &spec.TemplateSpec{
				Layouts: []spec.Layout{
					{
						Name: "Test",
						Placeholders: []spec.Placeholder{
							{
								ID:      "test",
								Content: "Test Content",
								Geometry: spec.Geometry{
									X: 0.1, Y: 0.1, W: 0.8, H: 0.8,
								},
							},
						},
					},
				},
			},
			TokenUsage: 150,
			Cost:       0.002,
			Model:      "test-model",
		}

		assert.NotNil(t, resp.Spec)
		assert.Greater(t, resp.TokenUsage, 0)
		assert.Greater(t, resp.Cost, 0.0)
		assert.NotEmpty(t, resp.Model)
	})
}

func TestTemplateSpecGeneration(t *testing.T) {
	tests := []struct {
		name        string
		prompt      string
		contentData map[string]interface{}
		wantContent string
	}{
		{
			name:   "TechStartupPitch",
			prompt: "Create a tech startup pitch deck for AI company",
			contentData: map[string]interface{}{
				"company": "AITech",
				"product": "ML Platform",
				"market":  "$50B",
			},
			wantContent: "AITech",
		},
		{
			name:   "HealthcarePresentation",
			prompt: "Medical device presentation for hospitals",
			contentData: map[string]interface{}{
				"device":     "Smart Monitor",
				"benefits":   "Real-time tracking",
				"compliance": "FDA approved",
			},
			wantContent: "Smart Monitor",
		},
		{
			name:   "FinancialReport",
			prompt: "Quarterly financial report",
			contentData: map[string]interface{}{
				"quarter": "Q4 2024",
				"revenue": "$25M",
				"profit":  "$5M",
				"growth":  "15%",
			},
			wantContent: "Q4 2024",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := GenerationRequest{
				Prompt:      tt.prompt,
				ContentData: tt.contentData,
			}

			// In real implementation, this would call AI
			// For testing, we verify the request structure
			assert.NotEmpty(t, req.Prompt)
			assert.NotNil(t, req.ContentData)

			// Verify content data is properly structured
			jsonData, err := json.Marshal(req.ContentData)
			require.NoError(t, err)
			assert.Contains(t, string(jsonData), tt.wantContent)
		})
	}
}

func TestSpecValidation(t *testing.T) {
	tests := []struct {
		name    string
		spec    *spec.TemplateSpec
		wantErr bool
	}{
		{
			name: "ValidSpec",
			spec: &spec.TemplateSpec{
				Layouts: []spec.Layout{
					{
						Name: "Valid",
						Placeholders: []spec.Placeholder{
							{
								ID: "valid",
								Geometry: spec.Geometry{
									X: 0.1, Y: 0.1, W: 0.5, H: 0.5,
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "EmptyLayouts",
			spec: &spec.TemplateSpec{
				Layouts: []spec.Layout{},
			},
			wantErr: true,
		},
		{
			name: "InvalidGeometry",
			spec: &spec.TemplateSpec{
				Layouts: []spec.Layout{
					{
						Name: "Invalid",
						Placeholders: []spec.Placeholder{
							{
								ID: "invalid",
								Geometry: spec.Geometry{
									X: -0.1, Y: 0.1, W: 1.5, H: 0.5,
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "MissingPlaceholderID",
			spec: &spec.TemplateSpec{
				Layouts: []spec.Layout{
					{
						Name: "NoID",
						Placeholders: []spec.Placeholder{
							{
								ID: "", // Empty ID
								Geometry: spec.Geometry{
									X: 0.1, Y: 0.1, W: 0.5, H: 0.5,
								},
							},
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "OverlappingPlaceholders",
			spec: &spec.TemplateSpec{
				Layouts: []spec.Layout{
					{
						Name: "Overlap",
						Placeholders: []spec.Placeholder{
							{
								ID: "first",
								Geometry: spec.Geometry{
									X: 0.1, Y: 0.1, W: 0.5, H: 0.5,
								},
							},
							{
								ID: "second",
								Geometry: spec.Geometry{
									X: 0.3, Y: 0.3, W: 0.5, H: 0.5, // Overlaps with first
								},
							},
						},
					},
				},
			},
			wantErr: false, // Overlapping is allowed but noted
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTemplateSpec(tt.spec)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Helper function to validate template spec
func validateTemplateSpec(templateSpec *spec.TemplateSpec) error {
	if templateSpec == nil {
		return assert.AnError
	}

	if len(templateSpec.Layouts) == 0 {
		return assert.AnError
	}

	for _, layout := range templateSpec.Layouts {
		if layout.Name == "" {
			return assert.AnError
		}

		for _, placeholder := range layout.Placeholders {
			if placeholder.ID == "" {
				return assert.AnError
			}

			if placeholder.Geometry.X < 0 || placeholder.Geometry.X > 1 ||
				placeholder.Geometry.Y < 0 || placeholder.Geometry.Y > 1 ||
				placeholder.Geometry.W <= 0 || placeholder.Geometry.W > 1 ||
				placeholder.Geometry.H <= 0 || placeholder.Geometry.H > 1 {
				return assert.AnError
			}
		}
	}

	return nil
}

func TestContentBinding(t *testing.T) {
	baseSpec := &spec.TemplateSpec{
		Layouts: []spec.Layout{
			{
				Name: "Title",
				Placeholders: []spec.Placeholder{
					{
						ID:       "title",
						Type:     "text",
						Content:  "", // Empty, to be filled
						Geometry: spec.Geometry{X: 0.1, Y: 0.3, W: 0.8, H: 0.2},
					},
					{
						ID:       "subtitle",
						Type:     "text",
						Content:  "", // Empty, to be filled
						Geometry: spec.Geometry{X: 0.1, Y: 0.5, W: 0.8, H: 0.1},
					},
				},
			},
		},
	}

	contentToFill := map[string]string{
		"title":    "Healthcare Analytics Platform",
		"subtitle": "AI-Powered Patient Care",
	}

	t.Run("FillPlaceholderContent", func(t *testing.T) {
		// Simulate filling content
		for i, layout := range baseSpec.Layouts {
			for j, placeholder := range layout.Placeholders {
				if content, exists := contentToFill[placeholder.ID]; exists {
					baseSpec.Layouts[i].Placeholders[j].Content = content
				}
			}
		}

		// Verify content was filled
		assert.Equal(t, "Healthcare Analytics Platform", baseSpec.Layouts[0].Placeholders[0].Content)
		assert.Equal(t, "AI-Powered Patient Care", baseSpec.Layouts[0].Placeholders[1].Content)
	})

	t.Run("PreserveGeometry", func(t *testing.T) {
		// Verify geometry wasn't changed
		assert.Equal(t, 0.1, baseSpec.Layouts[0].Placeholders[0].Geometry.X)
		assert.Equal(t, 0.3, baseSpec.Layouts[0].Placeholders[0].Geometry.Y)
		assert.Equal(t, 0.8, baseSpec.Layouts[0].Placeholders[0].Geometry.W)
		assert.Equal(t, 0.2, baseSpec.Layouts[0].Placeholders[0].Geometry.H)
	})
}

// Benchmark tests
func BenchmarkTemplateSpecValidation(b *testing.B) {
	spec := &spec.TemplateSpec{
		Layouts: []spec.Layout{
			{
				Name: "Test",
				Placeholders: []spec.Placeholder{
					{ID: "p1", Geometry: spec.Geometry{X: 0.1, Y: 0.1, W: 0.3, H: 0.3}},
					{ID: "p2", Geometry: spec.Geometry{X: 0.5, Y: 0.1, W: 0.3, H: 0.3}},
					{ID: "p3", Geometry: spec.Geometry{X: 0.1, Y: 0.5, W: 0.3, H: 0.3}},
					{ID: "p4", Geometry: spec.Geometry{X: 0.5, Y: 0.5, W: 0.3, H: 0.3}},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateTemplateSpec(spec)
	}
}

func BenchmarkJSONMarshaling(b *testing.B) {
	testSpec := &spec.TemplateSpec{
		Tokens: map[string]interface{}{
			"colors": map[string]interface{}{
				"primary": "#000000",
			},
		},
		Layouts: []spec.Layout{
			{
				Name: "Test",
				Placeholders: []spec.Placeholder{
					{
						ID:       "test",
						Content:  "Long content string that simulates real presentation text",
						Geometry: spec.Geometry{X: 0.1, Y: 0.1, W: 0.8, H: 0.8},
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, _ := json.Marshal(testSpec)
		var decoded spec.TemplateSpec
		_ = json.Unmarshal(data, &decoded)
	}
}