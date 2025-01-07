package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDHeader is the header name for request ID
const RequestIDHeader = "X-Request-ID"

// RequestLogger returns a request logging middleware
func RequestLogger(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate or use existing request ID
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Header(RequestIDHeader, requestID)

		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Log request start
		slog.Info("Request started",
			"service", serviceName,
			"request_id", requestID,
			"method", c.Request.Method,
			"path", path,
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Collect response info
		status := c.Writer.Status()
		size := c.Writer.Size()

		// Determine log level based on status
		logAttrs := []any{
			"service", serviceName,
			"request_id", requestID,
			"method", c.Request.Method,
			"path", path,
			"query", query,
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"size", size,
			"client_ip", c.ClientIP(),
		}

		// Add user ID if authenticated
		if userID := GetUserID(c); userID != "" {
			logAttrs = append(logAttrs, "user_id", userID)
		}

		// Add error if any
		if len(c.Errors) > 0 {
			logAttrs = append(logAttrs, "errors", c.Errors.String())
		}

		// Log based on status code
		switch {
		case status >= 500:
			slog.Error("Request completed with server error", logAttrs...)
		case status >= 400:
			slog.Warn("Request completed with client error", logAttrs...)
		default:
			slog.Info("Request completed", logAttrs...)
		}
	}
}

// GetRequestID retrieves the request ID from the context
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}
