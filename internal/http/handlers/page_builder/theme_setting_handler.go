package page_builder

import (
	"github.com/gin-gonic/gin"
	"github.com/owner/go-cms/internal/core/domain"
	usecases "github.com/owner/go-cms/internal/core/usecases/page_builder"
	"github.com/owner/go-cms/pkg/response"
)

type ThemeSettingHandler struct {
	themeUseCase *usecases.ThemeSettingUseCase
}

func NewThemeSettingHandler(themeUseCase *usecases.ThemeSettingUseCase) *ThemeSettingHandler {
	return &ThemeSettingHandler{
		themeUseCase: themeUseCase,
	}
}

// GetAllThemes gets all theme settings
// @Summary Get all theme settings
// @Tags theme-settings
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]ThemeSettingResponse}
// @Router /theme-settings [get]
func (h *ThemeSettingHandler) GetAllThemes(c *gin.Context) {
	themes, err := h.themeUseCase.GetAllThemes(c.Request.Context())
	if err != nil {
		response.Error(c, WrapError(err, "Failed to get themes"))
		return
	}

	themeResponses := make([]*ThemeSettingResponse, len(themes))
	for i, t := range themes {
		themeResponses[i] = ToThemeSettingResponse(t)
	}

	response.Success(c, map[string]interface{}{
		"data": themeResponses,
	})
}

// GetActiveTheme gets the active theme
// @Summary Get active theme
// @Tags theme-settings
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=ThemeSettingResponse}
// @Router /theme-settings/active [get]
func (h *ThemeSettingHandler) GetActiveTheme(c *gin.Context) {
	theme, err := h.themeUseCase.GetActiveTheme(c.Request.Context())
	if err != nil {
		response.Error(c, WrapError(err, "No active theme found"))
		return
	}

	response.Success(c, ToThemeSettingResponse(theme))
}

// UpdateTheme updates a theme setting
// @Summary Update a theme setting
// @Tags theme-settings
// @Accept json
// @Produce json
// @Param id path string true "Theme ID"
// @Param request body UpdateThemeRequest true "Update theme request"
// @Success 200 {object} response.Response{data=ThemeSettingResponse}
// @Router /theme-settings/{id} [put]
func (h *ThemeSettingHandler) UpdateTheme(c *gin.Context) {
	id, err := ParseUUID(c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}

	var req UpdateThemeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body")
		return
	}

	theme := &domain.ThemeSetting{
		UUIDModel: domain.UUIDModel{ID: id},
		Name:      req.Name,
		Config:    req.Config,
	}

	if err := h.themeUseCase.UpdateTheme(c.Request.Context(), theme); err != nil {
		response.Error(c, WrapError(err, "Failed to update theme"))
		return
	}

	response.Success(c, ToThemeSettingResponse(theme))
}

// ActivateTheme activates a theme
// @Summary Activate a theme
// @Tags theme-settings
// @Accept json
// @Produce json
// @Param id path string true "Theme ID"
// @Success 200 {object} response.Response
// @Router /theme-settings/{id}/activate [post]
func (h *ThemeSettingHandler) ActivateTheme(c *gin.Context) {
	id, err := ParseUUID(c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}

	if err := h.themeUseCase.ActivateTheme(c.Request.Context(), id); err != nil {
		response.Error(c, WrapError(err, "Failed to activate theme"))
		return
	}

	response.Success(c, gin.H{"message": "Theme activated successfully"})
}
