package domain

import (
	"time"

	"github.com/google/uuid"
)

// WebSocketEventType represents the type of WebSocket event
type WebSocketEventType string

const (
	WSEventNotification     WebSocketEventType = "notification"
	WSEventNotificationRead WebSocketEventType = "notification_read"
	WSEventUserOnline       WebSocketEventType = "user_online"
	WSEventUserOffline      WebSocketEventType = "user_offline"
	WSEventSystemMessage    WebSocketEventType = "system_message"
	WSEventPing             WebSocketEventType = "ping"
	WSEventPong             WebSocketEventType = "pong"
)

// WebSocketMessage represents a message sent through WebSocket
type WebSocketMessage struct {
	Type      WebSocketEventType `json:"type"`
	Payload   interface{}        `json:"payload"`
	Timestamp time.Time          `json:"timestamp"`
	MessageID string             `json:"message_id,omitempty"`
}

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	ID         string
	UserID     uuid.UUID
	Connection interface{} // Will be *websocket.Conn
	Send       chan []byte
	CreatedAt  time.Time
}

// NotificationEvent represents a notification event for WebSocket
type NotificationEvent struct {
	Notification *Notification `json:"notification"`
	Action       string        `json:"action"` // created, updated, deleted, read
}

// SystemMessage represents a system-wide message
type SystemMessage struct {
	Message   string    `json:"message"`
	Type      string    `json:"type"` // info, warning, error
	Timestamp time.Time `json:"timestamp"`
}

// UserPresence represents user online/offline status
type UserPresence struct {
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username,omitempty"`
	Status    string    `json:"status"` // online, offline
	Timestamp time.Time `json:"timestamp"`
}
