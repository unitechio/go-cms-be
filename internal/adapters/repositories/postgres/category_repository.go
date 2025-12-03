package postgres

import (
	"context"
	"fmt"

	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/pkg/pagination"
	"gorm.io/gorm"
)

// CategoryRepository handles category data operations
type CategoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{db: db}
}

// Create creates a new category
func (r *CategoryRepository) Create(ctx context.Context, category *domain.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

// GetByID retrieves a category by ID
func (r *CategoryRepository) GetByID(ctx context.Context, id uint) (*domain.Category, error) {
	var category domain.Category
	err := r.db.WithContext(ctx).
		Preload("Parent").
		First(&category, id).Error
	if err != nil {
		return nil, err
	}

	// Get children count
	var count int64
	r.db.WithContext(ctx).Model(&domain.Category{}).Where("parent_id = ?", id).Count(&count)
	category.ChildrenCount = int(count)

	return &category, nil
}

// GetBySlug retrieves a category by slug
func (r *CategoryRepository) GetBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	var category domain.Category
	err := r.db.WithContext(ctx).
		Preload("Parent").
		Where("slug = ?", slug).
		First(&category).Error
	if err != nil {
		return nil, err
	}

	// Get children count
	var count int64
	r.db.WithContext(ctx).Model(&domain.Category{}).Where("parent_id = ?", category.ID).Count(&count)
	category.ChildrenCount = int(count)

	return &category, nil
}

// Update updates a category
func (r *CategoryRepository) Update(ctx context.Context, category *domain.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

// Delete soft deletes a category
func (r *CategoryRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Category{}, id).Error
}

// List retrieves categories with pagination and filters
func (r *CategoryRepository) List(ctx context.Context, filters map[string]interface{}, page *pagination.OffsetPagination) ([]domain.Category, int64, error) {
	var categories []domain.Category
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Category{})

	// Apply filters
	if search, ok := filters["search"].(string); ok && search != "" {
		query = query.Where("name LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if categoryType, ok := filters["type"].(string); ok && categoryType != "" {
		query = query.Where("type = ?", categoryType)
	}

	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}

	if parentID, ok := filters["parent_id"]; ok {
		if parentID == nil || parentID == "null" {
			query = query.Where("parent_id IS NULL")
		} else if pid, ok := parentID.(uint); ok {
			query = query.Where("parent_id = ?", pid)
		}
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := query.
		Preload("Parent").
		Order("\"order\" ASC, created_at DESC").
		Offset(page.GetOffset()).
		Limit(page.Limit).
		Find(&categories).Error

	if err != nil {
		return nil, 0, err
	}

	// Get children count for each category
	for i := range categories {
		var count int64
		r.db.WithContext(ctx).Model(&domain.Category{}).Where("parent_id = ?", categories[i].ID).Count(&count)
		categories[i].ChildrenCount = int(count)
	}

	return categories, total, nil
}

// GetTree retrieves categories in tree structure
func (r *CategoryRepository) GetTree(ctx context.Context, categoryType string) ([]domain.Category, error) {
	var categories []domain.Category

	query := r.db.WithContext(ctx).Model(&domain.Category{})

	// Filter by type if provided
	if categoryType != "" {
		query = query.Where("type = ?", categoryType)
	}

	// Get all categories ordered by parent and order
	err := query.
		Order("parent_id ASC, \"order\" ASC, created_at DESC").
		Find(&categories).Error

	if err != nil {
		return nil, err
	}

	// Build tree structure
	return r.buildTree(categories, nil), nil
}

// buildTree recursively builds a tree structure from flat categories
func (r *CategoryRepository) buildTree(categories []domain.Category, parentID *uint) []domain.Category {
	tree := make([]domain.Category, 0)

	for i := range categories {
		// Check if this category belongs to the current parent level
		if (parentID == nil && categories[i].ParentID == nil) ||
			(parentID != nil && categories[i].ParentID != nil && *categories[i].ParentID == *parentID) {

			// Get children for this category
			categoryID := categories[i].ID
			children := r.buildTree(categories, &categoryID)
			categories[i].Children = children
			categories[i].ChildrenCount = len(children)

			tree = append(tree, categories[i])
		}
	}

	return tree
}

// GetActiveCategories retrieves all active categories
func (r *CategoryRepository) GetActiveCategories(ctx context.Context, categoryType string) ([]domain.Category, error) {
	var categories []domain.Category

	query := r.db.WithContext(ctx).
		Where("status = ?", domain.CategoryStatusActive)

	if categoryType != "" {
		query = query.Where("type = ?", categoryType)
	}

	err := query.
		Preload("Parent").
		Order("\"order\" ASC, name ASC").
		Find(&categories).Error

	if err != nil {
		return nil, err
	}

	// Get children count for each category
	for i := range categories {
		var count int64
		r.db.WithContext(ctx).Model(&domain.Category{}).Where("parent_id = ?", categories[i].ID).Count(&count)
		categories[i].ChildrenCount = int(count)
	}

	return categories, nil
}

// Reorder updates the order and parent of a category
func (r *CategoryRepository) Reorder(ctx context.Context, id uint, parentID *uint, order int) error {
	updates := map[string]interface{}{
		"parent_id": parentID,
		"order":     order,
	}

	return r.db.WithContext(ctx).
		Model(&domain.Category{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// CheckSlugExists checks if a slug already exists (excluding a specific ID)
func (r *CategoryRepository) CheckSlugExists(ctx context.Context, slug string, excludeID uint) (bool, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&domain.Category{}).Where("slug = ?", slug)

	if excludeID > 0 {
		query = query.Where("id != ?", excludeID)
	}

	err := query.Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetChildrenCount returns the number of children for a category
func (r *CategoryRepository) GetChildrenCount(ctx context.Context, parentID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Category{}).
		Where("parent_id = ?", parentID).
		Count(&count).Error

	return count, err
}

// HasChildren checks if a category has children
func (r *CategoryRepository) HasChildren(ctx context.Context, id uint) (bool, error) {
	count, err := r.GetChildrenCount(ctx, id)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ValidateParent validates that the parent category exists and prevents circular references
func (r *CategoryRepository) ValidateParent(ctx context.Context, categoryID uint, parentID *uint) error {
	if parentID == nil {
		return nil
	}

	// Check if parent exists
	var parent domain.Category
	if err := r.db.WithContext(ctx).First(&parent, *parentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("parent category not found")
		}
		return err
	}

	// Prevent setting self as parent
	if categoryID == *parentID {
		return fmt.Errorf("category cannot be its own parent")
	}

	// Prevent circular reference by checking if the parent is a descendant
	if categoryID > 0 {
		isDescendant, err := r.isDescendant(ctx, categoryID, *parentID)
		if err != nil {
			return err
		}
		if isDescendant {
			return fmt.Errorf("circular reference detected: parent cannot be a descendant")
		}
	}

	return nil
}

// isDescendant checks if potentialDescendant is a descendant of ancestor
func (r *CategoryRepository) isDescendant(ctx context.Context, ancestor uint, potentialDescendant uint) (bool, error) {
	var category domain.Category
	if err := r.db.WithContext(ctx).First(&category, potentialDescendant).Error; err != nil {
		return false, err
	}

	if category.ParentID == nil {
		return false, nil
	}

	if *category.ParentID == ancestor {
		return true, nil
	}

	return r.isDescendant(ctx, ancestor, *category.ParentID)
}
