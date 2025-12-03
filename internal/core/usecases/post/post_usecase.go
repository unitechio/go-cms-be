package post

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/pkg/errors"
	"github.com/owner/go-cms/pkg/logger"
	"github.com/owner/go-cms/pkg/pagination"
	"go.uber.org/zap"
)

// UseCase defines the post use case interface
type UseCase interface {
	// Post CRUD
	CreatePost(ctx context.Context, req CreatePostRequest) (*domain.Post, error)
	GetPost(ctx context.Context, id uint) (*domain.Post, error)
	GetPostBySlug(ctx context.Context, slug string) (*domain.Post, error)
	UpdatePost(ctx context.Context, id uint, req UpdatePostRequest) (*domain.Post, error)
	DeletePost(ctx context.Context, id uint) error
	ListPosts(ctx context.Context, filter repositories.PostFilter, page *pagination.OffsetPagination) ([]*domain.Post, int64, error)

	// Post Status
	PublishPost(ctx context.Context, id uint) error
	SchedulePost(ctx context.Context, id uint, scheduledAt time.Time) error
	ArchivePost(ctx context.Context, id uint) error

	// Media Management
	AttachMedia(ctx context.Context, postID, mediaID uint, order int) error
	DetachMedia(ctx context.Context, postID, mediaID uint) error
	GetPostMedia(ctx context.Context, postID uint) ([]*domain.Media, error)

	// Statistics
	IncrementViewCount(ctx context.Context, postID uint) error
}

// useCase implements the UseCase interface
type useCase struct {
	postRepo  repositories.PostRepository
	mediaRepo repositories.MediaRepository
	userRepo  repositories.UserRepository
}

// NewUseCase creates a new post use case
func NewUseCase(
	postRepo repositories.PostRepository,
	mediaRepo repositories.MediaRepository,
	userRepo repositories.UserRepository,
) UseCase {
	return &useCase{
		postRepo:  postRepo,
		mediaRepo: mediaRepo,
		userRepo:  userRepo,
	}
}

// CreatePostRequest represents a create post request
type CreatePostRequest struct {
	Title           string    `json:"title" binding:"required"`
	Content         string    `json:"content"`
	Excerpt         string    `json:"excerpt"`
	FeaturedImage   string    `json:"featured_image"`
	Status          string    `json:"status"` // draft, published, scheduled
	AuthorID        uuid.UUID `json:"author_id" binding:"required"`
	MetaTitle       string    `json:"meta_title"`
	MetaDescription string    `json:"meta_description"`
	MetaKeywords    string    `json:"meta_keywords"`
	Tags            string    `json:"tags"`
	Categories      string    `json:"categories"`
}

// UpdatePostRequest represents an update post request
type UpdatePostRequest struct {
	Title           *string `json:"title"`
	Content         *string `json:"content"`
	Excerpt         *string `json:"excerpt"`
	FeaturedImage   *string `json:"featured_image"`
	Status          *string `json:"status"`
	MetaTitle       *string `json:"meta_title"`
	MetaDescription *string `json:"meta_description"`
	MetaKeywords    *string `json:"meta_keywords"`
	Tags            *string `json:"tags"`
	Categories      *string `json:"categories"`
}

// CreatePost creates a new post
func (uc *useCase) CreatePost(ctx context.Context, req CreatePostRequest) (*domain.Post, error) {
	// Validate author exists
	author, err := uc.userRepo.GetByID(ctx, req.AuthorID)
	if err != nil {
		logger.Error("Author not found", zap.Error(err), zap.String("author_id", req.AuthorID.String()))
		return nil, errors.Wrap(err, errors.ErrCodeNotFound, "author not found", 404)
	}

	// Generate slug from title
	postSlug := slug.Make(req.Title)

	// Check if slug already exists
	existingPost, err := uc.postRepo.GetBySlug(ctx, postSlug)
	if err == nil && existingPost != nil {
		// Append timestamp to make slug unique
		postSlug = fmt.Sprintf("%s-%d", postSlug, time.Now().Unix())
	}

	// Validate status
	status := domain.PostStatus(req.Status)
	if status == "" {
		status = domain.PostStatusDraft
	}

	// Create post
	post := &domain.Post{
		Title:           req.Title,
		Slug:            postSlug,
		Content:         req.Content,
		Excerpt:         req.Excerpt,
		FeaturedImage:   req.FeaturedImage,
		Status:          status,
		AuthorID:        req.AuthorID,
		MetaTitle:       req.MetaTitle,
		MetaDescription: req.MetaDescription,
		MetaKeywords:    req.MetaKeywords,
		Tags:            req.Tags,
		Categories:      req.Categories,
	}

	// Set published_at if status is published
	if status == domain.PostStatusPublished {
		now := time.Now()
		post.PublishedAt = &now
	}

	if err := uc.postRepo.Create(ctx, post); err != nil {
		logger.Error("Failed to create post", zap.Error(err))
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to create post", 500)
	}

	// Load author relationship
	post.Author = *author

	logger.Info("Post created successfully",
		zap.Uint("post_id", post.ID),
		zap.String("title", post.Title),
		zap.String("author", author.Email))

	return post, nil
}

// GetPost gets a post by ID
func (uc *useCase) GetPost(ctx context.Context, id uint) (*domain.Post, error) {
	post, err := uc.postRepo.GetByID(ctx, id)
	if err != nil {
		logger.Error("Failed to get post", zap.Error(err), zap.Uint("id", id))
		return nil, err
	}
	return post, nil
}

// GetPostBySlug gets a post by slug
func (uc *useCase) GetPostBySlug(ctx context.Context, slug string) (*domain.Post, error) {
	post, err := uc.postRepo.GetBySlug(ctx, slug)
	if err != nil {
		logger.Error("Failed to get post by slug", zap.Error(err), zap.String("slug", slug))
		return nil, err
	}
	return post, nil
}

// UpdatePost updates a post
func (uc *useCase) UpdatePost(ctx context.Context, id uint, req UpdatePostRequest) (*domain.Post, error) {
	// Get existing post
	post, err := uc.postRepo.GetByID(ctx, id)
	if err != nil {
		logger.Error("Post not found", zap.Error(err), zap.Uint("id", id))
		return nil, err
	}

	// Update fields if provided
	if req.Title != nil {
		post.Title = *req.Title
		// Regenerate slug if title changed
		newSlug := slug.Make(*req.Title)
		if newSlug != post.Slug {
			// Check if new slug already exists
			existingPost, _ := uc.postRepo.GetBySlug(ctx, newSlug)
			if existingPost != nil && existingPost.ID != post.ID {
				newSlug = fmt.Sprintf("%s-%d", newSlug, time.Now().Unix())
			}
			post.Slug = newSlug
		}
	}

	if req.Content != nil {
		post.Content = *req.Content
	}

	if req.Excerpt != nil {
		post.Excerpt = *req.Excerpt
	}

	if req.FeaturedImage != nil {
		post.FeaturedImage = *req.FeaturedImage
	}

	if req.Status != nil {
		post.Status = domain.PostStatus(*req.Status)
	}

	if req.MetaTitle != nil {
		post.MetaTitle = *req.MetaTitle
	}

	if req.MetaDescription != nil {
		post.MetaDescription = *req.MetaDescription
	}

	if req.MetaKeywords != nil {
		post.MetaKeywords = *req.MetaKeywords
	}

	if req.Tags != nil {
		post.Tags = *req.Tags
	}

	if req.Categories != nil {
		post.Categories = *req.Categories
	}

	// Update post
	if err := uc.postRepo.Update(ctx, post); err != nil {
		logger.Error("Failed to update post", zap.Error(err), zap.Uint("id", id))
		return nil, errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to update post", 500)
	}

	logger.Info("Post updated successfully", zap.Uint("post_id", post.ID))

	return post, nil
}

// DeletePost deletes a post
func (uc *useCase) DeletePost(ctx context.Context, id uint) error {
	// Check if post exists
	_, err := uc.postRepo.GetByID(ctx, id)
	if err != nil {
		logger.Error("Post not found", zap.Error(err), zap.Uint("id", id))
		return err
	}

	if err := uc.postRepo.Delete(ctx, id); err != nil {
		logger.Error("Failed to delete post", zap.Error(err), zap.Uint("id", id))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to delete post", 500)
	}

	logger.Info("Post deleted successfully", zap.Uint("post_id", id))

	return nil
}

// ListPosts lists posts with filters and pagination
func (uc *useCase) ListPosts(ctx context.Context, filter repositories.PostFilter, page *pagination.OffsetPagination) ([]*domain.Post, int64, error) {
	posts, total, err := uc.postRepo.List(ctx, filter, page)
	if err != nil {
		logger.Error("Failed to list posts", zap.Error(err))
		return nil, 0, err
	}
	return posts, total, nil
}

// PublishPost publishes a post
func (uc *useCase) PublishPost(ctx context.Context, id uint) error {
	// Check if post exists
	post, err := uc.postRepo.GetByID(ctx, id)
	if err != nil {
		logger.Error("Post not found", zap.Error(err), zap.Uint("id", id))
		return err
	}

	// Validate post can be published
	if post.Title == "" {
		return errors.New(errors.ErrCodeBadRequest, "post title is required", 400)
	}

	if err := uc.postRepo.Publish(ctx, id); err != nil {
		logger.Error("Failed to publish post", zap.Error(err), zap.Uint("id", id))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to publish post", 500)
	}

	logger.Info("Post published successfully", zap.Uint("post_id", id))

	return nil
}

// SchedulePost schedules a post for future publication
func (uc *useCase) SchedulePost(ctx context.Context, id uint, scheduledAt time.Time) error {
	// Check if post exists
	_, err := uc.postRepo.GetByID(ctx, id)
	if err != nil {
		logger.Error("Post not found", zap.Error(err), zap.Uint("id", id))
		return err
	}

	// Validate scheduled time is in the future
	if scheduledAt.Before(time.Now()) {
		return errors.New(errors.ErrCodeBadRequest, "scheduled time must be in the future", 400)
	}

	if err := uc.postRepo.Schedule(ctx, id, scheduledAt.Format(time.RFC3339)); err != nil {
		logger.Error("Failed to schedule post", zap.Error(err), zap.Uint("id", id))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to schedule post", 500)
	}

	logger.Info("Post scheduled successfully",
		zap.Uint("post_id", id),
		zap.Time("scheduled_at", scheduledAt))

	return nil
}

// ArchivePost archives a post
func (uc *useCase) ArchivePost(ctx context.Context, id uint) error {
	// Check if post exists
	_, err := uc.postRepo.GetByID(ctx, id)
	if err != nil {
		logger.Error("Post not found", zap.Error(err), zap.Uint("id", id))
		return err
	}

	if err := uc.postRepo.UpdateStatus(ctx, id, domain.PostStatusArchived); err != nil {
		logger.Error("Failed to archive post", zap.Error(err), zap.Uint("id", id))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to archive post", 500)
	}

	logger.Info("Post archived successfully", zap.Uint("post_id", id))

	return nil
}

// AttachMedia attaches media to a post
func (uc *useCase) AttachMedia(ctx context.Context, postID, mediaID uint, order int) error {
	// Validate post exists
	_, err := uc.postRepo.GetByID(ctx, postID)
	if err != nil {
		logger.Error("Post not found", zap.Error(err), zap.Uint("post_id", postID))
		return err
	}

	// Validate media exists
	_, err = uc.mediaRepo.GetByID(ctx, mediaID)
	if err != nil {
		logger.Error("Media not found", zap.Error(err), zap.Uint("media_id", mediaID))
		return err
	}

	if err := uc.postRepo.AttachMedia(ctx, postID, mediaID, order); err != nil {
		logger.Error("Failed to attach media to post", zap.Error(err),
			zap.Uint("post_id", postID), zap.Uint("media_id", mediaID))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to attach media", 500)
	}

	logger.Info("Media attached to post successfully",
		zap.Uint("post_id", postID), zap.Uint("media_id", mediaID))

	return nil
}

// DetachMedia detaches media from a post
func (uc *useCase) DetachMedia(ctx context.Context, postID, mediaID uint) error {
	if err := uc.postRepo.DetachMedia(ctx, postID, mediaID); err != nil {
		logger.Error("Failed to detach media from post", zap.Error(err),
			zap.Uint("post_id", postID), zap.Uint("media_id", mediaID))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to detach media", 500)
	}

	logger.Info("Media detached from post successfully",
		zap.Uint("post_id", postID), zap.Uint("media_id", mediaID))

	return nil
}

// GetPostMedia gets all media for a post
func (uc *useCase) GetPostMedia(ctx context.Context, postID uint) ([]*domain.Media, error) {
	media, err := uc.postRepo.GetPostMedia(ctx, postID)
	if err != nil {
		logger.Error("Failed to get post media", zap.Error(err), zap.Uint("post_id", postID))
		return nil, err
	}
	return media, nil
}

// IncrementViewCount increments the view count for a post
func (uc *useCase) IncrementViewCount(ctx context.Context, postID uint) error {
	if err := uc.postRepo.IncrementViewCount(ctx, postID); err != nil {
		logger.Error("Failed to increment view count", zap.Error(err), zap.Uint("post_id", postID))
		return errors.Wrap(err, errors.ErrCodeDatabaseError, "failed to increment view count", 500)
	}
	return nil
}
