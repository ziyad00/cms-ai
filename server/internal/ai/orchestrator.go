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
	// First attempt at generation
	resp, err := o.client.GenerateTemplateSpec(ctx, req)
	if err == nil {
		return resp, nil
	}

	// If generation fails, try to repair with a fallback approach
	return o.generateWithFallback(ctx, req)
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

func (o *orchestrator) generateWithFallback(ctx context.Context, req GenerationRequest) (*GenerationResponse, error) {
	// Use a simpler prompt for fallback
	fallbackPrompt := fmt.Sprintf("Generate a basic presentation template for: %s\n\nCreate a simple TemplateSpec with title and subtitle placeholders.", req.Prompt)

	fallbackReq := GenerationRequest{
		Prompt: fallbackPrompt,
		RTL:    req.RTL,
	}

	return o.client.GenerateTemplateSpec(ctx, fallbackReq)
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
	// We keep this intentionally small: reuse the existing generation pipeline by
	// asking the model to fill placeholder.content based on a template + content blob.
	// No fallback here; caller expects AI to be required.

	b, err := json.Marshal(templateSpec)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal template spec: %w", err)
	}

	bindReq := GenerationRequest{
		Prompt: fmt.Sprintf("Bind the following content into the provided TemplateSpec by filling placeholders.content. Do not change geometry or placeholder IDs. Return ONLY valid JSON TemplateSpec.\n\nCONTENT:\n%s\n\nTEMPLATE_SPEC_JSON:\n%s", content, string(b)),
		RTL:    false,
	}

	resp, err := s.orchestrator.GenerateTemplateSpec(ctx, bindReq)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to bind deck spec: %w", err)
	}

	meteringEvent := store.MeteringEvent{
		ID:       newID("met"),
		OrgID:    orgID,
		UserID:   userID,
		Type:     "ai_bind",
		Quantity: resp.TokenUsage,
	}
	_, _ = s.store.Metering().Record(ctx, meteringEvent)

	return resp.Spec, resp, nil
}

func newID(prefix string) string {
	// This should match the ID generation used elsewhere in the codebase
	// For now, using a simple implementation
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}
