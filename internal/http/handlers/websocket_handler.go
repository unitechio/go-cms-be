package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports"
	"go.uber.org/zap"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Implement proper origin checking in production
		return true
	},
}

type WebSocketHandler struct {
	wsService ports.WebSocketService
	logger    *zap.Logger
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(wsService ports.WebSocketService, logger *zap.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		wsService: wsService,
		logger:    logger,
	}
}

// HandleWebSocket handles WebSocket connections
// @Summary WebSocket connection endpoint
// @Description Establish a WebSocket connection for real-time notifications
// @Tags WebSocket
// @Security BearerAuth
// @Param user_id query string true "User ID"
// @Success 101 {string} string "Switching Protocols"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /ws [get]
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// Get user ID from query or context (set by auth middleware)
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		// Try to get from context (set by auth middleware)
		if userID, exists := c.Get("user_id"); exists {
			if uid, ok := userID.(uuid.UUID); ok {
				userIDStr = uid.String()
			}
		}
	}

	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id format"})
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade connection",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)
		return
	}

	// Create client
	client := &domain.WebSocketClient{
		ID:         uuid.New().String(),
		UserID:     userID,
		Connection: conn,
		Send:       make(chan []byte, 256),
		CreatedAt:  time.Now(),
	}

	// Register client
	h.wsService.RegisterClient(client)

	// Start goroutines for reading and writing
	go h.writePump(client, conn)
	go h.readPump(client, conn)
}

// readPump pumps messages from the WebSocket connection to the hub
func (h *WebSocketHandler) readPump(client *domain.WebSocketClient, conn *websocket.Conn) {
	defer func() {
		h.wsService.UnregisterClient(client.ID)
		conn.Close()
	}()

	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetReadLimit(maxMessageSize)
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				h.logger.Error("WebSocket read error",
					zap.Error(err),
					zap.String("client_id", client.ID),
				)
			}
			break
		}

		// Handle incoming messages (e.g., ping/pong, acknowledgments)
		h.logger.Debug("Received WebSocket message",
			zap.String("client_id", client.ID),
			zap.ByteString("message", message),
		)

		// You can add custom message handling here
		// For example, handle ping messages from client
	}
}

// writePump pumps messages from the hub to the WebSocket connection
func (h *WebSocketHandler) writePump(client *domain.WebSocketClient, conn *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.Send:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(client.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-client.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// GetOnlineUsers returns the list of currently online users
// @Summary Get online users
// @Description Get the list of currently online users
// @Tags WebSocket
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "Online users"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /ws/online-users [get]
func (h *WebSocketHandler) GetOnlineUsers(c *gin.Context) {
	users := h.wsService.GetOnlineUsers()
	c.JSON(http.StatusOK, gin.H{
		"online_users": users,
		"count":        len(users),
	})
}

// GetConnectionStats returns WebSocket connection statistics
// @Summary Get connection statistics
// @Description Get WebSocket connection statistics
// @Tags WebSocket
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "Connection statistics"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /ws/stats [get]
func (h *WebSocketHandler) GetConnectionStats(c *gin.Context) {
	totalConnections := h.wsService.GetTotalConnections()
	onlineUsers := h.wsService.GetOnlineUsers()

	c.JSON(http.StatusOK, gin.H{
		"total_connections": totalConnections,
		"online_users":      len(onlineUsers),
		"users":             onlineUsers,
	})
}

// BroadcastMessage broadcasts a message to all connected clients (admin only)
// @Summary Broadcast message
// @Description Broadcast a message to all connected clients (admin only)
// @Tags WebSocket
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param message body domain.SystemMessage true "System message"
// @Success 200 {object} map[string]interface{} "Message broadcasted"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Router /ws/broadcast [post]
func (h *WebSocketHandler) BroadcastMessage(c *gin.Context) {
	var req domain.SystemMessage
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Timestamp = time.Now()

	message := &domain.WebSocketMessage{
		Type:    domain.WSEventSystemMessage,
		Payload: req,
	}

	if err := h.wsService.Broadcast(message); err != nil {
		h.logger.Error("Failed to broadcast message",
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to broadcast message"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Message broadcasted successfully",
		"sent_to": h.wsService.GetTotalConnections(),
	})
}
