package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3Client wraps MinIO client for object storage operations.
type S3Client struct {
	client     *minio.Client
	bucket     string
	publicURL  string // e.g., http://localhost:9010 or http://100.86.223.10:9010
}

// NewS3Client creates a new S3 client connection.
func NewS3Client(endpoint, accessKey, secretKey, bucket, publicURL string, useSSL bool) (*S3Client, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio connect: %w", err)
	}

	// Ensure bucket exists
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("bucket check: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("create bucket: %w", err)
		}
		// Set public read policy
		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Principal": {"AWS": ["*"]},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}]
		}`, bucket)
		client.SetBucketPolicy(ctx, bucket, policy)
	}

	return &S3Client{
		client:    client,
		bucket:    bucket,
		publicURL: publicURL,
	}, nil
}

// UploadFile uploads a local file to S3 and returns the public URL.
func (s *S3Client) UploadFile(ctx context.Context, localPath, objectKey string) (string, error) {
	file, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("stat file: %w", err)
	}

	// Detect content type
	contentType := "application/octet-stream"
	switch filepath.Ext(localPath) {
	case ".png":
		contentType = "image/png"
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".webp":
		contentType = "image/webp"
	}

	_, err = s.client.PutObject(ctx, s.bucket, objectKey, file, stat.Size(), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("put object: %w", err)
	}

	return s.publicURL + "/" + s.bucket + "/" + objectKey, nil
}

// UploadBytes uploads raw bytes to S3 and returns the public URL.
func (s *S3Client) UploadBytes(ctx context.Context, data []byte, objectKey, contentType string) (string, error) {
	_, err := s.client.PutObject(ctx, s.bucket, objectKey, io.NopCloser(io.Reader(nil)), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		// Try with a bytes reader
		_, err = s.client.PutObject(ctx, s.bucket, objectKey, bytesReader(data), int64(len(data)), minio.PutObjectOptions{
			ContentType: contentType,
		})
		if err != nil {
			return "", fmt.Errorf("put bytes: %w", err)
		}
	}
	return s.publicURL + "/" + s.bucket + "/" + objectKey, nil
}

func bytesReader(data []byte) *bytesReadCloser {
	return &bytesReadCloser{data: data}
}

type bytesReadCloser struct {
	data   []byte
	offset int
}

func (b *bytesReadCloser) Read(p []byte) (int, error) {
	if b.offset >= len(b.data) {
		return 0, io.EOF
	}
	n := copy(p, b.data[b.offset:])
	b.offset += n
	return n, nil
}

func (b *bytesReadCloser) Close() error { return nil }

// PublicURL returns the base public URL for the bucket.
func (s *S3Client) PublicURL() string {
	return s.publicURL + "/" + s.bucket
}
