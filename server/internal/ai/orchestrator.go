package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ziyad/cms-ai/server/internal/spec"
	"github.com/ziyad/cms-ai/server/internal/store"
)

type Orchestrator interface {
	GenerateTemplateSpec(ctx context.Context, req GenerationRequest) (*GenerationResponse, error)
	GenerateJSON(ctx context.Context, prompt string) (string, error)
	RepairTemplateSpec(ctx context.Context, invalidSpec *spec.TemplateSpec, errors []spec.ValidationError) (*spec.TemplateSpec, error)
}

type orchestrator struct {
	client *HuggingFaceClient
}

func NewOrchestrator() Orchestrator {
	// Check if we should use mock mode
	if os.Getenv("USE_MOCK_AI") == "true" {
		return NewMockOrchestrator()
	}

	apiKey := os.Getenv("HUGGINGFACE_API_KEY")
	model := os.Getenv("HUGGINGFACE_MODEL")
	if model == "" {
		model = "moonshotai/Kimi-K2-Instruct-0905"
	}

	// If no API key, use mock mode to avoid costs
	if apiKey == "" {
		return NewMockOrchestrator()
	}

	return &orchestrator{
		client: NewHuggingFaceClient(apiKey, model),
	}
}

func (o *orchestrator) GenerateTemplateSpec(ctx context.Context, req GenerationRequest) (*GenerationResponse, error) {
	// 1. Primary AI Attempt
	resp, err := o.client.GenerateTemplateSpec(ctx, req)
	if err == nil {
		return resp, nil
	}

	// 2. Guaranteed Static Fallback (No AI)
	// If AI is unreachable or fails, return a basic structural template
	return o.generateStaticSafetyNet(req), nil
}

func (o *orchestrator) generateStaticSafetyNet(req GenerationRequest) *GenerationResponse {
	title := "New Presentation"
	if req.Prompt != "" {
		title = req.Prompt
	}

	spec := &spec.TemplateSpec{
		Tokens: map[string]interface{}{
			"colors": map[string]interface{}{
				"primary":    "#2563eb",
				"background": "#ffffff",
				"text":       "#1f2937",
				"accent":     "#10b981",
			},
			"fonts": map[string]interface{}{
				"heading": "Arial",
				"body":    "Helvetica",
			},
		},
		Constraints: spec.Constraints{SafeMargin: 0.05},
		Layouts: []spec.Layout{
			{
				Name: "Title Slide",
				Placeholders: []spec.Placeholder{
					{ID: "title", Type: "text", Content: title, Geometry: spec.Geometry{X: 0.1, Y: 0.3, W: 0.8, H: 0.2}},
					{ID: "subtitle", Type: "text", Content: "Generated via Safety Fallback", Geometry: spec.Geometry{X: 0.1, Y: 0.5, W: 0.8, H: 0.1}},
				},
			},
			{
				Name: "Content Slide",
				Placeholders: []spec.Placeholder{
					{ID: "title", Type: "text", Content: "Overview", Geometry: spec.Geometry{X: 0.1, Y: 0.1, W: 0.8, H: 0.1}},
					{ID: "body", Type: "text", Content: "AI generation is currently unavailable. You can edit this content manually.", Geometry: spec.Geometry{X: 0.1, Y: 0.25, W: 0.8, H: 0.5}},
				},
			},
		},
	}

	return &GenerationResponse{
		Spec:       spec,
		TokenUsage: 0,
		Cost:       0,
		Model:      "static-fallback",
		Timestamp:  time.Now(),
	}
}

func (o *orchestrator) GenerateJSON(ctx context.Context, prompt string) (string, error) {
	return o.client.GenerateRaw(ctx, prompt)
}

func (o *orchestrator) RepairTemplateSpec(ctx context.Context, invalidSpec *spec.TemplateSpec, errors []spec.ValidationError) (*spec.TemplateSpec, error) {
	// Create a repair request with error details
	repairPrompt := o.buildRepairPrompt(invalidSpec, errors)

	repairReq := GenerationRequest{
		Prompt: repairPrompt,
		RTL:    false,
	}

	resp, err := o.client.GenerateTemplateSpec(ctx, repairReq)
	if err != nil {
		return nil, fmt.Errorf("failed to repair template spec: %w", err)
	}

	return resp.Spec, nil
}

func (o *orchestrator) buildRepairPrompt(invalidSpec *spec.TemplateSpec, errors []spec.ValidationError) string {
	prompt := "The following TemplateSpec has validation errors. Please fix them:\n\n"
	prompt += "Errors:\n"
	for _, err := range errors {
		prompt += fmt.Sprintf("- %s at %s\n", err.Message, err.Path)
	}
	prompt += "\nInvalid TemplateSpec:\n"

	// Convert spec to JSON for repair
	if specJSON, err := json.Marshal(invalidSpec); err == nil {
		prompt += string(specJSON)
	}

	prompt += "\n\nProvide a corrected TemplateSpec that fixes all validation errors:"

	return prompt
}

// AIServiceInterface defines the interface for AI template generation
type AIServiceInterface interface {
	GenerateTemplateForRequest(ctx context.Context, orgID, userID string, req GenerationRequest, brandKitID string) (*spec.TemplateSpec, *GenerationResponse, error)
	BindDeckSpec(ctx context.Context, orgID, userID string, templateSpec *spec.TemplateSpec, content string) (*spec.TemplateSpec, *GenerationResponse, error)
}

// AIService handles AI generation for templates
type AIService struct {
	orchestrator Orchestrator
	store        store.Store
}

func NewAIService(store store.Store) *AIService {
	return &AIService{
		orchestrator: NewOrchestrator(),
		store:        store,
	}
}

// GenerateTemplateForRequest handles the synchronous generation of a template spec
func (s *AIService) GenerateTemplateForRequest(ctx context.Context, orgID, userID string, req GenerationRequest, brandKitID string) (*spec.TemplateSpec, *GenerationResponse, error) {
	// Load brand kit if specified
	if brandKitID != "" {
		brandKits, err := s.store.BrandKits().List(ctx, orgID)
		if err == nil {
			for _, bk := range brandKits {
				if bk.ID == brandKitID {
					if tokens, ok := bk.Tokens.(map[string]any); ok {
						req.BrandKit = tokens
					}
					break
				}
			}
		}
	}

	// Generate the template spec
	resp, err := s.orchestrator.GenerateTemplateSpec(ctx, req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate template spec: %w", err)
	}

	// Record token usage
	meteringEvent := store.MeteringEvent{
		ID:       newID("met"),
		OrgID:    orgID,
		UserID:   userID,
		Type:     "ai_generation",
		Quantity: resp.TokenUsage,
	}
	_, _ = s.store.Metering().Record(ctx, meteringEvent)

	return resp.Spec, resp, nil
}

func (s *AIService) BindDeckSpec(ctx context.Context, orgID, userID string, templateSpec *spec.TemplateSpec, content string) (*spec.TemplateSpec, *GenerationResponse, error) {
	b, err := json.Marshal(templateSpec)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal template spec: %w", err)
	}

	bindReq := GenerationRequest{
		Prompt: fmt.Sprintf("Bind the following content into the provided TemplateSpec by filling placeholders.content. Do not change geometry or placeholder IDs. Return ONLY valid JSON TemplateSpec.\n\nCONTENT:\n%s\n\nTEMPLATE_SPEC_JSON:\n%s", content, string(b)),
		RTL:    false,
	}

	resp, err := s.orchestrator.GenerateTemplateSpec(ctx, bindReq)
	if err == nil {
		return resp.Spec, resp, nil
	}

	// Fallback: If AI fails to bind, return the original template spec
	// The user can then edit the empty placeholders in the UI
	return templateSpec, &GenerationResponse{Spec: templateSpec, Model: "binding-fallback"}, nil
}

func newID(prefix string) string {
	// This should match the ID generation used elsewhere in the codebase
	// For now, using a simple implementation
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}
