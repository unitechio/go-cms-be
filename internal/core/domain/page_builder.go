package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PageStatus represents the status of a page
type PageStatus string

const (
	PageStatusDraft     PageStatus = "draft"
	PageStatusPublished PageStatus = "published"
	PageStatusArchived  PageStatus = "archived"
)

// Page represents a dynamic page
type Page struct {
	UUIDModel
	Title          string     `gorm:"size:500;not null" json:"title"`
	Slug           string     `gorm:"uniqueIndex;size:500;not null" json:"slug"`
	Template       string     `gorm:"size:100;default:'default'" json:"template"`
	Status         PageStatus `gorm:"type:varchar(20);default:'draft'" json:"status"`
	SeoTitle       string     `gorm:"size:500" json:"seo_title"`
	SeoDescription string     `gorm:"type:text" json:"seo_description"`
	OgImage        string     `gorm:"size:500" json:"og_image"`
	AuthorID       uuid.UUID  `gorm:"type:uuid;not null" json:"author_id"`
	PublishedAt    *time.Time `json:"published_at,omitempty"`

	// Relationships
	Author User        `gorm:"foreignKey:AuthorID;references:ID" json:"author,omitempty"`
	Blocks []PageBlock `gorm:"foreignKey:PageID;references:ID;constraint:OnDelete:CASCADE" json:"blocks,omitempty"`
}

// TableName specifies the table name for Page
func (Page) TableName() string {
	return "pages"
}

// Block represents a reusable block definition
type Block struct {
	UUIDModel
	Name            string          `gorm:"size:200;not null" json:"name"`
	Type            string          `gorm:"size:100;not null" json:"type"`
	Category        string          `gorm:"size:100" json:"category"`
	Schema          json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"schema"`
	PreviewTemplate string          `gorm:"type:text" json:"preview_template"`
	IsGlobal        bool            `gorm:"default:false" json:"is_global"`
}

// TableName specifies the table name for Block
func (Block) TableName() string {
	return "blocks"
}

// PageBlock represents a block instance on a page
type PageBlock struct {
	ID            uuid.UUID       `gorm:"type:uuid;primarykey;default:gen_random_uuid()" json:"id"`
	CreatedAt     time.Time       `json:"created_at"`
	PageID        uuid.UUID       `gorm:"type:uuid;not null" json:"page_id"`
	BlockID       uuid.UUID       `gorm:"type:uuid;not null" json:"block_id"`
	ParentBlockID *uuid.UUID      `gorm:"type:uuid" json:"parent_block_id,omitempty"`
	Order         int             `gorm:"default:0" json:"order"`
	Config        json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"config"`
	Language      string          `gorm:"size:10;default:'en'" json:"language"`

	// Relationships
	Block       Block        `gorm:"foreignKey:BlockID;references:ID" json:"block,omitempty"`
	ParentBlock *PageBlock   `gorm:"foreignKey:ParentBlockID;references:ID" json:"parent_block,omitempty"`
	Children    []*PageBlock `gorm:"foreignKey:ParentBlockID;references:ID" json:"children,omitempty"`
}

// TableName specifies the table name for PageBlock
func (PageBlock) TableName() string {
	return "page_blocks"
}

// BeforeCreate hook to generate UUID if not set
func (u *PageBlock) BeforeCreate(tx *gorm.DB) error {
	if u.ID == (uuid.UUID{}) {
		u.ID = uuid.New()
	}
	return nil
}

// PageVersion represents a snapshot of a page version
type PageVersion struct {
	ID                uuid.UUID       `gorm:"type:uuid;primarykey;default:gen_random_uuid()" json:"id"`
	CreatedAt         time.Time       `json:"created_at"`
	PageID            uuid.UUID       `gorm:"type:uuid;not null" json:"page_id"`
	VersionNumber     int             `gorm:"not null" json:"version_number"`
	Title             string          `gorm:"size:500;not null" json:"title"`
	Slug              string          `gorm:"size:500;not null" json:"slug"`
	BlocksSnapshot    json.RawMessage `gorm:"type:jsonb;not null" json:"blocks_snapshot"`
	CreatedBy         uuid.UUID       `gorm:"type:uuid;not null" json:"created_by"`
	ChangeDescription string          `gorm:"type:text" json:"change_description"`

	// Relationships
	Creator User `gorm:"foreignKey:CreatedBy;references:ID" json:"creator,omitempty"`
}

// TableName specifies the table name for PageVersion
func (PageVersion) TableName() string {
	return "page_versions"
}

// BeforeCreate hook to generate UUID if not set
func (u *PageVersion) BeforeCreate(tx *gorm.DB) error {
	if u.ID == (uuid.UUID{}) {
		u.ID = uuid.New()
	}
	return nil
}

// ThemeSetting represents theme configuration
type ThemeSetting struct {
	UUIDModel
	Name     string          `gorm:"size:200;not null" json:"name"`
	Config   json.RawMessage `gorm:"type:jsonb;default:'{}'" json:"config"`
	IsActive bool            `gorm:"default:false" json:"is_active"`
}

// TableName specifies the table name for ThemeSetting
func (ThemeSetting) TableName() string {
	return "theme_settings"
}
