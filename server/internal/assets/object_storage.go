package assets

import (
	"context"
	"io"
	"time"
)

// ObjectStorage represents an object storage backend (S3, GCS, etc.)
type ObjectStorage interface {
	// Upload uploads data to the specified key and returns metadata
	Upload(ctx context.Context, key string, data []byte, contentType string) (*ObjectMetadata, error)

	// UploadStream uploads data from a reader to the specified key
	UploadStream(ctx context.Context, key string, reader io.Reader, contentType string) (*ObjectMetadata, error)

	// GetURL returns a signed URL for accessing the object
	GetURL(ctx context.Context, key string, expiration time.Duration) (string, error)

	// Download retrieves the object data
	Download(ctx context.Context, key string) ([]byte, error)

	// DownloadStream retrieves the object as a reader
	DownloadStream(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete removes the object
	Delete(ctx context.Context, key string) error

	// Exists checks if the object exists
	Exists(ctx context.Context, key string) (bool, error)

	// ListObjects lists objects with a given prefix
	ListObjects(ctx context.Context, prefix string) ([]*ObjectMetadata, error)

	// GetMetadata retrieves object metadata without downloading
	GetMetadata(ctx context.Context, key string) (*ObjectMetadata, error)
}

// ObjectMetadata contains information about a stored object
type ObjectMetadata struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	ETag         string    `json:"etag"`
	ContentType  string    `json:"contentType"`
	LastModified time.Time `json:"lastModified"`
	URL          string    `json:"url,omitempty"` // Signed URL if requested
}

// StorageConfig holds configuration for object storage backends
type StorageConfig struct {
	Type     string            `json:"type"` // s3, gcs, local
	Settings map[string]string `json:"settings"`

	// Common settings
	Bucket string `json:"bucket"`
	Region string `json:"region"`

	// S3 settings
	AccessKeyID string `json:"accessKeyId"`
	SecretKey   string `json:"secretKey"`
	Endpoint    string `json:"endpoint,omitempty"` // For MinIO

	// GCS settings
	CredentialsJSON string `json:"credentialsJson"`

	// Local filesystem settings
	BasePath string `json:"basePath"`

	// URL settings
	URLExpiration time.Duration `json:"urlExpiration"`
	PublicBaseURL string        `json:"publicBaseUrl"` // For local dev
}

// PresignedURLOptions contains options for generating presigned URLs
type PresignedURLOptions struct {
	Expiration time.Duration     `json:"expiration"`
	Method     string            `json:"method"` // GET, PUT
	Headers    map[string]string `json:"headers,omitempty"`
}
