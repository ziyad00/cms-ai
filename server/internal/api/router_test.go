package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateEndpoint_OK(t *testing.T) {
	s := NewServer()
	h := s.Handler()

	body := map[string]any{
		"tokens":      map[string]any{"colors": map[string]any{"primary": "#3366FF"}},
		"constraints": map[string]any{"safeMargin": 0.05},
		"layouts": []any{
			map[string]any{
				"name": "Title",
				"placeholders": []any{
					map[string]any{"id": "title", "geometry": map[string]any{"x": 0.1, "y": 0.2, "w": 0.8, "h": 0.2}},
					map[string]any{"id": "subtitle", "geometry": map[string]any{"x": 0.1, "y": 0.45, "w": 0.8, "h": 0.15}},
				},
			},
		},
	}

	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/v1/templates/validate", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	addTestAuth(req, "user-1", "org-1", "Editor")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestValidateEndpoint_InvalidOverlap(t *testing.T) {
	s := NewServer()
	h := s.Handler()

	body := map[string]any{
		"tokens":      map[string]any{"colors": map[string]any{"primary": "#3366FF"}},
		"constraints": map[string]any{"safeMargin": 0.05},
		"layouts": []any{
			map[string]any{
				"name": "Bad",
				"placeholders": []any{
					map[string]any{"id": "a", "geometry": map[string]any{"x": 0.1, "y": 0.2, "w": 0.6, "h": 0.3}},
					map[string]any{"id": "b", "geometry": map[string]any{"x": 0.5, "y": 0.3, "w": 0.4, "h": 0.3}},
				},
			},
		},
	}

	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/v1/templates/validate", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	addTestAuth(req, "user-1", "org-1", "Editor")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(bytes.ToLower(w.Body.Bytes()), []byte("overlap")) {
		t.Fatalf("expected overlap in response, got: %s", w.Body.String())
	}
}
