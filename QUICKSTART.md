# 🚀 Quick Start Guide - GO CMS

## ✅ Trạng Thái: SẴN SÀNG CHẠY!

Hệ thống CRM với Go backend đã được implement với đầy đủ tính năng authentication.

## 📋 Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Make (optional)

## 🏃 Quick Start

### 1. Clone & Setup
```bash
cd GO_CMS
cp .env.example .env
```

### 2. Start Infrastructure
```bash
docker-compose up -d
```

Điều này sẽ khởi động:
- PostgreSQL (port 5432)
- Redis (port 6379)
- MinIO (port 9000, console: 9001)

### 3. Install Dependencies
```bash
go mod download
```

### 4. Run Application
```bash
go run cmd/server/main.go
```

Hoặc sử dụng Make:
```bash
make run
```

## 🌐 Access Points

- **API**: http://localhost:8080
- **Health Check**: http://localhost:8080/health
- **Swagger Docs**: http://localhost:8080/swagger/index.html
- **MinIO Console**: http://localhost:9001 (minioadmin/minioadmin)

## 🧪 Test Authentication Flow

### 1. Register User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!",
    "first_name": "Test",
    "last_name": "User"
  }'
```

**Response**: Bạn sẽ nhận được access_token và refresh_token. OTP sẽ được in ra trong logs.

### 2. Verify Email
Kiểm tra logs để lấy OTP code, sau đó:
```bash
curl -X POST http://localhost:8080/api/v1/auth/verify-email \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "code": "YOUR_OTP_CODE"
  }'
```

### 3. Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!"
  }'
```

### 4. Get Current User
```bash
curl -X GET http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## 📚 Available Endpoints

### Public (No Auth)
- `POST /api/v1/auth/register` - Register
- `POST /api/v1/auth/login` - Login
- `POST /api/v1/auth/verify-email` - Verify Email
- `POST /api/v1/auth/resend-otp` - Resend OTP
- `POST /api/v1/auth/refresh` - Refresh Token
- `POST /api/v1/auth/forgot-password` - Forgot Password
- `POST /api/v1/auth/reset-password` - Reset Password

### Protected (Auth Required)
- `POST /api/v1/auth/logout` - Logout
- `POST /api/v1/auth/change-password` - Change Password
- `GET /api/v1/auth/me` - Get Current User
- `PUT /api/v1/auth/me` - Update Profile
- `POST /api/v1/auth/2fa/enable` - Enable 2FA
- `POST /api/v1/auth/2fa/verify` - Verify 2FA
- `POST /api/v1/auth/2fa/disable` - Disable 2FA

## 🎯 Features Implemented

✅ **Authentication**
- Email/Password registration
- Email verification with OTP
- Login with 2FA support
- JWT access + refresh tokens
- Password management (forgot, reset, change)

✅ **Authorization**
- Hierarchical RBAC (5 levels)
- Permission-based access control
- Redis caching for permissions

✅ **Infrastructure**
- PostgreSQL with auto-migration
- Redis caching
- MinIO file storage
- Graceful shutdown
- Structured logging

✅ **Performance**
- Cursor-based pagination
- Connection pooling
- Database indexing
- Redis caching

## 📖 Documentation

- **Full Summary**: See `FINAL_SUMMARY.md`
- **Implementation Plan**: See `.agent/workflows/implementation-plan.md`
- **Progress**: See `PROGRESS.md`
- **Swagger**: http://localhost:8080/swagger/index.html (after running)

## 🛠️ Development

### Run Tests
```bash
make test
```

### Generate Swagger Docs
```bash
make swagger
```

### Database Migrations
```bash
make migrate-up
make migrate-down
```

### Code Formatting
```bash
make fmt
make lint
```

## 🗂️ Project Structure

```
GO_CMS/
├── cmd/server/          # Application entry point
├── internal/
│   ├── adapters/        # Handlers & Repositories
│   ├── core/           # Domain & Use Cases
│   ├── infrastructure/ # DB, Cache, Storage, Middleware
│   └── config/         # Configuration
├── pkg/                # Shared utilities
├── docs/               # Swagger documentation
└── migrations/         # Database migrations
```

## 🔐 Default Credentials

**Super Admin** (created via seed data):
- Tạo user với email bất kỳ
- Verify email
- Assign role `super_admin` qua database hoặc API

## 🐛 Troubleshooting

### Database Connection Error
```bash
# Check if PostgreSQL is running
docker-compose ps

# View logs
docker-compose logs postgres
```

### Redis Connection Error
```bash
# Check if Redis is running
docker-compose ps

# View logs
docker-compose logs redis
```

### Port Already in Use
```bash
# Change ports in .env file
SERVER_PORT=8081
```

## 📝 Environment Variables

Key variables in `.env`:
```env
# Server
SERVER_PORT=8080

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=go_cms

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# MinIO
MINIO_ENDPOINT=localhost:9000

# JWT
JWT_SECRET=your-secret-key
JWT_ACCESS_TOKEN_EXPIRE=15m
JWT_REFRESH_TOKEN_EXPIRE=7d
```

## 🎓 Tech Stack

- **Language**: Go 1.21+
- **Framework**: Gin
- **ORM**: GORM
- **Database**: PostgreSQL 15+
- **Cache**: Redis 7+
- **Storage**: MinIO
- **Logging**: Zap
- **Documentation**: Swagger

## 📞 Support

For issues or questions, check:
1. `FINAL_SUMMARY.md` - Complete feature list
2. Logs in console
3. Swagger documentation
4. Source code comments

## 🎉 Success!

If you see:
```
INFO    Server starting {"address": "0.0.0.0:8080", "environment": ""}
```

Your application is running successfully! 🚀

Visit http://localhost:8080/health to verify.
