package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel contains common fields for all models with sequence-based ID
type BaseModel struct {
	ID        uint           `gorm:"primarykey;autoIncrement" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate hook to set ID from sequence
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == 0 {
		// ID will be set by database sequence trigger
		// We don't set it here, let PostgreSQL handle it
	}
	return nil
}

// UUIDModel contains common fields for models with UUID primary key
type UUIDModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primarykey;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate hook to generate UUID if not set
func (u *UUIDModel) BeforeCreate(tx *gorm.DB) error {
	if u.ID == (uuid.UUID{}) {
		u.ID = uuid.New()
	}
	return nil
}

// UserStatus represents user account status
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusPending   UserStatus = "pending"
)

// User represents a system user with UUID
type User struct {
	UUIDModel
	Email            string     `gorm:"uniqueIndex;not null" json:"email"`
	Password         string     `gorm:"not null" json:"-"`
	FirstName        string     `gorm:"size:100" json:"first_name"`
	LastName         string     `gorm:"size:100" json:"last_name"`
	Phone            string     `gorm:"size:20" json:"phone"`
	Avatar           string     `json:"avatar"`
	DepartmentID     *uint      `gorm:"index" json:"department_id,omitempty"` // Foreign key to departments table
	Position         string     `gorm:"size:100" json:"position"`             // Job title/position
	Status           UserStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	EmailVerified    bool       `gorm:"default:false" json:"email_verified"`
	EmailVerifiedAt  *time.Time `json:"email_verified_at,omitempty"`
	TwoFactorEnabled bool       `gorm:"default:false" json:"two_factor_enabled"`
	TwoFactorSecret  string     `json:"-"`
	LastLoginAt      *time.Time `json:"last_login_at,omitempty"`
	LastLoginIP      string     `gorm:"size:45" json:"last_login_ip,omitempty"`

	// Relationships
	Department  *Department  `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	Roles       []Role       `gorm:"many2many:user_roles;" json:"roles,omitempty"`
	Permissions []Permission `gorm:"many2many:user_permissions;" json:"permissions,omitempty"`
	Posts       []Post       `gorm:"foreignKey:AuthorID" json:"posts,omitempty"`
	Media       []Media      `gorm:"foreignKey:UploadedBy" json:"media,omitempty"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}

// Customer represents a customer in the CRM
type Customer struct {
	BaseModel
	Email      string     `gorm:"uniqueIndex;not null" json:"email"`
	FirstName  string     `gorm:"size:100;not null" json:"first_name"`
	LastName   string     `gorm:"size:100;not null" json:"last_name"`
	Phone      string     `gorm:"size:20" json:"phone"`
	Company    string     `gorm:"size:200" json:"company"`
	Address    string     `json:"address"`
	City       string     `gorm:"size:100" json:"city"`
	State      string     `gorm:"size:100" json:"state"`
	Country    string     `gorm:"size:100" json:"country"`
	PostalCode string     `gorm:"size:20" json:"postal_code"`
	Status     UserStatus `gorm:"type:varchar(20);default:'active'" json:"status"`
	Notes      string     `gorm:"type:text" json:"notes"`
	Tags       string     `json:"tags"`                                   // JSON array of tags
	Source     string     `gorm:"size:100" json:"source"`                 // Where the customer came from
	AssignedTo *uuid.UUID `gorm:"type:uuid" json:"assigned_to,omitempty"` // Assigned user UUID

	// Relationships
	AssignedUser *User `gorm:"foreignKey:AssignedTo;references:ID" json:"assigned_user,omitempty"`
}

// TableName specifies the table name for Customer
func (Customer) TableName() string {
	return "customers"
}

// RoleLevel represents the hierarchical level of a role
type RoleLevel string

const (
	RoleLevelOrganization RoleLevel = "organization"
	RoleLevelDepartment   RoleLevel = "department"
	RoleLevelService      RoleLevel = "service"
	RoleLevelAction       RoleLevel = "action"
)

// Role represents a user role with hierarchical structure
type Role struct {
	BaseModel
	Name        string    `gorm:"uniqueIndex;not null" json:"name"`
	DisplayName string    `gorm:"size:200" json:"display_name"`
	Description string    `gorm:"type:text" json:"description"`
	Level       RoleLevel `gorm:"type:varchar(20);not null" json:"level"`
	ParentID    *uint     `json:"parent_id,omitempty"`
	IsSystem    bool      `gorm:"default:false" json:"is_system"` // System roles cannot be deleted

	// Relationships
	Parent      *Role        `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children    []Role       `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
	Users       []User       `gorm:"many2many:user_roles;" json:"users,omitempty"`
}

// TableName specifies the table name for Role
func (Role) TableName() string {
	return "roles"
}

// Permission represents a granular permission
type Permission struct {
	BaseModel
	Resource    string `gorm:"size:100;not null" json:"resource"` // e.g., "users", "posts", "customers"
	Action      string `gorm:"size:50;not null" json:"action"`    // e.g., "create", "read", "update", "delete"
	Description string `gorm:"type:text" json:"description"`
	Module      string `gorm:"size:100" json:"module"`     // e.g., "crm", "content", "admin"
	Department  string `gorm:"size:100" json:"department"` // Department level
	Service     string `gorm:"size:100" json:"service"`    // Service level within department

	// Relationships
	Roles []Role `gorm:"many2many:role_permissions;" json:"roles,omitempty"`
	Users []User `gorm:"many2many:user_permissions;" json:"users,omitempty"`
}

// TableName specifies the table name for Permission
func (Permission) TableName() string {
	return "permissions"
}

// GetPermissionKey returns a unique key for the permission
func (p *Permission) GetPermissionKey() string {
	return p.Module + ":" + p.Department + ":" + p.Service + ":" + p.Resource + ":" + p.Action
}

// UserRole represents the many-to-many relationship between users and roles
type UserRole struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	RoleID    uint      `gorm:"primaryKey" json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
	DeletedAt time.Time `json:"deleted_at"`
}

// TableName specifies the table name for UserRole
func (UserRole) TableName() string {
	return "user_roles"
}

// RolePermission represents the many-to-many relationship between roles and permissions
type RolePermission struct {
	RoleID       uint      `gorm:"primaryKey" json:"role_id"`
	PermissionID uint      `gorm:"primaryKey" json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// TableName specifies the table name for RolePermission
func (RolePermission) TableName() string {
	return "role_permissions"
}

// UserPermission represents direct user permissions (override role permissions)
type UserPermission struct {
	UserID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	PermissionID uint      `gorm:"primaryKey" json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// TableName specifies the table name for UserPermission
func (UserPermission) TableName() string {
	return "user_permissions"
}

// OTP represents a one-time password
type OTP struct {
	BaseModel
	Email     string     `gorm:"index;not null" json:"email"`
	Code      string     `gorm:"not null" json:"-"`
	Type      string     `gorm:"size:50;not null" json:"type"` // e.g., "email_verification", "password_reset"
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	Used      bool       `gorm:"default:false" json:"used"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
}

// TableName specifies the table name for OTP
func (OTP) TableName() string {
	return "otps"
}

// IsExpired checks if the OTP is expired
func (o *OTP) IsExpired() bool {
	return time.Now().After(o.ExpiresAt)
}

// RefreshToken represents a refresh token for JWT
type RefreshToken struct {
	BaseModel
	UserID    uuid.UUID  `gorm:"type:uuid;index;not null" json:"user_id"`
	Token     string     `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time  `gorm:"not null" json:"expires_at"`
	Revoked   bool       `gorm:"default:false" json:"revoked"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`

	// Relationships
	User User `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

// TableName specifies the table name for RefreshToken
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// IsExpired checks if the refresh token is expired
func (r *RefreshToken) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}
