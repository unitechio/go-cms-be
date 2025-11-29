package ports

import (
	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
)

// WebSocketService defines the interface for WebSocket operations
type WebSocketService interface {
	// RegisterClient registers a new WebSocket client
	RegisterClient(client *domain.WebSocketClient)

	// UnregisterClient unregisters a WebSocket client
	UnregisterClient(clientID string)

	// SendToUser sends a message to a specific user (all their connections)
	SendToUser(userID uuid.UUID, message *domain.WebSocketMessage) error

	// SendToClient sends a message to a specific client connection
	SendToClient(clientID string, message *domain.WebSocketMessage) error

	// Broadcast sends a message to all connected clients
	Broadcast(message *domain.WebSocketMessage) error

	// BroadcastExcept sends a message to all clients except the specified one
	BroadcastExcept(excludeClientID string, message *domain.WebSocketMessage) error

	// GetOnlineUsers returns the list of currently online users
	GetOnlineUsers() []uuid.UUID

	// IsUserOnline checks if a user is currently online
	IsUserOnline(userID uuid.UUID) bool

	// GetUserConnections returns the number of active connections for a user
	GetUserConnections(userID uuid.UUID) int

	// GetTotalConnections returns the total number of active connections
	GetTotalConnections() int

	// NotifyNotification sends a notification event to user(s)
	NotifyNotification(notification *domain.Notification, action string) error
}
