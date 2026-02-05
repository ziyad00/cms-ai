package assets

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/ziyad/cms-ai/server/internal/store"
)

// AIEnhancedRenderer wraps the Python renderer with company context extraction
type AIEnhancedRenderer struct {
	pythonRenderer *PythonPPTXRenderer
	store          store.Store
}

// NewAIEnhancedRenderer creates a new AI-enhanced renderer
func NewAIEnhancedRenderer(st store.Store) *AIEnhancedRenderer {
	return &AIEnhancedRenderer{
		pythonRenderer: &PythonPPTXRenderer{
			PythonPath:        "python3",
			ScriptPath:        "/app/tools/renderer/simple_test.py",
			HuggingFaceAPIKey: os.Getenv("HUGGING_FACE_API_KEY"),
		},
		store: st,
	}
}

// RenderPPTX renders with AI enhancement when possible
func (r *AIEnhancedRenderer) RenderPPTX(ctx context.Context, spec any, outPath string) error {
	// Try to extract company context from spec
	company := r.extractCompanyContext(spec)

	if company != nil {
		log.Printf("AI-enhanced rendering with company context: %s", company.Name)
		return r.pythonRenderer.RenderPPTXWithCompany(ctx, spec, outPath, company)
	}

	// Fallback to basic rendering
	return r.pythonRenderer.RenderPPTX(ctx, spec, outPath)
}

// RenderPPTXBytes renders to bytes
func (r *AIEnhancedRenderer) RenderPPTXBytes(ctx context.Context, spec any) ([]byte, error) {
	tmpFile, err := os.CreateTemp("", "render-*.pptx")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if err := r.RenderPPTX(ctx, spec, tmpFile.Name()); err != nil {
		return nil, err
	}

	return os.ReadFile(tmpFile.Name())
}

// GenerateSlideThumbnails generates thumbnails (delegates to Python renderer)
func (r *AIEnhancedRenderer) GenerateSlideThumbnails(ctx context.Context, spec any) ([][]byte, error) {
	return r.pythonRenderer.GenerateSlideThumbnails(ctx, spec)
}

// extractCompanyContext attempts to extract company information from the spec
func (r *AIEnhancedRenderer) extractCompanyContext(spec any) *CompanyContext {
	// Convert spec to map if needed
	var specMap map[string]interface{}

	switch v := spec.(type) {
	case []byte:
		if err := json.Unmarshal(v, &specMap); err != nil {
			return nil
		}
	case json.RawMessage:
		if err := json.Unmarshal([]byte(v), &specMap); err != nil {
			return nil
		}
	case map[string]interface{}:
		specMap = v
	default:
		// Try to marshal and unmarshal
		data, err := json.Marshal(spec)
		if err != nil {
			return nil
		}
		if err := json.Unmarshal(data, &specMap); err != nil {
			return nil
		}
	}

	// Look for tokens.company in the spec
	if tokens, ok := specMap["tokens"].(map[string]interface{}); ok {
		if companyData, ok := tokens["company"].(map[string]interface{}); ok {
			company := &CompanyContext{}

			// Extract company fields
			if name, ok := companyData["name"].(string); ok {
				company.Name = name
			}
			if industry, ok := companyData["industry"].(string); ok {
				company.Industry = industry
			}
			// Description and Mission fields might not exist in CompanyContext
			// We'll just use Name and Industry for now
			if values, ok := companyData["values"].([]interface{}); ok {
				for _, v := range values {
					if str, ok := v.(string); ok {
						company.Values = append(company.Values, str)
					}
				}
			}

			// Only return if we have meaningful data
			if company.Name != "" || company.Industry != "" {
				return company
			}
		}
	}

	// Try to extract from brand kit tokens if available
	if tokens, ok := specMap["tokens"].(map[string]interface{}); ok {
		if brandKit, ok := tokens["brandKit"].(map[string]interface{}); ok {
			company := &CompanyContext{}

			// Extract from brand kit
			if name, ok := brandKit["companyName"].(string); ok {
				company.Name = name
			}
			if industry, ok := brandKit["industry"].(string); ok {
				company.Industry = industry
			}

			if company.Name != "" || company.Industry != "" {
				return company
			}
		}
	}

	// Try to infer from content
	company := r.inferCompanyFromContent(specMap)
	if company != nil && (company.Name != "" || company.Industry != "") {
		return company
	}

	return nil
}

// inferCompanyFromContent tries to infer company context from slide content
func (r *AIEnhancedRenderer) inferCompanyFromContent(specMap map[string]interface{}) *CompanyContext {
	company := &CompanyContext{}

	// Look through layouts for clues
	if layouts, ok := specMap["layouts"].([]interface{}); ok {
		for _, layout := range layouts {
			if layoutMap, ok := layout.(map[string]interface{}); ok {
				if placeholders, ok := layoutMap["placeholders"].([]interface{}); ok {
					for _, placeholder := range placeholders {
						if phMap, ok := placeholder.(map[string]interface{}); ok {
							content := fmt.Sprintf("%v", phMap["content"])

							// Look for industry keywords
							if company.Industry == "" {
								if containsHealthcareKeywords(content) {
									company.Industry = "Healthcare"
								} else if containsFinanceKeywords(content) {
									company.Industry = "Finance"
								} else if containsTechKeywords(content) {
									company.Industry = "Technology"
								} else if containsEducationKeywords(content) {
									company.Industry = "Education"
								}
							}
						}
					}
				}
			}
		}
	}

	return company
}

func containsHealthcareKeywords(content string) bool {
	keywords := []string{"health", "medical", "patient", "hospital", "clinic", "doctor", "treatment", "diagnosis", "HIPAA"}
	for _, keyword := range keywords {
		if containsIgnoreCase(content, keyword) {
			return true
		}
	}
	return false
}

func containsFinanceKeywords(content string) bool {
	keywords := []string{"finance", "banking", "investment", "portfolio", "trading", "asset", "capital", "ROI", "revenue"}
	for _, keyword := range keywords {
		if containsIgnoreCase(content, keyword) {
			return true
		}
	}
	return false
}

func containsTechKeywords(content string) bool {
	keywords := []string{"software", "API", "cloud", "digital", "platform", "technology", "data", "analytics", "AI", "machine learning"}
	for _, keyword := range keywords {
		if containsIgnoreCase(content, keyword) {
			return true
		}
	}
	return false
}

func containsEducationKeywords(content string) bool {
	keywords := []string{"education", "learning", "training", "curriculum", "student", "teacher", "course", "academy", "university"}
	for _, keyword := range keywords {
		if containsIgnoreCase(content, keyword) {
			return true
		}
	}
	return false
}

func containsIgnoreCase(s, substr string) bool {
	// Simple case-insensitive contains (could be optimized)
	return len(s) > 0 && len(substr) > 0 &&
		(len(s) >= len(substr)) &&
		containsHelper(s, substr)
}

func containsHelper(s, substr string) bool {
	// Convert both to lowercase for comparison
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if toLower(s[i+j]) != toLower(substr[j]) {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + 32
	}
	return c
}