package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/owner/go-cms/pkg/errors"
	"github.com/owner/go-cms/pkg/pagination"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// ErrorInfo represents error information in the response
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Meta represents metadata in the response (e.g., pagination)
type Meta struct {
	Pagination *PaginationMeta `json:"pagination,omitempty"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Total       int64  `json:"total"`
	PerPage     int    `json:"per_page"`
	CurrentPage int    `json:"current_page"`
	LastPage    int    `json:"last_page"`
	NextCursor  string `json:"next_cursor,omitempty"`
	PrevCursor  string `json:"prev_cursor,omitempty"`
}

// Success sends a successful response
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// SuccessWithMeta sends a successful response with metadata
func SuccessWithMeta(c *gin.Context, data interface{}, meta *Meta) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// SuccessWithPagination sends a successful response with pagination metadata
func SuccessWithPagination(c *gin.Context, data interface{}, total int64, page *pagination.OffsetPagination) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Pagination: &PaginationMeta{
				Total:       total,
				PerPage:     page.PerPage,
				CurrentPage: page.Page,
				LastPage:    page.GetLastPage(),
			},
		},
	})
}

// Created sends a created response
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// NoContent sends a no content response
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error sends an error response
func Error(c *gin.Context, err error) {
	if appErr := errors.GetAppError(err); appErr != nil {
		c.JSON(appErr.Status, Response{
			Success: false,
			Error: &ErrorInfo{
				Code:    appErr.Code,
				Message: appErr.Message,
			},
		})
		return
	}

	// Default to internal server error
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    errors.ErrCodeInternal,
			Message: "Internal server error",
		},
	})
}

// BadRequest sends a bad request error response
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    errors.ErrCodeBadRequest,
			Message: message,
		},
	})
}

// Unauthorized sends an unauthorized error response
func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    errors.ErrCodeUnauthorized,
			Message: message,
		},
	})
}

// Forbidden sends a forbidden error response
func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    errors.ErrCodeForbidden,
			Message: message,
		},
	})
}

// NotFound sends a not found error response
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    errors.ErrCodeNotFound,
			Message: message,
		},
	})
}

// ValidationError sends a validation error response
func ValidationError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    errors.ErrCodeValidation,
			Message: message,
		},
	})
}

// InternalError sends an internal server error response
func InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    errors.ErrCodeInternal,
			Message: message,
		},
	})
}
