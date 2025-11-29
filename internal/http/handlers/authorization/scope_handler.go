package authorization

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/owner/go-cms/internal/core/usecases/authorization"
	"github.com/owner/go-cms/pkg/pagination"
	"github.com/owner/go-cms/pkg/response"
)

// ScopeHandler handles scope-related HTTP requests
type ScopeHandler struct {
	scopeUseCase authorization.ScopeUseCase
}

// NewScopeHandler creates a new scope handler
func NewScopeHandler(scopeUseCase authorization.ScopeUseCase) *ScopeHandler {
	return &ScopeHandler{
		scopeUseCase: scopeUseCase,
	}
}

// CreateScope godoc
// @Summary Create a new scope
// @Description Create a new permission scope
// @Tags scopes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body authorization.CreateScopeRequest true "Create scope request"
// @Success 201 {object} response.Response{data=domain.Scope}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /scopes [post]
func (h *ScopeHandler) CreateScope(c *gin.Context) {
	var req authorization.CreateScopeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	scope, err := h.scopeUseCase.CreateScope(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, scope)
}

// GetScope godoc
// @Summary Get scope by ID
// @Description Get scope details by ID
// @Tags scopes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Scope ID"
// @Success 200 {object} response.Response{data=domain.Scope}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /scopes/{id} [get]
func (h *ScopeHandler) GetScope(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ValidationError(c, "Invalid scope ID")
		return
	}

	scope, err := h.scopeUseCase.GetScope(c.Request.Context(), uint(id))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, scope)
}

// GetScopeByCode godoc
// @Summary Get scope by code
// @Description Get scope details by code
// @Tags scopes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code path string true "Scope Code"
// @Success 200 {object} response.Response{data=domain.Scope}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /scopes/code/{code} [get]
func (h *ScopeHandler) GetScopeByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.ValidationError(c, "Scope code is required")
		return
	}

	scope, err := h.scopeUseCase.GetScopeByCode(c.Request.Context(), code)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, scope)
}

// ListScopes godoc
// @Summary List scopes
// @Description List all scopes with pagination
// @Tags scopes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} response.Response{data=pagination.Result[domain.Scope]}
// @Failure 500 {object} response.Response
// @Router /scopes [get]
func (h *ScopeHandler) ListScopes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	params := pagination.Params{
		Page:  page,
		Limit: limit,
	}

	result, err := h.scopeUseCase.ListScopes(c.Request.Context(), params)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// UpdateScope godoc
// @Summary Update scope
// @Description Update an existing scope
// @Tags scopes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Scope ID"
// @Param request body authorization.UpdateScopeRequest true "Update scope request"
// @Success 200 {object} response.Response{data=domain.Scope}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /scopes/{id} [put]
func (h *ScopeHandler) UpdateScope(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ValidationError(c, "Invalid scope ID")
		return
	}

	var req authorization.UpdateScopeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	scope, err := h.scopeUseCase.UpdateScope(c.Request.Context(), uint(id), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, scope)
}

// DeleteScope godoc
// @Summary Delete scope
// @Description Delete a scope
// @Tags scopes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Scope ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /scopes/{id} [delete]
func (h *ScopeHandler) DeleteScope(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ValidationError(c, "Invalid scope ID")
		return
	}

	if err := h.scopeUseCase.DeleteScope(c.Request.Context(), uint(id)); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"message": "Scope deleted successfully",
	})
}

// ListAllScopes godoc
// @Summary List all scopes
// @Description List all scopes without pagination
// @Tags scopes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]domain.Scope}
// @Failure 500 {object} response.Response
// @Router /scopes/all [get]
func (h *ScopeHandler) ListAllScopes(c *gin.Context) {
	scopes, err := h.scopeUseCase.ListAllScopes(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, scopes)
}
