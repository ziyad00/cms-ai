package assets

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Storage implements ObjectStorage for AWS S3
type S3Storage struct {
	client  *s3.Client
	bucket  string
	config  StorageConfig
	baseURL string
}

// NewS3Storage creates a new S3 storage backend
func NewS3Storage(ctx context.Context, config StorageConfig) (*S3Storage, error) {
	if config.Type != "s3" {
		return nil, fmt.Errorf("invalid storage type: %s", config.Type)
	}

	if config.Bucket == "" {
		return nil, fmt.Errorf("bucket name is required for S3 storage")
	}

	// Create AWS configuration
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(config.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Override credentials if provided
	if config.AccessKeyID != "" && config.SecretKey != "" {
		cfg.Credentials = credentials.NewStaticCredentialsProvider(
			config.AccessKeyID,
			config.SecretKey,
			"",
		)
	}

	// Create S3 client
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if config.Endpoint != "" {
			o.BaseEndpoint = aws.String(config.Endpoint)
		}
	})

	// Ensure bucket exists
	if err := ensureBucketExists(ctx, client, config.Bucket, config.Region); err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	baseURL := config.Settings["publicBaseURL"]
	if baseURL == "" {
		// Default to virtual-hosted style URL
		baseURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com", config.Bucket, config.Region)
	}

	return &S3Storage{
		client:  client,
		bucket:  config.Bucket,
		config:  config,
		baseURL: baseURL,
	}, nil
}

// Upload uploads data to S3
func (s *S3Storage) Upload(ctx context.Context, key string, data []byte, contentType string) (*ObjectMetadata, error) {
	return s.UploadStream(ctx, key, bytes.NewReader(data), contentType)
}

// UploadStream uploads data from a reader to S3
func (s *S3Storage) UploadStream(ctx context.Context, key string, reader io.Reader, contentType string) (*ObjectMetadata, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload object: %w", err)
	}

	// Get object metadata
	return s.GetMetadata(ctx, key)
}

// GetURL returns a presigned URL for accessing the object
func (s *S3Storage) GetURL(ctx context.Context, key string, expiration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return req.URL, nil
}

// Download retrieves the object data
func (s *S3Storage) Download(ctx context.Context, key string) ([]byte, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download object: %w", err)
	}
	defer result.Body.Close()

	data, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object data: %w", err)
	}

	return data, nil
}

// DownloadStream retrieves the object as a reader
func (s *S3Storage) DownloadStream(ctx context.Context, key string) (io.ReadCloser, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download object: %w", err)
	}

	return result.Body, nil
}

// Delete removes the object from S3
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// Exists checks if the object exists in S3
func (s *S3Storage) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &noSuchKey) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}

	return true, nil
}

// ListObjects lists objects with a given prefix
func (s *S3Storage) ListObjects(ctx context.Context, prefix string) ([]*ObjectMetadata, error) {
	result, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	var objects []*ObjectMetadata
	for _, obj := range result.Contents {
		metadata := &ObjectMetadata{
			Key:          *obj.Key,
			Size:         aws.ToInt64(obj.Size),
			ETag:         *obj.ETag,
			LastModified: *obj.LastModified,
		}
		objects = append(objects, metadata)
	}

	return objects, nil
}

// GetMetadata retrieves object metadata without downloading
func (s *S3Storage) GetMetadata(ctx context.Context, key string) (*ObjectMetadata, error) {
	result, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object metadata: %w", err)
	}

	metadata := &ObjectMetadata{
		Key:          key,
		Size:         aws.ToInt64(result.ContentLength),
		ETag:         *result.ETag,
		ContentType:  *result.ContentType,
		LastModified: *result.LastModified,
	}

	return metadata, nil
}

// ensureBucketExists creates the bucket if it doesn't exist
func ensureBucketExists(ctx context.Context, client *s3.Client, bucket, region string) error {
	// Check if bucket exists
	_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err == nil {
		return nil // Bucket exists
	}

	// Try to create the bucket
	var cfg *s3.CreateBucketInput
	if region == "us-east-1" {
		// us-east-1 doesn't support LocationConstraint
		cfg = &s3.CreateBucketInput{
			Bucket: aws.String(bucket),
		}
	} else {
		cfg = &s3.CreateBucketInput{
			Bucket: aws.String(bucket),
			CreateBucketConfiguration: &types.CreateBucketConfiguration{
				LocationConstraint: types.BucketLocationConstraint(region),
			},
		}
	}

	_, err = client.CreateBucket(ctx, cfg)
	if err != nil {
		// Check if bucket already exists (might be owned by someone else)
		var bucketAlreadyExists *types.BucketAlreadyExists
		var bucketAlreadyOwnedByYou *types.BucketAlreadyOwnedByYou
		if errors.As(err, &bucketAlreadyExists) || errors.As(err, &bucketAlreadyOwnedByYou) {
			return nil // Bucket exists, that's fine
		}
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	return nil
}
