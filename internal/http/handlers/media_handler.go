package handlers

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/internal/core/usecases/media"
	"github.com/owner/go-cms/internal/http/middleware"
	"github.com/owner/go-cms/pkg/pagination"
	"github.com/owner/go-cms/pkg/response"
)

// MediaHandler handles media-related HTTP requests
type MediaHandler struct {
	mediaUseCase media.UseCase
}

// NewMediaHandler creates a new media handler
func NewMediaHandler(mediaUseCase media.UseCase) *MediaHandler {
	return &MediaHandler{
		mediaUseCase: mediaUseCase,
	}
}

// ListMedia godoc
// @Summary List media
// @Description Get list of media files with pagination and filters
// @Tags media
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param type query string false "Media type (image, video, document, audio, other)"
// @Param search query string false "Search in filename"
// @Success 200 {object} response.Response{data=[]domain.Media}
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /media [get]
func (h *MediaHandler) ListMedia(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	pag := &pagination.OffsetPagination{
		Page:    page,
		PerPage: limit,
		Limit:   limit,
	}

	filter := repositories.MediaFilter{
		Search: c.Query("search"),
	}

	if mediaType := c.Query("type"); mediaType != "" {
		filter.Type = domain.MediaType(mediaType)
	}

	media, total, err := h.mediaUseCase.ListMedia(c.Request.Context(), filter, pag)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithPagination(c, media, total, pag)
}

// UploadMedia godoc
// @Summary Upload media
// @Description Upload a media file
// @Tags media
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Media file"
// @Success 201 {object} response.Response{data=domain.Media}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /media/upload [post]
func (h *MediaHandler) UploadMedia(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.ValidationError(c, "File is required")
		return
	}

	uploaderID := middleware.MustGetUserID(c)

	result, err := h.mediaUseCase.UploadMedia(c.Request.Context(), file, uploaderID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, result)
}

// GetMedia godoc
// @Summary Get media
// @Description Get media by ID
// @Tags media
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Media ID"
// @Success 200 {object} response.Response{data=domain.Media}
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /media/{id} [get]
func (h *MediaHandler) GetMedia(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ValidationError(c, "Invalid media ID")
		return
	}

	media, err := h.mediaUseCase.GetMedia(c.Request.Context(), uint(id))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, media)
}

// DeleteMedia godoc
// @Summary Delete media
// @Description Delete a media file
// @Tags media
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Media ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /media/{id} [delete]
func (h *MediaHandler) DeleteMedia(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ValidationError(c, "Invalid media ID")
		return
	}

	if err := h.mediaUseCase.DeleteMedia(c.Request.Context(), uint(id)); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"message": "Media deleted successfully",
	})
}

// GetPresignedURL godoc
// @Summary Get presigned URL
// @Description Get a presigned URL for media access
// @Tags media
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Media ID"
// @Param expiry query int false "Expiry in seconds" default(3600)
// @Success 200 {object} response.Response{data=object{url=string}}
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /media/{id}/presigned-url [get]
func (h *MediaHandler) GetPresignedURL(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ValidationError(c, "Invalid media ID")
		return
	}

	expiry, _ := strconv.Atoi(c.DefaultQuery("expiry", "3600"))

	url, err := h.mediaUseCase.GetPresignedURL(c.Request.Context(), uint(id), time.Duration(expiry)*time.Second)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"url": url,
	})
}
