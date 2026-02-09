// +build ignore

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
	fmt.Println("🐍 Testing Python PPTX Renderer with Rich Visuals")
	fmt.Println("=" + string(make([]rune, 51)) + "=")

	// Create output directory
	outputDir := "./test_outputs"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Error creating output directory: %v", err)
	}

	// Test data for healthcare presentation
	healthcareData := map[string]any{
		"tokens": map[string]any{
			"colors": map[string]any{
				"primary":    "#2D7DB3",
				"secondary":  "#38B2AC",
				"background": "#E8F8F5",
				"text":       "#2D3748",
			},
			"company": map[string]any{
				"name":     "HealthTech Medical Systems",
				"industry": "healthcare",
			},
		},
		"layouts": []map[string]any{
			{
				"name": "title-slide",
				"placeholders": []map[string]any{
					{
						"id":   "title",
						"type": "title",
						"geometry": map[string]float64{
							"x": 1.0, "y": 2.0, "w": 8.0, "h": 2.0,
						},
						"content": "Healthcare Innovation Presentation",
					},
					{
						"id":   "subtitle",
						"type": "subtitle",
						"geometry": map[string]float64{
							"x": 1.0, "y": 4.5, "w": 8.0, "h": 1.5,
						},
						"content": "Advanced Medical Technology Solutions with Rich Visual Design",
					},
				},
			},
			{
				"name": "content-slide",
				"placeholders": []map[string]any{
					{
						"id":   "title",
						"type": "title",
						"geometry": map[string]float64{
							"x": 1.0, "y": 0.5, "w": 8.0, "h": 1.0,
						},
						"content": "Key Benefits & Features",
					},
					{
						"id":   "content",
						"type": "body",
						"geometry": map[string]float64{
							"x": 1.0, "y": 2.0, "w": 8.0, "h": 4.5,
						},
						"content": "40% improvement in patient outcomes\nCost-effective implementation strategy\nScalable across multiple healthcare systems\nAI-powered diagnostic algorithms\nReal-time patient monitoring with IoT integration",
					},
				},
			},
		},
	}

	// Test data for technology presentation
	techData := map[string]any{
		"tokens": map[string]any{
			"colors": map[string]any{
				"primary":    "#667EEA",
				"secondary":  "#764BA2",
				"background": "#E8EDFF",
				"text":       "#1A202C",
			},
			"company": map[string]any{
				"name":     "TechStartup Innovations",
				"industry": "technology",
			},
		},
		"layouts": []map[string]any{
			{
				"name": "title-slide",
				"placeholders": []map[string]any{
					{
						"id":   "title",
						"type": "title",
						"geometry": map[string]float64{
							"x": 1.0, "y": 2.0, "w": 8.0, "h": 2.0,
						},
						"content": "Tech Innovation Platform",
					},
					{
						"id":   "subtitle",
						"type": "subtitle",
						"geometry": map[string]float64{
							"x": 1.0, "y": 4.5, "w": 8.0, "h": 1.5,
						},
						"content": "Next-Generation Software Solutions with Circuit Board Visuals",
					},
				},
			},
		},
	}

	// Create Python renderer
	pythonRenderer := assets.PythonPPTXRenderer{
		PythonPath: "/usr/bin/python3", // Use system Python that has python-pptx
		ScriptPath: "tools/renderer/render_pptx.py",
	}

	ctx := context.Background()

	// Test 1: Healthcare presentation
	fmt.Println("🏥 Generating Healthcare presentation with Python renderer...")
	healthcarePath := filepath.Join(outputDir, "python_healthcare_rich.pptx")
	if err := pythonRenderer.RenderPPTX(ctx, healthcareData, healthcarePath); err != nil {
		log.Printf("Error generating healthcare presentation: %v", err)
	} else {
		fmt.Printf("   ✅ Saved: python_healthcare_rich.pptx\n")
	}

	// Test 2: Technology presentation
	fmt.Println("💻 Generating Technology presentation with Python renderer...")
	techPath := filepath.Join(outputDir, "python_technology_rich.pptx")
	if err := pythonRenderer.RenderPPTX(ctx, techData, techPath); err != nil {
		log.Printf("Error generating technology presentation: %v", err)
	} else {
		fmt.Printf("   ✅ Saved: python_technology_rich.pptx\n")
	}

	// Test 3: Compare with Go renderer
	fmt.Println("🔄 Generating same presentations with Go renderer for comparison...")
	goRenderer := assets.NewGoPPTXRenderer()

	goHealthcarePath := filepath.Join(outputDir, "go_healthcare_comparison.pptx")
	if err := goRenderer.RenderPPTX(ctx, healthcareData, goHealthcarePath); err != nil {
		log.Printf("Error generating Go healthcare presentation: %v", err)
	} else {
		fmt.Printf("   ✅ Saved: go_healthcare_comparison.pptx\n")
	}

	goTechPath := filepath.Join(outputDir, "go_technology_comparison.pptx")
	if err := goRenderer.RenderPPTX(ctx, techData, goTechPath); err != nil {
		log.Printf("Error generating Go technology presentation: %v", err)
	} else {
		fmt.Printf("   ✅ Saved: go_technology_comparison.pptx\n")
	}

	fmt.Println("\n" + "=" + string(make([]rune, 51)) + "=")
	fmt.Println("🎉 PYTHON RENDERER TEST COMPLETE")
	fmt.Println("📁 Generated presentations in ./test_outputs/")
	fmt.Println("🔍 Compare the presentations:")
	fmt.Println("   • Python: Rich backgrounds, shapes, themed colors")
	fmt.Println("   • Go: Limited visual styling due to gooxml constraints")
	fmt.Println("")
	fmt.Println("💡 Recommendation: Use Python renderer for visual-rich presentations!")
}