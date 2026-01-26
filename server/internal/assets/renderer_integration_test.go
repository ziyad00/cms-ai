package assets

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGoPPTXRenderer_RendersBulletsAndMultipleSlides(t *testing.T) {
	r := GoPPTXRenderer{}

	spec := map[string]any{
		"tokens": map[string]any{
			"colors": map[string]any{
				"primary":    "#2E75B6",
				"secondary":  "#5A6C7D",
				"background": "#FFFFFF",
				"text":       "#2C3E50",
			},
		},
		"layouts": []map[string]any{
			{
				"name": "Title",
				"placeholders": []map[string]any{
					{
						"id":       "title",
						"type":     "text",
						"content":  "Technical Proposal – Web Application Development",
						"geometry": map[string]any{"x": 0.1, "y": 0.15, "w": 0.8, "h": 0.15},
					},
					{
						"id":       "body",
						"type":     "text",
						"content":  "Submitted by: Sandrock\nClient: Example Authority",
						"geometry": map[string]any{"x": 0.1, "y": 0.35, "w": 0.8, "h": 0.3},
					},
				},
			},
			{
				"name": "Executive Summary",
				"placeholders": []map[string]any{
					{
						"id":       "title",
						"type":     "text",
						"content":  "Executive Summary",
						"geometry": map[string]any{"x": 0.05, "y": 0.08, "w": 0.9, "h": 0.12},
					},
					{
						"id":       "content",
						"type":     "text",
						"content":  "This proposal outlines the approach.\nSecure, scalable, maintainable.\nAligned with objectives.",
						"geometry": map[string]any{"x": 0.08, "y": 0.22, "w": 0.84, "h": 0.65},
					},
				},
			},
		},
	}

	b, err := json.Marshal(spec)
	require.NoError(t, err)

	pptx, err := r.RenderPPTXBytes(context.Background(), b)
	require.NoError(t, err)
	require.Greater(t, len(pptx), 1000)

	zr, err := zip.NewReader(bytes.NewReader(pptx), int64(len(pptx)))
	require.NoError(t, err)

	var slideXML []byte
	for _, f := range zr.File {
		if f.Name == "ppt/slides/slide1.xml" {
			rc, err := f.Open()
			require.NoError(t, err)
			slideXML, err = io.ReadAll(rc)
			require.NoError(t, err)
			rc.Close()
			break
		}
	}
	require.NotEmpty(t, slideXML)
	require.True(t, strings.Contains(string(slideXML), "Technical Proposal"))

	// Ensure we have multiple slides
	count := 0
	for _, f := range zr.File {
		if strings.HasPrefix(f.Name, "ppt/slides/slide") && strings.HasSuffix(f.Name, ".xml") {
			count++
		}
	}
	require.GreaterOrEqual(t, count, 2)
}
