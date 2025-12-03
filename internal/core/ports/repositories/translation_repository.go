package repositories

import (
	"context"

	"github.com/owner/go-cms/internal/core/domain"
)

// TranslationRepository defines the interface for translation operations
type TranslationRepository interface {
	// Get a translation by key and locale
	Get(ctx context.Context, key, locale, namespace string) (*domain.Translation, error)

	// Get all translations for a locale
	GetByLocale(ctx context.Context, locale, namespace string) ([]*domain.Translation, error)

	// Get all translations for a key (all locales)
	GetByKey(ctx context.Context, key, namespace string) ([]*domain.Translation, error)

	// Create or update a translation
	Upsert(ctx context.Context, translation *domain.Translation) error

	// Bulk upsert translations
	BulkUpsert(ctx context.Context, translations []*domain.Translation) error

	// Delete a translation
	Delete(ctx context.Context, key, locale, namespace string) error

	// Get all namespaces
	GetNamespaces(ctx context.Context) ([]string, error)
}
