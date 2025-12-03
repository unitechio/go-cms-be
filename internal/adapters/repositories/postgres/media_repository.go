package postgres

import (
	"context"

	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/pkg/errors"
	"github.com/owner/go-cms/pkg/logger"
	"github.com/owner/go-cms/pkg/pagination"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type mediaRepository struct {
	db *gorm.DB
}

// NewMediaRepository creates a new media repository
func NewMediaRepository(db *gorm.DB) repositories.MediaRepository {
	return &mediaRepository{db: db}
}

// Create creates a new media record
func (r *mediaRepository) Create(ctx context.Context, media *domain.Media) error {
	if err := r.db.WithContext(ctx).Create(media).Error; err != nil {
		logger.Error("Failed to create media", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to create media", 500)
	}
	return nil
}

// GetByID gets a media by ID
func (r *mediaRepository) GetByID(ctx context.Context, id uint) (*domain.Media, error) {
	var media domain.Media
	if err := r.db.WithContext(ctx).
		Preload("Uploader").
		First(&media, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		logger.Error("Failed to get media by ID", zap.Error(err), zap.Uint("id", id))
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get media", 500)
	}
	return &media, nil
}

// Update updates a media record
func (r *mediaRepository) Update(ctx context.Context, media *domain.Media) error {
	if err := r.db.WithContext(ctx).Save(media).Error; err != nil {
		logger.Error("Failed to update media", zap.Error(err), zap.Uint("id", media.ID))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to update media", 500)
	}
	return nil
}

// Delete deletes a media record (soft delete)
func (r *mediaRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Media{}, id).Error; err != nil {
		logger.Error("Failed to delete media", zap.Error(err), zap.Uint("id", id))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to delete media", 500)
	}
	return nil
}

// List lists media with offset pagination
func (r *mediaRepository) List(ctx context.Context, filter repositories.MediaFilter, page *pagination.OffsetPagination) ([]*domain.Media, int64, error) {
	var media []*domain.Media
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Media{})
	query = r.applyFilters(query, filter)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		logger.Error("Failed to count media", zap.Error(err))
		return nil, 0, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to count media", 500)
	}

	// Apply pagination
	if page != nil {
		query = query.Offset(page.GetOffset()).Limit(page.Limit)
	}

	// Fetch media
	if err := query.
		Preload("Uploader").
		Order("created_at DESC").
		Find(&media).Error; err != nil {
		logger.Error("Failed to list media", zap.Error(err))
		return nil, 0, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to list media", 500)
	}

	return media, total, nil
}

// ListWithCursor lists media with cursor pagination
func (r *mediaRepository) ListWithCursor(ctx context.Context, filter repositories.MediaFilter, cursor *pagination.Cursor, limit int) ([]*domain.Media, *pagination.Cursor, error) {
	var media []*domain.Media

	query := r.db.WithContext(ctx).Model(&domain.Media{})
	query = r.applyFilters(query, filter)

	// Apply cursor
	if cursor != nil && cursor.After != "" {
		query = query.Where("id > ?", cursor.After)
	}

	// Fetch media
	if err := query.
		Preload("Uploader").
		Order("id ASC").
		Limit(limit + 1).
		Find(&media).Error; err != nil {
		logger.Error("Failed to list media with cursor", zap.Error(err))
		return nil, nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to list media", 500)
	}

	// Build next cursor
	var nextCursor *pagination.Cursor
	if len(media) > limit {
		media = media[:limit]
		nextCursor = &pagination.Cursor{
			After: string(rune(media[len(media)-1].ID)),
		}
	}

	return media, nextCursor, nil
}

// applyFilters applies filters to query
func (r *mediaRepository) applyFilters(query *gorm.DB, filter repositories.MediaFilter) *gorm.DB {
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}

	if filter.UploaderID != nil {
		query = query.Where("uploaded_by = ?", *filter.UploaderID)
	}

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("file_name ILIKE ? OR original_name ILIKE ?",
			searchPattern, searchPattern)
	}

	if len(filter.IDs) > 0 {
		query = query.Where("id IN ?", filter.IDs)
	}

	if len(filter.Tags) > 0 {
		for _, tag := range filter.Tags {
			query = query.Where("tags LIKE ?", "%"+tag+"%")
		}
	}

	return query
}

// GetByType gets media by type
func (r *mediaRepository) GetByType(ctx context.Context, mediaType domain.MediaType) ([]*domain.Media, error) {
	var media []*domain.Media

	if err := r.db.WithContext(ctx).
		Where("type = ?", mediaType).
		Preload("Uploader").
		Order("created_at DESC").
		Find(&media).Error; err != nil {
		logger.Error("Failed to get media by type", zap.Error(err), zap.String("type", string(mediaType)))
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get media by type", 500)
	}

	return media, nil
}

// GetByUploader gets media by uploader
func (r *mediaRepository) GetByUploader(ctx context.Context, uploaderID uint) ([]*domain.Media, error) {
	var media []*domain.Media

	if err := r.db.WithContext(ctx).
		Where("uploaded_by = ?", uploaderID).
		Preload("Uploader").
		Order("created_at DESC").
		Find(&media).Error; err != nil {
		logger.Error("Failed to get media by uploader", zap.Error(err), zap.Uint("uploader_id", uploaderID))
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get media by uploader", 500)
	}

	return media, nil
}

// GetByBucket gets media by bucket
func (r *mediaRepository) GetByBucket(ctx context.Context, bucket string) ([]*domain.Media, error) {
	var media []*domain.Media

	if err := r.db.WithContext(ctx).
		Where("bucket = ?", bucket).
		Preload("Uploader").
		Order("created_at DESC").
		Find(&media).Error; err != nil {
		logger.Error("Failed to get media by bucket", zap.Error(err), zap.String("bucket", bucket))
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get media by bucket", 500)
	}

	return media, nil
}

// GetByObjectKey gets media by object key
func (r *mediaRepository) GetByObjectKey(ctx context.Context, objectKey string) (*domain.Media, error) {
	var media domain.Media

	if err := r.db.WithContext(ctx).
		Where("object_key = ?", objectKey).
		Preload("Uploader").
		First(&media).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		logger.Error("Failed to get media by object key", zap.Error(err), zap.String("object_key", objectKey))
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get media by object key", 500)
	}

	return &media, nil
}
