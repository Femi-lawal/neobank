package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// UserRateLimiter implements per-user rate limiting
type UserRateLimiter struct {
	limits      map[string]*userLimit
	mu          sync.RWMutex
	maxRequests int
	window      time.Duration
}

type userLimit struct {
	requests    int
	windowStart time.Time
}

// NewUserRateLimiter creates a new per-user rate limiter
func NewUserRateLimiter(maxRequests int, window time.Duration) *UserRateLimiter {
	return &UserRateLimiter{
		limits:      make(map[string]*userLimit),
		maxRequests: maxRequests,
		window:      window,
	}
}

// DefaultUserRateLimiter returns a rate limiter with secure defaults
func DefaultUserRateLimiter() *UserRateLimiter {
	return NewUserRateLimiter(100, time.Minute) // 100 requests per minute per user
}

// Allow checks if a user is allowed to make a request
func (rl *UserRateLimiter) Allow(userID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	limit, exists := rl.limits[userID]

	if !exists {
		rl.limits[userID] = &userLimit{
			requests:    1,
			windowStart: now,
		}
		return true
	}

	// Reset window if expired
	if now.Sub(limit.windowStart) > rl.window {
		limit.requests = 1
		limit.windowStart = now
		return true
	}

	// Check limit
	if limit.requests >= rl.maxRequests {
		return false
	}

	limit.requests++
	return true
}

// Remaining returns the number of remaining requests for a user
func (rl *UserRateLimiter) Remaining(userID string) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	limit, exists := rl.limits[userID]
	if !exists {
		return rl.maxRequests
	}

	// Check if window expired
	if time.Since(limit.windowStart) > rl.window {
		return rl.maxRequests
	}

	remaining := rl.maxRequests - limit.requests
	if remaining < 0 {
		return 0
	}
	return remaining
}

// UserRateLimitMiddleware creates a Gin middleware for per-user rate limiting
func UserRateLimitMiddleware(limiter *UserRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by auth middleware)
		userID := c.GetString("userID")

		// Fall back to IP if no user ID
		if userID == "" {
			userID = "ip:" + c.ClientIP()
		}

		if !limiter.Allow(userID) {
			c.Header("X-RateLimit-Limit", string(rune(limiter.maxRequests)))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", "60")

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": "Too many requests. Please try again later.",
			})
			return
		}

		// Add rate limit headers
		remaining := limiter.Remaining(userID)
		c.Header("X-RateLimit-Limit", string(rune(limiter.maxRequests)))
		c.Header("X-RateLimit-Remaining", string(rune(remaining)))

		c.Next()
	}
}

// Cleanup removes expired entries
func (rl *UserRateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for userID, limit := range rl.limits {
		if now.Sub(limit.windowStart) > rl.window {
			delete(rl.limits, userID)
		}
	}
}

// SensitiveEndpointRateLimiter applies stricter limits to sensitive endpoints
type SensitiveEndpointRateLimiter struct {
	limiter *UserRateLimiter
}

// NewSensitiveEndpointRateLimiter creates a rate limiter for sensitive endpoints
func NewSensitiveEndpointRateLimiter() *SensitiveEndpointRateLimiter {
	return &SensitiveEndpointRateLimiter{
		limiter: NewUserRateLimiter(10, time.Minute), // 10 requests per minute
	}
}

// SensitiveRateLimitMiddleware applies strict rate limiting to sensitive endpoints
func SensitiveRateLimitMiddleware(limiter *SensitiveEndpointRateLimiter) gin.HandlerFunc {
	sensitiveEndpoints := map[string]bool{
		"/login":           true,
		"/register":        true,
		"/password/reset":  true,
		"/password/change": true,
		"/transfer":        true,
	}

	return func(c *gin.Context) {
		// Only apply to sensitive endpoints
		if !sensitiveEndpoints[c.FullPath()] {
			c.Next()
			return
		}

		userID := c.GetString("userID")
		if userID == "" {
			userID = "ip:" + c.ClientIP()
		}

		if !limiter.limiter.Allow(userID) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many attempts. Please wait before trying again.",
			})
			return
		}

		c.Next()
	}
}
