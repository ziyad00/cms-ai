// +build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ziyad/cms-ai/server/internal/assets"
)

func main() {
	fmt.Println("🧠 Testing Olama AI Bridge Integration")
	fmt.Println("=" + string(make([]rune, 40)) + "=")

	// Initialize olama AI bridge
	olamaAI := assets.NewOlamaAIBridge()

	// Check if olama AI is available
	fmt.Printf("🔍 Checking olama AI availability...\n")
	if olamaAI.IsAvailable() {
		fmt.Printf("   ✅ Olama AI is available!\n")
		fmt.Printf("   📁 Olama path: %s\n", "/Users/ziyad/Documents/olama")

		// Check environment variables
		apiKey := os.Getenv("DO_AI_API_KEY")
		if apiKey != "" {
			fmt.Printf("   🔑 API key found: %s...%s\n", apiKey[:8], apiKey[len(apiKey)-8:])
		} else {
			fmt.Printf("   ⚠️  API key not found (set DO_AI_API_KEY environment variable)\n")
		}
	} else {
		fmt.Printf("   ❌ Olama AI not available\n")
		fmt.Printf("   Required environment variables: %v\n", olamaAI.GetRequiredEnvVars())
		fmt.Printf("   Will use fallback rule-based analysis\n")
	}

	fmt.Println()

	// Test data
	testData := map[string]any{
		"slides": []any{
			map[string]any{
				"title": "AI-Powered Healthcare Innovation",
				"content": []any{
					"Machine learning algorithms for patient diagnosis",
					"Real-time monitoring with IoT integration",
					"40% improvement in patient outcomes",
					"HIPAA-compliant data management systems",
				},
			},
			map[string]any{
				"title": "Technology Platform Architecture",
				"content": []any{
					"Microservices architecture with API gateway",
					"Cloud-native deployment on AWS",
					"Kubernetes orchestration for scalability",
					"Advanced security protocols",
				},
			},
		},
	}

	companyInfo := assets.CompanyContext{
		Name:     "HealthTech AI Solutions",
		Industry: "healthcare",
		Colors: map[string]string{
			"primary":   "#2D7DB3",
			"secondary": "#38B2AC",
		},
		Values:      []string{"innovation", "trust", "patient-centered"},
		Personality: "innovative yet trustworthy",
	}

	// Test AI analysis
	fmt.Printf("🔬 Testing AI design analysis...\n")
	designIdentity, err := olamaAI.AnalyzeContentForDesign(testData, companyInfo)
	if err != nil {
		log.Printf("   ❌ Error: %v", err)
		return
	}

	// Display results
	fmt.Printf("   ✅ Analysis complete!\n\n")
	fmt.Printf("📊 AI DESIGN ANALYSIS RESULTS:\n")
	fmt.Printf("   Industry: %s\n", designIdentity.Industry)
	fmt.Printf("   Formality: %s\n", designIdentity.Formality)
	fmt.Printf("   Style: %s\n", designIdentity.Style)
	fmt.Printf("   Colors: %s\n", designIdentity.ColorPreference)
	fmt.Printf("   Audience: %s\n", designIdentity.Audience)
	fmt.Printf("   Visual Metaphor: %s\n", designIdentity.VisualMetaphor)
	fmt.Printf("   Emotional Tone: %s\n", designIdentity.EmotionalTone)
	fmt.Printf("   Reasoning: %s\n", designIdentity.Reasoning)

	fmt.Println("\n" + "=" + string(make([]rune, 40)) + "=")
	if olamaAI.IsAvailable() && os.Getenv("DO_AI_API_KEY") != "" {
		fmt.Println("🎉 OLAMA AI INTEGRATION SUCCESS!")
		fmt.Println("🧠 Using real AI for design analysis")
	} else {
		fmt.Println("⚠️  Using fallback rule-based analysis")
		fmt.Println("💡 Set DO_AI_API_KEY to enable real AI")
	}
}