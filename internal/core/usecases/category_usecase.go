package usecases

import (
	"context"
	"fmt"

	"github.com/gosimple/slug"
	"github.com/owner/go-cms/internal/adapters/repositories/postgres"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/pkg/pagination"
)

// CategoryUseCase handles category business logic
type CategoryUseCase struct {
	repo *postgres.CategoryRepository
}

// NewCategoryUseCase creates a new category use case
func NewCategoryUseCase(repo *postgres.CategoryRepository) *CategoryUseCase {
	return &CategoryUseCase{repo: repo}
}

// CreateCategory creates a new category
func (uc *CategoryUseCase) CreateCategory(ctx context.Context, req *CreateCategoryRequest) (*domain.Category, error) {
	// Generate slug if not provided
	if req.Slug == "" {
		req.Slug = slug.Make(req.Name)
	} else {
		req.Slug = slug.Make(req.Slug)
	}

	// Check if slug already exists
	exists, err := uc.repo.CheckSlugExists(ctx, req.Slug, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to check slug existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("category with slug '%s' already exists", req.Slug)
	}

	// Validate parent if provided
	if req.ParentID != nil {
		if err := uc.repo.ValidateParent(ctx, 0, req.ParentID); err != nil {
			return nil, fmt.Errorf("invalid parent: %w", err)
		}
	}

	// Set default order if not provided
	if req.Order == nil {
		defaultOrder := 0
		req.Order = &defaultOrder
	}

	// Set default status if not provided
	if req.Status == "" {
		req.Status = domain.CategoryStatusActive
	}

	// Set default type if not provided
	if req.Type == "" {
		req.Type = domain.CategoryTypeBlog
	}

	category := &domain.Category{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		ParentID:    req.ParentID,
		Order:       *req.Order,
		Type:        req.Type,
		Status:      req.Status,
		Icon:        req.Icon,
		Color:       req.Color,
	}

	if err := uc.repo.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return category, nil
}

// GetCategory retrieves a category by ID
func (uc *CategoryUseCase) GetCategory(ctx context.Context, id uint) (*domain.Category, error) {
	category, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	return category, nil
}

// UpdateCategory updates an existing category
func (uc *CategoryUseCase) UpdateCategory(ctx context.Context, id uint, req *UpdateCategoryRequest) (*domain.Category, error) {
	// Get existing category
	category, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	// Update fields if provided
	if req.Name != nil {
		category.Name = *req.Name
		// Regenerate slug if name changed and slug not explicitly provided
		if req.Slug == nil {
			category.Slug = slug.Make(*req.Name)
		}
	}

	if req.Slug != nil {
		newSlug := slug.Make(*req.Slug)
		// Check if slug already exists (excluding current category)
		exists, err := uc.repo.CheckSlugExists(ctx, newSlug, id)
		if err != nil {
			return nil, fmt.Errorf("failed to check slug existence: %w", err)
		}
		if exists {
			return nil, fmt.Errorf("category with slug '%s' already exists", newSlug)
		}
		category.Slug = newSlug
	}

	if req.Description != nil {
		category.Description = *req.Description
	}

	if req.ParentID != nil {
		// Validate parent
		if err := uc.repo.ValidateParent(ctx, id, req.ParentID); err != nil {
			return nil, fmt.Errorf("invalid parent: %w", err)
		}
		category.ParentID = req.ParentID
	}

	if req.Order != nil {
		category.Order = *req.Order
	}

	if req.Type != nil {
		category.Type = *req.Type
	}

	if req.Status != nil {
		category.Status = *req.Status
	}

	if req.Icon != nil {
		category.Icon = *req.Icon
	}

	if req.Color != nil {
		category.Color = *req.Color
	}

	if err := uc.repo.Update(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return category, nil
}

// DeleteCategory deletes a category
func (uc *CategoryUseCase) DeleteCategory(ctx context.Context, id uint) error {
	// Check if category has children
	hasChildren, err := uc.repo.HasChildren(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check children: %w", err)
	}

	if hasChildren {
		return fmt.Errorf("cannot delete category with children")
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}

// ListCategories retrieves categories with pagination and filters
func (uc *CategoryUseCase) ListCategories(ctx context.Context, filters map[string]interface{}, page *pagination.OffsetPagination) ([]domain.Category, int64, error) {
	categories, total, err := uc.repo.List(ctx, filters, page)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list categories: %w", err)
	}
	return categories, total, nil
}

// GetCategoryTree retrieves categories in tree structure
func (uc *CategoryUseCase) GetCategoryTree(ctx context.Context, categoryType string) ([]domain.Category, error) {
	categories, err := uc.repo.GetTree(ctx, categoryType)
	if err != nil {
		return nil, fmt.Errorf("failed to get category tree: %w", err)
	}
	return categories, nil
}

// GetActiveCategories retrieves all active categories
func (uc *CategoryUseCase) GetActiveCategories(ctx context.Context, categoryType string) ([]domain.Category, error) {
	categories, err := uc.repo.GetActiveCategories(ctx, categoryType)
	if err != nil {
		return nil, fmt.Errorf("failed to get active categories: %w", err)
	}
	return categories, nil
}

// ReorderCategory updates the order and parent of a category
func (uc *CategoryUseCase) ReorderCategory(ctx context.Context, id uint, req *ReorderCategoryRequest) (*domain.Category, error) {
	// Validate parent if provided
	if req.ParentID != nil {
		if err := uc.repo.ValidateParent(ctx, id, req.ParentID); err != nil {
			return nil, fmt.Errorf("invalid parent: %w", err)
		}
	}

	if err := uc.repo.Reorder(ctx, id, req.ParentID, req.Order); err != nil {
		return nil, fmt.Errorf("failed to reorder category: %w", err)
	}

	return uc.repo.GetByID(ctx, id)
}

// Request DTOs

type CreateCategoryRequest struct {
	Name        string                `json:"name" binding:"required"`
	Slug        string                `json:"slug"`
	Description string                `json:"description"`
	ParentID    *uint                 `json:"parent_id"`
	Order       *int                  `json:"order"`
	Type        domain.CategoryType   `json:"type" binding:"required"`
	Status      domain.CategoryStatus `json:"status"`
	Icon        string                `json:"icon"`
	Color       string                `json:"color"`
}

type UpdateCategoryRequest struct {
	Name        *string                `json:"name"`
	Slug        *string                `json:"slug"`
	Description *string                `json:"description"`
	ParentID    *uint                  `json:"parent_id"`
	Order       *int                   `json:"order"`
	Type        *domain.CategoryType   `json:"type"`
	Status      *domain.CategoryStatus `json:"status"`
	Icon        *string                `json:"icon"`
	Color       *string                `json:"color"`
}

type ReorderCategoryRequest struct {
	ParentID *uint `json:"parent_id"`
	Order    int   `json:"order" binding:"required"`
}
