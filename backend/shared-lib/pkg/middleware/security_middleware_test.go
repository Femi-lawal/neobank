package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// =====================================
// Idempotency Middleware Tests
// =====================================

func TestIdempotency_RequiresKeyForConfiguredPaths(t *testing.T) {
	r := gin.New()
	store := NewInMemoryIdempotencyStore()
	config := IdempotencyConfig{
		HeaderName:    "X-Idempotency-Key",
		TTL:           time.Hour,
		RequiredPaths: []string{"/api/v1/transfer"},
		OptionalPaths: []string{},
	}
	r.Use(Idempotency(store, config))
	r.POST("/api/v1/transfer", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/transfer", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response["error"], "Idempotency key required")
}

func TestIdempotency_AcceptsRequestWithKey(t *testing.T) {
	r := gin.New()
	store := NewInMemoryIdempotencyStore()
	config := DefaultIdempotencyConfig()
	r.Use(Idempotency(store, config))
	r.POST("/api/v1/transfer", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success", "id": "123"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/transfer", nil)
	req.Header.Set("X-Idempotency-Key", "unique-key-123")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIdempotency_ReturnsCachedResponseForSameKey(t *testing.T) {
	r := gin.New()
	store := NewInMemoryIdempotencyStore()
	config := DefaultIdempotencyConfig()

	callCount := 0
	r.Use(Idempotency(store, config))
	r.POST("/api/v1/transfer", func(c *gin.Context) {
		callCount++
		c.JSON(http.StatusOK, gin.H{"status": "success", "call": callCount})
	})

	idempotencyKey := "unique-key-456"

	// First request
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodPost, "/api/v1/transfer", nil)
	req1.Header.Set("X-Idempotency-Key", idempotencyKey)
	r.ServeHTTP(w1, req1)

	// Second request with same key
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodPost, "/api/v1/transfer", nil)
	req2.Header.Set("X-Idempotency-Key", idempotencyKey)
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Equal(t, "true", w2.Header().Get("X-Idempotent-Replayed"))
	assert.Equal(t, 1, callCount) // Handler should only be called once
}

func TestIdempotency_SkipsGETRequests(t *testing.T) {
	r := gin.New()
	store := NewInMemoryIdempotencyStore()
	config := DefaultIdempotencyConfig()
	r.Use(Idempotency(store, config))
	r.GET("/api/v1/accounts", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"accounts": []string{}})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
	// No idempotency key needed for GET
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// =====================================
// CSRF Middleware Tests
// =====================================

func TestCSRF_GeneratesTokenOnGET(t *testing.T) {
	r := gin.New()
	store := NewInMemoryCSRFStore()
	config := DefaultCSRFConfig()
	config.Secure = false // For testing
	r.Use(CSRFProtection(config, store))
	r.GET("/page", func(c *gin.Context) {
		token, _ := c.Get("csrf_token")
		c.JSON(http.StatusOK, gin.H{"csrf_token": token})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/page", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check cookie is set
	cookies := w.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "csrf_token" {
			csrfCookie = c
			break
		}
	}
	assert.NotNil(t, csrfCookie)
	assert.NotEmpty(t, csrfCookie.Value)
}

func TestCSRF_BlocksPOSTWithoutToken(t *testing.T) {
	r := gin.New()
	store := NewInMemoryCSRFStore()
	config := DefaultCSRFConfig()
	config.Secure = false
	r.Use(CSRFProtection(config, store))
	r.POST("/submit", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/submit", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCSRF_AllowsPOSTWithValidToken(t *testing.T) {
	r := gin.New()
	store := NewInMemoryCSRFStore()
	config := DefaultCSRFConfig()
	config.Secure = false
	r.Use(CSRFProtection(config, store))
	r.POST("/submit", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Generate a valid token
	token, _ := generateCSRFToken(32)
	store.Set(token, time.Hour)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/submit", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: token})
	req.Header.Set("X-CSRF-Token", token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCSRF_BlocksTokenMismatch(t *testing.T) {
	r := gin.New()
	store := NewInMemoryCSRFStore()
	config := DefaultCSRFConfig()
	config.Secure = false
	r.Use(CSRFProtection(config, store))
	r.POST("/submit", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Generate a valid token
	token, _ := generateCSRFToken(32)
	store.Set(token, time.Hour)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/submit", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: token})
	req.Header.Set("X-CSRF-Token", "wrong-token")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCSRF_SkipsHealthEndpoint(t *testing.T) {
	r := gin.New()
	store := NewInMemoryCSRFStore()
	config := DefaultCSRFConfig()
	r.Use(CSRFProtection(config, store))
	r.POST("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/health", nil)
	r.ServeHTTP(w, req)

	// /health is in skip paths, so should succeed
	assert.Equal(t, http.StatusOK, w.Code)
}

// =====================================
// Input Validation Tests
// =====================================

func TestInputValidation_BlocksSQLInjection(t *testing.T) {
	r := gin.New()
	r.Use(InputValidation(nil))
	r.GET("/users", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	testCases := []string{
		"'; DROP TABLE users; --",
		"1'; DROP TABLE users--",
		"UNION SELECT * FROM users",
	}

	for _, payload := range testCases {
		t.Run(payload, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/users?id="+payload, nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestInputValidation_BlocksXSS(t *testing.T) {
	r := gin.New()
	r.Use(InputValidation(nil))
	r.GET("/search", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	testCases := []string{
		"<script>alert('XSS')</script>",
		"javascript:alert('XSS')",
	}

	for _, payload := range testCases {
		t.Run(payload, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/search?q="+payload, nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestInputValidation_AllowsValidInput(t *testing.T) {
	r := gin.New()
	r.Use(InputValidation(nil))
	r.GET("/search", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "query": c.Query("q")})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/search?q=valid+search+term", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInputValidation_BlocksLongURL(t *testing.T) {
	r := gin.New()
	config := DefaultValidationConfig()
	config.MaxURLLength = 100
	r.Use(InputValidation(config))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Create a very long URL
	longQuery := ""
	for i := 0; i < 200; i++ {
		longQuery += "a"
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test?q="+longQuery, nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusRequestURITooLong, w.Code)
}

// =====================================
// Password Validation Tests
// =====================================

func TestValidatePassword_RequiresMinLength(t *testing.T) {
	valid, msg := ValidatePassword("Short1!")
	assert.False(t, valid)
	assert.Contains(t, msg, "8 characters")
}

func TestValidatePassword_RequiresUppercase(t *testing.T) {
	valid, msg := ValidatePassword("lowercase1!")
	assert.False(t, valid)
	assert.Contains(t, msg, "uppercase")
}

func TestValidatePassword_RequiresLowercase(t *testing.T) {
	valid, msg := ValidatePassword("UPPERCASE1!")
	assert.False(t, valid)
	assert.Contains(t, msg, "lowercase")
}

func TestValidatePassword_RequiresDigit(t *testing.T) {
	valid, msg := ValidatePassword("NoDigits!")
	assert.False(t, valid)
	assert.Contains(t, msg, "digit")
}

func TestValidatePassword_RequiresSpecialChar(t *testing.T) {
	valid, msg := ValidatePassword("NoSpecial1")
	assert.False(t, valid)
	assert.Contains(t, msg, "special")
}

func TestValidatePassword_AcceptsValidPassword(t *testing.T) {
	valid, msg := ValidatePassword("SecurePass123!")
	assert.True(t, valid)
	assert.Empty(t, msg)
}

// =====================================
// Email Validation Tests
// =====================================

func TestValidateEmail_AcceptsValidEmails(t *testing.T) {
	validEmails := []string{
		"test@example.com",
		"user.name@domain.org",
		"user+tag@example.co.uk",
	}

	for _, email := range validEmails {
		t.Run(email, func(t *testing.T) {
			assert.True(t, ValidateEmail(email))
		})
	}
}

func TestValidateEmail_RejectsInvalidEmails(t *testing.T) {
	invalidEmails := []string{
		"not-an-email",
		"missing@domain",
		"@nodomain.com",
		"spaces in@email.com",
	}

	for _, email := range invalidEmails {
		t.Run(email, func(t *testing.T) {
			assert.False(t, ValidateEmail(email))
		})
	}
}

// =====================================
// UUID Validation Tests
// =====================================

func TestValidateUUID_AcceptsValidUUIDs(t *testing.T) {
	validUUIDs := []string{
		"550e8400-e29b-41d4-a716-446655440000",
		"123e4567-e89b-12d3-a456-426614174000",
	}

	for _, uuid := range validUUIDs {
		t.Run(uuid, func(t *testing.T) {
			assert.True(t, ValidateUUID(uuid))
		})
	}
}

func TestValidateUUID_RejectsInvalidUUIDs(t *testing.T) {
	invalidUUIDs := []string{
		"not-a-uuid",
		"550e8400-e29b-41d4-a716",
		"550e8400e29b41d4a716446655440000", // No hyphens
	}

	for _, uuid := range invalidUUIDs {
		t.Run(uuid, func(t *testing.T) {
			assert.False(t, ValidateUUID(uuid))
		})
	}
}

// =====================================
// Audit Logging Tests
// =====================================

func TestAuditLogger_LogsEvent(t *testing.T) {
	logger := NewAuditLoggerWithConfig(AuditLoggerConfig{
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	})

	event := &AuditEvent{
		Timestamp:  time.Now(),
		EventType:  AuditEventLogin,
		Severity:   AuditSeverityInfo,
		UserID:     "user-123",
		Action:     "USER_LOGIN",
		Resource:   "/auth/login",
		IP:         "192.168.1.1",
		StatusCode: 200,
		Success:    true,
	}

	// Should not panic
	require.NotPanics(t, func() {
		logger.Log(event)
	})

	assert.Equal(t, "test-service", event.ServiceName)
}

func TestClassifyEvent_IdentifiesLoginEvent(t *testing.T) {
	eventType, severity := classifyEvent("POST", "/auth/login", 200)
	assert.Equal(t, AuditEventLogin, eventType)
	assert.Equal(t, AuditSeverityInfo, severity)
}

func TestClassifyEvent_IdentifiesFailedLogin(t *testing.T) {
	eventType, severity := classifyEvent("POST", "/auth/login", 401)
	assert.Equal(t, AuditEventLoginFailed, eventType)
	assert.Equal(t, AuditSeverityWarning, severity)
}

func TestClassifyEvent_IdentifiesTransferEvent(t *testing.T) {
	eventType, severity := classifyEvent("POST", "/api/v1/transfer", 200)
	assert.Equal(t, AuditEventTransferInit, eventType)
	assert.Equal(t, AuditSeverityInfo, severity)
}

func TestClassifyEvent_IdentifiesRateLimitExceeded(t *testing.T) {
	eventType, severity := classifyEvent("GET", "/api/v1/anything", 429)
	assert.Equal(t, AuditEventRateLimitExceeded, eventType)
	assert.Equal(t, AuditSeverityWarning, severity)
}

// =====================================
// Security Headers Tests
// =====================================

func TestSecurityHeaders_SetsAllHeaders(t *testing.T) {
	r := gin.New()
	r.Use(SecurityHeaders())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.NotEmpty(t, w.Header().Get("Content-Security-Policy"))
}

// =====================================
// Secure Compare Tests
// =====================================

func TestSecureCompare_MatchingStrings(t *testing.T) {
	assert.True(t, secureCompare("abc123", "abc123"))
}

func TestSecureCompare_DifferentStrings(t *testing.T) {
	assert.False(t, secureCompare("abc123", "xyz789"))
}

func TestSecureCompare_DifferentLengths(t *testing.T) {
	assert.False(t, secureCompare("short", "much longer string"))
}

// =====================================
// User Rate Limiter Tests
// =====================================

func TestUserRateLimiter_AllowsWithinLimit(t *testing.T) {
	limiter := NewUserRateLimiter(10, time.Minute)

	for i := 0; i < 10; i++ {
		assert.True(t, limiter.Allow("user-123"), "Request %d should be allowed", i)
	}
}

func TestUserRateLimiter_BlocksOverLimit(t *testing.T) {
	limiter := NewUserRateLimiter(5, time.Minute)

	// Use up all requests
	for i := 0; i < 5; i++ {
		limiter.Allow("user-456")
	}

	// Next request should be blocked
	assert.False(t, limiter.Allow("user-456"))
}

func TestUserRateLimiter_SeparatesUsers(t *testing.T) {
	limiter := NewUserRateLimiter(2, time.Minute)

	// User 1 uses their limit
	limiter.Allow("user-1")
	limiter.Allow("user-1")
	assert.False(t, limiter.Allow("user-1"))

	// User 2 should still have their limit
	assert.True(t, limiter.Allow("user-2"))
	assert.True(t, limiter.Allow("user-2"))
}

func TestUserRateLimiter_Remaining(t *testing.T) {
	limiter := NewUserRateLimiter(10, time.Minute)

	assert.Equal(t, 10, limiter.Remaining("new-user"))

	limiter.Allow("new-user")
	assert.Equal(t, 9, limiter.Remaining("new-user"))
}

// =====================================
// Integration Test: Full Request Flow
// =====================================

func TestFullSecurityMiddlewareStack(t *testing.T) {
	r := gin.New()

	// Apply all security middleware
	r.Use(SecurityHeaders())
	r.Use(RequestID())
	r.Use(InputValidation(nil))
	r.Use(RateLimit())

	r.GET("/secure", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "secure"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/secure", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify all security headers are present
	assert.NotEmpty(t, w.Header().Get("X-Frame-Options"))
	assert.NotEmpty(t, w.Header().Get("X-Content-Type-Options"))
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestBodyWithIdempotency(t *testing.T) {
	r := gin.New()
	store := NewInMemoryIdempotencyStore()
	config := DefaultIdempotencyConfig()
	r.Use(Idempotency(store, config))

	r.POST("/api/v1/transfer", func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.BindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok", "amount": body["amount"]})
	})

	payload := `{"amount": 100, "to": "account-456"}`

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/transfer", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Idempotency-Key", "transfer-key-1")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, float64(100), response["amount"])
}
