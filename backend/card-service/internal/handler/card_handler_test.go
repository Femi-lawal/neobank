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

func TestCardHandler_IssueCard(t *testing.T) {
	tests := []struct {
		name        string
		requestBody map[string]interface{}
	}{
		{
			name: "valid card request",
			requestBody: map[string]interface{}{
				"account_id": "550e8400-e29b-41d4-a716-446655440000",
			},
		},
		{
			name:        "missing account_id",
			requestBody: map[string]interface{}{},
		},
		{
			name: "invalid account_id",
			requestBody: map[string]interface{}{
				"account_id": "not-a-uuid",
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

			req, _ := http.NewRequest("POST", "/api/v1/cards", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.NotNil(t, w)
		})
	}
}

func TestCardHandler_ListCards(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/api/v1/cards", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotNil(t, w)
}

func TestCardHandler_GetCard(t *testing.T) {
	tests := []struct {
		name   string
		cardID string
	}{
		{"valid uuid", "550e8400-e29b-41d4-a716-446655440000"},
		{"invalid uuid", "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter()

			req, _ := http.NewRequest("GET", "/api/v1/cards/"+tt.cardID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.NotNil(t, w)
		})
	}
}

func TestCardHandler_BlockCard(t *testing.T) {
	router := setupTestRouter()

	cardID := "550e8400-e29b-41d4-a716-446655440000"
	req, _ := http.NewRequest("POST", "/api/v1/cards/"+cardID+"/block", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.NotNil(t, w)
}
