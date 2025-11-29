# 🔔 Notification & Realtime System - Usage Guide

## 📖 Tổng quan

Hệ thống Notification và Realtime đã được tích hợp thành công vào Go CMS Backend với các tính năng:

- ✅ **Notification System**: Quản lý thông báo cho người dùng
- ✅ **WebSocket Realtime**: Gửi thông báo real-time qua WebSocket
- ✅ **User Presence**: Theo dõi trạng thái online/offline của người dùng
- ✅ **Broadcasting**: Gửi thông báo đến tất cả người dùng

## 🚀 Khởi động Server

```bash
# Development mode
make dev

# Production mode
make run
```

## 📡 WebSocket Connection

### Kết nối WebSocket

```javascript
// Frontend JavaScript example
const token = 'your-jwt-token';
const ws = new WebSocket(`ws://localhost:8080/api/v1/ws?user_id=${userId}`);

// Hoặc với Authorization header (recommended)
const ws = new WebSocket('ws://localhost:8080/api/v1/ws');
// Gửi token sau khi kết nối
ws.onopen = () => {
  ws.send(JSON.stringify({
    type: 'auth',
    token: token
  }));
};

// Nhận messages
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Received:', data);
  
  switch(data.type) {
    case 'notification':
      handleNotification(data.payload);
      break;
    case 'user_online':
      handleUserOnline(data.payload);
      break;
    case 'user_offline':
      handleUserOffline(data.payload);
      break;
  }
};
```

### WebSocket Event Types

```typescript
type WebSocketEventType = 
  | 'notification'          // Thông báo mới
  | 'notification_read'     // Thông báo đã đọc
  | 'user_online'           // User online
  | 'user_offline'          // User offline
  | 'system_message'        // Thông báo hệ thống
  | 'ping'                  // Heartbeat ping
  | 'pong';                 // Heartbeat pong
```

## 🔔 Notification API

### 1. Tạo Notification

```bash
POST /api/v1/notifications
Authorization: Bearer {token}
Content-Type: application/json

{
  "user_id": "uuid-here",  // null để broadcast
  "type": "info",          // info, success, warning, error
  "priority": "normal",    // low, normal, high, urgent
  "title": "Welcome!",
  "message": "Welcome to our system",
  "data": "{\"key\": \"value\"}",  // Optional JSON data
  "link": "/dashboard",    // Optional action link
  "image_url": "https://...",  // Optional image
  "expires_at": "2024-12-31T23:59:59Z"  // Optional expiration
}
```

### 2. Lấy Notifications của User

```bash
GET /api/v1/notifications/me?after=123&limit=20
Authorization: Bearer {token}
```

Response:
```json
{
  "data": [
    {
      "id": 1,
      "user_id": "uuid",
      "type": "info",
      "priority": "normal",
      "title": "Welcome!",
      "message": "Welcome to our system",
      "is_read": false,
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "next_cursor": {
    "id": 123,
    "has_more": true
  }
}
```

### 3. Đánh dấu đã đọc

```bash
POST /api/v1/notifications/{id}/read
Authorization: Bearer {token}
```

### 4. Đánh dấu tất cả đã đọc

```bash
POST /api/v1/notifications/mark-all-read
Authorization: Bearer {token}
```

### 5. Lấy số lượng chưa đọc

```bash
GET /api/v1/notifications/unread-count
Authorization: Bearer {token}
```

Response:
```json
{
  "unread_count": 5
}
```

### 6. Lấy thống kê

```bash
GET /api/v1/notifications/stats
Authorization: Bearer {token}
```

Response:
```json
{
  "total": 100,
  "unread": 5,
  "read": 95,
  "by_type": {
    "info": 50,
    "success": 30,
    "warning": 15,
    "error": 5
  },
  "by_priority": {
    "low": 20,
    "normal": 60,
    "high": 15,
    "urgent": 5
  }
}
```

### 7. Broadcast Notification (Admin only)

```bash
POST /api/v1/notifications/broadcast
Authorization: Bearer {admin-token}
Content-Type: application/json

{
  "type": "warning",
  "priority": "high",
  "title": "System Maintenance",
  "message": "System will be down for maintenance at 2 AM"
}
```

## 👥 User Presence API

### 1. Lấy danh sách users online

```bash
GET /api/v1/ws/online-users
Authorization: Bearer {token}
```

Response:
```json
{
  "online_users": [
    "uuid-1",
    "uuid-2",
    "uuid-3"
  ],
  "count": 3
}
```

### 2. Lấy thống kê kết nối

```bash
GET /api/v1/ws/stats
Authorization: Bearer {token}
```

Response:
```json
{
  "total_connections": 10,
  "online_users": 5,
  "users": ["uuid-1", "uuid-2", ...]
}
```

### 3. Broadcast System Message (Admin only)

```bash
POST /api/v1/ws/broadcast
Authorization: Bearer {admin-token}
Content-Type: application/json

{
  "message": "Server will restart in 5 minutes",
  "type": "warning"
}
```

## 💡 Use Cases

### 1. Gửi thông báo khi user đăng ký

```go
notification := &domain.CreateNotificationRequest{
    UserID:   &newUser.ID,
    Type:     domain.NotificationTypeSuccess,
    Priority: domain.NotificationPriorityNormal,
    Title:    "Welcome!",
    Message:  "Your account has been created successfully",
    Link:     utils.StringPtr("/dashboard"),
}

notificationUseCase.CreateNotification(ctx, notification)
// Notification sẽ tự động được gửi real-time qua WebSocket
```

### 2. Gửi thông báo khi có order mới

```go
notification := &domain.CreateNotificationRequest{
    UserID:   &adminUserID,
    Type:     domain.NotificationTypeInfo,
    Priority: domain.NotificationPriorityHigh,
    Title:    "New Order",
    Message:  fmt.Sprintf("Order #%s has been placed", orderID),
    Link:     utils.StringPtr(fmt.Sprintf("/orders/%s", orderID)),
    Data:     utils.StringPtr(fmt.Sprintf(`{"order_id": "%s"}`, orderID)),
}

notificationUseCase.CreateNotification(ctx, notification)
```

### 3. Gửi thông báo hệ thống

```go
notification := &domain.CreateNotificationRequest{
    UserID:   nil,  // Broadcast to all users
    Type:     domain.NotificationTypeWarning,
    Priority: domain.NotificationPriorityUrgent,
    Title:    "System Maintenance",
    Message:  "System will be down for maintenance",
    ExpiresAt: utils.TimePtr(time.Now().Add(24 * time.Hour)),
}

notificationUseCase.BroadcastNotification(ctx, notification)
```

## 🔧 Configuration

### Environment Variables

```env
# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Database (notifications table will be auto-migrated)
DB_HOST=localhost
DB_PORT=5432
DB_NAME=go_cms
DB_USER=postgres
DB_PASSWORD=postgres
```

### WebSocket Configuration

WebSocket Hub tự động khởi động khi server start:
- Ping/Pong heartbeat: 54 seconds
- Write timeout: 10 seconds
- Read timeout: 60 seconds
- Max message size: 512 bytes

## 📊 Database Schema

### Notifications Table

```sql
CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id),  -- NULL for broadcast
    type VARCHAR(20) NOT NULL DEFAULT 'info',
    priority VARCHAR(20) NOT NULL DEFAULT 'normal',
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    data JSONB,
    is_read BOOLEAN DEFAULT FALSE,
    read_at TIMESTAMP,
    link VARCHAR(500),
    image_url VARCHAR(500),
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_is_read ON notifications(is_read);
CREATE INDEX idx_notifications_created_at ON notifications(created_at DESC);
CREATE INDEX idx_notifications_user_read ON notifications(user_id, is_read, created_at DESC);
```

## 🎯 Frontend Integration Example

### React Hook

```typescript
import { useEffect, useState } from 'react';

export function useNotifications() {
  const [notifications, setNotifications] = useState([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const [ws, setWs] = useState<WebSocket | null>(null);

  useEffect(() => {
    // Fetch initial notifications
    fetch('/api/v1/notifications/me', {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })
    .then(res => res.json())
    .then(data => setNotifications(data.data));

    // Fetch unread count
    fetch('/api/v1/notifications/unread-count', {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })
    .then(res => res.json())
    .then(data => setUnreadCount(data.unread_count));

    // Connect WebSocket
    const websocket = new WebSocket('ws://localhost:8080/api/v1/ws');
    
    websocket.onmessage = (event) => {
      const data = JSON.parse(event.data);
      
      if (data.type === 'notification') {
        const notification = data.payload.notification;
        setNotifications(prev => [notification, ...prev]);
        setUnreadCount(prev => prev + 1);
        
        // Show toast notification
        showToast(notification.title, notification.message);
      }
    };

    setWs(websocket);

    return () => {
      websocket.close();
    };
  }, []);

  const markAsRead = async (id: number) => {
    await fetch(`/api/v1/notifications/${id}/read`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
    
    setNotifications(prev =>
      prev.map(n => n.id === id ? { ...n, is_read: true } : n)
    );
    setUnreadCount(prev => prev - 1);
  };

  return { notifications, unreadCount, markAsRead };
}
```

## 🐛 Troubleshooting

### WebSocket không kết nối được

1. Kiểm tra CORS configuration
2. Kiểm tra firewall/proxy settings
3. Kiểm tra authentication token

### Notifications không gửi real-time

1. Kiểm tra WebSocket Hub đã start chưa
2. Kiểm tra user đã kết nối WebSocket chưa
3. Check logs để xem có lỗi gì không

### Performance issues

1. Sử dụng pagination khi fetch notifications
2. Cleanup expired notifications định kỳ:
   ```go
   notificationUseCase.CleanupExpiredNotifications(ctx)
   ```
3. Limit số lượng notifications per user

## 📚 API Documentation

Swagger documentation có sẵn tại: `http://localhost:8080/swagger/index.html`

## 🎉 Hoàn thành!

Hệ thống Notification và Realtime đã sẵn sàng sử dụng. Chúc bạn code vui vẻ! 🚀
