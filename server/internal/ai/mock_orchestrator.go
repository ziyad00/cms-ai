package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ziyad/cms-ai/server/internal/spec"
)

// MockOrchestrator provides deterministic AI responses for testing and development
type MockOrchestrator struct {
	// UseMockResponses enables mock mode
	UseMockResponses bool
	// CustomResponses allows injecting specific responses for testing
	CustomResponses map[string]*spec.TemplateSpec
}

// NewMockOrchestrator creates a new mock orchestrator
func NewMockOrchestrator() *MockOrchestrator {
	return &MockOrchestrator{
		UseMockResponses: true,
		CustomResponses:  make(map[string]*spec.TemplateSpec),
	}
}

// GenerateTemplateSpec returns mock template specs based on prompt content
func (m *MockOrchestrator) GenerateTemplateSpec(ctx context.Context, req GenerationRequest) (*GenerationResponse, error) {
	// Check for custom response first
	if customSpec, exists := m.CustomResponses[req.Prompt]; exists {
		return &GenerationResponse{
			Spec:       customSpec,
			TokenUsage: 100,
			Cost:       0.0, // No cost for mocks
			Model:      "mock",
			Timestamp:  time.Now(),
		}, nil
	}

	// Generate appropriate mock based on content analysis
	templateSpec := m.generateMockSpec(req)

	return &GenerationResponse{
		Spec:       templateSpec,
		TokenUsage: 100,
		Cost:       0.0,
		Model:      "mock",
		Timestamp:  time.Now(),
	}, nil
}

// generateMockSpec creates a mock spec based on request content
func (m *MockOrchestrator) generateMockSpec(req GenerationRequest) *spec.TemplateSpec {
	// Analyze prompt and content to determine appropriate mock
	promptLower := strings.ToLower(req.Prompt)

	// Determine industry and style
	industry := m.detectIndustry(promptLower, req.ContentData)
	colors := m.getColorsForIndustry(industry)

	// Extract company info if available
	companyName := "Sample Company"
	if company, ok := req.ContentData["company"].(string); ok {
		companyName = company
	}

	// Create layouts based on content
	layouts := m.generateLayouts(req, companyName, industry)

	return &spec.TemplateSpec{
		Tokens: map[string]interface{}{
			"colors": colors,
			"company": map[string]interface{}{
				"name":        companyName,
				"industry":    industry,
				"description": fmt.Sprintf("Mock %s presentation", industry),
			},
			"fonts": map[string]interface{}{
				"heading": "Arial",
				"body":    "Calibri",
			},
		},
		Constraints: spec.Constraints{
			SafeMargin: 0.05,
		},
		Layouts: layouts,
	}
}

// detectIndustry determines industry from prompt and content
func (m *MockOrchestrator) detectIndustry(prompt string, contentData map[string]interface{}) string {
	// Check prompt for industry keywords
	if strings.Contains(prompt, "health") || strings.Contains(prompt, "medical") ||
	   strings.Contains(prompt, "patient") || strings.Contains(prompt, "hospital") {
		return "Healthcare"
	}

	if strings.Contains(prompt, "financ") || strings.Contains(prompt, "invest") ||
	   strings.Contains(prompt, "bank") || strings.Contains(prompt, "money") {
		return "Finance"
	}

	if strings.Contains(prompt, "tech") || strings.Contains(prompt, "software") ||
	   strings.Contains(prompt, "api") || strings.Contains(prompt, "cloud") {
		return "Technology"
	}

	if strings.Contains(prompt, "education") || strings.Contains(prompt, "learn") ||
	   strings.Contains(prompt, "student") || strings.Contains(prompt, "training") {
		return "Education"
	}

	if strings.Contains(prompt, "security") || strings.Contains(prompt, "cyber") {
		return "Security"
	}

	// Check content data for clues
	for key, value := range contentData {
		keyLower := strings.ToLower(key)
		valueLower := strings.ToLower(fmt.Sprintf("%v", value))

		if strings.Contains(keyLower, "patient") || strings.Contains(valueLower, "medical") {
			return "Healthcare"
		}
		if strings.Contains(keyLower, "revenue") || strings.Contains(valueLower, "profit") {
			return "Finance"
		}
	}

	return "Corporate"
}

// getColorsForIndustry returns appropriate colors for the industry
func (m *MockOrchestrator) getColorsForIndustry(industry string) map[string]interface{} {
	switch industry {
	case "Healthcare":
		return map[string]interface{}{
			"primary":    "#48BB78",
			"secondary":  "#68D391",
			"background": "#FFFFFF",
			"text":       "#2D3748",
			"accent":     "#4299E1",
			"light":      "#F0FFF4",
		}
	case "Finance":
		return map[string]interface{}{
			"primary":    "#1B5E20",
			"secondary":  "#2E7D32",
			"background": "#FFFFFF",
			"text":       "#1B5E20",
			"accent":     "#FFB300",
			"light":      "#F1F8E9",
		}
	case "Technology":
		return map[string]interface{}{
			"primary":    "#667EEA",
			"secondary":  "#764BA2",
			"background": "#F7FAFC",
			"text":       "#1A202C",
			"accent":     "#4FD1C7",
			"light":      "#EDF2F7",
		}
	case "Security":
		return map[string]interface{}{
			"primary":    "#C53030",
			"secondary":  "#2D3748",
			"background": "#1A202C",
			"text":       "#F7FAFC",
			"accent":     "#E53E3E",
			"light":      "#4A5568",
		}
	case "Education":
		return map[string]interface{}{
			"primary":    "#2B6CB0",
			"secondary":  "#ED8936",
			"background": "#FFFBF0",
			"text":       "#2D3748",
			"accent":     "#38A169",
			"light":      "#FFF5F5",
		}
	default: // Corporate
		return map[string]interface{}{
			"primary":    "#2E75B6",
			"secondary":  "#5A6C7D",
			"background": "#FFFFFF",
			"text":       "#2C3E50",
			"accent":     "#3498DB",
			"light":      "#F8F9FA",
		}
	}
}

// generateLayouts creates appropriate layouts based on content
func (m *MockOrchestrator) generateLayouts(req GenerationRequest, companyName, industry string) []spec.Layout {
	layouts := []spec.Layout{}

	// Title slide
	titleContent := companyName
	subtitleContent := fmt.Sprintf("%s Presentation", industry)

	// Use content data if available
	if title, ok := req.ContentData["title"].(string); ok {
		titleContent = title
	}
	if subtitle, ok := req.ContentData["subtitle"].(string); ok {
		subtitleContent = subtitle
	} else if tagline, ok := req.ContentData["tagline"].(string); ok {
		subtitleContent = tagline
	}

	layouts = append(layouts, spec.Layout{
		Name: "Title Slide",
		Placeholders: []spec.Placeholder{
			{
				ID:      "title",
				Type:    "text",
				Content: titleContent,
				Geometry: spec.Geometry{
					X: 0.1, Y: 0.3, W: 0.8, H: 0.15,
				},
			},
			{
				ID:      "subtitle",
				Type:    "text",
				Content: subtitleContent,
				Geometry: spec.Geometry{
					X: 0.1, Y: 0.5, W: 0.8, H: 0.1,
				},
			},
		},
	})

	// Add content slides based on available data
	if features, ok := req.ContentData["features"]; ok {
		layouts = append(layouts, m.createFeatureSlide(features))
	}

	if benefits, ok := req.ContentData["benefits"]; ok {
		layouts = append(layouts, m.createBenefitsSlide(benefits))
	}

	// Add metrics slide if financial data present
	if _, ok := req.ContentData["revenue"]; ok {
		layouts = append(layouts, m.createMetricsSlide(req.ContentData))
	}

	// Add generic content slide if we have other data
	if len(layouts) == 1 && len(req.ContentData) > 2 {
		layouts = append(layouts, m.createContentSlide(req.ContentData))
	}

	// Always add at least one content slide
	if len(layouts) == 1 {
		layouts = append(layouts, spec.Layout{
			Name: "Content Slide",
			Placeholders: []spec.Placeholder{
				{
					ID:      "slide_title",
					Type:    "text",
					Content: "Overview",
					Geometry: spec.Geometry{
						X: 0.1, Y: 0.1, W: 0.8, H: 0.1,
					},
				},
				{
					ID:      "content",
					Type:    "text",
					Content: m.generateBulletPoints(industry),
					Geometry: spec.Geometry{
						X: 0.1, Y: 0.25, W: 0.8, H: 0.5,
					},
				},
			},
		})
	}

	return layouts
}

// createFeatureSlide creates a features slide
func (m *MockOrchestrator) createFeatureSlide(features interface{}) spec.Layout {
	content := ""
	switch v := features.(type) {
	case string:
		content = v
	case []string:
		for _, feature := range v {
			content += "• " + feature + "\n"
		}
	case []interface{}:
		for _, feature := range v {
			content += "• " + fmt.Sprintf("%v", feature) + "\n"
		}
	default:
		content = fmt.Sprintf("%v", features)
	}

	return spec.Layout{
		Name: "Features",
		Placeholders: []spec.Placeholder{
			{
				ID:       "slide_title",
				Type:     "text",
				Content:  "Key Features",
				Geometry: spec.Geometry{X: 0.1, Y: 0.1, W: 0.8, H: 0.1},
			},
			{
				ID:       "content",
				Type:     "text",
				Content:  content,
				Geometry: spec.Geometry{X: 0.1, Y: 0.25, W: 0.8, H: 0.5},
			},
		},
	}
}

// createBenefitsSlide creates a benefits slide
func (m *MockOrchestrator) createBenefitsSlide(benefits interface{}) spec.Layout {
	content := ""
	switch v := benefits.(type) {
	case string:
		content = "• " + v
	case []string:
		for _, benefit := range v {
			content += "• " + benefit + "\n"
		}
	default:
		content = fmt.Sprintf("• %v", benefits)
	}

	return spec.Layout{
		Name: "Benefits",
		Placeholders: []spec.Placeholder{
			{
				ID:       "slide_title",
				Type:     "text",
				Content:  "Benefits",
				Geometry: spec.Geometry{X: 0.1, Y: 0.1, W: 0.8, H: 0.1},
			},
			{
				ID:       "content",
				Type:     "text",
				Content:  content,
				Geometry: spec.Geometry{X: 0.1, Y: 0.25, W: 0.8, H: 0.5},
			},
		},
	}
}

// createMetricsSlide creates a metrics/financial slide
func (m *MockOrchestrator) createMetricsSlide(data map[string]interface{}) spec.Layout {
	placeholders := []spec.Placeholder{
		{
			ID:       "slide_title",
			Type:     "text",
			Content:  "Key Metrics",
			Geometry: spec.Geometry{X: 0.1, Y: 0.1, W: 0.8, H: 0.1},
		},
	}

	// Add revenue if present
	if revenueVal, ok := data["revenue"]; ok {
		placeholders = append(placeholders, spec.Placeholder{
			ID:       "revenue",
			Type:     "text",
			Content:  fmt.Sprintf("Revenue: %v", revenueVal),
			Geometry: spec.Geometry{X: 0.1, Y: 0.3, W: 0.35, H: 0.15},
		})
	}

	// Add growth if present
	if growth, ok := data["growth"]; ok {
		placeholders = append(placeholders, spec.Placeholder{
			ID:       "growth",
			Type:     "text",
			Content:  fmt.Sprintf("Growth: %v", growth),
			Geometry: spec.Geometry{X: 0.55, Y: 0.3, W: 0.35, H: 0.15},
		})
	}

	// Add other metrics
	y := 0.5
	for key, value := range data {
		if key != "revenue" && key != "growth" && key != "company" && key != "title" {
			placeholders = append(placeholders, spec.Placeholder{
				ID:       key,
				Type:     "text",
				Content:  fmt.Sprintf("%s: %v", key, value),
				Geometry: spec.Geometry{X: 0.1, Y: y, W: 0.8, H: 0.1},
			})
			y += 0.12
			if y > 0.8 {
				break
			}
		}
	}

	return spec.Layout{
		Name:         "Metrics",
		Placeholders: placeholders,
	}
}

// createContentSlide creates a generic content slide from data
func (m *MockOrchestrator) createContentSlide(data map[string]interface{}) spec.Layout {
	content := ""
	for key, value := range data {
		if key != "company" && key != "title" && key != "subtitle" {
			content += fmt.Sprintf("• %s: %v\n", key, value)
		}
	}

	return spec.Layout{
		Name: "Details",
		Placeholders: []spec.Placeholder{
			{
				ID:       "slide_title",
				Type:     "text",
				Content:  "Details",
				Geometry: spec.Geometry{X: 0.1, Y: 0.1, W: 0.8, H: 0.1},
			},
			{
				ID:       "content",
				Type:     "text",
				Content:  content,
				Geometry: spec.Geometry{X: 0.1, Y: 0.25, W: 0.8, H: 0.5},
			},
		},
	}
}

// generateBulletPoints generates industry-specific bullet points
func (m *MockOrchestrator) generateBulletPoints(industry string) string {
	switch industry {
	case "Healthcare":
		return "• Patient-centered care solutions\n• HIPAA compliant systems\n• Real-time monitoring capabilities\n• Evidence-based outcomes"
	case "Finance":
		return "• Risk management strategies\n• Portfolio optimization\n• Regulatory compliance\n• ROI maximization"
	case "Technology":
		return "• Scalable cloud infrastructure\n• API-first architecture\n• Machine learning integration\n• DevOps best practices"
	case "Security":
		return "• Threat detection and prevention\n• Zero-trust architecture\n• Compliance automation\n• Incident response protocols"
	case "Education":
		return "• Interactive learning platforms\n• Student engagement tools\n• Assessment and analytics\n• Curriculum management"
	default:
		return "• Strategic planning\n• Operational excellence\n• Stakeholder engagement\n• Performance metrics"
	}
}

// SetCustomResponse allows setting a custom response for a specific prompt
func (m *MockOrchestrator) SetCustomResponse(prompt string, spec *spec.TemplateSpec) {
	m.CustomResponses[prompt] = spec
}

// ClearCustomResponses clears all custom responses
func (m *MockOrchestrator) ClearCustomResponses() {
	m.CustomResponses = make(map[string]*spec.TemplateSpec)
}

// GenerateJSON generates raw JSON for testing
func (m *MockOrchestrator) GenerateJSON(ctx context.Context, prompt string) (string, error) {
	// Generate a simple JSON response
	mockJSON := map[string]interface{}{
		"generated": true,
		"mock":      true,
		"prompt":    prompt,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	data, err := json.Marshal(mockJSON)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// RepairTemplateSpec attempts to repair an invalid template spec
func (m *MockOrchestrator) RepairTemplateSpec(ctx context.Context, invalidSpec *spec.TemplateSpec, errors []spec.ValidationError) (*spec.TemplateSpec, error) {
	// For mock, just return a valid spec
	if invalidSpec == nil {
		return m.generateMockSpec(GenerationRequest{
			Prompt: "Default repair template",
		}), nil
	}

	// Fix common issues
	repairedSpec := *invalidSpec

	// Ensure at least one layout
	if len(repairedSpec.Layouts) == 0 {
		repairedSpec.Layouts = []spec.Layout{
			{
				Name: "Default",
				Placeholders: []spec.Placeholder{
					{
						ID:       "title",
						Type:     "text",
						Content:  "Repaired Content",
						Geometry: spec.Geometry{X: 0.1, Y: 0.3, W: 0.8, H: 0.2},
					},
				},
			},
		}
	}

	// Fix invalid geometries
	for i, layout := range repairedSpec.Layouts {
		for j, placeholder := range layout.Placeholders {
			// Ensure ID exists
			if placeholder.ID == "" {
				repairedSpec.Layouts[i].Placeholders[j].ID = fmt.Sprintf("placeholder_%d_%d", i, j)
			}

			// Fix geometry bounds
			g := &repairedSpec.Layouts[i].Placeholders[j].Geometry
			if g.X < 0 {
				g.X = 0
			}
			if g.X > 1 {
				g.X = 0.9
			}
			if g.Y < 0 {
				g.Y = 0
			}
			if g.Y > 1 {
				g.Y = 0.9
			}
			if g.W <= 0 {
				g.W = 0.1
			}
			if g.W > 1 {
				g.W = 1 - g.X
			}
			if g.H <= 0 {
				g.H = 0.1
			}
			if g.H > 1 {
				g.H = 1 - g.Y
			}
		}
	}

	return &repairedSpec, nil
}