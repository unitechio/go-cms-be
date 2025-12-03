package repositories

import (
	"context"
	"time"

	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/pkg/errors"
	"github.com/owner/go-cms/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Add to existing MediaRepository interface

// GetByHash gets media by SHA256 hash
func (r *mediaRepository) GetByHash(ctx context.Context, hash string) (*domain.Media, error) {
	var media domain.Media
	if err := r.db.WithContext(ctx).
		Where("file_hash = ?", hash).
		First(&media).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Not found is not an error for deduplication
		}
		logger.Error("Failed to get media by hash", zap.Error(err), zap.String("hash", hash))
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get media by hash", 500)
	}
	return &media, nil
}

// FindUnused finds media files not referenced in the last N days
func (r *mediaRepository) FindUnused(ctx context.Context, cutoffDate time.Time) ([]*domain.Media, error) {
	var media []*domain.Media

	// Find media that:
	// 1. Created before cutoff date
	// 2. Not attached to any posts
	// 3. Not used in documents
	if err := r.db.WithContext(ctx).
		Where("created_at < ?", cutoffDate).
		Where("id NOT IN (SELECT DISTINCT media_id FROM post_media)").
		Where("id NOT IN (SELECT DISTINCT media_id FROM documents WHERE media_id IS NOT NULL)").
		Find(&media).Error; err != nil {
		logger.Error("Failed to find unused media", zap.Error(err))
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to find unused media", 500)
	}

	return media, nil
}

// UpdateHash updates the file hash for a media record
func (r *mediaRepository) UpdateHash(ctx context.Context, id uint, hash string) error {
	if err := r.db.WithContext(ctx).
		Model(&domain.Media{}).
		Where("id = ?", id).
		Update("file_hash", hash).Error; err != nil {
		logger.Error("Failed to update media hash", zap.Error(err), zap.Uint("id", id))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to update media hash", 500)
	}
	return nil
}
