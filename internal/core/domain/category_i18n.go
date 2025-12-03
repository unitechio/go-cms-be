package domain

// Update Category to support translations

// NameTranslations        TranslatableContent `gorm:"type:jsonb" json:"name_translations,omitempty"`
// DescriptionTranslations TranslatableContent `gorm:"type:jsonb" json:"description_translations,omitempty"`

// GetName returns translated name
func (c *Category) GetName(locale string) string {
	if c.NameTranslations != nil {
		return c.NameTranslations.Get(locale, LocaleDefault)
	}
	return c.Name
}

// GetDescription returns translated description
func (c *Category) GetDescription(locale string) string {
	if c.DescriptionTranslations != nil {
		return c.DescriptionTranslations.Get(locale, LocaleDefault)
	}
	return c.Description
}

// Migration SQL:
/*
ALTER TABLE categories ADD COLUMN name_translations JSONB;
ALTER TABLE categories ADD COLUMN description_translations JSONB;
*/
