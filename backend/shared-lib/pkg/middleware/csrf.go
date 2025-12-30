package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// CSRFConfig holds CSRF protection configuration
type CSRFConfig struct {
	// TokenLength is the length of the CSRF token in bytes (will be hex encoded, so 32 bytes = 64 chars)
	TokenLength int
	// CookieName is the name of the cookie that stores the CSRF token
	CookieName string
	// HeaderName is the name of the header where the token should be sent
	HeaderName string
	// FormFieldName is the name of the form field for the token
	FormFieldName string
	// TokenExpiry is how long a token is valid
	TokenExpiry time.Duration
	// Secure sets the Secure flag on the cookie
	Secure bool
	// SameSite sets the SameSite attribute on the cookie
	SameSite http.SameSite
	// SkipPaths are paths that don't require CSRF protection
	SkipPaths []string
	// ExemptMethods are HTTP methods that don't require CSRF protection
	ExemptMethods []string
}

// DefaultCSRFConfig returns secure default CSRF configuration
func DefaultCSRFConfig() CSRFConfig {
	return CSRFConfig{
		TokenLength:   32,
		CookieName:    "csrf_token",
		HeaderName:    "X-CSRF-Token",
		FormFieldName: "_csrf",
		TokenExpiry:   24 * time.Hour,
		Secure:        true,
		SameSite:      http.SameSiteStrictMode,
		SkipPaths:     []string{"/health", "/metrics"},
		ExemptMethods: []string{"GET", "HEAD", "OPTIONS"},
	}
}

// CSRFTokenStore interface for storing CSRF tokens
type CSRFTokenStore interface {
	Get(token string) (bool, error)
	Set(token string, expiry time.Duration) error
	Delete(token string) error
}

// InMemoryCSRFStore is an in-memory CSRF token store
type InMemoryCSRFStore struct {
	tokens map[string]time.Time
	mu     sync.RWMutex
}

// NewInMemoryCSRFStore creates a new in-memory CSRF store
func NewInMemoryCSRFStore() *InMemoryCSRFStore {
	store := &InMemoryCSRFStore{
		tokens: make(map[string]time.Time),
	}
	go store.cleanup()
	return store
}

func (s *InMemoryCSRFStore) Get(token string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	expiry, exists := s.tokens[token]
	if !exists {
		return false, nil
	}
	if time.Now().After(expiry) {
		return false, nil
	}
	return true, nil
}

func (s *InMemoryCSRFStore) Set(token string, expiry time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[token] = time.Now().Add(expiry)
	return nil
}

func (s *InMemoryCSRFStore) Delete(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, token)
	return nil
}

func (s *InMemoryCSRFStore) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for token, expiry := range s.tokens {
			if now.After(expiry) {
				delete(s.tokens, token)
			}
		}
		s.mu.Unlock()
	}
}

// generateCSRFToken generates a cryptographically secure random token
func generateCSRFToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CSRFProtection returns a CSRF protection middleware
func CSRFProtection(config CSRFConfig, store CSRFTokenStore) gin.HandlerFunc {
	if store == nil {
		store = NewInMemoryCSRFStore()
	}

	return func(c *gin.Context) {
		// Check if path should be skipped
		for _, path := range config.SkipPaths {
			if strings.HasPrefix(c.Request.URL.Path, path) {
				c.Next()
				return
			}
		}

		// Check if method is exempt
		for _, method := range config.ExemptMethods {
			if c.Request.Method == method {
				// For GET requests, ensure a CSRF token is set
				ensureCSRFToken(c, config, store)
				c.Next()
				return
			}
		}

		// For non-exempt methods, validate CSRF token
		cookieToken, err := c.Cookie(config.CookieName)
		if err != nil || cookieToken == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "CSRF validation failed",
				"message": "Missing CSRF token cookie",
			})
			return
		}

		// Get token from header or form
		headerToken := c.GetHeader(config.HeaderName)
		formToken := c.PostForm(config.FormFieldName)

		submittedToken := headerToken
		if submittedToken == "" {
			submittedToken = formToken
		}

		if submittedToken == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "CSRF validation failed",
				"message": "Missing CSRF token in request",
			})
			return
		}

		// Constant-time comparison to prevent timing attacks
		if !secureCompare(cookieToken, submittedToken) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "CSRF validation failed",
				"message": "CSRF token mismatch",
			})
			return
		}

		// Verify token exists in store and is not expired
		valid, err := store.Get(cookieToken)
		if err != nil || !valid {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "CSRF validation failed",
				"message": "CSRF token expired or invalid",
			})
			return
		}

		// Token is valid, continue
		c.Next()
	}
}

// ensureCSRFToken ensures a CSRF token exists in the response
func ensureCSRFToken(c *gin.Context, config CSRFConfig, store CSRFTokenStore) {
	// Check if token already exists
	existingToken, err := c.Cookie(config.CookieName)
	if err == nil && existingToken != "" {
		// Verify token is still valid
		valid, _ := store.Get(existingToken)
		if valid {
			// Set token in context for templates
			c.Set("csrf_token", existingToken)
			return
		}
	}

	// Generate new token
	token, err := generateCSRFToken(config.TokenLength)
	if err != nil {
		return
	}

	// Store token
	store.Set(token, config.TokenExpiry)

	// Set cookie
	c.SetSameSite(config.SameSite)
	c.SetCookie(
		config.CookieName,
		token,
		int(config.TokenExpiry.Seconds()),
		"/",
		"",
		config.Secure,
		false, // Not HttpOnly - JS needs to read it
	)

	// Set token in context for templates
	c.Set("csrf_token", token)
}

// secureCompare performs constant-time string comparison
func secureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}

	return result == 0
}

// GetCSRFToken returns a handler that provides a new CSRF token
func GetCSRFToken(config CSRFConfig, store CSRFTokenStore) gin.HandlerFunc {
	if store == nil {
		store = NewInMemoryCSRFStore()
	}

	return func(c *gin.Context) {
		// Generate new token
		token, err := generateCSRFToken(config.TokenLength)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to generate CSRF token",
			})
			return
		}

		// Store token
		if err := store.Set(token, config.TokenExpiry); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to store CSRF token",
			})
			return
		}

		// Set cookie
		c.SetSameSite(config.SameSite)
		c.SetCookie(
			config.CookieName,
			token,
			int(config.TokenExpiry.Seconds()),
			"/",
			"",
			config.Secure,
			false,
		)

		c.JSON(http.StatusOK, gin.H{
			"csrf_token": token,
		})
	}
}

// ValidateCSRFToken validates a CSRF token without middleware context
func ValidateCSRFToken(store CSRFTokenStore, cookieToken, submittedToken string) error {
	if cookieToken == "" || submittedToken == "" {
		return errors.New("missing CSRF token")
	}

	if !secureCompare(cookieToken, submittedToken) {
		return errors.New("CSRF token mismatch")
	}

	valid, err := store.Get(cookieToken)
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("CSRF token expired or invalid")
	}

	return nil
}
