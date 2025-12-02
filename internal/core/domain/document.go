package domain

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
)

// Document represents a file in the system
type Document struct {
	ID           uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	DocumentCode string     `json:"document_code" gorm:"size:100;not null;uniqueIndex"`
	EntityType   string     `json:"entity_type" gorm:"size:50;index"` // "order", "customer", "contract", etc.
	EntityID     uint       `json:"entity_id" gorm:"index"`           // ID of the related entity
	DocumentName string     `json:"document_name" gorm:"size:255;not null"`
	DocumentPath string     `json:"document_path" gorm:"size:500;not null"`
	DocumentType string     `json:"document_type" gorm:"size:100;not null"` // MIME type or file extension
	FileSize     int64      `json:"file_size" gorm:"not null"`              // Size in bytes
	UploadedBy   uuid.UUID  `json:"uploaded_by" gorm:"type:char(36);not null;index"`
	CreatedAt    time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    *time.Time `json:"deleted_at" gorm:"index"`

	// Polymorphic relation
	AttachableID   uint   `json:"-"`
	AttachableType string `json:"-"`

	// Relations
	Uploader            User                 `json:"uploader" gorm:"foreignKey:UploadedBy"`
	DocumentPermissions []DocumentPermission `json:"document_permissions" gorm:"foreignKey:DocumentID"`
}

// DocumentPermission defines access control for documents
type DocumentPermission struct {
	ID              uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	DocumentID      uint      `json:"document_id" gorm:"not null;index"`
	UserID          uuid.UUID `json:"user_id" gorm:"type:char(36);not null;index"` // If null, applies to a role
	JobTitle        string    `json:"job_title" gorm:"index"`                      // If null, applies to a specific user
	PermissionLevel string    `json:"permission_level" gorm:"size:20;not null"`    // view, edit, comment, owner
	CreatedBy       uuid.UUID `json:"created_by" gorm:"type:char(36);not null;index"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	Document Document `json:"-" gorm:"foreignKey:DocumentID"`
	User     User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Creator  User     `json:"creator" gorm:"foreignKey:CreatedBy"`
}

// DocumentComment allows users to add comments to documents
type DocumentComment struct {
	ID         uint       `json:"id" gorm:"primaryKey;autoIncrement"`
	DocumentID uint       `json:"document_id" gorm:"not null;index"`
	UserID     uuid.UUID  `json:"user_id" gorm:"type:char(36);not null;index"` // If null, applies to a role
	Comment    string     `json:"comment" gorm:"type:text;not null"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt  *time.Time `json:"deleted_at" gorm:"index"`

	// Relations
	Document Document `json:"-" gorm:"foreignKey:DocumentID"`
	User     User     `json:"user" gorm:"foreignKey:UserID"`
}

// DocumentVersion tracks document revisions
type DocumentVersion struct {
	ID            uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	DocumentID    uint      `json:"document_id" gorm:"not null;index"`
	VersionNumber int       `json:"version_number" gorm:"not null"`
	DocumentPath  string    `json:"document_path" gorm:"size:500;not null"`
	FileSize      int64     `json:"file_size" gorm:"not null"`
	ChangedBy     uuid.UUID `json:"changed_by" gorm:"type:char(36);not null;index"`
	ChangeNote    string    `json:"change_note" gorm:"type:text"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`

	// Relations
	Document Document `json:"-" gorm:"foreignKey:DocumentID"`
	User     User     `json:"user" gorm:"foreignKey:ChangedBy"`
}

// Constants for permission levels
const (
	PermissionView    = "view"
	PermissionComment = "comment"
	PermissionEdit    = "edit"
	PermissionOwner   = "owner"
)

type DocumentRepository interface {
	Create(ctx context.Context, doc *Document) error
	GetByID(ctx context.Context, id string) (*Document, error)
	List(ctx context.Context, userID string, limit, offset int) ([]*Document, int64, error)
	Delete(ctx context.Context, id string) error
}

type StorageService interface {
	Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error
	GetPresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error)
	Delete(ctx context.Context, objectName string) error
}
