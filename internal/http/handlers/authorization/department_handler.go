package authorization

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/owner/go-cms/internal/core/usecases/authorization"
	"github.com/owner/go-cms/pkg/pagination"
	"github.com/owner/go-cms/pkg/response"
)

// DepartmentHandler handles department-related HTTP requests
type DepartmentHandler struct {
	departmentUseCase authorization.DepartmentUseCase
}

// NewDepartmentHandler creates a new department handler
func NewDepartmentHandler(departmentUseCase authorization.DepartmentUseCase) *DepartmentHandler {
	return &DepartmentHandler{
		departmentUseCase: departmentUseCase,
	}
}

// CreateDepartment godoc
// @Summary Create a new department
// @Description Create a new department within a module
// @Tags departments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body authorization.CreateDepartmentRequest true "Create department request"
// @Success 201 {object} response.Response{data=domain.Department}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /departments [post]
func (h *DepartmentHandler) CreateDepartment(c *gin.Context) {
	var req authorization.CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	department, err := h.departmentUseCase.CreateDepartment(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, department)
}

// GetDepartment godoc
// @Summary Get department by ID
// @Description Get department details by ID
// @Tags departments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Department ID"
// @Success 200 {object} response.Response{data=domain.Department}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /departments/{id} [get]
func (h *DepartmentHandler) GetDepartment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ValidationError(c, "Invalid department ID")
		return
	}

	department, err := h.departmentUseCase.GetDepartment(c.Request.Context(), uint(id))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, department)
}

// GetDepartmentByCode godoc
// @Summary Get department by code
// @Description Get department details by code
// @Tags departments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code path string true "Department Code"
// @Success 200 {object} response.Response{data=domain.Department}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /departments/code/{code} [get]
func (h *DepartmentHandler) GetDepartmentByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.ValidationError(c, "Department code is required")
		return
	}

	department, err := h.departmentUseCase.GetDepartmentByCode(c.Request.Context(), code)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, department)
}

// ListDepartments godoc
// @Summary List departments
// @Description List all departments with pagination
// @Tags departments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} response.Response{data=pagination.Result[domain.Department]}
// @Failure 500 {object} response.Response
// @Router /departments [get]
func (h *DepartmentHandler) ListDepartments(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	params := pagination.Params{
		Page:  page,
		Limit: limit,
	}

	result, err := h.departmentUseCase.ListDepartments(c.Request.Context(), params)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// ListDepartmentsByModule godoc
// @Summary List departments by module
// @Description List all departments belonging to a specific module
// @Tags departments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param moduleId path int true "Module ID"
// @Success 200 {object} response.Response{data=[]domain.Department}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /modules/{id}/departments [get]
func (h *DepartmentHandler) ListDepartmentsByModule(c *gin.Context) {
	moduleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ValidationError(c, "Invalid module ID")
		return
	}

	departments, err := h.departmentUseCase.ListDepartmentsByModule(c.Request.Context(), uint(moduleID))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, departments)
}

// UpdateDepartment godoc
// @Summary Update department
// @Description Update an existing department
// @Tags departments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Department ID"
// @Param request body authorization.UpdateDepartmentRequest true "Update department request"
// @Success 200 {object} response.Response{data=domain.Department}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /departments/{id} [put]
func (h *DepartmentHandler) UpdateDepartment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ValidationError(c, "Invalid department ID")
		return
	}

	var req authorization.UpdateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	department, err := h.departmentUseCase.UpdateDepartment(c.Request.Context(), uint(id), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, department)
}

// DeleteDepartment godoc
// @Summary Delete department
// @Description Delete a department
// @Tags departments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Department ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /departments/{id} [delete]
func (h *DepartmentHandler) DeleteDepartment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ValidationError(c, "Invalid department ID")
		return
	}

	if err := h.departmentUseCase.DeleteDepartment(c.Request.Context(), uint(id)); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"message": "Department deleted successfully",
	})
}

// ListActiveDepartments godoc
// @Summary List active departments
// @Description List all active departments
// @Tags departments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]domain.Department}
// @Failure 500 {object} response.Response
// @Router /departments/active [get]
func (h *DepartmentHandler) ListActiveDepartments(c *gin.Context) {
	departments, err := h.departmentUseCase.ListActiveDepartments(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, departments)
}
