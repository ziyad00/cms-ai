package assets

import (
	"testing"
)

func TestAIDesignAnalyzer_AnalyzeContentThemes(t *testing.T) {
	analyzer := NewAIDesignAnalyzer()

	tests := []struct {
		name           string
		content        string
		expectedTheme  ThemeType
		expectedStrength int
	}{
		{
			name:          "Technology content",
			content:       "Our API architecture uses cloud databases and backend systems for digital platform development",
			expectedTheme: ThemeTechnology,
			expectedStrength: 7,
		},
		{
			name:          "Business content",
			content:       "Strategic governance for stakeholder management and ROI optimization in corporate revenue markets",
			expectedTheme: ThemeBusiness,
			expectedStrength: 7,
		},
		{
			name:          "Security content",
			content:       "Security compliance requires encryption and risk authentication to prevent vulnerability threats",
			expectedTheme: ThemeSecurity,
			expectedStrength: 7,
		},
		{
			name:          "Healthcare content",
			content:       "Medical diagnosis and patient treatment in clinical healthcare pharmaceutical therapy",
			expectedTheme: ThemeHealthcare,
			expectedStrength: 7,
		},
		{
			name:          "Finance content",
			content:       "Financial investment banking budget costs and capital asset portfolio trading",
			expectedTheme: ThemeFinance,
			expectedStrength: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData := map[string]any{
				"slides": []any{
					map[string]any{
						"title": tt.content,
						"content": []any{tt.content},
					},
				},
			}

			analysis := analyzer.analyzeContentThemes(tt.content, jsonData)

			if analysis.DominantTheme != tt.expectedTheme {
				t.Errorf("Expected theme %v, got %v", tt.expectedTheme, analysis.DominantTheme)
			}

			if analysis.ThemeStrength < tt.expectedStrength {
				t.Errorf("Expected strength >= %d, got %d", tt.expectedStrength, analysis.ThemeStrength)
			}

			if analysis.SlideCount != 1 {
				t.Errorf("Expected 1 slide, got %d", analysis.SlideCount)
			}
		})
	}
}

func TestAIDesignAnalyzer_GenerateDesignIdentity(t *testing.T) {
	analyzer := NewAIDesignAnalyzer()

	analysis := ContentThemeAnalysis{
		DominantTheme: ThemeTechnology,
		ThemeStrength: 5,
		Complexity:    "medium",
		SlideCount:    10,
		ContentLength: 500,
	}

	company := CompanyContext{
		Industry: "Software Development",
		Name:     "TechCorp",
	}

	identity := analyzer.generateDesignIdentity(analysis, company)

	if identity.Industry != "Technology/Software" {
		t.Errorf("Expected Technology/Software industry, got %s", identity.Industry)
	}

	if identity.Formality != "Modern Professional" {
		t.Errorf("Expected Modern Professional formality, got %s", identity.Formality)
	}

	if identity.ColorPreference != "Cool blues and grays with electric accents" {
		t.Errorf("Unexpected color preference: %s", identity.ColorPreference)
	}

	if identity.Reasoning == "" {
		t.Error("Expected reasoning to be populated")
	}
}

func TestAIDesignAnalyzer_ExtractKeyConcepts(t *testing.T) {
	analyzer := NewAIDesignAnalyzer()

	content := "architecture database development platform technology systems"
	concepts := analyzer.extractKeyConcepts(content, 3)

	if len(concepts) > 3 {
		t.Errorf("Expected max 3 concepts, got %d", len(concepts))
	}

	// Should extract words longer than 6 characters
	expectedConcepts := []string{"architecture", "database", "development"}
	for i, concept := range concepts {
		if i < len(expectedConcepts) && concept != expectedConcepts[i] {
			t.Errorf("Expected concept %s, got %s", expectedConcepts[i], concept)
		}
	}
}

func TestAIDesignAnalyzer_GetThemeName(t *testing.T) {
	analyzer := NewAIDesignAnalyzer()

	tests := []struct {
		theme ThemeType
		expected string
	}{
		{ThemeTechnology, "Technology"},
		{ThemeBusiness, "Business"},
		{ThemeSecurity, "Security"},
		{ThemeHealthcare, "Healthcare"},
		{ThemeFinance, "Finance"},
	}

	for _, tt := range tests {
		name := analyzer.GetThemeName(tt.theme)
		if name != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, name)
		}
	}
}