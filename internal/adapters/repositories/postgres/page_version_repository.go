package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"gorm.io/gorm"
)

type pageVersionRepository struct {
	db *gorm.DB
}

// NewPageVersionRepository creates a new page version repository instance
func NewPageVersionRepository(db *gorm.DB) repositories.PageVersionRepository {
	return &pageVersionRepository{db: db}
}

func (r *pageVersionRepository) CreateVersion(ctx context.Context, version *domain.PageVersion) error {
	return r.db.WithContext(ctx).Create(version).Error
}

func (r *pageVersionRepository) GetVersionsByPageID(ctx context.Context, pageID uuid.UUID) ([]*domain.PageVersion, error) {
	var versions []*domain.PageVersion
	if err := r.db.WithContext(ctx).
		Preload("Creator").
		Where("page_id = ?", pageID).
		Order("version_number DESC").
		Find(&versions).Error; err != nil {
		return nil, err
	}
	return versions, nil
}

func (r *pageVersionRepository) GetVersionByID(ctx context.Context, id uuid.UUID) (*domain.PageVersion, error) {
	var version domain.PageVersion
	if err := r.db.WithContext(ctx).Preload("Creator").First(&version, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &version, nil
}

func (r *pageVersionRepository) GetLatestVersionNumber(ctx context.Context, pageID uuid.UUID) (int, error) {
	var version domain.PageVersion
	err := r.db.WithContext(ctx).
		Where("page_id = ?", pageID).
		Order("version_number DESC").
		First(&version).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}

	return version.VersionNumber, nil
}
