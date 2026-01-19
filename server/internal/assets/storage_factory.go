package assets

import (
	"context"
	"fmt"
	"os"
	"time"
)

// StorageFactory creates object storage backends based on configuration
type StorageFactory struct{}

// NewStorageFactory creates a new storage factory
func NewStorageFactory() *StorageFactory {
	return &StorageFactory{}
}

// CreateStorage creates an object storage backend based on environment variables
func (f *StorageFactory) CreateStorage(ctx context.Context) (ObjectStorage, error) {
	storageType := os.Getenv("STORAGE_TYPE")
	if storageType == "" {
		storageType = "local" // Default to local storage
	}

	config := StorageConfig{
		Type:          storageType,
		Bucket:        os.Getenv("S3_BUCKET"),
		Region:        os.Getenv("AWS_REGION"),
		AccessKeyID:   os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretKey:     os.Getenv("AWS_SECRET_KEY"),
		Endpoint:      os.Getenv("S3_ENDPOINT"),
		BasePath:      os.Getenv("LOCAL_STORAGE_PATH"),
		URLExpiration: 15 * time.Minute, // Default 15 minutes
		Settings:      make(map[string]string),
	}

	// Add public base URL from settings
	if publicURL := os.Getenv("PUBLIC_BASE_URL"); publicURL != "" {
		config.Settings["publicBaseURL"] = publicURL
	}

	switch storageType {
	case "s3":
		return NewS3Storage(ctx, config)
	case "gcs":
		return NewGCSStorage(ctx, config)
	case "local":
		return NewLocalStorage(config)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageType)
	}
}

// CreateStorageWithConfig creates storage with explicit configuration
func (f *StorageFactory) CreateStorageWithConfig(ctx context.Context, config StorageConfig) (ObjectStorage, error) {
	switch config.Type {
	case "s3":
		return NewS3Storage(ctx, config)
	case "gcs":
		return NewGCSStorage(ctx, config)
	case "local":
		return NewLocalStorage(config)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", config.Type)
	}
}
