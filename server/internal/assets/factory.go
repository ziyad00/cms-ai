package assets

import (
	"context"
	"os"
)

// StorageFactory creates storage instances
type StorageFactory struct{}

// NewStorageFactory creates a new storage factory
func NewStorageFactory() *StorageFactory {
	return &StorageFactory{}
}

// CreateStorage creates a storage instance based on environment variables
func (f *StorageFactory) CreateStorage(ctx context.Context) (ObjectStorage, error) {
	storageType := os.Getenv("STORAGE_TYPE")
	if storageType == "" {
		storageType = "local"
	}

	switch storageType {
	case "s3":
		return NewS3Storage()
	case "gcs":
		return NewGCSStorage()
	case "local":
		fallthrough
	default:
		return NewLocalStorage()
	}
}
