package errors

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AppError represents a structured application error
type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    any    `json:"details,omitempty"`
	HTTPStatus int    `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// WithDetails returns a new error with additional details
func (e *AppError) WithDetails(details any) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    e.Message,
		Details:    details,
		HTTPStatus: e.HTTPStatus,
	}
}

// WithMessage returns a new error with a custom message
func (e *AppError) WithMessage(msg string) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    msg,
		Details:    e.Details,
		HTTPStatus: e.HTTPStatus,
	}
}

// =============================================================================
// Standard Error Definitions
// =============================================================================

// Authentication Errors
var (
	ErrUnauthorized = &AppError{
		Code:       "UNAUTHORIZED",
		Message:    "Authentication required",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrInvalidToken = &AppError{
		Code:       "INVALID_TOKEN",
		Message:    "Invalid or expired authentication token",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrForbidden = &AppError{
		Code:       "FORBIDDEN",
		Message:    "You do not have permission to access this resource",
		HTTPStatus: http.StatusForbidden,
	}
)

// Validation Errors
var (
	ErrValidation = &AppError{
		Code:       "VALIDATION_ERROR",
		Message:    "Request validation failed",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInvalidRequest = &AppError{
		Code:       "INVALID_REQUEST",
		Message:    "Invalid request format",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrMissingField = &AppError{
		Code:       "MISSING_FIELD",
		Message:    "Required field is missing",
		HTTPStatus: http.StatusBadRequest,
	}
)

// Resource Errors
var (
	ErrNotFound = &AppError{
		Code:       "NOT_FOUND",
		Message:    "Resource not found",
		HTTPStatus: http.StatusNotFound,
	}

	ErrAlreadyExists = &AppError{
		Code:       "ALREADY_EXISTS",
		Message:    "Resource already exists",
		HTTPStatus: http.StatusConflict,
	}
)

// Business Logic Errors
var (
	ErrInsufficientFunds = &AppError{
		Code:       "INSUFFICIENT_FUNDS",
		Message:    "Insufficient funds for this transaction",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrAccountFrozen = &AppError{
		Code:       "ACCOUNT_FROZEN",
		Message:    "Account is frozen and cannot process transactions",
		HTTPStatus: http.StatusForbidden,
	}

	ErrTransferLimit = &AppError{
		Code:       "TRANSFER_LIMIT_EXCEEDED",
		Message:    "Transfer amount exceeds limit",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrSameAccount = &AppError{
		Code:       "SAME_ACCOUNT",
		Message:    "Cannot transfer to the same account",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInvalidAmount = &AppError{
		Code:       "INVALID_AMOUNT",
		Message:    "Invalid transaction amount",
		HTTPStatus: http.StatusBadRequest,
	}
)

// Server Errors
var (
	ErrInternal = &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    "An internal error occurred",
		HTTPStatus: http.StatusInternalServerError,
	}

	ErrServiceUnavailable = &AppError{
		Code:       "SERVICE_UNAVAILABLE",
		Message:    "Service is temporarily unavailable",
		HTTPStatus: http.StatusServiceUnavailable,
	}

	ErrTimeout = &AppError{
		Code:       "TIMEOUT",
		Message:    "Request timed out",
		HTTPStatus: http.StatusGatewayTimeout,
	}
)

// Rate Limiting
var (
	ErrRateLimited = &AppError{
		Code:       "RATE_LIMITED",
		Message:    "Too many requests, please slow down",
		HTTPStatus: http.StatusTooManyRequests,
	}
)

// =============================================================================
// Error Response Helpers
// =============================================================================

// ErrorResponse is the standard error response format
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody is the error details in the response
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// RespondWithError writes an error response to the gin context
func RespondWithError(c *gin.Context, err *AppError) {
	c.AbortWithStatusJSON(err.HTTPStatus, ErrorResponse{
		Error: ErrorBody{
			Code:    err.Code,
			Message: err.Message,
			Details: err.Details,
		},
	})
}

// ErrorMiddleware handles panics and converts them to proper error responses
func ErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				RespondWithError(c, ErrInternal.WithDetails(fmt.Sprintf("%v", r)))
			}
		}()
		c.Next()
	}
}

// NewError creates a custom error with the given code, message, and status
func NewError(code string, message string, status int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: status,
	}
}

// IsAppError checks if the error is an AppError and returns it
func IsAppError(err error) (*AppError, bool) {
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}
