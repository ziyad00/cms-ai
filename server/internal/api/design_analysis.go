package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ziyad/cms-ai/server/internal/assets"
)

type DesignAnalysisRequest struct {
	Content     string            `json:"content"`
	Title       string            `json:"title"`
	SlideNumber int               `json:"slide_number"`
	TotalSlides int               `json:"total_slides"`
	BrandKit    map[string]string `json:"brand_kit,omitempty"`
}

type DesignAnalysisResponse struct {
	LayoutType       string                     `json:"layout_type"`
	ColorScheme      assets.ColorScheme         `json:"color_scheme"`
	ContentAnalysis  assets.ContentAnalysis     `json:"content_analysis"`
	SmartLayout      assets.SmartLayout         `json:"smart_layout"`
	DesignSuggestions []string                  `json:"design_suggestions"`
}

type EnhancedDesignAnalysisResponse struct {
	LayoutType        string                     `json:"layout_type"`
	ColorScheme       assets.ColorScheme         `json:"color_scheme"`
	ContentAnalysis   assets.ContentAnalysis     `json:"content_analysis"`
	SmartLayout       assets.SmartLayout         `json:"smart_layout"`
	DesignSuggestions []string                   `json:"design_suggestions"`
	DesignIdentity    assets.DesignIdentity      `json:"design_identity"`
	RecommendedTheme  string                     `json:"recommended_theme"`
	ThemeDescription  string                     `json:"theme_description"`
	TypographyReport  map[string]any             `json:"typography_report"`
	IndustryElements  []assets.VisualElement     `json:"industry_elements"`
}

func (s *Server) AnalyzeDesign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DesignAnalysisRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Content == "" {
		http.Error(w, "Content is required", http.StatusBadRequest)
		return
	}

	// Create enhanced AI design system
	aiAnalyzer := assets.NewAIDesignAnalyzer()
	layoutGenerator := assets.NewSmartLayoutGenerator()
	templateLibrary := assets.NewDesignTemplateLibrary()
	typographySystem := assets.NewAdvancedTypographySystem()

	// Convert request to format expected by AI analyzer
	jsonData := map[string]any{
		"slides": []map[string]any{
			{
				"title":   req.Title,
				"content": []string{req.Content},
			},
		},
	}

	// Convert brand kit to company context
	companyInfo := assets.CompanyContext{}
	for k, v := range req.BrandKit {
		switch k {
		case "name":
			companyInfo.Name = v
		case "industry":
			companyInfo.Industry = v
		}
	}

	// Perform AI design analysis
	designIdentity, err := aiAnalyzer.AnalyzeContentForDesign(jsonData, companyInfo)
	if err != nil {
		http.Error(w, "Design analysis failed", http.StatusInternalServerError)
		return
	}

	// Get theme based on analysis
	designTheme := templateLibrary.GetThemeForAnalysis(designIdentity)

	// Analyze content for smart layout
	contentAnalyzer := assets.NewSmartContentAnalyzer()
	contentAnalysis := contentAnalyzer.AnalyzeContent(req.Content)

	// Generate smart layout with theme integration
	smartLayout := layoutGenerator.GenerateLayout(req.Title, req.Content, req.SlideNumber, req.TotalSlides)

	// Override layout colors with theme colors
	smartLayout.ColorScheme = assets.ColorScheme{
		Primary:    designTheme.Colors["primary"],
		Secondary:  designTheme.Colors["secondary"],
		Background: designTheme.Colors["background"],
		Text:       designTheme.Colors["text"],
		Accent:     designTheme.Colors["accent"],
	}

	// Generate typography recommendations
	typographyReport := typographySystem.GenerateTypographyReport(req.Content, designTheme.Name)

	// Generate enhanced design suggestions
	suggestions := generateEnhancedDesignSuggestions(contentAnalysis, smartLayout, *designIdentity, designTheme, typographyReport)

	// Prepare enhanced response
	response := EnhancedDesignAnalysisResponse{
		LayoutType:        getLayoutTypeName(smartLayout.Type),
		ColorScheme:       smartLayout.ColorScheme,
		ContentAnalysis:   contentAnalysis,
		SmartLayout:       smartLayout,
		DesignSuggestions: suggestions,
		DesignIdentity:    *designIdentity,
		RecommendedTheme:  designTheme.Name,
		ThemeDescription:  designTheme.Description,
		TypographyReport:  typographyReport,
		IndustryElements:  generateIndustryElements(designIdentity.Industry),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func getLayoutTypeName(layoutType assets.LayoutType) string {
	switch layoutType {
	case assets.LayoutTitle:
		return "title"
	case assets.LayoutContent:
		return "content"
	case assets.LayoutComparison:
		return "comparison"
	case assets.LayoutTimeline:
		return "timeline"
	case assets.LayoutMetrics:
		return "metrics"
	case assets.LayoutQuote:
		return "quote"
	case assets.LayoutDataVisualization:
		return "data_visualization"
	default:
		return "content"
	}
}

func generateDesignSuggestions(analysis assets.ContentAnalysis, layout assets.SmartLayout) []string {
	suggestions := []string{}

	// Suggestions based on content analysis
	switch analysis.ContentType {
	case assets.ContentDataDriven:
		suggestions = append(suggestions, "Consider adding charts or graphs to visualize the data")
		suggestions = append(suggestions, "Use larger fonts for key metrics")
	case assets.ContentListItems:
		suggestions = append(suggestions, "Consider using icons to enhance bullet points")
		suggestions = append(suggestions, "Space items evenly for better readability")
	case assets.ContentComparison:
		suggestions = append(suggestions, "Use contrasting colors to highlight differences")
		suggestions = append(suggestions, "Consider a side-by-side layout for clear comparison")
	case assets.ContentTimeline:
		suggestions = append(suggestions, "Use a horizontal timeline for better flow")
		suggestions = append(suggestions, "Add progress indicators or connecting lines")
	}

	// Suggestions based on sentiment
	switch analysis.Sentiment {
	case "positive":
		suggestions = append(suggestions, "Use vibrant colors to reinforce positive messaging")
	case "urgent":
		suggestions = append(suggestions, "Use bold typography and attention-grabbing colors")
		suggestions = append(suggestions, "Consider adding visual emphasis elements")
	case "negative":
		suggestions = append(suggestions, "Use calming colors to balance negative content")
	}

	// Suggestions based on complexity
	switch analysis.Complexity {
	case "complex":
		suggestions = append(suggestions, "Break content into smaller chunks for better readability")
		suggestions = append(suggestions, "Use hierarchy to guide the reader through complex information")
	case "simple":
		suggestions = append(suggestions, "Consider larger fonts and more white space")
		suggestions = append(suggestions, "Add visual elements to enhance the simple message")
	}

	// General design suggestions
	if analysis.WordCount > 100 {
		suggestions = append(suggestions, "Consider splitting content across multiple slides")
	}

	if analysis.HasNumbers {
		suggestions = append(suggestions, "Highlight important numbers with larger fonts or colors")
	}

	if analysis.HasDates {
		suggestions = append(suggestions, "Consider using a timeline layout for date-based content")
	}

	return suggestions
}

func generateEnhancedDesignSuggestions(analysis assets.ContentAnalysis, layout assets.SmartLayout, identity assets.DesignIdentity, theme assets.DesignTheme, typographyReport map[string]any) []string {
	suggestions := []string{}

	// Industry-specific suggestions
	suggestions = append(suggestions, fmt.Sprintf("Use %s visual style to match %s industry", identity.Style, identity.Industry))
	suggestions = append(suggestions, fmt.Sprintf("Target %s with %s tone", identity.Audience, identity.EmotionalTone))

	// Visual metaphor suggestions
	suggestions = append(suggestions, fmt.Sprintf("Incorporate %s as visual metaphors", identity.VisualMetaphor))

	// Color scheme suggestions
	suggestions = append(suggestions, fmt.Sprintf("Apply %s color scheme", identity.ColorPreference))

	// Typography suggestions from report
	if recs, ok := typographyReport["recommendations"].([]string); ok {
		suggestions = append(suggestions, recs...)
	}

	// Theme-specific suggestions
	suggestions = append(suggestions, fmt.Sprintf("Use %s background patterns for %s theme", theme.BackgroundDesign.Type.String(), theme.Name))

	// Content-based suggestions
	switch analysis.ContentType {
	case assets.ContentDataDriven:
		suggestions = append(suggestions, "Consider adding data visualization elements")
	case assets.ContentListItems:
		suggestions = append(suggestions, "Use consistent bullet styling and spacing")
	case assets.ContentComparison:
		suggestions = append(suggestions, "Implement side-by-side layout with contrasting colors")
	}

	return suggestions
}

func generateIndustryElements(industry string) []assets.VisualElement {
	visualRenderer := assets.NewSmartVisualRenderer()
	return visualRenderer.GenerateIndustrySpecificElements(industry)
}