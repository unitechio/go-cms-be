package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/pkg/errors"
	"github.com/owner/go-cms/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type userSettingsRepository struct {
	db *gorm.DB
}

// NewUserSettingsRepository creates a new user settings repository
func NewUserSettingsRepository(db *gorm.DB) repositories.UserSettingsRepository {
	return &userSettingsRepository{db: db}
}

func (r *userSettingsRepository) GetByUserAndKey(ctx context.Context, userID uuid.UUID, key string) (*domain.UserSetting, error) {
	var setting domain.UserSetting
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND setting_key = ?", userID, key).
		First(&setting).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		logger.Error("Failed to get user setting", zap.Error(err), zap.String("user_id", userID.String()), zap.String("key", key))
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get user setting", 500)
	}
	return &setting, nil
}

func (r *userSettingsRepository) GetAllByUser(ctx context.Context, userID uuid.UUID) ([]*domain.UserSetting, error) {
	var settings []*domain.UserSetting
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Find(&settings).Error; err != nil {
		logger.Error("Failed to get user settings", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get user settings", 500)
	}
	return settings, nil
}

func (r *userSettingsRepository) Upsert(ctx context.Context, setting *domain.UserSetting) error {
	if err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "setting_key"}},
			DoUpdates: clause.AssignmentColumns([]string{"setting_value", "updated_at"}),
		}).
		Create(setting).Error; err != nil {
		logger.Error("Failed to upsert user setting", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to upsert user setting", 500)
	}
	return nil
}

func (r *userSettingsRepository) Delete(ctx context.Context, userID uuid.UUID, key string) error {
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND setting_key = ?", userID, key).
		Delete(&domain.UserSetting{}).Error; err != nil {
		logger.Error("Failed to delete user setting", zap.Error(err), zap.String("user_id", userID.String()), zap.String("key", key))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to delete user setting", 500)
	}
	return nil
}

func (r *userSettingsRepository) BulkUpsert(ctx context.Context, settings []*domain.UserSetting) error {
	if len(settings) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "setting_key"}},
			DoUpdates: clause.AssignmentColumns([]string{"setting_value", "updated_at"}),
		}).
		Create(&settings).Error; err != nil {
		logger.Error("Failed to bulk upsert user settings", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to bulk upsert user settings", 500)
	}
	return nil
}
