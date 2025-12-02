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

type BlockHandler struct {
	blockUseCase *usecases.BlockUseCase
}

func NewBlockHandler(blockUseCase *usecases.BlockUseCase) *BlockHandler {
	return &BlockHandler{
		blockUseCase: blockUseCase,
	}
}

// CreateBlock creates a new block
// @Summary Create a new block
// @Tags blocks
// @Accept json
// @Produce json
// @Param request body CreateBlockRequest true "Block creation request"
// @Success 201 {object} response.Response{data=BlockResponse}
// @Router /blocks [post]
func (h *BlockHandler) CreateBlock(c *gin.Context) {
	var req CreateBlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	block := &domain.Block{
		Name:            req.Name,
		Type:            req.Type,
		Category:        req.Category,
		Schema:          req.Schema,
		PreviewTemplate: req.PreviewTemplate,
		IsGlobal:        req.IsGlobal,
	}

	if err := h.blockUseCase.CreateBlock(c.Request.Context(), block); err != nil {
		response.InternalError(c, "Failed to create block")
		return
	}

	response.Created(c, ToBlockResponse(block))
}

// GetBlock gets a block by ID
// @Summary Get a block by ID
// @Tags blocks
// @Accept json
// @Produce json
// @Param id path string true "Block ID"
// @Success 200 {object} response.Response{data=BlockResponse}
// @Router /blocks/{id} [get]
func (h *BlockHandler) GetBlock(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid block ID")
		return
	}

	block, err := h.blockUseCase.GetBlock(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "Block not found")
		return
	}

	response.Success(c, ToBlockResponse(block))
}

// UpdateBlock updates a block
// @Summary Update a block
// @Tags blocks
// @Accept json
// @Produce json
// @Param id path string true "Block ID"
// @Param request body UpdateBlockRequest true "Block update request"
// @Success 200 {object} response.Response{data=BlockResponse}
// @Router /blocks/{id} [put]
func (h *BlockHandler) UpdateBlock(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid block ID")
		return
	}

	var req UpdateBlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	block := &domain.Block{
		UUIDModel:       domain.UUIDModel{ID: id},
		Name:            req.Name,
		Type:            req.Type,
		Category:        req.Category,
		Schema:          req.Schema,
		PreviewTemplate: req.PreviewTemplate,
		IsGlobal:        req.IsGlobal,
	}

	if err := h.blockUseCase.UpdateBlock(c.Request.Context(), block); err != nil {
		response.InternalError(c, "Failed to update block")
		return
	}

	updatedBlock, _ := h.blockUseCase.GetBlock(c.Request.Context(), id)
	response.Success(c, ToBlockResponse(updatedBlock))
}

// DeleteBlock deletes a block
// @Summary Delete a block
// @Tags blocks
// @Accept json
// @Produce json
// @Param id path string true "Block ID"
// @Success 200 {object} response.Response
// @Router /blocks/{id} [delete]
func (h *BlockHandler) DeleteBlock(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid block ID")
		return
	}

	if err := h.blockUseCase.DeleteBlock(c.Request.Context(), id); err != nil {
		response.InternalError(c, "Failed to delete block")
		return
	}

	response.Success(c, nil)
}

// ListBlocks lists blocks
// @Summary List blocks
// @Tags blocks
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Page limit"
// @Param search query string false "Search query"
// @Param category query string false "Category filter"
// @Param type query string false "Type filter"
// @Success 200 {object} response.Response{data=[]BlockResponse}
// @Router /blocks [get]
func (h *BlockHandler) ListBlocks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	search := c.Query("search")
	category := c.Query("category")
	blockType := c.Query("type")

	filter := repositories.BlockFilter{
		Search:   search,
		Category: category,
		Type:     blockType,
	}

	pag := &pagination.OffsetPagination{
		Page:    page,
		Limit:   limit,
		PerPage: limit,
		Total:   0,
	}
	blocks, total, err := h.blockUseCase.ListBlocks(c.Request.Context(), filter, pag)
	if err != nil {
		response.InternalError(c, "Failed to list blocks")
		return
	}

	blockResponses := make([]*BlockResponse, len(blocks))
	for i, b := range blocks {
		blockResponses[i] = ToBlockResponse(b)
	}

	response.SuccessWithPagination(c, blockResponses, total, &pagination.OffsetPagination{
		Page:    page,
		PerPage: limit,
		Total:   total,
	})
}
