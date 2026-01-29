package assets

import (
	"testing"
)

func TestSmartContentAnalyzer_AnalyzeContent(t *testing.T) {
	analyzer := NewSmartContentAnalyzer()

	tests := []struct {
		name           string
		content        string
		expectedType   ContentType
		expectedSentiment string
		expectedComplexity string
		expectedNumbers bool
	}{
		{
			name:           "Simple text",
			content:        "Hello world",
			expectedType:   ContentTextHeavy,
			expectedSentiment: "neutral",
			expectedComplexity: "simple",
			expectedNumbers: false,
		},
		{
			name:           "Quote content",
			content:        "\"Innovation is key to success\", said the CEO",
			expectedType:   ContentQuote,
			expectedSentiment: "positive",
			expectedComplexity: "simple",
			expectedNumbers: false,
		},
		{
			name:           "Urgent content",
			content:        "URGENT: Critical security breach detected immediately",
			expectedType:   ContentTextHeavy,
			expectedSentiment: "urgent",
			expectedComplexity: "simple",
			expectedNumbers: false,
		},
		{
			name:           "Complex technical content",
			content:        "The microservices architecture utilizes containerized deployment with Kubernetes orchestration, implementing distributed systems patterns for scalability and fault tolerance across multiple availability zones",
			expectedType:   ContentTextHeavy,
			expectedSentiment: "neutral",
			expectedComplexity: "medium",
			expectedNumbers: false,
		},
		{
			name:           "Content with numbers",
			content:        "Revenue data shows 25% increase to $1.2M in Q4 2023",
			expectedType:   ContentDataDriven,
			expectedSentiment: "positive",
			expectedComplexity: "simple",
			expectedNumbers: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.AnalyzeContent(tt.content)

			if result.ContentType != tt.expectedType {
				t.Errorf("Expected content type %v, got %v", tt.expectedType, result.ContentType)
			}

			if result.Sentiment != tt.expectedSentiment {
				t.Errorf("Expected sentiment %s, got %s", tt.expectedSentiment, result.Sentiment)
			}

			if result.Complexity != tt.expectedComplexity {
				t.Errorf("Expected complexity %s, got %s", tt.expectedComplexity, result.Complexity)
			}

			if result.HasNumbers != tt.expectedNumbers {
				t.Errorf("Expected has_numbers %v, got %v", tt.expectedNumbers, result.HasNumbers)
			}

			if result.WordCount <= 0 {
				t.Error("Expected positive word count")
			}
		})
	}
}

func TestSmartContentAnalyzer_DetermineSentiment(t *testing.T) {
	analyzer := NewSmartContentAnalyzer()

	tests := []struct {
		name      string
		content   string
		expected  string
	}{
		{
			name:     "Positive sentiment",
			content:  "Great success! Our growth exceeded expectations",
			expected: "positive",
		},
		{
			name:     "Negative sentiment",
			content:  "We face several problems and risks in this project",
			expected: "negative",
		},
		{
			name:     "Urgent sentiment",
			content:  "This is urgent! We need immediate action for the deadline",
			expected: "urgent",
		},
		{
			name:     "Neutral sentiment",
			content:  "Here are the quarterly results for review",
			expected: "neutral",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.AnalyzeContent(tt.content)
			if result.Sentiment != tt.expected {
				t.Errorf("Expected sentiment %s, got %s", tt.expected, result.Sentiment)
			}
		})
	}
}

func TestSmartContentAnalyzer_ExtractKeyConcepts(t *testing.T) {
	analyzer := NewSmartContentAnalyzer()

	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "Business concepts",
			content:  "Our strategy focuses on market growth and customer value",
			expected: []string{"business"},
		},
		{
			name:     "Technology concepts",
			content:  "The API connects to our cloud database system",
			expected: []string{"technology"},
		},
		{
			name:     "Finance concepts",
			content:  "Budget analysis shows investment returns and costs",
			expected: []string{"finance"},
		},
		{
			name:     "Multiple concepts",
			content:  "Our technology platform drives business growth through smart investment",
			expected: []string{"business", "technology", "finance"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.AnalyzeContent(tt.content)

			// Check if all expected concepts are present
			for _, expected := range tt.expected {
				found := false
				for _, concept := range result.KeyConcepts {
					if concept == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected concept %s not found in %v", expected, result.KeyConcepts)
				}
			}
		})
	}
}

func TestSmartLayoutGenerator_GenerateLayout(t *testing.T) {
	generator := NewSmartLayoutGenerator()

	tests := []struct {
		name            string
		title           string
		content         string
		expectedType    LayoutType
		expectedPlaceholders int
	}{
		{
			name:            "Basic content layout",
			title:           "Project Overview",
			content:         "This project aims to improve our systems",
			expectedType:    LayoutContent,
			expectedPlaceholders: 1,
		},
		{
			name:            "Comparison layout",
			title:           "Product Comparison",
			content:         "Product A vs Product B: performance differences",
			expectedType:    LayoutComparison,
			expectedPlaceholders: 2,
		},
		{
			name:            "Timeline layout",
			title:           "Project Timeline",
			content:         "First phase: planning, then development, finally deployment",
			expectedType:    LayoutTimeline,
			expectedPlaceholders: 1,
		},
		{
			name:            "Quote layout",
			title:           "Inspiration",
			content:         "\"Innovation distinguishes between a leader and a follower.\"",
			expectedType:    LayoutQuote,
			expectedPlaceholders: 2, // quote + attribution
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layout := generator.GenerateLayout(tt.title, tt.content, 1, 5)

			if layout.Type != tt.expectedType {
				t.Errorf("Expected layout type %v, got %v", tt.expectedType, layout.Type)
			}

			if len(layout.Content) != tt.expectedPlaceholders {
				t.Errorf("Expected %d content placeholders, got %d", tt.expectedPlaceholders, len(layout.Content))
			}

			// Verify title placeholder is present
			if layout.Title.ID != "title" {
				t.Errorf("Expected title placeholder with ID 'title', got '%s'", layout.Title.ID)
			}

			// Verify color scheme is set
			if layout.ColorScheme.Primary == "" {
				t.Error("Expected primary color to be set")
			}
		})
	}
}

func TestSmartLayoutGenerator_ColorSchemeSelection(t *testing.T) {
	generator := NewSmartLayoutGenerator()

	tests := []struct {
		name     string
		content  string
		expected string // Expected primary color pattern
	}{
		{
			name:     "Positive content gets green colors",
			content:  "Excellent growth and great success in our expansion",
			expected: "#2E8B57", // Sea Green for positive
		},
		{
			name:     "Urgent content gets red colors",
			content:  "Urgent deadline! We need immediate action now",
			expected: "#DC143C", // Crimson for urgent
		},
		{
			name:     "Negative content gets blue colors",
			content:  "Several problems and risks need attention",
			expected: "#4682B4", // Steel Blue for negative
		},
		{
			name:     "Neutral content gets default colors",
			content:  "Here are the quarterly metrics for review",
			expected: "#2C3E50", // Dark Blue Gray for neutral
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			layout := generator.GenerateLayout("Test", tt.content, 1, 1)

			if layout.ColorScheme.Primary != tt.expected {
				t.Errorf("Expected primary color %s, got %s", tt.expected, layout.ColorScheme.Primary)
			}
		})
	}
}