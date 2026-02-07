package assets

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"baliance.com/gooxml/measurement"
	"baliance.com/gooxml/presentation"
)

type Renderer interface {
	RenderPPTX(ctx context.Context, spec any, outPath string) error
	RenderPPTXBytes(ctx context.Context, spec any) ([]byte, error)
	GenerateSlideThumbnails(ctx context.Context, spec any) ([][]byte, error)
}

func specToJSONBytes(spec any) ([]byte, error) {
	// In our stores, spec_json may come back as []byte from the DB driver.
	// If we json.Marshal([]byte), it becomes a base64 string and breaks parsing.
	switch v := spec.(type) {
	case []byte:
		return v, nil
	case json.RawMessage:
		return []byte(v), nil
	default:
		return json.Marshal(spec)
	}
}

type PythonPPTXRenderer struct {
	PythonPath         string
	ScriptPath         string
	HuggingFaceAPIKey  string
}

// NewPythonPPTXRenderer creates a new Python renderer with smart path resolution
// Similar to GoPPTXRenderer, this provides Railway vs local development fallback
func NewPythonPPTXRenderer(huggingFaceAPIKey string) *PythonPPTXRenderer {
	// Primary path: Railway container environment
	railwayScriptPath := "/app/tools/renderer/render_pptx.py"

	// Check if Railway path exists
	if _, err := os.Stat(railwayScriptPath); err == nil {
		log.Printf("Python renderer: Using Railway container path: %s", railwayScriptPath)
		return &PythonPPTXRenderer{
			PythonPath:        "python3",
			ScriptPath:        railwayScriptPath,
			HuggingFaceAPIKey: huggingFaceAPIKey,
		}
	}

	// Fallback: Local development path resolution
	// Navigate up directories to find tools/renderer/render_pptx.py
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("Python renderer: Failed to get working directory: %v", err)
		// Use Railway path as fallback (will fail but provide clear error)
		return &PythonPPTXRenderer{
			PythonPath:        "python3",
			ScriptPath:        railwayScriptPath,
			HuggingFaceAPIKey: huggingFaceAPIKey,
		}
	}

	// Search for tools directory going up the directory tree
	searchDir := wd
	for i := 0; i < 5; i++ { // Limit search to prevent infinite loops
		localScriptPath := filepath.Join(searchDir, "tools", "renderer", "render_pptx.py")
		if _, err := os.Stat(localScriptPath); err == nil {
			log.Printf("Python renderer: Using local development path: %s", localScriptPath)
			return &PythonPPTXRenderer{
				PythonPath:        "python3",
				ScriptPath:        localScriptPath,
				HuggingFaceAPIKey: huggingFaceAPIKey,
			}
		}

		// Try web directory for Next.js deployment structure
		webScriptPath := filepath.Join(searchDir, "web", "tools", "renderer", "render_pptx.py")
		if _, err := os.Stat(webScriptPath); err == nil {
			log.Printf("Python renderer: Using web deployment path: %s", webScriptPath)
			return &PythonPPTXRenderer{
				PythonPath:        "python3",
				ScriptPath:        webScriptPath,
				HuggingFaceAPIKey: huggingFaceAPIKey,
			}
		}

		// Move up one directory
		parent := filepath.Dir(searchDir)
		if parent == searchDir {
			break // Reached filesystem root
		}
		searchDir = parent
	}

	log.Printf("Python renderer: No local script found, using Railway path as fallback: %s", railwayScriptPath)
	return &PythonPPTXRenderer{
		PythonPath:        "python3",
		ScriptPath:        railwayScriptPath,
		HuggingFaceAPIKey: huggingFaceAPIKey,
	}
}

func (r PythonPPTXRenderer) RenderPPTX(ctx context.Context, spec any, outPath string) error {
	return r.RenderPPTXWithCompany(ctx, spec, outPath, nil)
}

func (r PythonPPTXRenderer) RenderPPTXWithCompany(ctx context.Context, spec any, outPath string, company *CompanyContext) error {
	python := r.PythonPath
	if python == "" {
		python = "python3"
	}
	script := r.ScriptPath
	if script == "" {
		// Use Railway deployment path by default, fall back to local path
		script = "/app/tools/renderer/render_pptx.py"
		if _, err := os.Stat(script); err != nil {
			// Fall back to local development path (use absolute path)
			script = filepath.Join("server", "tools", "renderer", "render_pptx.py")

			// If still not found, try the current working directory's parent
			if _, err := os.Stat(script); err != nil {
				wd, _ := os.Getwd()
				// Navigate up to find the server directory
				for wd != "/" && wd != "" {
					testScript := filepath.Join(wd, "tools", "renderer", "render_pptx.py")
					if _, err := os.Stat(testScript); err == nil {
						script = testScript
						break
					}
					wd = filepath.Dir(wd)
				}
			}
		}
	}


	tmpDir := filepath.Dir(outPath)
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return err
	}

	// Create temporary spec file
	tmpSpec, err := os.CreateTemp(tmpDir, "spec-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmpSpec.Name())
	defer tmpSpec.Close()

	b, err := specToJSONBytes(spec)
	if err != nil {
		return err
	}
	if _, err := tmpSpec.Write(b); err != nil {
		return err
	}
	if err := tmpSpec.Close(); err != nil {
		return err
	}

	// Create command arguments
	args := []string{script, tmpSpec.Name(), outPath}

	// Add company info if provided
	var tmpCompany *os.File
	if company != nil {
		tmpCompany, err = os.CreateTemp(tmpDir, "company-*.json")
		if err != nil {
			return err
		}
		defer os.Remove(tmpCompany.Name())
		defer tmpCompany.Close()

		companyBytes, err := json.Marshal(company)
		if err != nil {
			return err
		}
		if _, err := tmpCompany.Write(companyBytes); err != nil {
			return err
		}
		if err := tmpCompany.Close(); err != nil {
			return err
		}

		args = append(args, "--company-info", tmpCompany.Name())
	}

	// Add Hugging Face API key if available
	if r.HuggingFaceAPIKey != "" {
		args = append(args, "--hf-api-key", r.HuggingFaceAPIKey)
	}

	// Check if script file exists
	if _, err := os.Stat(script); err != nil {
		return fmt.Errorf("script file not found: %v", err)
	}

	cmd := exec.CommandContext(ctx, python, args...)
	// Set working directory based on environment
	workDir := "/app" // Railway deployment root
	if strings.Contains(script, "tools/renderer/render_pptx.py") && !strings.HasPrefix(script, "/app/") {
		// Local development - use current directory
		workDir = ""
	}
	if workDir != "" {
		cmd.Dir = workDir
	}
	cmd.Env = append(os.Environ(),
		"PYTHONUNBUFFERED=1",
		"HUGGING_FACE_API_KEY="+r.HuggingFaceAPIKey,
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		stderrStr := stderr.String()
		if stderrStr != "" {
			return fmt.Errorf("python renderer failed: %s", stderrStr)
		}
		return fmt.Errorf("python renderer failed: %v", err)
	}
	return nil
}

func (r PythonPPTXRenderer) RenderPPTXBytes(ctx context.Context, spec any) ([]byte, error) {
	tmpFile, err := os.CreateTemp("", "render-*.pptx")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if err := r.RenderPPTX(ctx, spec, tmpFile.Name()); err != nil {
		return nil, err
	}

	return os.ReadFile(tmpFile.Name())
}

// GenerateSlideThumbnails creates preview thumbnails for each slide
// For Python renderer, this returns placeholder thumbnails
func (r PythonPPTXRenderer) GenerateSlideThumbnails(ctx context.Context, spec any) ([][]byte, error) {
	specBytes, err := specToJSONBytes(spec)
	if err != nil {
		return nil, err
	}

	var templateSpec struct {
		Layouts []struct {
			Name string `json:"name"`
		} `json:"layouts"`
	}

	if err := json.Unmarshal(specBytes, &templateSpec); err != nil {
		return nil, err
	}

	if len(templateSpec.Layouts) == 0 {
		return nil, errors.New("no layouts found in template spec")
	}

	var thumbnails [][]byte

	// Generate placeholder thumbnail for each layout
	for i := range templateSpec.Layouts {
		img := image.NewRGBA(image.Rect(0, 0, 400, 300))

		// Fill background
		for y := 0; y < 300; y++ {
			for x := 0; x < 400; x++ {
				img.Set(x, y, color.RGBA{230, 230, 250, 255})
			}
		}

		// Add slide number
		slideNumX, slideNumY := 20, 20
		for dy := 0; dy < 30; dy++ {
			for dx := 0; dx < 30; dx++ {
				if slideNumX+dx < 400 && slideNumY+dy < 300 {
					img.Set(slideNumX+dx, slideNumY+dy, color.RGBA{100, 100, 200, 255})
				}
			}
		}

		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return nil, fmt.Errorf("failed to encode thumbnail %d: %w", i+1, err)
		}

		thumbnails = append(thumbnails, buf.Bytes())
	}

	return thumbnails, nil
}

type GoPPTXRenderer struct {
	layoutGenerator       *SmartLayoutGenerator
	aiDesignAnalyzer      *AIDesignAnalyzer
	olamaAI               *OlamaAIBridge
	backgroundRenderer    *AdvancedBackgroundRenderer
	visualRenderer        *SmartVisualRenderer
	visualEnhancer        *VisualEnhancementRenderer
	typographySystem      *AdvancedTypographySystem
	templateLibrary       *DesignTemplateLibrary
}

func NewGoPPTXRenderer() *GoPPTXRenderer {
	return &GoPPTXRenderer{
		layoutGenerator:    NewSmartLayoutGenerator(),
		aiDesignAnalyzer:   NewAIDesignAnalyzer(),
		olamaAI:            NewOlamaAIBridge(),
		backgroundRenderer: NewAdvancedBackgroundRenderer(),
		visualRenderer:     NewSmartVisualRenderer(),
		visualEnhancer:     NewVisualEnhancementRenderer(),
		typographySystem:   NewAdvancedTypographySystem(),
		templateLibrary:    NewDesignTemplateLibrary(),
	}
}

func (r GoPPTXRenderer) RenderPPTX(ctx context.Context, spec any, outPath string) error {
	data, err := r.RenderPPTXBytes(ctx, spec)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}

	return os.WriteFile(outPath, data, 0o644)
}

func (r GoPPTXRenderer) RenderPPTXBytes(ctx context.Context, spec any) ([]byte, error) {
	// Parse the template spec
	specBytes, err := specToJSONBytes(spec)
	if err != nil {
		return nil, err
	}

	var templateSpec struct {
		Layouts []struct {
			Name         string `json:"name"`
			Placeholders []struct {
				ID       string `json:"id"`
				Type     string `json:"type"`
				Content  string `json:"content"`
				Geometry struct {
					X float64 `json:"x"`
					Y float64 `json:"y"`
					W float64 `json:"w"`
					H float64 `json:"h"`
				} `json:"geometry"`
			} `json:"placeholders"`
		} `json:"layouts"`
		Tokens struct {
			Colors struct {
				Primary    string `json:"primary"`
				Secondary  string `json:"secondary"`
				Background string `json:"background"`
				Text       string `json:"text"`
			} `json:"colors"`
		} `json:"tokens"`
	}

	if err := json.Unmarshal(specBytes, &templateSpec); err != nil {
		return nil, err
	}

	if len(templateSpec.Layouts) == 0 {
		return nil, errors.New("no layouts found in template spec")
	}

	// Create a new presentation with custom slide master
	ppt := presentation.New()

	// Note: Slide background will be applied per slide due to gooxml limitations

	// Perform AI design analysis using olama's AI if available
	jsonData := r.specToMap(templateSpec)
	companyInfo := CompanyContext{} // Could be extracted from brand kit

	var designIdentity *DesignIdentity
	var aiErr error

	// Try olama AI first (if HUGGINGFACE_API_KEY is available)
	if r.olamaAI.IsAvailable() && os.Getenv("HUGGINGFACE_API_KEY") != "" {
		designIdentity, aiErr = r.olamaAI.AnalyzeContentForDesign(jsonData, companyInfo)
		if aiErr != nil {
			// Log the error but fall back to the regular AI analyzer
			log.Printf("Olama AI analysis failed, falling back to regular analyzer: %v", aiErr)
		}
	}

	// Fall back to regular AI design analyzer if olama failed or is not available
	if designIdentity == nil || aiErr != nil {
		designIdentity, aiErr = r.aiDesignAnalyzer.AnalyzeContentForDesign(jsonData, companyInfo)
		if aiErr != nil {
			return nil, fmt.Errorf("AI design analysis failed: %v", aiErr)
		}
	}

	designTheme := r.templateLibrary.GetThemeForAnalysis(designIdentity)

	// Add a slide for each layout using advanced AI design
	for i, layout := range templateSpec.Layouts {
		slide := ppt.AddSlide()

		// Extract title and content for smart analysis
		var title, content string
		for _, ph := range layout.Placeholders {
			if strings.Contains(strings.ToLower(ph.ID), "title") {
				title = ph.Content
			} else {
				if content != "" {
					content += "\n"
				}
				content += ph.Content
			}
		}

		// Apply slide background first (using text box background)
		r.visualEnhancer.AddSlideBackground(slide, designTheme.Colors["background"])

		// Apply advanced visual elements and enhancements
		slideType := r.determineSlideType(title, content, i)
		r.visualEnhancer.ApplySlideEnhancements(slide, designTheme, slideType)
		r.visualRenderer.ApplyVisualElements(slide, designTheme, slideType)

		// Generate smart layout with industry-specific adjustments
		smartLayout := r.layoutGenerator.GenerateLayout(title, content, i+1, len(templateSpec.Layouts))

		// Override colors with theme colors
		smartLayout.ColorScheme = ColorScheme{
			Primary:    designTheme.Colors["primary"],
			Secondary:  designTheme.Colors["secondary"],
			Background: designTheme.Colors["background"],
			Text:       designTheme.Colors["text"],
			Accent:     designTheme.Colors["accent"],
		}

		// Add title with advanced typography
		if title != "" {
			titleBox := slide.AddTextBox()
			r.configureAdvancedTextBox(titleBox, smartLayout.Title, title, smartLayout.ColorScheme, designTheme)
		}

		// Add content with advanced typography and industry-specific styling
		for j, contentConfig := range smartLayout.Content {
			contentBox := slide.AddTextBox()
			contentText := content
			if j < len(layout.Placeholders)-1 {
				contentLines := strings.Split(content, "\n")
				if j < len(contentLines) {
					contentText = contentLines[j]
				}
			}
			r.configureAdvancedTextBox(contentBox, contentConfig, contentText, smartLayout.ColorScheme, designTheme)
		}
	}

	// Save to temp file
	tmpFile, err := os.CreateTemp("", "render-*.pptx")
	if err != nil {
		return nil, err
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()

	if err := ppt.SaveToFile(tmpPath); err != nil {
		os.Remove(tmpPath)
		return nil, err
	}

	// Read the file back
	data, err := os.ReadFile(tmpPath)
	if err != nil {
		os.Remove(tmpPath)
		return nil, err
	}

	// Clean up temp file
	os.Remove(tmpPath)

	return data, nil
}

func (r GoPPTXRenderer) applySlideBackground(ppt presentation.Presentation, theme DesignTheme) {
	// Apply background styling through slide master
	// This is the proper way to set backgrounds in PowerPoint
	if len(ppt.SlideMasters()) > 0 {
		master := ppt.SlideMasters()[0]

		// Apply theme background color to slide master
		bgColor := theme.Colors["background"]
		if bgColor != "" {
			// Note: gooxml has limited slide master background API
			// In a full implementation, this would use master.Background()
			// or manipulate the XML directly for rich backgrounds
			r.attemptSlideBackground(master, bgColor)
		}
	}
}

func (r GoPPTXRenderer) attemptSlideBackground(master presentation.SlideMaster, bgColor string) {
	// Attempt to set background through available gooxml APIs
	// Note: This may have limited effect due to gooxml constraints
	// Background setting in PowerPoint typically requires slide master manipulation
}

func (r GoPPTXRenderer) applySmartBackground(slide presentation.Slide, bg BackgroundConfig, colors ColorScheme) {
	// Apply background styling based on smart analysis
	// Note: gooxml has limited background styling options
	// This is a simplified implementation
}

func (r GoPPTXRenderer) configureTextBox(textBox presentation.TextBox, config PlaceholderConfig, text string, colors ColorScheme) {
	// Position and size (convert relative coords to 10x7.5in slide)
	props := textBox.Properties()
	x := measurement.Distance(config.X * 10 * measurement.Inch)
	y := measurement.Distance(config.Y * 7.5 * measurement.Inch)
	w := measurement.Distance(config.W * 10 * measurement.Inch)
	h := measurement.Distance(config.H * 7.5 * measurement.Inch)
	props.SetPosition(x, y)
	props.SetSize(w, h)
	props.SetNoFill()
	props.LineProperties().SetNoFill()

	// Add text with smart formatting
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		para := textBox.AddParagraph()

		// Smart bullet formatting for multi-line content
		if len(lines) > 1 && i > 0 && !config.Bold {
			para.Properties().SetBulletChar("•")
		}

		run := para.AddRun()
		run.SetText(line)

		// Smart typography
		rp := run.Properties()
		if config.Bold {
			rp.SetBold(true)
		}
		rp.SetSize(measurement.Distance(config.FontSize) * measurement.Point)
	}
}

// Helper methods for the enhanced renderer

func (r GoPPTXRenderer) specToMap(spec any) map[string]any {
	// Convert spec to map for AI analysis
	specBytes, _ := json.Marshal(spec)
	var specMap map[string]any
	json.Unmarshal(specBytes, &specMap)
	return specMap
}

func (r GoPPTXRenderer) determineSlideType(title, content string, slideIndex int) string {
	if slideIndex == 0 {
		return "title"
	}

	lowerContent := strings.ToLower(content)
	if strings.Contains(lowerContent, "summary") || strings.Contains(lowerContent, "conclusion") {
		return "conclusion"
	}
	if strings.Contains(lowerContent, "agenda") || strings.Contains(lowerContent, "outline") {
		return "agenda"
	}

	return "content"
}

func (r GoPPTXRenderer) configureAdvancedTextBox(textBox presentation.TextBox, config PlaceholderConfig, text string, colors ColorScheme, theme DesignTheme) {
	// Position and size (convert relative coords to 10x7.5in slide)
	props := textBox.Properties()
	x := measurement.Distance(config.X * 10 * measurement.Inch)
	y := measurement.Distance(config.Y * 7.5 * measurement.Inch)
	w := measurement.Distance(config.W * 10 * measurement.Inch)
	h := measurement.Distance(config.H * 7.5 * measurement.Inch)
	props.SetPosition(x, y)
	props.SetSize(w, h)
	props.SetNoFill()
	props.LineProperties().SetNoFill()

	// Determine optimal typography style
	position := config.ID
	style := r.typographySystem.GetOptimalStyle(text, position, theme.Name)

	// Apply advanced typography
	r.typographySystem.ApplyTypography(textBox, text, style, theme.Name)
}

func (r GoPPTXRenderer) parseColor(hexColor string) color.RGBA {
	// Remove # if present
	if strings.HasPrefix(hexColor, "#") {
		hexColor = hexColor[1:]
	}

	// Default to black if parsing fails
	if len(hexColor) != 6 {
		return color.RGBA{0, 0, 0, 255}
	}

	// Parse RGB values (simplified - would need proper hex parsing)
	return color.RGBA{0, 0, 0, 255} // Placeholder - would implement proper color parsing
}

// GenerateSlideThumbnails creates preview thumbnails for each slide
func (r GoPPTXRenderer) GenerateSlideThumbnails(ctx context.Context, spec any) ([][]byte, error) {
	// Parse the template spec
	specBytes, err := specToJSONBytes(spec)
	if err != nil {
		return nil, err
	}

	var templateSpec struct {
		Layouts []struct {
			Name         string `json:"name"`
			Placeholders []struct {
				ID       string `json:"id"`
				Type     string `json:"type"`
				Content  string `json:"content"`
				Geometry struct {
					X float64 `json:"x"`
					Y float64 `json:"y"`
					W float64 `json:"w"`
					H float64 `json:"h"`
				} `json:"geometry"`
			} `json:"placeholders"`
		} `json:"layouts"`
	}

	if err := json.Unmarshal(specBytes, &templateSpec); err != nil {
		return nil, err
	}

	if len(templateSpec.Layouts) == 0 {
		return nil, errors.New("no layouts found in template spec")
	}

	var thumbnails [][]byte

	// Generate a thumbnail for each layout (slide)
	for i, layout := range templateSpec.Layouts {
		// Create a simple PNG thumbnail
		img := image.NewRGBA(image.Rect(0, 0, 400, 300))

		// Fill background with a light gray color
		for y := 0; y < 300; y++ {
			for x := 0; x < 400; x++ {
				img.Set(x, y, color.RGBA{240, 240, 240, 255})
			}
		}

		// Add a border
		for x := 0; x < 400; x++ {
			img.Set(x, 0, color.RGBA{100, 100, 100, 255})
			img.Set(x, 299, color.RGBA{100, 100, 100, 255})
		}
		for y := 0; y < 300; y++ {
			img.Set(0, y, color.RGBA{100, 100, 100, 255})
			img.Set(399, y, color.RGBA{100, 100, 100, 255})
		}

		// Add slide number indicator
		slideNumX := 20
		slideNumY := 20
		for dy := 0; dy < 30; dy++ {
			for dx := 0; dx < 30; dx++ {
				if slideNumX+dx < 400 && slideNumY+dy < 300 {
					img.Set(slideNumX+dx, slideNumY+dy, color.RGBA{50, 50, 200, 255})
				}
			}
		}

		// Add placeholder indicators
		for _, ph := range layout.Placeholders {
			if ph.Type == "text" {
				// Calculate position relative to 400x300 thumbnail
				phX := int(ph.Geometry.X * 400)
				phY := int(ph.Geometry.Y * 300)
				phW := int(ph.Geometry.W * 400)
				phH := int(ph.Geometry.H * 300)

				// Ensure placeholder fits within bounds
				if phX < 0 {
					phX = 0
				}
				if phY < 0 {
					phY = 0
				}
				if phX+phW > 400 {
					phW = 400 - phX
				}
				if phY+phH > 300 {
					phH = 300 - phY
				}

				// Draw placeholder rectangle
				for dy := 0; dy < phH; dy++ {
					for dx := 0; dx < phW; dx++ {
						if phX+dx < 400 && phY+dy < 300 {
							img.Set(phX+dx, phY+dy, color.RGBA{200, 200, 255, 255})
						}
					}
				}
			}
		}

		// Convert image to PNG bytes
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return nil, fmt.Errorf("failed to encode thumbnail %d: %w", i+1, err)
		}

		thumbnails = append(thumbnails, buf.Bytes())
	}

	return thumbnails, nil
}
