package security

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"
)

// DataMasker provides utilities for masking sensitive data
type DataMasker struct{}

// NewDataMasker creates a new data masker
func NewDataMasker() *DataMasker {
	return &DataMasker{}
}

// MaskCardNumber masks a credit card number, showing only last 4 digits
func (dm *DataMasker) MaskCardNumber(cardNumber string) string {
	cleaned := strings.ReplaceAll(cardNumber, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")

	if len(cleaned) < 4 {
		return "****"
	}

	masked := strings.Repeat("*", len(cleaned)-4) + cleaned[len(cleaned)-4:]
	return masked
}

// MaskEmail masks an email address, showing only first 2 chars and domain
func (dm *DataMasker) MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***@***"
	}

	local := parts[0]
	domain := parts[1]

	if len(local) <= 2 {
		return local + "***@" + domain
	}

	return local[:2] + strings.Repeat("*", len(local)-2) + "@" + domain
}

// MaskPhone masks a phone number, showing only last 4 digits
func (dm *DataMasker) MaskPhone(phone string) string {
	// Extract only digits
	var digits strings.Builder
	for _, r := range phone {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		}
	}

	d := digits.String()
	if len(d) < 4 {
		return "****"
	}

	return strings.Repeat("*", len(d)-4) + d[len(d)-4:]
}

// MaskSSN masks a Social Security Number
func (dm *DataMasker) MaskSSN(ssn string) string {
	cleaned := strings.ReplaceAll(ssn, "-", "")
	if len(cleaned) < 4 {
		return "***-**-****"
	}

	return "***-**-" + cleaned[len(cleaned)-4:]
}

// MaskAccountNumber masks an account number, showing only last 4 digits
func (dm *DataMasker) MaskAccountNumber(accountNumber string) string {
	if len(accountNumber) <= 4 {
		return accountNumber
	}

	return strings.Repeat("*", len(accountNumber)-4) + accountNumber[len(accountNumber)-4:]
}

// CSRFToken represents a CSRF token
type CSRFToken struct {
	Token     string
	ExpiresAt time.Time
}

// CSRFProtection provides CSRF token management
type CSRFProtection struct {
	tokens      map[string]*CSRFToken
	mu          sync.RWMutex
	tokenExpiry time.Duration
}

// NewCSRFProtection creates a new CSRF protection manager
func NewCSRFProtection(tokenExpiry time.Duration) *CSRFProtection {
	return &CSRFProtection{
		tokens:      make(map[string]*CSRFToken),
		tokenExpiry: tokenExpiry,
	}
}

// GenerateToken generates a new CSRF token for a session
func (cp *CSRFProtection) GenerateToken(sessionID string) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	token := hex.EncodeToString(bytes)

	cp.mu.Lock()
	defer cp.mu.Unlock()

	cp.tokens[sessionID] = &CSRFToken{
		Token:     token,
		ExpiresAt: time.Now().Add(cp.tokenExpiry),
	}

	return token, nil
}

// ValidateToken validates a CSRF token
func (cp *CSRFProtection) ValidateToken(sessionID, token string) error {
	cp.mu.RLock()
	defer cp.mu.RUnlock()

	stored, exists := cp.tokens[sessionID]
	if !exists {
		return errors.New("no CSRF token found for session")
	}

	if time.Now().After(stored.ExpiresAt) {
		return errors.New("CSRF token expired")
	}

	if stored.Token != token {
		return errors.New("invalid CSRF token")
	}

	return nil
}

// Cleanup removes expired tokens
func (cp *CSRFProtection) Cleanup() {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	now := time.Now()
	for sessionID, token := range cp.tokens {
		if now.After(token.ExpiresAt) {
			delete(cp.tokens, sessionID)
		}
	}
}

// SecureCompare performs a constant-time string comparison
// to prevent timing attacks
func SecureCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}

	return result == 0
}

// SanitizeForLog removes sensitive data from strings before logging
func SanitizeForLog(input string) string {
	masker := NewDataMasker()

	// Simple pattern matching for common sensitive data
	// In production, use more sophisticated detection

	// Mask card numbers (16 digits)
	if len(input) >= 16 {
		input = masker.MaskCardNumber(input)
	}

	// Mask SSN patterns
	if strings.Contains(input, "-") && len(input) == 11 {
		input = masker.MaskSSN(input)
	}

	return input
}
