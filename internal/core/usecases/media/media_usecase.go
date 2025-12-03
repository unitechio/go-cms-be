package media

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/pkg/errors"
	"github.com/owner/go-cms/pkg/logger"
	"github.com/owner/go-cms/pkg/pagination"
	"go.uber.org/zap"
)

// StorageService defines the interface for file storage operations
type StorageService interface {
	Upload(ctx context.Context, file io.Reader, filename, contentType string) (string, error)
	Delete(ctx context.Context, objectKey string) error
	GetPresignedURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error)
}

// UseCase defines the media use case interface
type UseCase interface {
	// Media CRUD
	UploadMedia(ctx context.Context, file *multipart.FileHeader, uploaderID uuid.UUID) (*domain.Media, error)
	GetMedia(ctx context.Context, id uint) (*domain.Media, error)
	UpdateMedia(ctx context.Context, id uint, req UpdateMediaRequest) (*domain.Media, error)
	DeleteMedia(ctx context.Context, id uint) error
	ListMedia(ctx context.Context, filter repositories.MediaFilter, page *pagination.OffsetPagination) ([]*domain.Media, int64, error)

	// Media Operations
	GetPresignedURL(ctx context.Context, id uint, expiry time.Duration) (string, error)
	GetMediaByType(ctx context.Context, mediaType domain.MediaType) ([]*domain.Media, error)

	// Optimization features
	UploadWithOptimization(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*domain.Media, error)
	CleanupUnusedFiles(ctx context.Context, days int) (int, error)
}

// useCase implements the UseCase interface
type useCase struct {
	mediaRepo      repositories.MediaRepository
	storageService StorageService
	bucketName     string
}

// NewUseCase creates a new media use case
func NewUseCase(
	mediaRepo repositories.MediaRepository,
	storageService StorageService,
	bucketName string,
) UseCase {
	return &useCase{
		mediaRepo:      mediaRepo,
		storageService: storageService,
		bucketName:     bucketName,
	}
}

// UpdateMediaRequest represents an update media request
type UpdateMediaRequest struct {
	Alt         *string `json:"alt"`
	Caption     *string `json:"caption"`
	Description *string `json:"description"`
	Tags        *string `json:"tags"`
}

// UploadMedia uploads a media file
func (uc *useCase) UploadMedia(ctx context.Context, file *multipart.FileHeader, uploaderID uuid.UUID) (*domain.Media, error) {
	// Open the file
	src, err := file.Open()
	if err != nil {
		logger.Error("Failed to open uploaded file", zap.Error(err))
		return nil, errors.Wrap(err, errors.ErrCodeUploadFailed, "failed to open file", 500)
	}
	defer src.Close()

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d_%s%s", time.Now().Unix(), uuid.New().String()[:8], ext)

	// Determine media type from mime type
	mediaType := getMediaTypeFromMime(file.Header.Get("Content-Type"))

	// Upload to storage
	objectKey, err := uc.storageService.Upload(ctx, src, filename, file.Header.Get("Content-Type"))
	if err != nil {
		logger.Error("Failed to upload file to storage", zap.Error(err))
		return nil, errors.Wrap(err, errors.ErrCodeUploadFailed, "failed to upload file", 500)
	}

	// Create media record
	media := &domain.Media{
		FileName:     filename,
		OriginalName: file.Filename,
		FilePath:     objectKey,
		FileSize:     file.Size,
		MimeType:     file.Header.Get("Content-Type"),
		Type:         mediaType,
		UploadedBy:   uploaderID,
		Bucket:       uc.bucketName,
		ObjectKey:    objectKey,
	}

	// TODO: Extract image dimensions, video duration, etc.
	// This would require additional libraries like image/jpeg, image/png, etc.

	if err := uc.mediaRepo.Create(ctx, media); err != nil {
		// If database insert fails, try to delete from storage
		_ = uc.storageService.Delete(ctx, objectKey)
		logger.Error("Failed to create media record", zap.Error(err))
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to create media record", 500)
	}

	logger.Info("Media uploaded successfully",
		zap.Uint("media_id", media.ID),
		zap.String("filename", filename),
		zap.String("uploader_id", uploaderID.String()))

	return media, nil
}

// GetMedia gets a media by ID
func (uc *useCase) GetMedia(ctx context.Context, id uint) (*domain.Media, error) {
	media, err := uc.mediaRepo.GetByID(ctx, id)
	if err != nil {
		logger.Error("Failed to get media", zap.Error(err), zap.Uint("id", id))
		return nil, err
	}
	return media, nil
}

// UpdateMedia updates media metadata
func (uc *useCase) UpdateMedia(ctx context.Context, id uint, req UpdateMediaRequest) (*domain.Media, error) {
	// Get existing media
	media, err := uc.mediaRepo.GetByID(ctx, id)
	if err != nil {
		logger.Error("Media not found", zap.Error(err), zap.Uint("id", id))
		return nil, err
	}

	// Update fields if provided
	if req.Alt != nil {
		media.Alt = *req.Alt
	}

	if req.Caption != nil {
		media.Caption = *req.Caption
	}

	if req.Description != nil {
		media.Description = *req.Description
	}

	if req.Tags != nil {
		media.Tags = *req.Tags
	}

	// Update media
	if err := uc.mediaRepo.Update(ctx, media); err != nil {
		logger.Error("Failed to update media", zap.Error(err), zap.Uint("id", id))
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to update media", 500)
	}

	logger.Info("Media updated successfully", zap.Uint("media_id", media.ID))

	return media, nil
}

// DeleteMedia deletes a media file
func (uc *useCase) DeleteMedia(ctx context.Context, id uint) error {
	// Get media to retrieve object key
	media, err := uc.mediaRepo.GetByID(ctx, id)
	if err != nil {
		logger.Error("Media not found", zap.Error(err), zap.Uint("id", id))
		return err
	}

	// Delete from storage
	if err := uc.storageService.Delete(ctx, media.ObjectKey); err != nil {
		logger.Warn("Failed to delete file from storage, continuing with database deletion",
			zap.Error(err), zap.String("object_key", media.ObjectKey))
	}

	// Delete from database
	if err := uc.mediaRepo.Delete(ctx, id); err != nil {
		logger.Error("Failed to delete media record", zap.Error(err), zap.Uint("id", id))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to delete media", 500)
	}

	logger.Info("Media deleted successfully", zap.Uint("media_id", id))

	return nil
}

// ListMedia lists media with filters and pagination
func (uc *useCase) ListMedia(ctx context.Context, filter repositories.MediaFilter, page *pagination.OffsetPagination) ([]*domain.Media, int64, error) {
	media, total, err := uc.mediaRepo.List(ctx, filter, page)
	if err != nil {
		logger.Error("Failed to list media", zap.Error(err))
		return nil, 0, err
	}
	return media, total, nil
}

// GetPresignedURL gets a presigned URL for media access
func (uc *useCase) GetPresignedURL(ctx context.Context, id uint, expiry time.Duration) (string, error) {
	// Get media
	media, err := uc.mediaRepo.GetByID(ctx, id)
	if err != nil {
		logger.Error("Media not found", zap.Error(err), zap.Uint("id", id))
		return "", err
	}

	// Get presigned URL
	url, err := uc.storageService.GetPresignedURL(ctx, media.ObjectKey, expiry)
	if err != nil {
		logger.Error("Failed to get presigned URL", zap.Error(err), zap.String("object_key", media.ObjectKey))
		return "", errors.Wrap(err, errors.ErrCodeInternal, "failed to get presigned URL", 500)
	}

	return url, nil
}

// GetMediaByType gets media by type
func (uc *useCase) GetMediaByType(ctx context.Context, mediaType domain.MediaType) ([]*domain.Media, error) {
	media, err := uc.mediaRepo.GetByType(ctx, mediaType)
	if err != nil {
		logger.Error("Failed to get media by type", zap.Error(err), zap.String("type", string(mediaType)))
		return nil, err
	}
	return media, nil
}

// getMediaTypeFromMime determines media type from MIME type
func getMediaTypeFromMime(mimeType string) domain.MediaType {
	mimeType = strings.ToLower(mimeType)

	if strings.HasPrefix(mimeType, "image/") {
		return domain.MediaTypeImage
	}

	if strings.HasPrefix(mimeType, "video/") {
		return domain.MediaTypeVideo
	}

	if strings.HasPrefix(mimeType, "audio/") {
		return domain.MediaTypeAudio
	}

	// Document types
	documentMimes := []string{
		"application/pdf",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.ms-powerpoint",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"text/plain",
	}

	for _, docMime := range documentMimes {
		if mimeType == docMime {
			return domain.MediaTypeDocument
		}
	}

	return domain.MediaTypeOther
}
