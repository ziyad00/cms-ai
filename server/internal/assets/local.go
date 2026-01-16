package assets

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// LocalStorage implements ObjectStorage using the local filesystem
type LocalStorage struct {
	basePath string
}

// NewLocalStorage creates a new local storage instance
func NewLocalStorage() (*LocalStorage, error) {
	basePath := os.Getenv("STORAGE_PATH")
	if basePath == "" {
		basePath = "/app/uploads"
	}
	
	// Ensure directory exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}
	
	return &LocalStorage{basePath: basePath}, nil
}

func (s *LocalStorage) Upload(ctx context.Context, key string, data []byte, contentType string) (*ObjectMetadata, error) {
	path := filepath.Join(s.basePath, key)
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}
	
	if err := os.WriteFile(path, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}
	
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	
	return &ObjectMetadata{
		Key:          key,
		Size:         info.Size(),
		ETag:         fmt.Sprintf("%x", info.ModTime().Unix()),
		ContentType:  contentType,
		LastModified: info.ModTime(),
	}, nil
}

func (s *LocalStorage) UploadStream(ctx context.Context, key string, reader io.Reader, contentType string) (*ObjectMetadata, error) {
	path := filepath.Join(s.basePath, key)
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}
	
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	
	size, err := io.Copy(file, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to write data: %w", err)
	}
	
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	
	return &ObjectMetadata{
		Key:          key,
		Size:         size,
		ETag:         fmt.Sprintf("%x", info.ModTime().Unix()),
		ContentType:  contentType,
		LastModified: info.ModTime(),
	}, nil
}

func (s *LocalStorage) GetURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	// For local storage, return a file:// URL or a path
	// In production, this would be a signed URL
	path := filepath.Join(s.basePath, key)
	return fmt.Sprintf("file://%s", path), nil
}

func (s *LocalStorage) Download(ctx context.Context, key string) ([]byte, error) {
	path := filepath.Join(s.basePath, key)
	return os.ReadFile(path)
}

func (s *LocalStorage) DownloadStream(ctx context.Context, key string) (io.ReadCloser, error) {
	path := filepath.Join(s.basePath, key)
	return os.Open(path)
}

func (s *LocalStorage) Delete(ctx context.Context, key string) error {
	path := filepath.Join(s.basePath, key)
	return os.Remove(path)
}

func (s *LocalStorage) Exists(ctx context.Context, key string) (bool, error) {
	path := filepath.Join(s.basePath, key)
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *LocalStorage) ListObjects(ctx context.Context, prefix string) ([]string, error) {
	dir := filepath.Join(s.basePath, prefix)
	var keys []string
	
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			rel, err := filepath.Rel(s.basePath, path)
			if err != nil {
				return err
			}
			keys = append(keys, rel)
		}
		return nil
	})
	
	return keys, err
}
