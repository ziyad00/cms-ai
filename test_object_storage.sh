#!/bin/bash

# Test script for object storage functionality
echo "Testing Object Storage Integration..."
echo "=================================="

# Set up test environment
export STORAGE_TYPE=local
export LOCAL_STORAGE_PATH="/tmp/cms-ai-test-assets"

# Clean up any previous test data
rm -rf "$LOCAL_STORAGE_PATH"

echo "1. Testing local storage backend..."
go test ./server/internal/assets/... -v -run TestLocalStorageFactory

echo ""
echo "2. Testing storage backend selection..."
go test ./server/internal/api/... -v -run TestStorageBackendSelection

echo ""
echo "3. Testing object storage operations..."
go test ./server/internal/api/... -v -run TestObjectStorageOperations

echo ""
echo "4. Testing signed URL integration..."
go test ./server/internal/api/... -v -run TestSignedURLIntegration

echo ""
echo "5. Testing S3 backend (if credentials available)..."
if [ -n "$AWS_ACCESS_KEY_ID" ] && [ -n "$AWS_SECRET_KEY" ] && [ -n "$S3_BUCKET" ]; then
    export STORAGE_TYPE=s3
    export AWS_REGION=${AWS_REGION:-us-east-1}
    echo "S3 credentials found, testing S3 backend..."
    go test ./server/internal/assets/... -v -run TestS3Storage || echo "S3 test failed - check credentials and bucket permissions"
else
    echo "No S3 credentials found, skipping S3 backend test"
    echo "To test S3: export AWS_ACCESS_KEY_ID=xxx AWS_SECRET_KEY=xxx S3_BUCKET=bucket-name AWS_REGION=us-east-1"
fi

echo ""
echo "6. Running all API tests to ensure no regressions..."
go test ./server/internal/api/... -v

echo ""
echo "Object storage testing complete!"
echo ""
echo "To test different storage backends:"
echo "- Local:  STORAGE_TYPE=local go run ./cmd/server"
echo "- S3:     STORAGE_TYPE=s3 S3_BUCKET=my-bucket AWS_REGION=us-east-1 go run ./cmd/server"
echo "- GCS:    STORAGE_TYPE=gcs GCS_BUCKET=my-bucket go run ./cmd/server"
echo ""
echo "Environment variables for configuration:"
echo "- STORAGE_TYPE: local, s3, gcs"
echo "- LOCAL_STORAGE_PATH: Path for local storage (default: ./assets)"
echo "- S3_BUCKET: S3 bucket name"
echo "- AWS_REGION: AWS region"
echo "- AWS_ACCESS_KEY_ID: AWS access key"
echo "- AWS_SECRET_KEY: AWS secret key"
echo "- S3_ENDPOINT: Custom S3 endpoint (for MinIO)"
echo "- PUBLIC_BASE_URL: Base URL for signed URLs"