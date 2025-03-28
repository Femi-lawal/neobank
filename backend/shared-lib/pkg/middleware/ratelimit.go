package middleware

import (
	"strconv"
	"sync"
	"time"

	"github.com/femi-lawal/new_bank/backend/shared-lib/pkg/errors"
	"github.com/gin-gonic/gin"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	BurstSize         int
	CleanupInterval   time.Duration
}

// DefaultRateLimitConfig returns default rate limiting settings
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerMinute: 60,
		BurstSize:         10,
		CleanupInterval:   5 * time.Minute,
	}
}

// clientInfo tracks request counts for a client
type clientInfo struct {
	tokens    float64
	lastCheck time.Time
}

// rateLimiter implements a token bucket rate limiter
type rateLimiter struct {
	clients map[string]*clientInfo
	mu      sync.RWMutex
	config  RateLimitConfig
	rate    float64 // tokens per second
}

// newRateLimiter creates a new rate limiter
func newRateLimiter(config RateLimitConfig) *rateLimiter {
	rl := &rateLimiter{
		clients: make(map[string]*clientInfo),
		config:  config,
		rate:    float64(config.RequestsPerMinute) / 60.0,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// allow checks if a request is allowed for the given key
func (rl *rateLimiter) allow(key string) (bool, int, int, time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	client, exists := rl.clients[key]

	if !exists {
		client = &clientInfo{
			tokens:    float64(rl.config.BurstSize),
			lastCheck: now,
		}
		rl.clients[key] = client
	}

	// Calculate tokens to add based on time elapsed
	elapsed := now.Sub(client.lastCheck).Seconds()
	client.tokens += elapsed * rl.rate
	client.lastCheck = now

	// Cap tokens at burst size
	if client.tokens > float64(rl.config.BurstSize) {
		client.tokens = float64(rl.config.BurstSize)
	}

	// Check if request is allowed
	remaining := int(client.tokens)
	resetTime := now.Add(time.Duration(float64(time.Minute) / rl.rate))

	if client.tokens < 1 {
		return false, rl.config.RequestsPerMinute, 0, resetTime
	}

	client.tokens--
	return true, rl.config.RequestsPerMinute, remaining - 1, resetTime
}

// cleanup removes old entries periodically
func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(rl.config.CleanupInterval)
	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-rl.config.CleanupInterval)
		for key, client := range rl.clients {
			if client.lastCheck.Before(cutoff) {
				delete(rl.clients, key)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit returns a rate limiting middleware with default config
func RateLimit() gin.HandlerFunc {
	return RateLimitWithConfig(DefaultRateLimitConfig())
}

// RateLimitWithConfig returns a rate limiting middleware with custom config
func RateLimitWithConfig(config RateLimitConfig) gin.HandlerFunc {
	limiter := newRateLimiter(config)

	return func(c *gin.Context) {
		// Use client IP as the rate limit key
		key := c.ClientIP()

		// Check if authenticated and use user ID instead
		if userID := GetUserID(c); userID != "" {
			key = "user:" + userID
		}

		allowed, limit, remaining, resetTime := limiter.allow(key)

		// Set rate limit headers (BE-006: use strconv.Itoa for proper int-to-string conversion)
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", resetTime.Format(time.RFC3339))

		if !allowed {
			c.Header("Retry-After", "60")
			errors.RespondWithError(c, errors.ErrRateLimited)
			return
		}

		c.Next()
	}
}

// KeyFunc is a function that returns the rate limit key for a request
type KeyFunc func(c *gin.Context) string

// RateLimitWithKeyFunc returns a rate limiting middleware with custom key function
func RateLimitWithKeyFunc(config RateLimitConfig, keyFunc KeyFunc) gin.HandlerFunc {
	limiter := newRateLimiter(config)

	return func(c *gin.Context) {
		key := keyFunc(c)
		allowed, limit, remaining, resetTime := limiter.allow(key)

		// Set rate limit headers (BE-006: use strconv.Itoa for proper int-to-string conversion)
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", resetTime.Format(time.RFC3339))

		if !allowed {
			c.Header("Retry-After", "60")
			errors.RespondWithError(c, errors.ErrRateLimited)
			return
		}

		c.Next()
	}
}
