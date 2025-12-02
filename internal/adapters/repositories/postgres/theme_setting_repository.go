package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"gorm.io/gorm"
)

type themeSettingRepository struct {
	db *gorm.DB
}

// NewThemeSettingRepository creates a new theme setting repository instance
func NewThemeSettingRepository(db *gorm.DB) repositories.ThemeSettingRepository {
	return &themeSettingRepository{db: db}
}

func (r *themeSettingRepository) Create(ctx context.Context, theme *domain.ThemeSetting) error {
	return r.db.WithContext(ctx).Create(theme).Error
}

func (r *themeSettingRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ThemeSetting, error) {
	var theme domain.ThemeSetting
	if err := r.db.WithContext(ctx).First(&theme, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &theme, nil
}

func (r *themeSettingRepository) Update(ctx context.Context, theme *domain.ThemeSetting) error {
	return r.db.WithContext(ctx).Save(theme).Error
}

func (r *themeSettingRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.ThemeSetting{}, "id = ?", id).Error
}

func (r *themeSettingRepository) List(ctx context.Context) ([]*domain.ThemeSetting, error) {
	var themes []*domain.ThemeSetting
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&themes).Error; err != nil {
		return nil, err
	}
	return themes, nil
}

func (r *themeSettingRepository) GetActive(ctx context.Context) (*domain.ThemeSetting, error) {
	var theme domain.ThemeSetting
	if err := r.db.WithContext(ctx).Where("is_active = ?", true).First(&theme).Error; err != nil {
		return nil, err
	}
	return &theme, nil
}

func (r *themeSettingRepository) Activate(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Deactivate all themes
		if err := tx.Model(&domain.ThemeSetting{}).Where("is_active = ?", true).Update("is_active", false).Error; err != nil {
			return err
		}

		// Activate the selected theme
		if err := tx.Model(&domain.ThemeSetting{}).Where("id = ?", id).Update("is_active", true).Error; err != nil {
			return err
		}

		return nil
	})
}
