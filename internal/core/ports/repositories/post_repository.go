package repositories

import (
	"context"

	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/pkg/pagination"
)

// PostRepository defines the interface for post data operations
type PostRepository interface {
	// Basic CRUD
	Create(ctx context.Context, post *domain.Post) error
	GetByID(ctx context.Context, id uint) (*domain.Post, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Post, error)
	Update(ctx context.Context, post *domain.Post) error
	Delete(ctx context.Context, id uint) error

	// List operations
	List(ctx context.Context, filter PostFilter, page *pagination.OffsetPagination) ([]*domain.Post, int64, error)
	ListWithCursor(ctx context.Context, filter PostFilter, cursor *pagination.Cursor, limit int) ([]*domain.Post, *pagination.Cursor, error)

	// Status operations
	UpdateStatus(ctx context.Context, postID uint, status domain.PostStatus) error
	Publish(ctx context.Context, postID uint) error
	Schedule(ctx context.Context, postID uint, scheduledAt string) error

	// Media operations
	AttachMedia(ctx context.Context, postID, mediaID uint, order int) error
	DetachMedia(ctx context.Context, postID, mediaID uint) error
	GetPostMedia(ctx context.Context, postID uint) ([]*domain.Media, error)

	// Statistics
	IncrementViewCount(ctx context.Context, postID uint) error
	IncrementLikeCount(ctx context.Context, postID uint) error
	IncrementCommentCount(ctx context.Context, postID uint) error

	// Author operations
	GetByAuthor(ctx context.Context, authorID uint) ([]*domain.Post, error)
}

// PostFilter represents filters for post queries
type PostFilter struct {
	Status     domain.PostStatus
	AuthorID   *uint
	Search     string // Search in title, content, excerpt
	Tags       []string
	Categories []string
	IDs        []uint
}

// MediaRepository defines the interface for media data operations
type MediaRepository interface {
	// Basic CRUD
	Create(ctx context.Context, media *domain.Media) error
	GetByID(ctx context.Context, id uint) (*domain.Media, error)
	Update(ctx context.Context, media *domain.Media) error
	Delete(ctx context.Context, id uint) error

	// List operations
	List(ctx context.Context, filter MediaFilter, page *pagination.OffsetPagination) ([]*domain.Media, int64, error)
	ListWithCursor(ctx context.Context, filter MediaFilter, cursor *pagination.Cursor, limit int) ([]*domain.Media, *pagination.Cursor, error)

	// Type operations
	GetByType(ctx context.Context, mediaType domain.MediaType) ([]*domain.Media, error)

	// Uploader operations
	GetByUploader(ctx context.Context, uploaderID uint) ([]*domain.Media, error)

	// Storage operations
	GetByBucket(ctx context.Context, bucket string) ([]*domain.Media, error)
	GetByObjectKey(ctx context.Context, objectKey string) (*domain.Media, error)
}

// MediaFilter represents filters for media queries
type MediaFilter struct {
	Type       domain.MediaType
	UploaderID *uint
	Search     string // Search in file_name, original_name
	Tags       []string
	IDs        []uint
}

// CategoryRepository defines the interface for category data operations
type CategoryRepository interface {
	// Basic CRUD
	Create(ctx context.Context, category *domain.Category) error
	GetByID(ctx context.Context, id uint) (*domain.Category, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Category, error)
	Update(ctx context.Context, category *domain.Category) error
	Delete(ctx context.Context, id uint) error

	// List operations
	List(ctx context.Context) ([]*domain.Category, error)
	GetHierarchy(ctx context.Context) ([]*domain.Category, error)
	GetChildren(ctx context.Context, parentID uint) ([]*domain.Category, error)
	GetRoots(ctx context.Context) ([]*domain.Category, error)
}

// TagRepository defines the interface for tag data operations
type TagRepository interface {
	// Basic CRUD
	Create(ctx context.Context, tag *domain.Tag) error
	GetByID(ctx context.Context, id uint) (*domain.Tag, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Tag, error)
	GetByName(ctx context.Context, name string) (*domain.Tag, error)
	Update(ctx context.Context, tag *domain.Tag) error
	Delete(ctx context.Context, id uint) error

	// List operations
	List(ctx context.Context) ([]*domain.Tag, error)
	Search(ctx context.Context, query string) ([]*domain.Tag, error)
}

// PostScheduleRepository defines the interface for post schedule data operations
type PostScheduleRepository interface {
	// Basic CRUD
	Create(ctx context.Context, schedule *domain.PostSchedule) error
	GetByID(ctx context.Context, id uint) (*domain.PostSchedule, error)
	GetByPostID(ctx context.Context, postID uint) (*domain.PostSchedule, error)
	Update(ctx context.Context, schedule *domain.PostSchedule) error
	Delete(ctx context.Context, id uint) error

	// List operations
	GetPendingSchedules(ctx context.Context) ([]*domain.PostSchedule, error)
	GetSchedulesByStatus(ctx context.Context, status string) ([]*domain.PostSchedule, error)

	// Execution
	MarkAsExecuted(ctx context.Context, id uint) error
	MarkAsFailed(ctx context.Context, id uint, errorMsg string) error
	IncrementRetryCount(ctx context.Context, id uint) error
}
