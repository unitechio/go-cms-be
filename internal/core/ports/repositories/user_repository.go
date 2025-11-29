package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/pkg/pagination"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Basic CRUD
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uuid.UUID) error

	// List operations
	List(ctx context.Context, filter UserFilter, page *pagination.OffsetPagination) ([]*domain.User, int64, error)
	ListWithCursor(ctx context.Context, filter UserFilter, cursor *pagination.Cursor, limit int) ([]*domain.User, *pagination.Cursor, error)

	// Role operations
	AssignRole(ctx context.Context, userID uuid.UUID, roleID uint) error
	RemoveRole(ctx context.Context, userID uuid.UUID, roleID uint) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error)

	// Permission operations
	AssignPermission(ctx context.Context, userID uuid.UUID, permissionID uint) error
	RemovePermission(ctx context.Context, userID uuid.UUID, permissionID uint) error
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*domain.Permission, error)

	// Authentication
	UpdateLastLogin(ctx context.Context, userID uuid.UUID, ip string) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error
	Enable2FA(ctx context.Context, userID uuid.UUID, secret string) error
	Disable2FA(ctx context.Context, userID uuid.UUID) error

	// Status
	UpdateStatus(ctx context.Context, userID uuid.UUID, status domain.UserStatus) error
	VerifyEmail(ctx context.Context, userID uuid.UUID) error
}

// UserFilter represents filters for user queries
type UserFilter struct {
	Email  string
	Status domain.UserStatus
	Search string // Search in email, first_name, last_name
	RoleID *uint
	IDs    []uuid.UUID
}

// CustomerRepository defines the interface for customer data operations
type CustomerRepository interface {
	// Basic CRUD
	Create(ctx context.Context, customer *domain.Customer) error
	GetByID(ctx context.Context, id uint) (*domain.Customer, error)
	GetByEmail(ctx context.Context, email string) (*domain.Customer, error)
	Update(ctx context.Context, customer *domain.Customer) error
	Delete(ctx context.Context, id uint) error

	// List operations
	List(ctx context.Context, filter CustomerFilter, page *pagination.OffsetPagination) ([]*domain.Customer, int64, error)
	ListWithCursor(ctx context.Context, filter CustomerFilter, cursor *pagination.Cursor, limit int) ([]*domain.Customer, *pagination.Cursor, error)

	// Assignment
	AssignToUser(ctx context.Context, customerID uint, userID uuid.UUID) error
	UnassignFromUser(ctx context.Context, customerID uint) error
	GetByAssignedUser(ctx context.Context, userID uuid.UUID) ([]*domain.Customer, error)

	// Status
	UpdateStatus(ctx context.Context, customerID uint, status domain.UserStatus) error
}

// CustomerFilter represents filters for customer queries
type CustomerFilter struct {
	Email      string
	Phone      string
	Status     domain.UserStatus
	Search     string // Search in email, first_name, last_name, company
	AssignedTo *uuid.UUID
	Source     string
	IDs        []uint
}

// RoleRepository defines the interface for role data operations
type RoleRepository interface {
	// Basic CRUD
	Create(ctx context.Context, role *domain.Role) error
	GetByID(ctx context.Context, id uint) (*domain.Role, error)
	GetByName(ctx context.Context, name string) (*domain.Role, error)
	Update(ctx context.Context, role *domain.Role) error
	Delete(ctx context.Context, id uint) error

	// List operations
	List(ctx context.Context, filter RoleFilter) ([]*domain.Role, error)
	GetHierarchy(ctx context.Context) ([]*domain.Role, error)
	GetChildren(ctx context.Context, parentID uint) ([]*domain.Role, error)

	// Permission operations
	AssignPermission(ctx context.Context, roleID, permissionID uint) error
	RemovePermission(ctx context.Context, roleID, permissionID uint) error
	GetRolePermissions(ctx context.Context, roleID uint) ([]*domain.Permission, error)

	// User operations
	GetRoleUsers(ctx context.Context, roleID uint) ([]*domain.User, error)
}

// RoleFilter represents filters for role queries
type RoleFilter struct {
	Name     string
	Level    domain.RoleLevel
	ParentID *uint
	IsSystem *bool
}

// PermissionRepository defines the interface for permission data operations
type PermissionRepository interface {
	// Basic CRUD
	Create(ctx context.Context, permission *domain.Permission) error
	GetByID(ctx context.Context, id uint) (*domain.Permission, error)
	Update(ctx context.Context, permission *domain.Permission) error
	Delete(ctx context.Context, id uint) error

	// List operations
	List(ctx context.Context, filter PermissionFilter) ([]*domain.Permission, error)
	GetByKey(ctx context.Context, key string) (*domain.Permission, error)

	// Module/Department/Service operations
	GetByModule(ctx context.Context, module string) ([]*domain.Permission, error)
	GetByDepartment(ctx context.Context, department string) ([]*domain.Permission, error)
	GetByService(ctx context.Context, service string) ([]*domain.Permission, error)
}

// PermissionFilter represents filters for permission queries
type PermissionFilter struct {
	Module     string
	Department string
	Service    string
	Resource   string
	Action     string
}

// OTPRepository defines the interface for OTP data operations
type OTPRepository interface {
	Create(ctx context.Context, otp *domain.OTP) error
	GetByEmail(ctx context.Context, email, otpType string) (*domain.OTP, error)
	MarkAsUsed(ctx context.Context, id uint) error
	DeleteExpired(ctx context.Context) error
	DeleteByEmail(ctx context.Context, email, otpType string) error
}

// RefreshTokenRepository defines the interface for refresh token data operations
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.RefreshToken, error)
	Revoke(ctx context.Context, token string) error
	RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}
