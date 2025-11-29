package domain

import (
	"time"

	"gorm.io/gorm"
)

// Module represents a system module/feature area
type Module struct {
	BaseModel
	Code        string `gorm:"uniqueIndex;size:50;not null" json:"code"` // e.g., "crm", "content", "admin"
	Name        string `gorm:"size:100;not null" json:"name"`
	DisplayName string `gorm:"size:200" json:"display_name"`
	Description string `gorm:"type:text" json:"description"`
	Icon        string `gorm:"size:100" json:"icon"`
	Color       string `gorm:"size:20" json:"color"`
	Order       int    `gorm:"default:0" json:"order"`
	IsActive    bool   `gorm:"default:true" json:"is_active"`
	IsSystem    bool   `gorm:"default:false" json:"is_system"` // System modules cannot be deleted

	// Relationships
	Departments []Department `gorm:"foreignKey:ModuleID" json:"departments,omitempty"`
	// Permissions []Permission `gorm:"foreignKey:ModuleID" json:"permissions,omitempty"`
}

// TableName specifies the table name for Module
func (Module) TableName() string {
	return "modules"
}

// Department represents a department/division within an organization
type Department struct {
	BaseModel
	ModuleID    uint   `gorm:"index;not null" json:"module_id"`
	Code        string `gorm:"uniqueIndex;size:50;not null" json:"code"` // e.g., "sales", "editorial", "it"
	Name        string `gorm:"size:100;not null" json:"name"`
	DisplayName string `gorm:"size:200" json:"display_name"`
	Description string `gorm:"type:text" json:"description"`
	ParentID    *uint  `gorm:"index" json:"parent_id,omitempty"`  // For hierarchical departments
	ManagerID   *uint  `gorm:"index" json:"manager_id,omitempty"` // Department manager (User ID would be UUID in real case)
	IsActive    bool   `gorm:"default:true" json:"is_active"`
	IsSystem    bool   `gorm:"default:false" json:"is_system"`

	// Relationships
	Module   Module       `gorm:"foreignKey:ModuleID" json:"module,omitempty"`
	Parent   *Department  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []Department `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Services []Service    `gorm:"foreignKey:DepartmentID" json:"services,omitempty"`
}

// TableName specifies the table name for Department
func (Department) TableName() string {
	return "departments"
}

// Service represents a specific service/functionality within a department
type Service struct {
	BaseModel
	DepartmentID uint   `gorm:"index;not null" json:"department_id"`
	Code         string `gorm:"uniqueIndex;size:50;not null" json:"code"` // e.g., "user_management", "post_management"
	Name         string `gorm:"size:100;not null" json:"name"`
	DisplayName  string `gorm:"size:200" json:"display_name"`
	Description  string `gorm:"type:text" json:"description"`
	Endpoint     string `gorm:"size:200" json:"endpoint"` // API endpoint prefix
	IsActive     bool   `gorm:"default:true" json:"is_active"`
	IsSystem     bool   `gorm:"default:false" json:"is_system"`

	// Relationships
	Department Department `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	// Permissions []Permission `gorm:"foreignKey:ServiceID" json:"permissions,omitempty"`
}

// TableName specifies the table name for Service
func (Service) TableName() string {
	return "services"
}

// ScopeLevel represents the scope/level at which a permission applies
type ScopeLevel string

const (
	ScopeLevelOrganization ScopeLevel = "organization" // Entire organization
	ScopeLevelDepartment   ScopeLevel = "department"   // Department level
	ScopeLevelTeam         ScopeLevel = "team"         // Team level
	ScopeLevelPersonal     ScopeLevel = "personal"     // Personal/own resources only
)

// Scope represents permission scope configuration
type Scope struct {
	BaseModel
	Code        string     `gorm:"uniqueIndex;size:50;not null" json:"code"`
	Name        string     `gorm:"size:100;not null" json:"name"`
	DisplayName string     `gorm:"size:200" json:"display_name"`
	Description string     `gorm:"type:text" json:"description"`
	Level       ScopeLevel `gorm:"type:varchar(20);not null" json:"level"`
	Priority    int        `gorm:"default:0" json:"priority"` // Higher priority = broader scope
	IsSystem    bool       `gorm:"default:false" json:"is_system"`

	// Relationships
	// Permissions []Permission `gorm:"foreignKey:ScopeID" json:"permissions,omitempty"`
}

// TableName specifies the table name for Scope
func (Scope) TableName() string {
	return "scopes"
}

// PermissionAction represents available actions
type PermissionAction string

const (
	ActionCreate  PermissionAction = "create"
	ActionRead    PermissionAction = "read"
	ActionUpdate  PermissionAction = "update"
	ActionDelete  PermissionAction = "delete"
	ActionExecute PermissionAction = "execute"
	ActionManage  PermissionAction = "manage" // Full control
	ActionApprove PermissionAction = "approve"
	ActionPublish PermissionAction = "publish"
	ActionExport  PermissionAction = "export"
	ActionImport  PermissionAction = "import"
)

// EnhancedPermission represents a granular permission with full hierarchy
// This will replace the existing Permission model
type EnhancedPermission struct {
	BaseModel
	ModuleID     uint             `gorm:"index;not null" json:"module_id"`
	DepartmentID uint             `gorm:"index;not null" json:"department_id"`
	ServiceID    uint             `gorm:"index;not null" json:"service_id"`
	ScopeID      uint             `gorm:"index;not null" json:"scope_id"`
	Resource     string           `gorm:"size:100;not null" json:"resource"` // e.g., "users", "posts"
	Action       PermissionAction `gorm:"type:varchar(50);not null" json:"action"`
	Code         string           `gorm:"uniqueIndex;size:200;not null" json:"code"` // Auto-generated unique code
	DisplayName  string           `gorm:"size:200" json:"display_name"`
	Description  string           `gorm:"type:text" json:"description"`
	IsSystem     bool             `gorm:"default:false" json:"is_system"`

	// Relationships
	Module     Module     `gorm:"foreignKey:ModuleID" json:"module,omitempty"`
	Department Department `gorm:"foreignKey:DepartmentID" json:"department,omitempty"`
	Service    Service    `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	Scope      Scope      `gorm:"foreignKey:ScopeID" json:"scope,omitempty"`
	Roles      []Role     `gorm:"many2many:role_enhanced_permissions;" json:"roles,omitempty"`
}

// TableName specifies the table name for EnhancedPermission
func (EnhancedPermission) TableName() string {
	return "enhanced_permissions"
}

// BeforeCreate hook to auto-generate permission code
func (ep *EnhancedPermission) BeforeCreate(tx *gorm.DB) error {
	if ep.Code == "" {
		// Auto-generate code: module:department:service:scope:resource:action
		// Example: crm:sales:customers:department:customers:read
		var module Module
		var department Department
		var service Service
		var scope Scope

		tx.First(&module, ep.ModuleID)
		tx.First(&department, ep.DepartmentID)
		tx.First(&service, ep.ServiceID)
		tx.First(&scope, ep.ScopeID)

		ep.Code = module.Code + ":" + department.Code + ":" + service.Code + ":" + scope.Code + ":" + ep.Resource + ":" + string(ep.Action)
	}
	return nil
}

// RoleEnhancedPermission represents the many-to-many relationship
type RoleEnhancedPermission struct {
	RoleID       uint      `gorm:"primaryKey" json:"role_id"`
	PermissionID uint      `gorm:"primaryKey" json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// TableName specifies the table name for RoleEnhancedPermission
func (RoleEnhancedPermission) TableName() string {
	return "role_enhanced_permissions"
}

// UserEnhancedPermission represents direct user permissions
type UserEnhancedPermission struct {
	UserID       uint       `gorm:"primaryKey" json:"user_id"` // Will be UUID in migration
	PermissionID uint       `gorm:"primaryKey" json:"permission_id"`
	GrantedBy    *uint      `gorm:"index" json:"granted_by,omitempty"` // Who granted this permission
	GrantedAt    time.Time  `json:"granted_at"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"` // Optional expiration
	IsRevoked    bool       `gorm:"default:false" json:"is_revoked"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
	RevokedBy    *uint      `json:"revoked_by,omitempty"`
}

// TableName specifies the table name for UserEnhancedPermission
func (UserEnhancedPermission) TableName() string {
	return "user_enhanced_permissions"
}
