package postgres

import (
	"context"
	"time"

	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/pkg/pagination"
	"gorm.io/gorm"
)

type auditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *gorm.DB) repositories.AuditLogRepository {
	return &auditLogRepository{db: db}
}

// Create creates a new audit log entry
func (r *auditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// List retrieves audit logs with filters and pagination
func (r *auditLogRepository) List(ctx context.Context, filter repositories.AuditLogFilter, page *pagination.OffsetPagination) ([]*domain.AuditLog, int64, error) {
	var logs []*domain.AuditLog
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.AuditLog{})

	// Apply filters
	query = r.applyFilters(query, filter)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	offset := (page.Page - 1) * page.Limit
	if err := query.
		Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(page.Limit).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// ListWithCursor retrieves audit logs with cursor-based pagination
func (r *auditLogRepository) ListWithCursor(ctx context.Context, filter repositories.AuditLogFilter, cursor *pagination.Cursor, limit int) ([]*domain.AuditLog, *pagination.Cursor, error) {
	var logs []*domain.AuditLog

	query := r.db.WithContext(ctx).Model(&domain.AuditLog{})

	// Apply filters
	query = r.applyFilters(query, filter)

	// Apply cursor
	if cursor != nil && cursor.After != "" {
		query = query.Where("id > ?", cursor.After)
	}

	// Fetch one extra to determine if there's a next page
	if err := query.
		Preload("User").
		Order("id ASC").
		Limit(limit + 1).
		Find(&logs).Error; err != nil {
		return nil, nil, err
	}

	// Determine next cursor
	var nextCursor *pagination.Cursor
	if len(logs) > limit {
		logs = logs[:limit]
		lastLog := logs[len(logs)-1]
		nextCursor = &pagination.Cursor{
			After:   string(rune(lastLog.ID)),
			HasMore: true,
		}
	}

	return logs, nextCursor, nil
}

// GetByID retrieves an audit log by ID
func (r *auditLogRepository) GetByID(ctx context.Context, id uint) (*domain.AuditLog, error) {
	var log domain.AuditLog
	if err := r.db.WithContext(ctx).
		Preload("User").
		First(&log, id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

// GetByUserID retrieves audit logs for a specific user
func (r *auditLogRepository) GetByUserID(ctx context.Context, userID uint, limit int) ([]*domain.AuditLog, error) {
	var logs []*domain.AuditLog
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// GetByResource retrieves audit logs for a specific resource
func (r *auditLogRepository) GetByResource(ctx context.Context, resource string, resourceID uint) ([]*domain.AuditLog, error) {
	var logs []*domain.AuditLog
	if err := r.db.WithContext(ctx).
		Where("resource = ? AND resource_id = ?", resource, resourceID).
		Preload("User").
		Order("created_at DESC").
		Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

// DeleteOlderThan deletes audit logs older than specified days
func (r *auditLogRepository) DeleteOlderThan(ctx context.Context, days int) error {
	cutoffDate := time.Now().AddDate(0, 0, -days)
	return r.db.WithContext(ctx).
		Where("created_at < ?", cutoffDate).
		Delete(&domain.AuditLog{}).Error
}

// applyFilters applies filters to the query
func (r *auditLogRepository) applyFilters(query *gorm.DB, filter repositories.AuditLogFilter) *gorm.DB {
	if filter.UserID != nil {
		query = query.Where("user_id = ?", *filter.UserID)
	}

	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}

	if filter.Resource != "" {
		query = query.Where("resource = ?", filter.Resource)
	}

	if filter.ResourceID != nil {
		query = query.Where("resource_id = ?", *filter.ResourceID)
	}

	if filter.IPAddress != "" {
		query = query.Where("ip_address = ?", filter.IPAddress)
	}

	if filter.DateFrom != "" {
		query = query.Where("created_at >= ?", filter.DateFrom)
	}

	if filter.DateTo != "" {
		query = query.Where("created_at <= ?", filter.DateTo)
	}

	return query
}
