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

type moduleRepository struct {
	db *gorm.DB
}

// NewModuleRepository creates a new module repository
func NewModuleRepository(db *gorm.DB) repositories.ModuleRepository {
	return &moduleRepository{db: db}
}

func (r *moduleRepository) Create(ctx context.Context, module *domain.Module) error {
	if err := r.db.WithContext(ctx).Create(module).Error; err != nil {
		logger.Error("Failed to create module", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to create module", 500)
	}
	return nil
}

func (r *moduleRepository) GetByID(ctx context.Context, id uint) (*domain.Module, error) {
	var module domain.Module
	if err := r.db.WithContext(ctx).
		Preload("Departments").
		First(&module, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrCodeNotFound, "module not found", 404)
		}
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get module", 500)
	}
	return &module, nil
}

func (r *moduleRepository) GetByCode(ctx context.Context, code string) (*domain.Module, error) {
	var module domain.Module
	if err := r.db.WithContext(ctx).
		Where("code = ?", code).
		Preload("Departments").
		First(&module).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrCodeNotFound, "module not found", 404)
		}
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get module", 500)
	}
	return &module, nil
}

func (r *moduleRepository) List(ctx context.Context, params pagination.Params) (*pagination.Result[domain.Module], error) {
	var modules []domain.Module
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Module{})

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to count modules", 500)
	}

	// Apply pagination
	offset := (params.Page - 1) * params.Limit
	if err := query.Order("\"order\" ASC, id ASC").
		Offset(offset).
		Limit(params.Limit).
		Find(&modules).Error; err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to list modules", 500)
	}

	return &pagination.Result[domain.Module]{
		Data:  modules,
		Total: total,
		Page:  params.Page,
		Limit: params.Limit,
	}, nil
}

func (r *moduleRepository) Update(ctx context.Context, module *domain.Module) error {
	if err := r.db.WithContext(ctx).Save(module).Error; err != nil {
		logger.Error("Failed to update module", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to update module", 500)
	}
	return nil
}

func (r *moduleRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Module{}, id).Error; err != nil {
		logger.Error("Failed to delete module", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to delete module", 500)
	}
	return nil
}

func (r *moduleRepository) ListActive(ctx context.Context) ([]domain.Module, error) {
	var modules []domain.Module
	if err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("\"order\" ASC").
		Find(&modules).Error; err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to list active modules", 500)
	}
	return modules, nil
}
