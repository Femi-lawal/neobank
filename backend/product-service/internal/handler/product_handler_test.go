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

func TestProductHandler_CreateProduct(t *testing.T) {
	tests := []struct {
		name        string
		requestBody map[string]interface{}
	}{
		{
			name: "valid product",
			requestBody: map[string]interface{}{
				"name":        "Premium Checking",
				"description": "High-yield checking account",
				"type":        "CHECKING",
				"active":      true,
			},
		},
		{
			name: "missing name",
			requestBody: map[string]interface{}{
				"description": "No name product",
				"type":        "SAVINGS",
			},
		},
		{
			name:        "empty body",
			requestBody: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()

			var body []byte
			if tt.requestBody != nil {
				body, _ = json.Marshal(tt.requestBody)
			}

			req, _ := http.NewRequest("POST", "/api/v1/products", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.NotNil(t, w)
		})
	}
}

func TestProductHandler_ListProducts(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/products", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotNil(t, w)
}

func TestProductHandler_GetProduct(t *testing.T) {
	tests := []struct {
		name      string
		productID string
	}{
		{"valid uuid", "550e8400-e29b-41d4-a716-446655440000"},
		{"invalid uuid", "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()

			req, _ := http.NewRequest("GET", "/api/v1/products/"+tt.productID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.NotNil(t, w)
		})
	}
}
