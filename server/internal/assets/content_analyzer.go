package assets

import (
	"regexp"
	"strings"
)

type ContentType int

const (
	ContentTextHeavy ContentType = iota
	ContentDataDriven
	ContentListItems
	ContentComparison
	ContentTimeline
	ContentHierarchy
	ContentQuote
	ContentImageText
)

type ContentAnalysis struct {
	ContentType     ContentType
	WordCount       int
	Complexity      string // simple, medium, complex
	Sentiment       string // positive, neutral, negative, urgent
	KeyConcepts     []string
	HasNumbers      bool
	HasDates        bool
	HierarchyLevel  int
	VisualWeight    float64
}

type SmartContentAnalyzer struct{}

func NewSmartContentAnalyzer() *SmartContentAnalyzer {
	return &SmartContentAnalyzer{}
}

func (a *SmartContentAnalyzer) AnalyzeContent(text string) ContentAnalysis {
	words := strings.Fields(text)
	wordCount := len(words)

	analysis := ContentAnalysis{
		WordCount:    wordCount,
		HasNumbers:   a.containsNumbers(text),
		HasDates:     a.containsDates(text),
		KeyConcepts:  a.extractKeyConcepts(text),
		Complexity:   a.determineComplexity(text, wordCount),
		Sentiment:    a.determineSentiment(text),
		ContentType:  a.determineContentType(text),
		VisualWeight: a.calculateVisualWeight(text, wordCount),
	}

	return analysis
}

func (a *SmartContentAnalyzer) containsNumbers(text string) bool {
	numberPattern := regexp.MustCompile(`\d+[%$]?|\$\d+|\d+\.\d+`)
	return numberPattern.MatchString(text)
}

func (a *SmartContentAnalyzer) containsDates(text string) bool {
	datePatterns := []*regexp.Regexp{
		regexp.MustCompile(`\d{1,2}/\d{1,2}/\d{2,4}`),
		regexp.MustCompile(`\d{4}-\d{1,2}-\d{1,2}`),
		regexp.MustCompile(`(?i)(january|february|march|april|may|june|july|august|september|october|november|december)`),
		regexp.MustCompile(`(?i)(jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec)\s+\d{1,2}`),
		regexp.MustCompile(`\d{4}`), // Years
	}

	for _, pattern := range datePatterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	return false
}

func (a *SmartContentAnalyzer) extractKeyConcepts(text string) []string {
	concepts := []string{}
	lowerText := strings.ToLower(text)

	// Business concepts
	businessTerms := []string{"strategy", "growth", "revenue", "roi", "market", "customer", "value", "profit"}
	for _, term := range businessTerms {
		if strings.Contains(lowerText, term) {
			concepts = append(concepts, "business")
			break
		}
	}

	// Technology concepts
	techTerms := []string{"api", "database", "cloud", "digital", "software", "platform", "system", "technology"}
	for _, term := range techTerms {
		if strings.Contains(lowerText, term) {
			concepts = append(concepts, "technology")
			break
		}
	}

	// Financial concepts
	financeTerms := []string{"budget", "cost", "investment", "financial", "money", "pricing", "funding"}
	for _, term := range financeTerms {
		if strings.Contains(lowerText, term) {
			concepts = append(concepts, "finance")
			break
		}
	}

	return concepts
}

func (a *SmartContentAnalyzer) determineComplexity(text string, wordCount int) string {
	if wordCount < 20 {
		return "simple"
	}
	if wordCount < 100 {
		return "medium"
	}
	return "complex"
}

func (a *SmartContentAnalyzer) determineSentiment(text string) string {
	lowerText := strings.ToLower(text)

	positiveWords := []string{"success", "growth", "improve", "excellent", "great", "increase", "opportunity"}
	negativeWords := []string{"problem", "issue", "decrease", "fail", "risk", "challenge", "concern"}
	urgentWords := []string{"urgent", "immediate", "critical", "asap", "deadline", "emergency"}

	positiveCount := 0
	negativeCount := 0
	urgentCount := 0

	for _, word := range positiveWords {
		if strings.Contains(lowerText, word) {
			positiveCount++
		}
	}

	for _, word := range negativeWords {
		if strings.Contains(lowerText, word) {
			negativeCount++
		}
	}

	for _, word := range urgentWords {
		if strings.Contains(lowerText, word) {
			urgentCount++
		}
	}

	if urgentCount > 0 {
		return "urgent"
	}
	if positiveCount > negativeCount {
		return "positive"
	}
	if negativeCount > positiveCount {
		return "negative"
	}
	return "neutral"
}

func (a *SmartContentAnalyzer) determineContentType(text string) ContentType {
	lowerText := strings.ToLower(text)

	// Check for lists (bullet points, numbers)
	if strings.Contains(text, "•") || strings.Contains(text, "-") ||
	   regexp.MustCompile(`^\d+\.`).MatchString(text) {
		return ContentListItems
	}

	// Check for comparisons
	comparisonWords := []string{"vs", "versus", "compared to", "difference", "better", "worse"}
	for _, word := range comparisonWords {
		if strings.Contains(lowerText, word) {
			return ContentComparison
		}
	}

	// Check for timeline content
	timelineWords := []string{"first", "then", "next", "finally", "step", "phase", "timeline"}
	for _, word := range timelineWords {
		if strings.Contains(lowerText, word) {
			return ContentTimeline
		}
	}

	// Check for data-driven content
	if a.containsNumbers(text) && (strings.Contains(lowerText, "chart") ||
	   strings.Contains(lowerText, "graph") || strings.Contains(lowerText, "data")) {
		return ContentDataDriven
	}

	// Check for quotes
	if strings.Contains(text, "\"") || strings.Contains(text, "'") {
		return ContentQuote
	}

	// Default to text heavy
	return ContentTextHeavy
}

func (a *SmartContentAnalyzer) calculateVisualWeight(text string, wordCount int) float64 {
	// Base weight on word count
	weight := float64(wordCount) / 100.0

	// Adjust based on content characteristics
	lowerText := strings.ToLower(text)

	if strings.Contains(lowerText, "important") || strings.Contains(lowerText, "key") {
		weight += 0.3
	}

	if a.containsNumbers(text) {
		weight += 0.2
	}

	// Clamp between 0 and 1
	if weight > 1.0 {
		weight = 1.0
	}
	if weight < 0.1 {
		weight = 0.1
	}

	return weight
}