package assets

import (
	"baliance.com/gooxml/color"
	"baliance.com/gooxml/measurement"
	"baliance.com/gooxml/presentation"
)

// VisualEnhancementRenderer creates visual elements using available gooxml features
type VisualEnhancementRenderer struct{}

func NewVisualEnhancementRenderer() *VisualEnhancementRenderer {
	return &VisualEnhancementRenderer{}
}

// ApplySlideEnhancements adds visual elements that gooxml supports
func (v *VisualEnhancementRenderer) ApplySlideEnhancements(slide presentation.Slide, theme DesignTheme, slideType string) {
	// Create colored header bar using a text box with background
	v.addHeaderBar(slide, theme)

	// Add accent elements using text boxes with colored backgrounds
	if theme.StyleProperties.AccentShapes {
		v.addAccentElements(slide, theme, slideType)
	}
}

func (v *VisualEnhancementRenderer) addHeaderBar(slide presentation.Slide, theme DesignTheme) {
	// Create a thin header bar using a text box with colored fill
	headerBox := slide.AddTextBox()
	props := headerBox.Properties()

	// Position at top of slide, full width, thin height
	props.SetPosition(0, 0)
	props.SetSize(
		measurement.Distance(10)*measurement.Inch,
		measurement.Distance(0.2)*measurement.Inch,
	)

	// Set background color to primary theme color
	primaryColor := color.FromHex(theme.Colors["primary"])
	props.SetSolidFill(primaryColor)
	props.LineProperties().SetNoFill()

	// Add empty paragraph to make the box visible
	para := headerBox.AddParagraph()
	run := para.AddRun()
	run.SetText(" ") // Space to ensure visibility
	run.Properties().SetSize(measurement.Distance(1) * measurement.Point)
}

func (v *VisualEnhancementRenderer) addAccentElements(slide presentation.Slide, theme DesignTheme, slideType string) {
	switch slideType {
	case "title":
		v.addTitleAccent(slide, theme)
	case "content":
		v.addContentAccent(slide, theme)
	case "conclusion":
		v.addConclusionAccent(slide, theme)
	}
}

func (v *VisualEnhancementRenderer) addTitleAccent(slide presentation.Slide, theme DesignTheme) {
	// Add a colored accent box on the right side
	accentBox := slide.AddTextBox()
	props := accentBox.Properties()

	// Position on right side
	props.SetPosition(
		measurement.Distance(9)*measurement.Inch,
		measurement.Distance(1)*measurement.Inch,
	)
	props.SetSize(
		measurement.Distance(0.8)*measurement.Inch,
		measurement.Distance(5)*measurement.Inch,
	)

	// Set accent color
	accentColor := color.FromHex(theme.Colors["accent"])
	props.SetSolidFill(accentColor)
	props.LineProperties().SetNoFill()

	// Add minimal content
	para := accentBox.AddParagraph()
	run := para.AddRun()
	run.SetText(" ")
	run.Properties().SetSize(measurement.Distance(1) * measurement.Point)
}

func (v *VisualEnhancementRenderer) addContentAccent(slide presentation.Slide, theme DesignTheme) {
	// Add a small colored accent box in the corner
	accentBox := slide.AddTextBox()
	props := accentBox.Properties()

	// Position in bottom right corner
	props.SetPosition(
		measurement.Distance(8.5)*measurement.Inch,
		measurement.Distance(6.8)*measurement.Inch,
	)
	props.SetSize(
		measurement.Distance(1.2)*measurement.Inch,
		measurement.Distance(0.5)*measurement.Inch,
	)

	// Set secondary color with some transparency
	secondaryColor := color.FromHex(theme.Colors["secondary"])
	props.SetSolidFill(secondaryColor)
	props.LineProperties().SetNoFill()

	// Add minimal content
	para := accentBox.AddParagraph()
	run := para.AddRun()
	run.SetText(" ")
	run.Properties().SetSize(measurement.Distance(1) * measurement.Point)
}

func (v *VisualEnhancementRenderer) addConclusionAccent(slide presentation.Slide, theme DesignTheme) {
	// Add decorative elements for conclusion slides
	v.addContentAccent(slide, theme)

	// Add bottom border
	bottomBox := slide.AddTextBox()
	props := bottomBox.Properties()

	// Position at bottom
	props.SetPosition(
		measurement.Distance(0.5)*measurement.Inch,
		measurement.Distance(7.2)*measurement.Inch,
	)
	props.SetSize(
		measurement.Distance(9)*measurement.Inch,
		measurement.Distance(0.1)*measurement.Inch,
	)

	// Set primary color
	primaryColor := color.FromHex(theme.Colors["primary"])
	props.SetSolidFill(primaryColor)
	props.LineProperties().SetNoFill()

	// Add minimal content
	para := bottomBox.AddParagraph()
	run := para.AddRun()
	run.SetText(" ")
	run.Properties().SetSize(measurement.Distance(1) * measurement.Point)
}

// AddSlideBackground attempts to set slide background through available methods
func (v *VisualEnhancementRenderer) AddSlideBackground(slide presentation.Slide, bgColor string) {
	// gooxml doesn't support slide.background.fill.solid() like python-pptx
	// Use a full-slide text box as background workaround
	v.createBackgroundRectangle(slide, bgColor)
}

func (v *VisualEnhancementRenderer) createBackgroundRectangle(slide presentation.Slide, bgColor string) {
	// Create a background using a large colored text box positioned first
	// This acts as a slide background since gooxml doesn't support slide.background

	// Note: Since we can't add shapes, we'll make a very large text box with background
	// that covers the entire slide and position content on top
	bgBox := slide.AddTextBox()
	props := bgBox.Properties()

	// Position at slide origin and cover entire slide area
	props.SetPosition(0, 0)
	props.SetSize(
		measurement.Distance(10)*measurement.Inch,
		measurement.Distance(7.5)*measurement.Inch,
	)

	// Set the background color
	bgColorObj := color.FromHex(bgColor)
	props.SetSolidFill(bgColorObj)
	props.LineProperties().SetNoFill()

	// Add empty content to make the background visible
	para := bgBox.AddParagraph()

	// The issue might be that we need actual text content
	// Let's add a very small, transparent text
	run := para.AddRun()
	run.SetText("\u00A0") // Non-breaking space
	runProps := run.Properties()
	runProps.SetSize(measurement.Distance(1) * measurement.Point)

	// Try to make text color same as background so it's invisible
	textColor := color.FromHex(bgColor)
	runProps.SetSolidFill(textColor)
}