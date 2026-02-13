package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ziyad/cms-ai/server/internal/spec"
)

func authHeaders(req *http.Request) {
	addTestAuth(req, "user-1", "org-1", "Editor")
}

func TestCreateThenListTemplates(t *testing.T) {
	s := NewServer()
	h := s.Handler()

	payload := map[string]any{"name": "Corporate minimal template"}
	b, _ := json.Marshal(payload)

	createReq := httptest.NewRequest(http.MethodPost, "/v1/templates", bytes.NewReader(b))
	createReq.Header.Set("Content-Type", "application/json")
	authHeaders(createReq)
	createW := httptest.NewRecorder()
	h.ServeHTTP(createW, createReq)
	if createW.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", createW.Code, createW.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/v1/templates", nil)
	authHeaders(listReq)
	listW := httptest.NewRecorder()
	h.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", listW.Code, listW.Body.String())
	}
	if !bytes.Contains(listW.Body.Bytes(), []byte("templates")) {
		t.Fatalf("expected templates in response, got: %s", listW.Body.String())
	}
}

func TestAuthRequired(t *testing.T) {
	s := NewServer()
	h := s.Handler()

	req := httptest.NewRequest(http.MethodGet, "/v1/templates", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAssetDownloadUnauthorized(t *testing.T) {
	s := NewServer()
	h := s.Handler()

	req := httptest.NewRequest(http.MethodGet, "/v1/assets/test-asset-1", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestBuildDeckSpecFromOutline_UsesLayoutHint(t *testing.T) {
	tplSpec := &spec.TemplateSpec{
		Layouts: []spec.Layout{
			{
				Name: "Base",
				Placeholders: []spec.Placeholder{
					{ID: "title", Type: "text"},
					{ID: "body", Type: "text"},
				},
			},
		},
	}

	outline := &DeckOutline{
		Slides: []SlideOutline{
			{SlideNumber: 1, Title: "Welcome", Content: []string{"Hello"}, LayoutHint: "title"},
			{SlideNumber: 2, Title: "Roadmap", Content: []string{"Q1", "Q2"}, LayoutHint: "timeline"},
			{SlideNumber: 3, Title: "Results", Content: []string{"50%"}, LayoutHint: "metrics"},
		},
	}

	result := buildDeckSpecFromOutline(tplSpec, outline)

	if len(result.Layouts) != 3 {
		t.Fatalf("expected 3 layouts, got %d", len(result.Layouts))
	}
	if result.Layouts[0].Name != "title" {
		t.Errorf("layout 0: expected name 'title', got %q", result.Layouts[0].Name)
	}
	if result.Layouts[1].Name != "timeline" {
		t.Errorf("layout 1: expected name 'timeline', got %q", result.Layouts[1].Name)
	}
	if result.Layouts[2].Name != "metrics" {
		t.Errorf("layout 2: expected name 'metrics', got %q", result.Layouts[2].Name)
	}
}

func TestBuildDeckSpecFromOutline_EmptyLayoutHintDefaultsToSimple(t *testing.T) {
	tplSpec := &spec.TemplateSpec{
		Layouts: []spec.Layout{
			{
				Name: "Base",
				Placeholders: []spec.Placeholder{
					{ID: "title", Type: "text"},
					{ID: "body", Type: "text"},
				},
			},
		},
	}

	outline := &DeckOutline{
		Slides: []SlideOutline{
			{SlideNumber: 1, Title: "Slide One", Content: []string{"Bullet A"}},
			{SlideNumber: 2, Title: "Slide Two", Content: []string{"Bullet B"}, LayoutHint: ""},
		},
	}

	result := buildDeckSpecFromOutline(tplSpec, outline)

	if len(result.Layouts) != 2 {
		t.Fatalf("expected 2 layouts, got %d", len(result.Layouts))
	}
	if result.Layouts[0].Name != "simple" {
		t.Errorf("layout 0: expected 'simple', got %q", result.Layouts[0].Name)
	}
	if result.Layouts[1].Name != "simple" {
		t.Errorf("layout 1: expected 'simple', got %q", result.Layouts[1].Name)
	}
}

func TestBuildDeckSpecFromOutline_PlaceholderContentFilled(t *testing.T) {
	tplSpec := &spec.TemplateSpec{
		Layouts: []spec.Layout{
			{
				Name: "Base",
				Placeholders: []spec.Placeholder{
					{ID: "title", Type: "text"},
					{ID: "body", Type: "text"},
				},
			},
		},
	}

	outline := &DeckOutline{
		Slides: []SlideOutline{
			{SlideNumber: 1, Title: "My Title", Content: []string{"Line 1", "Line 2"}, LayoutHint: "comparison"},
		},
	}

	result := buildDeckSpecFromOutline(tplSpec, outline)

	if len(result.Layouts) != 1 {
		t.Fatalf("expected 1 layout, got %d", len(result.Layouts))
	}
	phs := result.Layouts[0].Placeholders
	if len(phs) != 2 {
		t.Fatalf("expected 2 placeholders, got %d", len(phs))
	}
	if phs[0].Content != "My Title" {
		t.Errorf("title placeholder: expected 'My Title', got %q", phs[0].Content)
	}
	if phs[1].Content != "Line 1\nLine 2" {
		t.Errorf("body placeholder: expected 'Line 1\\nLine 2', got %q", phs[1].Content)
	}
}
