package postgres

import (
	"context"

	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/pkg/errors"
	"github.com/owner/go-cms/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type permissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(db *gorm.DB) repositories.PermissionRepository {
	return &permissionRepository{db: db}
}

// Create creates a new permission
func (r *permissionRepository) Create(ctx context.Context, permission *domain.Permission) error {
	if err := r.db.WithContext(ctx).Create(permission).Error; err != nil {
		logger.Error("Failed to create permission", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to create permission", 500)
	}
	return nil
}

// GetByID retrieves a permission by ID
func (r *permissionRepository) GetByID(ctx context.Context, id uint) (*domain.Permission, error) {
	var permission domain.Permission
	err := r.db.WithContext(ctx).
		Preload("Roles").
		Preload("Users").
		First(&permission, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrCodeNotFound, "permission not found", 404)
		}
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get permission", 500)
	}
	return &permission, nil
}

// GetByKey retrieves a permission by its unique key
func (r *permissionRepository) GetByKey(ctx context.Context, key string) (*domain.Permission, error) {
	var permission domain.Permission
	err := r.db.WithContext(ctx).
		Where("CONCAT(module, ':', department, ':', service, ':', resource, ':', action) = ?", key).
		First(&permission).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrCodeNotFound, "permission not found", 404)
		}
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get permission", 500)
	}
	return &permission, nil
}

// List retrieves all permissions with filters
func (r *permissionRepository) List(ctx context.Context, filter repositories.PermissionFilter) ([]*domain.Permission, error) {
	var permissions []*domain.Permission
	query := r.db.WithContext(ctx).Model(&domain.Permission{})

	if filter.Module != "" {
		query = query.Where("module = ?", filter.Module)
	}
	if filter.Department != "" {
		query = query.Where("department = ?", filter.Department)
	}
	if filter.Service != "" {
		query = query.Where("service = ?", filter.Service)
	}
	if filter.Resource != "" {
		query = query.Where("resource = ?", filter.Resource)
	}
	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}

	if err := query.Order("module, department, service, resource, action").Find(&permissions).Error; err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to list permissions", 500)
	}

	return permissions, nil
}

// GetByModule retrieves permissions by module
func (r *permissionRepository) GetByModule(ctx context.Context, module string) ([]*domain.Permission, error) {
	var permissions []*domain.Permission
	err := r.db.WithContext(ctx).
		Where("module = ?", module).
		Order("resource, action").
		Find(&permissions).Error
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get permissions by module", 500)
	}
	return permissions, nil
}

// GetByDepartment retrieves permissions by department
func (r *permissionRepository) GetByDepartment(ctx context.Context, department string) ([]*domain.Permission, error) {
	var permissions []*domain.Permission
	err := r.db.WithContext(ctx).
		Where("department = ?", department).
		Order("service, resource, action").
		Find(&permissions).Error
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get permissions by department", 500)
	}
	return permissions, nil
}

// GetByService retrieves permissions by service
func (r *permissionRepository) GetByService(ctx context.Context, service string) ([]*domain.Permission, error) {
	var permissions []*domain.Permission
	err := r.db.WithContext(ctx).
		Where("service = ?", service).
		Order("resource, action").
		Find(&permissions).Error
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get permissions by service", 500)
	}
	return permissions, nil
}

// Update updates an existing permission
func (r *permissionRepository) Update(ctx context.Context, permission *domain.Permission) error {
	if err := r.db.WithContext(ctx).Save(permission).Error; err != nil {
		logger.Error("Failed to update permission", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to update permission", 500)
	}
	return nil
}

// Delete soft deletes a permission
func (r *permissionRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Permission{}, id).Error; err != nil {
		logger.Error("Failed to delete permission", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to delete permission", 500)
	}
	return nil
}
