package authorization

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/owner/go-cms/internal/core/usecases/authorization"
	"github.com/owner/go-cms/pkg/pagination"
	"github.com/owner/go-cms/pkg/response"
)

// ModuleHandler handles module-related HTTP requests
type ModuleHandler struct {
	moduleUseCase authorization.ModuleUseCase
}

// NewModuleHandler creates a new module handler
func NewModuleHandler(moduleUseCase authorization.ModuleUseCase) *ModuleHandler {
	return &ModuleHandler{
		moduleUseCase: moduleUseCase,
	}
}

// CreateModule godoc
// @Summary Create a new module
// @Description Create a new system module
// @Tags modules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body authorization.CreateModuleRequest true "Create module request"
// @Success 201 {object} response.Response{data=domain.Module}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /modules [post]
func (h *ModuleHandler) CreateModule(c *gin.Context) {
	var req authorization.CreateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	module, err := h.moduleUseCase.CreateModule(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, module)
}

// GetModule godoc
// @Summary Get module by ID
// @Description Get module details by ID
// @Tags modules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Module ID"
// @Success 200 {object} response.Response{data=domain.Module}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /modules/{id} [get]
func (h *ModuleHandler) GetModule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ValidationError(c, "Invalid module ID")
		return
	}

	module, err := h.moduleUseCase.GetModule(c.Request.Context(), uint(id))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, module)
}

// GetModuleByCode godoc
// @Summary Get module by code
// @Description Get module details by code
// @Tags modules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code path string true "Module Code"
// @Success 200 {object} response.Response{data=domain.Module}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /modules/code/{code} [get]
func (h *ModuleHandler) GetModuleByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.ValidationError(c, "Module code is required")
		return
	}

	module, err := h.moduleUseCase.GetModuleByCode(c.Request.Context(), code)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, module)
}

// ListModules godoc
// @Summary List modules
// @Description List all modules with pagination
// @Tags modules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} response.Response{data=pagination.Result[domain.Module]}
// @Failure 500 {object} response.Response
// @Router /modules [get]
func (h *ModuleHandler) ListModules(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	params := pagination.Params{
		Page:  page,
		Limit: limit,
	}

	result, err := h.moduleUseCase.ListModules(c.Request.Context(), params)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// UpdateModule godoc
// @Summary Update module
// @Description Update an existing module
// @Tags modules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Module ID"
// @Param request body authorization.UpdateModuleRequest true "Update module request"
// @Success 200 {object} response.Response{data=domain.Module}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /modules/{id} [put]
func (h *ModuleHandler) UpdateModule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ValidationError(c, "Invalid module ID")
		return
	}

	var req authorization.UpdateModuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	module, err := h.moduleUseCase.UpdateModule(c.Request.Context(), uint(id), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, module)
}

// DeleteModule godoc
// @Summary Delete module
// @Description Delete a module
// @Tags modules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Module ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /modules/{id} [delete]
func (h *ModuleHandler) DeleteModule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ValidationError(c, "Invalid module ID")
		return
	}

	if err := h.moduleUseCase.DeleteModule(c.Request.Context(), uint(id)); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"message": "Module deleted successfully",
	})
}

// ListActiveModules godoc
// @Summary List active modules
// @Description List all active modules
// @Tags modules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]domain.Module}
// @Failure 500 {object} response.Response
// @Router /modules/active [get]
func (h *ModuleHandler) ListActiveModules(c *gin.Context) {
	modules, err := h.moduleUseCase.ListActiveModules(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, modules)
}
