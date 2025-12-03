package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/owner/go-cms/pkg/response"
)

// Add to existing MediaHandler

// UploadOptimized godoc
// @Summary Upload media with optimization
// @Description Upload a file with automatic compression, deduplication, and image optimization
// @Tags media
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "File to upload"
// @Success 200 {object} response.Response{data=domain.Media}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /media/upload-optimized [post]
func (h *MediaHandler) UploadOptimized(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, err)
		return
	}
	defer file.Close()

	media, err := h.mediaUseCase.UploadWithOptimization(c.Request.Context(), file, header)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, media)
}

// CleanupUnused godoc
// @Summary Cleanup unused files
// @Description Remove files not referenced in the last N days
// @Tags media
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param days query int false "Days threshold" default(30)
// @Success 200 {object} response.Response{data=map[string]int}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /media/cleanup [post]
func (h *MediaHandler) CleanupUnused(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))

	deleted, err := h.mediaUseCase.CleanupUnusedFiles(c.Request.Context(), days)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"deleted": deleted,
		"message": "Cleanup completed successfully",
	})
}
