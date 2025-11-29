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

type roleRepository struct {
	db *gorm.DB
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *gorm.DB) repositories.RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) Create(ctx context.Context, role *domain.Role) error {
	if err := r.db.WithContext(ctx).Create(role).Error; err != nil {
		logger.Error("Failed to create role", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to create role", 500)
	}
	return nil
}

func (r *roleRepository) GetByID(ctx context.Context, id uint) (*domain.Role, error) {
	var role domain.Role
	if err := r.db.WithContext(ctx).
		Preload("Permissions").
		Preload("Parent").
		Preload("Children").
		First(&role, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrCodeNotFound, "role not found", 404)
		}
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get role", 500)
	}
	return &role, nil
}

func (r *roleRepository) GetByName(ctx context.Context, name string) (*domain.Role, error) {
	var role domain.Role
	if err := r.db.WithContext(ctx).
		Where("name = ?", name).
		Preload("Permissions").
		First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrCodeNotFound, "role not found", 404)
		}
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get role", 500)
	}
	return &role, nil
}

func (r *roleRepository) Update(ctx context.Context, role *domain.Role) error {
	if err := r.db.WithContext(ctx).Save(role).Error; err != nil {
		logger.Error("Failed to update role", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to update role", 500)
	}
	return nil
}

func (r *roleRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Role{}, id).Error; err != nil {
		logger.Error("Failed to delete role", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to delete role", 500)
	}
	return nil
}

func (r *roleRepository) List(ctx context.Context, filter repositories.RoleFilter) ([]*domain.Role, error) {
	var roles []*domain.Role
	query := r.db.WithContext(ctx).Model(&domain.Role{}).Preload("Permissions")

	if filter.Name != "" {
		query = query.Where("name LIKE ?", "%"+filter.Name+"%")
	}
	if filter.Level != "" {
		query = query.Where("level = ?", filter.Level)
	}
	if filter.ParentID != nil {
		query = query.Where("parent_id = ?", *filter.ParentID)
	}
	if filter.IsSystem != nil {
		query = query.Where("is_system = ?", *filter.IsSystem)
	}

	if err := query.Order("id ASC").Find(&roles).Error; err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to list roles", 500)
	}

	return roles, nil
}

func (r *roleRepository) GetHierarchy(ctx context.Context) ([]*domain.Role, error) {
	var roles []*domain.Role
	// Get root roles (no parent)
	if err := r.db.WithContext(ctx).
		Where("parent_id IS NULL").
		Preload("Children").
		Preload("Children.Children"). // Load 2 levels deep for now
		Find(&roles).Error; err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get role hierarchy", 500)
	}
	return roles, nil
}

func (r *roleRepository) GetChildren(ctx context.Context, parentID uint) ([]*domain.Role, error) {
	var roles []*domain.Role
	if err := r.db.WithContext(ctx).
		Where("parent_id = ?", parentID).
		Find(&roles).Error; err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get child roles", 500)
	}
	return roles, nil
}

func (r *roleRepository) AssignPermission(ctx context.Context, roleID, permissionID uint) error {
	rolePermission := domain.RolePermission{
		RoleID:       roleID,
		PermissionID: permissionID,
	}
	if err := r.db.WithContext(ctx).Create(&rolePermission).Error; err != nil {
		logger.Error("Failed to assign permission to role", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to assign permission", 500)
	}
	return nil
}

func (r *roleRepository) RemovePermission(ctx context.Context, roleID, permissionID uint) error {
	if err := r.db.WithContext(ctx).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&domain.RolePermission{}).Error; err != nil {
		logger.Error("Failed to remove permission from role", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to remove permission", 500)
	}
	return nil
}

func (r *roleRepository) GetRolePermissions(ctx context.Context, roleID uint) ([]*domain.Permission, error) {
	var permissions []*domain.Permission
	if err := r.db.WithContext(ctx).
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&permissions).Error; err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get role permissions", 500)
	}
	return permissions, nil
}

func (r *roleRepository) GetRoleUsers(ctx context.Context, roleID uint) ([]*domain.User, error) {
	var users []*domain.User
	if err := r.db.WithContext(ctx).
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Where("user_roles.role_id = ?", roleID).
		Find(&users).Error; err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get role users", 500)
	}
	return users, nil
}
