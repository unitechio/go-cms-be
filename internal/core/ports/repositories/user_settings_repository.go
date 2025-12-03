package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
)

// UserSettingsRepository defines the interface for user settings data operations
type UserSettingsRepository interface {
	// Get a specific setting by user ID and key
	GetByUserAndKey(ctx context.Context, userID uuid.UUID, key string) (*domain.UserSetting, error)

	// Get all settings for a user
	GetAllByUser(ctx context.Context, userID uuid.UUID) ([]*domain.UserSetting, error)

	// Create or update a setting
	Upsert(ctx context.Context, setting *domain.UserSetting) error

	// Delete a setting
	Delete(ctx context.Context, userID uuid.UUID, key string) error

	// Bulk upsert settings
	BulkUpsert(ctx context.Context, settings []*domain.UserSetting) error
}
