package handlers

import (
	"github.com/gin-gonic/gin"

	"github.com/owner/go-cms/internal/core/usecases/translation"
	"github.com/owner/go-cms/pkg/response"
)

type TranslationHandler struct {
	translationUseCase translation.UseCase
}

func NewTranslationHandler(translationUseCase translation.UseCase) *TranslationHandler {
	return &TranslationHandler{
		translationUseCase: translationUseCase,
	}
}

// GetTranslations godoc
// @Summary Get all translations for a locale
// @Description Get all translations for a specific locale and namespace
// @Tags translations
// @Accept json
// @Produce json
// @Param locale path string true "Locale (en, vi, fr)"
// @Param namespace query string false "Namespace" default(common)
// @Success 200 {object} response.Response{data=map[string]string}
// @Failure 500 {object} response.Response
// @Router /translations/{locale} [get]
func (h *TranslationHandler) GetTranslations(c *gin.Context) {
	locale := c.Param("locale")
	namespace := c.DefaultQuery("namespace", "common")

	translations, err := h.translationUseCase.GetTranslations(c.Request.Context(), locale, namespace)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, translations)
}

// SetTranslation godoc
// @Summary Set a translation
// @Description Create or update a translation
// @Tags translations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{key:string,locale:string,namespace:string,value:string} true "Translation"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /translations [post]
func (h *TranslationHandler) SetTranslation(c *gin.Context) {
	var req struct {
		Key       string `json:"key" binding:"required"`
		Locale    string `json:"locale" binding:"required"`
		Namespace string `json:"namespace"`
		Value     string `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.translationUseCase.SetTranslation(c.Request.Context(), req.Key, req.Locale, req.Namespace, req.Value); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Translation set successfully"})
}

// BulkSetTranslations godoc
// @Summary Bulk set translations
// @Description Create or update multiple translations at once
// @Tags translations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{namespace:string,translations:map[string]map[string]string} true "Bulk Translations"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /translations/bulk [post]
func (h *TranslationHandler) BulkSetTranslations(c *gin.Context) {
	var req struct {
		Namespace    string                       `json:"namespace"`
		Translations map[string]map[string]string `json:"translations" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, err)
		return
	}

	if err := h.translationUseCase.BulkSetTranslations(c.Request.Context(), req.Translations, req.Namespace); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Translations set successfully"})
}

// GetSupportedLocales godoc
// @Summary Get supported locales
// @Description Get list of supported locales
// @Tags translations
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]string}
// @Router /translations/locales [get]
func (h *TranslationHandler) GetSupportedLocales(c *gin.Context) {
	locales := h.translationUseCase.GetSupportedLocales(c.Request.Context())
	response.Success(c, locales)
}

// GetNamespaces godoc
// @Summary Get namespaces
// @Description Get list of available namespaces
// @Tags translations
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]string}
// @Router /translations/namespaces [get]
func (h *TranslationHandler) GetNamespaces(c *gin.Context) {
	namespaces, err := h.translationUseCase.GetNamespaces(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, namespaces)
}
