package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/internal/core/usecases/authorization"
	"github.com/owner/go-cms/pkg/response"
)

// PermissionHandler handles HTTP requests for permissions
type PermissionHandler struct {
	useCase *authorization.PermissionUseCase
}

// NewPermissionHandler creates a new permission handler
func NewPermissionHandler(useCase *authorization.PermissionUseCase) *PermissionHandler {
	return &PermissionHandler{useCase: useCase}
}

// CreatePermission godoc
// @Summary Create a new permission
// @Description Create a new permission in the system
// @Tags permissions
// @Accept json
// @Produce json
// @Param permission body domain.Permission true "Permission object"
// @Success 201 {object} response.Response{data=domain.Permission}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /permissions [post]
func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var permission domain.Permission
	if err := c.ShouldBindJSON(&permission); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if err := h.useCase.CreatePermission(c.Request.Context(), &permission); err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, permission)
}

// GetPermission godoc
// @Summary Get permission by ID
// @Description Get detailed information about a permission
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path int true "Permission ID"
// @Success 200 {object} response.Response{data=domain.Permission}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /permissions/{id} [get]
func (h *PermissionHandler) GetPermission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid permission ID")
		return
	}

	permission, err := h.useCase.GetPermission(c.Request.Context(), uint(id))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, permission)
}

// ListPermissions godoc
// @Summary List all permissions
// @Description Get a list of all permissions with optional filters
// @Tags permissions
// @Accept json
// @Produce json
// @Param module query string false "Filter by module"
// @Param department query string false "Filter by department"
// @Param service query string false "Filter by service"
// @Param resource query string false "Filter by resource"
// @Param action query string false "Filter by action"
// @Success 200 {object} response.Response{data=[]domain.Permission}
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /permissions [get]
func (h *PermissionHandler) ListPermissions(c *gin.Context) {
	filter := repositories.PermissionFilter{
		Module:     c.Query("module"),
		Department: c.Query("department"),
		Service:    c.Query("service"),
		Resource:   c.Query("resource"),
		Action:     c.Query("action"),
	}

	permissions, err := h.useCase.ListPermissions(c.Request.Context(), filter)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, permissions)
}

// GetPermissionsByModule godoc
// @Summary Get permissions by module
// @Description Get all permissions for a specific module
// @Tags permissions
// @Accept json
// @Produce json
// @Param module path string true "Module name"
// @Success 200 {object} response.Response{data=[]domain.Permission}
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /permissions/module/{module} [get]
func (h *PermissionHandler) GetPermissionsByModule(c *gin.Context) {
	module := c.Param("module")
	if module == "" {
		response.BadRequest(c, "Module is required")
		return
	}

	permissions, err := h.useCase.GetPermissionsByModule(c.Request.Context(), module)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, permissions)
}

// UpdatePermission godoc
// @Summary Update a permission
// @Description Update an existing permission
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path int true "Permission ID"
// @Param permission body domain.Permission true "Permission update object"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /permissions/{id} [put]
func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid permission ID")
		return
	}

	var updates domain.Permission
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if err := h.useCase.UpdatePermission(c.Request.Context(), uint(id), &updates); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Permission updated successfully"})
}

// DeletePermission godoc
// @Summary Delete a permission
// @Description Delete a permission from the system
// @Tags permissions
// @Accept json
// @Produce json
// @Param id path int true "Permission ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /permissions/{id} [delete]
func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid permission ID")
		return
	}

	if err := h.useCase.DeletePermission(c.Request.Context(), uint(id)); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Permission deleted successfully"})
}
