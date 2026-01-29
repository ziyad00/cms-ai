package assets

import (
	"fmt"
	"regexp"
	"strings"
)

type ThemeType int

const (
	ThemeTechnology ThemeType = iota
	ThemeBusiness
	ThemeSecurity
	ThemeInnovation
	ThemeHealthcare
	ThemeFinance
	ThemeGovernment
	ThemeEducation
)

type DesignIdentity struct {
	Industry        string  `json:"industry"`
	Formality       string  `json:"formality"`
	Style          string  `json:"style"`
	ColorPreference string  `json:"color_preference"`
	Audience       string  `json:"audience"`
	VisualMetaphor string  `json:"visual_metaphor"`
	EmotionalTone  string  `json:"emotional_tone"`
	Reasoning      string  `json:"reasoning"`
}

type ContentThemeAnalysis struct {
	DominantTheme   ThemeType
	ThemeStrength   int
	Complexity      string
	KeyConcepts     []string
	ContentLength   int
	SlideCount      int
}

type CompanyContext struct {
	Name        string            `json:"name"`
	Industry    string            `json:"industry"`
	Colors      map[string]string `json:"colors"`
	Logo        string            `json:"logo"`
	Values      []string          `json:"values"`
	Personality string            `json:"personality"`
}

type AIDesignAnalyzer struct {
	analyzer *SmartContentAnalyzer
}

func NewAIDesignAnalyzer() *AIDesignAnalyzer {
	return &AIDesignAnalyzer{
		analyzer: NewSmartContentAnalyzer(),
	}
}

func (a *AIDesignAnalyzer) AnalyzeContentForDesign(jsonData map[string]any, companyInfo CompanyContext) (*DesignIdentity, error) {
	// Extract content from slides
	content := a.extractAllContent(jsonData)

	// Perform theme analysis
	themeAnalysis := a.analyzeContentThemes(content, jsonData)

	// Generate design identity
	return a.generateDesignIdentity(themeAnalysis, companyInfo), nil
}

func (a *AIDesignAnalyzer) extractAllContent(jsonData map[string]any) string {
	var allContent []string

	slides, ok := jsonData["slides"].([]any)
	if !ok {
		return ""
	}

	for _, slideData := range slides {
		slide, ok := slideData.(map[string]any)
		if !ok {
			continue
		}

		if title, ok := slide["title"].(string); ok {
			allContent = append(allContent, title)
		}

		if content, ok := slide["content"].([]any); ok {
			for _, item := range content {
				if str, ok := item.(string); ok {
					allContent = append(allContent, str)
				}
			}
		}
	}

	return strings.Join(allContent, " ")
}

func (a *AIDesignAnalyzer) analyzeContentThemes(content string, jsonData map[string]any) ContentThemeAnalysis {
	lowerContent := strings.ToLower(content)

	// Define theme keywords
	themeKeywords := map[ThemeType][]string{
		ThemeTechnology: {"api", "database", "architecture", "backend", "frontend", "cloud", "digital", "software", "platform", "system", "technology", "tech", "coding", "development"},
		ThemeBusiness:   {"strategy", "governance", "stakeholder", "management", "roi", "value", "business", "corporate", "revenue", "profit", "market", "customer", "sales"},
		ThemeSecurity:   {"security", "encryption", "compliance", "risk", "authentication", "vulnerability", "threat", "protection", "firewall", "breach"},
		ThemeInnovation: {"ai", "machine learning", "innovation", "automation", "future", "artificial intelligence", "ml", "algorithm", "data science", "neural"},
		ThemeHealthcare: {"medical", "health", "patient", "clinical", "diagnosis", "treatment", "healthcare", "medicine", "pharmaceutical", "therapy"},
		ThemeFinance:    {"finance", "financial", "investment", "banking", "budget", "cost", "funding", "capital", "asset", "portfolio", "trading"},
		ThemeGovernment: {"government", "public", "policy", "regulation", "compliance", "authority", "administration", "civic", "municipal"},
		ThemeEducation:  {"education", "learning", "student", "teaching", "curriculum", "academic", "university", "training", "knowledge"},
	}

	// Count keyword occurrences
	themeCounts := make(map[ThemeType]int)
	for theme, keywords := range themeKeywords {
		count := 0
		for _, keyword := range keywords {
			if strings.Contains(lowerContent, keyword) {
				count++
			}
		}
		themeCounts[theme] = count
	}

	// Find dominant theme
	dominantTheme := ThemeBusiness // default
	maxCount := 0
	for theme, count := range themeCounts {
		if count > maxCount {
			maxCount = count
			dominantTheme = theme
		}
	}

	// Extract key concepts
	keyConcepts := a.extractKeyConcepts(content, 10)

	// Determine complexity
	wordCount := len(strings.Fields(content))
	complexity := "low"
	if wordCount > 1000 {
		complexity = "high"
	} else if wordCount > 500 {
		complexity = "medium"
	}

	slides, _ := jsonData["slides"].([]any)

	return ContentThemeAnalysis{
		DominantTheme: dominantTheme,
		ThemeStrength: maxCount,
		Complexity:    complexity,
		KeyConcepts:   keyConcepts,
		ContentLength: wordCount,
		SlideCount:    len(slides),
	}
}

func (a *AIDesignAnalyzer) extractKeyConcepts(content string, limit int) []string {
	words := strings.Fields(strings.ToLower(content))

	// Find words longer than 6 characters (likely to be meaningful concepts)
	var concepts []string
	conceptMap := make(map[string]bool)

	for _, word := range words {
		// Clean word of punctuation
		cleaned := regexp.MustCompile(`[^a-zA-Z0-9]`).ReplaceAllString(word, "")
		if len(cleaned) > 6 && !conceptMap[cleaned] {
			concepts = append(concepts, cleaned)
			conceptMap[cleaned] = true

			if len(concepts) >= limit {
				break
			}
		}
	}

	return concepts
}

func (a *AIDesignAnalyzer) generateDesignIdentity(analysis ContentThemeAnalysis, companyInfo CompanyContext) *DesignIdentity {
	// Map theme to industry and style recommendations
	themeMapping := map[ThemeType]struct {
		industry        string
		formality       string
		style          string
		colorPreference string
		audience       string
		visualMetaphor string
		emotionalTone  string
	}{
		ThemeTechnology: {
			industry:        "Technology/Software",
			formality:       "Modern Professional",
			style:          "Clean, geometric design with subtle tech patterns",
			colorPreference: "Cool blues and grays with electric accents",
			audience:       "Technical professionals and developers",
			visualMetaphor: "Circuit patterns, network connections, data flow",
			emotionalTone:  "Innovative, precise, forward-thinking",
		},
		ThemeBusiness: {
			industry:        "Corporate/Consulting",
			formality:       "Highly Professional",
			style:          "Conservative, structured design with business hierarchy",
			colorPreference: "Navy blue, charcoal gray with gold accents",
			audience:       "C-level executives and business stakeholders",
			visualMetaphor: "Growth charts, building blocks, ascending arrows",
			emotionalTone:  "Trustworthy, authoritative, results-driven",
		},
		ThemeSecurity: {
			industry:        "Cybersecurity/Compliance",
			formality:       "Serious Professional",
			style:          "Strong, secure design with protective elements",
			colorPreference: "Dark blues and reds with security-themed accents",
			audience:       "Security professionals and risk managers",
			visualMetaphor: "Shields, locks, fortress walls, encrypted patterns",
			emotionalTone:  "Secure, vigilant, protective",
		},
		ThemeInnovation: {
			industry:        "AI/Research & Development",
			formality:       "Creative Professional",
			style:          "Futuristic design with AI-inspired elements",
			colorPreference: "Purple gradients with neon accents",
			audience:       "Researchers, innovators, and tech enthusiasts",
			visualMetaphor: "Neural networks, molecular structures, abstract patterns",
			emotionalTone:  "Cutting-edge, intelligent, transformative",
		},
		ThemeHealthcare: {
			industry:        "Healthcare/Medical",
			formality:       "Clinical Professional",
			style:          "Clean, healing-focused design with medical elements",
			colorPreference: "Medical blues and greens with calming tones",
			audience:       "Healthcare professionals and patients",
			visualMetaphor: "Cross symbols, heartbeat lines, molecular diagrams",
			emotionalTone:  "Caring, precise, life-affirming",
		},
		ThemeFinance: {
			industry:        "Financial Services/Banking",
			formality:       "Conservative Professional",
			style:          "Sophisticated design with financial growth themes",
			colorPreference: "Deep greens and golds with prosperity accents",
			audience:       "Financial advisors, investors, and clients",
			visualMetaphor: "Growth charts, currency symbols, stability pillars",
			emotionalTone:  "Prosperous, stable, growth-oriented",
		},
		ThemeGovernment: {
			industry:        "Government/Public Sector",
			formality:       "Institutional Professional",
			style:          "Authoritative design with civic elements",
			colorPreference: "Patriotic colors with institutional blues",
			audience:       "Government officials and public servants",
			visualMetaphor: "Institutional columns, civic symbols, service emblems",
			emotionalTone:  "Authoritative, trustworthy, service-oriented",
		},
		ThemeEducation: {
			industry:        "Education/Training",
			formality:       "Academic Professional",
			style:          "Learning-focused design with educational elements",
			colorPreference: "Warm oranges and blues with knowledge accents",
			audience:       "Educators, students, and administrators",
			visualMetaphor: "Books, lightbulbs, growth trees, knowledge pathways",
			emotionalTone:  "Enlightening, growth-focused, accessible",
		},
	}

	mapping := themeMapping[analysis.DominantTheme]

	// Adjust formality based on complexity
	if analysis.Complexity == "high" && analysis.SlideCount > 15 {
		mapping.formality = "Highly " + mapping.formality
	}

	// Generate reasoning
	reasoning := fmt.Sprintf(
		"Content analysis revealed %s as the dominant theme (strength: %d). With %d slides and %s complexity content, this presentation targets %s. The visual approach emphasizes %s to convey %s emotions, making it suitable for %s.",
		mapping.industry,
		analysis.ThemeStrength,
		analysis.SlideCount,
		analysis.Complexity,
		strings.ToLower(mapping.audience),
		strings.ToLower(mapping.visualMetaphor),
		strings.ToLower(mapping.emotionalTone),
		strings.ToLower(mapping.industry),
	)

	// Incorporate company context if available
	if companyInfo.Industry != "" {
		reasoning += fmt.Sprintf(" Company context (%s industry) reinforces this design direction.", companyInfo.Industry)
	}

	return &DesignIdentity{
		Industry:        mapping.industry,
		Formality:       mapping.formality,
		Style:          mapping.style,
		ColorPreference: mapping.colorPreference,
		Audience:       mapping.audience,
		VisualMetaphor: mapping.visualMetaphor,
		EmotionalTone:  mapping.emotionalTone,
		Reasoning:      reasoning,
	}
}

func (a *AIDesignAnalyzer) GetThemeName(theme ThemeType) string {
	names := map[ThemeType]string{
		ThemeTechnology: "Technology",
		ThemeBusiness:   "Business",
		ThemeSecurity:   "Security",
		ThemeInnovation: "Innovation",
		ThemeHealthcare: "Healthcare",
		ThemeFinance:    "Finance",
		ThemeGovernment: "Government",
		ThemeEducation:  "Education",
	}
	return names[theme]
}