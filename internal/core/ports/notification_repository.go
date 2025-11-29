package ports

import (
	"context"

	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/pkg/pagination"
)

type NotificationRepository interface {
	Create(ctx context.Context, notification *domain.Notification) error
	GetByID(ctx context.Context, id uint) (*domain.Notification, error)

	// GetByUserID retrieves notifications for a specific user with pagination
	GetByUserID(ctx context.Context, userID uuid.UUID, cursor *pagination.Cursor) ([]*domain.Notification, *pagination.Cursor, error)

	// GetAll retrieves all notifications with pagination and filters
	GetAll(ctx context.Context, filter *domain.NotificationFilter, cursor *pagination.Cursor) ([]*domain.Notification, *pagination.Cursor, error)

	// Update updates a notification
	Update(ctx context.Context, notification *domain.Notification) error

	// Delete deletes a notification by ID
	Delete(ctx context.Context, id uint) error

	// DeleteByUserID deletes all notifications for a user
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error

	// MarkAsRead marks a notification as read
	MarkAsRead(ctx context.Context, id uint) error

	// MarkAsUnread marks a notification as unread
	MarkAsUnread(ctx context.Context, id uint) error

	// MarkAllAsRead marks all notifications as read for a user
	MarkAllAsRead(ctx context.Context, userID uuid.UUID) error

	// GetUnreadCount gets the count of unread notifications for a user
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error)

	// GetStats gets notification statistics for a user
	GetStats(ctx context.Context, userID uuid.UUID) (*domain.NotificationStats, error)

	// DeleteExpired deletes all expired notifications
	DeleteExpired(ctx context.Context) (int64, error)

	// GetBroadcastNotifications retrieves broadcast notifications (UserID is nil)
	GetBroadcastNotifications(ctx context.Context, cursor *pagination.Cursor) ([]*domain.Notification, *pagination.Cursor, error)
}

// NotificationFilter represents filters for notification queries
type NotificationFilter struct {
	UserID *uint
	Type   string
	Read   *bool
}
