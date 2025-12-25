package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOClient wraps the MinIO client with helper methods
type MinIOClient struct {
	client     *minio.Client
	bucketName string
	publicURL  string
}

// Config holds MinIO configuration
type Config struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	BucketName string
	UseSSL     bool
	PublicURL  string // Public URL for accessing files (e.g., http://localhost:9000)
}

// NewMinIOClient creates a new MinIO client instance
func NewMinIOClient(cfg Config) (*MinIOClient, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	// Ensure bucket exists
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}

		// Set bucket policy to allow public read
		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Principal": {"AWS": ["*"]},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}]
		}`, cfg.BucketName)

		err = client.SetBucketPolicy(ctx, cfg.BucketName, policy)
		if err != nil {
			// Non-fatal: bucket created but policy might fail in some setups
			fmt.Printf("Warning: failed to set bucket policy: %v\n", err)
		}
	}

	publicURL := cfg.PublicURL
	if publicURL == "" {
		protocol := "http"
		if cfg.UseSSL {
			protocol = "https"
		}
		publicURL = fmt.Sprintf("%s://%s", protocol, cfg.Endpoint)
	}

	return &MinIOClient{
		client:     client,
		bucketName: cfg.BucketName,
		publicURL:  publicURL,
	}, nil
}

// UploadFile uploads a file and returns the public URL
func (m *MinIOClient) UploadFile(ctx context.Context, reader io.Reader, size int64, contentType string, originalName string) (string, error) {
	// Generate unique filename
	ext := filepath.Ext(originalName)
	objectName := fmt.Sprintf("uploads/%s/%s%s",
		time.Now().Format("2006/01/02"),
		uuid.New().String(),
		ext,
	)

	// Upload file
	_, err := m.client.PutObject(ctx, m.bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Return public URL
	url := fmt.Sprintf("%s/%s/%s", m.publicURL, m.bucketName, objectName)
	return url, nil
}

// NewMinIOClientFromEnv creates a MinIO client from environment variables
func NewMinIOClientFromEnv() (*MinIOClient, error) {
	cfg := Config{
		Endpoint:   getEnv("MINIO_ENDPOINT", "localhost:9000"),
		AccessKey:  getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		SecretKey:  getEnv("MINIO_SECRET_KEY", "minioadmin"),
		BucketName: getEnv("MINIO_BUCKET", "codeecho"),
		UseSSL:     getEnv("MINIO_USE_SSL", "false") == "true",
		PublicURL:  getEnv("MINIO_PUBLIC_URL", ""),
	}

	return NewMinIOClient(cfg)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
