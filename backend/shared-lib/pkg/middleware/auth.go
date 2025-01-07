package middleware

import (
	"log/slog"
	"strings"

	"github.com/femi-lawal/new_bank/backend/shared-lib/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// ContextKey is the type for context keys
type ContextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
	// EmailKey is the context key for email
	EmailKey ContextKey = "email"
	// ClaimsKey is the context key for full claims
	ClaimsKey ContextKey = "claims"
)

// JWTAuthConfig holds configuration for the JWT middleware
type JWTAuthConfig struct {
	SecretKey    string
	TokenLookup  string // "header:Authorization" or "cookie:token"
	TokenPrefix  string // "Bearer "
	SkipPaths    []string
	ErrorHandler func(*gin.Context, error)
}

// DefaultJWTConfig returns a default JWT configuration
func DefaultJWTConfig(secretKey string) JWTAuthConfig {
	return JWTAuthConfig{
		SecretKey:   secretKey,
		TokenLookup: "header:Authorization",
		TokenPrefix: "Bearer ",
		SkipPaths:   []string{"/health", "/metrics", "/auth/login", "/auth/register"},
	}
}

// JWTAuth returns a JWT authentication middleware
func JWTAuth(secretKey string) gin.HandlerFunc {
	return JWTAuthWithConfig(DefaultJWTConfig(secretKey))
}

// JWTAuthWithConfig returns a JWT middleware with custom config
func JWTAuthWithConfig(config JWTAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth for certain paths
		for _, path := range config.SkipPaths {
			if strings.HasPrefix(c.Request.URL.Path, path) {
				c.Next()
				return
			}
		}

		// Extract token
		tokenString := extractToken(c, config)
		if tokenString == "" {
			slog.Debug("No token found in request", "path", c.Request.URL.Path)
			errors.RespondWithError(c, errors.ErrUnauthorized)
			return
		}

		// Parse and validate token
		claims, err := validateToken(tokenString, config.SecretKey)
		if err != nil {
			slog.Debug("Invalid token", "error", err.Error())
			errors.RespondWithError(c, errors.ErrInvalidToken)
			return
		}

		// Set user info in context
		c.Set(string(UserIDKey), claims.UserID)
		c.Set(string(EmailKey), claims.Email)
		c.Set(string(ClaimsKey), claims)

		slog.Debug("Authenticated request", "user_id", claims.UserID, "path", c.Request.URL.Path)
		c.Next()
	}
}

// extractToken extracts the JWT token from the request
func extractToken(c *gin.Context, config JWTAuthConfig) string {
	parts := strings.Split(config.TokenLookup, ":")
	if len(parts) != 2 {
		return ""
	}

	switch parts[0] {
	case "header":
		header := c.GetHeader(parts[1])
		if header == "" {
			return ""
		}
		if config.TokenPrefix != "" && strings.HasPrefix(header, config.TokenPrefix) {
			return strings.TrimPrefix(header, config.TokenPrefix)
		}
		return header

	case "cookie":
		cookie, err := c.Cookie(parts[1])
		if err != nil {
			return ""
		}
		return cookie

	case "query":
		return c.Query(parts[1])
	}

	return ""
}

// validateToken parses and validates a JWT token
func validateToken(tokenString string, secretKey string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// GetUserID retrieves the user ID from the context
func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get(string(UserIDKey)); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// GetEmail retrieves the email from the context
func GetEmail(c *gin.Context) string {
	if email, exists := c.Get(string(EmailKey)); exists {
		if e, ok := email.(string); ok {
			return e
		}
	}
	return ""
}

// GetClaims retrieves the full claims from the context
func GetClaims(c *gin.Context) *Claims {
	if claims, exists := c.Get(string(ClaimsKey)); exists {
		if c, ok := claims.(*Claims); ok {
			return c
		}
	}
	return nil
}

// OptionalAuth is similar to JWTAuth but doesn't reject unauthenticated requests
func OptionalAuth(secretKey string) gin.HandlerFunc {
	config := DefaultJWTConfig(secretKey)
	return func(c *gin.Context) {
		tokenString := extractToken(c, config)
		if tokenString != "" {
			claims, err := validateToken(tokenString, config.SecretKey)
			if err == nil {
				c.Set(string(UserIDKey), claims.UserID)
				c.Set(string(EmailKey), claims.Email)
				c.Set(string(ClaimsKey), claims)
			}
		}
		c.Next()
	}
}
