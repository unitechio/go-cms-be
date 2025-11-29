package authorization

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/owner/go-cms/internal/core/usecases/authorization"
	"github.com/owner/go-cms/pkg/pagination"
	"github.com/owner/go-cms/pkg/response"
)

// ServiceHandler handles service-related HTTP requests
type ServiceHandler struct {
	serviceUseCase authorization.ServiceUseCase
}

// NewServiceHandler creates a new service handler
func NewServiceHandler(serviceUseCase authorization.ServiceUseCase) *ServiceHandler {
	return &ServiceHandler{
		serviceUseCase: serviceUseCase,
	}
}

// CreateService godoc
// @Summary Create a new service
// @Description Create a new service within a department
// @Tags services
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body authorization.CreateServiceRequest true "Create service request"
// @Success 201 {object} response.Response{data=domain.Service}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /services [post]
func (h *ServiceHandler) CreateService(c *gin.Context) {
	var req authorization.CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	service, err := h.serviceUseCase.CreateService(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, service)
}

// GetService godoc
// @Summary Get service by ID
// @Description Get service details by ID
// @Tags services
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Service ID"
// @Success 200 {object} response.Response{data=domain.Service}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /services/{id} [get]
func (h *ServiceHandler) GetService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ValidationError(c, "Invalid service ID")
		return
	}

	service, err := h.serviceUseCase.GetService(c.Request.Context(), uint(id))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, service)
}

// GetServiceByCode godoc
// @Summary Get service by code
// @Description Get service details by code
// @Tags services
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param code path string true "Service Code"
// @Success 200 {object} response.Response{data=domain.Service}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /services/code/{code} [get]
func (h *ServiceHandler) GetServiceByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.ValidationError(c, "Service code is required")
		return
	}

	service, err := h.serviceUseCase.GetServiceByCode(c.Request.Context(), code)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, service)
}

// ListServices godoc
// @Summary List services
// @Description List all services with pagination
// @Tags services
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} response.Response{data=pagination.Result[domain.Service]}
// @Failure 500 {object} response.Response
// @Router /services [get]
func (h *ServiceHandler) ListServices(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	params := pagination.Params{
		Page:  page,
		Limit: limit,
	}

	result, err := h.serviceUseCase.ListServices(c.Request.Context(), params)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// ListServicesByDepartment godoc
// @Summary List services by department
// @Description List all services belonging to a specific department
// @Tags services
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param deptId path int true "Department ID"
// @Success 200 {object} response.Response{data=[]domain.Service}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /departments/{id}/services [get]
func (h *ServiceHandler) ListServicesByDepartment(c *gin.Context) {
	deptID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ValidationError(c, "Invalid department ID")
		return
	}

	services, err := h.serviceUseCase.ListServicesByDepartment(c.Request.Context(), uint(deptID))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, services)
}

// UpdateService godoc
// @Summary Update service
// @Description Update an existing service
// @Tags services
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Service ID"
// @Param request body authorization.UpdateServiceRequest true "Update service request"
// @Success 200 {object} response.Response{data=domain.Service}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /services/{id} [put]
func (h *ServiceHandler) UpdateService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ValidationError(c, "Invalid service ID")
		return
	}

	var req authorization.UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	service, err := h.serviceUseCase.UpdateService(c.Request.Context(), uint(id), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, service)
}

// DeleteService godoc
// @Summary Delete service
// @Description Delete a service
// @Tags services
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Service ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /services/{id} [delete]
func (h *ServiceHandler) DeleteService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ValidationError(c, "Invalid service ID")
		return
	}

	if err := h.serviceUseCase.DeleteService(c.Request.Context(), uint(id)); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"message": "Service deleted successfully",
	})
}

// ListActiveServices godoc
// @Summary List active services
// @Description List all active services
// @Tags services
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]domain.Service}
// @Failure 500 {object} response.Response
// @Router /services/active [get]
func (h *ServiceHandler) ListActiveServices(c *gin.Context) {
	services, err := h.serviceUseCase.ListActiveServices(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, services)
}
