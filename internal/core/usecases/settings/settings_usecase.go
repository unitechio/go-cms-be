package settings

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/pkg/errors"
	"github.com/owner/go-cms/pkg/logger"
	"go.uber.org/zap"
)

type UseCase interface {
	GetSetting(ctx context.Context, userID uuid.UUID, key string) (*domain.UserSetting, error)
	GetAllSettings(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error)
	UpdateSetting(ctx context.Context, userID uuid.UUID, key string, value interface{}) error
	DeleteSetting(ctx context.Context, userID uuid.UUID, key string) error
	BulkUpdateSettings(ctx context.Context, userID uuid.UUID, settings map[string]interface{}) error
	GetDefaultSettings(ctx context.Context) map[string]interface{}
}

type useCase struct {
	settingsRepo repositories.UserSettingsRepository
	userRepo     repositories.UserRepository
}

func NewUseCase(settingsRepo repositories.UserSettingsRepository, userRepo repositories.UserRepository) UseCase {
	return &useCase{
		settingsRepo: settingsRepo,
		userRepo:     userRepo,
	}
}

func (uc *useCase) GetSetting(ctx context.Context, userID uuid.UUID, key string) (*domain.UserSetting, error) {
	// Verify user exists
	if _, err := uc.userRepo.GetByID(ctx, userID); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeNotFound, "user not found", 404)
	}

	setting, err := uc.settingsRepo.GetByUserAndKey(ctx, userID, key)
	if err != nil {
		if err == errors.ErrNotFound {
			// Return default setting
			defaultValue := uc.getDefaultValue(key)
			return &domain.UserSetting{
				UserID:       userID,
				SettingKey:   key,
				SettingValue: defaultValue,
			}, nil
		}
		return nil, err
	}

	return setting, nil
}

func (uc *useCase) GetAllSettings(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	// Verify user exists
	if _, err := uc.userRepo.GetByID(ctx, userID); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeNotFound, "user not found", 404)
	}

	settings, err := uc.settingsRepo.GetAllByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Convert to map
	result := uc.GetDefaultSettings(ctx)
	for _, setting := range settings {
		result[setting.SettingKey] = setting.SettingValue
	}

	return result, nil
}

func (uc *useCase) UpdateSetting(ctx context.Context, userID uuid.UUID, key string, value interface{}) error {
	// Verify user exists
	if _, err := uc.userRepo.GetByID(ctx, userID); err != nil {
		return errors.Wrap(err, errors.ErrCodeNotFound, "user not found", 404)
	}

	// Validate setting key
	if !uc.isValidSettingKey(key) {
		return errors.New(errors.ErrCodeBadRequest, "invalid setting key", 400)
	}

	// Convert value to SettingValue
	settingValue, err := uc.convertToSettingValue(value)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeBadRequest, "invalid setting value", 400)
	}

	setting := &domain.UserSetting{
		UserID:       userID,
		SettingKey:   key,
		SettingValue: settingValue,
	}

	if err := uc.settingsRepo.Upsert(ctx, setting); err != nil {
		logger.Error("Failed to update setting", zap.Error(err), zap.String("user_id", userID.String()), zap.String("key", key))
		return err
	}

	logger.Info("Setting updated successfully", zap.String("user_id", userID.String()), zap.String("key", key))
	return nil
}

func (uc *useCase) DeleteSetting(ctx context.Context, userID uuid.UUID, key string) error {
	// Verify user exists
	if _, err := uc.userRepo.GetByID(ctx, userID); err != nil {
		return errors.Wrap(err, errors.ErrCodeNotFound, "user not found", 404)
	}

	if err := uc.settingsRepo.Delete(ctx, userID, key); err != nil {
		logger.Error("Failed to delete setting", zap.Error(err), zap.String("user_id", userID.String()), zap.String("key", key))
		return err
	}

	logger.Info("Setting deleted successfully", zap.String("user_id", userID.String()), zap.String("key", key))
	return nil
}

func (uc *useCase) BulkUpdateSettings(ctx context.Context, userID uuid.UUID, settings map[string]interface{}) error {
	// Verify user exists
	if _, err := uc.userRepo.GetByID(ctx, userID); err != nil {
		return errors.Wrap(err, errors.ErrCodeNotFound, "user not found", 404)
	}

	var settingsList []*domain.UserSetting
	for key, value := range settings {
		if !uc.isValidSettingKey(key) {
			continue
		}

		settingValue, err := uc.convertToSettingValue(value)
		if err != nil {
			continue
		}

		settingsList = append(settingsList, &domain.UserSetting{
			UserID:       userID,
			SettingKey:   key,
			SettingValue: settingValue,
		})
	}

	if err := uc.settingsRepo.BulkUpsert(ctx, settingsList); err != nil {
		logger.Error("Failed to bulk update settings", zap.Error(err), zap.String("user_id", userID.String()))
		return err
	}

	logger.Info("Settings bulk updated successfully", zap.String("user_id", userID.String()), zap.Int("count", len(settingsList)))
	return nil
}

func (uc *useCase) GetDefaultSettings(ctx context.Context) map[string]interface{} {
	return map[string]interface{}{
		domain.SettingKeyLanguage: domain.SettingValue{
			"locale": "en",
		},
		domain.SettingKeyTheme: domain.SettingValue{
			"mode": "light",
			"font": "Inter",
		},
		domain.SettingKeyNotifications: domain.SettingValue{
			"email":     true,
			"push":      false,
			"frequency": "instant",
			"types": map[string]bool{
				"posts":     true,
				"comments":  true,
				"documents": true,
				"system":    true,
			},
		},
		domain.SettingKeyDisplay: domain.SettingValue{
			"density": "comfortable",
			"sidebar": "expanded",
		},
	}
}

func (uc *useCase) isValidSettingKey(key string) bool {
	validKeys := []string{
		domain.SettingKeyLanguage,
		domain.SettingKeyTheme,
		domain.SettingKeyNotifications,
		domain.SettingKeyDisplay,
	}

	for _, validKey := range validKeys {
		if key == validKey {
			return true
		}
	}
	return false
}

func (uc *useCase) convertToSettingValue(value interface{}) (domain.SettingValue, error) {
	// If already a map, convert directly
	if m, ok := value.(map[string]interface{}); ok {
		return domain.SettingValue(m), nil
	}

	// Otherwise, marshal and unmarshal
	bytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	var settingValue domain.SettingValue
	if err := json.Unmarshal(bytes, &settingValue); err != nil {
		return nil, err
	}

	return settingValue, nil
}

func (uc *useCase) getDefaultValue(key string) domain.SettingValue {
	defaults := uc.GetDefaultSettings(context.Background())
	if value, ok := defaults[key]; ok {
		if sv, ok := value.(domain.SettingValue); ok {
			return sv
		}
	}
	return domain.SettingValue{}
}
