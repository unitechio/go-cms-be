package page_builder

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	usecases "github.com/owner/go-cms/internal/core/usecases/page_builder"
	"github.com/owner/go-cms/pkg/pagination"
	"github.com/owner/go-cms/pkg/response"
)

type PageHandler struct {
	pageUseCase *usecases.PageUseCase
}

func NewPageHandler(pageUseCase *usecases.PageUseCase) *PageHandler {
	return &PageHandler{
		pageUseCase: pageUseCase,
	}
}

// CreatePage creates a new page
// @Summary Create a new page
// @Tags pages
// @Accept json
// @Produce json
// @Param request body CreatePageRequest true "Page creation request"
// @Success 201 {object} response.Response{data=PageResponse}
// @Router /pages [post]
func (h *PageHandler) CreatePage(c *gin.Context) {
	var req CreatePageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	// Get current user ID from context (middleware should set this)
	// For now assuming it's set as "userID"
	userIDStr := c.GetString("userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		// Fallback or error if auth middleware not fully integrated yet
		// response.Error(c, http.StatusUnauthorized, "Unauthorized", nil)
		// return
		// For development without auth, generate a random UUID or use a fixed one
		userID = uuid.New()
	}

	page := &domain.Page{
		Title:          req.Title,
		Slug:           req.Slug,
		Template:       req.Template,
		Status:         domain.PageStatus(req.Status),
		SeoTitle:       req.SeoTitle,
		SeoDescription: req.SeoDescription,
		OgImage:        req.OgImage,
		AuthorID:       userID,
	}

	if err := h.pageUseCase.CreatePage(c.Request.Context(), page); err != nil {
		response.InternalError(c, "Failed to create page")
		return
	}

	response.Created(c, ToPageResponse(page))
}

// GetPage gets a page by ID
// @Summary Get a page by ID
// @Tags pages
// @Accept json
// @Produce json
// @Param id path string true "Page ID"
// @Success 200 {object} response.Response{data=PageResponse}
// @Router /pages/{id} [get]
func (h *PageHandler) GetPage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid page ID")
		return
	}

	page, err := h.pageUseCase.GetPage(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "Page not found")
		return
	}

	response.Success(c, ToPageResponse(page))
}

// UpdatePage updates a page
// @Summary Update a page
// @Tags pages
// @Accept json
// @Produce json
// @Param id path string true "Page ID"
// @Param request body UpdatePageRequest true "Page update request"
// @Success 200 {object} response.Response{data=PageResponse}
// @Router /pages/{id} [put]
func (h *PageHandler) UpdatePage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid page ID")
		return
	}

	var req UpdatePageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	page := &domain.Page{
		UUIDModel:      domain.UUIDModel{ID: id},
		Title:          req.Title,
		Slug:           req.Slug,
		Template:       req.Template,
		Status:         domain.PageStatus(req.Status),
		SeoTitle:       req.SeoTitle,
		SeoDescription: req.SeoDescription,
		OgImage:        req.OgImage,
	}

	if err := h.pageUseCase.UpdatePage(c.Request.Context(), page); err != nil {
		response.InternalError(c, "Failed to update page")
		return
	}

	// Fetch updated page to return full object
	updatedPage, _ := h.pageUseCase.GetPage(c.Request.Context(), id)
	response.Success(c, ToPageResponse(updatedPage))
}

// DeletePage deletes a page
// @Summary Delete a page
// @Tags pages
// @Accept json
// @Produce json
// @Param id path string true "Page ID"
// @Success 200 {object} response.Response
// @Router /pages/{id} [delete]
func (h *PageHandler) DeletePage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid page ID")
		return
	}

	if err := h.pageUseCase.DeletePage(c.Request.Context(), id); err != nil {
		response.InternalError(c, "Failed to delete page")
		return
	}

	response.Success(c, nil)
}

// ListPages lists pages
// @Summary List pages
// @Tags pages
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Page limit"
// @Param search query string false "Search query"
// @Param status query string false "Status filter"
// @Success 200 {object} response.Response{data=[]PageResponse}
// @Router /pages [get]
func (h *PageHandler) ListPages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	status := c.Query("status")

	filter := repositories.PageFilter{
		Search: search,
		Status: domain.PageStatus(status),
	}

	pag := &pagination.OffsetPagination{
		Page:    page,
		Limit:   limit,
		PerPage: limit,
		Total:   0,
	}
	pages, total, err := h.pageUseCase.ListPages(c.Request.Context(), filter, pag)
	if err != nil {
		response.InternalError(c, "Failed to list pages")
		return
	}

	pageResponses := make([]*PageResponse, len(pages))
	for i, p := range pages {
		pageResponses[i] = ToPageResponse(p)
	}

	response.SuccessWithPagination(c, pageResponses, total, &pagination.OffsetPagination{
		Page:    page,
		PerPage: limit,
		Total:   total,
	})
}

// DuplicatePage duplicates a page
// @Summary Duplicate a page
// @Tags pages
// @Accept json
// @Produce json
// @Param id path string true "Page ID"
// @Success 200 {object} response.Response{data=PageResponse}
// @Router /pages/{id}/duplicate [post]
func (h *PageHandler) DuplicatePage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid page ID")
		return
	}

	newPage, err := h.pageUseCase.DuplicatePage(c.Request.Context(), id)
	if err != nil {
		response.InternalError(c, "Failed to duplicate page")
		return
	}

	response.Created(c, ToPageResponse(newPage))
}

// PublishPage publishes a page
// @Summary Publish a page
// @Tags pages
// @Accept json
// @Produce json
// @Param id path string true "Page ID"
// @Success 200 {object} response.Response{data=PageResponse}
// @Router /pages/{id}/publish [post]
func (h *PageHandler) PublishPage(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid page ID")
		return
	}

	userIDStr := c.GetString("userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		userID = uuid.New() // Fallback
	}

	if err := h.pageUseCase.PublishPage(c.Request.Context(), id, userID); err != nil {
		response.InternalError(c, "Failed to publish page")
		return
	}

	updatedPage, _ := h.pageUseCase.GetPage(c.Request.Context(), id)
	response.Success(c, ToPageResponse(updatedPage))
}
