package assets

import (
	"fmt"
)

type TypographyConfig struct {
	FontName string  `json:"font_name"`
	FontSize int     `json:"font_size"`
	Bold     bool    `json:"bold"`
	Color    string  `json:"color"`
}

type StyleProperties struct {
	BackgroundType   string  `json:"background_type"`
	AccentShapes     bool    `json:"accent_shapes"`
	HeaderStyle      string  `json:"header_style"`
	LayoutSpacing    string  `json:"layout_spacing"`
	BorderRadius     float64 `json:"border_radius"`
	ShadowIntensity  float64 `json:"shadow_intensity"`
}

type DesignTheme struct {
	Name            string                     `json:"name"`
	Description     string                     `json:"description"`
	Colors          map[string]string          `json:"colors"`
	Typography      map[string]TypographyConfig `json:"typography"`
	StyleProperties StyleProperties            `json:"style_properties"`
	BackgroundDesign BackgroundDesign          `json:"background_design"`
	Watermark       map[string]any            `json:"watermark"`
	FrameElements   []DecorativeElement       `json:"frame_elements"`
}

type DesignTemplateLibrary struct{}

func NewDesignTemplateLibrary() *DesignTemplateLibrary {
	return &DesignTemplateLibrary{}
}

func (d *DesignTemplateLibrary) GetCorporateTheme() DesignTheme {
	return DesignTheme{
		Name:        "Corporate Professional",
		Description: "Conservative, professional design suitable for corporate and government presentations",
		Colors: map[string]string{
			"primary":    "#2E75B6", // Professional blue
			"secondary":  "#5A6C7D", // Muted blue-gray
			"background": "#FFFFFF", // Clean white
			"text":       "#2C3E50", // Dark blue-gray
			"accent":     "#3498DB", // Bright blue
			"light":      "#F8F9FA", // Light gray
		},
		Typography: map[string]TypographyConfig{
			"title_slide": {FontName: "Calibri", FontSize: 36, Bold: true, Color: "primary"},
			"slide_title": {FontName: "Calibri", FontSize: 24, Bold: true, Color: "primary"},
			"body_text":   {FontName: "Calibri", FontSize: 14, Bold: false, Color: "text"},
			"caption":     {FontName: "Calibri", FontSize: 11, Bold: false, Color: "secondary"},
		},
		StyleProperties: StyleProperties{
			BackgroundType:  "solid",
			AccentShapes:    true,
			HeaderStyle:     "minimal",
			LayoutSpacing:   "generous",
			BorderRadius:    2.0,
			ShadowIntensity: 0.1,
		},
		BackgroundDesign: BackgroundDesign{
			Type:           BackgroundCorporateBars,
			PrimaryColor:   "#FFFFFF",
			SecondaryColor: "#2E75B6",
			PatternOpacity: 0.08,
		},
		FrameElements: []DecorativeElement{
			{
				ShapeType: "rectangle",
				Position:  map[string]float64{"x": 0, "y": 0, "width": 1, "height": 0.02},
				Color:     "#2E75B6",
				Opacity:   1.0,
			},
		},
	}
}

func (d *DesignTemplateLibrary) GetModernTechTheme() DesignTheme {
	return DesignTheme{
		Name:        "Modern Tech",
		Description: "Contemporary design with gradients, suitable for tech companies and startups",
		Colors: map[string]string{
			"primary":    "#667EEA", // Gradient purple
			"secondary":  "#764BA2", // Deep purple
			"background": "#E8EDFF", // Light tech blue background
			"text":       "#1A202C", // Near black
			"accent":     "#4FD1C7", // Teal accent
			"light":      "#F0F4FF", // Very light tech blue
		},
		Typography: map[string]TypographyConfig{
			"title_slide": {FontName: "Segoe UI", FontSize: 40, Bold: true, Color: "primary"},
			"slide_title": {FontName: "Segoe UI", FontSize: 28, Bold: true, Color: "primary"},
			"body_text":   {FontName: "Segoe UI", FontSize: 16, Bold: false, Color: "text"},
			"caption":     {FontName: "Segoe UI", FontSize: 12, Bold: false, Color: "secondary"},
		},
		StyleProperties: StyleProperties{
			BackgroundType:  "gradient",
			AccentShapes:    true,
			HeaderStyle:     "modern",
			LayoutSpacing:   "tight",
			BorderRadius:    8.0,
			ShadowIntensity: 0.2,
		},
		BackgroundDesign: BackgroundDesign{
			Type:           BackgroundTechCircuit,
			PrimaryColor:   "#F7FAFC",
			SecondaryColor: "#4FD1C7",
			PatternOpacity: 0.1,
		},
		FrameElements: []DecorativeElement{
			{
				ShapeType: "circle",
				Position:  map[string]float64{"x": 0.9, "y": 0.1, "width": 0.08, "height": 0.08},
				Color:     "#4FD1C7",
				Opacity:   0.3,
			},
		},
	}
}

func (d *DesignTemplateLibrary) GetHealthcareTheme() DesignTheme {
	return DesignTheme{
		Name:        "Healthcare Professional",
		Description: "Clean, calming design optimized for medical and healthcare presentations",
		Colors: map[string]string{
			"primary":    "#2D7DB3", // Medical blue
			"secondary":  "#38B2AC", // Teal
			"background": "#E8F8F5", // Light healthcare green background
			"text":       "#2D3748", // Dark gray
			"accent":     "#48BB78", // Healing green
			"light":      "#F0FFF8", // Very light mint
		},
		Typography: map[string]TypographyConfig{
			"title_slide": {FontName: "Arial", FontSize: 32, Bold: true, Color: "primary"},
			"slide_title": {FontName: "Arial", FontSize: 24, Bold: true, Color: "primary"},
			"body_text":   {FontName: "Arial", FontSize: 14, Bold: false, Color: "text"},
			"caption":     {FontName: "Arial", FontSize: 11, Bold: false, Color: "secondary"},
		},
		StyleProperties: StyleProperties{
			BackgroundType:  "solid",
			AccentShapes:    false,
			HeaderStyle:     "clean",
			LayoutSpacing:   "balanced",
			BorderRadius:    4.0,
			ShadowIntensity: 0.05,
		},
		BackgroundDesign: BackgroundDesign{
			Type:           BackgroundMedicalCurves,
			PrimaryColor:   "#FEFEFE",
			SecondaryColor: "#38B2AC",
			PatternOpacity: 0.12,
		},
		FrameElements: []DecorativeElement{
			{
				ShapeType: "line",
				Position:  map[string]float64{"x": 0.05, "y": 0.15, "width": 0.03, "height": 0.7},
				Color:     "#48BB78",
				Opacity:   0.6,
			},
		},
	}
}

func (d *DesignTemplateLibrary) GetFinancialTheme() DesignTheme {
	return DesignTheme{
		Name:        "Financial Services",
		Description: "Sophisticated design emphasizing trust, growth, and financial stability",
		Colors: map[string]string{
			"primary":    "#1B5E20", // Deep green
			"secondary":  "#2E7D32", // Forest green
			"background": "#E8F5E8", // Light financial green background
			"text":       "#1B5E20", // Dark green
			"accent":     "#FFB300", // Gold accent
			"light":      "#F0F8F0", // Very light green
		},
		Typography: map[string]TypographyConfig{
			"title_slide": {FontName: "Times New Roman", FontSize: 36, Bold: true, Color: "primary"},
			"slide_title": {FontName: "Times New Roman", FontSize: 26, Bold: true, Color: "primary"},
			"body_text":   {FontName: "Times New Roman", FontSize: 14, Bold: false, Color: "text"},
			"caption":     {FontName: "Times New Roman", FontSize: 12, Bold: false, Color: "secondary"},
		},
		StyleProperties: StyleProperties{
			BackgroundType:  "subtle_texture",
			AccentShapes:    true,
			HeaderStyle:     "elegant",
			LayoutSpacing:   "generous",
			BorderRadius:    3.0,
			ShadowIntensity: 0.15,
		},
		BackgroundDesign: BackgroundDesign{
			Type:           BackgroundDiagonalLines,
			PrimaryColor:   "#FFFDE7",
			SecondaryColor: "#2E7D32",
			PatternOpacity: 0.1,
		},
		FrameElements: []DecorativeElement{
			{
				ShapeType: "rectangle",
				Position:  map[string]float64{"x": 0.8, "y": 0.8, "width": 0.15, "height": 0.15},
				Color:     "#FFB300",
				Opacity:   0.2,
			},
		},
	}
}

func (d *DesignTemplateLibrary) GetSecurityTheme() DesignTheme {
	return DesignTheme{
		Name:        "Cybersecurity",
		Description: "Strong, secure design emphasizing protection and reliability",
		Colors: map[string]string{
			"primary":    "#C53030", // Security red
			"secondary":  "#2D3748", // Dark gray
			"background": "#1A202C", // Dark background
			"text":       "#F7FAFC", // Light text
			"accent":     "#E53E3E", // Bright red
			"light":      "#4A5568", // Medium gray
		},
		Typography: map[string]TypographyConfig{
			"title_slide": {FontName: "Arial", FontSize: 34, Bold: true, Color: "primary"},
			"slide_title": {FontName: "Arial", FontSize: 26, Bold: true, Color: "accent"},
			"body_text":   {FontName: "Arial", FontSize: 14, Bold: false, Color: "text"},
			"caption":     {FontName: "Arial", FontSize: 11, Bold: false, Color: "light"},
		},
		StyleProperties: StyleProperties{
			BackgroundType:  "dark_gradient",
			AccentShapes:    true,
			HeaderStyle:     "strong",
			LayoutSpacing:   "tight",
			BorderRadius:    0.0,
			ShadowIntensity: 0.3,
		},
		BackgroundDesign: BackgroundDesign{
			Type:           BackgroundDiagonalLines,
			PrimaryColor:   "#1A202C",
			SecondaryColor: "#C53030",
			PatternOpacity: 0.15,
		},
		FrameElements: []DecorativeElement{
			{
				ShapeType: "polygon",
				Position:  map[string]float64{"x": 0.02, "y": 0.02, "width": 0.05, "height": 0.05},
				Color:     "#E53E3E",
				Opacity:   0.8,
			},
		},
	}
}

func (d *DesignTemplateLibrary) GetEducationTheme() DesignTheme {
	return DesignTheme{
		Name:        "Educational",
		Description: "Friendly, accessible design perfect for learning and training materials",
		Colors: map[string]string{
			"primary":    "#2B6CB0", // Education blue
			"secondary":  "#ED8936", // Warm orange
			"background": "#FFFBF0", // Warm white
			"text":       "#2D3748", // Dark gray
			"accent":     "#38A169", // Growth green
			"light":      "#FFF5F5", // Light peach
		},
		Typography: map[string]TypographyConfig{
			"title_slide": {FontName: "Verdana", FontSize: 32, Bold: true, Color: "primary"},
			"slide_title": {FontName: "Verdana", FontSize: 22, Bold: true, Color: "primary"},
			"body_text":   {FontName: "Verdana", FontSize: 14, Bold: false, Color: "text"},
			"caption":     {FontName: "Verdana", FontSize: 11, Bold: false, Color: "secondary"},
		},
		StyleProperties: StyleProperties{
			BackgroundType:  "warm_gradient",
			AccentShapes:    true,
			HeaderStyle:     "friendly",
			LayoutSpacing:   "comfortable",
			BorderRadius:    6.0,
			ShadowIntensity: 0.1,
		},
		BackgroundDesign: BackgroundDesign{
			Type:           BackgroundSolid,
			PrimaryColor:   "#FFFBF0",
			SecondaryColor: "#ED8936",
			PatternOpacity: 0.05,
		},
		FrameElements: []DecorativeElement{
			{
				ShapeType: "circle",
				Position:  map[string]float64{"x": 0.05, "y": 0.9, "width": 0.06, "height": 0.06},
				Color:     "#38A169",
				Opacity:   0.4,
			},
		},
	}
}

func (d *DesignTemplateLibrary) GetThemeForAnalysis(analysis *DesignIdentity) DesignTheme {
	switch analysis.Industry {
	case "Technology/Software":
		return d.GetModernTechTheme()
	case "Corporate/Consulting":
		return d.GetCorporateTheme()
	case "Healthcare/Medical":
		return d.GetHealthcareTheme()
	case "Financial Services/Banking":
		return d.GetFinancialTheme()
	case "Cybersecurity/Compliance":
		return d.GetSecurityTheme()
	case "Education/Training":
		return d.GetEducationTheme()
	default:
		return d.GetCorporateTheme() // Default to corporate
	}
}

func (d *DesignTemplateLibrary) GetAllThemes() []DesignTheme {
	return []DesignTheme{
		d.GetCorporateTheme(),
		d.GetModernTechTheme(),
		d.GetHealthcareTheme(),
		d.GetFinancialTheme(),
		d.GetSecurityTheme(),
		d.GetEducationTheme(),
	}
}

func (d *DesignTemplateLibrary) GetThemeByName(name string) (*DesignTheme, error) {
	themes := map[string]func() DesignTheme{
		"Corporate Professional": d.GetCorporateTheme,
		"Modern Tech":           d.GetModernTechTheme,
		"Healthcare Professional": d.GetHealthcareTheme,
		"Financial Services":    d.GetFinancialTheme,
		"Cybersecurity":         d.GetSecurityTheme,
		"Educational":           d.GetEducationTheme,
	}

	if themeFunc, exists := themes[name]; exists {
		theme := themeFunc()
		return &theme, nil
	}

	return nil, fmt.Errorf("theme not found: %s", name)
}