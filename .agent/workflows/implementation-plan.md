---
description: Go CRM System Implementation Plan
---

# Go CRM System - Implementation Plan

## Architecture Overview
- **Pattern**: Clean Architecture + Domain-Driven Design
- **Framework**: Gin (HTTP Router)
- **ORM**: GORM
- **Database**: PostgreSQL (with partitioning & indexing)
- **Cache**: Redis
- **Storage**: MinIO
- **Documentation**: Swagger/OpenAPI

## Project Structure
```
GO_CMS/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── core/                       # Business logic layer
│   │   ├── domain/                 # Domain models & interfaces
│   │   │   ├── user.go
│   │   │   ├── customer.go
│   │   │   ├── role.go
│   │   │   ├── permission.go
│   │   │   ├── post.go
│   │   │   ├── media.go
│   │   │   └── audit_log.go
│   │   ├── ports/                  # Repository & Service interfaces
│   │   │   ├── repositories/
│   │   │   └── services/
│   │   └── usecases/               # Business use cases
│   │       ├── auth/
│   │       ├── user/
│   │       ├── customer/
│   │       ├── permission/
│   │       ├── post/
│   │       └── media/
│   ├── adapters/                   # External adapters
│   │   ├── handlers/               # HTTP handlers
│   │   │   ├── auth_handler.go
│   │   │   ├── user_handler.go
│   │   │   ├── customer_handler.go
│   │   │   ├── permission_handler.go
│   │   │   ├── post_handler.go
│   │   │   └── media_handler.go
│   │   ├── repositories/           # Database implementations
│   │   │   ├── postgres/
│   │   │   └── redis/
│   │   └── external/               # External services
│   │       ├── minio/
│   │       └── email/
│   ├── infrastructure/             # Infrastructure layer
│   │   ├── database/
│   │   │   ├── postgres.go
│   │   │   └── migrations/
│   │   ├── cache/
│   │   │   └── redis.go
│   │   ├── storage/
│   │   │   └── minio.go
│   │   ├── middleware/
│   │   │   ├── timeout.go
│   │   │   ├── cors.go
│   │   │   ├── auth.go
│   │   │   ├── authorize.go
│   │   │   ├── logger.go
│   │   │   └── recovery.go
│   │   └── router/
│   │       └── router.go
│   └── config/
│       └── config.go               # Configuration management
├── pkg/                            # Shared packages
│   ├── logger/
│   ├── errors/
│   ├── response/
│   ├── pagination/
│   ├── validator/
│   └── utils/
├── docs/                           # Swagger documentation
├── migrations/                     # Database migrations
├── scripts/                        # Utility scripts
├── .env.example
├── .gitignore
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Implementation Phases

### Phase 1: Project Setup & Infrastructure
1. Initialize Go module
2. Setup configuration management
3. Database connection (PostgreSQL)
4. Redis connection
5. MinIO setup
6. Logger setup
7. Error handling utilities
8. Response utilities

### Phase 2: Core Domain Models
1. User domain (with 2FA, OTP)
2. Customer domain
3. Role & Permission domain (hierarchical RBAC)
4. Post domain
5. Media domain
6. Audit log domain

### Phase 3: Middleware Implementation
1. Timeout middleware
2. CORS middleware
3. Authentication middleware (JWT)
4. Authorization middleware (permission-based)
5. Logger middleware
6. Recovery middleware

### Phase 4: Authentication & Authorization System
1. Email/Password authentication
2. OTP generation & verification
3. 2FA (TOTP) implementation
4. JWT token management
5. Refresh token mechanism
6. Hierarchical permission system:
   - Organization level
   - Department level
   - Service level
   - Action level

### Phase 5: User & Customer Management
1. User CRUD operations
2. Customer CRUD operations
3. User-Role assignment
4. Permission management
5. Audit logging

### Phase 6: Content Management
1. Post CRUD operations
2. Media upload (images, documents)
3. Content scheduling
4. Job queue for scheduled posts

### Phase 7: Advanced Features
1. Cursor-based pagination
2. Advanced filtering & search
3. Database indexing strategy
4. Table partitioning
5. Redis caching layer
6. File upload handling (multiple types)

### Phase 8: Documentation & Testing
1. Swagger API documentation
2. API handler comments
3. Unit tests
4. Integration tests

## Key Technical Requirements

### Database Optimization
- **Indexing**: Composite indexes on frequently queried columns
- **Partitioning**: Time-based partitioning for audit logs and posts
- **Transactions**: Full ACID compliance with proper transaction management
- **Connection Pooling**: Optimized connection pool settings

### Caching Strategy
- **User sessions**: Redis with TTL
- **Permission cache**: Redis with invalidation on updates
- **Query results**: Selective caching for expensive queries

### Pagination
- **Cursor-based**: Using encoded cursor for efficient pagination
- **Configurable page size**: With max limits

### Security
- **Password hashing**: bcrypt
- **JWT**: RS256 algorithm
- **2FA**: TOTP (Time-based One-Time Password)
- **OTP**: Random 6-digit with expiration
- **Rate limiting**: Per endpoint and per user

### Logging
- **Structured logging**: JSON format for ElasticSearch
- **Log levels**: DEBUG, INFO, WARN, ERROR
- **Request tracing**: Correlation ID for request tracking
- **Audit logs**: All sensitive operations logged

### File Upload
- **Supported types**: Images (jpg, png, gif), Documents (pdf, docx, xlsx), Videos (mp4)
- **Size limits**: Configurable per file type
- **Virus scanning**: Optional integration
- **CDN integration**: MinIO with presigned URLs

## Development Workflow
1. Create domain models
2. Define repository interfaces
3. Implement repositories
4. Create use cases
5. Implement handlers
6. Add middleware
7. Write tests
8. Generate Swagger docs
9. Optimize queries
10. Add caching

## Dependencies
```
- github.com/gin-gonic/gin
- gorm.io/gorm
- gorm.io/driver/postgres
- github.com/redis/go-redis/v9
- github.com/minio/minio-go/v7
- github.com/golang-jwt/jwt/v5
- github.com/pquerna/otp
- github.com/swaggo/swag
- github.com/swaggo/gin-swagger
- go.uber.org/zap (logging)
- github.com/spf13/viper (config)
- golang.org/x/crypto (bcrypt)
```
