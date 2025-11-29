package authorization

import (
	"context"

	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/pkg/errors"
	"github.com/owner/go-cms/pkg/logger"
	"go.uber.org/zap"
)

// PermissionUseCase handles permission business logic
type PermissionUseCase struct {
	permissionRepo repositories.PermissionRepository
}

// NewPermissionUseCase creates a new permission use case
func NewPermissionUseCase(permissionRepo repositories.PermissionRepository) *PermissionUseCase {
	return &PermissionUseCase{
		permissionRepo: permissionRepo,
	}
}

// CreatePermission creates a new permission
func (uc *PermissionUseCase) CreatePermission(ctx context.Context, permission *domain.Permission) error {
	// Validate permission
	if permission.Resource == "" || permission.Action == "" {
		return errors.New(errors.ErrCodeValidation, "resource and action are required", 400)
	}

	// Check if permission with same key already exists
	key := permission.GetPermissionKey()
	existing, err := uc.permissionRepo.GetByKey(ctx, key)
	if err == nil && existing != nil {
		return errors.New(errors.ErrCodeConflict, "permission with this key already exists", 409)
	}

	if err := uc.permissionRepo.Create(ctx, permission); err != nil {
		logger.Error("Failed to create permission", zap.Error(err))
		return err
	}

	logger.Info("Permission created successfully",
		zap.String("resource", permission.Resource),
		zap.String("action", permission.Action),
		zap.Uint("id", permission.ID))
	return nil
}

// GetPermission retrieves a permission by ID
func (uc *PermissionUseCase) GetPermission(ctx context.Context, id uint) (*domain.Permission, error) {
	permission, err := uc.permissionRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return permission, nil
}

// GetPermissionByKey retrieves a permission by its unique key
func (uc *PermissionUseCase) GetPermissionByKey(ctx context.Context, key string) (*domain.Permission, error) {
	permission, err := uc.permissionRepo.GetByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	return permission, nil
}

// ListPermissions retrieves all permissions with optional filters
func (uc *PermissionUseCase) ListPermissions(ctx context.Context, filter repositories.PermissionFilter) ([]*domain.Permission, error) {
	permissions, err := uc.permissionRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

// GetPermissionsByModule retrieves permissions by module
func (uc *PermissionUseCase) GetPermissionsByModule(ctx context.Context, module string) ([]*domain.Permission, error) {
	permissions, err := uc.permissionRepo.GetByModule(ctx, module)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

// GetPermissionsByDepartment retrieves permissions by department
func (uc *PermissionUseCase) GetPermissionsByDepartment(ctx context.Context, department string) ([]*domain.Permission, error) {
	permissions, err := uc.permissionRepo.GetByDepartment(ctx, department)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

// GetPermissionsByService retrieves permissions by service
func (uc *PermissionUseCase) GetPermissionsByService(ctx context.Context, service string) ([]*domain.Permission, error) {
	permissions, err := uc.permissionRepo.GetByService(ctx, service)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

// UpdatePermission updates an existing permission
func (uc *PermissionUseCase) UpdatePermission(ctx context.Context, id uint, updates *domain.Permission) error {
	// Get existing permission
	existing, err := uc.permissionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Update fields
	if updates.Description != "" {
		existing.Description = updates.Description
	}
	if updates.Module != "" {
		existing.Module = updates.Module
	}
	if updates.Department != "" {
		existing.Department = updates.Department
	}
	if updates.Service != "" {
		existing.Service = updates.Service
	}

	if err := uc.permissionRepo.Update(ctx, existing); err != nil {
		logger.Error("Failed to update permission", zap.Error(err))
		return err
	}

	logger.Info("Permission updated successfully", zap.Uint("id", id))
	return nil
}

// DeletePermission deletes a permission
func (uc *PermissionUseCase) DeletePermission(ctx context.Context, id uint) error {
	// Get existing permission
	_, err := uc.permissionRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := uc.permissionRepo.Delete(ctx, id); err != nil {
		logger.Error("Failed to delete permission", zap.Error(err))
		return err
	}

	logger.Info("Permission deleted successfully", zap.Uint("id", id))
	return nil
}
