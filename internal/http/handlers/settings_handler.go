package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/usecases/settings"
	"github.com/owner/go-cms/pkg/response"
)

type SettingsHandler struct {
	settingsUseCase settings.UseCase
}

func NewSettingsHandler(settingsUseCase settings.UseCase) *SettingsHandler {
	return &SettingsHandler{
		settingsUseCase: settingsUseCase,
	}
}

// GetAllSettings godoc
// @Summary Get all user settings
// @Description Get all settings for the authenticated user
// @Tags settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path string true "User ID"
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /users/{user_id}/settings [get]
func (h *SettingsHandler) GetAllSettings(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		response.Error(c, err)
		return
	}

	settings, err := h.settingsUseCase.GetAllSettings(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, settings)
}

// GetSetting godoc
// @Summary Get a specific setting
// @Description Get a specific setting by key for the authenticated user
// @Tags settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path string true "User ID"
// @Param key path string true "Setting Key"
// @Success 200 {object} response.Response{data=domain.UserSetting}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /users/{user_id}/settings/{key} [get]
func (h *SettingsHandler) GetSetting(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		response.Error(c, err)
		return
	}

	key := c.Param("key")
	setting, err := h.settingsUseCase.GetSetting(c.Request.Context(), userID, key)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, setting)
}

// UpdateSetting godoc
// @Summary Update a setting
// @Description Update a specific setting for the authenticated user
// @Tags settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path string true "User ID"
// @Param key path string true "Setting Key"
// @Param request body map[string]interface{} true "Setting Value"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /users/{user_id}/settings/{key} [put]
func (h *SettingsHandler) UpdateSetting(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		response.Error(c, err)
		return
	}

	key := c.Param("key")

	var value map[string]interface{}
	if err := c.ShouldBindJSON(&value); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.settingsUseCase.UpdateSetting(c.Request.Context(), userID, key, value); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Setting updated successfully"})
}

// DeleteSetting godoc
// @Summary Delete a setting
// @Description Delete a specific setting for the authenticated user
// @Tags settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path string true "User ID"
// @Param key path string true "Setting Key"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /users/{user_id}/settings/{key} [delete]
func (h *SettingsHandler) DeleteSetting(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		response.Error(c, err)
		return
	}

	key := c.Param("key")

	if err := h.settingsUseCase.DeleteSetting(c.Request.Context(), userID, key); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Setting deleted successfully"})
}

// BulkUpdateSettings godoc
// @Summary Bulk update settings
// @Description Update multiple settings at once for the authenticated user
// @Tags settings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path string true "User ID"
// @Param request body map[string]interface{} true "Settings Map"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /users/{user_id}/settings/bulk [post]
func (h *SettingsHandler) BulkUpdateSettings(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		response.Error(c, err)
		return
	}

	var settings map[string]interface{}
	if err := c.ShouldBindJSON(&settings); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.settingsUseCase.BulkUpdateSettings(c.Request.Context(), userID, settings); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Settings updated successfully"})
}

// GetDefaultSettings godoc
// @Summary Get default settings
// @Description Get the default settings configuration
// @Tags settings
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Router /settings/defaults [get]
func (h *SettingsHandler) GetDefaultSettings(c *gin.Context) {
	defaults := h.settingsUseCase.GetDefaultSettings(c.Request.Context())
	response.Success(c, defaults)
}
