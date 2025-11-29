package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/internal/core/usecases/authorization"
	"github.com/owner/go-cms/pkg/response"
)

// RoleHandler handles HTTP requests for roles
type RoleHandler struct {
	useCase *authorization.RoleUseCase
}

// NewRoleHandler creates a new role handler
func NewRoleHandler(useCase *authorization.RoleUseCase) *RoleHandler {
	return &RoleHandler{useCase: useCase}
}

// CreateRole godoc
// @Summary Create a new role
// @Description Create a new role in the system
// @Tags roles
// @Accept json
// @Produce json
// @Param role body domain.Role true "Role object"
// @Success 201 {object} response.Response{data=domain.Role}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /roles [post]
func (h *RoleHandler) CreateRole(c *gin.Context) {
	var role domain.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if err := h.useCase.CreateRole(c.Request.Context(), &role); err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, role)
}

// GetRole godoc
// @Summary Get role by ID
// @Description Get detailed information about a role
// @Tags roles
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Success 200 {object} response.Response{data=domain.Role}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /roles/{id} [get]
func (h *RoleHandler) GetRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid role ID")
		return
	}

	role, err := h.useCase.GetRole(c.Request.Context(), uint(id))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, role)
}

// ListRoles godoc
// @Summary List all roles
// @Description Get a list of all roles with optional filters
// @Tags roles
// @Accept json
// @Produce json
// @Param name query string false "Filter by name"
// @Param level query string false "Filter by level"
// @Param parent_id query int false "Filter by parent ID"
// @Param is_system query bool false "Filter by system flag"
// @Success 200 {object} response.Response{data=[]domain.Role}
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /roles [get]
func (h *RoleHandler) ListRoles(c *gin.Context) {
	filter := repositories.RoleFilter{
		Name:  c.Query("name"),
		Level: domain.RoleLevel(c.Query("level")),
	}

	if parentID := c.Query("parent_id"); parentID != "" {
		id, err := strconv.ParseUint(parentID, 10, 32)
		if err == nil {
			uid := uint(id)
			filter.ParentID = &uid
		}
	}

	if isSystem := c.Query("is_system"); isSystem != "" {
		val := isSystem == "true"
		filter.IsSystem = &val
	}

	roles, err := h.useCase.ListRoles(c.Request.Context(), filter)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, roles)
}

// GetRoleHierarchy godoc
// @Summary Get role hierarchy
// @Description Get the complete role hierarchy tree
// @Tags roles
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]domain.Role}
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /roles/hierarchy [get]
func (h *RoleHandler) GetRoleHierarchy(c *gin.Context) {
	roles, err := h.useCase.GetRoleHierarchy(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, roles)
}

// UpdateRole godoc
// @Summary Update a role
// @Description Update an existing role
// @Tags roles
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Param role body domain.Role true "Role update object"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /roles/{id} [put]
func (h *RoleHandler) UpdateRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid role ID")
		return
	}

	var updates domain.Role
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if err := h.useCase.UpdateRole(c.Request.Context(), uint(id), &updates); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Role updated successfully"})
}

// DeleteRole godoc
// @Summary Delete a role
// @Description Delete a role from the system
// @Tags roles
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Success 200 {object} response.Response
// @Failure 403 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /roles/{id} [delete]
func (h *RoleHandler) DeleteRole(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid role ID")
		return
	}

	if err := h.useCase.DeleteRole(c.Request.Context(), uint(id)); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Role deleted successfully"})
}

// GetRolePermissions godoc
// @Summary Get role permissions
// @Description Get all permissions assigned to a role
// @Tags roles
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Success 200 {object} response.Response{data=[]domain.Permission}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /roles/{id}/permissions [get]
func (h *RoleHandler) GetRolePermissions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid role ID")
		return
	}

	permissions, err := h.useCase.GetRolePermissions(c.Request.Context(), uint(id))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, permissions)
}

// AssignPermission godoc
// @Summary Assign permission to role
// @Description Assign a permission to a role
// @Tags roles
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Param request body object{permission_id=uint} true "Permission assignment"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /roles/{id}/permissions [post]
func (h *RoleHandler) AssignPermission(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid role ID")
		return
	}

	var req struct {
		PermissionID uint `json:"permission_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if err := h.useCase.AssignPermission(c.Request.Context(), uint(roleID), req.PermissionID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Permission assigned successfully"})
}

// RemovePermission godoc
// @Summary Remove permission from role
// @Description Remove a permission from a role
// @Tags roles
// @Accept json
// @Produce json
// @Param id path int true "Role ID"
// @Param permissionId path int true "Permission ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /roles/{id}/permissions/{permissionId} [delete]
func (h *RoleHandler) RemovePermission(c *gin.Context) {
	roleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid role ID")
		return
	}

	permissionID, err := strconv.ParseUint(c.Param("permissionId"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid permission ID")
		return
	}

	if err := h.useCase.RemovePermission(c.Request.Context(), uint(roleID), uint(permissionID)); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Permission removed successfully"})
}
