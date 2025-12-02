package page_builder

import (
	"github.com/gin-gonic/gin"
	usecases "github.com/owner/go-cms/internal/core/usecases/page_builder"
	"github.com/owner/go-cms/pkg/response"
)

type PageVersionHandler struct {
	versionUseCase *usecases.PageVersionUseCase
}

func NewPageVersionHandler(versionUseCase *usecases.PageVersionUseCase) *PageVersionHandler {
	return &PageVersionHandler{
		versionUseCase: versionUseCase,
	}
}

// GetVersionHistory gets version history for a page
// @Summary Get version history for a page
// @Tags page-versions
// @Accept json
// @Produce json
// @Param pageId path string true "Page ID"
// @Success 200 {object} response.Response{data=[]PageVersionResponse}
// @Router /pages/{pageId}/versions [get]
func (h *PageVersionHandler) GetVersionHistory(c *gin.Context) {
	pageID, err := ParseUUID(c.Param("pageId"))
	if err != nil {
		response.Error(c, err)
		return
	}

	versions, err := h.versionUseCase.GetVersionHistory(c.Request.Context(), pageID)
	if err != nil {
		response.Error(c, WrapError(err, "Failed to get version history"))
		return
	}

	versionResponses := make([]*PageVersionResponse, len(versions))
	for i, v := range versions {
		versionResponses[i] = ToPageVersionResponse(v)
	}

	response.Success(c, map[string]interface{}{
		"data": versionResponses,
	})
}

// RevertToVersion reverts a page to a specific version
// @Summary Revert a page to a specific version
// @Tags page-versions
// @Accept json
// @Produce json
// @Param pageId path string true "Page ID"
// @Param versionId path string true "Version ID"
// @Success 200 {object} response.Response
// @Router /pages/{pageId}/versions/{versionId}/revert [post]
func (h *PageVersionHandler) RevertToVersion(c *gin.Context) {
	pageID, err := ParseUUID(c.Param("pageId"))
	if err != nil {
		response.Error(c, err)
		return
	}

	versionID, err := ParseUUID(c.Param("versionId"))
	if err != nil {
		response.Error(c, err)
		return
	}

	if err := h.versionUseCase.RevertToVersion(c.Request.Context(), pageID, versionID); err != nil {
		response.Error(c, WrapError(err, "Failed to revert to version"))
		return
	}

	response.Success(c, gin.H{"message": "Page reverted to version successfully"})
}
