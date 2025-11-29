# GO CMS - Enterprise CRM System

A comprehensive CRM (Customer Relationship Management) system built with Go, following Clean Architecture and Domain-Driven Design principles.

## Features

### 🔐 Authentication & Authorization
- **Multi-factor Authentication**: Email/Password, OTP, and 2FA (TOTP)
- **Hierarchical RBAC**: Deep permission system at organization, department, service, and action levels
- **JWT-based**: Secure token-based authentication with refresh tokens
- **Session Management**: Redis-backed session storage

### 👥 User & Customer Management
- Complete user lifecycle management
- Customer relationship tracking
- Role and permission assignment
- Activity and audit logging

### 📝 Content Management
- Post creation and management
- Media upload and storage (MinIO)
- Content scheduling with job queue
- Category and tag system

### 🚀 Performance & Scalability
- **Cursor-based Pagination**: Efficient data retrieval for large datasets
- **Database Optimization**: Composite indexes and table partitioning
- **Redis Caching**: Intelligent caching for permissions and frequently accessed data
- **Connection Pooling**: Optimized database connections

### 📊 Monitoring & Logging
- **Structured Logging**: JSON format for ElasticSearch integration
- **Request Tracing**: Correlation ID for distributed tracing
- **Audit Logs**: Complete audit trail for compliance
- **Activity Tracking**: User activity monitoring

## Tech Stack

- **Language**: Go 1.21+
- **Framework**: Gin
- **ORM**: GORM
- **Database**: PostgreSQL 15+
- **Cache**: Redis 7+
- **Storage**: MinIO
- **Documentation**: Swagger/OpenAPI

## Architecture

```
├── cmd/                    # Application entry points
├── internal/              # Private application code
│   ├── core/             # Business logic
│   │   ├── domain/       # Domain models
│   │   ├── ports/        # Interfaces
│   │   └── usecases/     # Use cases
│   ├── adapters/         # External adapters
│   │   ├── handlers/     # HTTP handlers
│   │   └── repositories/ # Data repositories
│   ├── infrastructure/   # Infrastructure layer
│   │   ├── database/     # Database setup
│   │   ├── cache/        # Redis cache
│   │   ├── storage/      # MinIO storage
│   │   └── middleware/   # HTTP middleware
│   └── config/          # Configuration
└── pkg/                  # Public packages
    ├── logger/          # Logging utilities
    ├── errors/          # Error handling
    ├── response/        # API responses
    ├── pagination/      # Pagination utilities
    └── utils/           # Helper functions
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Make (optional, for convenience)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/owner/go-cms.git
cd go-cms
```

2. Copy environment file:
```bash
cp .env.example .env
```

3. Start infrastructure services:
```bash
docker-compose up -d
```

4. Install dependencies:
```bash
go mod download
```

5. Run database migrations:
```bash
make migrate-up
```

6. Start the application:
```bash
make run
```

The API will be available at `http://localhost:8080`

## API Documentation

Swagger documentation is available at `http://localhost:8080/swagger/index.html` when the application is running.

To regenerate Swagger docs:
```bash
make swagger
```

## Development

### Running Tests
```bash
make test
```

### Code Coverage
```bash
make test-coverage
```

### Linting
```bash
make lint
```

### Format Code
```bash
make fmt
```

## Database Migrations

Create a new migration:
```bash
make migrate-create name=create_users_table
```

Run migrations:
```bash
make migrate-up
```

Rollback migrations:
```bash
make migrate-down
```

## Permission System

The system implements a hierarchical permission structure:

```
module:department:service:resource:action
```

Example permissions:
- `crm:sales:customers:customers:create`
- `content:editorial:posts:posts:publish`
- `admin:system:users:users:delete`

This allows for granular access control at multiple organizational levels.

## Environment Variables

See `.env.example` for all available configuration options.

Key variables:
- `DB_HOST`: PostgreSQL host
- `REDIS_HOST`: Redis host
- `MINIO_ENDPOINT`: MinIO endpoint
- `JWT_SECRET`: JWT signing secret
- `LOG_LEVEL`: Logging level (debug, info, warn, error)

## Docker Support

Build and run with Docker:
```bash
docker build -t go-cms .
docker run -p 8080:8080 go-cms
```

Or use Docker Compose for the full stack:
```bash
docker-compose up
```

## Performance Optimization

### Database
- Composite indexes on frequently queried columns
- Table partitioning for large tables (audit logs)
- Connection pooling with configurable limits
- Prepared statements for better performance

### Caching
- Permission caching with automatic invalidation
- Session caching in Redis
- Query result caching for expensive operations

### Pagination
- Cursor-based pagination for efficient large dataset handling
- Configurable page sizes with maximum limits

## Security

- Password hashing with bcrypt
- JWT with RS256 algorithm
- 2FA using TOTP
- Rate limiting per endpoint
- CORS configuration
- SQL injection prevention via ORM
- XSS protection

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support, email support@example.com or create an issue in the repository.

## Roadmap

- [ ] GraphQL API support
- [ ] WebSocket for real-time notifications
- [ ] Advanced reporting and analytics
- [ ] Multi-tenancy support
- [ ] API rate limiting per user
- [ ] Email templates and notifications
- [ ] File virus scanning
- [ ] Advanced search with ElasticSearch
- [ ] Kubernetes deployment manifests
- [ ] CI/CD pipeline configuration
# go-cms-be
