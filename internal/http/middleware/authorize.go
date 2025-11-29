package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/infrastructure/cache"
	"github.com/owner/go-cms/pkg/errors"
	"github.com/owner/go-cms/pkg/logger"
	"github.com/owner/go-cms/pkg/response"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PermissionChecker is an interface for checking permissions
type PermissionChecker interface {
	HasPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error)
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
}

// DefaultPermissionChecker implements PermissionChecker
type DefaultPermissionChecker struct {
	db *gorm.DB
}

// NewPermissionChecker creates a new permission checker
func NewPermissionChecker(db *gorm.DB) *DefaultPermissionChecker {
	return &DefaultPermissionChecker{db: db}
}

// HasPermission checks if a user has a specific permission
func (pc *DefaultPermissionChecker) HasPermission(ctx context.Context, userID uuid.UUID, permission string) (bool, error) {
	// Try to get from cache first
	cachedPerms, err := cache.GetPermissions(ctx, userID)

	if err == nil && len(cachedPerms) > 0 {
		// Check if permission exists in cache
		if _, exists := cachedPerms[permission]; exists {
			return true, nil
		}
		// If not found in cache, it means user doesn't have this permission
		return false, nil
	}

	// Cache miss, query database
	permissions, err := pc.GetUserPermissions(ctx, userID)
	if err != nil {
		return false, err
	}

	// Cache the permissions
	_ = cache.CachePermissions(ctx, userID, permissions, 15*time.Minute)

	// Check if user has the permission
	for _, perm := range permissions {
		if perm == permission {
			return true, nil
		}
	}

	return false, nil
}

// GetUserPermissions gets all permissions for a user
func (pc *DefaultPermissionChecker) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error) {
	var permissions []string

	// Get permissions from user roles
	query := `
		SELECT DISTINCT CONCAT(p.module, ':', p.department, ':', p.service, ':', p.resource, ':', p.action) as permission
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		INNER JOIN user_roles ur ON rp.role_id = ur.role_id
		WHERE ur.user_id = ? AND ur.deleted_at IS NULL

		UNION

		SELECT DISTINCT CONCAT(p.module, ':', p.department, ':', p.service, ':', p.resource, ':', p.action) as permission
		FROM permissions p
		INNER JOIN user_permissions up ON p.id = up.permission_id
		WHERE up.user_id = ?
	`

	var permissionRows []struct {
		Permission string
	}

	if err := pc.db.WithContext(ctx).Raw(query, userID, userID).Scan(&permissionRows).Error; err != nil {
		return nil, err
	}

	for _, row := range permissionRows {
		permissions = append(permissions, row.Permission)
	}

	return permissions, nil
}

// AuthorizeMiddleware checks if the user has the required permission
func AuthorizeMiddleware(checker PermissionChecker, requiredPermission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			response.Error(c, errors.ErrUnauthorized)
			c.Abort()
			return
		}

		// Check permission
		hasPermission, err := checker.HasPermission(c.Request.Context(), userID, requiredPermission)
		if err != nil {
			logger.Error("Failed to check permission",
				zap.String("user_id", userID.String()),
				zap.String("permission", requiredPermission),
				zap.Error(err),
			)
			response.Error(c, errors.ErrInternal)
			c.Abort()
			return
		}

		if !hasPermission {
			logger.Warn("Permission denied",
				zap.String("user_id", userID.String()),
				zap.String("permission", requiredPermission),
			)
			response.Error(c, errors.ErrInsufficientPermissions)
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuthorizeAnyMiddleware checks if the user has any of the required permissions
func AuthorizeAnyMiddleware(checker PermissionChecker, requiredPermissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			response.Error(c, errors.ErrUnauthorized)
			c.Abort()
			return
		}

		// Check if user has any of the required permissions
		for _, permission := range requiredPermissions {
			hasPermission, err := checker.HasPermission(c.Request.Context(), userID, permission)
			if err != nil {
				logger.Error("Failed to check permission",
					zap.String("user_id", userID.String()),
					zap.String("permission", permission),
					zap.Error(err),
				)
				continue
			}

			if hasPermission {
				c.Next()
				return
			}
		}

		logger.Warn("Permission denied - none of required permissions found",
			zap.String("user_id", userID.String()),
			zap.Strings("permissions", requiredPermissions),
		)
		response.Error(c, errors.ErrInsufficientPermissions)
		c.Abort()
	}
}

// AuthorizeAllMiddleware checks if the user has all of the required permissions
func AuthorizeAllMiddleware(checker PermissionChecker, requiredPermissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			response.Error(c, errors.ErrUnauthorized)
			c.Abort()
			return
		}

		// Check if user has all required permissions
		for _, permission := range requiredPermissions {
			hasPermission, err := checker.HasPermission(c.Request.Context(), userID, permission)
			if err != nil {
				logger.Error("Failed to check permission",
					zap.String("user_id", userID.String()),
					zap.String("permission", permission),
					zap.Error(err),
				)
				response.Error(c, errors.ErrInternal)
				c.Abort()
				return
			}

			if !hasPermission {
				logger.Warn("Permission denied - missing required permission",
					zap.String("user_id", userID.String()),
					zap.String("permission", permission),
				)
				response.Error(c, errors.ErrInsufficientPermissions)
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// BuildPermission builds a permission string from components
func BuildPermission(module, department, service, resource, action string) string {
	return fmt.Sprintf("%s:%s:%s:%s:%s", module, department, service, resource, action)
}

// ParsePermission parses a permission string into components
func ParsePermission(permission string) (module, department, service, resource, action string) {
	parts := strings.Split(permission, ":")
	if len(parts) != 5 {
		return "", "", "", "", ""
	}
	return parts[0], parts[1], parts[2], parts[3], parts[4]
}

// Permission constants for common operations
const (
	// CRM module permissions
	PermissionCRMCustomersCreate = "crm:sales:customers:customers:create"
	PermissionCRMCustomersRead   = "crm:sales:customers:customers:read"
	PermissionCRMCustomersUpdate = "crm:sales:customers:customers:update"
	PermissionCRMCustomersDelete = "crm:sales:customers:customers:delete"

	// Content module permissions
	PermissionContentPostsCreate  = "content:editorial:posts:posts:create"
	PermissionContentPostsRead    = "content:editorial:posts:posts:read"
	PermissionContentPostsUpdate  = "content:editorial:posts:posts:update"
	PermissionContentPostsDelete  = "content:editorial:posts:posts:delete"
	PermissionContentPostsPublish = "content:editorial:posts:posts:publish"

	// Media module permissions
	PermissionContentMediaUpload = "content:editorial:media:media:upload"
	PermissionContentMediaRead   = "content:editorial:media:media:read"
	PermissionContentMediaDelete = "content:editorial:media:media:delete"

	// Admin module permissions
	PermissionAdminUsersCreate = "admin:system:users:users:create"
	PermissionAdminUsersRead   = "admin:system:users:users:read"
	PermissionAdminUsersUpdate = "admin:system:users:users:update"
	PermissionAdminUsersDelete = "admin:system:users:users:delete"

	PermissionAdminRolesCreate = "admin:system:roles:roles:create"
	PermissionAdminRolesRead   = "admin:system:roles:roles:read"
	PermissionAdminRolesUpdate = "admin:system:roles:roles:update"
	PermissionAdminRolesDelete = "admin:system:roles:roles:delete"

	PermissionAdminPermissionsManage = "admin:system:permissions:permissions:manage"
)
