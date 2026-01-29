package assets

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// OlamaAIBridge calls olama's AI design generator for real AI analysis
type OlamaAIBridge struct {
	OlamaPath   string // Path to olama directory
	PythonPath  string // Python executable path
	APIKey      string // DigitalOcean AI API key (from env)
}

func NewOlamaAIBridge() *OlamaAIBridge {
	return &OlamaAIBridge{
		OlamaPath:  "/Users/ziyad/Documents/olama", // Default olama path
		PythonPath: "/usr/bin/python3",             // System Python
		APIKey:     os.Getenv("DO_AI_API_KEY"),     // Get from environment
	}
}

// AnalyzeContentForDesign calls olama's AI design generator
func (o *OlamaAIBridge) AnalyzeContentForDesign(jsonData map[string]any, companyInfo CompanyContext) (*DesignIdentity, error) {
	if o.APIKey == "" {
		return nil, fmt.Errorf("DO_AI_API_KEY environment variable is required for AI analysis")
	}

	// Create temporary files for olama communication
	tempDir, err := os.MkdirTemp("", "olama-ai-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Write presentation data to temp file
	presentationFile := filepath.Join(tempDir, "presentation.json")
	presentationData, err := json.Marshal(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal presentation data: %v", err)
	}
	if err := os.WriteFile(presentationFile, presentationData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write presentation file: %v", err)
	}

	// Write company info to temp file
	companyFile := filepath.Join(tempDir, "company.json")
	companyData, err := json.Marshal(companyInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal company data: %v", err)
	}
	if err := os.WriteFile(companyFile, companyData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write company file: %v", err)
	}

	// Create olama AI bridge script
	bridgeScript := filepath.Join(tempDir, "ai_bridge.py")
	scriptContent := fmt.Sprintf(`
import sys
import json
import asyncio
sys.path.append('%s')

from app.services.proposals.ai_design_generator import AIDesignGenerator

async def main():
    # Load input data
    with open('%s', 'r') as f:
        presentation_data = json.load(f)

    with open('%s', 'r') as f:
        company_data = json.load(f)

    # Initialize AI generator with API key
    ai_generator = AIDesignGenerator(api_key='%s')

    try:
        # Call olama's AI analysis
        result = await ai_generator.analyze_content_for_unique_design(
            presentation_data,
            company_data
        )

        # Output result as JSON
        print(json.dumps(result, indent=2))

    except Exception as e:
        # Return error in JSON format
        print(json.dumps({
            "error": str(e),
            "fallback": True
        }))

if __name__ == '__main__':
    asyncio.run(main())
`, o.OlamaPath, presentationFile, companyFile, o.APIKey)

	if err := os.WriteFile(bridgeScript, []byte(scriptContent), 0755); err != nil {
		return nil, fmt.Errorf("failed to write bridge script: %v", err)
	}

	// Execute olama AI analysis
	cmd := exec.CommandContext(context.Background(), o.PythonPath, bridgeScript)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("DO_AI_API_KEY=%s", o.APIKey),
		"PYTHONPATH="+o.OlamaPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("olama AI analysis failed: %v\nOutput: %s", err, string(output))
	}

	// Parse AI response
	var aiResult map[string]any
	if err := json.Unmarshal(output, &aiResult); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %v\nOutput: %s", err, string(output))
	}

	// Check for error in response
	if errVal, hasError := aiResult["error"]; hasError {
		return nil, fmt.Errorf("olama AI analysis failed: %v", errVal)
	}

	// Convert AI result to DesignIdentity
	return o.convertAIResponseToDesignIdentity(aiResult), nil
}

// convertAIResponseToDesignIdentity converts olama's AI response to our DesignIdentity struct
func (o *OlamaAIBridge) convertAIResponseToDesignIdentity(aiResult map[string]any) *DesignIdentity {
	getString := func(key string, defaultVal string) string {
		if val, ok := aiResult[key].(string); ok {
			return val
		}
		return defaultVal
	}

	return &DesignIdentity{
		Industry:        getString("industry", "Corporate/Consulting"),
		Formality:       getString("formality", "Professional"),
		Style:          getString("style", "Clean, professional design"),
		ColorPreference: getString("color_preference", "Professional blues and grays"),
		Audience:       getString("audience", "Business professionals"),
		VisualMetaphor: getString("visual_metaphor", "Professional elements"),
		EmotionalTone:  getString("emotional_tone", "Trustworthy, professional"),
		Reasoning:      getString("reasoning", "AI-powered design analysis"),
	}
}

// IsAvailable checks if olama AI analysis is available
func (o *OlamaAIBridge) IsAvailable() bool {
	// Check if olama directory exists
	if _, err := os.Stat(o.OlamaPath); os.IsNotExist(err) {
		return false
	}

	// Check if AI generator file exists
	aiGenPath := filepath.Join(o.OlamaPath, "app/services/proposals/ai_design_generator.py")
	if _, err := os.Stat(aiGenPath); os.IsNotExist(err) {
		return false
	}

	// Check if API key is available
	return o.APIKey != ""
}

// GetRequiredEnvVars returns the environment variables needed for AI analysis
func (o *OlamaAIBridge) GetRequiredEnvVars() []string {
	return []string{
		"DO_AI_API_KEY", // DigitalOcean AI API key
	}
}