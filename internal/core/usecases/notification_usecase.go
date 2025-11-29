package usecases

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports"
	"github.com/owner/go-cms/pkg/pagination"
	"go.uber.org/zap"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
	ErrUnauthorized         = errors.New("unauthorized to access this notification")
)

type NotificationUseCase struct {
	repo      ports.NotificationRepository
	wsService ports.WebSocketService
	logger    *zap.Logger
}

// NewNotificationUseCase creates a new notification use case
func NewNotificationUseCase(
	repo ports.NotificationRepository,
	wsService ports.WebSocketService,
	logger *zap.Logger,
) *NotificationUseCase {
	return &NotificationUseCase{
		repo:      repo,
		wsService: wsService,
		logger:    logger,
	}
}

// CreateNotification creates a new notification
func (uc *NotificationUseCase) CreateNotification(ctx context.Context, req *domain.CreateNotificationRequest) (*domain.Notification, error) {
	notification := &domain.Notification{
		UserID:    req.UserID,
		Type:      req.Type,
		Priority:  req.Priority,
		Title:     req.Title,
		Message:   req.Message,
		Data:      req.Data,
		Link:      req.Link,
		ImageURL:  req.ImageURL,
		ExpiresAt: req.ExpiresAt,
		IsRead:    false,
	}

	// Set default priority if not provided
	if notification.Priority == "" {
		notification.Priority = domain.NotificationPriorityNormal
	}

	err := uc.repo.Create(ctx, notification)
	if err != nil {
		uc.logger.Error("Failed to create notification",
			zap.Error(err),
			zap.String("title", req.Title),
		)
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	// Send real-time notification via WebSocket
	if err := uc.wsService.NotifyNotification(notification, "created"); err != nil {
		uc.logger.Warn("Failed to send real-time notification",
			zap.Error(err),
			zap.Uint("notification_id", notification.ID),
		)
	}

	uc.logger.Info("Notification created",
		zap.Uint("id", notification.ID),
		zap.String("title", notification.Title),
	)

	return notification, nil
}

// GetNotification retrieves a notification by ID
func (uc *NotificationUseCase) GetNotification(ctx context.Context, id uint, userID uuid.UUID) (*domain.Notification, error) {
	notification, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrNotificationNotFound
	}

	// Check if user has access to this notification
	if notification.UserID != nil && *notification.UserID != userID {
		return nil, ErrUnauthorized
	}

	return notification, nil
}

// GetUserNotifications retrieves notifications for a specific user
func (uc *NotificationUseCase) GetUserNotifications(ctx context.Context, userID uuid.UUID, cursor *pagination.Cursor) ([]*domain.Notification, *pagination.Cursor, error) {
	return uc.repo.GetByUserID(ctx, userID, cursor)
}

// GetAllNotifications retrieves all notifications with filters
func (uc *NotificationUseCase) GetAllNotifications(ctx context.Context, filter *domain.NotificationFilter, cursor *pagination.Cursor) ([]*domain.Notification, *pagination.Cursor, error) {
	return uc.repo.GetAll(ctx, filter, cursor)
}

// UpdateNotification updates a notification
func (uc *NotificationUseCase) UpdateNotification(ctx context.Context, id uint, userID uuid.UUID, req *domain.UpdateNotificationRequest) (*domain.Notification, error) {
	notification, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrNotificationNotFound
	}

	// Check if user has access to this notification
	if notification.UserID != nil && *notification.UserID != userID {
		return nil, ErrUnauthorized
	}

	// Update fields
	if req.Type != nil {
		notification.Type = *req.Type
	}
	if req.Priority != nil {
		notification.Priority = *req.Priority
	}
	if req.Title != nil {
		notification.Title = *req.Title
	}
	if req.Message != nil {
		notification.Message = *req.Message
	}
	if req.Data != nil {
		notification.Data = req.Data
	}
	if req.Link != nil {
		notification.Link = req.Link
	}
	if req.ImageURL != nil {
		notification.ImageURL = req.ImageURL
	}
	if req.IsRead != nil {
		if *req.IsRead {
			notification.MarkAsRead()
		} else {
			notification.MarkAsUnread()
		}
	}
	if req.ExpiresAt != nil {
		notification.ExpiresAt = req.ExpiresAt
	}

	err = uc.repo.Update(ctx, notification)
	if err != nil {
		uc.logger.Error("Failed to update notification",
			zap.Error(err),
			zap.Uint("id", id),
		)
		return nil, fmt.Errorf("failed to update notification: %w", err)
	}

	// Send real-time update via WebSocket
	if err := uc.wsService.NotifyNotification(notification, "updated"); err != nil {
		uc.logger.Warn("Failed to send real-time notification update",
			zap.Error(err),
			zap.Uint("notification_id", notification.ID),
		)
	}

	uc.logger.Info("Notification updated",
		zap.Uint("id", id),
	)

	return notification, nil
}

// DeleteNotification deletes a notification
func (uc *NotificationUseCase) DeleteNotification(ctx context.Context, id uint, userID uuid.UUID) error {
	notification, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return ErrNotificationNotFound
	}

	// Check if user has access to this notification
	if notification.UserID != nil && *notification.UserID != userID {
		return ErrUnauthorized
	}

	err = uc.repo.Delete(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to delete notification",
			zap.Error(err),
			zap.Uint("id", id),
		)
		return fmt.Errorf("failed to delete notification: %w", err)
	}

	// Send real-time deletion via WebSocket
	if err := uc.wsService.NotifyNotification(notification, "deleted"); err != nil {
		uc.logger.Warn("Failed to send real-time notification deletion",
			zap.Error(err),
			zap.Uint("notification_id", notification.ID),
		)
	}

	uc.logger.Info("Notification deleted",
		zap.Uint("id", id),
	)

	return nil
}

// MarkAsRead marks a notification as read
func (uc *NotificationUseCase) MarkAsRead(ctx context.Context, id uint, userID uuid.UUID) error {
	notification, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return ErrNotificationNotFound
	}

	// Check if user has access to this notification
	if notification.UserID != nil && *notification.UserID != userID {
		return ErrUnauthorized
	}

	err = uc.repo.MarkAsRead(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to mark notification as read",
			zap.Error(err),
			zap.Uint("id", id),
		)
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}

	// Reload notification to get updated data
	notification, _ = uc.repo.GetByID(ctx, id)

	// Send real-time update via WebSocket
	if err := uc.wsService.NotifyNotification(notification, "read"); err != nil {
		uc.logger.Warn("Failed to send real-time notification read status",
			zap.Error(err),
			zap.Uint("notification_id", notification.ID),
		)
	}

	return nil
}

// MarkAsUnread marks a notification as unread
func (uc *NotificationUseCase) MarkAsUnread(ctx context.Context, id uint, userID uuid.UUID) error {
	notification, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return ErrNotificationNotFound
	}

	// Check if user has access to this notification
	if notification.UserID != nil && *notification.UserID != userID {
		return ErrUnauthorized
	}

	err = uc.repo.MarkAsUnread(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to mark notification as unread",
			zap.Error(err),
			zap.Uint("id", id),
		)
		return fmt.Errorf("failed to mark notification as unread: %w", err)
	}

	// Reload notification to get updated data
	notification, _ = uc.repo.GetByID(ctx, id)

	// Send real-time update via WebSocket
	if err := uc.wsService.NotifyNotification(notification, "unread"); err != nil {
		uc.logger.Warn("Failed to send real-time notification unread status",
			zap.Error(err),
			zap.Uint("notification_id", notification.ID),
		)
	}

	return nil
}

// MarkAllAsRead marks all notifications as read for a user
func (uc *NotificationUseCase) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	err := uc.repo.MarkAllAsRead(ctx, userID)
	if err != nil {
		uc.logger.Error("Failed to mark all notifications as read",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	uc.logger.Info("All notifications marked as read",
		zap.String("user_id", userID.String()),
	)

	return nil
}

// GetUnreadCount gets the count of unread notifications for a user
func (uc *NotificationUseCase) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return uc.repo.GetUnreadCount(ctx, userID)
}

// GetStats gets notification statistics for a user
func (uc *NotificationUseCase) GetStats(ctx context.Context, userID uuid.UUID) (*domain.NotificationStats, error) {
	return uc.repo.GetStats(ctx, userID)
}

// DeleteAllUserNotifications deletes all notifications for a user
func (uc *NotificationUseCase) DeleteAllUserNotifications(ctx context.Context, userID uuid.UUID) error {
	err := uc.repo.DeleteByUserID(ctx, userID)
	if err != nil {
		uc.logger.Error("Failed to delete all user notifications",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)
		return fmt.Errorf("failed to delete all user notifications: %w", err)
	}

	uc.logger.Info("All user notifications deleted",
		zap.String("user_id", userID.String()),
	)

	return nil
}

// CleanupExpiredNotifications deletes all expired notifications
func (uc *NotificationUseCase) CleanupExpiredNotifications(ctx context.Context) (int64, error) {
	count, err := uc.repo.DeleteExpired(ctx)
	if err != nil {
		uc.logger.Error("Failed to cleanup expired notifications",
			zap.Error(err),
		)
		return 0, fmt.Errorf("failed to cleanup expired notifications: %w", err)
	}

	if count > 0 {
		uc.logger.Info("Expired notifications cleaned up",
			zap.Int64("count", count),
		)
	}

	return count, nil
}

// BroadcastNotification creates and broadcasts a notification to all users
func (uc *NotificationUseCase) BroadcastNotification(ctx context.Context, req *domain.CreateNotificationRequest) (*domain.Notification, error) {
	// Ensure UserID is nil for broadcast
	req.UserID = nil

	return uc.CreateNotification(ctx, req)
}

// GetBroadcastNotifications retrieves broadcast notifications
func (uc *NotificationUseCase) GetBroadcastNotifications(ctx context.Context, cursor *pagination.Cursor) ([]*domain.Notification, *pagination.Cursor, error) {
	return uc.repo.GetBroadcastNotifications(ctx, cursor)
}
