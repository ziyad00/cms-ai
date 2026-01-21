package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ziyad/cms-ai/server/internal/assets"
)

func main() {
	renderer := assets.GoPPTXRenderer{}

	// Use the spec from the template response
	spec := map[string]interface{}{
		"constraints": map[string]interface{}{
			"safeMargin": 0.05,
		},
		"layouts": []map[string]interface{}{
			{
				"name": "Title / Hero",
				"placeholders": []map[string]interface{}{
					{
						"id":   "title",
						"type": "text",
						"geometry": map[string]interface{}{
							"x": 0.1,
							"y": 0.2,
							"w": 0.8,
							"h": 0.2,
						},
					},
					{
						"id":   "subtitle",
						"type": "text",
						"geometry": map[string]interface{}{
							"x": 0.1,
							"y": 0.45,
							"w": 0.8,
							"h": 0.15,
						},
					},
				},
			},
		},
		"tokens": map[string]interface{}{
			"colors": map[string]interface{}{
				"background": "#FFFFFF",
				"primary":    "#3366FF",
				"text":       "#111111",
			},
		},
	}

	data, err := renderer.RenderPPTXBytes(context.Background(), spec)
	if err != nil {
		log.Fatalf("Render failed: %v", err)
	}

	fmt.Printf("Successfully rendered PPTX with %d bytes\n", len(data))

	// Save to file to test
	err = os.WriteFile("test_output.pptx", data, 0644)
	if err != nil {
		log.Fatalf("Failed to write file: %v", err)
	}

	fmt.Println("Saved to test_output.pptx")
}
