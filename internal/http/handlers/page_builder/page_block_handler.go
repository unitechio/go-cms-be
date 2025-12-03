package page_builder

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	usecases "github.com/owner/go-cms/internal/core/usecases/page_builder"
	"github.com/owner/go-cms/pkg/response"
)

type PageBlockHandler struct {
	pageBlockUseCase *usecases.PageBlockUseCase
}

func NewPageBlockHandler(pageBlockUseCase *usecases.PageBlockUseCase) *PageBlockHandler {
	return &PageBlockHandler{
		pageBlockUseCase: pageBlockUseCase,
	}
}

// AddBlockToPage adds a block to a page
// @Summary Add a block to a page
// @Tags page-blocks
// @Accept json
// @Produce json
// @Param pageId path string true "Page ID"
// @Param request body AddPageBlockRequest true "Add block request"
// @Success 201 {object} response.Response{data=PageBlockResponse}
// @Router /pages/{pageId}/blocks [post]
func (h *PageBlockHandler) AddBlockToPage(c *gin.Context) {
	pageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid page ID")
		return
	}

	var req AddPageBlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	pageBlock := &domain.PageBlock{
		PageID:        pageID,
		BlockID:       req.BlockID,
		ParentBlockID: req.ParentBlockID,
		Order:         req.Order,
		Config:        req.Config,
		Language:      req.Language,
	}

	if pageBlock.Language == "" {
		pageBlock.Language = "en"
	}

	if err := h.pageBlockUseCase.AddBlockToPage(c.Request.Context(), pageBlock); err != nil {
		response.InternalError(c, "Failed to add block to page")
		return
	}

	response.Created(c, ToPageBlockResponse(pageBlock))
}

// UpdatePageBlock updates a page block configuration
// @Summary Update a page block
// @Tags page-blocks
// @Accept json
// @Produce json
// @Param pageId path string true "Page ID"
// @Param blockId path string true "Page Block ID"
// @Param request body UpdatePageBlockRequest true "Update block request"
// @Success 200 {object} response.Response{data=PageBlockResponse}
// @Router /pages/{pageId}/blocks/{blockId} [put]
func (h *PageBlockHandler) UpdatePageBlock(c *gin.Context) {
	pageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid page ID")
		return
	}

	blockID, err := uuid.Parse(c.Param("blockId"))
	if err != nil {
		response.BadRequest(c, "Invalid block ID")
		return
	}

	var req UpdatePageBlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	pageBlock := &domain.PageBlock{
		ID:     blockID,
		PageID: pageID,
		Config: req.Config,
		Order:  req.Order,
	}

	if err := h.pageBlockUseCase.UpdatePageBlock(c.Request.Context(), pageBlock); err != nil {
		response.InternalError(c, "Failed to update page block")
		return
	}

	response.Success(c, ToPageBlockResponse(pageBlock))
}

// RemoveBlockFromPage removes a block from a page
// @Summary Remove a block from a page
// @Tags page-blocks
// @Accept json
// @Produce json
// @Param pageId path string true "Page ID"
// @Param blockId path string true "Page Block ID"
// @Success 200 {object} response.Response
// @Router /pages/{pageId}/blocks/{blockId} [delete]
func (h *PageBlockHandler) RemoveBlockFromPage(c *gin.Context) {
	blockID, err := uuid.Parse(c.Param("blockId"))
	if err != nil {
		response.BadRequest(c, "Invalid block ID")
		return
	}

	if err := h.pageBlockUseCase.RemoveBlockFromPage(c.Request.Context(), blockID); err != nil {
		response.InternalError(c, "Failed to remove block from page")
		return
	}

	response.Success(c, nil)
}

// ReorderBlocks reorders blocks on a page
// @Summary Reorder blocks on a page
// @Tags page-blocks
// @Accept json
// @Produce json
// @Param pageId path string true "Page ID"
// @Param request body ReorderBlocksRequest true "Reorder request"
// @Success 200 {object} response.Response
// @Router /pages/{pageId}/blocks/reorder [put]
func (h *PageBlockHandler) ReorderBlocks(c *gin.Context) {
	pageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid page ID")
		return
	}

	var req ReorderBlocksRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	blockOrders := make(map[uuid.UUID]int)
	for _, block := range req.Blocks {
		blockOrders[block.ID] = block.Order
	}

	if err := h.pageBlockUseCase.ReorderBlocks(c.Request.Context(), pageID, blockOrders); err != nil {
		response.InternalError(c, "Failed to reorder blocks")
		return
	}

	response.Success(c, nil)
}
