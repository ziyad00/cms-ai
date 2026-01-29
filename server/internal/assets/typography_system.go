package assets

import (
	"strings"

	"baliance.com/gooxml/color"
	"baliance.com/gooxml/measurement"
	"baliance.com/gooxml/presentation"
)

type FontFamily int

const (
	FontCalibri FontFamily = iota
	FontArial
	FontSegoeUI
	FontTimesNewRoman
	FontVerdana
	FontHelvetica
	FontGeorgia
	FontTahoma
)

type TextStyle int

const (
	StyleTitleSlide TextStyle = iota
	StyleSlideTitle
	StyleBodyText
	StyleCaption
	StyleEmphasis
	StyleQuote
	StyleCode
	StyleListItem
)

type TypographyRule struct {
	FontFamily     FontFamily `json:"font_family"`
	FontSize       int        `json:"font_size"`
	Bold           bool       `json:"bold"`
	Italic         bool       `json:"italic"`
	Color          string     `json:"color"`
	LineHeight     float64    `json:"line_height"`
	LetterSpacing  float64    `json:"letter_spacing"`
	TextTransform  string     `json:"text_transform"` // uppercase, lowercase, capitalize
}

type AdvancedTypographySystem struct {
	themeRules     map[string]map[TextStyle]TypographyRule
	fontMappings   map[FontFamily]string
	contentAnalyzer *SmartContentAnalyzer
}

func NewAdvancedTypographySystem() *AdvancedTypographySystem {
	system := &AdvancedTypographySystem{
		themeRules:      make(map[string]map[TextStyle]TypographyRule),
		contentAnalyzer: NewSmartContentAnalyzer(),
	}

	system.initializeFontMappings()
	system.initializeThemeRules()

	return system
}

func (t *AdvancedTypographySystem) initializeFontMappings() {
	t.fontMappings = map[FontFamily]string{
		FontCalibri:       "Calibri",
		FontArial:         "Arial",
		FontSegoeUI:       "Segoe UI",
		FontTimesNewRoman: "Times New Roman",
		FontVerdana:       "Verdana",
		FontHelvetica:     "Helvetica",
		FontGeorgia:       "Georgia",
		FontTahoma:        "Tahoma",
	}
}

func (t *AdvancedTypographySystem) initializeThemeRules() {
	// Corporate Professional Typography
	t.themeRules["Corporate Professional"] = map[TextStyle]TypographyRule{
		StyleTitleSlide: {FontCalibri, 36, true, false, "#2E75B6", 1.2, 0, ""},
		StyleSlideTitle: {FontCalibri, 24, true, false, "#2E75B6", 1.3, 0, ""},
		StyleBodyText:   {FontCalibri, 14, false, false, "#2C3E50", 1.5, 0, ""},
		StyleCaption:    {FontCalibri, 11, false, false, "#5A6C7D", 1.4, 0, ""},
		StyleEmphasis:   {FontCalibri, 14, true, false, "#3498DB", 1.5, 0, ""},
		StyleQuote:      {FontCalibri, 16, false, true, "#2E75B6", 1.6, 0.5, ""},
	}

	// Modern Tech Typography
	t.themeRules["Modern Tech"] = map[TextStyle]TypographyRule{
		StyleTitleSlide: {FontSegoeUI, 40, true, false, "#667EEA", 1.1, 0, ""},
		StyleSlideTitle: {FontSegoeUI, 28, true, false, "#667EEA", 1.2, 0, ""},
		StyleBodyText:   {FontSegoeUI, 16, false, false, "#1A202C", 1.4, 0, ""},
		StyleCaption:    {FontSegoeUI, 12, false, false, "#764BA2", 1.3, 0, ""},
		StyleEmphasis:   {FontSegoeUI, 16, true, false, "#4FD1C7", 1.4, 0, ""},
		StyleCode:       {FontTahoma, 12, false, false, "#1A202C", 1.2, 0, ""},
	}

	// Healthcare Professional Typography
	t.themeRules["Healthcare Professional"] = map[TextStyle]TypographyRule{
		StyleTitleSlide: {FontArial, 32, true, false, "#2D7DB3", 1.3, 0, ""},
		StyleSlideTitle: {FontArial, 24, true, false, "#2D7DB3", 1.3, 0, ""},
		StyleBodyText:   {FontArial, 14, false, false, "#2D3748", 1.6, 0, ""},
		StyleCaption:    {FontArial, 11, false, false, "#38B2AC", 1.5, 0, ""},
		StyleEmphasis:   {FontArial, 14, true, false, "#48BB78", 1.6, 0, ""},
	}

	// Financial Services Typography
	t.themeRules["Financial Services"] = map[TextStyle]TypographyRule{
		StyleTitleSlide: {FontTimesNewRoman, 36, true, false, "#1B5E20", 1.2, 0, ""},
		StyleSlideTitle: {FontTimesNewRoman, 26, true, false, "#1B5E20", 1.3, 0, ""},
		StyleBodyText:   {FontTimesNewRoman, 14, false, false, "#1B5E20", 1.5, 0, ""},
		StyleCaption:    {FontTimesNewRoman, 12, false, false, "#2E7D32", 1.4, 0, ""},
		StyleEmphasis:   {FontTimesNewRoman, 14, true, false, "#FFB300", 1.5, 0, ""},
	}

	// Cybersecurity Typography
	t.themeRules["Cybersecurity"] = map[TextStyle]TypographyRule{
		StyleTitleSlide: {FontArial, 34, true, false, "#C53030", 1.1, 0, "uppercase"},
		StyleSlideTitle: {FontArial, 26, true, false, "#E53E3E", 1.2, 0, ""},
		StyleBodyText:   {FontArial, 14, false, false, "#F7FAFC", 1.4, 0, ""},
		StyleCaption:    {FontArial, 11, false, false, "#4A5568", 1.3, 0, ""},
		StyleEmphasis:   {FontArial, 14, true, false, "#E53E3E", 1.4, 0, "uppercase"},
	}

	// Educational Typography
	t.themeRules["Educational"] = map[TextStyle]TypographyRule{
		StyleTitleSlide: {FontVerdana, 32, true, false, "#2B6CB0", 1.3, 0, ""},
		StyleSlideTitle: {FontVerdana, 22, true, false, "#2B6CB0", 1.4, 0, ""},
		StyleBodyText:   {FontVerdana, 14, false, false, "#2D3748", 1.6, 0, ""},
		StyleCaption:    {FontVerdana, 11, false, false, "#ED8936", 1.5, 0, ""},
		StyleEmphasis:   {FontVerdana, 14, true, false, "#38A169", 1.6, 0, ""},
	}
}

func (t *AdvancedTypographySystem) ApplyTypography(textBox presentation.TextBox, content string, style TextStyle, themeName string) error {
	// Get typography rule for theme and style
	rule, exists := t.getTypographyRule(themeName, style)
	if !exists {
		// Fallback to corporate theme
		rule, _ = t.getTypographyRule("Corporate Professional", style)
	}

	// Analyze content for dynamic adjustments
	analysis := t.contentAnalyzer.AnalyzeContent(content)
	adjustedRule := t.adjustRuleForContent(rule, analysis, content)

	// Apply typography to text box
	return t.applyRuleToTextBox(textBox, content, adjustedRule)
}

func (t *AdvancedTypographySystem) getTypographyRule(themeName string, style TextStyle) (TypographyRule, bool) {
	themeRules, themeExists := t.themeRules[themeName]
	if !themeExists {
		return TypographyRule{}, false
	}

	rule, styleExists := themeRules[style]
	return rule, styleExists
}

func (t *AdvancedTypographySystem) adjustRuleForContent(rule TypographyRule, analysis ContentAnalysis, content string) TypographyRule {
	adjustedRule := rule

	// Adjust font size based on content length and complexity
	wordCount := analysis.WordCount

	if wordCount > 100 && rule.FontSize > 16 {
		// Reduce font size for long content
		adjustedRule.FontSize = rule.FontSize - 2
	} else if wordCount < 20 && rule.FontSize < 24 {
		// Increase font size for short content
		adjustedRule.FontSize = rule.FontSize + 2
	}

	// Adjust based on content urgency
	if analysis.Sentiment == "urgent" {
		adjustedRule.Bold = true
		if rule.FontSize < 20 {
			adjustedRule.FontSize += 2
		}
	}

	// Adjust for numeric content
	if analysis.HasNumbers {
		adjustedRule.FontFamily = FontTahoma // Better for numbers
		if strings.Contains(strings.ToLower(content), "$") ||
		   strings.Contains(strings.ToLower(content), "%") {
			adjustedRule.Bold = true
		}
	}

	// Adjust line height for complexity
	switch analysis.Complexity {
	case "high":
		adjustedRule.LineHeight = rule.LineHeight + 0.2
	case "simple":
		adjustedRule.LineHeight = rule.LineHeight - 0.1
		if adjustedRule.LineHeight < 1.1 {
			adjustedRule.LineHeight = 1.1
		}
	}

	return adjustedRule
}

func (t *AdvancedTypographySystem) applyRuleToTextBox(textBox presentation.TextBox, content string, rule TypographyRule) error {
	// Apply text transformation
	processedContent := t.applyTextTransform(content, rule.TextTransform)

	// Split content into lines
	lines := strings.Split(processedContent, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		para := textBox.AddParagraph()

		// Add bullet for multi-line content (except first line)
		if len(lines) > 1 && i > 0 && !rule.Bold {
			para.Properties().SetBulletChar("•")
		}

		run := para.AddRun()
		run.SetText(line)

		// Apply typography properties
		rp := run.Properties()

		// Font family
		fontName := t.fontMappings[rule.FontFamily]
		rp.SetFont(fontName)

		// Font size
		rp.SetSize(measurement.Distance(rule.FontSize) * measurement.Point)

		// Font weight
		if rule.Bold {
			rp.SetBold(true)
		}

		// Font style (italic not available in this gooxml version)
		// if rule.Italic {
		// 	rp.SetItalic(true)
		// }

		// Text color
		if rule.Color != "" {
			color := t.parseHexColor(rule.Color)
			rp.SetSolidFill(color)
		}
	}

	return nil
}

func (t *AdvancedTypographySystem) applyTextTransform(text string, transform string) string {
	switch transform {
	case "uppercase":
		return strings.ToUpper(text)
	case "lowercase":
		return strings.ToLower(text)
	case "capitalize":
		return strings.Title(strings.ToLower(text))
	default:
		return text
	}
}

func (t *AdvancedTypographySystem) GetOptimalStyle(content string, position string, theme string) TextStyle {
	analysis := t.contentAnalyzer.AnalyzeContent(content)

	// Determine style based on position and content analysis
	switch {
	case strings.Contains(strings.ToLower(position), "title") && analysis.WordCount < 10:
		return StyleTitleSlide
	case strings.Contains(strings.ToLower(position), "title"):
		return StyleSlideTitle
	case analysis.ContentType == ContentQuote:
		return StyleQuote
	case analysis.HasNumbers && analysis.WordCount < 50:
		return StyleEmphasis
	case strings.Contains(strings.ToLower(content), "code") ||
		 strings.Contains(strings.ToLower(content), "api"):
		return StyleCode
	case analysis.WordCount < 30:
		return StyleCaption
	default:
		return StyleBodyText
	}
}

func (t *AdvancedTypographySystem) GenerateTypographyReport(content string, theme string) map[string]any {
	analysis := t.contentAnalyzer.AnalyzeContent(content)

	recommendations := []string{}

	// Font size recommendations
	if analysis.WordCount > 100 {
		recommendations = append(recommendations, "Consider reducing font size for long content")
	}

	// Font family recommendations
	if analysis.HasNumbers {
		recommendations = append(recommendations, "Use monospace or sans-serif fonts for numerical content")
	}

	// Hierarchy recommendations
	if analysis.Complexity == "high" {
		recommendations = append(recommendations, "Establish clear typography hierarchy for complex content")
	}

	// Readability recommendations
	if analysis.Sentiment == "urgent" {
		recommendations = append(recommendations, "Use bold typography and increased font size for urgent content")
	}

	return map[string]any{
		"word_count":      analysis.WordCount,
		"complexity":      analysis.Complexity,
		"has_numbers":     analysis.HasNumbers,
		"content_type":    analysis.ContentType,
		"sentiment":       analysis.Sentiment,
		"recommendations": recommendations,
		"suggested_styles": map[string]string{
			"title":   t.getStyleName(StyleSlideTitle),
			"body":    t.getStyleName(StyleBodyText),
			"caption": t.getStyleName(StyleCaption),
		},
	}
}

func (t *AdvancedTypographySystem) getStyleName(style TextStyle) string {
	names := map[TextStyle]string{
		StyleTitleSlide: "Title Slide",
		StyleSlideTitle: "Slide Title",
		StyleBodyText:   "Body Text",
		StyleCaption:    "Caption",
		StyleEmphasis:   "Emphasis",
		StyleQuote:      "Quote",
		StyleCode:       "Code",
		StyleListItem:   "List Item",
	}
	return names[style]
}

func (t *AdvancedTypographySystem) parseHexColor(hexColor string) color.Color {
	// Use gooxml's color parsing
	return color.FromHex(hexColor)
}