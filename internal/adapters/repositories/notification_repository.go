package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports"
	"github.com/owner/go-cms/pkg/pagination"
	"gorm.io/gorm"
)

type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *gorm.DB) ports.NotificationRepository {
	return &notificationRepository{db: db}
}

// Create creates a new notification
func (r *notificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

// GetByID retrieves a notification by ID
func (r *notificationRepository) GetByID(ctx context.Context, id uint) (*domain.Notification, error) {
	var notification domain.Notification
	err := r.db.WithContext(ctx).
		Preload("User").
		First(&notification, id).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

// GetByUserID retrieves notifications for a specific user with pagination
func (r *notificationRepository) GetByUserID(ctx context.Context, userID uuid.UUID, cursor *pagination.Cursor) ([]*domain.Notification, *pagination.Cursor, error) {
	var notifications []*domain.Notification

	query := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("id DESC")

	// Apply cursor pagination
	limit := 20
	if cursor != nil && cursor.ID > 0 {
		query = query.Where("id < ?", cursor.ID)
	}

	err := query.Limit(limit + 1).Find(&notifications).Error
	if err != nil {
		return nil, nil, err
	}

	// Build next cursor
	var nextCursor *pagination.Cursor
	if len(notifications) > limit {
		notifications = notifications[:limit]
		lastID := notifications[len(notifications)-1].ID
		nextCursor = &pagination.Cursor{
			ID:      lastID,
			HasMore: true,
		}
	}

	return notifications, nextCursor, nil
}

// GetAll retrieves all notifications with pagination and filters
func (r *notificationRepository) GetAll(ctx context.Context, filter *domain.NotificationFilter, cursor *pagination.Cursor) ([]*domain.Notification, *pagination.Cursor, error) {
	var notifications []*domain.Notification

	query := r.db.WithContext(ctx).Model(&domain.Notification{})

	// Apply filters
	if filter != nil {
		if filter.UserID != nil {
			query = query.Where("user_id = ?", *filter.UserID)
		}
		if filter.Type != nil {
			query = query.Where("type = ?", *filter.Type)
		}
		if filter.Priority != nil {
			query = query.Where("priority = ?", *filter.Priority)
		}
		if filter.IsRead != nil {
			query = query.Where("is_read = ?", *filter.IsRead)
		}
		if filter.FromDate != nil {
			query = query.Where("created_at >= ?", *filter.FromDate)
		}
		if filter.ToDate != nil {
			query = query.Where("created_at <= ?", *filter.ToDate)
		}
	}

	query = query.Order("id DESC")

	// Apply cursor pagination
	limit := 20
	if cursor != nil && cursor.ID > 0 {
		query = query.Where("id < ?", cursor.ID)
	}

	err := query.Limit(limit + 1).Preload("User").Find(&notifications).Error
	if err != nil {
		return nil, nil, err
	}

	// Build next cursor
	var nextCursor *pagination.Cursor
	if len(notifications) > limit {
		notifications = notifications[:limit]
		lastID := notifications[len(notifications)-1].ID
		nextCursor = &pagination.Cursor{
			ID:      lastID,
			HasMore: true,
		}
	}

	return notifications, nextCursor, nil
}

// Update updates a notification
func (r *notificationRepository) Update(ctx context.Context, notification *domain.Notification) error {
	return r.db.WithContext(ctx).Save(notification).Error
}

// Delete deletes a notification by ID
func (r *notificationRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Notification{}, id).Error
}

// DeleteByUserID deletes all notifications for a user
func (r *notificationRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&domain.Notification{}).Error
}

// MarkAsRead marks a notification as read
func (r *notificationRepository) MarkAsRead(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error
}

// MarkAsUnread marks a notification as unread
func (r *notificationRepository) MarkAsUnread(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_read": false,
			"read_at": nil,
		}).Error
}

// MarkAllAsRead marks all notifications as read for a user
func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": now,
		}).Error
}

// GetUnreadCount gets the count of unread notifications for a user
func (r *notificationRepository) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

// GetStats gets notification statistics for a user
func (r *notificationRepository) GetStats(ctx context.Context, userID uuid.UUID) (*domain.NotificationStats, error) {
	stats := &domain.NotificationStats{
		ByType:     make(map[domain.NotificationType]int64),
		ByPriority: make(map[domain.NotificationPriority]int64),
	}

	// Total count
	err := r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("user_id = ?", userID).
		Count(&stats.Total).Error
	if err != nil {
		return nil, err
	}

	// Unread count
	err = r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&stats.Unread).Error
	if err != nil {
		return nil, err
	}

	stats.Read = stats.Total - stats.Unread

	// Count by type
	var typeResults []struct {
		Type  domain.NotificationType
		Count int64
	}
	err = r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Select("type, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("type").
		Scan(&typeResults).Error
	if err != nil {
		return nil, err
	}

	for _, result := range typeResults {
		stats.ByType[result.Type] = result.Count
	}

	// Count by priority
	var priorityResults []struct {
		Priority domain.NotificationPriority
		Count    int64
	}
	err = r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Select("priority, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("priority").
		Scan(&priorityResults).Error
	if err != nil {
		return nil, err
	}

	for _, result := range priorityResults {
		stats.ByPriority[result.Priority] = result.Count
	}

	return stats, nil
}

// DeleteExpired deletes all expired notifications
func (r *notificationRepository) DeleteExpired(ctx context.Context) (int64, error) {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Where("expires_at IS NOT NULL AND expires_at < ?", now).
		Delete(&domain.Notification{})

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

// GetBroadcastNotifications retrieves broadcast notifications (UserID is nil)
func (r *notificationRepository) GetBroadcastNotifications(ctx context.Context, cursor *pagination.Cursor) ([]*domain.Notification, *pagination.Cursor, error) {
	var notifications []*domain.Notification

	query := r.db.WithContext(ctx).
		Where("user_id IS NULL").
		Order("id DESC")

	// Apply cursor pagination
	limit := 20
	if cursor != nil && cursor.ID > 0 {
		query = query.Where("id < ?", cursor.ID)
	}

	err := query.Limit(limit + 1).Find(&notifications).Error
	if err != nil {
		return nil, nil, err
	}

	// Build next cursor
	var nextCursor *pagination.Cursor
	if len(notifications) > limit {
		notifications = notifications[:limit]
		lastID := notifications[len(notifications)-1].ID
		nextCursor = &pagination.Cursor{
			ID:      lastID,
			HasMore: true,
		}
	}

	return notifications, nextCursor, nil
}
