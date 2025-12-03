package authorization

import (
	"context"

	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/pkg/errors"
	"github.com/owner/go-cms/pkg/logger"
	"go.uber.org/zap"
)

// RoleUseCase handles role business logic
type RoleUseCase struct {
	roleRepo       repositories.RoleRepository
	permissionRepo repositories.PermissionRepository
}

// NewRoleUseCase creates a new role use case
func NewRoleUseCase(roleRepo repositories.RoleRepository, permissionRepo repositories.PermissionRepository) *RoleUseCase {
	return &RoleUseCase{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
	}
}

// CreateRole creates a new role
func (uc *RoleUseCase) CreateRole(ctx context.Context, role *domain.Role) error {
	// Validate role
	if role.Name == "" {
		return errors.New(errors.ErrCodeValidation, "role name is required", 400)
	}

	// Check if role with same name already exists
	existing, err := uc.roleRepo.GetByName(ctx, role.Name)
	if err == nil && existing != nil {
		return errors.New(errors.ErrCodeConflict, "role with this name already exists", 409)
	}

	// Validate parent role if specified
	if role.ParentID != nil {
		parent, err := uc.roleRepo.GetByID(ctx, *role.ParentID)
		if err != nil {
			return errors.Wrap(err, errors.ErrCodeNotFound, "parent role not found", 404)
		}
		// Ensure parent exists
		if parent == nil {
			return errors.New(errors.ErrCodeNotFound, "parent role not found", 404)
		}
	}

	if err := uc.roleRepo.Create(ctx, role); err != nil {
		logger.Error("Failed to create role", zap.Error(err))
		return err
	}

	logger.Info("Role created successfully", zap.String("name", role.Name), zap.Uint("id", role.ID))
	return nil
}

// GetRole retrieves a role by ID
func (uc *RoleUseCase) GetRole(ctx context.Context, id uint) (*domain.Role, error) {
	role, err := uc.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return role, nil
}

// GetRoleByName retrieves a role by name
func (uc *RoleUseCase) GetRoleByName(ctx context.Context, name string) (*domain.Role, error) {
	role, err := uc.roleRepo.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return role, nil
}

// ListRoles retrieves all roles with optional filters
func (uc *RoleUseCase) ListRoles(ctx context.Context, filter repositories.RoleFilter) ([]*domain.Role, error) {
	roles, err := uc.roleRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// GetRoleHierarchy retrieves the role hierarchy
func (uc *RoleUseCase) GetRoleHierarchy(ctx context.Context) ([]*domain.Role, error) {
	roles, err := uc.roleRepo.GetHierarchy(ctx)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// UpdateRole updates an existing role
func (uc *RoleUseCase) UpdateRole(ctx context.Context, id uint, updates *domain.Role) error {
	// Get existing role
	existing, err := uc.roleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if it's a system role
	if existing.IsSystem {
		return errors.New(errors.ErrCodeForbidden, "cannot modify system role", 403)
	}

	// Update fields
	if updates.Name != "" {
		existing.Name = updates.Name
	}
	if updates.DisplayName != "" {
		existing.DisplayName = updates.DisplayName
	}
	if updates.Description != "" {
		existing.Description = updates.Description
	}
	if updates.Level != "" {
		existing.Level = updates.Level
	}

	if err := uc.roleRepo.Update(ctx, existing); err != nil {
		logger.Error("Failed to update role", zap.Error(err))
		return err
	}

	logger.Info("Role updated successfully", zap.Uint("id", id))
	return nil
}

// DeleteRole deletes a role
func (uc *RoleUseCase) DeleteRole(ctx context.Context, id uint) error {
	// Get existing role
	role, err := uc.roleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if it's a system role
	if role.IsSystem {
		return errors.New(errors.ErrCodeForbidden, "cannot delete system role", 403)
	}

	// Check if role has children
	children, err := uc.roleRepo.GetChildren(ctx, id)
	if err != nil {
		return err
	}
	if len(children) > 0 {
		return errors.New(errors.ErrCodeConflict, "cannot delete role with child roles", 409)
	}

	if err := uc.roleRepo.Delete(ctx, id); err != nil {
		logger.Error("Failed to delete role", zap.Error(err))
		return err
	}

	logger.Info("Role deleted successfully", zap.Uint("id", id))
	return nil
}

// GetRolePermissions retrieves all permissions for a role
func (uc *RoleUseCase) GetRolePermissions(ctx context.Context, roleID uint) ([]*domain.Permission, error) {
	permissions, err := uc.roleRepo.GetRolePermissions(ctx, roleID)
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

func (uc *RoleUseCase) AssignPermission(ctx context.Context, roleID, permissionID uint) error {
	_, err := uc.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return err
	}

	_, err = uc.permissionRepo.GetByID(ctx, permissionID)
	if err != nil {
		return err
	}

	if err := uc.roleRepo.AssignPermission(ctx, roleID, permissionID); err != nil {
		logger.Error("Failed to assign permission to role", zap.Error(err))
		return err
	}

	logger.Info("Permission assigned to role", zap.Uint("roleID", roleID), zap.Uint("permissionID", permissionID))
	return nil
}

// RemovePermission removes a permission from a role
func (uc *RoleUseCase) RemovePermission(ctx context.Context, roleID, permissionID uint) error {
	if err := uc.roleRepo.RemovePermission(ctx, roleID, permissionID); err != nil {
		logger.Error("Failed to remove permission from role", zap.Error(err))
		return err
	}

	logger.Info("Permission removed from role", zap.Uint("roleID", roleID), zap.Uint("permissionID", permissionID))
	return nil
}

// GetRoleUsers retrieves all users with a specific role
func (uc *RoleUseCase) GetRoleUsers(ctx context.Context, roleID uint) ([]*domain.User, error) {
	users, err := uc.roleRepo.GetRoleUsers(ctx, roleID)
	if err != nil {
		return nil, err
	}
	return users, nil
}
