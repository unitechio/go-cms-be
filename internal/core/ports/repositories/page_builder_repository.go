package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/pkg/pagination"
)

// PageRepository defines the interface for page data operations
type PageRepository interface {
	// Basic CRUD
	Create(ctx context.Context, page *domain.Page) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Page, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Page, error)
	Update(ctx context.Context, page *domain.Page) error
	Delete(ctx context.Context, id uuid.UUID) error

	// List operations
	List(ctx context.Context, filter PageFilter, page *pagination.OffsetPagination) ([]*domain.Page, int64, error)

	// Advanced operations
	Duplicate(ctx context.Context, originalPageID uuid.UUID, newTitle, newSlug string) (*domain.Page, error)
	Publish(ctx context.Context, id uuid.UUID) error
	GetWithBlocks(ctx context.Context, id uuid.UUID) (*domain.Page, error)
}

// PageFilter represents filters for page queries
type PageFilter struct {
	Status domain.PageStatus
	Search string // Search in title and slug
}

// BlockRepository defines the interface for block data operations
type BlockRepository interface {
	// Basic CRUD
	Create(ctx context.Context, block *domain.Block) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Block, error)
	Update(ctx context.Context, block *domain.Block) error
	Delete(ctx context.Context, id uuid.UUID) error

	// List operations
	List(ctx context.Context, filter BlockFilter, page *pagination.OffsetPagination) ([]*domain.Block, int64, error)
}

// BlockFilter represents filters for block queries
type BlockFilter struct {
	Search   string
	Category string
	Type     string
}

// PageBlockRepository defines the interface for page block data operations
type PageBlockRepository interface {
	AddBlockToPage(ctx context.Context, pageBlock *domain.PageBlock) error
	UpdatePageBlock(ctx context.Context, pageBlock *domain.PageBlock) error
	RemoveBlockFromPage(ctx context.Context, id uuid.UUID) error
	GetPageBlocks(ctx context.Context, pageID uuid.UUID) ([]*domain.PageBlock, error)
	ReorderBlocks(ctx context.Context, pageID uuid.UUID, blockOrders map[uuid.UUID]int) error
}

// PageVersionRepository defines the interface for page version data operations
type PageVersionRepository interface {
	CreateVersion(ctx context.Context, version *domain.PageVersion) error
	GetVersionsByPageID(ctx context.Context, pageID uuid.UUID) ([]*domain.PageVersion, error)
	GetVersionByID(ctx context.Context, id uuid.UUID) (*domain.PageVersion, error)
	GetLatestVersionNumber(ctx context.Context, pageID uuid.UUID) (int, error)
}

// ThemeSettingRepository defines the interface for theme setting data operations
type ThemeSettingRepository interface {
	Create(ctx context.Context, theme *domain.ThemeSetting) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.ThemeSetting, error)
	Update(ctx context.Context, theme *domain.ThemeSetting) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*domain.ThemeSetting, error)
	GetActive(ctx context.Context) (*domain.ThemeSetting, error)
	Activate(ctx context.Context, id uuid.UUID) error
}
