package repositories

import (
	"context"

	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/pkg/pagination"
)

// AuditLogRepository defines the interface for audit log data operations
type AuditLogRepository interface {
	// Create
	Create(ctx context.Context, log *domain.AuditLog) error

	// List operations
	List(ctx context.Context, filter AuditLogFilter, page *pagination.OffsetPagination) ([]*domain.AuditLog, int64, error)
	ListWithCursor(ctx context.Context, filter AuditLogFilter, cursor *pagination.Cursor, limit int) ([]*domain.AuditLog, *pagination.Cursor, error)

	// Get operations
	GetByID(ctx context.Context, id uint) (*domain.AuditLog, error)
	GetByUserID(ctx context.Context, userID uint, limit int) ([]*domain.AuditLog, error)
	GetByResource(ctx context.Context, resource string, resourceID uint) ([]*domain.AuditLog, error)

	// Cleanup
	DeleteOlderThan(ctx context.Context, days int) error
}

// AuditLogFilter represents filters for audit log queries
type AuditLogFilter struct {
	UserID     *uint
	Action     domain.AuditAction
	Resource   string
	ResourceID *uint
	IPAddress  string
	DateFrom   string
	DateTo     string
}

// SystemSettingRepository defines the interface for system setting data operations
type SystemSettingRepository interface {
	// Basic CRUD
	Create(ctx context.Context, setting *domain.SystemSetting) error
	GetByKey(ctx context.Context, key string) (*domain.SystemSetting, error)
	Update(ctx context.Context, setting *domain.SystemSetting) error
	Delete(ctx context.Context, key string) error

	// List operations
	List(ctx context.Context, category string) ([]*domain.SystemSetting, error)
	GetPublicSettings(ctx context.Context) ([]*domain.SystemSetting, error)
	GetByCategory(ctx context.Context, category string) ([]*domain.SystemSetting, error)

	// Bulk operations
	BulkUpdate(ctx context.Context, settings []*domain.SystemSetting) error
}

// ActivityLogRepository defines the interface for activity log data operations
type ActivityLogRepository interface {
	// Create
	Create(ctx context.Context, log *domain.ActivityLog) error

	// List operations
	List(ctx context.Context, filter ActivityLogFilter, page *pagination.OffsetPagination) ([]*domain.ActivityLog, int64, error)
	GetByUserID(ctx context.Context, userID uint, limit int) ([]*domain.ActivityLog, error)

	// Cleanup
	DeleteOlderThan(ctx context.Context, days int) error
}

// ActivityLogFilter represents filters for activity log queries
type ActivityLogFilter struct {
	UserID   *uint
	Activity string
	DateFrom string
	DateTo   string
}

// EmailTemplateRepository defines the interface for email template data operations
type EmailTemplateRepository interface {
	// Basic CRUD
	Create(ctx context.Context, template *domain.EmailTemplate) error
	GetByID(ctx context.Context, id uint) (*domain.EmailTemplate, error)
	GetByName(ctx context.Context, name string) (*domain.EmailTemplate, error)
	Update(ctx context.Context, template *domain.EmailTemplate) error
	Delete(ctx context.Context, id uint) error

	// List operations
	List(ctx context.Context, category string) ([]*domain.EmailTemplate, error)
	GetActive(ctx context.Context) ([]*domain.EmailTemplate, error)
	GetByCategory(ctx context.Context, category string) ([]*domain.EmailTemplate, error)
}

// EmailLogRepository defines the interface for email log data operations
type EmailLogRepository interface {
	// Create
	Create(ctx context.Context, log *domain.EmailLog) error

	// List operations
	List(ctx context.Context, filter EmailLogFilter, page *pagination.OffsetPagination) ([]*domain.EmailLog, int64, error)
	GetByID(ctx context.Context, id uint) (*domain.EmailLog, error)

	// Status operations
	UpdateStatus(ctx context.Context, id uint, status string) error
	MarkAsSent(ctx context.Context, id uint) error
	MarkAsFailed(ctx context.Context, id uint, errorMsg string) error

	// Cleanup
	DeleteOlderThan(ctx context.Context, days int) error
}

// EmailLogFilter represents filters for email log queries
type EmailLogFilter struct {
	To         string
	Status     string
	TemplateID *uint
	DateFrom   string
	DateTo     string
}
