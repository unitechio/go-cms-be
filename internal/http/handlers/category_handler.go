package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/owner/go-cms/internal/core/usecases"
	"github.com/owner/go-cms/pkg/pagination"
	"github.com/owner/go-cms/pkg/response"
)

// CategoryHandler handles category HTTP requests
type CategoryHandler struct {
	useCase *usecases.CategoryUseCase
}

// NewCategoryHandler creates a new category handler
func NewCategoryHandler(useCase *usecases.CategoryUseCase) *CategoryHandler {
	return &CategoryHandler{useCase: useCase}
}

// CreateCategory godoc
// @Summary Create a new category
// @Description Create a new category
// @Tags categories
// @Accept json
// @Produce json
// @Param category body usecases.CreateCategoryRequest true "Category data"
// @Success 201 {object} response.Response{data=domain.Category}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /categories [post]
// @Security BearerAuth
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var req usecases.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	category, err := h.useCase.CreateCategory(c.Request.Context(), &req)
	if err != nil {
		response.InternalError(c, "Failed to create category: "+err.Error())
		return
	}

	response.Created(c, category)
}

// GetCategory godoc
// @Summary Get a category by ID
// @Description Get a category by ID
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} response.Response{data=domain.Category}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /categories/{id} [get]
// @Security BearerAuth
func (h *CategoryHandler) GetCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid category ID: "+err.Error())
		return
	}

	category, err := h.useCase.GetCategory(c.Request.Context(), uint(id))
	if err != nil {
		response.NotFound(c, "Category not found: "+err.Error())
		return
	}

	response.Success(c, category)
}

// UpdateCategory godoc
// @Summary Update a category
// @Description Update a category
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param category body usecases.UpdateCategoryRequest true "Category data"
// @Success 200 {object} response.Response{data=domain.Category}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /categories/{id} [put]
// @Security BearerAuth
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid category ID: "+err.Error())
		return
	}

	var req usecases.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	category, err := h.useCase.UpdateCategory(c.Request.Context(), uint(id), &req)
	if err != nil {
		response.InternalError(c, "Failed to update category: "+err.Error())
		return
	}

	response.Success(c, category)
}

// DeleteCategory godoc
// @Summary Delete a category
// @Description Delete a category
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /categories/{id} [delete]
// @Security BearerAuth
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid category ID: "+err.Error())
		return
	}

	if err := h.useCase.DeleteCategory(c.Request.Context(), uint(id)); err != nil {
		response.InternalError(c, "Failed to delete category: "+err.Error())
		return
	}

	response.Success(c, gin.H{"message": "Category deleted successfully"})
}

// ListCategories godoc
// @Summary List categories with pagination
// @Description List categories with pagination and filters
// @Tags categories
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search term"
// @Param type query string false "Category type (blog, header, footer, sidebar)"
// @Param status query string false "Category status (active, inactive)"
// @Param parent_id query string false "Parent category ID (use 'null' for root categories)"
// @Success 200 {object} response.Response{data=[]domain.Category}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /categories [get]
// @Security BearerAuth
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	pag := &pagination.OffsetPagination{
		Page:    page,
		PerPage: limit,
		Limit:   limit,
	}

	// Build filters
	filters := make(map[string]interface{})

	if search := c.Query("search"); search != "" {
		filters["search"] = search
	}

	if categoryType := c.Query("type"); categoryType != "" {
		filters["type"] = categoryType
	}

	if status := c.Query("status"); status != "" {
		filters["status"] = status
	}

	if parentIDStr := c.Query("parent_id"); parentIDStr != "" {
		if parentIDStr == "null" {
			filters["parent_id"] = nil
		} else {
			parentID, err := strconv.ParseUint(parentIDStr, 10, 32)
			if err == nil {
				filters["parent_id"] = uint(parentID)
			}
		}
	}

	categories, total, err := h.useCase.ListCategories(c.Request.Context(), filters, pag)
	if err != nil {
		response.InternalError(c, "Failed to list categories: "+err.Error())
		return
	}

	response.SuccessWithPagination(c, categories, total, pag)
}

// GetCategoryTree godoc
// @Summary Get categories in tree structure
// @Description Get categories organized in a hierarchical tree structure
// @Tags categories
// @Accept json
// @Produce json
// @Param type query string false "Category type (blog, header, footer, sidebar)"
// @Success 200 {object} response.Response{data=[]domain.Category}
// @Failure 500 {object} response.Response
// @Router /categories/tree [get]
// @Security BearerAuth
func (h *CategoryHandler) GetCategoryTree(c *gin.Context) {
	categoryType := c.Query("type")

	categories, err := h.useCase.GetCategoryTree(c.Request.Context(), categoryType)
	if err != nil {
		response.InternalError(c, "Failed to get category tree: "+err.Error())
		return
	}

	response.Success(c, categories)
}

// GetActiveCategories godoc
// @Summary Get active categories
// @Description Get all active categories (no pagination)
// @Tags categories
// @Accept json
// @Produce json
// @Param type query string false "Category type (blog, header, footer, sidebar)"
// @Success 200 {object} response.Response{data=[]domain.Category}
// @Failure 500 {object} response.Response
// @Router /categories/active [get]
// @Security BearerAuth
func (h *CategoryHandler) GetActiveCategories(c *gin.Context) {
	categoryType := c.Query("type")

	categories, err := h.useCase.GetActiveCategories(c.Request.Context(), categoryType)
	if err != nil {
		response.InternalError(c, "Failed to get active categories: "+err.Error())
		return
	}

	response.Success(c, categories)
}

// ReorderCategory godoc
// @Summary Reorder a category
// @Description Update the order and parent of a category
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param reorder body usecases.ReorderCategoryRequest true "Reorder data"
// @Success 200 {object} response.Response{data=domain.Category}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /categories/{id}/reorder [put]
// @Security BearerAuth
func (h *CategoryHandler) ReorderCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid category ID: "+err.Error())
		return
	}

	var req usecases.ReorderCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	category, err := h.useCase.ReorderCategory(c.Request.Context(), uint(id), &req)
	if err != nil {
		response.InternalError(c, "Failed to reorder category: "+err.Error())
		return
	}

	response.Success(c, category)
}
