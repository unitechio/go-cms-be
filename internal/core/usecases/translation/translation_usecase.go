package translation

import (
	"context"

	"go.uber.org/zap"

	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/pkg/errors"
	"github.com/owner/go-cms/pkg/logger"
)

type UseCase interface {
	GetTranslation(ctx context.Context, key, locale, namespace string) (string, error)
	GetTranslations(ctx context.Context, locale, namespace string) (map[string]string, error)
	SetTranslation(ctx context.Context, key, locale, namespace, value string) error
	BulkSetTranslations(ctx context.Context, translations map[string]map[string]string, namespace string) error
	DeleteTranslation(ctx context.Context, key, locale, namespace string) error
	GetSupportedLocales(ctx context.Context) []string
	GetNamespaces(ctx context.Context) ([]string, error)
}

type useCase struct {
	translationRepo repositories.TranslationRepository
}

func NewUseCase(translationRepo repositories.TranslationRepository) UseCase {
	return &useCase{
		translationRepo: translationRepo,
	}
}

func (uc *useCase) GetTranslation(ctx context.Context, key, locale, namespace string) (string, error) {
	if namespace == "" {
		namespace = "common"
	}

	translation, err := uc.translationRepo.Get(ctx, key, locale, namespace)
	if err != nil {
		if err == errors.ErrNotFound {
			// Return key as fallback
			return key, nil
		}
		return "", err
	}

	return translation.Value, nil
}

func (uc *useCase) GetTranslations(ctx context.Context, locale, namespace string) (map[string]string, error) {
	if namespace == "" {
		namespace = "common"
	}

	translations, err := uc.translationRepo.GetByLocale(ctx, locale, namespace)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, t := range translations {
		result[t.Key] = t.Value
	}

	return result, nil
}

func (uc *useCase) SetTranslation(ctx context.Context, key, locale, namespace, value string) error {
	if namespace == "" {
		namespace = "common"
	}

	if !domain.IsValidLocale(locale) {
		return errors.New(errors.ErrCodeBadRequest, "invalid locale", 400)
	}

	translation := &domain.Translation{
		Key:       key,
		Locale:    locale,
		Namespace: namespace,
		Value:     value,
	}

	if err := uc.translationRepo.Upsert(ctx, translation); err != nil {
		logger.Error("Failed to set translation", zap.Error(err))
		return err
	}

	logger.Info("Translation set successfully",
		zap.String("key", key),
		zap.String("locale", locale),
		zap.String("namespace", namespace))

	return nil
}

func (uc *useCase) BulkSetTranslations(ctx context.Context, translations map[string]map[string]string, namespace string) error {
	if namespace == "" {
		namespace = "common"
	}

	var translationList []*domain.Translation
	for key, locales := range translations {
		for locale, value := range locales {
			if !domain.IsValidLocale(locale) {
				continue
			}

			translationList = append(translationList, &domain.Translation{
				Key:       key,
				Locale:    locale,
				Namespace: namespace,
				Value:     value,
			})
		}
	}

	if err := uc.translationRepo.BulkUpsert(ctx, translationList); err != nil {
		logger.Error("Failed to bulk set translations", zap.Error(err))
		return err
	}

	logger.Info("Bulk translations set successfully",
		zap.Int("count", len(translationList)),
		zap.String("namespace", namespace))

	return nil
}

func (uc *useCase) DeleteTranslation(ctx context.Context, key, locale, namespace string) error {
	if namespace == "" {
		namespace = "common"
	}

	if err := uc.translationRepo.Delete(ctx, key, locale, namespace); err != nil {
		logger.Error("Failed to delete translation", zap.Error(err))
		return err
	}

	logger.Info("Translation deleted successfully",
		zap.String("key", key),
		zap.String("locale", locale))

	return nil
}

func (uc *useCase) GetSupportedLocales(ctx context.Context) []string {
	return domain.SupportedLocales
}

func (uc *useCase) GetNamespaces(ctx context.Context) ([]string, error) {
	return uc.translationRepo.GetNamespaces(ctx)
}
