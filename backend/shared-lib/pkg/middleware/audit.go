package middleware

import (
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditEvent represents a security audit log entry
type AuditEvent struct {
	Timestamp  time.Time              `json:"timestamp"`
	RequestID  string                 `json:"request_id"`
	UserID     string                 `json:"user_id,omitempty"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource"`
	Method     string                 `json:"method"`
	Path       string                 `json:"path"`
	IP         string                 `json:"ip"`
	UserAgent  string                 `json:"user_agent"`
	StatusCode int                    `json:"status_code"`
	Duration   int64                  `json:"duration_ms"`
	Success    bool                   `json:"success"`
	ErrorMsg   string                 `json:"error,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// AuditLogger provides security audit logging
type AuditLogger struct {
	// In production, this would write to a secure audit log store
	// like Elasticsearch, Splunk, or a SIEM system
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger() *AuditLogger {
	return &AuditLogger{}
}

// Log writes an audit event
func (a *AuditLogger) Log(event *AuditEvent) {
	// Serialize and write to audit log
	data, _ := json.Marshal(event)

	// In production, write to:
	// - Secure audit log database
	// - SIEM system (Splunk, ELK, etc.)
	// - Immutable storage for compliance

	// For now, output to structured log
	println("[AUDIT]", string(data))
}

// AuditMiddleware logs all security-relevant actions
func AuditMiddleware(logger *AuditLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Process request
		c.Next()

		// Build audit event
		event := &AuditEvent{
			Timestamp:  startTime,
			RequestID:  c.GetString("requestID"),
			UserID:     c.GetString("userID"),
			Action:     classifyAction(c.Request.Method, c.FullPath()),
			Resource:   c.FullPath(),
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			IP:         c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			StatusCode: c.Writer.Status(),
			Duration:   time.Since(startTime).Milliseconds(),
			Success:    c.Writer.Status() < 400,
		}

		// Add error if present
		if len(c.Errors) > 0 {
			event.ErrorMsg = c.Errors.Last().Error()
		}

		// Log security-sensitive endpoints always
		if isSensitiveEndpoint(c.FullPath()) || c.Writer.Status() >= 400 {
			logger.Log(event)
		}
	}
}

// classifyAction determines the action type for audit logging
func classifyAction(method, path string) string {
	sensitiveActions := map[string]string{
		"/login":    "USER_LOGIN",
		"/register": "USER_REGISTER",
		"/logout":   "USER_LOGOUT",
		"/password": "PASSWORD_CHANGE",
		"/transfer": "MONEY_TRANSFER",
		"/cards":    "CARD_OPERATION",
		"/accounts": "ACCOUNT_OPERATION",
	}

	for pattern, action := range sensitiveActions {
		if contains(path, pattern) {
			return action
		}
	}

	return method + "_" + path
}

// isSensitiveEndpoint checks if the endpoint requires audit logging
func isSensitiveEndpoint(path string) bool {
	sensitive := []string{
		"/login", "/register", "/logout",
		"/password", "/transfer", "/cards",
		"/accounts", "/profile", "/admin",
	}

	for _, s := range sensitive {
		if contains(path, s) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && contains(s[1:], substr) || s[:len(substr)] == substr)
}
