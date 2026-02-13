package api

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ziyad/cms-ai/server/internal/assets"
	"github.com/ziyad/cms-ai/server/internal/store"
)

// TestSignedURLIntegration tests the full signed URL flow
func TestSignedURLIntegration(t *testing.T) {
	s := NewServer()

	// Replace with mock storage
	mockStorage := NewMockObjectStorage()
	s.ObjectStorage = mockStorage

	_ = s.Handler()

	// First, create a test asset in the mock storage
	testData := []byte("test pptx content")
	metadata, err := mockStorage.Upload(context.Background(), "test-asset.pptx", testData, "application/vnd.openxmlformats-officedocument.presentationml.presentation")
	if err != nil {
		t.Fatalf("Failed to upload test asset: %v", err)
	}

	// Create a test asset record in the store (using memory store)
	asset := store.Asset{
		ID:    "test-asset-id",
		OrgID: "test-org",
		Type:  store.AssetPPTX,
		Path:  metadata.Key,
		Mime:  metadata.ContentType,
	}
	createdAsset, err := s.Store.Assets().Create(context.Background(), asset)
	if err != nil {
		t.Fatalf("Failed to create asset record: %v", err)
	}

	// Test 1: Get download URL (should return signed URL)
	req := httptest.NewRequest("GET", "/v1/assets/"+createdAsset.ID+"/download-url", nil)
	addTestAuth(req, "test-user", "test-org", "Editor")
	w := httptest.NewRecorder()

	// Use the full router
	h := s.Handler()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Test 2: Direct asset download (should redirect to signed URL)
	req2 := httptest.NewRequest("GET", "/v1/assets/"+createdAsset.ID, nil)
	addTestAuth(req2, "test-user", "test-org", "Editor")
	w2 := httptest.NewRecorder()

	h.ServeHTTP(w2, req2)

	if w2.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected redirect status %d, got %d: %s", http.StatusTemporaryRedirect, w2.Code, w2.Body.String())
	}

	location := w2.Header().Get("Location")
	if location == "" {
		t.Error("Expected Location header in redirect")
	}

	if location != "https://mock-signed-url.com/"+metadata.Key {
		t.Errorf("Expected mock signed URL, got %s", location)
	}
}

// TestStorageBackendSelection tests different storage backends
func TestStorageBackendSelection(t *testing.T) {
	tests := []struct {
		name        string
		storageType string
		envVars     map[string]string
		expectError bool
	}{
		{
			name:        "Local storage default",
			storageType: "",
			envVars:     map[string]string{},
			expectError: false,
		},
		{
			name:        "Local storage explicit",
			storageType: "local",
			envVars: map[string]string{
				"LOCAL_STORAGE_PATH": "/tmp/test-assets",
			},
			expectError: false,
		},
{
			name:        "Invalid storage type",
			storageType: "invalid",
			envVars:     map[string]string{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}
			if tt.storageType != "" {
				t.Setenv("STORAGE_TYPE", tt.storageType)
			} else {
				t.Setenv("STORAGE_TYPE", "")
			}

			factory := assets.NewStorageFactory()
			storage, err := factory.CreateStorage(context.Background())

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if storage == nil {
				t.Error("Expected storage but got nil")
			}
		})
	}
}

// TestObjectStorageOperations tests the full object storage lifecycle
func TestObjectStorageOperations(t *testing.T) {
	ctx := context.Background()
	factory := assets.NewStorageFactory()

	// Use local storage for this test
	t.Setenv("STORAGE_TYPE", "local")
	t.Setenv("LOCAL_STORAGE_PATH", t.TempDir())

	storage, err := factory.CreateStorage(ctx)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	testKey := "test/operations/file.txt"
	testData := []byte("This is test content for object storage operations")
	contentType := "text/plain"

	// Test upload
	metadata, err := storage.Upload(ctx, testKey, testData, contentType)
	if err != nil {
		t.Fatalf("Upload failed: %v", err)
	}

	if metadata.Key != testKey {
		t.Errorf("Expected key %s, got %s", testKey, metadata.Key)
	}

	if metadata.Size != int64(len(testData)) {
		t.Errorf("Expected size %d, got %d", len(testData), metadata.Size)
	}

	if metadata.ContentType != contentType {
		t.Errorf("Expected content type %s, got %s", contentType, metadata.ContentType)
	}

	// Test exists
	exists, err := storage.Exists(ctx, testKey)
	if err != nil {
		t.Fatalf("Exists check failed: %v", err)
	}
	if !exists {
		t.Error("File should exist")
	}

	// Test download
	downloadedData, err := storage.Download(ctx, testKey)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	if !bytes.Equal(downloadedData, testData) {
		t.Errorf("Downloaded data mismatch: expected %s, got %s", string(testData), string(downloadedData))
	}

	// Test metadata check
	retrievedMetadata, err := storage.GetMetadata(ctx, testKey)
	if err != nil {
		t.Fatalf("GetMetadata failed: %v", err)
	}

	if retrievedMetadata.Key != testKey {
		t.Errorf("Metadata key mismatch: expected %s, got %s", testKey, retrievedMetadata.Key)
	}

	// Test signed URL generation
	url, err := storage.GetURL(ctx, testKey, 15*time.Minute)
	if err != nil {
		t.Fatalf("GetURL failed: %v", err)
	}

	if url == "" {
		t.Error("Expected non-empty URL")
	}

	// Test list operations
	objects, err := storage.ListObjects(ctx, "test/operations/")
	if err != nil {
		t.Fatalf("ListObjects failed: %v", err)
	}

	if len(objects) != 1 {
		t.Errorf("Expected 1 object in list, got %d", len(objects))
	}

	if objects[0].Key != testKey {
		t.Errorf("Listed key mismatch: expected %s, got %s", testKey, objects[0].Key)
	}

	// Test delete
	err = storage.Delete(ctx, testKey)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	exists, err = storage.Exists(ctx, testKey)
	if err != nil {
		t.Fatalf("Post-delete exists check failed: %v", err)
	}
	if exists {
		t.Error("File should not exist after deletion")
	}
}
