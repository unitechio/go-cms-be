package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/pkg/errors"
	"github.com/owner/go-cms/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// otpRepository implements the OTPRepository interface
type otpRepository struct {
	db *gorm.DB
}

// NewOTPRepository creates a new OTP repository
func NewOTPRepository(db *gorm.DB) repositories.OTPRepository {
	return &otpRepository{db: db}
}

// Create creates a new OTP
func (r *otpRepository) Create(ctx context.Context, otp *domain.OTP) error {
	if err := r.db.WithContext(ctx).Create(otp).Error; err != nil {
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to create OTP", 500)
	}
	return nil
}

// GetByEmail retrieves an OTP by email and type
func (r *otpRepository) GetByEmail(ctx context.Context, email, otpType string) (*domain.OTP, error) {
	var otp domain.OTP
	if err := r.db.WithContext(ctx).
		Where("email = ? AND type = ? AND used = ?", email, otpType, false).
		Order("created_at DESC").
		First(&otp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get OTP", 500)
	}
	return &otp, nil
}

// MarkAsUsed marks an OTP as used
func (r *otpRepository) MarkAsUsed(ctx context.Context, id uint) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Model(&domain.OTP{}).Where("id = ?", id).Updates(map[string]interface{}{
		"used":    true,
		"used_at": now,
	}).Error; err != nil {
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to mark OTP as used", 500)
	}
	return nil
}

// DeleteExpired deletes expired OTPs
func (r *otpRepository) DeleteExpired(ctx context.Context) error {
	if err := r.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&domain.OTP{}).Error; err != nil {
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to delete expired OTPs", 500)
	}
	return nil
}

// DeleteByEmail deletes OTPs by email and type
func (r *otpRepository) DeleteByEmail(ctx context.Context, email, otpType string) error {
	if err := r.db.WithContext(ctx).Where("email = ? AND type = ?", email, otpType).Delete(&domain.OTP{}).Error; err != nil {
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to delete OTP", 500)
	}
	return nil
}

// refreshTokenRepository implements the RefreshTokenRepository interface
type refreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a new refresh token repository
func NewRefreshTokenRepository(db *gorm.DB) repositories.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

// Create creates a new refresh token
func (r *refreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	if err := r.db.WithContext(ctx).Create(token).Error; err != nil {
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to create refresh token", 500)
	}
	return nil
}

// GetByToken retrieves a refresh token by token string
func (r *refreshTokenRepository) GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	var refreshToken domain.RefreshToken
	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&refreshToken).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get refresh token", 500)
	}
	return &refreshToken, nil
}

// GetByUserID gets all refresh tokens for a user
func (r *refreshTokenRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.RefreshToken, error) {
	var tokens []*domain.RefreshToken

	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND revoked = ?", userID, false).
		Find(&tokens).Error; err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get refresh tokens", 500)
	}

	return tokens, nil
}

// Revoke revokes a refresh token
func (r *refreshTokenRepository) Revoke(ctx context.Context, token string) error {
	updates := map[string]interface{}{
		"revoked":    true,
		"revoked_at": gorm.Expr("NOW()"),
	}

	if err := r.db.WithContext(ctx).Model(&domain.RefreshToken{}).
		Where("token = ?", token).
		Updates(updates).Error; err != nil {
		logger.Error("Failed to revoke refresh token", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to revoke refresh token", 500)
	}

	return nil
}

// RevokeAllByUserID revokes all refresh tokens for a user
func (r *refreshTokenRepository) RevokeAllByUserID(ctx context.Context, userID uuid.UUID) error {
	updates := map[string]interface{}{
		"revoked":    true,
		"revoked_at": gorm.Expr("NOW()"),
	}

	if err := r.db.WithContext(ctx).Model(&domain.RefreshToken{}).
		Where("user_id = ?", userID).
		Updates(updates).Error; err != nil {
		logger.Error("Failed to revoke all refresh tokens", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to revoke all refresh tokens", 500)
	}

	return nil
}

// DeleteExpired deletes expired refresh tokens
func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	if err := r.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&domain.RefreshToken{}).Error; err != nil {
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to delete expired refresh tokens", 500)
	}
	return nil
}
