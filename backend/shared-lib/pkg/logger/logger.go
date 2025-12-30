package logger

import (
	"context"
	"log/slog"
	"os"
	"regexp"
	"strings"
)

var Log *slog.Logger

// PIIPatterns contains regex patterns for PII detection
var PIIPatterns = map[string]*regexp.Regexp{
	"email":       regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
	"card_number": regexp.MustCompile(`\b(?:\d{4}[-\s]?){3}\d{4}\b`),
	"ssn":         regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`),
	"phone":       regexp.MustCompile(`\b(?:\+1[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}\b`),
	"password":    regexp.MustCompile(`(?i)password["\s:=]+[^\s,}"']+`),
	"api_key":     regexp.MustCompile(`(?i)api[_-]?key["\s:=]+[^\s,}"']+`),
	"secret":      regexp.MustCompile(`(?i)secret["\s:=]+[^\s,}"']+`),
	"token":       regexp.MustCompile(`(?i)token["\s:=]+[^\s,}"']+`),
	"bearer":      regexp.MustCompile(`(?i)bearer\s+[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+`),
}

// SensitiveFields lists field names that should always be redacted
var SensitiveFields = map[string]bool{
	"password":        true,
	"pass":            true,
	"pwd":             true,
	"secret":          true,
	"token":           true,
	"api_key":         true,
	"apikey":          true,
	"authorization":   true,
	"auth":            true,
	"credit_card":     true,
	"card_number":     true,
	"cvv":             true,
	"ssn":             true,
	"social_security": true,
	"pin":             true,
}

// PIIRedactingHandler wraps slog.Handler to redact PII
type PIIRedactingHandler struct {
	slog.Handler
}

func (h *PIIRedactingHandler) Handle(ctx context.Context, r slog.Record) error {
	// Create a new record with redacted attributes
	newRecord := slog.NewRecord(r.Time, r.Level, r.Message, r.PC)

	r.Attrs(func(a slog.Attr) bool {
		newRecord.AddAttrs(redactAttr(a))
		return true
	})

	return h.Handler.Handle(ctx, newRecord)
}

func (h *PIIRedactingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	redactedAttrs := make([]slog.Attr, len(attrs))
	for i, attr := range attrs {
		redactedAttrs[i] = redactAttr(attr)
	}
	return &PIIRedactingHandler{h.Handler.WithAttrs(redactedAttrs)}
}

func (h *PIIRedactingHandler) WithGroup(name string) slog.Handler {
	return &PIIRedactingHandler{h.Handler.WithGroup(name)}
}

// redactAttr redacts sensitive data from a slog.Attr
func redactAttr(a slog.Attr) slog.Attr {
	key := strings.ToLower(a.Key)

	// Check if field name is sensitive
	if SensitiveFields[key] {
		return slog.String(a.Key, "[REDACTED]")
	}

	// For string values, check for PII patterns
	if a.Value.Kind() == slog.KindString {
		value := a.Value.String()
		redacted := RedactPII(value)
		if redacted != value {
			return slog.String(a.Key, redacted)
		}
	}

	return a
}

// RedactPII redacts personally identifiable information from a string
func RedactPII(input string) string {
	result := input

	// Redact each pattern
	for name, pattern := range PIIPatterns {
		result = pattern.ReplaceAllStringFunc(result, func(match string) string {
			return redactMatch(name, match)
		})
	}

	return result
}

// redactMatch returns a redacted version of the matched PII
func redactMatch(piiType, match string) string {
	switch piiType {
	case "email":
		// Keep first 2 chars and domain
		parts := strings.Split(match, "@")
		if len(parts) == 2 {
			local := parts[0]
			if len(local) > 2 {
				return local[:2] + "***@" + parts[1]
			}
			return local + "***@" + parts[1]
		}
		return "[REDACTED_EMAIL]"
	case "card_number":
		// Show last 4 digits only
		cleaned := strings.ReplaceAll(strings.ReplaceAll(match, "-", ""), " ", "")
		if len(cleaned) >= 4 {
			return "****-****-****-" + cleaned[len(cleaned)-4:]
		}
		return "[REDACTED_CARD]"
	case "ssn":
		return "***-**-" + match[len(match)-4:]
	case "phone":
		// Show last 4 digits
		cleaned := strings.Map(func(r rune) rune {
			if r >= '0' && r <= '9' {
				return r
			}
			return -1
		}, match)
		if len(cleaned) >= 4 {
			return "***-***-" + cleaned[len(cleaned)-4:]
		}
		return "[REDACTED_PHONE]"
	case "password", "api_key", "secret", "token", "bearer":
		return "[REDACTED]"
	default:
		return "[REDACTED]"
	}
}

// InitLogger initializes the logger with PII redaction
func InitLogger(serviceName string, debug bool) {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	if debug {
		opts.Level = slog.LevelDebug
	}

	baseHandler := slog.NewJSONHandler(os.Stdout, opts)
	redactingHandler := &PIIRedactingHandler{Handler: baseHandler}

	Log = slog.New(redactingHandler).With(
		slog.String("service", serviceName),
	)
	slog.SetDefault(Log)
}

// InitLoggerWithoutRedaction initializes logger without PII redaction (for testing)
func InitLoggerWithoutRedaction(serviceName string, debug bool) {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	if debug {
		opts.Level = slog.LevelDebug
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	Log = slog.New(handler).With(
		slog.String("service", serviceName),
	)
	slog.SetDefault(Log)
}
