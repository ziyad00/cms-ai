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

// Test data for different industries
var testSlideData = map[string]any{
	"tokens": map[string]any{
		"colors": map[string]any{
			"primary":    "#2E75B6",
			"secondary":  "#5A6C7D",
			"background": "#FFFFFF",
			"text":       "#2C3E50",
		},
	},
	"layouts": []map[string]any{
		{
			"name": "title-slide",
			"placeholders": []map[string]any{
				{
					"id":      "title",
					"type":    "title",
					"x":       1.0,
					"y":       2.0,
					"width":   8.0,
					"height":  2.0,
					"content": "Industry-Specific Design Showcase",
				},
				{
					"id":      "subtitle",
					"type":    "subtitle",
					"x":       1.0,
					"y":       4.5,
					"width":   8.0,
					"height":  1.5,
					"content": "Demonstrating intelligent theme selection and AI-powered design analysis",
				},
			},
		},
		{
			"name": "content-slide",
			"placeholders": []map[string]any{
				{
					"id":      "title",
					"type":    "title",
					"x":       1.0,
					"y":       0.5,
					"width":   8.0,
					"height":  1.0,
					"content": "Smart Content Analysis Results",
				},
				{
					"id":      "content",
					"type":    "body",
					"x":       1.0,
					"y":       2.0,
					"width":   8.0,
					"height":  4.5,
					"content": "AI-powered design analysis with 8 industry themes\nAdvanced background rendering with pattern support\nIntelligent content analysis for sentiment and complexity\nContent-aware typography adjustments",
				},
			},
		},
		{
			"name": "data-slide",
			"placeholders": []map[string]any{
				{
					"id":      "title",
					"type":    "title",
					"x":       1.0,
					"y":       0.5,
					"width":   8.0,
					"height":  1.0,
					"content": "Market Analysis & ROI Projections",
				},
				{
					"id":      "content",
					"type":    "body",
					"x":       1.0,
					"y":       2.0,
					"width":   8.0,
					"height":  4.5,
					"content": "Market data shows 35% growth potential in Q4 2024\nCompetitive advantage: 78% efficiency improvement\nROI projection: 240% return over 18-month period\nCustomer satisfaction rating: 92% approval",
				},
			},
		},
	},
}

// Test companies for different industries
var testCompanies = []map[string]any{
	{
		"name":             "HealthTech Medical Systems",
		"industry":         "healthcare",
		"style_preference": "trustworthy",
		"colors": map[string]string{
			"primary":   "#48BB78",
			"secondary": "#68D391",
		},
	},
	{
		"name":             "FinTech Solutions Corp",
		"industry":         "finance",
		"style_preference": "conservative",
		"colors": map[string]string{
			"primary":   "#1A365D",
			"secondary": "#2C5282",
		},
	},
	{
		"name":             "TechStartup Innovations",
		"industry":         "technology",
		"style_preference": "dynamic",
		"colors": map[string]string{
			"primary":   "#667EEA",
			"secondary": "#764BA2",
		},
	},
	{
		"name":             "SecureShield Cybersecurity",
		"industry":         "security",
		"style_preference": "serious",
		"colors": map[string]string{
			"primary":   "#C53030",
			"secondary": "#E53E3E",
		},
	},
	{
		"name":             "EduTech Learning Platform",
		"industry":         "education",
		"style_preference": "friendly",
		"colors": map[string]string{
			"primary":   "#2B6CB0",
			"secondary": "#3182CE",
		},
	},
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("🎨 CMS-AI SMART FEATURES TEST SUITE")
	fmt.Println("=" + string(make([]rune, 59)) + "=")
	fmt.Println()

	// Run all tests
	testSmartContentAnalysis()
	fmt.Println()
	testAIDesignAnalysis()
	fmt.Println()
	testTypographySystem()
	fmt.Println()
	testIndustryThemes()
	fmt.Println()
	testMultiSlideGeneration()

	fmt.Println("\n" + "=" + string(make([]rune, 59)) + "=")
	fmt.Println("🎉 SMART FEATURES TEST SUITE COMPLETE")
	fmt.Println("📁 Generated test presentations in ./test_outputs/")
	fmt.Println("🔍 Check presentations for:")
	fmt.Println("  • Industry-specific color schemes and typography")
	fmt.Println("  • Content-aware layout adjustments")
	fmt.Println("  • Smart background patterns (when gooxml supports)")
	fmt.Println("  • AI-powered design recommendations")
}

func testSmartContentAnalysis() {
	fmt.Println("📊 Testing Smart Content Analysis...")

	analyzer := assets.NewSmartContentAnalyzer()

	testContent := []struct {
		name    string
		content string
	}{
		{"Healthcare Content", "Medical diagnosis and patient treatment in clinical healthcare settings"},
		{"Finance Content", "Investment portfolio analysis shows 25% ROI growth this quarter"},
		{"Technology Content", "API architecture with cloud database integration and microservices"},
		{"Urgent Content", "URGENT: Critical security breach detected - immediate action required"},
		{"Quote Content", "\"Innovation distinguishes between leaders and followers\" - Steve Jobs"},
		{"Complex Content", "The distributed microservices architecture utilizes containerized deployment with Kubernetes orchestration, implementing resilient patterns for scalability and fault tolerance across multiple availability zones with comprehensive monitoring and observability"},
	}

	for _, test := range testContent {
		analysis := analyzer.AnalyzeContent(test.content)
		fmt.Printf("  🔹 %s:\n", test.name)
		fmt.Printf("      Type: %d, Sentiment: %s, Complexity: %s\n",
			analysis.ContentType, analysis.Sentiment, analysis.Complexity)
		fmt.Printf("      Words: %d, Numbers: %v, Key Concepts: %v\n",
			analysis.WordCount, analysis.HasNumbers, analysis.KeyConcepts)
	}
}

func testAIDesignAnalysis() {
	fmt.Println("🎯 Testing AI Design Analysis...")

	analyzer := assets.NewAIDesignAnalyzer()

	// Test theme detection
	for _, company := range testCompanies[:3] { // Test first 3
		companyContext := assets.CompanyContext{
			Name:     company["name"].(string),
			Industry: company["industry"].(string),
			Colors: map[string]string{
				"primary":   company["colors"].(map[string]string)["primary"],
				"secondary": company["colors"].(map[string]string)["secondary"],
			},
		}

		identity, err := analyzer.AnalyzeContentForDesign(testSlideData, companyContext)
		if err != nil {
			log.Printf("Error analyzing design for %s: %v", company["name"], err)
			continue
		}

		fmt.Printf("  🏢 %s (%s):\n", company["name"], company["industry"])
		fmt.Printf("      Industry: %s, Formality: %s\n", identity.Industry, identity.Formality)
		fmt.Printf("      Colors: %s\n", identity.ColorPreference)
		fmt.Printf("      Style: %s\n", identity.Style[:50] + "...")
	}
}

func testTypographySystem() {
	fmt.Println("✍️  Testing Advanced Typography System...")

	typography := assets.NewAdvancedTypographySystem()

	testTexts := []struct {
		name     string
		content  string
		position string
		theme    string
	}{
		{"Corporate Title", "Quarterly Business Review", "title", "Corporate Professional"},
		{"Tech Body", "API integration with microservices architecture", "body", "Modern Tech"},
		{"Healthcare Caption", "Patient outcomes improved by 40%", "caption", "Healthcare Professional"},
		{"Finance Emphasis", "ROI increased $2.5M annually", "emphasis", "Financial Services"},
		{"Security Alert", "CRITICAL: Security breach detected", "alert", "Cybersecurity"},
	}

	for _, test := range testTexts {
		optimalStyle := typography.GetOptimalStyle(test.content, test.position, test.theme)
		report := typography.GenerateTypographyReport(test.content, test.theme)

		fmt.Printf("  📝 %s:\n", test.name)
		fmt.Printf("      Style: %s, Theme: %s\n", getStyleName(optimalStyle), test.theme)
		fmt.Printf("      Recommendations: %v\n", report["recommendations"])
	}
}

func testIndustryThemes() {
	fmt.Println("🏭 Testing Industry-Specific Themes...")

	// Create output directory
	outputDir := "./test_outputs"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Printf("Error creating output directory: %v", err)
		return
	}

	renderer := assets.NewGoPPTXRenderer()

	for idx, company := range testCompanies {
		industryName := company["industry"].(string)
		companyName := company["name"].(string)

		fmt.Printf("  🏢 Generating %s presentation (%s theme)...\n", companyName, industryName)

		// Create industry-specific test data with company info
		industryTestData := map[string]any{
			"tokens": map[string]any{
				"colors": company["colors"],
				"company": map[string]any{
					"name":     companyName,
					"industry": industryName,
				},
			},
			"layouts": testSlideData["layouts"],
		}

		// Generate PPTX with smart features
		ctx := context.Background()
		pptxBytes, err := renderer.RenderPPTXBytes(ctx, industryTestData)
		if err != nil {
			log.Printf("Error generating PPTX for %s: %v", companyName, err)
			continue
		}

		// Save the presentation
		safeName := fmt.Sprintf("smart_features_%s_%d.pptx", industryName, idx+1)
		outputPath := filepath.Join(outputDir, safeName)
		if err := os.WriteFile(outputPath, pptxBytes, 0644); err != nil {
			log.Printf("Error saving %s: %v", outputPath, err)
			continue
		}

		fmt.Printf("      ✅ Saved: %s (%s bytes)\n", safeName, formatBytes(len(pptxBytes)))
	}
}

func testMultiSlideGeneration() {
	fmt.Println("📑 Testing Multi-Slide Smart Generation...")

	// Create larger dataset for multi-slide testing
	multiSlideData := map[string]any{
		"tokens": testSlideData["tokens"],
		"layouts": []map[string]any{
			{
				"name": "title-slide",
				"placeholders": []map[string]any{
					{
						"id": "title", "type": "title", "x": 1.0, "y": 2.0, "width": 8.0, "height": 2.0,
						"content": "Healthcare Innovation Presentation",
					},
					{
						"id": "subtitle", "type": "subtitle", "x": 1.0, "y": 4.5, "width": 8.0, "height": 1.5,
						"content": "Advanced Medical Technology Solutions",
					},
				},
			},
			{
				"name": "executive-summary",
				"placeholders": []map[string]any{
					{
						"id": "title", "type": "title", "x": 1.0, "y": 0.5, "width": 8.0, "height": 1.0,
						"content": "Executive Summary",
					},
					{
						"id": "content", "type": "body", "x": 1.0, "y": 2.0, "width": 8.0, "height": 4.5,
						"content": "Revolutionary healthcare technology platform\n40% improvement in patient outcomes\nCost-effective implementation strategy\nScalable across multiple healthcare systems",
					},
				},
			},
			{
				"name": "market-analysis",
				"placeholders": []map[string]any{
					{
						"id": "title", "type": "title", "x": 1.0, "y": 0.5, "width": 8.0, "height": 1.0,
						"content": "Market Analysis & Financial Projections",
					},
					{
						"id": "content", "type": "body", "x": 1.0, "y": 2.0, "width": 8.0, "height": 4.5,
						"content": "Global healthcare market: $8.3 trillion opportunity\nDigital health segment growing 25% annually\nTarget market size: $2.1 billion addressable market\nProjected ROI: 240% return over 18 months",
					},
				},
			},
			{
				"name": "technology-platform",
				"placeholders": []map[string]any{
					{
						"id": "title", "type": "title", "x": 1.0, "y": 0.5, "width": 8.0, "height": 1.0,
						"content": "Technology Platform Architecture",
					},
					{
						"id": "content", "type": "body", "x": 1.0, "y": 2.0, "width": 8.0, "height": 4.5,
						"content": "AI-powered diagnostic algorithms and machine learning\nReal-time patient monitoring with IoT integration\nSecure HIPAA-compliant data management systems\nSeamless integration with existing hospital infrastructure",
					},
				},
			},
			{
				"name": "implementation-timeline",
				"placeholders": []map[string]any{
					{
						"id": "title", "type": "title", "x": 1.0, "y": 0.5, "width": 8.0, "height": 1.0,
						"content": "Implementation Timeline & Roadmap",
					},
					{
						"id": "content", "type": "body", "x": 1.0, "y": 2.0, "width": 8.0, "height": 4.5,
						"content": "Q1 2024: System design and technical architecture\nQ2 2024: Development phase and rigorous testing\nQ3 2024: Pilot program with select healthcare partners\nQ4 2024: Full deployment and nationwide rollout",
					},
				},
			},
		},
	}

	renderer := assets.NewGoPPTXRenderer()

	fmt.Println("  📑 Generating 5-slide healthcare presentation with smart features...")

	ctx := context.Background()
	pptxBytes, err := renderer.RenderPPTXBytes(ctx, multiSlideData)
	if err != nil {
		log.Printf("Error generating multi-slide PPTX: %v", err)
		return
	}

	outputDir := "./test_outputs"
	outputPath := filepath.Join(outputDir, "multi_slide_healthcare_showcase.pptx")
	if err := os.WriteFile(outputPath, pptxBytes, 0644); err != nil {
		log.Printf("Error saving multi-slide presentation: %v", err)
		return
	}

	fmt.Printf("      ✅ Generated: multi_slide_healthcare_showcase.pptx (%s bytes)\n", formatBytes(len(pptxBytes)))
	fmt.Println("      🔍 Verify all 5 slides have:")
	fmt.Println("         • Healthcare-themed colors and typography")
	fmt.Println("         • Content-aware layout adjustments")
	fmt.Println("         • Smart background patterns (medical theme)")
	fmt.Println("         • Consistent design identity throughout")
}

// Helper functions
func getStyleName(style assets.TextStyle) string {
	names := map[assets.TextStyle]string{
		0: "Title Slide",
		1: "Slide Title",
		2: "Body Text",
		3: "Caption",
		4: "Emphasis",
		5: "Quote",
		6: "Code",
		7: "List Item",
	}
	return names[style]
}

func formatBytes(bytes int) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
}