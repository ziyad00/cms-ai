package assets

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LocalObjectStorage implements ObjectStorage for local filesystem
type LocalObjectStorage struct {
	basePath string
	baseURL  string
}

// NewLocalStorage creates a new local filesystem storage backend
func NewLocalStorage(config StorageConfig) (*LocalObjectStorage, error) {
	if config.Type != "local" {
		return nil, fmt.Errorf("invalid storage type: %s", config.Type)
	}

	basePath := config.BasePath
	if basePath == "" {
		basePath = "./assets"
	}

	// Ensure base path exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	baseURL := config.Settings["publicBaseURL"]
	if baseURL == "" {
		// For development, use relative URLs
		baseURL = "/assets"
	}

	return &LocalObjectStorage{
		basePath: absPath,
		baseURL:  baseURL,
	}, nil
}

// getFullPath returns the full filesystem path for a given key
func (l *LocalObjectStorage) getFullPath(key string) string {
	return filepath.Join(l.basePath, key)
}

// Upload uploads data to local filesystem
func (l *LocalObjectStorage) Upload(ctx context.Context, key string, data []byte, contentType string) (*ObjectMetadata, error) {
	return l.UploadStream(ctx, key, strings.NewReader(string(data)), contentType)
}

// UploadStream uploads data from a reader to local filesystem
func (l *LocalObjectStorage) UploadStream(ctx context.Context, key string, reader io.Reader, contentType string) (*ObjectMetadata, error) {
	fullPath := l.getFullPath(key)

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy data
	size, err := io.Copy(file, reader)
	if err != nil {
		// Clean up on error
		os.Remove(fullPath)
		return nil, fmt.Errorf("failed to write data: %w", err)
	}

	// Get file info
	info, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	metadata := &ObjectMetadata{
		Key:          key,
		Size:         size,
		ETag:         fmt.Sprintf(`"%x"`, info.ModTime().UnixNano()),
		ContentType:  contentType,
		LastModified: info.ModTime(),
		URL:          l.baseURL + "/" + key,
	}

	return metadata, nil
}

// GetURL returns a public URL for accessing the object
func (l *LocalObjectStorage) GetURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	// For local storage, ignore expiration and return public URL
	return l.baseURL + "/" + key, nil
}

// Download retrieves the object data
func (l *LocalObjectStorage) Download(ctx context.Context, key string) ([]byte, error) {
	fullPath := l.getFullPath(key)

	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("object not found: %s", key)
		}
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

// DownloadStream retrieves the object as a reader
func (l *LocalObjectStorage) DownloadStream(ctx context.Context, key string) (io.ReadCloser, error) {
	fullPath := l.getFullPath(key)

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("object not found: %s", key)
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// Delete removes the object from local filesystem
func (l *LocalObjectStorage) Delete(ctx context.Context, key string) error {
	fullPath := l.getFullPath(key)

	err := os.Remove(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Object doesn't exist, that's fine
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// Exists checks if the object exists in local filesystem
func (l *LocalObjectStorage) Exists(ctx context.Context, key string) (bool, error) {
	fullPath := l.getFullPath(key)

	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return true, nil
}

// ListObjects lists objects with a given prefix
func (l *LocalObjectStorage) ListObjects(ctx context.Context, prefix string) ([]*ObjectMetadata, error) {
	baseDir := l.getFullPath(prefix)

	var objects []*ObjectMetadata

	err := filepath.Walk(l.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and the base directory itself
		if info.IsDir() || path == l.basePath {
			return nil
		}

		// Check if path has the desired prefix
		if strings.HasPrefix(path, baseDir) {
			// Get relative key
			key, err := filepath.Rel(l.basePath, path)
			if err != nil {
				return err
			}

			// Convert to forward slashes for consistency
			key = strings.ReplaceAll(key, string(filepath.Separator), "/")

			metadata := &ObjectMetadata{
				Key:          key,
				Size:         info.Size(),
				ETag:         fmt.Sprintf(`"%x"`, info.ModTime().UnixNano()),
				LastModified: info.ModTime(),
				URL:          l.baseURL + "/" + key,
			}
			objects = append(objects, metadata)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return objects, nil
}

// GetMetadata retrieves object metadata without downloading
func (l *LocalObjectStorage) GetMetadata(ctx context.Context, key string) (*ObjectMetadata, error) {
	fullPath := l.getFullPath(key)

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("object not found: %s", key)
		}
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	metadata := &ObjectMetadata{
		Key:          key,
		Size:         info.Size(),
		ETag:         fmt.Sprintf(`"%x"`, info.ModTime().UnixNano()),
		LastModified: info.ModTime(),
		URL:          l.baseURL + "/" + key,
	}

	return metadata, nil
}
