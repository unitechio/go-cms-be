package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/internal/core/usecases/customer"
	"github.com/owner/go-cms/pkg/pagination"
	"github.com/owner/go-cms/pkg/response"
)

// CustomerHandler handles customer-related HTTP requests
type CustomerHandler struct {
	customerUseCase customer.UseCase
}

// NewCustomerHandler creates a new customer handler
func NewCustomerHandler(customerUseCase customer.UseCase) *CustomerHandler {
	return &CustomerHandler{
		customerUseCase: customerUseCase,
	}
}

// ListCustomers godoc
// @Summary List customers
// @Description Get list of customers with pagination and filters
// @Tags customers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param status query string false "Customer status"
// @Param search query string false "Search in name, email, company"
// @Success 200 {object} response.Response{data=[]domain.Customer}
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /customers [get]
func (h *CustomerHandler) ListCustomers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	pag := &pagination.OffsetPagination{
		Page:    page,
		PerPage: limit,
		Limit:   limit,
	}

	filter := repositories.CustomerFilter{
		Search: c.Query("search"),
		Source: c.Query("source"),
	}

	if status := c.Query("status"); status != "" {
		filter.Status = domain.UserStatus(status)
	}

	customers, total, err := h.customerUseCase.ListCustomers(c.Request.Context(), filter, pag)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithPagination(c, customers, total, pag)
}

// CreateCustomer godoc
// @Summary Create customer
// @Description Create a new customer
// @Tags customers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body customer.CreateCustomerRequest true "Create customer request"
// @Success 201 {object} response.Response{data=domain.Customer}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /customers [post]
func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	var req customer.CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	result, err := h.customerUseCase.CreateCustomer(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, result)
}

// GetCustomer godoc
// @Summary Get customer
// @Description Get customer by ID
// @Tags customers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Customer ID"
// @Success 200 {object} response.Response{data=domain.Customer}
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /customers/{id} [get]
func (h *CustomerHandler) GetCustomer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ValidationError(c, "Invalid customer ID")
		return
	}

	customer, err := h.customerUseCase.GetCustomer(c.Request.Context(), uint(id))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, customer)
}

// UpdateCustomer godoc
// @Summary Update customer
// @Description Update an existing customer
// @Tags customers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Customer ID"
// @Param request body customer.UpdateCustomerRequest true "Update customer request"
// @Success 200 {object} response.Response{data=domain.Customer}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /customers/{id} [put]
func (h *CustomerHandler) UpdateCustomer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ValidationError(c, "Invalid customer ID")
		return
	}

	var req customer.UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	result, err := h.customerUseCase.UpdateCustomer(c.Request.Context(), uint(id), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// DeleteCustomer godoc
// @Summary Delete customer
// @Description Delete a customer
// @Tags customers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Customer ID"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /customers/{id} [delete]
func (h *CustomerHandler) DeleteCustomer(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.ValidationError(c, "Invalid customer ID")
		return
	}

	if err := h.customerUseCase.DeleteCustomer(c.Request.Context(), uint(id)); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"message": "Customer deleted successfully",
	})
}
