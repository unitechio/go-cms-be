package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports"
	"go.uber.org/zap"
)

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients mapped by client ID
	clients map[string]*domain.WebSocketClient

	// User ID to client IDs mapping (one user can have multiple connections)
	userClients map[uuid.UUID]map[string]bool

	// Register requests from the clients
	register chan *domain.WebSocketClient

	// Unregister requests from clients
	unregister chan string

	// Broadcast messages to all clients
	broadcast chan []byte

	// Send message to specific user
	sendToUser chan *userMessage

	// Send message to specific client
	sendToClient chan *clientMessage

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Logger
	logger *zap.Logger
}

type userMessage struct {
	userID  uuid.UUID
	message []byte
}

type clientMessage struct {
	clientID string
	message  []byte
}

// NewHub creates a new WebSocket hub
func NewHub(logger *zap.Logger) ports.WebSocketService {
	return &Hub{
		clients:      make(map[string]*domain.WebSocketClient),
		userClients:  make(map[uuid.UUID]map[string]bool),
		register:     make(chan *domain.WebSocketClient),
		unregister:   make(chan string),
		broadcast:    make(chan []byte),
		sendToUser:   make(chan *userMessage),
		sendToClient: make(chan *clientMessage),
		logger:       logger,
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client

			// Add to user clients mapping
			if h.userClients[client.UserID] == nil {
				h.userClients[client.UserID] = make(map[string]bool)
			}
			h.userClients[client.UserID][client.ID] = true

			h.mu.Unlock()

			h.logger.Info("Client registered",
				zap.String("client_id", client.ID),
				zap.String("user_id", client.UserID.String()),
			)

			// Broadcast user online status
			h.broadcastUserPresence(client.UserID, "online")

		case clientID := <-h.unregister:
			h.mu.Lock()
			if client, ok := h.clients[clientID]; ok {
				// Remove from user clients mapping
				if userClients, exists := h.userClients[client.UserID]; exists {
					delete(userClients, clientID)
					if len(userClients) == 0 {
						delete(h.userClients, client.UserID)
						// Broadcast user offline status only if no more connections
						go h.broadcastUserPresence(client.UserID, "offline")
					}
				}

				close(client.Send)
				delete(h.clients, clientID)

				h.logger.Info("Client unregistered",
					zap.String("client_id", clientID),
					zap.String("user_id", client.UserID.String()),
				)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.Send <- message:
				default:
					// Client's send channel is full, skip
					h.logger.Warn("Client send channel full",
						zap.String("client_id", client.ID),
					)
				}
			}
			h.mu.RUnlock()

		case msg := <-h.sendToUser:
			h.mu.RLock()
			if clientIDs, ok := h.userClients[msg.userID]; ok {
				for clientID := range clientIDs {
					if client, exists := h.clients[clientID]; exists {
						select {
						case client.Send <- msg.message:
						default:
							h.logger.Warn("Client send channel full",
								zap.String("client_id", clientID),
							)
						}
					}
				}
			}
			h.mu.RUnlock()

		case msg := <-h.sendToClient:
			h.mu.RLock()
			if client, ok := h.clients[msg.clientID]; ok {
				select {
				case client.Send <- msg.message:
				default:
					h.logger.Warn("Client send channel full",
						zap.String("client_id", msg.clientID),
					)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// RegisterClient registers a new WebSocket client
func (h *Hub) RegisterClient(client *domain.WebSocketClient) {
	h.register <- client
}

// UnregisterClient unregisters a WebSocket client
func (h *Hub) UnregisterClient(clientID string) {
	h.unregister <- clientID
}

// SendToUser sends a message to a specific user (all their connections)
func (h *Hub) SendToUser(userID uuid.UUID, message *domain.WebSocketMessage) error {
	message.Timestamp = time.Now()
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.sendToUser <- &userMessage{
		userID:  userID,
		message: data,
	}
	return nil
}

// SendToClient sends a message to a specific client connection
func (h *Hub) SendToClient(clientID string, message *domain.WebSocketMessage) error {
	message.Timestamp = time.Now()
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.sendToClient <- &clientMessage{
		clientID: clientID,
		message:  data,
	}
	return nil
}

// Broadcast sends a message to all connected clients
func (h *Hub) Broadcast(message *domain.WebSocketMessage) error {
	message.Timestamp = time.Now()
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.broadcast <- data
	return nil
}

// BroadcastExcept sends a message to all clients except the specified one
func (h *Hub) BroadcastExcept(excludeClientID string, message *domain.WebSocketMessage) error {
	message.Timestamp = time.Now()
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for clientID, client := range h.clients {
		if clientID != excludeClientID {
			select {
			case client.Send <- data:
			default:
				h.logger.Warn("Client send channel full",
					zap.String("client_id", clientID),
				)
			}
		}
	}
	return nil
}

// GetOnlineUsers returns the list of currently online users
func (h *Hub) GetOnlineUsers() []uuid.UUID {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]uuid.UUID, 0, len(h.userClients))
	for userID := range h.userClients {
		users = append(users, userID)
	}
	return users
}

// IsUserOnline checks if a user is currently online
func (h *Hub) IsUserOnline(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()

	_, exists := h.userClients[userID]
	return exists
}

// GetUserConnections returns the number of active connections for a user
func (h *Hub) GetUserConnections(userID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if clients, ok := h.userClients[userID]; ok {
		return len(clients)
	}
	return 0
}

// GetTotalConnections returns the total number of active connections
func (h *Hub) GetTotalConnections() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.clients)
}

// NotifyNotification sends a notification event to user(s)
func (h *Hub) NotifyNotification(notification *domain.Notification, action string) error {
	event := &domain.NotificationEvent{
		Notification: notification,
		Action:       action,
	}

	message := &domain.WebSocketMessage{
		Type:    domain.WSEventNotification,
		Payload: event,
	}

	// If notification is for a specific user, send only to that user
	if notification.UserID != nil {
		return h.SendToUser(*notification.UserID, message)
	}

	// Otherwise, broadcast to all users
	return h.Broadcast(message)
}

// broadcastUserPresence broadcasts user online/offline status
func (h *Hub) broadcastUserPresence(userID uuid.UUID, status string) {
	presence := &domain.UserPresence{
		UserID:    userID,
		Status:    status,
		Timestamp: time.Now(),
	}

	var eventType domain.WebSocketEventType
	if status == "online" {
		eventType = domain.WSEventUserOnline
	} else {
		eventType = domain.WSEventUserOffline
	}

	message := &domain.WebSocketMessage{
		Type:    eventType,
		Payload: presence,
	}

	if err := h.Broadcast(message); err != nil {
		h.logger.Error("Failed to broadcast user presence",
			zap.String("user_id", userID.String()),
			zap.String("status", status),
			zap.Error(err),
		)
	}
}
