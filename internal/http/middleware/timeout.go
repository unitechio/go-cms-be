package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/owner/go-cms/pkg/response"
)

// TimeoutMiddleware adds a timeout to requests
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Replace request context
		c.Request = c.Request.WithContext(ctx)

		// Channel to signal completion
		finished := make(chan struct{})

		go func() {
			c.Next()
			finished <- struct{}{}
		}()

		select {
		case <-finished:
			// Request completed successfully
			return
		case <-ctx.Done():
			// Request timed out
			c.AbortWithStatusJSON(http.StatusRequestTimeout, response.Response{
				Success: false,
				Error: &response.ErrorInfo{
					Code:    "TIMEOUT",
					Message: "Request timeout",
				},
			})
		}
	}
}
