package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/pkg/pagination"
	"gorm.io/gorm"
)

type blockRepository struct {
	db *gorm.DB
}

// NewBlockRepository creates a new block repository instance
func NewBlockRepository(db *gorm.DB) repositories.BlockRepository {
	return &blockRepository{db: db}
}

func (r *blockRepository) Create(ctx context.Context, block *domain.Block) error {
	return r.db.WithContext(ctx).Create(block).Error
}

func (r *blockRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Block, error) {
	var block domain.Block
	if err := r.db.WithContext(ctx).First(&block, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &block, nil
}

func (r *blockRepository) Update(ctx context.Context, block *domain.Block) error {
	return r.db.WithContext(ctx).Save(block).Error
}

func (r *blockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Block{}, "id = ?", id).Error
}

func (r *blockRepository) List(ctx context.Context, filter repositories.BlockFilter, page *pagination.OffsetPagination) ([]*domain.Block, int64, error) {
	var blocks []*domain.Block
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Block{})

	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}

	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}

	if filter.Search != "" {
		search := "%" + filter.Search + "%"
		query = query.Where("name ILIKE ?", search)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page != nil {
		query = query.Offset(page.GetOffset()).Limit(page.Limit)
	}

	if err := query.Order("created_at DESC").Find(&blocks).Error; err != nil {
		return nil, 0, err
	}

	return blocks, total, nil
}
