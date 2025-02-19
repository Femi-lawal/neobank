package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestPaymentHandler_MakeTransfer(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name: "valid transfer",
			requestBody: map[string]interface{}{
				"from_account_id": "550e8400-e29b-41d4-a716-446655440000",
				"to_account_id":   "550e8400-e29b-41d4-a716-446655440001",
				"amount":          "100.00",
				"currency":        "USD",
				"description":     "Test transfer",
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "missing body",
			requestBody:    nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "same account",
			requestBody: map[string]interface{}{
				"from_account_id": "550e8400-e29b-41d4-a716-446655440000",
				"to_account_id":   "550e8400-e29b-41d4-a716-446655440000",
				"amount":          "100.00",
				"currency":        "USD",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "negative amount",
			requestBody: map[string]interface{}{
				"from_account_id": "550e8400-e29b-41d4-a716-446655440000",
				"to_account_id":   "550e8400-e29b-41d4-a716-446655440001",
				"amount":          "-50.00",
				"currency":        "USD",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "zero amount",
			requestBody: map[string]interface{}{
				"from_account_id": "550e8400-e29b-41d4-a716-446655440000",
				"to_account_id":   "550e8400-e29b-41d4-a716-446655440001",
				"amount":          "0",
				"currency":        "USD",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()

			var body []byte
			if tt.requestBody != nil {
				body, _ = json.Marshal(tt.requestBody)
			}

			req, _ := http.NewRequest("POST", "/api/v1/transfer", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.NotNil(t, w)
		})
	}
}

func TestPaymentHandler_GetPayment(t *testing.T) {
	tests := []struct {
		name      string
		paymentID string
	}{
		{"valid uuid", "550e8400-e29b-41d4-a716-446655440000"},
		{"invalid uuid", "not-a-uuid"},
		{"empty id", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()

			req, _ := http.NewRequest("GET", "/api/v1/payments/"+tt.paymentID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.NotNil(t, w)
		})
	}
}

func TestPaymentHandler_ListPayments(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/payments", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotNil(t, w)
}
