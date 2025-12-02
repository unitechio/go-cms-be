package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/pkg/pagination"
	"gorm.io/gorm"
)

type pageRepository struct {
	db *gorm.DB
}

// NewPageRepository creates a new page repository instance
func NewPageRepository(db *gorm.DB) repositories.PageRepository {
	return &pageRepository{db: db}
}

func (r *pageRepository) Create(ctx context.Context, page *domain.Page) error {
	return r.db.WithContext(ctx).Create(page).Error
}

func (r *pageRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Page, error) {
	var page domain.Page
	if err := r.db.WithContext(ctx).Preload("Author").Preload("Blocks").Preload("Blocks.Block").First(&page, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &page, nil
}

func (r *pageRepository) GetBySlug(ctx context.Context, slug string) (*domain.Page, error) {
	var page domain.Page
	if err := r.db.WithContext(ctx).Preload("Author").Preload("Blocks").Preload("Blocks.Block").First(&page, "slug = ?", slug).Error; err != nil {
		return nil, err
	}
	return &page, nil
}

func (r *pageRepository) Update(ctx context.Context, page *domain.Page) error {
	return r.db.WithContext(ctx).Save(page).Error
}

func (r *pageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Page{}, "id = ?", id).Error
}

func (r *pageRepository) List(ctx context.Context, filter repositories.PageFilter, page *pagination.OffsetPagination) ([]*domain.Page, int64, error) {
	var pages []*domain.Page
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Page{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.Search != "" {
		search := "%" + filter.Search + "%"
		query = query.Where("title ILIKE ? OR slug ILIKE ?", search, search)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page != nil {
		query = query.Offset(page.GetOffset()).Limit(page.Limit)
	}

	if err := query.Preload("Author").Order("created_at DESC").Find(&pages).Error; err != nil {
		return nil, 0, err
	}

	return pages, total, nil
}

func (r *pageRepository) Duplicate(ctx context.Context, originalPageID uuid.UUID, newTitle, newSlug string) (*domain.Page, error) {
	var originalPage domain.Page
	if err := r.db.WithContext(ctx).Preload("Blocks").First(&originalPage, "id = ?", originalPageID).Error; err != nil {
		return nil, err
	}

	var newPage domain.Page
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create new page
		newPage = domain.Page{
			Title:          newTitle,
			Slug:           newSlug,
			Template:       originalPage.Template,
			Status:         domain.PageStatusDraft,
			SeoTitle:       originalPage.SeoTitle,
			SeoDescription: originalPage.SeoDescription,
			OgImage:        originalPage.OgImage,
			AuthorID:       originalPage.AuthorID, // Keep original author or should be passed? Assuming same author for now
		}

		if err := tx.Create(&newPage).Error; err != nil {
			return err
		}

		// Duplicate blocks
		for _, block := range originalPage.Blocks {
			newBlock := domain.PageBlock{
				PageID:        newPage.ID,
				BlockID:       block.BlockID,
				ParentBlockID: nil, // Reset parent for now as we need to map old parent IDs to new ones if we support nested
				Order:         block.Order,
				Config:        block.Config,
				Language:      block.Language,
			}
			// Note: Nested blocks duplication logic would be more complex, skipping deep nesting for MVP
			if err := tx.Create(&newBlock).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &newPage, nil
}

func (r *pageRepository) Publish(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&domain.Page{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       domain.PageStatusPublished,
		"published_at": now,
	}).Error
}

func (r *pageRepository) GetWithBlocks(ctx context.Context, id uuid.UUID) (*domain.Page, error) {
	var page domain.Page
	// Preload blocks and their definitions, ordered by order field
	if err := r.db.WithContext(ctx).
		Preload("Blocks", func(db *gorm.DB) *gorm.DB {
			return db.Order("\"order\" ASC") // Order by order column
		}).
		Preload("Blocks.Block").
		First(&page, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &page, nil
}
