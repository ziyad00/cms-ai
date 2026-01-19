package assets

import (
	"context"
	"testing"
	"time"
)

func TestLocalStorageFactory(t *testing.T) {
	ctx := context.Background()
	factory := NewStorageFactory()

	// Test local storage creation
	storage, err := factory.CreateStorage(ctx)
	if err != nil {
		t.Fatalf("Failed to create local storage: %v", err)
	}

	// Test basic operations
	testData := []byte("test data for local storage")
	metadata, err := storage.Upload(ctx, "test-key", testData, "text/plain")
	if err != nil {
		t.Fatalf("Failed to upload: %v", err)
	}

	if metadata.Key != "test-key" {
		t.Errorf("Expected key 'test-key', got '%s'", metadata.Key)
	}

	if metadata.Size != int64(len(testData)) {
		t.Errorf("Expected size %d, got %d", len(testData), metadata.Size)
	}

	// Test download
	downloadedData, err := storage.Download(ctx, "test-key")
	if err != nil {
		t.Fatalf("Failed to download: %v", err)
	}

	if string(downloadedData) != string(testData) {
		t.Errorf("Downloaded data mismatch: expected %s, got %s", string(testData), string(downloadedData))
	}

	// Test signed URL (local storage returns public URL)
	url, err := storage.GetURL(ctx, "test-key", 15*time.Minute)
	if err != nil {
		t.Fatalf("Failed to get URL: %v", err)
	}

	if url == "" {
		t.Error("Expected non-empty URL")
	}
}

func TestStorageFactoryConfig(t *testing.T) {
	ctx := context.Background()
	factory := NewStorageFactory()

	// Test with explicit config
	config := StorageConfig{
		Type:     "local",
		BasePath: "/tmp/test-assets",
		Settings: map[string]string{
			"publicBaseURL": "http://localhost:8080/assets",
		},
	}

	storage, err := factory.CreateStorageWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create storage with config: %v", err)
	}

	// This should be a LocalObjectStorage
	if storage == nil {
		t.Fatal("Storage is nil")
	}

	// Test invalid storage type
	invalidConfig := StorageConfig{
		Type: "invalid",
	}

	_, err = factory.CreateStorageWithConfig(ctx, invalidConfig)
	if err == nil {
		t.Error("Expected error for invalid storage type")
	}
}
