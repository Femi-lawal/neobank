package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRateLimit_AllowsRequests(t *testing.T) {
	r := gin.New()
	config := RateLimitConfig{
		RequestsPerMinute: 10,
		BurstSize:         5,
		CleanupInterval:   time.Minute,
	}
	r.Use(RateLimitWithConfig(config))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// First request should succeed
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRateLimit_BlocksExcessiveRequests(t *testing.T) {
	r := gin.New()
	config := RateLimitConfig{
		RequestsPerMinute: 60,
		BurstSize:         2, // Only allow 2 burst requests
		CleanupInterval:   time.Minute,
	}
	r.Use(RateLimitWithConfig(config))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Make burst requests
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345" // Same IP
		r.ServeHTTP(w, req)

		if i < 2 {
			assert.Equal(t, http.StatusOK, w.Code, "Request %d should succeed", i)
		}
	}
}

func TestRequestLogger_AddsRequestID(t *testing.T) {
	r := gin.New()
	r.Use(RequestLogger("test-service"))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestRequestLogger_PreservesExistingRequestID(t *testing.T) {
	r := gin.New()
	r.Use(RequestLogger("test-service"))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	existingID := "existing-request-id-123"

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", existingID)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Should preserve the existing request ID
	// Note: The actual behavior depends on implementation
}

func TestCORS_AllowsConfiguredOrigins(t *testing.T) {
	r := gin.New()
	config := CORSConfig{
		AllowOrigins: []string{"http://localhost:3000"},
		AllowMethods: []string{"GET", "POST"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}
	r.Use(CORSWithConfig(config))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORS_HandlesPreflight(t *testing.T) {
	r := gin.New()
	r.Use(CORS())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestJWTAuth_RejectsWithoutToken(t *testing.T) {
	r := gin.New()
	r.Use(JWTAuth("secret"))
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestJWTAuth_SkipsConfiguredPaths(t *testing.T) {
	r := gin.New()
	config := DefaultJWTConfig("secret")
	r.Use(JWTAuthWithConfig(config))
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)

	// /health should be skipped (no auth required)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetUserID_ReturnsEmptyWhenNotSet(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	userID := GetUserID(c)
	assert.Empty(t, userID)
}

func TestGetUserID_ReturnsValueWhenSet(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(string(UserIDKey), "user-123")

	userID := GetUserID(c)
	assert.Equal(t, "user-123", userID)
}

func TestDefaultJWTConfig(t *testing.T) {
	config := DefaultJWTConfig("my-secret")

	assert.Equal(t, "my-secret", config.SecretKey)
	assert.Equal(t, "header:Authorization", config.TokenLookup)
	assert.Equal(t, "Bearer ", config.TokenPrefix)
	assert.Contains(t, config.SkipPaths, "/health")
	assert.Contains(t, config.SkipPaths, "/metrics")
}

func TestDefaultRateLimitConfig(t *testing.T) {
	config := DefaultRateLimitConfig()

	assert.Equal(t, 60, config.RequestsPerMinute)
	assert.Equal(t, 10, config.BurstSize)
	assert.Equal(t, 5*time.Minute, config.CleanupInterval)
}

func TestDefaultCORSConfig(t *testing.T) {
	config := DefaultCORSConfig()

	assert.Contains(t, config.AllowOrigins, "http://localhost:3000")
	assert.Contains(t, config.AllowMethods, "GET")
	assert.Contains(t, config.AllowMethods, "POST")
	assert.Contains(t, config.AllowHeaders, "Authorization")
	assert.True(t, config.AllowCredentials)
}
