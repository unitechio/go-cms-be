package domain

import (
	"time"

	"github.com/google/uuid"
)

// PostStatus represents the status of a post
type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusScheduled PostStatus = "scheduled"
	PostStatusPublished PostStatus = "published"
	PostStatusArchived  PostStatus = "archived"
)

// Post represents a content post
type Post struct {
	BaseModel
	Title           string     `gorm:"size:500;not null" json:"title"`
	Slug            string     `gorm:"uniqueIndex;size:500;not null" json:"slug"`
	Content         string     `gorm:"type:text" json:"content"`
	Excerpt         string     `gorm:"type:text" json:"excerpt"`
	FeaturedImage   string     `json:"featured_image"`
	Status          PostStatus `gorm:"type:varchar(20);default:'draft'" json:"status"`
	AuthorID        uuid.UUID  `gorm:"type:uuid;not null" json:"author_id"`
	PublishedAt     *time.Time `json:"published_at,omitempty"`
	ScheduledAt     *time.Time `json:"scheduled_at,omitempty"`
	ViewCount       int64      `gorm:"default:0" json:"view_count"`
	LikeCount       int64      `gorm:"default:0" json:"like_count"`
	CommentCount    int64      `gorm:"default:0" json:"comment_count"`
	MetaTitle       string     `gorm:"size:500" json:"meta_title"`
	MetaDescription string     `gorm:"type:text" json:"meta_description"`
	MetaKeywords    string     `json:"meta_keywords"`
	Tags            string     `json:"tags"`       // JSON array of tags
	Categories      string     `json:"categories"` // JSON array of category IDs

	// Relationships
	Author User    `gorm:"foreignKey:AuthorID;references:ID" json:"author,omitempty"`
	Media  []Media `gorm:"many2many:post_media;" json:"media,omitempty"`
}

// TableName specifies the table name for Post
func (Post) TableName() string {
	return "posts"
}

// IsPublished checks if the post is published
func (p *Post) IsPublished() bool {
	return p.Status == PostStatusPublished && p.PublishedAt != nil && p.PublishedAt.Before(time.Now())
}

// IsScheduled checks if the post is scheduled
func (p *Post) IsScheduled() bool {
	return p.Status == PostStatusScheduled && p.ScheduledAt != nil && p.ScheduledAt.After(time.Now())
}

// MediaType represents the type of media
type MediaType string

const (
	MediaTypeImage    MediaType = "image"
	MediaTypeVideo    MediaType = "video"
	MediaTypeDocument MediaType = "document"
	MediaTypeAudio    MediaType = "audio"
	MediaTypeOther    MediaType = "other"
)

// Media represents uploaded media files
type Media struct {
	BaseModel
	FileName     string    `gorm:"size:500;not null" json:"file_name"`
	OriginalName string    `gorm:"size:500;not null" json:"original_name"`
	FilePath     string    `gorm:"not null" json:"file_path"`
	FileSize     int64     `gorm:"not null" json:"file_size"`
	MimeType     string    `gorm:"size:100;not null" json:"mime_type"`
	Type         MediaType `gorm:"type:varchar(20);not null" json:"type"`
	Width        int       `json:"width,omitempty"`     // For images/videos
	Height       int       `json:"height,omitempty"`    // For images/videos
	Duration     int       `json:"duration,omitempty"`  // For videos/audio (in seconds)
	Thumbnail    string    `json:"thumbnail,omitempty"` // Thumbnail path for videos
	UploadedBy   uuid.UUID `gorm:"type:uuid;not null" json:"uploaded_by"`
	Bucket       string    `gorm:"size:100" json:"bucket"`     // MinIO bucket
	ObjectKey    string    `gorm:"size:500" json:"object_key"` // MinIO object key
	URL          string    `json:"url,omitempty"`              // Public URL
	Alt          string    `gorm:"size:500" json:"alt"`        // Alt text for images
	Caption      string    `gorm:"type:text" json:"caption"`
	Description  string    `gorm:"type:text" json:"description"`
	Tags         string    `json:"tags"` // JSON array of tags

	// Relationships
	Uploader User   `gorm:"foreignKey:UploadedBy;references:ID" json:"uploader,omitempty"`
	Posts    []Post `gorm:"many2many:post_media;" json:"posts,omitempty"`
}

// TableName specifies the table name for Media
func (Media) TableName() string {
	return "media"
}

// PostMedia represents the many-to-many relationship between posts and media
type PostMedia struct {
	PostID    uint      `gorm:"primaryKey" json:"post_id"`
	MediaID   uint      `gorm:"primaryKey" json:"media_id"`
	Order     int       `gorm:"default:0" json:"order"` // Display order
	CreatedAt time.Time `json:"created_at"`
}

// TableName specifies the table name for PostMedia
func (PostMedia) TableName() string {
	return "post_media"
}

// CategoryType represents the type of category
type CategoryType string

const (
	CategoryTypeBlog    CategoryType = "blog"
	CategoryTypeHeader  CategoryType = "header"
	CategoryTypeFooter  CategoryType = "footer"
	CategoryTypeSidebar CategoryType = "sidebar"
)

// CategoryStatus represents the status of a category
type CategoryStatus string

const (
	CategoryStatusActive   CategoryStatus = "active"
	CategoryStatusInactive CategoryStatus = "inactive"
)

// Category represents a content category
type Category struct {
	BaseModel
	Name        string         `gorm:"size:200;not null" json:"name"`
	Slug        string         `gorm:"uniqueIndex;size:200;not null" json:"slug"`
	Description string         `gorm:"type:text" json:"description"`
	ParentID    *uint          `json:"parent_id,omitempty"`
	Order       int            `gorm:"default:0" json:"order"`
	Type        CategoryType   `gorm:"type:varchar(20);default:'blog';not null" json:"type"`
	Status      CategoryStatus `gorm:"type:varchar(20);default:'active';not null" json:"status"`
	Icon        string         `json:"icon"`
	Color       string         `gorm:"size:20" json:"color"`

	// Relationships
	Parent        *Category  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children      []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	ChildrenCount int        `gorm:"-" json:"children_count,omitempty"` // Computed field
}

// TableName specifies the table name for Category
func (Category) TableName() string {
	return "categories"
}

// Tag represents a content tag
type Tag struct {
	BaseModel
	Name  string `gorm:"uniqueIndex;size:100;not null" json:"name"`
	Slug  string `gorm:"uniqueIndex;size:100;not null" json:"slug"`
	Color string `gorm:"size:20" json:"color"`
}

// TableName specifies the table name for Tag
func (Tag) TableName() string {
	return "tags"
}

// PostSchedule represents a scheduled post job
type PostSchedule struct {
	BaseModel
	PostID      uint       `gorm:"uniqueIndex;not null" json:"post_id"`
	ScheduledAt time.Time  `gorm:"not null" json:"scheduled_at"`
	Status      string     `gorm:"type:varchar(20);default:'pending'" json:"status"` // pending, processing, completed, failed
	ExecutedAt  *time.Time `json:"executed_at,omitempty"`
	Error       string     `gorm:"type:text" json:"error,omitempty"`
	RetryCount  int        `gorm:"default:0" json:"retry_count"`

	// Relationships
	Post Post `gorm:"foreignKey:PostID" json:"post,omitempty"`
}

// TableName specifies the table name for PostSchedule
func (PostSchedule) TableName() string {
	return "post_schedules"
}
