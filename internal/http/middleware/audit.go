package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/usecases/audit"
)

// responseWriter wraps gin.ResponseWriter to capture response body
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// AuditLogger is a middleware that logs all requests to audit_logs table
func AuditLogger(auditUseCase *audit.UseCase) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip audit logging for certain paths
		if shouldSkipAudit(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Record start time
		startTime := time.Now()

		// Capture request body for POST/PUT/PATCH
		var requestBody []byte
		if c.Request.Method != "GET" && c.Request.Method != "DELETE" {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// Restore the body for the next handler
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Wrap response writer to capture response body
		blw := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = blw

		// Process request
		c.Next()

		// Calculate duration and finish time
		duration := time.Since(startTime).Milliseconds()
		finishedAt := time.Now()

		// Get user ID from context (set by auth middleware)
		// NOTE: Skipping user_id for now due to type mismatch
		// User.ID is uuid.UUID but AuditLog.UserID is *uint
		// This will be nil for all requests until we align the types
		var userID *uint = nil

		// Determine action based on HTTP method
		action := getActionFromMethod(c.Request.Method)

		// Extract resource and resource ID from path
		resource, resourceID := extractResourceInfo(c.Request.URL.Path, c.Request.Method)

		// Prepare request body (sanitized)
		var reqBody *string
		if len(requestBody) > 0 {
			sanitized := sanitizeJSON(requestBody)
			if sanitized != "" {
				reqBody = &sanitized
			}
		}

		// Prepare response body (limit size to prevent huge logs)
		var respBody *string
		if blw.body.Len() > 0 && blw.body.Len() < 10000 { // Max 10KB
			respBodyStr := blw.body.String()
			respBody = &respBodyStr
		}

		// Get old and new values for update operations (structured data)
		var newValues *string
		if c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			if len(requestBody) > 0 {
				sanitized := sanitizeJSON(requestBody)
				if sanitized != "" {
					newValues = &sanitized
				}
			}
		}

		// Create audit log entry
		auditLog := &domain.AuditLog{
			UserID:       userID,
			Action:       action,
			Resource:     resource,
			ResourceID:   resourceID,
			Description:  generateDescription(c.Request.Method, resource, c.Writer.Status()),
			IPAddress:    c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			Method:       c.Request.Method,
			Path:         c.Request.URL.Path,
			StatusCode:   c.Writer.Status(),
			Duration:     duration,
			RequestBody:  reqBody,
			ResponseBody: respBody,
			NewValues:    newValues,
			CreatedAt:    startTime,
			FinishedAt:   &finishedAt,
		}

		// Save audit log asynchronously to not block the response
		// Use background context to avoid context cancellation after request completes
		go func() {
			ctx := context.Background()
			if err := auditUseCase.Create(ctx, auditLog); err != nil {
				// Log error but don't fail the request
				// You might want to use a proper logger here
				println("Failed to create audit log:", err.Error())
			}
		}()
	}
}

// shouldSkipAudit determines if a path should skip audit logging
func shouldSkipAudit(path string) bool {
	skipPaths := []string{
		"/health",
		"/metrics",
		"/swagger",
		"/api/v1/ws",
		"/api/v1/ping",
		"/api/v1/audit-logs", // Don't log audit log queries to avoid recursion
	}

	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	return false
}

// getActionFromMethod maps HTTP method to audit action
func getActionFromMethod(method string) domain.AuditAction {
	switch method {
	case "POST":
		return domain.AuditActionCreate
	case "GET":
		return domain.AuditActionRead
	case "PUT", "PATCH":
		return domain.AuditActionUpdate
	case "DELETE":
		return domain.AuditActionDelete
	default:
		return domain.AuditActionRead
	}
}

// extractResourceInfo extracts resource type and ID from the path
func extractResourceInfo(path, method string) (string, *uint) {
	// Remove /api/v1/ prefix
	path = strings.TrimPrefix(path, "/api/v1/")

	// Split path into segments
	segments := strings.Split(path, "/")
	if len(segments) == 0 {
		return "unknown", nil
	}

	// First segment is usually the resource
	resource := segments[0]

	// Try to extract resource ID (usually the second segment for detail endpoints)
	var resourceID *uint
	if len(segments) >= 2 && method != "POST" {
		// Try to parse as uint
		var id uint
		if _, err := fmt.Sscanf(segments[1], "%d", &id); err == nil {
			resourceID = &id
		}
	}

	return resource, resourceID
}

// generateDescription generates a human-readable description
func generateDescription(method, resource string, statusCode int) string {
	action := ""
	switch method {
	case "POST":
		action = "Created"
	case "GET":
		action = "Viewed"
	case "PUT", "PATCH":
		action = "Updated"
	case "DELETE":
		action = "Deleted"
	default:
		action = "Accessed"
	}

	status := "successfully"
	if statusCode >= 400 {
		status = "with errors"
	}

	return fmt.Sprintf("%s %s %s", action, resource, status)
}

// sanitizeJSON sanitizes JSON by removing sensitive fields
func sanitizeJSON(data []byte) string {
	var jsonData map[string]interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		// If not valid JSON, return empty string
		return ""
	}

	// Remove sensitive fields
	sensitiveFields := []string{"password", "token", "secret", "api_key", "refresh_token", "access_token"}
	for _, field := range sensitiveFields {
		delete(jsonData, field)
	}

	sanitized, err := json.Marshal(jsonData)
	if err != nil {
		return ""
	}
	return string(sanitized)
}
