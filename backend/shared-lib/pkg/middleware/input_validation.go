package middleware

import (
	"net/http"
	"regexp"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
)

// InputValidationConfig configures input validation rules
type InputValidationConfig struct {
	MaxBodySize     int64
	MaxURLLength    int
	MaxHeaderSize   int
	AllowedMethods  []string
	BlockedPatterns []string
}

// DefaultValidationConfig returns secure defaults
func DefaultValidationConfig() *InputValidationConfig {
	return &InputValidationConfig{
		MaxBodySize:   1 * 1024 * 1024, // 1MB
		MaxURLLength:  2048,
		MaxHeaderSize: 8192,
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		BlockedPatterns: []string{
			`(?i)<script`,
			`(?i)javascript:`,
			`(?i)on\w+\s*=`,
			`(?i)data:text/html`,
			`(?i)vbscript:`,
			`\/\*[\s\S]*?\*\/`,
			`--`,
			`;--`,
			`(?i)UNION\s+SELECT`,
			`(?i)INSERT\s+INTO`,
			`(?i)DROP\s+TABLE`,
			`(?i)DELETE\s+FROM`,
			`(?i)UPDATE\s+.*SET`,
			`(?i)exec\s*\(`,
			`(?i)execute\s*\(`,
		},
	}
}

// InputValidation middleware validates and sanitizes all inputs
func InputValidation(config *InputValidationConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultValidationConfig()
	}

	// Compile regex patterns
	var blockedRegex []*regexp.Regexp
	for _, pattern := range config.BlockedPatterns {
		if re, err := regexp.Compile(pattern); err == nil {
			blockedRegex = append(blockedRegex, re)
		}
	}

	return func(c *gin.Context) {
		// Check HTTP method
		if !isAllowedMethod(c.Request.Method, config.AllowedMethods) {
			c.AbortWithStatusJSON(http.StatusMethodNotAllowed, gin.H{
				"error": "Method not allowed",
			})
			return
		}

		// Check URL length
		if len(c.Request.URL.String()) > config.MaxURLLength {
			c.AbortWithStatusJSON(http.StatusRequestURITooLong, gin.H{
				"error": "URL too long",
			})
			return
		}

		// Check for malicious patterns in URL
		if containsMaliciousPattern(c.Request.URL.String(), blockedRegex) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request",
			})
			return
		}

		// Check query parameters
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if containsMaliciousPattern(key, blockedRegex) || containsMaliciousPattern(value, blockedRegex) {
					c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
						"error": "Invalid query parameter",
					})
					return
				}
			}
		}

		// Check headers
		for key, values := range c.Request.Header {
			for _, value := range values {
				if containsMaliciousPattern(key, blockedRegex) || containsMaliciousPattern(value, blockedRegex) {
					c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
						"error": "Invalid header",
					})
					return
				}
			}
		}

		// Limit body size
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, config.MaxBodySize)

		c.Next()
	}
}

func isAllowedMethod(method string, allowed []string) bool {
	for _, m := range allowed {
		if m == method {
			return true
		}
	}
	return false
}

func containsMaliciousPattern(input string, patterns []*regexp.Regexp) bool {
	for _, re := range patterns {
		if re.MatchString(input) {
			return true
		}
	}
	return false
}

// SanitizeString removes potentially dangerous characters
func SanitizeString(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Remove control characters
	var cleaned strings.Builder
	for _, r := range input {
		if !unicode.IsControl(r) || r == '\n' || r == '\r' || r == '\t' {
			cleaned.WriteRune(r)
		}
	}

	return strings.TrimSpace(cleaned.String())
}

// ValidateEmail validates email format
func ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email) && len(email) <= 254
}

// ValidateUUID validates UUID format
func ValidateUUID(uuid string) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	return uuidRegex.MatchString(uuid)
}

// ValidatePassword checks password strength
func ValidatePassword(password string) (bool, string) {
	if len(password) < 8 {
		return false, "Password must be at least 8 characters"
	}
	if len(password) > 128 {
		return false, "Password must not exceed 128 characters"
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return false, "Password must contain at least one uppercase letter"
	}
	if !hasLower {
		return false, "Password must contain at least one lowercase letter"
	}
	if !hasDigit {
		return false, "Password must contain at least one digit"
	}
	if !hasSpecial {
		return false, "Password must contain at least one special character"
	}

	return true, ""
}
