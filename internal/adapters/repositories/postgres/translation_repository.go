package postgres

import (
	"context"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/pkg/errors"
	"github.com/owner/go-cms/pkg/logger"
)

type translationRepository struct {
	db *gorm.DB
}

// NewTranslationRepository creates a new translation repository
func NewTranslationRepository(db *gorm.DB) repositories.TranslationRepository {
	return &translationRepository{db: db}
}

func (r *translationRepository) Get(ctx context.Context, key, locale, namespace string) (*domain.Translation, error) {
	var translation domain.Translation
	if err := r.db.WithContext(ctx).
		Where("key = ? AND locale = ? AND namespace = ?", key, locale, namespace).
		First(&translation).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		logger.Error("Failed to get translation", zap.Error(err))
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get translation", 500)
	}
	return &translation, nil
}

func (r *translationRepository) GetByLocale(ctx context.Context, locale, namespace string) ([]*domain.Translation, error) {
	var translations []*domain.Translation
	query := r.db.WithContext(ctx).Where("locale = ?", locale)

	if namespace != "" {
		query = query.Where("namespace = ?", namespace)
	}

	if err := query.Find(&translations).Error; err != nil {
		logger.Error("Failed to get translations by locale", zap.Error(err))
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get translations", 500)
	}
	return translations, nil
}

func (r *translationRepository) GetByKey(ctx context.Context, key, namespace string) ([]*domain.Translation, error) {
	var translations []*domain.Translation
	query := r.db.WithContext(ctx).Where("key = ?", key)

	if namespace != "" {
		query = query.Where("namespace = ?", namespace)
	}

	if err := query.Find(&translations).Error; err != nil {
		logger.Error("Failed to get translations by key", zap.Error(err))
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get translations", 500)
	}
	return translations, nil
}

func (r *translationRepository) Upsert(ctx context.Context, translation *domain.Translation) error {
	if err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}, {Name: "locale"}, {Name: "namespace"}},
			DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
		}).
		Create(translation).Error; err != nil {
		logger.Error("Failed to upsert translation", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to upsert translation", 500)
	}
	return nil
}

func (r *translationRepository) BulkUpsert(ctx context.Context, translations []*domain.Translation) error {
	if len(translations) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "key"}, {Name: "locale"}, {Name: "namespace"}},
			DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
		}).
		Create(&translations).Error; err != nil {
		logger.Error("Failed to bulk upsert translations", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to bulk upsert translations", 500)
	}
	return nil
}

func (r *translationRepository) Delete(ctx context.Context, key, locale, namespace string) error {
	if err := r.db.WithContext(ctx).
		Where("key = ? AND locale = ? AND namespace = ?", key, locale, namespace).
		Delete(&domain.Translation{}).Error; err != nil {
		logger.Error("Failed to delete translation", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to delete translation", 500)
	}
	return nil
}

func (r *translationRepository) GetNamespaces(ctx context.Context) ([]string, error) {
	var namespaces []string
	if err := r.db.WithContext(ctx).
		Model(&domain.Translation{}).
		Distinct("namespace").
		Pluck("namespace", &namespaces).Error; err != nil {
		logger.Error("Failed to get namespaces", zap.Error(err))
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get namespaces", 500)
	}
	return namespaces, nil
}
