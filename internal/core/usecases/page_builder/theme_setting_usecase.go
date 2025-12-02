package page_builder

import (
	"context"

	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
)

type ThemeSettingUseCase struct {
	themeRepo repositories.ThemeSettingRepository
}

func NewThemeSettingUseCase(themeRepo repositories.ThemeSettingRepository) *ThemeSettingUseCase {
	return &ThemeSettingUseCase{
		themeRepo: themeRepo,
	}
}

func (uc *ThemeSettingUseCase) GetAllThemes(ctx context.Context) ([]*domain.ThemeSetting, error) {
	return uc.themeRepo.List(ctx)
}

func (uc *ThemeSettingUseCase) GetActiveTheme(ctx context.Context) (*domain.ThemeSetting, error) {
	return uc.themeRepo.GetActive(ctx)
}

func (uc *ThemeSettingUseCase) UpdateTheme(ctx context.Context, theme *domain.ThemeSetting) error {
	existingTheme, err := uc.themeRepo.GetByID(ctx, theme.ID)
	if err != nil {
		return err
	}

	existingTheme.Config = theme.Config
	// Name shouldn't change often, but can be allowed
	if theme.Name != "" {
		existingTheme.Name = theme.Name
	}

	return uc.themeRepo.Update(ctx, existingTheme)
}

func (uc *ThemeSettingUseCase) ActivateTheme(ctx context.Context, id uuid.UUID) error {
	return uc.themeRepo.Activate(ctx, id)
}
