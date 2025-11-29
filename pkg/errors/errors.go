package errors

import (
	"fmt"
	"net/http"
)

// AppError represents an application error
type AppError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"`
	Err     error  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// New creates a new AppError
func New(code, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
	}
}

// Wrap wraps an error with additional context
func Wrap(err error, code, message string, status int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Status:  status,
		Err:     err,
	}
}

// Common error codes
const (
	// General errors
	ErrCodeInternal     = "INTERNAL_ERROR"
	ErrCodeBadRequest   = "BAD_REQUEST"
	ErrCodeNotFound     = "NOT_FOUND"
	ErrCodeUnauthorized = "UNAUTHORIZED"
	ErrCodeForbidden    = "FORBIDDEN"
	ErrCodeConflict     = "CONFLICT"
	ErrCodeValidation   = "VALIDATION_ERROR"
	ErrCodeTimeout      = "TIMEOUT"

	// Authentication errors
	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
	ErrCodeInvalidToken       = "INVALID_TOKEN"
	ErrCodeExpiredToken       = "EXPIRED_TOKEN"
	ErrCodeInvalidOTP         = "INVALID_OTP"
	ErrCodeExpiredOTP         = "EXPIRED_OTP"
	ErrCodeInvalid2FA         = "INVALID_2FA"

	// Authorization errors
	ErrCodeInsufficientPermissions = "INSUFFICIENT_PERMISSIONS"
	ErrCodeAccessDenied            = "ACCESS_DENIED"

	// Resource errors
	ErrCodeUserNotFound     = "USER_NOT_FOUND"
	ErrCodeCustomerNotFound = "CUSTOMER_NOT_FOUND"
	ErrCodePostNotFound     = "POST_NOT_FOUND"
	ErrCodeMediaNotFound    = "MEDIA_NOT_FOUND"
	ErrCodeRoleNotFound     = "ROLE_NOT_FOUND"

	// Database errors
	ErrCodeDatabaseError    = "DATABASE_ERROR"
	ErrCodeDuplicateEntry   = "DUPLICATE_ENTRY"
	ErrCodeConstraintFailed = "CONSTRAINT_FAILED"

	// File upload errors
	ErrCodeInvalidFileType = "INVALID_FILE_TYPE"
	ErrCodeFileTooLarge    = "FILE_TOO_LARGE"
	ErrCodeUploadFailed    = "UPLOAD_FAILED"
)

// Predefined errors
var (
	// General errors
	ErrInternal     = New(ErrCodeInternal, "Internal server error", http.StatusInternalServerError)
	ErrBadRequest   = New(ErrCodeBadRequest, "Bad request", http.StatusBadRequest)
	ErrNotFound     = New(ErrCodeNotFound, "Resource not found", http.StatusNotFound)
	ErrUnauthorized = New(ErrCodeUnauthorized, "Unauthorized", http.StatusUnauthorized)
	ErrForbidden    = New(ErrCodeForbidden, "Forbidden", http.StatusForbidden)
	ErrConflict     = New(ErrCodeConflict, "Resource conflict", http.StatusConflict)
	ErrValidation   = New(ErrCodeValidation, "Validation error", http.StatusBadRequest)
	ErrTimeout      = New(ErrCodeTimeout, "Request timeout", http.StatusRequestTimeout)

	// Authentication errors
	ErrInvalidCredentials = New(ErrCodeInvalidCredentials, "Invalid email or password", http.StatusUnauthorized)
	ErrInvalidToken       = New(ErrCodeInvalidToken, "Invalid token", http.StatusUnauthorized)
	ErrExpiredToken       = New(ErrCodeExpiredToken, "Token has expired", http.StatusUnauthorized)
	ErrInvalidOTP         = New(ErrCodeInvalidOTP, "Invalid OTP", http.StatusUnauthorized)
	ErrExpiredOTP         = New(ErrCodeExpiredOTP, "OTP has expired", http.StatusUnauthorized)
	ErrInvalid2FA         = New(ErrCodeInvalid2FA, "Invalid 2FA code", http.StatusUnauthorized)

	// Authorization errors
	ErrInsufficientPermissions = New(ErrCodeInsufficientPermissions, "Insufficient permissions", http.StatusForbidden)
	ErrAccessDenied            = New(ErrCodeAccessDenied, "Access denied", http.StatusForbidden)

	// Resource errors
	ErrUserNotFound     = New(ErrCodeUserNotFound, "User not found", http.StatusNotFound)
	ErrCustomerNotFound = New(ErrCodeCustomerNotFound, "Customer not found", http.StatusNotFound)
	ErrPostNotFound     = New(ErrCodePostNotFound, "Post not found", http.StatusNotFound)
	ErrMediaNotFound    = New(ErrCodeMediaNotFound, "Media not found", http.StatusNotFound)
	ErrRoleNotFound     = New(ErrCodeRoleNotFound, "Role not found", http.StatusNotFound)

	// Database errors
	ErrDatabaseError    = New(ErrCodeDatabaseError, "Database error", http.StatusInternalServerError)
	ErrDuplicateEntry   = New(ErrCodeDuplicateEntry, "Duplicate entry", http.StatusConflict)
	ErrConstraintFailed = New(ErrCodeConstraintFailed, "Constraint failed", http.StatusBadRequest)

	// File upload errors
	ErrInvalidFileType = New(ErrCodeInvalidFileType, "Invalid file type", http.StatusBadRequest)
	ErrFileTooLarge    = New(ErrCodeFileTooLarge, "File too large", http.StatusBadRequest)
	ErrUploadFailed    = New(ErrCodeUploadFailed, "Upload failed", http.StatusInternalServerError)
)

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError returns the AppError from an error
func GetAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}
	return nil
}
