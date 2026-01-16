package assets

import (
	"context"
	"io"
	"time"
)

// ObjectMetadata contains metadata about a stored object
type ObjectMetadata struct {
	Key          string
	Size         int64
	ETag         string
	ContentType  string
	LastModified time.Time
}

// ObjectStorage interface for object storage operations
type ObjectStorage interface {
	// Upload uploads data to object storage
	Upload(ctx context.Context, key string, data []byte, contentType string) (*ObjectMetadata, error)
	
	// UploadStream uploads data from a stream
	UploadStream(ctx context.Context, key string, reader io.Reader, contentType string) (*ObjectMetadata, error)
	
	// GetURL generates a signed URL for downloading an object
	GetURL(ctx context.Context, key string, expiration time.Duration) (string, error)
	
	// Download downloads an object
	Download(ctx context.Context, key string) ([]byte, error)
	
	// DownloadStream downloads an object as a stream
	DownloadStream(ctx context.Context, key string) (io.ReadCloser, error)
	
	// Delete deletes an object
	Delete(ctx context.Context, key string) error
	
	// Exists checks if an object exists
	Exists(ctx context.Context, key string) (bool, error)
	
	// ListObjects lists objects with the given prefix
	ListObjects(ctx context.Context, prefix string) ([]string, error)
}

// Storage interface (legacy, for worker compatibility)
type Storage interface {
	Store(ctx context.Context, key string, data []byte) error
	Retrieve(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
}
