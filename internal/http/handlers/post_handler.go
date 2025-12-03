package handlers

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/internal/core/usecases/post"
	"github.com/owner/go-cms/internal/http/middleware"
	"github.com/owner/go-cms/pkg/pagination"
	"github.com/owner/go-cms/pkg/response"
)

// PostHandler handles post-related HTTP requests
type PostHandler struct {
	postUseCase post.UseCase
}

// NewPostHandler creates a new post handler
func NewPostHandler(postUseCase post.UseCase) *PostHandler {
	return &PostHandler{
		postUseCase: postUseCase,
	}
}

// ListPosts godoc
// @Summary List posts
// @Description Get list of posts with pagination and filters
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Post status (draft, published, scheduled, archived)"
// @Param author_id query string false "Author UUID"
// @Param search query string false "Search in title, content, excerpt"
// @Success 200 {object} response.Response{data=[]domain.Post}
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /posts [get]
func (h *PostHandler) ListPosts(c *gin.Context) {
	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	pag := &pagination.OffsetPagination{
		Page:    page,
		PerPage: limit,
		Limit:   limit,
	}

	// Parse filters
	filter := repositories.PostFilter{
		Search: c.Query("search"),
	}

	if status := c.Query("status"); status != "" {
		filter.Status = domain.PostStatus(status)
	}

	if authorIDStr := c.Query("author_id"); authorIDStr != "" {
		if authorID, err := uuid.Parse(authorIDStr); err == nil {
			authorIDUint := uint(authorID.ID())
			filter.AuthorID = &authorIDUint
		}
	}

	posts, total, err := h.postUseCase.ListPosts(c.Request.Context(), filter, pag)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithPagination(c, posts, total, pag)
}

// CreatePost godoc
// @Summary Create post
// @Description Create a new post
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body post.CreatePostRequest true "Create post request"
// @Success 201 {object} response.Response{data=domain.Post}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /posts [post]
func (h *PostHandler) CreatePost(c *gin.Context) {
	var req post.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	// Set author ID from authenticated user
	userID := middleware.MustGetUserID(c)
	req.AuthorID = userID

	result, err := h.postUseCase.CreatePost(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, result)
}

// GetPost godoc
// @Summary Get post
// @Description Get post by ID
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Success 200 {object} response.Response{data=domain.Post}
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /posts/{id} [get]
func (h *PostHandler) GetPost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ValidationError(c, "Invalid post ID")
		return
	}

	post, err := h.postUseCase.GetPost(c.Request.Context(), uint(id))
	if err != nil {
		response.Error(c, err)
		return
	}

	// Increment view count
	_ = h.postUseCase.IncrementViewCount(c.Request.Context(), uint(id))

	response.Success(c, post)
}

// UpdatePost godoc
// @Summary Update post
// @Description Update an existing post
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Param request body post.UpdatePostRequest true "Update post request"
// @Success 200 {object} response.Response{data=domain.Post}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /posts/{id} [put]
func (h *PostHandler) UpdatePost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ValidationError(c, "Invalid post ID")
		return
	}

	var req post.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	result, err := h.postUseCase.UpdatePost(c.Request.Context(), uint(id), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// DeletePost godoc
// @Summary Delete post
// @Description Delete a post
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /posts/{id} [delete]
func (h *PostHandler) DeletePost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ValidationError(c, "Invalid post ID")
		return
	}

	if err := h.postUseCase.DeletePost(c.Request.Context(), uint(id)); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"message": "Post deleted successfully",
	})
}

// PublishPost godoc
// @Summary Publish post
// @Description Publish a post immediately
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /posts/{id}/publish [post]
func (h *PostHandler) PublishPost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ValidationError(c, "Invalid post ID")
		return
	}

	if err := h.postUseCase.PublishPost(c.Request.Context(), uint(id)); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"message": "Post published successfully",
	})
}

// SchedulePost godoc
// @Summary Schedule post
// @Description Schedule a post for future publication
// @Tags posts
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Post ID"
// @Param request body object{scheduled_at=string} true "Schedule request (RFC3339 format)"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /posts/{id}/schedule [post]
func (h *PostHandler) SchedulePost(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ValidationError(c, "Invalid post ID")
		return
	}

	var req struct {
		ScheduledAt string `json:"scheduled_at" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		response.ValidationError(c, "Invalid scheduled_at format (use RFC3339)")
		return
	}

	if err := h.postUseCase.SchedulePost(c.Request.Context(), uint(id), scheduledAt); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"message": "Post scheduled successfully",
	})
}
