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

type serviceRepository struct {
	db *gorm.DB
}

// NewServiceRepository creates a new service repository
func NewServiceRepository(db *gorm.DB) repositories.ServiceRepository {
	return &serviceRepository{db: db}
}

func (r *serviceRepository) Create(ctx context.Context, service *domain.Service) error {
	if err := r.db.WithContext(ctx).Create(service).Error; err != nil {
		logger.Error("Failed to create service", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to create service", 500)
	}
	return nil
}

func (r *serviceRepository) GetByID(ctx context.Context, id uint) (*domain.Service, error) {
	var service domain.Service
	if err := r.db.WithContext(ctx).
		Preload("Department").
		Preload("Department.Module").
		First(&service, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrCodeNotFound, "service not found", 404)
		}
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get service", 500)
	}
	return &service, nil
}

func (r *serviceRepository) GetByCode(ctx context.Context, code string) (*domain.Service, error) {
	var service domain.Service
	if err := r.db.WithContext(ctx).
		Where("code = ?", code).
		Preload("Department").
		First(&service).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrCodeNotFound, "service not found", 404)
		}
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get service", 500)
	}
	return &service, nil
}

func (r *serviceRepository) List(ctx context.Context, params pagination.Params) (*pagination.Result[domain.Service], error) {
	var services []domain.Service
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Service{}).
		Preload("Department").
		Preload("Department.Module")

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to count services", 500)
	}

	// Apply pagination
	offset := (params.Page - 1) * params.Limit
	if err := query.Order("id ASC").
		Offset(offset).
		Limit(params.Limit).
		Find(&services).Error; err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to list services", 500)
	}

	return &pagination.Result[domain.Service]{
		Data:  services,
		Total: total,
		Page:  params.Page,
		Limit: params.Limit,
	}, nil
}

func (r *serviceRepository) ListByDepartment(ctx context.Context, departmentID uint) ([]domain.Service, error) {
	var services []domain.Service
	if err := r.db.WithContext(ctx).
		Where("department_id = ?", departmentID).
		Find(&services).Error; err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to list services by department", 500)
	}
	return services, nil
}

func (r *serviceRepository) Update(ctx context.Context, service *domain.Service) error {
	if err := r.db.WithContext(ctx).Save(service).Error; err != nil {
		logger.Error("Failed to update service", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to update service", 500)
	}
	return nil
}

func (r *serviceRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Service{}, id).Error; err != nil {
		logger.Error("Failed to delete service", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to delete service", 500)
	}
	return nil
}

func (r *serviceRepository) ListActive(ctx context.Context) ([]domain.Service, error) {
	var services []domain.Service
	if err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Preload("Department").
		Find(&services).Error; err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to list active services", 500)
	}
	return services, nil
}
