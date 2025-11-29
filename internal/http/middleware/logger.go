package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/owner/go-cms/pkg/logger"
	"go.uber.org/zap"
)

// LoggerMiddleware logs HTTP requests
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Get correlation ID or generate one
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			correlationID = generateCorrelationID()
		}
		c.Set("correlation_id", correlationID)
		c.Header("X-Correlation-ID", correlationID)

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status code
		statusCode := c.Writer.Status()

		// Get client IP
		clientIP := c.ClientIP()

		// Get user agent
		userAgent := c.Request.UserAgent()

		// Get method
		method := c.Request.Method

		// Get errors if any
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// Log fields
		fields := []zap.Field{
			zap.String("correlation_id", correlationID),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
			zap.String("user_agent", userAgent),
		}

		if errorMessage != "" {
			fields = append(fields, zap.String("error", errorMessage))
		}

		// Log based on status code
		if statusCode >= 500 {
			logger.Error("Server error", fields...)
		} else if statusCode >= 400 {
			logger.Warn("Client error", fields...)
		} else {
			logger.Info("Request completed", fields...)
		}
	}
}

// generateCorrelationID generates a unique correlation ID
func generateCorrelationID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
