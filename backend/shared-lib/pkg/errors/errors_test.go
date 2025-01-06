package errors

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAppError_Error(t *testing.T) {
	err := &AppError{
		Code:    "TEST_ERROR",
		Message: "This is a test error",
	}

	assert.Equal(t, "[TEST_ERROR] This is a test error", err.Error())
}

func TestAppError_WithDetails(t *testing.T) {
	original := ErrValidation
	details := map[string]string{"field": "email", "reason": "invalid format"}

	withDetails := original.WithDetails(details)

	assert.Equal(t, original.Code, withDetails.Code)
	assert.Equal(t, original.Message, withDetails.Message)
	assert.Equal(t, details, withDetails.Details)
	assert.Equal(t, original.HTTPStatus, withDetails.HTTPStatus)
	// Original should be unchanged
	assert.Nil(t, original.Details)
}

func TestAppError_WithMessage(t *testing.T) {
	original := ErrNotFound
	customMessage := "User not found"

	withMessage := original.WithMessage(customMessage)

	assert.Equal(t, original.Code, withMessage.Code)
	assert.Equal(t, customMessage, withMessage.Message)
	assert.Equal(t, original.HTTPStatus, withMessage.HTTPStatus)
	// Original should be unchanged
	assert.NotEqual(t, customMessage, original.Message)
}

func TestRespondWithError(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	RespondWithError(c, ErrUnauthorized)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "UNAUTHORIZED")
}

func TestRespondWithError_WithDetails(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	err := ErrValidation.WithDetails(map[string]string{"field": "email"})
	RespondWithError(c, err)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "VALIDATION_ERROR")
	assert.Contains(t, w.Body.String(), "email")
}

func TestNewError(t *testing.T) {
	err := NewError("CUSTOM_ERROR", "Custom error message", http.StatusTeapot)

	assert.Equal(t, "CUSTOM_ERROR", err.Code)
	assert.Equal(t, "Custom error message", err.Message)
	assert.Equal(t, http.StatusTeapot, err.HTTPStatus)
}

func TestIsAppError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantOk   bool
		wantCode string
	}{
		{
			name:     "is AppError",
			err:      ErrNotFound,
			wantOk:   true,
			wantCode: "NOT_FOUND",
		},
		{
			name:   "is not AppError",
			err:    assert.AnError,
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr, ok := IsAppError(tt.err)
			assert.Equal(t, tt.wantOk, ok)
			if ok {
				assert.Equal(t, tt.wantCode, appErr.Code)
			}
		})
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name   string
		err    *AppError
		status int
	}{
		{"Unauthorized", ErrUnauthorized, http.StatusUnauthorized},
		{"InvalidToken", ErrInvalidToken, http.StatusUnauthorized},
		{"Forbidden", ErrForbidden, http.StatusForbidden},
		{"Validation", ErrValidation, http.StatusBadRequest},
		{"InvalidRequest", ErrInvalidRequest, http.StatusBadRequest},
		{"NotFound", ErrNotFound, http.StatusNotFound},
		{"AlreadyExists", ErrAlreadyExists, http.StatusConflict},
		{"InsufficientFunds", ErrInsufficientFunds, http.StatusBadRequest},
		{"AccountFrozen", ErrAccountFrozen, http.StatusForbidden},
		{"TransferLimit", ErrTransferLimit, http.StatusBadRequest},
		{"SameAccount", ErrSameAccount, http.StatusBadRequest},
		{"InvalidAmount", ErrInvalidAmount, http.StatusBadRequest},
		{"Internal", ErrInternal, http.StatusInternalServerError},
		{"ServiceUnavailable", ErrServiceUnavailable, http.StatusServiceUnavailable},
		{"Timeout", ErrTimeout, http.StatusGatewayTimeout},
		{"RateLimited", ErrRateLimited, http.StatusTooManyRequests},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.status, tt.err.HTTPStatus)
			assert.NotEmpty(t, tt.err.Code)
			assert.NotEmpty(t, tt.err.Message)
		})
	}
}

func TestErrorMiddleware_RecoversPanic(t *testing.T) {
	r := gin.New()
	r.Use(ErrorMiddleware())
	r.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/panic", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "INTERNAL_ERROR")
}
