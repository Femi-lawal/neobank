package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// IdempotencyConfig holds configuration for idempotency
type IdempotencyConfig struct {
	// Header name for the idempotency key
	HeaderName string
	// How long to store idempotency keys
	TTL time.Duration
	// Required endpoints (paths that must have idempotency keys)
	RequiredPaths []string
	// Optional endpoints (paths where idempotency is encouraged but not required)
	OptionalPaths []string
}

// DefaultIdempotencyConfig returns default configuration
func DefaultIdempotencyConfig() IdempotencyConfig {
	return IdempotencyConfig{
		HeaderName: "X-Idempotency-Key",
		TTL:        24 * time.Hour,
		RequiredPaths: []string{
			"/api/v1/transfer",
			"/api/v1/payment",
			"/api/v1/cards/issue",
		},
		OptionalPaths: []string{
			"/api/v1/accounts",
		},
	}
}

// IdempotencyRecord stores the result of an idempotent operation
type IdempotencyRecord struct {
	Key           string
	RequestHash   string
	StatusCode    int
	ResponseBody  []byte
	ResponseHeaders map[string]string
	CreatedAt     time.Time
	ExpiresAt     time.Time
	UserID        string
}

// IdempotencyStore interface for storing idempotency records
type IdempotencyStore interface {
	Get(key string) (*IdempotencyRecord, bool)
	Set(key string, record *IdempotencyRecord)
	Delete(key string)
}

// InMemoryIdempotencyStore is an in-memory implementation (use Redis in production)
type InMemoryIdempotencyStore struct {
	records map[string]*IdempotencyRecord
	mu      sync.RWMutex
}

// NewInMemoryIdempotencyStore creates a new in-memory store
func NewInMemoryIdempotencyStore() *InMemoryIdempotencyStore {
	store := &InMemoryIdempotencyStore{
		records: make(map[string]*IdempotencyRecord),
	}
	// Start cleanup goroutine
	go store.cleanup()
	return store
}

func (s *InMemoryIdempotencyStore) Get(key string) (*IdempotencyRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, exists := s.records[key]
	if !exists {
		return nil, false
	}
	if time.Now().After(record.ExpiresAt) {
		return nil, false
	}
	return record, true
}

func (s *InMemoryIdempotencyStore) Set(key string, record *IdempotencyRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[key] = record
}

func (s *InMemoryIdempotencyStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.records, key)
}

func (s *InMemoryIdempotencyStore) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for key, record := range s.records {
			if now.After(record.ExpiresAt) {
				delete(s.records, key)
			}
		}
		s.mu.Unlock()
	}
}

// responseRecorder captures the response
type responseRecorder struct {
	gin.ResponseWriter
	body       []byte
	statusCode int
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return r.ResponseWriter.Write(b)
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

// Idempotency middleware ensures idempotent operations
func Idempotency(store IdempotencyStore, config IdempotencyConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply to mutating methods
		if c.Request.Method != http.MethodPost && c.Request.Method != http.MethodPut && c.Request.Method != http.MethodPatch {
			c.Next()
			return
		}

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		isRequired := isPathInList(path, config.RequiredPaths)
		isOptional := isPathInList(path, config.OptionalPaths)

		// Get idempotency key from header
		idempotencyKey := c.GetHeader(config.HeaderName)

		// If path requires idempotency key but none provided
		if isRequired && idempotencyKey == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   "Idempotency key required",
				"message": fmt.Sprintf("Please provide a unique %s header for this operation", config.HeaderName),
			})
			return
		}

		// If no idempotency key and not required, continue normally
		if idempotencyKey == "" && !isRequired {
			c.Next()
			return
		}

		// Only process idempotency for required or optional paths
		if !isRequired && !isOptional {
			c.Next()
			return
		}

		// Get user ID for key scoping
		userID := c.GetString(string(UserIDKey))
		scopedKey := fmt.Sprintf("%s:%s", userID, idempotencyKey)

		// Calculate request hash to detect conflicting requests
		requestHash := hashRequest(c)

		// Check for existing record
		if record, exists := store.Get(scopedKey); exists {
			// Verify request is identical
			if record.RequestHash != requestHash {
				c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{
					"error":   "Idempotency key conflict",
					"message": "The idempotency key has been used with different request parameters",
				})
				return
			}

			// Return cached response
			for k, v := range record.ResponseHeaders {
				c.Header(k, v)
			}
			c.Header("X-Idempotent-Replayed", "true")
			c.Data(record.StatusCode, "application/json", record.ResponseBody)
			c.Abort()
			return
		}

		// Record the response
		recorder := &responseRecorder{ResponseWriter: c.Writer, statusCode: 200}
		c.Writer = recorder

		c.Next()

		// Only store successful responses (2xx and some 4xx)
		if recorder.statusCode >= 200 && recorder.statusCode < 500 {
			headers := make(map[string]string)
			for k, v := range recorder.Header() {
				if len(v) > 0 {
					headers[k] = v[0]
				}
			}

			record := &IdempotencyRecord{
				Key:             scopedKey,
				RequestHash:     requestHash,
				StatusCode:      recorder.statusCode,
				ResponseBody:    recorder.body,
				ResponseHeaders: headers,
				CreatedAt:       time.Now(),
				ExpiresAt:       time.Now().Add(config.TTL),
				UserID:          userID,
			}
			store.Set(scopedKey, record)
		}
	}
}

func isPathInList(path string, list []string) bool {
	for _, p := range list {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

func hashRequest(c *gin.Context) string {
	h := sha256.New()
	h.Write([]byte(c.Request.Method))
	h.Write([]byte(c.Request.URL.Path))
	h.Write([]byte(c.Request.URL.RawQuery))
	
	// Hash body if present
	if c.Request.Body != nil {
		body, _ := io.ReadAll(c.Request.Body)
		h.Write(body)
		// Reset body for further processing
		c.Request.Body = io.NopCloser(strings.NewReader(string(body)))
	}
	
	return hex.EncodeToString(h.Sum(nil))
}
