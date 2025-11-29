package repositories

import (
	"context"

	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/pkg/pagination"
)

// ModuleRepository defines the interface for module data access
type ModuleRepository interface {
	Create(ctx context.Context, module *domain.Module) error
	GetByID(ctx context.Context, id uint) (*domain.Module, error)
	GetByCode(ctx context.Context, code string) (*domain.Module, error)
	List(ctx context.Context, params pagination.Params) (*pagination.Result[domain.Module], error)
	Update(ctx context.Context, module *domain.Module) error
	Delete(ctx context.Context, id uint) error
	ListActive(ctx context.Context) ([]domain.Module, error)
}

// DepartmentRepository defines the interface for department data access
type DepartmentRepository interface {
	Create(ctx context.Context, department *domain.Department) error
	GetByID(ctx context.Context, id uint) (*domain.Department, error)
	GetByCode(ctx context.Context, code string) (*domain.Department, error)
	List(ctx context.Context, params pagination.Params) (*pagination.Result[domain.Department], error)
	ListByModule(ctx context.Context, moduleID uint) ([]domain.Department, error)
	Update(ctx context.Context, department *domain.Department) error
	Delete(ctx context.Context, id uint) error
	ListActive(ctx context.Context) ([]domain.Department, error)
}

// ServiceRepository defines the interface for service data access
type ServiceRepository interface {
	Create(ctx context.Context, service *domain.Service) error
	GetByID(ctx context.Context, id uint) (*domain.Service, error)
	GetByCode(ctx context.Context, code string) (*domain.Service, error)
	List(ctx context.Context, params pagination.Params) (*pagination.Result[domain.Service], error)
	ListByDepartment(ctx context.Context, departmentID uint) ([]domain.Service, error)
	Update(ctx context.Context, service *domain.Service) error
	Delete(ctx context.Context, id uint) error
	ListActive(ctx context.Context) ([]domain.Service, error)
}

// ScopeRepository defines the interface for scope data access
type ScopeRepository interface {
	Create(ctx context.Context, scope *domain.Scope) error
	GetByID(ctx context.Context, id uint) (*domain.Scope, error)
	GetByCode(ctx context.Context, code string) (*domain.Scope, error)
	List(ctx context.Context, params pagination.Params) (*pagination.Result[domain.Scope], error)
	Update(ctx context.Context, scope *domain.Scope) error
	Delete(ctx context.Context, id uint) error
	ListAll(ctx context.Context) ([]domain.Scope, error)
}

// EnhancedPermissionRepository defines the interface for enhanced permission data access
type EnhancedPermissionRepository interface {
	Create(ctx context.Context, permission *domain.EnhancedPermission) error
	GetByID(ctx context.Context, id uint) (*domain.EnhancedPermission, error)
	GetByCode(ctx context.Context, code string) (*domain.EnhancedPermission, error)
	List(ctx context.Context, params pagination.Params) (*pagination.Result[domain.EnhancedPermission], error)
	ListByModule(ctx context.Context, moduleID uint) ([]domain.EnhancedPermission, error)
	ListByDepartment(ctx context.Context, departmentID uint) ([]domain.EnhancedPermission, error)
	ListByService(ctx context.Context, serviceID uint) ([]domain.EnhancedPermission, error)
	ListByScope(ctx context.Context, scopeID uint) ([]domain.EnhancedPermission, error)
	Update(ctx context.Context, permission *domain.EnhancedPermission) error
	Delete(ctx context.Context, id uint) error

	// Role-Permission relationships
	AssignToRole(ctx context.Context, roleID uint, permissionIDs []uint) error
	RemoveFromRole(ctx context.Context, roleID uint, permissionIDs []uint) error
	GetRolePermissions(ctx context.Context, roleID uint) ([]domain.EnhancedPermission, error)

	// User-Permission relationships (direct grants)
	GrantToUser(ctx context.Context, userID uint, permissionID uint, grantedBy uint) error
	RevokeFromUser(ctx context.Context, userID uint, permissionID uint, revokedBy uint) error
	GetUserDirectPermissions(ctx context.Context, userID uint) ([]domain.EnhancedPermission, error)
	GetUserAllPermissions(ctx context.Context, userID uint) ([]domain.EnhancedPermission, error) // From roles + direct
}
