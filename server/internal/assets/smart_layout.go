package assets

import ()

type LayoutType int

const (
	LayoutTitle LayoutType = iota
	LayoutContent
	LayoutComparison
	LayoutTimeline
	LayoutMetrics
	LayoutQuote
	LayoutDataVisualization
)

type SmartLayout struct {
	Type           LayoutType
	Title          PlaceholderConfig
	Content        []PlaceholderConfig
	Background     BackgroundConfig
	Typography     SmartTypographyConfig
	ColorScheme    ColorScheme
}

type PlaceholderConfig struct {
	ID       string
	Type     string
	X        float64
	Y        float64
	W        float64
	H        float64
	FontSize int
	Bold     bool
	Align    string
}

type BackgroundConfig struct {
	Type      string // solid, gradient, pattern
	Primary   string
	Secondary string
	Opacity   float64
}

type SmartTypographyConfig struct {
	TitleFont   string
	ContentFont string
	TitleSize   int
	ContentSize int
}

type ColorScheme struct {
	Primary     string
	Secondary   string
	Background  string
	Text        string
	Accent      string
}

type SmartLayoutGenerator struct {
	analyzer *SmartContentAnalyzer
}

func NewSmartLayoutGenerator() *SmartLayoutGenerator {
	return &SmartLayoutGenerator{
		analyzer: NewSmartContentAnalyzer(),
	}
}

func (g *SmartLayoutGenerator) GenerateLayout(title, content string, slideNumber, totalSlides int) SmartLayout {
	analysis := g.analyzer.AnalyzeContent(content)
	layoutType := g.mapContentTypeToLayout(analysis.ContentType)

	colorScheme := g.selectColorScheme(analysis)
	typography := g.selectTypography(analysis)

	switch layoutType {
	case LayoutTitle:
		return g.generateTitleLayout(title, content, colorScheme, typography)
	case LayoutComparison:
		return g.generateComparisonLayout(title, content, colorScheme, typography)
	case LayoutTimeline:
		return g.generateTimelineLayout(title, content, colorScheme, typography)
	case LayoutMetrics:
		return g.generateMetricsLayout(title, content, colorScheme, typography)
	case LayoutQuote:
		return g.generateQuoteLayout(title, content, colorScheme, typography)
	case LayoutDataVisualization:
		return g.generateDataVizLayout(title, content, colorScheme, typography)
	default:
		return g.generateContentLayout(title, content, colorScheme, typography, analysis)
	}
}

func (g *SmartLayoutGenerator) mapContentTypeToLayout(contentType ContentType) LayoutType {
	switch contentType {
	case ContentComparison:
		return LayoutComparison
	case ContentTimeline:
		return LayoutTimeline
	case ContentDataDriven:
		return LayoutDataVisualization
	case ContentQuote:
		return LayoutQuote
	default:
		return LayoutContent
	}
}

func (g *SmartLayoutGenerator) selectColorScheme(analysis ContentAnalysis) ColorScheme {
	// Select colors based on content sentiment and concepts
	switch analysis.Sentiment {
	case "positive":
		return ColorScheme{
			Primary:    "#2E8B57", // Sea Green
			Secondary:  "#90EE90", // Light Green
			Background: "#F0FFF0", // Honeydew
			Text:       "#2F4F4F", // Dark Slate Gray
			Accent:     "#FFD700", // Gold
		}
	case "urgent":
		return ColorScheme{
			Primary:    "#DC143C", // Crimson
			Secondary:  "#FF6347", // Tomato
			Background: "#FFF8DC", // Cornsilk
			Text:       "#8B0000", // Dark Red
			Accent:     "#FF4500", // Orange Red
		}
	case "negative":
		return ColorScheme{
			Primary:    "#4682B4", // Steel Blue
			Secondary:  "#87CEEB", // Sky Blue
			Background: "#F0F8FF", // Alice Blue
			Text:       "#2F4F4F", // Dark Slate Gray
			Accent:     "#FF6347", // Tomato
		}
	default: // neutral
		return ColorScheme{
			Primary:    "#2C3E50", // Dark Blue Gray
			Secondary:  "#3498DB", // Dodger Blue
			Background: "#FFFFFF", // White
			Text:       "#2C3E50", // Dark Blue Gray
			Accent:     "#E74C3C", // Red
		}
	}
}

func (g *SmartLayoutGenerator) selectTypography(analysis ContentAnalysis) SmartTypographyConfig {
	baseSize := 16
	titleSize := 28

	// Adjust based on complexity
	switch analysis.Complexity {
	case "simple":
		titleSize = 32
		baseSize = 18
	case "complex":
		titleSize = 24
		baseSize = 14
	}

	return SmartTypographyConfig{
		TitleFont:   "Arial",
		ContentFont: "Arial",
		TitleSize:   titleSize,
		ContentSize: baseSize,
	}
}

func (g *SmartLayoutGenerator) generateTitleLayout(title, content string, colorScheme ColorScheme, typography SmartTypographyConfig) SmartLayout {
	return SmartLayout{
		Type: LayoutTitle,
		Title: PlaceholderConfig{
			ID:       "title",
			Type:     "text",
			X:        0.1,
			Y:        0.3,
			W:        0.8,
			H:        0.2,
			FontSize: typography.TitleSize + 8,
			Bold:     true,
			Align:    "center",
		},
		Content: []PlaceholderConfig{
			{
				ID:       "subtitle",
				Type:     "text",
				X:        0.1,
				Y:        0.55,
				W:        0.8,
				H:        0.15,
				FontSize: typography.ContentSize + 4,
				Bold:     false,
				Align:    "center",
			},
		},
		ColorScheme: colorScheme,
		Background: BackgroundConfig{
			Type:    "gradient",
			Primary: colorScheme.Background,
			Secondary: colorScheme.Secondary,
			Opacity: 0.8,
		},
	}
}

func (g *SmartLayoutGenerator) generateContentLayout(title, content string, colorScheme ColorScheme, typography SmartTypographyConfig, analysis ContentAnalysis) SmartLayout {
	// Dynamic sizing based on content analysis
	titleHeight := 0.15
	contentY := titleHeight + 0.05
	contentHeight := 0.75 - contentY

	// Adjust for visual weight
	if analysis.VisualWeight > 0.7 {
		titleHeight = 0.12
		contentY = titleHeight + 0.03
		contentHeight = 0.8 - contentY
	}

	return SmartLayout{
		Type: LayoutContent,
		Title: PlaceholderConfig{
			ID:       "title",
			Type:     "text",
			X:        0.05,
			Y:        0.05,
			W:        0.9,
			H:        titleHeight,
			FontSize: typography.TitleSize,
			Bold:     true,
			Align:    "left",
		},
		Content: []PlaceholderConfig{
			{
				ID:       "content",
				Type:     "text",
				X:        0.05,
				Y:        contentY,
				W:        0.9,
				H:        contentHeight,
				FontSize: typography.ContentSize,
				Bold:     false,
				Align:    "left",
			},
		},
		ColorScheme: colorScheme,
		Background: BackgroundConfig{
			Type:    "solid",
			Primary: colorScheme.Background,
			Opacity: 1.0,
		},
	}
}

func (g *SmartLayoutGenerator) generateComparisonLayout(title, content string, colorScheme ColorScheme, typography SmartTypographyConfig) SmartLayout {
	return SmartLayout{
		Type: LayoutComparison,
		Title: PlaceholderConfig{
			ID:       "title",
			Type:     "text",
			X:        0.05,
			Y:        0.05,
			W:        0.9,
			H:        0.15,
			FontSize: typography.TitleSize,
			Bold:     true,
			Align:    "center",
		},
		Content: []PlaceholderConfig{
			{
				ID:       "left_column",
				Type:     "text",
				X:        0.05,
				Y:        0.25,
				W:        0.4,
				H:        0.65,
				FontSize: typography.ContentSize,
				Bold:     false,
				Align:    "left",
			},
			{
				ID:       "right_column",
				Type:     "text",
				X:        0.55,
				Y:        0.25,
				W:        0.4,
				H:        0.65,
				FontSize: typography.ContentSize,
				Bold:     false,
				Align:    "left",
			},
		},
		ColorScheme: colorScheme,
		Background: BackgroundConfig{
			Type:    "solid",
			Primary: colorScheme.Background,
			Opacity: 1.0,
		},
	}
}

func (g *SmartLayoutGenerator) generateMetricsLayout(title, content string, colorScheme ColorScheme, typography SmartTypographyConfig) SmartLayout {
	return SmartLayout{
		Type: LayoutMetrics,
		Title: PlaceholderConfig{
			ID:       "title",
			Type:     "text",
			X:        0.05,
			Y:        0.05,
			W:        0.9,
			H:        0.15,
			FontSize: typography.TitleSize,
			Bold:     true,
			Align:    "center",
		},
		Content: []PlaceholderConfig{
			{
				ID:       "metric_1",
				Type:     "text",
				X:        0.1,
				Y:        0.25,
				W:        0.35,
				H:        0.3,
				FontSize: typography.ContentSize + 8,
				Bold:     true,
				Align:    "center",
			},
			{
				ID:       "metric_2",
				Type:     "text",
				X:        0.55,
				Y:        0.25,
				W:        0.35,
				H:        0.3,
				FontSize: typography.ContentSize + 8,
				Bold:     true,
				Align:    "center",
			},
			{
				ID:       "description",
				Type:     "text",
				X:        0.1,
				Y:        0.6,
				W:        0.8,
				H:        0.3,
				FontSize: typography.ContentSize,
				Bold:     false,
				Align:    "center",
			},
		},
		ColorScheme: colorScheme,
		Background: BackgroundConfig{
			Type:    "gradient",
			Primary: colorScheme.Background,
			Secondary: colorScheme.Secondary,
			Opacity: 0.9,
		},
	}
}

func (g *SmartLayoutGenerator) generateTimelineLayout(title, content string, colorScheme ColorScheme, typography SmartTypographyConfig) SmartLayout {
	return SmartLayout{
		Type: LayoutTimeline,
		Title: PlaceholderConfig{
			ID:       "title",
			Type:     "text",
			X:        0.05,
			Y:        0.05,
			W:        0.9,
			H:        0.15,
			FontSize: typography.TitleSize,
			Bold:     true,
			Align:    "left",
		},
		Content: []PlaceholderConfig{
			{
				ID:       "timeline",
				Type:     "text",
				X:        0.1,
				Y:        0.25,
				W:        0.8,
				H:        0.65,
				FontSize: typography.ContentSize,
				Bold:     false,
				Align:    "left",
			},
		},
		ColorScheme: colorScheme,
		Background: BackgroundConfig{
			Type:    "solid",
			Primary: colorScheme.Background,
			Opacity: 1.0,
		},
	}
}

func (g *SmartLayoutGenerator) generateQuoteLayout(title, content string, colorScheme ColorScheme, typography SmartTypographyConfig) SmartLayout {
	return SmartLayout{
		Type: LayoutQuote,
		Title: PlaceholderConfig{
			ID:       "title",
			Type:     "text",
			X:        0.05,
			Y:        0.05,
			W:        0.9,
			H:        0.15,
			FontSize: typography.TitleSize,
			Bold:     true,
			Align:    "center",
		},
		Content: []PlaceholderConfig{
			{
				ID:       "quote",
				Type:     "text",
				X:        0.15,
				Y:        0.3,
				W:        0.7,
				H:        0.4,
				FontSize: typography.ContentSize + 6,
				Bold:     false,
				Align:    "center",
			},
			{
				ID:       "attribution",
				Type:     "text",
				X:        0.15,
				Y:        0.75,
				W:        0.7,
				H:        0.1,
				FontSize: typography.ContentSize - 2,
				Bold:     false,
				Align:    "right",
			},
		},
		ColorScheme: colorScheme,
		Background: BackgroundConfig{
			Type:    "gradient",
			Primary: colorScheme.Background,
			Secondary: colorScheme.Accent,
			Opacity: 0.3,
		},
	}
}

func (g *SmartLayoutGenerator) generateDataVizLayout(title, content string, colorScheme ColorScheme, typography SmartTypographyConfig) SmartLayout {
	return SmartLayout{
		Type: LayoutDataVisualization,
		Title: PlaceholderConfig{
			ID:       "title",
			Type:     "text",
			X:        0.05,
			Y:        0.05,
			W:        0.9,
			H:        0.12,
			FontSize: typography.TitleSize,
			Bold:     true,
			Align:    "left",
		},
		Content: []PlaceholderConfig{
			{
				ID:       "chart_area",
				Type:     "text",
				X:        0.05,
				Y:        0.2,
				W:        0.6,
				H:        0.6,
				FontSize: typography.ContentSize,
				Bold:     false,
				Align:    "center",
			},
			{
				ID:       "insights",
				Type:     "text",
				X:        0.7,
				Y:        0.2,
				W:        0.25,
				H:        0.6,
				FontSize: typography.ContentSize - 2,
				Bold:     false,
				Align:    "left",
			},
		},
		ColorScheme: colorScheme,
		Background: BackgroundConfig{
			Type:    "solid",
			Primary: colorScheme.Background,
			Opacity: 1.0,
		},
	}
}