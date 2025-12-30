package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactPII_RedactsEmail(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"john.doe@example.com", "jo***@example.com"},
		{"Contact: user@domain.org for info", "Contact: us***@domain.org for info"},
		{"a@b.com", "a***@b.com"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := RedactPII(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRedactPII_RedactsCardNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"4111111111111111", "****-****-****-1111"},
		{"4111-1111-1111-1111", "****-****-****-1111"},
		{"4111 1111 1111 1111", "****-****-****-1111"},
		{"Card: 4532015112830366", "Card: ****-****-****-0366"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := RedactPII(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRedactPII_RedactsSSN(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"123-45-6789", "***-**-6789"},
		{"SSN: 987-65-4321", "SSN: ***-**-4321"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := RedactPII(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRedactPII_RedactsPhone(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"555-123-4567", "***-***-4567"},
		// These have surrounding characters that remain
		{"(555) 123-4567", "(***-***-4567"},
		{"+1-555-123-4567", "+1-***-***-4567"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := RedactPII(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRedactPII_RedactsPasswords(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`password: mysecret123`, "[REDACTED]"},
		// JSON format keeps the quotes
		{`"password":"secret"`, `"[REDACTED]"`},
		{`api_key=abc123def`, "[REDACTED]"},
		{`secret: verysecret`, "[REDACTED]"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := RedactPII(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRedactPII_RedactsBearerTokens(t *testing.T) {
	input := "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N"
	result := RedactPII(input)

	assert.Contains(t, result, "[REDACTED]")
	assert.NotContains(t, result, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9")
}

func TestRedactPII_PreservesNonSensitiveData(t *testing.T) {
	tests := []string{
		"Hello, World!",
		"User ID: 12345",
		"Status: active",
		"Error: not found",
		"https://example.com/page",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			result := RedactPII(input)
			assert.Equal(t, input, result)
		})
	}
}

func TestRedactPII_HandlesMultiplePIITypes(t *testing.T) {
	input := "User john@example.com called from 555-123-4567"
	result := RedactPII(input)

	assert.NotContains(t, result, "john@example.com")
	assert.NotContains(t, result, "555-123-4567")
	assert.Contains(t, result, "jo***@example.com")
	assert.Contains(t, result, "***-***-4567")
}

func TestSensitiveFields_ContainsExpectedFields(t *testing.T) {
	expectedFields := []string{
		"password",
		"secret",
		"token",
		"api_key",
		"authorization",
		"cvv",
		"ssn",
		"pin",
	}

	for _, field := range expectedFields {
		assert.True(t, SensitiveFields[field], "Expected %s to be a sensitive field", field)
	}
}

func TestRedactMatch_EmailPattern(t *testing.T) {
	result := redactMatch("email", "johndoe@example.com")
	assert.Equal(t, "jo***@example.com", result)
}

func TestRedactMatch_CardPattern(t *testing.T) {
	result := redactMatch("card_number", "4111111111111111")
	assert.Equal(t, "****-****-****-1111", result)
}

func TestRedactMatch_SSNPattern(t *testing.T) {
	result := redactMatch("ssn", "123-45-6789")
	assert.Equal(t, "***-**-6789", result)
}

func TestRedactMatch_PhonePattern(t *testing.T) {
	result := redactMatch("phone", "555-123-4567")
	assert.Equal(t, "***-***-4567", result)
}

func TestRedactMatch_PasswordPattern(t *testing.T) {
	result := redactMatch("password", "password=secret123")
	assert.Equal(t, "[REDACTED]", result)
}

func TestRedactMatch_UnknownPattern(t *testing.T) {
	result := redactMatch("unknown", "some data")
	assert.Equal(t, "[REDACTED]", result)
}

func TestInitLogger_DoesNotPanic(t *testing.T) {
	assert.NotPanics(t, func() {
		InitLogger("test-service", true)
	})
}

func TestInitLoggerWithoutRedaction_DoesNotPanic(t *testing.T) {
	assert.NotPanics(t, func() {
		InitLoggerWithoutRedaction("test-service", true)
	})
}

// Benchmark tests
func BenchmarkRedactPII_Email(b *testing.B) {
	input := "Contact john.doe@example.com for more information"
	for i := 0; i < b.N; i++ {
		RedactPII(input)
	}
}

func BenchmarkRedactPII_MultipleTypes(b *testing.B) {
	input := "User john@example.com with card 4111111111111111 called from 555-123-4567"
	for i := 0; i < b.N; i++ {
		RedactPII(input)
	}
}

func BenchmarkRedactPII_NoSensitiveData(b *testing.B) {
	input := "Regular log message without any sensitive data"
	for i := 0; i < b.N; i++ {
		RedactPII(input)
	}
}
