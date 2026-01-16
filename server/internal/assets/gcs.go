package assets

import (
	"context"
	"fmt"
	"io"
	"time"
)

// GCSStorage implements ObjectStorage using Google Cloud Storage
// This is a placeholder - implement when GCS is needed
type GCSStorage struct{}

// NewGCSStorage creates a new GCS storage instance
func NewGCSStorage() (*GCSStorage, error) {
	return &GCSStorage{}, fmt.Errorf("GCS storage not yet implemented")
}

func (s *GCSStorage) Upload(ctx context.Context, key string, data []byte, contentType string) (*ObjectMetadata, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *GCSStorage) UploadStream(ctx context.Context, key string, reader io.Reader, contentType string) (*ObjectMetadata, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *GCSStorage) GetURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (s *GCSStorage) Download(ctx context.Context, key string) ([]byte, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *GCSStorage) DownloadStream(ctx context.Context, key string) (io.ReadCloser, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *GCSStorage) Delete(ctx context.Context, key string) error {
	return fmt.Errorf("not implemented")
}

func (s *GCSStorage) Exists(ctx context.Context, key string) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (s *GCSStorage) ListObjects(ctx context.Context, prefix string) ([]string, error) {
	return nil, fmt.Errorf("not implemented")
}
