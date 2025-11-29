package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/internal/core/usecases/audit"
	"github.com/owner/go-cms/pkg/pagination"
	"github.com/owner/go-cms/pkg/response"
)

// AuditLogHandler handles audit log HTTP requests
type AuditLogHandler struct {
	useCase *audit.UseCase
}

// NewAuditLogHandler creates a new audit log handler
func NewAuditLogHandler(useCase *audit.UseCase) *AuditLogHandler {
	return &AuditLogHandler{
		useCase: useCase,
	}
}

// ListAuditLogs godoc
// @Summary List audit logs
// @Description Get a list of audit logs with filters and pagination
// @Tags Audit Logs
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param user_id query int false "Filter by user ID"
// @Param action query string false "Filter by action"
// @Param resource query string false "Filter by resource"
// @Param resource_id query int false "Filter by resource ID"
// @Param ip_address query string false "Filter by IP address"
// @Param date_from query string false "Filter from date (YYYY-MM-DD)"
// @Param date_to query string false "Filter to date (YYYY-MM-DD)"
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]domain.AuditLog}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /audit-logs [get]
func (h *AuditLogHandler) ListAuditLogs(c *gin.Context) {
	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Parse filters
	filter := repositories.AuditLogFilter{
		Action:    domain.AuditAction(c.Query("action")),
		Resource:  c.Query("resource"),
		IPAddress: c.Query("ip_address"),
		DateFrom:  c.Query("date_from"),
		DateTo:    c.Query("date_to"),
	}

	if userID := c.Query("user_id"); userID != "" {
		if id, err := strconv.ParseUint(userID, 10, 32); err == nil {
			uid := uint(id)
			filter.UserID = &uid
		}
	}

	if resourceID := c.Query("resource_id"); resourceID != "" {
		if id, err := strconv.ParseUint(resourceID, 10, 32); err == nil {
			rid := uint(id)
			filter.ResourceID = &rid
		}
	}

	// Get audit logs
	logs, total, err := h.useCase.List(c.Request.Context(), filter, &pagination.OffsetPagination{
		Page:    page,
		Limit:   limit,
		PerPage: limit,
		Total:   0,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithPagination(c, logs, total, &pagination.OffsetPagination{
		Page:    page,
		PerPage: limit,
		Total:   total,
	})
}

// GetAuditLog godoc
// @Summary Get audit log by ID
// @Description Get a specific audit log by ID
// @Tags Audit Logs
// @Accept json
// @Produce json
// @Param id path int true "Audit Log ID"
// @Security BearerAuth
// @Success 200 {object} response.Response{data=domain.AuditLog}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /audit-logs/{id} [get]
func (h *AuditLogHandler) GetAuditLog(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid audit log ID")
		return
	}

	log, err := h.useCase.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, log)
}

// GetUserAuditLogs godoc
// @Summary Get audit logs for a user
// @Description Get audit logs for a specific user
// @Tags Audit Logs
// @Accept json
// @Produce json
// @Param user_id path int true "User ID"
// @Param limit query int false "Limit" default(50)
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]domain.AuditLog}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /audit-logs/user/{user_id} [get]
func (h *AuditLogHandler) GetUserAuditLogs(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 100 {
		limit = 50
	}

	logs, err := h.useCase.GetByUserID(c.Request.Context(), uint(userID), limit)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, logs)
}

// GetResourceAuditLogs godoc
// @Summary Get audit logs for a resource
// @Description Get audit logs for a specific resource
// @Tags Audit Logs
// @Accept json
// @Produce json
// @Param resource query string true "Resource type"
// @Param resource_id query int true "Resource ID"
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]domain.AuditLog}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /audit-logs/resource [get]
func (h *AuditLogHandler) GetResourceAuditLogs(c *gin.Context) {
	resource := c.Query("resource")
	if resource == "" {
		response.BadRequest(c, "Resource is required")
		return
	}

	resourceID, err := strconv.ParseUint(c.Query("resource_id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid resource ID")
		return
	}

	logs, err := h.useCase.GetByResource(c.Request.Context(), resource, uint(resourceID))
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, logs)
}

// CleanupOldLogs godoc
// @Summary Cleanup old audit logs
// @Description Delete audit logs older than specified days
// @Tags Audit Logs
// @Accept json
// @Produce json
// @Param days query int true "Days to keep" default(90)
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /audit-logs/cleanup [delete]
func (h *AuditLogHandler) CleanupOldLogs(c *gin.Context) {
	days, err := strconv.Atoi(c.DefaultQuery("days", "90"))
	if err != nil || days <= 0 {
		response.BadRequest(c, "Invalid days parameter")
		return
	}

	if err := h.useCase.DeleteOlderThan(c.Request.Context(), days); err != nil {
		response.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Old audit logs deleted successfully",
	})
}
