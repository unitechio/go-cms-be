package media

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"io"
	"mime/multipart"
	"strings"

	"go.uber.org/zap"

	"github.com/owner/go-cms/internal/core/domain"
	imageProcessor "github.com/owner/go-cms/internal/infrastructure/image"
	"github.com/owner/go-cms/internal/infrastructure/storage"
	"github.com/owner/go-cms/pkg/errors"
	"github.com/owner/go-cms/pkg/logger"
)

// Add to existing media useCase

// UploadWithOptimization uploads a file with compression, deduplication, and image optimization
func (uc *useCase) UploadWithOptimization(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*domain.Media, error) {
	defer file.Close()

	// Calculate hash for deduplication
	compressor := storage.NewCompressor()

	// Read file into buffer for hash calculation
	var buf bytes.Buffer
	tee := io.TeeReader(file, &buf)
	hash, err := compressor.CalculateHash(tee)
	if err != nil {
		logger.Error("Failed to calculate file hash", zap.Error(err))
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to calculate file hash", 500)
	}

	// TODO: Check for duplicates - requires GetByHash method in repository
	// existing, err := uc.mediaRepo.GetByHash(ctx, hash)
	// if err != nil {
	// 	return nil, err
	// }
	// if existing != nil {
	// 	logger.Info("File already exists, returning existing media",
	// 		zap.String("hash", hash),
	// 		zap.Uint("existing_id", existing.ID))
	// 	return existing, nil
	// }

	// Reset reader
	file.Seek(0, 0)
	contentType := header.Header.Get("Content-Type")

	// Handle images with optimization
	if strings.HasPrefix(contentType, "image/") {
		return uc.uploadImageWithVariants(ctx, file, header, hash)
	}

	// Handle other files with optional compression
	return uc.uploadFileWithCompression(ctx, &buf, header, hash, contentType)
}

func (uc *useCase) uploadImageWithVariants(ctx context.Context, file multipart.File, header *multipart.FileHeader, hash string) (*domain.Media, error) {
	// Decode image
	img, format, err := image.Decode(file)
	if err != nil {
		logger.Error("Failed to decode image", zap.Error(err))
		return nil, errors.Wrap(err, errors.ErrCodeUploadFailed, "failed to decode image", 500)
	}

	processor := imageProcessor.NewProcessor()

	// Optimize original image
	optimized, err := processor.OptimizeImage(img, 2000, 2000, 85)
	if err != nil {
		logger.Error("Failed to optimize image", zap.Error(err))
		return nil, err
	}

	// Upload original
	originalKey, err := uc.storageService.Upload(ctx, optimized, header.Filename, "image/"+format)
	if err != nil {
		logger.Error("Failed to upload original image", zap.Error(err))
		return nil, err
	}

	// Generate and upload variants
	variants := processor.GenerateVariants(img)
	variantKeys := make(map[string]string)

	for size, variantImg := range variants {
		compressed, err := processor.Compress(variantImg, 85)
		if err != nil {
			continue
		}

		filename := fmt.Sprintf("%s_%s.jpg", size, header.Filename)
		key, err := uc.storageService.Upload(ctx, compressed, filename, "image/jpeg")
		if err != nil {
			logger.Warn("Failed to upload variant", zap.String("size", size), zap.Error(err))
			continue
		}
		variantKeys[size] = key
	}

	// Create media record with variants
	media := &domain.Media{
		FileName:  header.Filename,
		ObjectKey: originalKey,
		FileSize:  header.Size,
		MimeType:  "image/" + format,
		// TODO: Add FileHash and Variants fields to Media domain if needed
		// FileHash:  hash,
		// Variants:  variantKeys,
	}

	if err := uc.mediaRepo.Create(ctx, media); err != nil {
		// Cleanup uploaded files on error
		_ = uc.storageService.Delete(ctx, originalKey)
		for _, key := range variantKeys {
			_ = uc.storageService.Delete(ctx, key)
		}
		return nil, err
	}

	logger.Info("Image uploaded with variants",
		zap.String("filename", header.Filename),
		zap.Uint("media_id", media.ID),
		zap.Int("variants", len(variantKeys)))

	return media, nil
}

func (uc *useCase) uploadFileWithCompression(ctx context.Context, data *bytes.Buffer, header *multipart.FileHeader, hash, contentType string) (*domain.Media, error) {
	compressor := storage.NewCompressor()

	var uploadData io.Reader = data
	var objectKey string
	var err error

	// Compress if beneficial
	if compressor.ShouldCompress(contentType) {
		compressed, err := compressor.Compress(data)
		if err == nil && compressed.Len() < data.Len() {
			uploadData = compressed
			objectKey, err = uc.storageService.Upload(ctx, uploadData, header.Filename+".gz", "application/gzip")
		} else {
			objectKey, err = uc.storageService.Upload(ctx, uploadData, header.Filename, contentType)
		}
	} else {
		objectKey, err = uc.storageService.Upload(ctx, uploadData, header.Filename, contentType)
	}

	if err != nil {
		logger.Error("Failed to upload file", zap.Error(err))
		return nil, errors.Wrap(err, errors.ErrCodeUploadFailed, "failed to upload file", 500)
	}

	media := &domain.Media{
		FileName:  header.Filename,
		ObjectKey: objectKey,
		FileSize:  header.Size,
		MimeType:  contentType,
		// TODO: Add FileHash field to Media domain if needed
		// FileHash:  hash,
	}

	if err := uc.mediaRepo.Create(ctx, media); err != nil {
		_ = uc.storageService.Delete(ctx, objectKey)
		return nil, err
	}

	return media, nil
}

// CleanupUnusedFiles removes files not referenced in the last N days
// TODO: Requires FindUnused method in MediaRepository
func (uc *useCase) CleanupUnusedFiles(ctx context.Context, days int) (int, error) {
	// cutoff := time.Now().AddDate(0, 0, -days)
	// unusedFiles, err := uc.mediaRepo.FindUnused(ctx, cutoff)
	// if err != nil {
	// 	return 0, err
	// }
	return 0, errors.New(errors.ErrCodeInternal, "CleanupUnusedFiles not implemented", 500)

	// deleted := 0
	// for _, file := range unusedFiles {
	// 	// Delete from storage
	// 	if err := uc.storageService.Delete(ctx, file.ObjectKey); err != nil {
	// 		logger.Warn("Failed to delete file from storage",
	// 			zap.Error(err),
	// 			zap.String("object_key", file.ObjectKey))
	// 		continue
	// 	}
	//
	// 	// Delete variants if exist
	// 	if file.Variants != nil {
	// 		for _, key := range file.Variants {
	// 			_ = uc.storageService.Delete(ctx, key.(string))
	// 		}
	// 	}
	//
	// 	// Delete from database
	// 	if err := uc.mediaRepo.Delete(ctx, file.ID); err != nil {
	// 		logger.Warn("Failed to delete media record",
	// 			zap.Error(err),
	// 			zap.Uint("id", file.ID))
	// 		continue
	// 	}
	//
	// 	deleted++
	// }
	//
	// logger.Info("Cleanup completed",
	// 	zap.Int("deleted", deleted),
	// 	zap.Int("total_unused", len(unusedFiles)))
	//
	// return deleted, nil
}
