package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"gorm.io/gorm"
)

type pageBlockRepository struct {
	db *gorm.DB
}

// NewPageBlockRepository creates a new page block repository instance
func NewPageBlockRepository(db *gorm.DB) repositories.PageBlockRepository {
	return &pageBlockRepository{db: db}
}

func (r *pageBlockRepository) AddBlockToPage(ctx context.Context, pageBlock *domain.PageBlock) error {
	return r.db.WithContext(ctx).Create(pageBlock).Error
}

func (r *pageBlockRepository) UpdatePageBlock(ctx context.Context, pageBlock *domain.PageBlock) error {
	return r.db.WithContext(ctx).Save(pageBlock).Error
}

func (r *pageBlockRepository) RemoveBlockFromPage(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.PageBlock{}, "id = ?", id).Error
}

func (r *pageBlockRepository) GetPageBlocks(ctx context.Context, pageID uuid.UUID) ([]*domain.PageBlock, error) {
	var blocks []*domain.PageBlock
	if err := r.db.WithContext(ctx).
		Preload("Block").
		Where("page_id = ?", pageID).
		Order("\"order\" ASC").
		Find(&blocks).Error; err != nil {
		return nil, err
	}
	return blocks, nil
}

func (r *pageBlockRepository) ReorderBlocks(ctx context.Context, pageID uuid.UUID, blockOrders map[uuid.UUID]int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for blockID, order := range blockOrders {
			if err := tx.Model(&domain.PageBlock{}).
				Where("id = ? AND page_id = ?", blockID, pageID).
				Update("order", order).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
