package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/owner/go-cms/pkg/logger"
	"github.com/owner/go-cms/pkg/response"
	"go.uber.org/zap"
)

// RecoveryMiddleware recovers from panics and logs the error
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("client_ip", c.ClientIP()),
				)

				// Return error response
				c.AbortWithStatusJSON(http.StatusInternalServerError, response.Response{
					Success: false,
					Error: &response.ErrorInfo{
						Code:    "INTERNAL_ERROR",
						Message: "Internal server error",
					},
				})
			}
		}()

		c.Next()
	}
}
