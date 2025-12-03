package domain

// Update Post struct to support translations
// Add these fields to existing Post struct:

// TitleTranslations   TranslatableContent `gorm:"type:jsonb" json:"title_translations,omitempty"`
// ContentTranslations TranslatableContent `gorm:"type:jsonb" json:"content_translations,omitempty"`
// ExcerptTranslations TranslatableContent `gorm:"type:jsonb" json:"excerpt_translations,omitempty"`

// GetTitle returns translated title
func (p *Post) GetTitle(locale string) string {
	if p.TitleTranslations != nil {
		return p.TitleTranslations.Get(locale, LocaleDefault)
	}
	return p.Title
}

// GetContent returns translated content
func (p *Post) GetContent(locale string) string {
	if p.ContentTranslations != nil {
		return p.ContentTranslations.Get(locale, LocaleDefault)
	}
	return p.Content
}

// GetExcerpt returns translated excerpt
func (p *Post) GetExcerpt(locale string) string {
	if p.ExcerptTranslations != nil {
		return p.ExcerptTranslations.Get(locale, LocaleDefault)
	}
	return p.Excerpt
}

// Migration SQL:
/*
ALTER TABLE posts ADD COLUMN title_translations JSONB;
ALTER TABLE posts ADD COLUMN content_translations JSONB;
ALTER TABLE posts ADD COLUMN excerpt_translations JSONB;

-- Example data:
-- title_translations: {"en": "Hello", "vi": "Xin chào", "fr": "Bonjour"}
*/
