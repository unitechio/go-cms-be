# 🔄 UUID & SEQUENCE MIGRATION - IN PROGRESS

## ✅ Đã Hoàn Thành

### 1. Domain Models Updated
- ✅ Created `UUIDModel` base model for User
- ✅ Created `BaseModel` with sequence support for other entities
- ✅ Updated `User` model to use UUID primary key
- ✅ Updated `Post.AuthorID` to UUID
- ✅ Updated `Media.UploadedBy` to UUID
- ✅ Updated `Customer.AssignedTo` to UUID
- ✅ Updated `RefreshToken.UserID` to UUID
- ✅ Updated `UserRole.UserID` to UUID
- ✅ Updated `UserPermission.UserID` to UUID

### 2. Database Sequence Strategy
```sql
-- Sequences will be created for each table
CREATE SEQUENCE customers_id_seq;
CREATE SEQUENCE roles_id_seq;
CREATE SEQUENCE permissions_id_seq;
CREATE SEQUENCE posts_id_seq;
CREATE SEQUENCE media_id_seq;
-- etc...

-- Tables will use sequence for ID
CREATE TABLE customers (
    id BIGINT PRIMARY KEY DEFAULT nextval('customers_id_seq'),
    ...
);
```

## 🔧 Cần Cập Nhật

### 1. Repository Interfaces
Cần thay đổi tất cả methods từ `uint` sang `uuid.UUID` cho User:

**File**: `internal/core/ports/repositories/user_repository.go`

```go
// Before
GetByID(ctx context.Context, id uint) (*domain.User, error)
VerifyEmail(ctx context.Context, userID uint) error

// After
GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
VerifyEmail(ctx context.Context, userID uuid.UUID) error
```

### 2. Repository Implementations
**File**: `internal/adapters/repositories/postgres/user_repository.go`

Cần cập nhật:
- Tất cả parameters từ `uint` → `uuid.UUID`
- Cursor pagination sử dụng UUID
- Foreign key references

### 3. Use Cases
**File**: `internal/core/usecases/auth/auth_usecase.go`

Cần cập nhật:
- Tất cả methods nhận `userID uint` → `userID uuid.UUID`
- JWT generation nhận UUID
- Logging với UUID

### 4. Utils Package
**File**: `pkg/utils/utils.go`

```go
// Before
func GenerateJWT(userID uint, email string, cfg *config.JWTConfig) (string, error)

// After
func GenerateJWT(userID uuid.UUID, email string, cfg *config.JWTConfig) (string, error)
```

### 5. Middleware
**File**: `internal/infrastructure/middleware/auth.go`

```go
// JWT Claims cần thay đổi
type Claims struct {
    UserID uuid.UUID `json:"user_id"` // was uint
    Email  string    `json:"email"`
    jwt.RegisteredClaims
}
```

### 6. Handlers
**File**: `internal/adapters/handlers/auth_handler.go`

- Parse UUID từ path parameters
- Validate UUID format

## 📋 Migration Steps (Recommended Order)

### Step 1: Install UUID Package ✅
```bash
go get github.com/google/uuid
```

### Step 2: Update Utils Package
- [ ] Update JWT generation to accept UUID
- [ ] Update JWT parsing to return UUID
- [ ] Update all helper functions

### Step 3: Update Repository Interfaces
- [ ] UserRepository - all methods
- [ ] CustomerRepository - AssignedTo field
- [ ] RefreshTokenRepository - UserID field
- [ ] All other repositories referencing User

### Step 4: Update Repository Implementations
- [ ] UserRepository implementation
- [ ] OTPRepository implementation
- [ ] RefreshTokenRepository implementation

### Step 5: Update Use Cases
- [ ] Auth use case - all methods
- [ ] Other use cases referencing User

### Step 6: Update Middleware
- [ ] Auth middleware - JWT claims
- [ ] Authorization middleware - user context

### Step 7: Update Handlers
- [ ] Auth handler - UUID parsing
- [ ] Other handlers

### Step 8: Database Migration
- [ ] Create sequences for all tables
- [ ] Create migration SQL files
- [ ] Update AutoMigrate function

## 🗄️ Database Sequences

### Create Sequences SQL
```sql
-- Create sequences for all tables
CREATE SEQUENCE IF NOT EXISTS customers_id_seq START 1;
CREATE SEQUENCE IF NOT EXISTS roles_id_seq START 1;
CREATE SEQUENCE IF NOT EXISTS permissions_id_seq START 1;
CREATE SEQUENCE IF NOT EXISTS posts_id_seq START 1;
CREATE SEQUENCE IF NOT EXISTS media_id_seq START 1;
CREATE SEQUENCE IF NOT EXISTS categories_id_seq START 1;
CREATE SEQUENCE IF NOT EXISTS tags_id_seq START 1;
CREATE SEQUENCE IF NOT EXISTS otps_id_seq START 1;
CREATE SEQUENCE IF NOT EXISTS refresh_tokens_id_seq START 1;
CREATE SEQUENCE IF NOT EXISTS audit_logs_id_seq START 1;
CREATE SEQUENCE IF NOT EXISTS notifications_id_seq START 1;
CREATE SEQUENCE IF NOT EXISTS activity_logs_id_seq START 1;
CREATE SEQUENCE IF NOT EXISTS email_templates_id_seq START 1;
CREATE SEQUENCE IF NOT EXISTS email_logs_id_seq START 1;
CREATE SEQUENCE IF NOT EXISTS system_settings_id_seq START 1;
CREATE SEQUENCE IF NOT EXISTS post_schedules_id_seq START 1;

-- Set default values for tables
ALTER TABLE customers ALTER COLUMN id SET DEFAULT nextval('customers_id_seq');
ALTER TABLE roles ALTER COLUMN id SET DEFAULT nextval('roles_id_seq');
-- etc...
```

### GORM Sequence Support
```go
// In BaseModel BeforeCreate hook
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
    if b.ID == 0 {
        var nextID uint
        tableName := tx.Statement.Table
        seqName := tableName + "_id_seq"
        
        err := tx.Raw("SELECT nextval(?)", seqName).Scan(&nextID).Error
        if err != nil {
            return err
        }
        b.ID = nextID
    }
    return nil
}
```

## 🎯 Benefits

### UUID for Users
- ✅ **Security**: Không thể đoán được user ID
- ✅ **Distributed**: Có thể generate offline
- ✅ **Unique**: Globally unique
- ✅ **Privacy**: Không lộ số lượng users

### Sequence for Others
- ✅ **Performance**: Faster than UUID for indexing
- ✅ **Readable**: Dễ debug và track
- ✅ **Compact**: Nhỏ hơn UUID (8 bytes vs 16 bytes)
- ✅ **Ordered**: Có thể sort theo thời gian tạo

## ⚠️ Breaking Changes

### API Responses
```json
// Before
{
  "id": 123,
  "email": "user@example.com"
}

// After (User)
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com"
}

// After (Customer, Post, etc.)
{
  "id": 123,  // Still integer from sequence
  "email": "customer@example.com"
}
```

### JWT Token
```json
// Before
{
  "user_id": 123,
  "email": "user@example.com"
}

// After
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com"
}
```

## 🚀 Next Actions

### Immediate (Required to compile)
1. Install UUID package: `go get github.com/google/uuid`
2. Update utils.GenerateJWT to accept UUID
3. Update repository interfaces
4. Update use cases
5. Update middleware

### Testing
1. Create migration SQL
2. Test with fresh database
3. Test all auth flows
4. Test relationships (User -> Posts, User -> Media)

### Documentation
1. Update API documentation
2. Update Swagger specs
3. Update README with UUID info

## 📝 Example Code Updates

### Utils - JWT Generation
```go
// pkg/utils/utils.go
func GenerateJWT(userID uuid.UUID, email string, cfg *config.JWTConfig) (string, error) {
    claims := &Claims{
        UserID: userID,
        Email:  email,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.AccessTokenExpire)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    // ... rest of implementation
}
```

### Repository - GetByID
```go
// internal/adapters/repositories/postgres/user_repository.go
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
    var user domain.User
    if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return nil, errors.ErrUserNotFound
        }
        return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get user", 500)
    }
    return &user, nil
}
```

### Use Case - Register
```go
// internal/core/usecases/auth/auth_usecase.go
// user.ID is now uuid.UUID, can use directly
accessToken, err := utils.GenerateJWT(user.ID, user.Email, &uc.config.JWT)
```

## 🎓 Summary

**Current Status**: Domain models updated, compilation errors present

**Required Work**: 
- Update ~50 function signatures
- Update JWT utils
- Update middleware
- Create database migrations
- Test thoroughly

**Estimated Time**: 2-3 hours for complete migration

**Risk Level**: Medium (breaking changes in API)

**Recommendation**: Complete migration in one go to avoid partial state
