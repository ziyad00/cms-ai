package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func authHeaders(req *http.Request) {
	req.Header.Set("X-User-Id", "user-1")
	req.Header.Set("X-Org-Id", "org-1")
	req.Header.Set("X-Role", "Editor")
}

func TestGenerateThenListTemplates(t *testing.T) {
	s := NewServer()
	h := s.Handler()

	payload := map[string]any{"prompt": "Corporate minimal template"}
	b, _ := json.Marshal(payload)

	genReq := httptest.NewRequest(http.MethodPost, "/v1/templates/generate", bytes.NewReader(b))
	genReq.Header.Set("Content-Type", "application/json")
	authHeaders(genReq)
	genW := httptest.NewRecorder()
	h.ServeHTTP(genW, genReq)
	if genW.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", genW.Code, genW.Body.String())
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
