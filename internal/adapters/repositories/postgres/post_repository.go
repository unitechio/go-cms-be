package postgres

import (
	"context"
	"time"

	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/pkg/errors"
	"github.com/owner/go-cms/pkg/logger"
	"github.com/owner/go-cms/pkg/pagination"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type postRepository struct {
	db *gorm.DB
}

// NewPostRepository creates a new post repository
func NewPostRepository(db *gorm.DB) repositories.PostRepository {
	return &postRepository{db: db}
}

// Create creates a new post
func (r *postRepository) Create(ctx context.Context, post *domain.Post) error {
	if err := r.db.WithContext(ctx).Create(post).Error; err != nil {
		logger.Error("Failed to create post", zap.Error(err))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to create post", 500)
	}
	return nil
}

// GetByID gets a post by ID
func (r *postRepository) GetByID(ctx context.Context, id uint) (*domain.Post, error) {
	var post domain.Post
	if err := r.db.WithContext(ctx).
		Preload("Author").
		Preload("Media").
		First(&post, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		logger.Error("Failed to get post by ID", zap.Error(err), zap.Uint("id", id))
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get post", 500)
	}
	return &post, nil
}

// GetBySlug gets a post by slug
func (r *postRepository) GetBySlug(ctx context.Context, slug string) (*domain.Post, error) {
	var post domain.Post
	if err := r.db.WithContext(ctx).
		Preload("Author").
		Preload("Media").
		Where("slug = ?", slug).
		First(&post).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.ErrNotFound
		}
		logger.Error("Failed to get post by slug", zap.Error(err), zap.String("slug", slug))
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get post", 500)
	}
	return &post, nil
}

// Update updates a post
func (r *postRepository) Update(ctx context.Context, post *domain.Post) error {
	if err := r.db.WithContext(ctx).Save(post).Error; err != nil {
		logger.Error("Failed to update post", zap.Error(err), zap.Uint("id", post.ID))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to update post", 500)
	}
	return nil
}

// Delete deletes a post (soft delete)
func (r *postRepository) Delete(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&domain.Post{}, id).Error; err != nil {
		logger.Error("Failed to delete post", zap.Error(err), zap.Uint("id", id))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to delete post", 500)
	}
	return nil
}

// List lists posts with offset pagination
func (r *postRepository) List(ctx context.Context, filter repositories.PostFilter, page *pagination.OffsetPagination) ([]*domain.Post, int64, error) {
	var posts []*domain.Post
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Post{})
	query = r.applyFilters(query, filter)

	// Count total
	if err := query.Count(&total).Error; err != nil {
		logger.Error("Failed to count posts", zap.Error(err))
		return nil, 0, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to count posts", 500)
	}

	// Apply pagination
	if page != nil {
		query = query.Offset(page.GetOffset()).Limit(page.Limit)
	}

	// Fetch posts
	if err := query.
		Preload("Author").
		Order("created_at DESC").
		Find(&posts).Error; err != nil {
		logger.Error("Failed to list posts", zap.Error(err))
		return nil, 0, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to list posts", 500)
	}

	return posts, total, nil
}

// ListWithCursor lists posts with cursor pagination
func (r *postRepository) ListWithCursor(ctx context.Context, filter repositories.PostFilter, cursor *pagination.Cursor, limit int) ([]*domain.Post, *pagination.Cursor, error) {
	var posts []*domain.Post

	query := r.db.WithContext(ctx).Model(&domain.Post{})
	query = r.applyFilters(query, filter)

	// Apply cursor
	if cursor != nil && cursor.After != "" {
		query = query.Where("id > ?", cursor.After)
	}

	// Fetch posts
	if err := query.
		Preload("Author").
		Order("id ASC").
		Limit(limit + 1).
		Find(&posts).Error; err != nil {
		logger.Error("Failed to list posts with cursor", zap.Error(err))
		return nil, nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to list posts", 500)
	}

	// Build next cursor
	var nextCursor *pagination.Cursor
	if len(posts) > limit {
		posts = posts[:limit]
		nextCursor = &pagination.Cursor{
			After: string(rune(posts[len(posts)-1].ID)),
		}
	}

	return posts, nextCursor, nil
}

// applyFilters applies filters to query
func (r *postRepository) applyFilters(query *gorm.DB, filter repositories.PostFilter) *gorm.DB {
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.AuthorID != nil {
		query = query.Where("author_id = ?", *filter.AuthorID)
	}

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("title ILIKE ? OR content ILIKE ? OR excerpt ILIKE ?",
			searchPattern, searchPattern, searchPattern)
	}

	if len(filter.IDs) > 0 {
		query = query.Where("id IN ?", filter.IDs)
	}

	// Note: Tags and Categories are stored as JSON strings
	// For proper filtering, consider using JSONB in PostgreSQL
	if len(filter.Tags) > 0 {
		for _, tag := range filter.Tags {
			query = query.Where("tags LIKE ?", "%"+tag+"%")
		}
	}

	if len(filter.Categories) > 0 {
		for _, category := range filter.Categories {
			query = query.Where("categories LIKE ?", "%"+category+"%")
		}
	}

	return query
}

// UpdateStatus updates the post status
func (r *postRepository) UpdateStatus(ctx context.Context, postID uint, status domain.PostStatus) error {
	if err := r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("id = ?", postID).
		Update("status", status).Error; err != nil {
		logger.Error("Failed to update post status", zap.Error(err), zap.Uint("id", postID))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to update post status", 500)
	}
	return nil
}

// Publish publishes a post
func (r *postRepository) Publish(ctx context.Context, postID uint) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("id = ?", postID).
		Updates(map[string]interface{}{
			"status":       domain.PostStatusPublished,
			"published_at": now,
		}).Error; err != nil {
		logger.Error("Failed to publish post", zap.Error(err), zap.Uint("id", postID))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to publish post", 500)
	}
	return nil
}

// Schedule schedules a post
func (r *postRepository) Schedule(ctx context.Context, postID uint, scheduledAt string) error {
	scheduledTime, err := time.Parse(time.RFC3339, scheduledAt)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeBadRequest, "invalid scheduled time format", 400)
	}

	if err := r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("id = ?", postID).
		Updates(map[string]interface{}{
			"status":       domain.PostStatusScheduled,
			"scheduled_at": scheduledTime,
		}).Error; err != nil {
		logger.Error("Failed to schedule post", zap.Error(err), zap.Uint("id", postID))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to schedule post", 500)
	}
	return nil
}

// AttachMedia attaches media to a post
func (r *postRepository) AttachMedia(ctx context.Context, postID, mediaID uint, order int) error {
	postMedia := &domain.PostMedia{
		PostID:  postID,
		MediaID: mediaID,
		Order:   order,
	}

	if err := r.db.WithContext(ctx).Create(postMedia).Error; err != nil {
		logger.Error("Failed to attach media to post", zap.Error(err),
			zap.Uint("post_id", postID), zap.Uint("media_id", mediaID))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to attach media", 500)
	}
	return nil
}

// DetachMedia detaches media from a post
func (r *postRepository) DetachMedia(ctx context.Context, postID, mediaID uint) error {
	if err := r.db.WithContext(ctx).
		Where("post_id = ? AND media_id = ?", postID, mediaID).
		Delete(&domain.PostMedia{}).Error; err != nil {
		logger.Error("Failed to detach media from post", zap.Error(err),
			zap.Uint("post_id", postID), zap.Uint("media_id", mediaID))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to detach media", 500)
	}
	return nil
}

// GetPostMedia gets all media for a post
func (r *postRepository) GetPostMedia(ctx context.Context, postID uint) ([]*domain.Media, error) {
	var media []*domain.Media

	if err := r.db.WithContext(ctx).
		Joins("JOIN post_media ON post_media.media_id = media.id").
		Where("post_media.post_id = ?", postID).
		Order("post_media.order ASC").
		Find(&media).Error; err != nil {
		logger.Error("Failed to get post media", zap.Error(err), zap.Uint("post_id", postID))
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get post media", 500)
	}

	return media, nil
}

// IncrementViewCount increments the view count
func (r *postRepository) IncrementViewCount(ctx context.Context, postID uint) error {
	if err := r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("id = ?", postID).
		UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error; err != nil {
		logger.Error("Failed to increment view count", zap.Error(err), zap.Uint("id", postID))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to increment view count", 500)
	}
	return nil
}

// IncrementLikeCount increments the like count
func (r *postRepository) IncrementLikeCount(ctx context.Context, postID uint) error {
	if err := r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("id = ?", postID).
		UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error; err != nil {
		logger.Error("Failed to increment like count", zap.Error(err), zap.Uint("id", postID))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to increment like count", 500)
	}
	return nil
}

// IncrementCommentCount increments the comment count
func (r *postRepository) IncrementCommentCount(ctx context.Context, postID uint) error {
	if err := r.db.WithContext(ctx).
		Model(&domain.Post{}).
		Where("id = ?", postID).
		UpdateColumn("comment_count", gorm.Expr("comment_count + ?", 1)).Error; err != nil {
		logger.Error("Failed to increment comment count", zap.Error(err), zap.Uint("id", postID))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to increment comment count", 500)
	}
	return nil
}

// GetByAuthor gets all posts by an author
func (r *postRepository) GetByAuthor(ctx context.Context, authorID uint) ([]*domain.Post, error) {
	var posts []*domain.Post

	if err := r.db.WithContext(ctx).
		Where("author_id = ?", authorID).
		Preload("Author").
		Order("created_at DESC").
		Find(&posts).Error; err != nil {
		logger.Error("Failed to get posts by author", zap.Error(err), zap.Uint("author_id", authorID))
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to get posts by author", 500)
	}

	return posts, nil
}
