package domain

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Translation represents a translation entry
type Translation struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Key       string    `gorm:"size:255;not null;index:idx_translation_key_locale" json:"key"`
	Locale    string    `gorm:"size:10;not null;index:idx_translation_key_locale" json:"locale"`
	Value     string    `gorm:"type:text;not null" json:"value"`
	Namespace string    `gorm:"size:100;default:'common'" json:"namespace"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName specifies the table name
func (Translation) TableName() string {
	return "translations"
}

// TranslatableContent represents JSONB content with translations
type TranslatableContent map[string]string

// Scan implements sql.Scanner interface
func (t *TranslatableContent) Scan(value interface{}) error {
	if value == nil {
		*t = make(TranslatableContent)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, t)
}

// Value implements driver.Valuer interface
func (t TranslatableContent) Value() (driver.Value, error) {
	return json.Marshal(t)
}

// Get returns translation for a locale, falls back to default
func (t TranslatableContent) Get(locale, defaultLocale string) string {
	if val, ok := t[locale]; ok && val != "" {
		return val
	}
	if val, ok := t[defaultLocale]; ok {
		return val
	}
	// Return first available translation
	for _, val := range t {
		if val != "" {
			return val
		}
	}
	return ""
}

// Set sets translation for a locale
func (t TranslatableContent) Set(locale, value string) {
	t[locale] = value
}

// Supported locales
const (
	LocaleEnglish    = "en"
	LocaleVietnamese = "vi"
	LocaleFrench     = "fr"
	LocaleDefault    = LocaleEnglish
)

var SupportedLocales = []string{LocaleEnglish, LocaleVietnamese, LocaleFrench}

// IsValidLocale checks if a locale is supported
func IsValidLocale(locale string) bool {
	for _, l := range SupportedLocales {
		if l == locale {
			return true
		}
	}
	return false
}
