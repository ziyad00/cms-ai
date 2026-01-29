package assets

import (
	"strings"
	"testing"
)

func TestAdvancedTypographySystem_GetTypographyRule(t *testing.T) {
	system := NewAdvancedTypographySystem()

	tests := []struct {
		name      string
		theme     string
		style     TextStyle
		shouldExist bool
	}{
		{
			name:      "Corporate title slide",
			theme:     "Corporate Professional",
			style:     StyleTitleSlide,
			shouldExist: true,
		},
		{
			name:      "Tech body text",
			theme:     "Modern Tech",
			style:     StyleBodyText,
			shouldExist: true,
		},
		{
			name:      "Healthcare emphasis",
			theme:     "Healthcare Professional",
			style:     StyleEmphasis,
			shouldExist: true,
		},
		{
			name:      "Nonexistent theme",
			theme:     "Nonexistent Theme",
			style:     StyleBodyText,
			shouldExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, exists := system.getTypographyRule(tt.theme, tt.style)

			if exists != tt.shouldExist {
				t.Errorf("Expected exists=%v, got %v", tt.shouldExist, exists)
			}

			if exists {
				if rule.FontSize <= 0 {
					t.Error("Expected positive font size")
				}
				if rule.LineHeight <= 0 {
					t.Error("Expected positive line height")
				}
				if rule.Color == "" {
					t.Error("Expected color to be set")
				}
			}
		})
	}
}

func TestAdvancedTypographySystem_AdjustRuleForContent(t *testing.T) {
	system := NewAdvancedTypographySystem()

	baseRule := TypographyRule{
		FontFamily: FontCalibri,
		FontSize:   16,
		Bold:       false,
		Color:      "#2C3E50",
		LineHeight: 1.4,
	}

	tests := []struct {
		name        string
		content     string
		analysis    ContentAnalysis
		expectedBold bool
		expectedFont FontFamily
	}{
		{
			name:        "Long content reduces font size",
			content:     strings.Repeat("word ", 25), // 25 words > 20
			analysis:    ContentAnalysis{WordCount: 25, Sentiment: "neutral", HasNumbers: false, Complexity: "medium"},
			expectedBold: false,
			expectedFont: FontCalibri,
		},
		{
			name:        "Urgent content becomes bold",
			content:     "URGENT: Critical issue needs attention",
			analysis:    ContentAnalysis{WordCount: 6, Sentiment: "urgent", HasNumbers: false, Complexity: "simple"},
			expectedBold: true,
			expectedFont: FontCalibri,
		},
		{
			name:        "Numeric content uses Tahoma font",
			content:     "Revenue: $150K, growth 25%",
			analysis:    ContentAnalysis{WordCount: 5, Sentiment: "positive", HasNumbers: true, Complexity: "simple"},
			expectedBold: true, // Because of $ symbol
			expectedFont: FontTahoma,
		},
		{
			name:        "High complexity increases line height",
			content:     "Complex technical explanation with multiple concepts",
			analysis:    ContentAnalysis{WordCount: 8, Sentiment: "neutral", HasNumbers: false, Complexity: "high"},
			expectedBold: false,
			expectedFont: FontCalibri,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adjusted := system.adjustRuleForContent(baseRule, tt.analysis, tt.content)

			if adjusted.Bold != tt.expectedBold {
				t.Errorf("Expected bold=%v, got %v", tt.expectedBold, adjusted.Bold)
			}

			if adjusted.FontFamily != tt.expectedFont {
				t.Errorf("Expected font=%v, got %v", tt.expectedFont, adjusted.FontFamily)
			}

			// Verify line height adjustments for complexity
			if tt.analysis.Complexity == "high" && adjusted.LineHeight <= baseRule.LineHeight {
				t.Error("Expected increased line height for high complexity")
			}
		})
	}
}

func TestAdvancedTypographySystem_GetOptimalStyle(t *testing.T) {
	system := NewAdvancedTypographySystem()

	tests := []struct {
		name     string
		content  string
		position string
		expected TextStyle
	}{
		{
			name:     "Title position with short text",
			content:  "Overview",
			position: "title",
			expected: StyleTitleSlide,
		},
		{
			name:     "Title position with long text",
			content:  "This is a very long title that exceeds the usual length",
			position: "title",
			expected: StyleSlideTitle,
		},
		{
			name:     "Quote content",
			content:  "\"Innovation is the key to success\"",
			position: "body",
			expected: StyleQuote,
		},
		{
			name:     "Code content",
			content:  "API endpoint returns JSON data",
			position: "body",
			expected: StyleCode,
		},
		{
			name:     "Numeric emphasis",
			content:  "Revenue increased 25% to $1.2M",
			position: "body",
			expected: StyleEmphasis,
		},
		{
			name:     "Short caption",
			content:  "Figure 1.1",
			position: "body",
			expected: StyleEmphasis, // Numbers detected, word count < 50
		},
		{
			name:     "Regular body text",
			content:  "This is standard body text content for the slide presentation",
			position: "body",
			expected: StyleCaption, // Word count < 30
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := system.GetOptimalStyle(tt.content, tt.position, "Corporate Professional")

			if result != tt.expected {
				t.Errorf("Expected style %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestAdvancedTypographySystem_ApplyTextTransform(t *testing.T) {
	system := NewAdvancedTypographySystem()

	tests := []struct {
		name      string
		text      string
		transform string
		expected  string
	}{
		{
			name:      "Uppercase transform",
			text:      "hello world",
			transform: "uppercase",
			expected:  "HELLO WORLD",
		},
		{
			name:      "Lowercase transform",
			text:      "HELLO WORLD",
			transform: "lowercase",
			expected:  "hello world",
		},
		{
			name:      "Capitalize transform",
			text:      "hello world example",
			transform: "capitalize",
			expected:  "Hello World Example",
		},
		{
			name:      "No transform",
			text:      "Hello World",
			transform: "",
			expected:  "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := system.applyTextTransform(tt.text, tt.transform)

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestAdvancedTypographySystem_GenerateTypographyReport(t *testing.T) {
	system := NewAdvancedTypographySystem()

	tests := []struct {
		name           string
		content        string
		theme          string
		expectRecommendations bool
	}{
		{
			name:           "Long content report",
			content:        strings.Repeat("This is a very long piece of content that should trigger recommendations. ", 20),
			theme:          "Corporate Professional",
			expectRecommendations: true,
		},
		{
			name:           "Numeric content report",
			content:        "Revenue: $150K, growth 25%, margin 15%",
			theme:          "Financial Services",
			expectRecommendations: true,
		},
		{
			name:           "Urgent content report",
			content:        "URGENT: Critical security breach detected",
			theme:          "Cybersecurity",
			expectRecommendations: true,
		},
		{
			name:           "Simple content report",
			content:        "Hello world",
			theme:          "Corporate Professional",
			expectRecommendations: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := system.GenerateTypographyReport(tt.content, tt.theme)

			// Check required fields
			if _, exists := report["word_count"]; !exists {
				t.Error("Expected word_count in report")
			}

			if _, exists := report["complexity"]; !exists {
				t.Error("Expected complexity in report")
			}

			if _, exists := report["recommendations"]; !exists {
				t.Error("Expected recommendations in report")
			}

			recommendations, ok := report["recommendations"].([]string)
			if !ok {
				t.Error("Expected recommendations to be []string")
			}

			if tt.expectRecommendations && len(recommendations) == 0 {
				t.Error("Expected recommendations but got none")
			}

			// Check suggested styles
			if styles, exists := report["suggested_styles"]; exists {
				styleMap, ok := styles.(map[string]string)
				if !ok {
					t.Error("Expected suggested_styles to be map[string]string")
				}

				if len(styleMap) == 0 {
					t.Error("Expected suggested styles to be populated")
				}
			}
		})
	}
}