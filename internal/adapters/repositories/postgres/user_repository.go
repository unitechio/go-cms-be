package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/pkg/errors"
	"github.com/owner/go-cms/pkg/logger"
	"github.com/owner/go-cms/pkg/pagination"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) repositories.UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		logger.Error("Failed to create user", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to create user", 500)
	}
	return nil
}

// GetByID gets a user by ID
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

// GetByEmail gets a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrUserNotFound
		}
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get user", 500)
	}
	return &user, nil
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
		logger.Error("Failed to update user", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to update user", 500)
	}
	return nil
}

// Delete deletes a user (soft delete)
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.User{}).Error; err != nil {
		logger.Error("Failed to delete user", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to delete user", 500)
	}
	return nil
}

// List lists users with offset pagination
func (r *userRepository) List(ctx context.Context, filter repositories.UserFilter, page *pagination.OffsetPagination) ([]*domain.User, int64, error) {
	logger.Info("ListUsers repo called", zap.Any("filter", filter), zap.Any("page", page))

	var users []*domain.User
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.User{})

	// Apply filters
	query = r.applyFilters(query, filter)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to count users", 500)
	}
	logger.Info("ListUsers total count", zap.Int64("total", total))

	// Apply pagination
	offset := (page.Page - 1) * page.Limit
	if err := query.Offset(offset).Limit(page.Limit).Find(&users).Error; err != nil {
		return nil, 0, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to list users", 500)
	}
	logger.Info("ListUsers found count", zap.Int("count", len(users)))

	return users, total, nil
}

// ListWithCursor lists users with cursor pagination
func (r *userRepository) ListWithCursor(ctx context.Context, filter repositories.UserFilter, cursor *pagination.Cursor, limit int) ([]*domain.User, *pagination.Cursor, error) {
	var users []*domain.User

	query := r.db.WithContext(ctx).Model(&domain.User{})

	// Apply filters
	query = r.applyFilters(query, filter)

	// Apply cursor
	if cursor != nil && cursor.After != "" {
		// Decode cursor (UUID string)
		afterID, err := uuid.Parse(cursor.After)
		if err != nil {
			return nil, nil, errors.Wrap(err, errors.ErrCodeValidation, "invalid cursor", 400)
		}
		query = query.Where("id > ?", afterID)
	}

	// Order and limit
	query = query.Order("id ASC").Limit(limit + 1)

	if err := query.Find(&users).Error; err != nil {
		return nil, nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to list users", 500)
	}

	// Check if there are more results
	hasMore := len(users) > limit
	if hasMore {
		users = users[:limit]
	}

	// Create next cursor
	var nextCursor *pagination.Cursor
	if hasMore && len(users) > 0 {
		lastUser := users[len(users)-1]
		nextCursor = &pagination.Cursor{
			After:   lastUser.ID.String(),
			HasMore: true,
		}
	}

	return users, nextCursor, nil
}

// applyFilters applies filters to query
func (r *userRepository) applyFilters(query *gorm.DB, filter repositories.UserFilter) *gorm.DB {
	if filter.Email != "" {
		query = query.Where("email = ?", filter.Email)
	}

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("email LIKE ? OR first_name LIKE ? OR last_name LIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	if len(filter.IDs) > 0 {
		query = query.Where("id IN ?", filter.IDs)
	}

	// if filter.RoleID != nil {
	// 	query = query.Joins("JOIN user_roles ON users.id = user_roles.user_id").
	// 		Where("user_roles.role_id = ?", *filter.RoleID)
	// }
	if filter.RoleID != nil && *filter.RoleID > 0 {
		query = query.
			Joins("LEFT JOIN user_roles ON users.id = user_roles.user_id").
			Where("user_roles.role_id = ?", *filter.RoleID)
	}

	return query
}

// AssignRole assigns a role to a user
func (r *userRepository) AssignRole(ctx context.Context, userID uuid.UUID, roleID uint) error {
	userRole := domain.UserRole{
		UserID: userID,
		RoleID: roleID,
	}

	if err := r.db.WithContext(ctx).Create(&userRole).Error; err != nil {
		logger.Error("Failed to assign role", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to assign role", 500)
	}

	return nil
}

// RemoveRole removes a role from a user
func (r *userRepository) RemoveRole(ctx context.Context, userID uuid.UUID, roleID uint) error {
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&domain.UserRole{}).Error; err != nil {
		logger.Error("Failed to remove role", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to remove role", 500)
	}

	return nil
}

// GetUserRoles gets all roles for a user
func (r *userRepository) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
	var roles []*domain.Role

	if err := r.db.WithContext(ctx).
		Joins("JOIN user_roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error; err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get user roles", 500)
	}

	return roles, nil
}

// AssignPermission assigns a permission to a user
func (r *userRepository) AssignPermission(ctx context.Context, userID uuid.UUID, permissionID uint) error {
	userPermission := domain.UserPermission{
		UserID:       userID,
		PermissionID: permissionID,
	}

	if err := r.db.WithContext(ctx).Create(&userPermission).Error; err != nil {
		logger.Error("Failed to assign permission", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to assign permission", 500)
	}

	return nil
}

// RemovePermission removes a permission from a user
func (r *userRepository) RemovePermission(ctx context.Context, userID uuid.UUID, permissionID uint) error {
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND permission_id = ?", userID, permissionID).
		Delete(&domain.UserPermission{}).Error; err != nil {
		logger.Error("Failed to remove permission", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to remove permission", 500)
	}

	return nil
}

// GetUserPermissions gets all permissions for a user
func (r *userRepository) GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]*domain.Permission, error) {
	var permissions []*domain.Permission

	if err := r.db.WithContext(ctx).
		Joins("JOIN user_permissions ON permissions.id = user_permissions.permission_id").
		Where("user_permissions.user_id = ?", userID).
		Find(&permissions).Error; err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get user permissions", 500)
	}

	return permissions, nil
}

// UpdateLastLogin updates the last login time and IP
func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, ip string) error {
	updates := map[string]interface{}{
		"last_login_at": gorm.Expr("NOW()"),
		"last_login_ip": ip,
	}

	if err := r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ?", userID).
		Updates(updates).Error; err != nil {
		logger.Error("Failed to update last login", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to update last login", 500)
	}

	return nil
}

// UpdatePassword updates the user password
func (r *userRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	if err := r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ?", userID).
		Update("password", hashedPassword).Error; err != nil {
		logger.Error("Failed to update password", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to update password", 500)
	}

	return nil
}

// Enable2FA enables 2FA for a user
func (r *userRepository) Enable2FA(ctx context.Context, userID uuid.UUID, secret string) error {
	updates := map[string]interface{}{
		"two_factor_enabled": true,
		"two_factor_secret":  secret,
	}

	if err := r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ?", userID).
		Updates(updates).Error; err != nil {
		logger.Error("Failed to enable 2FA", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to enable 2FA", 500)
	}

	return nil
}

// Disable2FA disables 2FA for a user
func (r *userRepository) Disable2FA(ctx context.Context, userID uuid.UUID) error {
	updates := map[string]interface{}{
		"two_factor_enabled": false,
		"two_factor_secret":  "",
	}

	if err := r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ?", userID).
		Updates(updates).Error; err != nil {
		logger.Error("Failed to disable 2FA", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to disable 2FA", 500)
	}

	return nil
}

// UpdateStatus updates the user status
func (r *userRepository) UpdateStatus(ctx context.Context, userID uuid.UUID, status domain.UserStatus) error {
	if err := r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ?", userID).
		Update("status", status).Error; err != nil {
		logger.Error("Failed to update status", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to update status", 500)
	}

	return nil
}

// VerifyEmail verifies a user's email
func (r *userRepository) VerifyEmail(ctx context.Context, userID uuid.UUID) error {
	updates := map[string]interface{}{
		"email_verified":    true,
		"email_verified_at": gorm.Expr("NOW()"),
	}

	if err := r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id = ?", userID).
		Updates(updates).Error; err != nil {
		logger.Error("Failed to verify email", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to verify email", 500)
	}

	return nil
}
