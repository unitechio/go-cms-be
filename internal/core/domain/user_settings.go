package domain

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// SettingValue represents a JSON value for settings
type SettingValue map[string]interface{}

// Scan implements sql.Scanner interface
func (s *SettingValue) Scan(value interface{}) error {
	if value == nil {
		*s = make(SettingValue)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, s)
}

// Value implements driver.Valuer interface
func (s SettingValue) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// UserSetting represents a user's setting
type UserSetting struct {
	ID           uint         `gorm:"primaryKey" json:"id"`
	UserID       uuid.UUID    `gorm:"type:uuid;not null;index" json:"user_id"`
	SettingKey   string       `gorm:"size:100;not null" json:"setting_key"`
	SettingValue SettingValue `gorm:"type:jsonb;not null" json:"setting_value"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`

	// Relationships
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName specifies the table name
func (UserSetting) TableName() string {
	return "user_settings"
}

// Predefined setting keys
const (
	SettingKeyLanguage      = "language"
	SettingKeyTheme         = "theme"
	SettingKeyNotifications = "notifications"
	SettingKeyDisplay       = "display"
)

// LanguageSetting represents language preferences
type LanguageSetting struct {
	Locale string `json:"locale"` // en, vi, fr, etc.
}

// NotificationSetting represents notification preferences
type NotificationSetting struct {
	Email     bool   `json:"email"`
	Push      bool   `json:"push"`
	Frequency string `json:"frequency"` // instant, daily, weekly
	Types     struct {
		Posts     bool `json:"posts"`
		Comments  bool `json:"comments"`
		Documents bool `json:"documents"`
		System    bool `json:"system"`
	} `json:"types"`
}

// DisplaySetting represents display preferences
type DisplaySetting struct {
	Density string `json:"density"` // compact, comfortable, spacious
	Sidebar string `json:"sidebar"` // expanded, collapsed
}
