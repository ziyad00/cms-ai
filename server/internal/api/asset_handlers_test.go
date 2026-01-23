package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ziyad/cms-ai/server/internal/assets"
)

// MockObjectStorage implements ObjectStorage for testing
type MockObjectStorage struct {
	assets map[string][]byte
}

func NewMockObjectStorage() *MockObjectStorage {
	return &MockObjectStorage{
		assets: make(map[string][]byte),
	}
}

func (m *MockObjectStorage) Upload(ctx context.Context, key string, data []byte, contentType string) (*assets.ObjectMetadata, error) {
	m.assets[key] = data
	return &assets.ObjectMetadata{
		Key:          key,
		Size:         int64(len(data)),
		ETag:         "mock-etag",
		ContentType:  contentType,
		LastModified: time.Now(),
	}, nil
}

func (m *MockObjectStorage) UploadStream(ctx context.Context, key string, reader io.Reader, contentType string) (*assets.ObjectMetadata, error) {
	// Simplified mock implementation
	data := []byte("mock stream data")
	return m.Upload(ctx, key, data, contentType)
}

func (m *MockObjectStorage) GetURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	// Return a mock signed URL
	return "https://mock-signed-url.com/" + key, nil
}

type LocalURLObjectStorage struct {
	assets map[string][]byte
}

func (l *LocalURLObjectStorage) Upload(ctx context.Context, key string, data []byte, contentType string) (*assets.ObjectMetadata, error) {
	if l.assets == nil {
		l.assets = make(map[string][]byte)
	}
	l.assets[key] = data
	return &assets.ObjectMetadata{Key: key, Size: int64(len(data)), ContentType: contentType, LastModified: time.Now()}, nil
}

func (l *LocalURLObjectStorage) UploadStream(ctx context.Context, key string, reader io.Reader, contentType string) (*assets.ObjectMetadata, error) {
	return nil, fmt.Errorf("not implemented")
}

func (l *LocalURLObjectStorage) GetURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	return "/assets/" + key, nil
}

func (l *LocalURLObjectStorage) Download(ctx context.Context, key string) ([]byte, error) {
	if l.assets == nil {
		return nil, fmt.Errorf("asset not found: %s", key)
	}
	b, ok := l.assets[key]
	if !ok {
		return nil, fmt.Errorf("asset not found: %s", key)
	}
	return b, nil
}

func (l *LocalURLObjectStorage) DownloadStream(ctx context.Context, key string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}

func (l *LocalURLObjectStorage) Delete(ctx context.Context, key string) error { return nil }
func (l *LocalURLObjectStorage) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}
func (l *LocalURLObjectStorage) ListObjects(ctx context.Context, prefix string) ([]*assets.ObjectMetadata, error) {
	return nil, nil
}
func (l *LocalURLObjectStorage) GetMetadata(ctx context.Context, key string) (*assets.ObjectMetadata, error) {
	return nil, nil
}

func (m *MockObjectStorage) Download(ctx context.Context, key string) ([]byte, error) {
	data, ok := m.assets[key]
	if !ok {
		return nil, fmt.Errorf("asset not found: %s", key)
	}
	return data, nil
}

func (m *MockObjectStorage) DownloadStream(ctx context.Context, key string) (io.ReadCloser, error) {
	data, err := m.Download(ctx, key)
	if err != nil {
		return nil, err
	}
	return &mockReadCloser{data: data}, nil
}

func (m *MockObjectStorage) Delete(ctx context.Context, key string) error {
	delete(m.assets, key)
	return nil
}

func (m *MockObjectStorage) Exists(ctx context.Context, key string) (bool, error) {
	_, ok := m.assets[key]
	return ok, nil
}

func (m *MockObjectStorage) ListObjects(ctx context.Context, prefix string) ([]*assets.ObjectMetadata, error) {
	var objects []*assets.ObjectMetadata
	for key, data := range m.assets {
		if len(prefix) == 0 || key[:len(prefix)] == prefix {
			objects = append(objects, &assets.ObjectMetadata{
				Key:          key,
				Size:         int64(len(data)),
				ETag:         "mock-etag",
				LastModified: time.Now(),
			})
		}
	}
	return objects, nil
}

func (m *MockObjectStorage) GetMetadata(ctx context.Context, key string) (*assets.ObjectMetadata, error) {
	data, ok := m.assets[key]
	if !ok {
		return nil, fmt.Errorf("asset not found: %s", key)
	}
	return &assets.ObjectMetadata{
		Key:          key,
		Size:         int64(len(data)),
		ETag:         "mock-etag",
		LastModified: time.Now(),
	}, nil
}

// mockReadCloser implements io.ReadCloser for testing
type mockReadCloser struct {
	data   []byte
	offset int
}

func (m *mockReadCloser) Read(p []byte) (int, error) {
	if m.offset >= len(m.data) {
		return 0, io.EOF
	}
	n := copy(p, m.data[m.offset:])
	m.offset += n
	return n, nil
}

func (m *mockReadCloser) Close() error {
	return nil
}

func TestAssetDownloadHandlers(t *testing.T) {
	s := NewServer()

	// Use a storage that returns a *local* URL (relative path), which should trigger
	// the handler's download fallback (not an HTTP redirect).
	localStorage := &LocalURLObjectStorage{}
	s.ObjectStorage = localStorage

	h := s.Handler()

	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		headers        map[string]string
	}{
		{
			name:           "Asset download without auth",
			method:         "GET",
			path:           "/v1/assets/test-asset-1",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Asset download with auth but no asset",
			method:         "GET",
			path:           "/v1/assets/nonexistent-asset",
			expectedStatus: http.StatusNotFound,
			headers: map[string]string{
				"X-User-Id": "user-1",
				"X-Org-Id":  "org-1",
				"X-Role":    "Editor",
			},
		},
		{
			name:           "Asset download with auth and local URL storage",
			method:         "GET",
			path:           "/v1/assets/test-asset-1",
			expectedStatus: http.StatusNotFound,
			headers: map[string]string{
				"X-User-Id": "user-1",
				"X-Org-Id":  "org-1",
				"X-Role":    "Editor",
			},
		},
		{
			name:           "Job asset download without auth",
			method:         "GET",
			path:           "/v1/jobs/test-job/assets/test-file.pptx",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Job asset download with auth but no job",
			method:         "GET",
			path:           "/v1/jobs/nonexistent-job/assets/test-file.pptx",
			expectedStatus: http.StatusNotFound,
			headers: map[string]string{
				"X-User-Id": "user-1",
				"X-Org-Id":  "org-1",
				"X-Role":    "Editor",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Fatalf("%s: expected status %d, got %d: %s", tt.name, tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}
