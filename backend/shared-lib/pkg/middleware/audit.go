package middleware

import (
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AuditEventType represents the type of audit event
type AuditEventType string

const (
	// Authentication events
	AuditEventLogin          AuditEventType = "USER_LOGIN"
	AuditEventLoginFailed    AuditEventType = "USER_LOGIN_FAILED"
	AuditEventLogout         AuditEventType = "USER_LOGOUT"
	AuditEventRegister       AuditEventType = "USER_REGISTER"
	AuditEventPasswordChange AuditEventType = "PASSWORD_CHANGE"
	AuditEventPasswordReset  AuditEventType = "PASSWORD_RESET"
	AuditEventMFAEnroll      AuditEventType = "MFA_ENROLL"
	AuditEventMFAVerify      AuditEventType = "MFA_VERIFY"
	AuditEventSessionCreate  AuditEventType = "SESSION_CREATE"
	AuditEventSessionRevoke  AuditEventType = "SESSION_REVOKE"

	// Account events
	AuditEventAccountCreate AuditEventType = "ACCOUNT_CREATE"
	AuditEventAccountUpdate AuditEventType = "ACCOUNT_UPDATE"
	AuditEventAccountClose  AuditEventType = "ACCOUNT_CLOSE"
	AuditEventAccountView   AuditEventType = "ACCOUNT_VIEW"

	// Money movement events
	AuditEventTransferInit     AuditEventType = "TRANSFER_INITIATED"
	AuditEventTransferComplete AuditEventType = "TRANSFER_COMPLETED"
	AuditEventTransferFailed   AuditEventType = "TRANSFER_FAILED"
	AuditEventPaymentInit      AuditEventType = "PAYMENT_INITIATED"
	AuditEventPaymentComplete  AuditEventType = "PAYMENT_COMPLETED"
	AuditEventPaymentFailed    AuditEventType = "PAYMENT_FAILED"

	// Card events
	AuditEventCardIssue     AuditEventType = "CARD_ISSUED"
	AuditEventCardActivate  AuditEventType = "CARD_ACTIVATED"
	AuditEventCardBlock     AuditEventType = "CARD_BLOCKED"
	AuditEventCardUnblock   AuditEventType = "CARD_UNBLOCKED"
	AuditEventCardPINChange AuditEventType = "CARD_PIN_CHANGED"

	// Admin events
	AuditEventAdminAction      AuditEventType = "ADMIN_ACTION"
	AuditEventPermissionChange AuditEventType = "PERMISSION_CHANGE"
	AuditEventRoleAssign       AuditEventType = "ROLE_ASSIGNED"
	AuditEventRoleRevoke       AuditEventType = "ROLE_REVOKED"

	// Security events
	AuditEventSuspiciousActivity AuditEventType = "SUSPICIOUS_ACTIVITY"
	AuditEventRateLimitExceeded  AuditEventType = "RATE_LIMIT_EXCEEDED"
	AuditEventUnauthorizedAccess AuditEventType = "UNAUTHORIZED_ACCESS"
	AuditEventInvalidInput       AuditEventType = "INVALID_INPUT_DETECTED"

	// Data access events
	AuditEventDataExport AuditEventType = "DATA_EXPORT"
	AuditEventDataView   AuditEventType = "SENSITIVE_DATA_VIEW"
	AuditEventAPICall    AuditEventType = "API_CALL"
)

// AuditSeverity represents the severity level of an audit event
type AuditSeverity string

const (
	AuditSeverityInfo     AuditSeverity = "INFO"
	AuditSeverityWarning  AuditSeverity = "WARNING"
	AuditSeverityError    AuditSeverity = "ERROR"
	AuditSeverityCritical AuditSeverity = "CRITICAL"
)

// AuditEvent represents a security audit log entry
type AuditEvent struct {
	Timestamp      time.Time              `json:"timestamp"`
	EventID        string                 `json:"event_id"`
	EventType      AuditEventType         `json:"event_type"`
	Severity       AuditSeverity          `json:"severity"`
	RequestID      string                 `json:"request_id"`
	TraceID        string                 `json:"trace_id,omitempty"`
	SpanID         string                 `json:"span_id,omitempty"`
	UserID         string                 `json:"user_id,omitempty"`
	Email          string                 `json:"email,omitempty"`
	SessionID      string                 `json:"session_id,omitempty"`
	Action         string                 `json:"action"`
	Resource       string                 `json:"resource"`
	ResourceID     string                 `json:"resource_id,omitempty"`
	Method         string                 `json:"method"`
	Path           string                 `json:"path"`
	IP             string                 `json:"ip"`
	UserAgent      string                 `json:"user_agent"`
	GeoLocation    string                 `json:"geo_location,omitempty"`
	StatusCode     int                    `json:"status_code"`
	Duration       int64                  `json:"duration_ms"`
	Success        bool                   `json:"success"`
	ErrorCode      string                 `json:"error_code,omitempty"`
	ErrorMsg       string                 `json:"error,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	ServiceName    string                 `json:"service_name"`
	ServiceVersion string                 `json:"service_version,omitempty"`
}

// AuditLogger provides security audit logging
type AuditLogger struct {
	serviceName    string
	serviceVersion string
}

// AuditLoggerConfig holds configuration for the audit logger
type AuditLoggerConfig struct {
	ServiceName    string
	ServiceVersion string
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger() *AuditLogger {
	return &AuditLogger{
		serviceName:    "unknown",
		serviceVersion: "1.0.0",
	}
}

// NewAuditLoggerWithConfig creates a new audit logger with configuration
func NewAuditLoggerWithConfig(config AuditLoggerConfig) *AuditLogger {
	return &AuditLogger{
		serviceName:    config.ServiceName,
		serviceVersion: config.ServiceVersion,
	}
}

// Log writes an audit event
func (a *AuditLogger) Log(event *AuditEvent) {
	event.ServiceName = a.serviceName
	event.ServiceVersion = a.serviceVersion

	// Serialize and write to audit log
	data, _ := json.Marshal(event)

	// In production, write to:
	// - Secure audit log database
	// - SIEM system (Splunk, ELK, etc.)
	// - Immutable storage for compliance

	// Use structured logging for audit events
	slog.Info("[AUDIT]",
		"event_type", event.EventType,
		"severity", event.Severity,
		"user_id", event.UserID,
		"action", event.Action,
		"resource", event.Resource,
		"status_code", event.StatusCode,
		"success", event.Success,
		"ip", event.IP,
		"data", string(data),
	)
}

// LogEvent creates and logs an audit event with specific type
func (a *AuditLogger) LogEvent(eventType AuditEventType, severity AuditSeverity, c *gin.Context, metadata map[string]interface{}) {
	event := &AuditEvent{
		Timestamp:      time.Now(),
		EventID:        generateEventID(),
		EventType:      eventType,
		Severity:       severity,
		RequestID:      c.GetString("requestID"),
		UserID:         c.GetString(string(UserIDKey)),
		Email:          c.GetString(string(EmailKey)),
		Action:         string(eventType),
		Resource:       c.FullPath(),
		Method:         c.Request.Method,
		Path:           c.Request.URL.Path,
		IP:             c.ClientIP(),
		UserAgent:      c.Request.UserAgent(),
		StatusCode:     c.Writer.Status(),
		Success:        c.Writer.Status() < 400,
		Metadata:       metadata,
		ServiceName:    a.serviceName,
		ServiceVersion: a.serviceVersion,
	}

	a.Log(event)
}

func generateEventID() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	id := make([]byte, 24)
	for i := range id {
		id[i] = chars[i%len(chars)]
	}
	return string(id)
}

// AuditMiddleware logs all security-relevant actions
func AuditMiddleware(logger *AuditLogger, serviceName string) gin.HandlerFunc {
	if logger == nil {
		logger = NewAuditLoggerWithConfig(AuditLoggerConfig{
			ServiceName:    serviceName,
			ServiceVersion: "1.0.0",
		})
	}

	return func(c *gin.Context) {
		startTime := time.Now()

		// Process request
		c.Next()

		// Determine event type and severity
		eventType, severity := classifyEvent(c.Request.Method, c.FullPath(), c.Writer.Status())

		// Build audit event
		event := &AuditEvent{
			Timestamp:   startTime,
			EventID:     generateEventID(),
			EventType:   eventType,
			Severity:    severity,
			RequestID:   c.GetString("requestID"),
			UserID:      c.GetString(string(UserIDKey)),
			Email:       c.GetString(string(EmailKey)),
			Action:      string(eventType),
			Resource:    c.FullPath(),
			Method:      c.Request.Method,
			Path:        c.Request.URL.Path,
			IP:          c.ClientIP(),
			UserAgent:   c.Request.UserAgent(),
			StatusCode:  c.Writer.Status(),
			Duration:    time.Since(startTime).Milliseconds(),
			Success:     c.Writer.Status() < 400,
			ServiceName: serviceName,
		}

		// Add error if present
		if len(c.Errors) > 0 {
			event.ErrorMsg = c.Errors.Last().Error()
			event.Severity = AuditSeverityError
		}

		// Log security-sensitive endpoints always, or any errors
		if isSensitiveEndpoint(c.FullPath()) || c.Writer.Status() >= 400 {
			logger.Log(event)
		}
	}
}

// classifyEvent determines the event type and severity
func classifyEvent(method, path string, statusCode int) (AuditEventType, AuditSeverity) {
	pathLower := strings.ToLower(path)

	// Determine severity based on status code
	severity := AuditSeverityInfo
	if statusCode >= 400 && statusCode < 500 {
		severity = AuditSeverityWarning
	} else if statusCode >= 500 {
		severity = AuditSeverityError
	}

	// Authentication events
	if strings.Contains(pathLower, "/login") {
		if statusCode >= 400 {
			return AuditEventLoginFailed, AuditSeverityWarning
		}
		return AuditEventLogin, AuditSeverityInfo
	}
	if strings.Contains(pathLower, "/register") {
		return AuditEventRegister, AuditSeverityInfo
	}
	if strings.Contains(pathLower, "/logout") {
		return AuditEventLogout, AuditSeverityInfo
	}
	if strings.Contains(pathLower, "/password") {
		return AuditEventPasswordChange, AuditSeverityWarning
	}

	// Money movement events
	if strings.Contains(pathLower, "/transfer") {
		if statusCode >= 400 {
			return AuditEventTransferFailed, AuditSeverityError
		}
		if method == "POST" {
			return AuditEventTransferInit, AuditSeverityInfo
		}
		return AuditEventTransferComplete, AuditSeverityInfo
	}
	if strings.Contains(pathLower, "/payment") {
		if statusCode >= 400 {
			return AuditEventPaymentFailed, AuditSeverityError
		}
		if method == "POST" {
			return AuditEventPaymentInit, AuditSeverityInfo
		}
		return AuditEventPaymentComplete, AuditSeverityInfo
	}

	// Card events
	if strings.Contains(pathLower, "/cards") {
		if strings.Contains(pathLower, "/block") || strings.Contains(pathLower, "/freeze") {
			return AuditEventCardBlock, AuditSeverityWarning
		}
		if strings.Contains(pathLower, "/unblock") || strings.Contains(pathLower, "/unfreeze") {
			return AuditEventCardUnblock, AuditSeverityInfo
		}
		if method == "POST" {
			return AuditEventCardIssue, AuditSeverityInfo
		}
		return AuditEventAPICall, severity
	}

	// Account events
	if strings.Contains(pathLower, "/accounts") {
		if method == "POST" {
			return AuditEventAccountCreate, AuditSeverityInfo
		}
		if method == "GET" {
			return AuditEventAccountView, AuditSeverityInfo
		}
		if method == "PUT" || method == "PATCH" {
			return AuditEventAccountUpdate, AuditSeverityInfo
		}
		if method == "DELETE" {
			return AuditEventAccountClose, AuditSeverityWarning
		}
	}

	// Rate limit exceeded
	if statusCode == 429 {
		return AuditEventRateLimitExceeded, AuditSeverityWarning
	}

	// Unauthorized access
	if statusCode == 401 || statusCode == 403 {
		return AuditEventUnauthorizedAccess, AuditSeverityWarning
	}

	return AuditEventAPICall, severity
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
