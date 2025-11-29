package domain

import (
	"time"

	"github.com/google/uuid"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeInfo    NotificationType = "info"
	NotificationTypeSuccess NotificationType = "success"
	NotificationTypeWarning NotificationType = "warning"
	NotificationTypeError   NotificationType = "error"
)

// NotificationPriority represents the priority level of notification
type NotificationPriority string

const (
	NotificationPriorityLow    NotificationPriority = "low"
	NotificationPriorityNormal NotificationPriority = "normal"
	NotificationPriorityHigh   NotificationPriority = "high"
	NotificationPriorityUrgent NotificationPriority = "urgent"
)

// Notification represents a notification entity
type Notification struct {
	ID        uint                 `json:"id" gorm:"primaryKey"`
	UserID    *uuid.UUID           `json:"user_id" gorm:"type:uuid;index"` // nil for broadcast notifications
	Type      NotificationType     `json:"type" gorm:"type:varchar(20);not null;default:'info'"`
	Priority  NotificationPriority `json:"priority" gorm:"type:varchar(20);not null;default:'normal'"`
	Title     string               `json:"title" gorm:"type:varchar(255);not null"`
	Message   string               `json:"message" gorm:"type:text;not null"`
	Data      *string              `json:"data,omitempty" gorm:"type:jsonb"` // Additional JSON data
	IsRead    bool                 `json:"is_read" gorm:"default:false;index"`
	ReadAt    *time.Time           `json:"read_at,omitempty"`
	Link      *string              `json:"link,omitempty" gorm:"type:varchar(500)"` // Optional action link
	ImageURL  *string              `json:"image_url,omitempty" gorm:"type:varchar(500)"`
	CreatedAt time.Time            `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time            `json:"updated_at" gorm:"autoUpdateTime"`
	ExpiresAt *time.Time           `json:"expires_at,omitempty" gorm:"index"` // Optional expiration

	// Relations
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
}

// TableName specifies the table name for Notification
func (Notification) TableName() string {
	return "notifications"
}

// IsExpired checks if the notification has expired
func (n *Notification) IsExpired() bool {
	if n.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*n.ExpiresAt)
}

// MarkAsRead marks the notification as read
func (n *Notification) MarkAsRead() {
	now := time.Now()
	n.IsRead = true
	n.ReadAt = &now
}

// MarkAsUnread marks the notification as unread
func (n *Notification) MarkAsUnread() {
	n.IsRead = false
	n.ReadAt = nil
}

// CreateNotificationRequest represents the request to create a notification
type CreateNotificationRequest struct {
	UserID    *uuid.UUID           `json:"user_id,omitempty"` // nil for broadcast
	Type      NotificationType     `json:"type" binding:"required,oneof=info success warning error"`
	Priority  NotificationPriority `json:"priority" binding:"omitempty,oneof=low normal high urgent"`
	Title     string               `json:"title" binding:"required,max=255"`
	Message   string               `json:"message" binding:"required"`
	Data      *string              `json:"data,omitempty"`
	Link      *string              `json:"link,omitempty"`
	ImageURL  *string              `json:"image_url,omitempty"`
	ExpiresAt *time.Time           `json:"expires_at,omitempty"`
}

// UpdateNotificationRequest represents the request to update a notification
type UpdateNotificationRequest struct {
	Type      *NotificationType     `json:"type,omitempty" binding:"omitempty,oneof=info success warning error"`
	Priority  *NotificationPriority `json:"priority,omitempty" binding:"omitempty,oneof=low normal high urgent"`
	Title     *string               `json:"title,omitempty" binding:"omitempty,max=255"`
	Message   *string               `json:"message,omitempty"`
	Data      *string               `json:"data,omitempty"`
	Link      *string               `json:"link,omitempty"`
	ImageURL  *string               `json:"image_url,omitempty"`
	IsRead    *bool                 `json:"is_read,omitempty"`
	ExpiresAt *time.Time            `json:"expires_at,omitempty"`
}

// NotificationFilter represents filters for querying notifications
type NotificationFilter struct {
	UserID   *uuid.UUID            `json:"user_id,omitempty"`
	Type     *NotificationType     `json:"type,omitempty"`
	Priority *NotificationPriority `json:"priority,omitempty"`
	IsRead   *bool                 `json:"is_read,omitempty"`
	FromDate *time.Time            `json:"from_date,omitempty"`
	ToDate   *time.Time            `json:"to_date,omitempty"`
}

// NotificationStats represents notification statistics
type NotificationStats struct {
	Total      int64                          `json:"total"`
	Unread     int64                          `json:"unread"`
	Read       int64                          `json:"read"`
	ByType     map[NotificationType]int64     `json:"by_type"`
	ByPriority map[NotificationPriority]int64 `json:"by_priority"`
}
