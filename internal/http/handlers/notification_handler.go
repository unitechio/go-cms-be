package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/usecases"
	"github.com/owner/go-cms/pkg/pagination"
	"go.uber.org/zap"
)

type NotificationHandler struct {
	useCase *usecases.NotificationUseCase
	logger  *zap.Logger
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(useCase *usecases.NotificationUseCase, logger *zap.Logger) *NotificationHandler {
	return &NotificationHandler{
		useCase: useCase,
		logger:  logger,
	}
}

// CreateNotification creates a new notification
// @Summary Create notification
// @Description Create a new notification for a user or broadcast to all users
// @Tags Notifications
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param notification body domain.CreateNotificationRequest true "Notification data"
// @Success 201 {object} domain.Notification "Notification created"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /notifications [post]
func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	var req domain.CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notification, err := h.useCase.CreateNotification(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}

	c.JSON(http.StatusCreated, notification)
}

// GetNotification retrieves a notification by ID
// @Summary Get notification
// @Description Get a notification by ID
// @Tags Notifications
// @Security BearerAuth
// @Produce json
// @Param id path int true "Notification ID"
// @Success 200 {object} domain.Notification "Notification details"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Not Found"
// @Router /notifications/{id} [get]
func (h *NotificationHandler) GetNotification(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	userID := h.getUserID(c)
	notification, err := h.useCase.GetNotification(c.Request.Context(), uint(id), userID)
	if err != nil {
		if err == usecases.ErrNotificationNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
			return
		}
		if err == usecases.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to access this notification"})
			return
		}
		h.logger.Error("Failed to get notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notification"})
		return
	}

	c.JSON(http.StatusOK, notification)
}

// GetMyNotifications retrieves notifications for the current user
// @Summary Get my notifications
// @Description Get notifications for the current user with pagination
// @Tags Notifications
// @Security BearerAuth
// @Produce json
// @Param after query int false "Cursor after ID"
// @Param limit query int false "Limit per page" default(20)
// @Success 200 {object} map[string]interface{} "Notifications list"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /notifications/me [get]
func (h *NotificationHandler) GetMyNotifications(c *gin.Context) {
	userID := h.getUserID(c)
	cursor := h.parseCursor(c)

	notifications, nextCursor, err := h.useCase.GetUserNotifications(c.Request.Context(), userID, cursor)
	if err != nil {
		h.logger.Error("Failed to get user notifications", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notifications"})
		return
	}

	response := gin.H{
		"data": notifications,
	}

	if nextCursor != nil {
		response["next_cursor"] = nextCursor
	}

	c.JSON(http.StatusOK, response)
}

// GetAllNotifications retrieves all notifications with filters (admin only)
// @Summary Get all notifications
// @Description Get all notifications with filters (admin only)
// @Tags Notifications
// @Security BearerAuth
// @Produce json
// @Param user_id query string false "Filter by user ID"
// @Param type query string false "Filter by type (info, success, warning, error)"
// @Param priority query string false "Filter by priority (low, normal, high, urgent)"
// @Param is_read query boolean false "Filter by read status"
// @Param after query int false "Cursor after ID"
// @Param limit query int false "Limit per page" default(20)
// @Success 200 {object} map[string]interface{} "Notifications list"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /notifications [get]
func (h *NotificationHandler) GetAllNotifications(c *gin.Context) {
	filter := &domain.NotificationFilter{}

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err == nil {
			filter.UserID = &userID
		}
	}

	if typeStr := c.Query("type"); typeStr != "" {
		notifType := domain.NotificationType(typeStr)
		filter.Type = &notifType
	}

	if priorityStr := c.Query("priority"); priorityStr != "" {
		priority := domain.NotificationPriority(priorityStr)
		filter.Priority = &priority
	}

	if isReadStr := c.Query("is_read"); isReadStr != "" {
		isRead := isReadStr == "true"
		filter.IsRead = &isRead
	}

	cursor := h.parseCursor(c)

	notifications, nextCursor, err := h.useCase.GetAllNotifications(c.Request.Context(), filter, cursor)
	if err != nil {
		h.logger.Error("Failed to get all notifications", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notifications"})
		return
	}

	response := gin.H{
		"data": notifications,
	}

	if nextCursor != nil {
		response["next_cursor"] = nextCursor
	}

	c.JSON(http.StatusOK, response)
}

// UpdateNotification updates a notification
// @Summary Update notification
// @Description Update a notification
// @Tags Notifications
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Notification ID"
// @Param notification body domain.UpdateNotificationRequest true "Notification update data"
// @Success 200 {object} domain.Notification "Updated notification"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Not Found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /notifications/{id} [put]
func (h *NotificationHandler) UpdateNotification(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	var req domain.UpdateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := h.getUserID(c)
	notification, err := h.useCase.UpdateNotification(c.Request.Context(), uint(id), userID, &req)
	if err != nil {
		if err == usecases.ErrNotificationNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
			return
		}
		if err == usecases.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to update this notification"})
			return
		}
		h.logger.Error("Failed to update notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notification"})
		return
	}

	c.JSON(http.StatusOK, notification)
}

// DeleteNotification deletes a notification
// @Summary Delete notification
// @Description Delete a notification
// @Tags Notifications
// @Security BearerAuth
// @Param id path int true "Notification ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Not Found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /notifications/{id} [delete]
func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	userID := h.getUserID(c)
	err = h.useCase.DeleteNotification(c.Request.Context(), uint(id), userID)
	if err != nil {
		if err == usecases.ErrNotificationNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
			return
		}
		if err == usecases.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to delete this notification"})
			return
		}
		h.logger.Error("Failed to delete notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notification"})
		return
	}

	c.Status(http.StatusNoContent)
}

// MarkAsRead marks a notification as read
// @Summary Mark notification as read
// @Description Mark a notification as read
// @Tags Notifications
// @Security BearerAuth
// @Param id path int true "Notification ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Not Found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /notifications/{id}/read [post]
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	userID := h.getUserID(c)
	err = h.useCase.MarkAsRead(c.Request.Context(), uint(id), userID)
	if err != nil {
		if err == usecases.ErrNotificationNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
			return
		}
		if err == usecases.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to access this notification"})
			return
		}
		h.logger.Error("Failed to mark notification as read", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification as read"})
		return
	}

	c.Status(http.StatusNoContent)
}

// MarkAsUnread marks a notification as unread
// @Summary Mark notification as unread
// @Description Mark a notification as unread
// @Tags Notifications
// @Security BearerAuth
// @Param id path int true "Notification ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 404 {object} map[string]interface{} "Not Found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /notifications/{id}/unread [post]
func (h *NotificationHandler) MarkAsUnread(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	userID := h.getUserID(c)
	err = h.useCase.MarkAsUnread(c.Request.Context(), uint(id), userID)
	if err != nil {
		if err == usecases.ErrNotificationNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
			return
		}
		if err == usecases.ErrUnauthorized {
			c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized to access this notification"})
			return
		}
		h.logger.Error("Failed to mark notification as unread", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification as unread"})
		return
	}

	c.Status(http.StatusNoContent)
}

// MarkAllAsRead marks all notifications as read for the current user
// @Summary Mark all notifications as read
// @Description Mark all notifications as read for the current user
// @Tags Notifications
// @Security BearerAuth
// @Success 204 "No Content"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /notifications/mark-all-read [post]
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := h.getUserID(c)
	err := h.useCase.MarkAllAsRead(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to mark all notifications as read", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark all notifications as read"})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetUnreadCount gets the count of unread notifications
// @Summary Get unread count
// @Description Get the count of unread notifications for the current user
// @Tags Notifications
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "Unread count"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /notifications/unread-count [get]
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID := h.getUserID(c)
	count, err := h.useCase.GetUnreadCount(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get unread count", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get unread count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"unread_count": count})
}

// GetStats gets notification statistics
// @Summary Get notification statistics
// @Description Get notification statistics for the current user
// @Tags Notifications
// @Security BearerAuth
// @Produce json
// @Success 200 {object} domain.NotificationStats "Notification statistics"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /notifications/stats [get]
func (h *NotificationHandler) GetStats(c *gin.Context) {
	userID := h.getUserID(c)
	stats, err := h.useCase.GetStats(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get notification stats", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get notification stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// DeleteAllMyNotifications deletes all notifications for the current user
// @Summary Delete all my notifications
// @Description Delete all notifications for the current user
// @Tags Notifications
// @Security BearerAuth
// @Success 204 "No Content"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /notifications/me [delete]
func (h *NotificationHandler) DeleteAllMyNotifications(c *gin.Context) {
	userID := h.getUserID(c)
	err := h.useCase.DeleteAllUserNotifications(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to delete all user notifications", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete all notifications"})
		return
	}

	c.Status(http.StatusNoContent)
}

// BroadcastNotification creates and broadcasts a notification to all users (admin only)
// @Summary Broadcast notification
// @Description Create and broadcast a notification to all users (admin only)
// @Tags Notifications
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param notification body domain.CreateNotificationRequest true "Notification data"
// @Success 201 {object} domain.Notification "Notification created and broadcasted"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /notifications/broadcast [post]
func (h *NotificationHandler) BroadcastNotification(c *gin.Context) {
	var req domain.CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notification, err := h.useCase.BroadcastNotification(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to broadcast notification", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to broadcast notification"})
		return
	}

	c.JSON(http.StatusCreated, notification)
}

// Helper methods
func (h *NotificationHandler) getUserID(c *gin.Context) uuid.UUID {
	userID, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil
	}

	if uid, ok := userID.(uuid.UUID); ok {
		return uid
	}

	return uuid.Nil
}

func (h *NotificationHandler) parseCursor(c *gin.Context) *pagination.Cursor {
	cursor := &pagination.Cursor{}

	if afterStr := c.Query("after"); afterStr != "" {
		if after, err := strconv.ParseUint(afterStr, 10, 32); err == nil {
			cursor.ID = uint(after)
		}
	}

	return cursor
}
