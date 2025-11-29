package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/owner/go-cms/internal/config"
	"github.com/owner/go-cms/pkg/logger"
	"go.uber.org/zap"
)

// Client is the global MinIO client
var Client *minio.Client

// Init initializes the MinIO connection
func Init(cfg *config.MinIOConfig) error {
	var err error

	// Initialize MinIO client
	Client, err = minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize MinIO client: %w", err)
	}

	// Create bucket if it doesn't exist
	ctx := context.Background()
	exists, err := Client.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = Client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		logger.Info("MinIO bucket created", zap.String("bucket", cfg.Bucket))
	}

	logger.Info("MinIO connected successfully",
		zap.String("endpoint", cfg.Endpoint),
		zap.String("bucket", cfg.Bucket),
	)

	return nil
}

// GetClient returns the MinIO client
func GetClient() *minio.Client {
	return Client
}

// UploadFile uploads a file to MinIO
func UploadFile(ctx context.Context, bucket, objectName string, reader io.Reader, size int64, contentType string) (*minio.UploadInfo, error) {
	info, err := Client.PutObject(ctx, bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return &info, nil
}

// DownloadFile downloads a file from MinIO
func DownloadFile(ctx context.Context, bucket, objectName string) (*minio.Object, error) {
	object, err := Client.GetObject(ctx, bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return object, nil
}

// DeleteFile deletes a file from MinIO
func DeleteFile(ctx context.Context, bucket, objectName string) error {
	err := Client.RemoveObject(ctx, bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetPresignedURL generates a presigned URL for file access
func GetPresignedURL(ctx context.Context, bucket, objectName string, expiry time.Duration) (string, error) {
	url, err := Client.PresignedGetObject(ctx, bucket, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// GetPresignedUploadURL generates a presigned URL for file upload
func GetPresignedUploadURL(ctx context.Context, bucket, objectName string, expiry time.Duration) (string, error) {
	url, err := Client.PresignedPutObject(ctx, bucket, objectName, expiry)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}

	return url.String(), nil
}

// ListFiles lists files in a bucket with prefix
func ListFiles(ctx context.Context, bucket, prefix string) ([]minio.ObjectInfo, error) {
	var objects []minio.ObjectInfo

	objectCh := Client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list files: %w", object.Err)
		}
		objects = append(objects, object)
	}

	return objects, nil
}

// FileExists checks if a file exists in MinIO
func FileExists(ctx context.Context, bucket, objectName string) (bool, error) {
	_, err := Client.StatObject(ctx, bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return true, nil
}

// GetFileInfo gets file information
func GetFileInfo(ctx context.Context, bucket, objectName string) (*minio.ObjectInfo, error) {
	info, err := Client.StatObject(ctx, bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &info, nil
}

// CopyFile copies a file within MinIO
func CopyFile(ctx context.Context, srcBucket, srcObject, destBucket, destObject string) error {
	src := minio.CopySrcOptions{
		Bucket: srcBucket,
		Object: srcObject,
	}

	dst := minio.CopyDestOptions{
		Bucket: destBucket,
		Object: destObject,
	}

	_, err := Client.CopyObject(ctx, dst, src)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// DeleteFiles deletes multiple files from MinIO
func DeleteFiles(ctx context.Context, bucket string, objectNames []string) error {
	objectsCh := make(chan minio.ObjectInfo)

	go func() {
		defer close(objectsCh)
		for _, name := range objectNames {
			objectsCh <- minio.ObjectInfo{
				Key: name,
			}
		}
	}()

	errorCh := Client.RemoveObjects(ctx, bucket, objectsCh, minio.RemoveObjectsOptions{})

	for err := range errorCh {
		if err.Err != nil {
			return fmt.Errorf("failed to delete file %s: %w", err.ObjectName, err.Err)
		}
	}

	return nil
}

// GenerateObjectName generates a unique object name for file storage
func GenerateObjectName(prefix, filename string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s/%d_%s", prefix, timestamp, filename)
}
