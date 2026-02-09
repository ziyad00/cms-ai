package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ziyad/cms-ai/server/internal/assets"
)

func main() {
	// Test spec for healthcare presentation
	testSpec := map[string]interface{}{
		"layouts": []map[string]interface{}{
			{
				"name": "title_slide",
				"placeholders": []map[string]interface{}{
					{
						"id":      "title",
						"type":    "title",
						"content": "AI-Powered Patient Care Platform",
						"geometry": map[string]interface{}{
							"x": 1.0, "y": 2.0, "w": 8.0, "h": 1.5,
						},
					},
					{
						"id":      "subtitle",
						"type":    "subtitle",
						"content": "Revolutionizing Healthcare with Machine Learning",
						"geometry": map[string]interface{}{
							"x": 1.0, "y": 4.0, "w": 8.0, "h": 1.0,
						},
					},
				},
			},
			{
				"name": "content_slide",
				"placeholders": []map[string]interface{}{
					{
						"id":      "slide_title",
						"type":    "title",
						"content": "Core Features",
						"geometry": map[string]interface{}{
							"x": 1.0, "y": 0.5, "w": 8.0, "h": 1.0,
						},
					},
					{
						"id":      "content",
						"type":    "body",
						"content": "Predictive analytics for early diagnosis\nReal-time patient monitoring dashboards\nHIPAA-compliant data security\nSeamless EHR integration",
						"geometry": map[string]interface{}{
							"x": 1.0, "y": 2.0, "w": 8.0, "h": 4.0,
						},
					},
				},
			},
		},
	}

	// Test company context for AI analysis
	testCompany := &assets.CompanyContext{
		Name:        "HealthTech Innovations",
		Industry:    "Healthcare Technology",
		Colors:      map[string]string{"primary": "#0066CC", "secondary": "#00CC66"},
		Values:      []string{"Innovation", "Patient Care", "Data Security"},
		Personality: "Professional and trustworthy",
	}

	// Test Python renderer with AI enhancement
	fmt.Println("🧪 Testing Go → Python AI integration...")

	pythonRenderer := &assets.PythonPPTXRenderer{
		PythonPath:        "python3",
		ScriptPath:        filepath.Join("renderer", "render_pptx.py"),
		HuggingFaceAPIKey: os.Getenv("HUGGING_FACE_API_KEY"),
	}

	// Test basic rendering
	outputPath := "test_go_python_basic.pptx"
	err := pythonRenderer.RenderPPTX(context.Background(), testSpec, outputPath)
	if err != nil {
		log.Printf("❌ Basic rendering failed: %v", err)
	} else {
		fmt.Printf("✅ Basic rendering successful: %s\n", outputPath)
	}

	// Test with company info for AI analysis
	outputPathAI := "test_go_python_ai.pptx"
	err = pythonRenderer.RenderPPTXWithCompany(context.Background(), testSpec, outputPathAI, testCompany)
	if err != nil {
		log.Printf("❌ AI-enhanced rendering failed: %v", err)
	} else {
		fmt.Printf("✅ AI-enhanced rendering successful: %s\n", outputPathAI)
	}

	fmt.Println("🎉 Go integration test completed!")
}