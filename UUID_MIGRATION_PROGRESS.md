# 🔄 UUID MIGRATION - PROGRESS UPDATE

## ✅ Đã Hoàn Thành (50%)

### 1. Domain Models ✅
- ✅ Created `UUIDModel` for User
- ✅ Created `BaseModel` with sequence support
- ✅ Updated User to use UUID
- ✅ Updated all User foreign keys (Post, Media, Customer, RefreshToken, etc.)

### 2. Repository Interfaces ✅
- ✅ UserRepository - all methods use uuid.UUID
- ✅ CustomerRepository - AssignedTo uses uuid.UUID
- ✅ RefreshTokenRepository - UserID uses uuid.UUID
- ✅ UserFilter.IDs - uses []uuid.UUID

### 3. Repository Implementations ✅
- ✅ UserRepository - Complete rewrite with UUID
- ✅ RefreshTokenRepository - UUID support
- ✅ OTPRepository - No changes needed

## 🔧 Còn Lại (50%)

### 1. Utils Package - JWT & Tokens
**File**: `pkg/utils/utils.go`

Cần update:
```go
// Current
func GenerateJWT(userID uint, email string, cfg *config.JWTConfig) (string, error)
func GenerateRefreshToken(userID uint, email string, cfg *config.JWTConfig) (string, error)

// Need to change to
func GenerateJWT(userID uuid.UUID, email string, cfg *config.JWTConfig) (string, error)
func GenerateRefreshToken(userID uuid.UUID, email string, cfg *config.JWTConfig) (string, error)

// JWT Claims
type Claims struct {
    UserID uuid.UUID `json:"user_id"` // was uint
    Email  string    `json:"email"`
    jwt.RegisteredClaims
}
```

### 2. Use Cases - Auth
**File**: `internal/core/usecases/auth/auth_usecase.go`

Cần update tất cả methods:
```go
// Change all userID parameters from uint to uuid.UUID
func (uc *useCase) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*domain.User, error)
func (uc *useCase) UpdateProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) (*domain.User, error)
func (uc *useCase) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error
func (uc *useCase) Enable2FA(ctx context.Context, userID uuid.UUID) (*Enable2FAResponse, error)
func (uc *useCase) Verify2FA(ctx context.Context, userID uuid.UUID, code string) error
func (uc *useCase) Disable2FA(ctx context.Context, userID uuid.UUID, password string) error
func (uc *useCase) Logout(ctx context.Context, userID uuid.UUID, token string) error
```

### 3. Middleware - Auth
**File**: `internal/infrastructure/middleware/auth.go`

Cần update:
```go
// JWT Claims
type Claims struct {
    UserID uuid.UUID `json:"user_id"`
    Email  string    `json:"email"`
    jwt.RegisteredClaims
}

// Context helpers
func GetUserID(c *gin.Context) (uuid.UUID, error)
func MustGetUserID(c *gin.Context) uuid.UUID
```

### 4. Handlers - Auth
**File**: `internal/adapters/handlers/auth_handler.go`

Cần update:
```go
// Parse UUID from context
userID := middleware.MustGetUserID(c) // returns uuid.UUID now
```

### 5. Pagination Package
**File**: `pkg/pagination/pagination.go`

Có vẻ struct không match. Cần check:
- `Cursor.After` field
- `Cursor.HasMore` field
- `OffsetPagination.Limit` field

## 🚀 Quick Fix Plan

### Step 1: Fix Utils (15 min)
```bash
# Update pkg/utils/utils.go
- Change GenerateJWT to accept uuid.UUID
- Change GenerateRefreshToken to accept uuid.UUID
- Update Claims struct
- Update ParseJWT to return uuid.UUID
```

### Step 2: Fix Use Cases (20 min)
```bash
# Update internal/core/usecases/auth/auth_usecase.go
- Change all userID parameters to uuid.UUID
- Update all calls to utils.GenerateJWT
- Update logging (zap.String("user_id", userID.String()))
```

### Step 3: Fix Middleware (10 min)
```bash
# Update internal/infrastructure/middleware/auth.go
- Update Claims struct
- Update GetUserID to return uuid.UUID
- Update MustGetUserID to return uuid.UUID
```

### Step 4: Fix Handlers (5 min)
```bash
# Update internal/adapters/handlers/auth_handler.go
- userID is now uuid.UUID from middleware
- No parsing needed, already UUID
```

### Step 5: Fix Pagination (5 min)
```bash
# Check pkg/pagination/pagination.go
- Ensure Cursor has After and HasMore fields
- Ensure OffsetPagination has Limit field
```

### Step 6: Run go mod tidy (1 min)
```bash
go mod tidy
go build ./...
```

## 📊 Estimated Time: ~1 hour

## 🎯 Next Action

**Tôi sẽ tiếp tục fix theo thứ tự:**
1. Utils package (JWT)
2. Use cases
3. Middleware
4. Handlers
5. Test compile

**Bạn muốn tôi:**
- ✅ **Tiếp tục fix** - Tôi sẽ fix tất cả ngay
- ⏸️ **Tạm dừng** - Để bạn review code
- 📝 **Giải thích thêm** - Về bất kỳ phần nào

Tôi sẽ tiếp tục fix ngay! 🚀
