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
	PythonPath string
	ScriptPath string
}

func (r PythonPPTXRenderer) RenderPPTX(ctx context.Context, spec any, outPath string) error {
	python := r.PythonPath
	if python == "" {
		python = "python3"
	}
	script := r.ScriptPath
	if script == "" {
		script = filepath.Join("tools", "renderer", "render_pptx.py")
	}

	tmpDir := filepath.Dir(outPath)
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return err
	}

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

	cmd := exec.CommandContext(ctx, python, script, tmpSpec.Name(), outPath)
	cmd.Env = append(os.Environ(), "PYTHONUNBUFFERED=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(out))
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

type GoPPTXRenderer struct{}

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

	// Create a new presentation
	ppt := presentation.New()

	// Add a slide for each layout
	for _, layout := range templateSpec.Layouts {
		slide := ppt.AddSlide()

		// Add placeholders as text boxes (simplified implementation)
		for _, ph := range layout.Placeholders {
			if ph.Type != "text" {
				continue
			}

			textBox := slide.AddTextBox()

			// Position and size (convert relative coords to 10x7.5in slide)
			props := textBox.Properties()
			x := measurement.Distance(ph.Geometry.X * 10 * measurement.Inch)
			y := measurement.Distance(ph.Geometry.Y * 7.5 * measurement.Inch)
			w := measurement.Distance(ph.Geometry.W * 10 * measurement.Inch)
			h := measurement.Distance(ph.Geometry.H * 7.5 * measurement.Inch)
			props.SetPosition(x, y)
			props.SetSize(w, h)
			props.SetNoFill()
			props.LineProperties().SetNoFill()

			content := ph.Content
			lines := strings.Split(content, "\n")
			for i, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				para := textBox.AddParagraph()
				if len(lines) > 1 {
					para.Properties().SetBulletChar("•")
				}
				if i > 0 {
					para.Properties().SetLevel(0)
				}
				run := para.AddRun()
				run.SetText(line)

				// Basic typography (keep it simple and readable)
				rp := run.Properties()
				if strings.Contains(strings.ToLower(ph.ID), "title") {
					rp.SetBold(true)
					rp.SetSize(28 * measurement.Point)
				} else {
					rp.SetSize(16 * measurement.Point)
				}
			}
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
