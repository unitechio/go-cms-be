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

type departmentRepository struct {
	db *gorm.DB
}

// NewDepartmentRepository creates a new department repository
func NewDepartmentRepository(db *gorm.DB) repositories.DepartmentRepository {
	return &departmentRepository{db: db}
}

func (r *departmentRepository) Create(ctx context.Context, department *domain.Department) error {
	if err := r.db.WithContext(ctx).Create(department).Error; err != nil {
		logger.Error("Failed to create department", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to create department", 500)
	}
	return nil
}

func (r *departmentRepository) GetByID(ctx context.Context, id uint) (*domain.Department, error) {
	var department domain.Department
	if err := r.db.WithContext(ctx).
		Preload("Module").
		Preload("Parent").
		Preload("Children").
		Preload("Services").
		First(&department, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrCodeNotFound, "department not found", 404)
		}
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get department", 500)
	}
	return &department, nil
}

func (r *departmentRepository) GetByCode(ctx context.Context, code string) (*domain.Department, error) {
	var department domain.Department
	if err := r.db.WithContext(ctx).
		Where("code = ?", code).
		Preload("Module").
		First(&department).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrCodeNotFound, "department not found", 404)
		}
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get department", 500)
	}
	return &department, nil
}

func (r *departmentRepository) List(ctx context.Context, params pagination.Params) (*pagination.Result[domain.Department], error) {
	var departments []domain.Department
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Department{}).Preload("Module")

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to count departments", 500)
	}

	// Apply pagination
	offset := (params.Page - 1) * params.Limit
	if err := query.Order("id ASC").
		Offset(offset).
		Limit(params.Limit).
		Find(&departments).Error; err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to list departments", 500)
	}

	return &pagination.Result[domain.Department]{
		Data:  departments,
		Total: total,
		Page:  params.Page,
		Limit: params.Limit,
	}, nil
}

func (r *departmentRepository) ListByModule(ctx context.Context, moduleID uint) ([]domain.Department, error) {
	var departments []domain.Department
	if err := r.db.WithContext(ctx).
		Where("module_id = ?", moduleID).
		Find(&departments).Error; err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to list departments by module", 500)
	}
	return departments, nil
}

func (r *departmentRepository) Update(ctx context.Context, department *domain.Department) error {
	if err := r.db.WithContext(ctx).Save(department).Error; err != nil {
		logger.Error("Failed to update department", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to update department", 500)
	}
	return nil
}

func (r *departmentRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Department{}, id).Error; err != nil {
		logger.Error("Failed to delete department", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to delete department", 500)
	}
	return nil
}

func (r *departmentRepository) ListActive(ctx context.Context) ([]domain.Department, error) {
	var departments []domain.Department
	if err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Preload("Module").
		Find(&departments).Error; err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to list active departments", 500)
	}
	return departments, nil
}
