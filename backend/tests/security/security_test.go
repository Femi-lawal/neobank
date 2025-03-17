package security_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPasswordValidation tests password strength requirements
func TestPasswordValidation(t *testing.T) {
	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{"empty password", "", false},
		{"too short", "Short1!", false},
		{"no uppercase", "lowercase1!", false},
		{"no lowercase", "UPPERCASE1!", false},
		{"no digit", "NoDigits!", false},
		{"no special char", "NoSpecial1", false},
		{"valid password", "SecurePass123!", true},
		{"long valid password", "ThisIsAVeryLongSecurePassword123!@#", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := validatePassword(tt.password)
			assert.Equal(t, tt.valid, valid, "password: %s", tt.password)
		})
	}
}

// TestSQLInjectionPrevention tests SQL injection patterns are blocked
func TestSQLInjectionPrevention(t *testing.T) {
	sqlInjectionPatterns := []string{
		"'; DROP TABLE users; --",
		"1 OR 1=1",
		"1; DELETE FROM users",
		"UNION SELECT * FROM users",
		"'; INSERT INTO users VALUES ('hacker', 'password'); --",
	}

	for _, pattern := range sqlInjectionPatterns {
		t.Run(pattern, func(t *testing.T) {
			assert.True(t, isMalicious(pattern), "should detect SQL injection: %s", pattern)
		})
	}
}

// TestXSSPrevention tests XSS patterns are blocked
func TestXSSPrevention(t *testing.T) {
	xssPatterns := []string{
		"<script>alert('XSS')</script>",
		"javascript:alert('XSS')",
		"<img src=x onerror=alert('XSS')>",
		"<body onload=alert('XSS')>",
		"<iframe src='javascript:alert(1)'>",
	}

	for _, pattern := range xssPatterns {
		t.Run(pattern, func(t *testing.T) {
			assert.True(t, isMalicious(pattern), "should detect XSS: %s", pattern)
		})
	}
}

// TestCardNumberMasking tests card numbers are properly masked
func TestCardNumberMasking(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"4111111111111111", "************1111"},
		{"4111-1111-1111-1111", "************1111"},
		{"4111 1111 1111 1111", "************1111"},
		{"1234", "1234"},
		{"", "****"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			masked := maskCardNumber(tt.input)
			assert.Equal(t, tt.expected, masked)
		})
	}
}

// TestEmailMasking tests emails are properly masked
func TestEmailMasking(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"john.doe@example.com", "jo******@example.com"},
		{"ab@test.com", "ab***@test.com"},
		{"a@b.com", "a***@b.com"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			masked := maskEmail(tt.input)
			assert.Equal(t, tt.expected, masked)
		})
	}
}

// TestAccountLockout tests account lockout functionality
func TestAccountLockout(t *testing.T) {
	t.Run("should lock after max attempts", func(t *testing.T) {
		// Simulate 5 failed attempts
		userID := "test-user"
		for i := 0; i < 5; i++ {
			locked := recordFailedAttempt(userID)
			if i < 4 {
				assert.False(t, locked, "should not be locked before 5 attempts")
			} else {
				assert.True(t, locked, "should be locked after 5 attempts")
			}
		}
	})

	t.Run("should clear on successful login", func(t *testing.T) {
		userID := "test-user-2"
		recordFailedAttempt(userID)
		recordFailedAttempt(userID)
		recordSuccessfulLogin(userID)

		locked := isAccountLocked(userID)
		assert.False(t, locked, "should not be locked after successful login")
	})
}

// TestCSRFTokenValidation tests CSRF token validation
func TestCSRFTokenValidation(t *testing.T) {
	t.Run("should validate correct token", func(t *testing.T) {
		sessionID := "session-123"
		token := generateCSRFToken(sessionID)

		err := validateCSRFToken(sessionID, token)
		assert.NoError(t, err)
	})

	t.Run("should reject wrong token", func(t *testing.T) {
		sessionID := "session-456"
		generateCSRFToken(sessionID)

		err := validateCSRFToken(sessionID, "wrong-token")
		assert.Error(t, err)
	})

	t.Run("should reject missing session", func(t *testing.T) {
		err := validateCSRFToken("nonexistent", "any-token")
		assert.Error(t, err)
	})
}

// Helper functions (these would be imported from actual packages)
func validatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasDigit = true
		case c == '!' || c == '@' || c == '#' || c == '$' || c == '%' || c == '^' || c == '&' || c == '*':
			hasSpecial = true
		}
	}
	return hasUpper && hasLower && hasDigit && hasSpecial
}

func isMalicious(input string) bool {
	for _, r := range input {
		if r == '\'' || r == '-' || r == ';' || r == '<' || r == '>' {
			return true
		}
	}
	return false
}

func maskCardNumber(card string) string {
	if len(card) < 4 {
		return "****"
	}
	return "************" + card[len(card)-4:]
}

func maskEmail(email string) string {
	for i, c := range email {
		if c == '@' {
			if i <= 2 {
				return email[:i] + "***" + email[i:]
			}
			return email[:2] + "******" + email[i:]
		}
	}
	return "***@***"
}

var lockouts = make(map[string]int)
var locked = make(map[string]bool)

func recordFailedAttempt(userID string) bool {
	lockouts[userID]++
	if lockouts[userID] >= 5 {
		locked[userID] = true
		return true
	}
	return false
}

func recordSuccessfulLogin(userID string) {
	delete(lockouts, userID)
	delete(locked, userID)
}

func isAccountLocked(userID string) bool {
	return locked[userID]
}

var csrfTokens = make(map[string]string)

func generateCSRFToken(sessionID string) string {
	token := "csrf-" + sessionID
	csrfTokens[sessionID] = token
	return token
}

func validateCSRFToken(sessionID, token string) error {
	stored, exists := csrfTokens[sessionID]
	if !exists {
		return assert.AnError
	}
	if stored != token {
		return assert.AnError
	}
	return nil
}
