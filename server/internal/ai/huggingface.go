package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ziyad/cms-ai/server/internal/spec"
)

type HuggingFaceClient struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

type GenerationRequest struct {
	Prompt      string                 `json:"prompt"`
	BrandKitID  string                 `json:"brandKitId,omitempty"`
	BrandKit    map[string]any         `json:"brandKit,omitempty"`
	Language    string                 `json:"language,omitempty"`
	Tone        string                 `json:"tone,omitempty"`
	RTL         bool                   `json:"rtl"`
	Tokens      map[string]any         `json:"tokens,omitempty"`
	ContentData map[string]interface{} `json:"contentData,omitempty"`
}

type GenerationResponse struct {
	Spec       *spec.TemplateSpec `json:"spec"`
	TokenUsage int                `json:"tokenUsage"`
	Cost       float64            `json:"cost"`
	Model      string             `json:"model"`
	Timestamp  time.Time          `json:"timestamp"`
}

type huggingFaceRequest struct {
	Inputs     string         `json:"inputs"`
	Parameters map[string]any `json:"parameters"`
}

type huggingFaceResponse struct {
	GeneratedText string `json:"generated_text"`
}

func NewHuggingFaceClient(apiKey, model string) *HuggingFaceClient {
	if apiKey == "" {
		apiKey = "hf_default" // Will be overridden by env var
	}
	if model == "" {
		model = "microsoft/DialoGPT-medium"
	}

	return &HuggingFaceClient{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://router.huggingface.co/models/" + model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (c *HuggingFaceClient) GenerateTemplateSpec(ctx context.Context, req GenerationRequest) (*GenerationResponse, error) {
	// Build system prompt with user requirements
	systemPrompt := c.buildSystemPrompt(req)

	// Prepare request for HuggingFace API
	hfReq := huggingFaceRequest{
		Inputs: systemPrompt + "\n\nUser request: " + req.Prompt,
		Parameters: map[string]any{
			"max_length":       2048,
			"temperature":      0.7,
			"return_full_text": false,
		},
	}

	// Marshal request
	reqBody, err := json.Marshal(hfReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// Make request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	// Parse HuggingFace response
	var hfResp []huggingFaceResponse
	if err := json.Unmarshal(respBody, &hfResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(hfResp) == 0 {
		return nil, fmt.Errorf("empty response from API")
	}

	generatedText := hfResp[0].GeneratedText

	// Parse the generated template spec
	templateSpec, err := c.parseTemplateSpec(generatedText)
	if err != nil {
		// If parsing fails, fall back to mock with user content
		fmt.Printf("Failed to parse AI response, using mock: %v\n", err)
		templateSpec = c.generateMockTemplateSpec(req)
	} else {
		// Validate the parsed spec
		if err := c.validateTemplateSpec(templateSpec); err != nil {
			fmt.Printf("Invalid template spec from AI, using mock: %v\n", err)
			templateSpec = c.generateMockTemplateSpec(req)
		}
	}

	// Calculate metrics
	tokenUsage := c.estimateTokenUsage(systemPrompt+req.Prompt, generatedText)
	cost := c.calculateCost(tokenUsage)

	return &GenerationResponse{
		Spec:       templateSpec,
		TokenUsage: tokenUsage,
		Cost:       cost,
		Model:      c.model,
		Timestamp:  time.Now(),
	}, nil
}

func (c *HuggingFaceClient) generateMockTemplateSpec(req GenerationRequest) *spec.TemplateSpec {
	// Create a simple template with user content
	layouts := []spec.Layout{
		{
			Name: "Title Slide",
			Placeholders: []spec.Placeholder{
				{
					ID:      "title",
					Type:    "text",
					Content: c.getContentValue(req.ContentData, "title", "Sales Presentation"),
					Geometry: spec.Geometry{X: 0.1, Y: 0.3, W: 0.8, H: 0.15},
				},
				{
					ID:      "subtitle",
					Type:    "text",
					Content: c.getContentValue(req.ContentData, "period", "Q1 2024"),
					Geometry: spec.Geometry{X: 0.1, Y: 0.5, W: 0.8, H: 0.1},
				},
			},
		},
		{
			Name: "Content Slide",
			Placeholders: []spec.Placeholder{
				{
					ID:      "revenue",
					Type:    "text",
					Content: "Revenue: " + c.getContentValue(req.ContentData, "revenue", "$2.5M"),
					Geometry: spec.Geometry{X: 0.1, Y: 0.2, W: 0.8, H: 0.1},
				},
				{
					ID:      "growth",
					Type:    "text",
					Content: "Growth: " + c.getContentValue(req.ContentData, "growth", "15%"),
					Geometry: spec.Geometry{X: 0.1, Y: 0.4, W: 0.8, H: 0.1},
				},
				{
					ID:      "deals",
					Type:    "text",
					Content: "Deals: " + c.getContentValue(req.ContentData, "deals", "40"),
					Geometry: spec.Geometry{X: 0.1, Y: 0.6, W: 0.8, H: 0.1},
				},
			},
		},
	}

	return &spec.TemplateSpec{
		Tokens: map[string]any{
			"colors": map[string]string{
				"primary":    "#2563eb",
				"background": "#ffffff",
				"text":       "#1f2937",
				"accent":     "#10b981",
			},
		},
		Constraints: spec.Constraints{
			SafeMargin: 0.05,
		},
		Layouts: layouts,
	}
}

func (c *HuggingFaceClient) getContentValue(contentData map[string]interface{}, key, defaultValue string) string {
	if contentData == nil {
		return defaultValue
	}
	if value, ok := contentData[key]; ok {
		if strValue, ok := value.(string); ok {
			return strValue
		}
	}
	return defaultValue
}

func (c *HuggingFaceClient) buildSystemPrompt(req GenerationRequest) string {
	examples := c.getFewShotExamples()

	prompt := `You are a professional presentation template designer. Generate a JSON TemplateSpec object based on the user's description.

The TemplateSpec must follow this exact structure:
{
  "tokens": {
    "colors": {
      "primary": "#hexcolor",
      "background": "#hexcolor",
      "text": "#hexcolor",
      "accent": "#hexcolor"
    },
    "fonts": {
      "heading": "Font Name",
      "body": "Font Name"
    },
    "logos": [],
    "images": []
  },
  "constraints": {
    "safeMargin": 0.05
  },
  "layouts": [
    {
      "name": "Layout Name",
      "placeholders": [
        {
          "id": "unique_id",
          "type": "text|image|logo",
          "content": "actual content from user data",
          "geometry": {
            "x": 0.0-1.0,
            "y": 0.0-1.0,
            "w": 0.0-1.0,
            "h": 0.0-1.0
          }
        }
      ]
    }
  ]
}

Rules:
- Geometry values are relative (0.0 to 1.0)
- Use descriptive placeholder IDs (title, subtitle, hero_image, etc.)
- Include multiple layout variations for different slide types
- Ensure placeholders don't overlap and respect safe margins
- Colors should be professional and accessible
- For RTL layouts, adjust positioning accordingly
- IMPORTANT: If contentData is provided, populate the "content" field of placeholders with actual user data`

	if req.Language != "" {
		prompt += fmt.Sprintf("\n- Generate content in %s language", req.Language)
	}
	if req.Tone != "" {
		prompt += fmt.Sprintf("\n- Use a %s tone", req.Tone)
	}
	if req.RTL {
		prompt += "\n- This is for RTL (right-to-left) layout - mirror horizontal positions"
	}
	if req.BrandKit != nil {
		prompt += "\n- Incorporate the provided brand kit colors and tokens"
	}
	if len(req.ContentData) > 0 {
		prompt += "\n- Use the following content data to populate placeholders:\n"
		for key, value := range req.ContentData {
			prompt += fmt.Sprintf("  %s: %v\n", key, value)
		}
	}

	prompt += "\n\n" + examples
	prompt += "\n\nGenerate ONLY the JSON TemplateSpec object, no explanations:"

	return prompt
}

func (c *HuggingFaceClient) getFewShotExamples() string {
	return `Examples:

User: "Create a modern tech startup pitch deck template"
ContentData: {"company": "TechCorp", "tagline": "Building the future"}
Response: {
  "tokens": {
    "colors": {
      "primary": "#2563eb",
      "background": "#ffffff",
      "text": "#1f2937",
      "accent": "#10b981"
    }
  },
  "constraints": {"safeMargin": 0.05},
  "layouts": [
    {
      "name": "Title Slide",
      "placeholders": [
        {"id": "title", "type": "text", "content": "TechCorp", "geometry": {"x": 0.1, "y": 0.3, "w": 0.8, "h": 0.15}},
        {"id": "subtitle", "type": "text", "content": "Building the future", "geometry": {"x": 0.1, "y": 0.5, "w": 0.8, "h": 0.1}},
        {"id": "logo", "type": "logo", "content": "", "geometry": {"x": 0.1, "y": 0.1, "w": 0.15, "h": 0.1}}
      ]
    }
  ]
}

User: "Sales report template with quarterly data"
ContentData: {"period": "Q4 2024", "revenue": "$2.5M", "growth": "15%"}`
}

func (c *HuggingFaceClient) parseTemplateSpec(generatedText string) (*spec.TemplateSpec, error) {
	// Look for JSON in the response
	jsonStart := bytes.Index([]byte(generatedText), []byte("{"))
	jsonEnd := bytes.LastIndex([]byte(generatedText), []byte("}"))

	if jsonStart == -1 || jsonEnd == -1 || jsonStart >= jsonEnd {
		return nil, fmt.Errorf("no valid JSON found in response")
	}

	jsonStr := generatedText[jsonStart : jsonEnd+1]

	var templateSpec spec.TemplateSpec
	if err := json.Unmarshal([]byte(jsonStr), &templateSpec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &templateSpec, nil
}

func (c *HuggingFaceClient) validateTemplateSpec(templateSpec *spec.TemplateSpec) error {
	if templateSpec == nil {
		return fmt.Errorf("template spec is nil")
	}

	if len(templateSpec.Layouts) == 0 {
		return fmt.Errorf("at least one layout is required")
	}

	for i, layout := range templateSpec.Layouts {
		if layout.Name == "" {
			return fmt.Errorf("layout %d: name is required", i)
		}

		for j, placeholder := range layout.Placeholders {
			if placeholder.ID == "" {
				return fmt.Errorf("layout %d, placeholder %d: id is required", i, j)
			}

			if placeholder.Geometry.X < 0 || placeholder.Geometry.X > 1 ||
				placeholder.Geometry.Y < 0 || placeholder.Geometry.Y > 1 ||
				placeholder.Geometry.W <= 0 || placeholder.Geometry.W > 1 ||
				placeholder.Geometry.H <= 0 || placeholder.Geometry.H > 1 {
				return fmt.Errorf("layout %d, placeholder %d: invalid geometry", i, j)
			}
		}
	}

	return nil
}

func (c *HuggingFaceClient) estimateTokenUsage(prompt, response string) int {
	// Rough estimation: ~4 characters per token
	return (len(prompt) + len(response)) / 4
}

func (c *HuggingFaceClient) calculateCost(tokens int) float64 {
	// Mixtral pricing: ~$0.50 per 1M input tokens, $1.50 per 1M output tokens
	// Assuming 30% input, 70% output
	inputTokens := int(float64(tokens) * 0.3)
	outputTokens := int(float64(tokens) * 0.7)

	return float64(inputTokens)*0.50/1000000 + float64(outputTokens)*1.50/1000000
}
