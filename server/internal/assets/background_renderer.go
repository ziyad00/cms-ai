package assets

import (
	"image/color"
	"strconv"

	"baliance.com/gooxml/presentation"
)

type BackgroundType int

const (
	BackgroundSolid BackgroundType = iota
	BackgroundGradient
	BackgroundDiagonalLines
	BackgroundHexagonGrid
	BackgroundMedicalCurves
	BackgroundTechCircuit
	BackgroundCorporateBars
	BackgroundDarkGradient
)

func (bt BackgroundType) String() string {
	switch bt {
	case BackgroundSolid:
		return "solid"
	case BackgroundGradient:
		return "gradient"
	case BackgroundDiagonalLines:
		return "diagonal lines"
	case BackgroundHexagonGrid:
		return "hexagon grid"
	case BackgroundMedicalCurves:
		return "medical curves"
	case BackgroundTechCircuit:
		return "tech circuit"
	case BackgroundCorporateBars:
		return "corporate bars"
	case BackgroundDarkGradient:
		return "dark gradient"
	default:
		return "solid"
	}
}

type DecorativeElement struct {
	ShapeType    string             `json:"shape_type"`    // rectangle, circle, line, polygon
	Position     map[string]float64 `json:"position"`      // x, y, width, height
	Color        string             `json:"color"`
	Opacity      float64            `json:"opacity"`
	Rotation     float64            `json:"rotation"`
	PatternData  map[string]any     `json:"pattern_data"`
}

type BackgroundDesign struct {
	Type               BackgroundType      `json:"type"`
	PrimaryColor       string              `json:"primary_color"`
	SecondaryColor     string              `json:"secondary_color"`
	PatternOpacity     float64             `json:"pattern_opacity"`
	DecorativeElements []DecorativeElement `json:"decorative_elements"`
}

type AdvancedBackgroundRenderer struct{}

func NewAdvancedBackgroundRenderer() *AdvancedBackgroundRenderer {
	return &AdvancedBackgroundRenderer{}
}

func (r *AdvancedBackgroundRenderer) ApplyBackgroundDesign(slide presentation.Slide, design BackgroundDesign) error {
	// Apply base background
	r.applyBaseBackground(slide, design)

	// Add pattern overlay
	r.addPatternOverlay(slide, design)

	// Add decorative elements
	r.addDecorativeElements(slide, design.DecorativeElements)

	return nil
}

func (r *AdvancedBackgroundRenderer) applyBaseBackground(slide presentation.Slide, design BackgroundDesign) {
	switch design.Type {
	case BackgroundSolid, BackgroundMedicalCurves:
		// Apply solid background
		r.applySolidBackground(slide, design.PrimaryColor)
	case BackgroundGradient, BackgroundDarkGradient:
		// Apply gradient background
		r.applyGradientBackground(slide, design.PrimaryColor, design.SecondaryColor)
	default:
		// Default to solid
		r.applySolidBackground(slide, design.PrimaryColor)
	}
}

func (r *AdvancedBackgroundRenderer) applySolidBackground(slide presentation.Slide, colorHex string) {
	// Note: gooxml has very limited slide background API
	// Background colors would need to be set at the slide master level
	// or through direct XML manipulation which is beyond current scope
}

func (r *AdvancedBackgroundRenderer) applyGradientBackground(slide presentation.Slide, primaryHex, secondaryHex string) {
	// Note: gooxml gradient support is limited
	// This would require more complex implementation with shapes
}

func (r *AdvancedBackgroundRenderer) addPatternOverlay(slide presentation.Slide, design BackgroundDesign) {
	switch design.Type {
	case BackgroundDiagonalLines:
		r.createDiagonalLinesPattern(slide, design)
	case BackgroundHexagonGrid:
		r.createHexagonGridPattern(slide, design)
	case BackgroundMedicalCurves:
		r.createMedicalCurvesPattern(slide, design)
	case BackgroundTechCircuit:
		r.createTechCircuitPattern(slide, design)
	case BackgroundCorporateBars:
		r.createCorporateBarsPattern(slide, design)
	}
}

func (r *AdvancedBackgroundRenderer) createDiagonalLinesPattern(slide presentation.Slide, design BackgroundDesign) {
	// Simplified implementation - gooxml has limited shape API
	// This would be implemented with proper shape creation in a full implementation
}

func (r *AdvancedBackgroundRenderer) createHexagonGridPattern(slide presentation.Slide, design BackgroundDesign) {
	// Simplified implementation - gooxml has limited shape API
}

func (r *AdvancedBackgroundRenderer) createMedicalCurvesPattern(slide presentation.Slide, design BackgroundDesign) {
	// Simplified implementation - gooxml has limited shape API
}

func (r *AdvancedBackgroundRenderer) createTechCircuitPattern(slide presentation.Slide, design BackgroundDesign) {
	// Simplified implementation - gooxml has limited shape API
}

func (r *AdvancedBackgroundRenderer) createCorporateBarsPattern(slide presentation.Slide, design BackgroundDesign) {
	// Simplified implementation - gooxml has limited shape API
}

func (r *AdvancedBackgroundRenderer) addDecorativeElements(slide presentation.Slide, elements []DecorativeElement) {
	for _, element := range elements {
		r.addDecorativeElement(slide, element)
	}
}

func (r *AdvancedBackgroundRenderer) addDecorativeElement(slide presentation.Slide, element DecorativeElement) {
	// Simplified implementation - gooxml has limited shape API
}

func (r *AdvancedBackgroundRenderer) setShapeColor(shape any, colorHex string, opacity float64) {
	// Simplified implementation - gooxml has limited shape API
}

func (r *AdvancedBackgroundRenderer) hexToRGB(hex string) color.RGBA {
	// Simple hex to RGB conversion
	if len(hex) != 6 {
		return color.RGBA{224, 224, 224, 255} // Light gray default
	}

	// Parse RGB components
	rStr := hex[0:2]
	gStr := hex[2:4]
	bStr := hex[4:6]

	red, _ := strconv.ParseUint(rStr, 16, 8)
	green, _ := strconv.ParseUint(gStr, 16, 8)
	blue, _ := strconv.ParseUint(bStr, 16, 8)

	return color.RGBA{uint8(red), uint8(green), uint8(blue), 255}
}

func (r *AdvancedBackgroundRenderer) GetBackgroundDesignForTheme(theme ThemeType) BackgroundDesign {
	switch theme {
	case ThemeTechnology:
		return BackgroundDesign{
			Type:           BackgroundTechCircuit,
			PrimaryColor:   "#F7FAFC",
			SecondaryColor: "#4FD1C7",
			PatternOpacity: 0.1,
		}
	case ThemeBusiness:
		return BackgroundDesign{
			Type:           BackgroundCorporateBars,
			PrimaryColor:   "#FFFFFF",
			SecondaryColor: "#2E75B6",
			PatternOpacity: 0.08,
		}
	case ThemeSecurity:
		return BackgroundDesign{
			Type:           BackgroundDiagonalLines,
			PrimaryColor:   "#1A202C",
			SecondaryColor: "#E53E3E",
			PatternOpacity: 0.15,
		}
	case ThemeHealthcare:
		return BackgroundDesign{
			Type:           BackgroundMedicalCurves,
			PrimaryColor:   "#F7FAFC",
			SecondaryColor: "#38B2AC",
			PatternOpacity: 0.12,
		}
	case ThemeFinance:
		return BackgroundDesign{
			Type:           BackgroundDiagonalLines,
			PrimaryColor:   "#FFFAF0",
			SecondaryColor: "#D69E2E",
			PatternOpacity: 0.1,
		}
	default:
		return BackgroundDesign{
			Type:           BackgroundSolid,
			PrimaryColor:   "#FFFFFF",
			SecondaryColor: "#E2E8F0",
			PatternOpacity: 0.05,
		}
	}
}