package logging

import (
	"io"
	"log/slog"
	"regexp"
	"strings"
)

// Logger wraps slog.Logger to provide structured logging throughout the application
type Logger struct {
	slog *slog.Logger
}

// NewLogger creates a new Logger with the specified level and output writer
func NewLogger(level slog.Level, output io.Writer) *Logger {
	handler := slog.NewJSONHandler(output, &slog.HandlerOptions{
		Level: level,
	})
	return &Logger{
		slog: slog.New(handler),
	}
}

// Debug logs a debug-level message with optional key-value pairs
func (l *Logger) Debug(msg string, args ...any) {
	l.slog.Debug(msg, args...)
}

// Info logs an info-level message with optional key-value pairs
func (l *Logger) Info(msg string, args ...any) {
	l.slog.Info(msg, args...)
}

// Warn logs a warning-level message with optional key-value pairs
func (l *Logger) Warn(msg string, args ...any) {
	l.slog.Warn(msg, args...)
}

// Error logs an error-level message with optional key-value pairs
func (l *Logger) Error(msg string, args ...any) {
	l.slog.Error(msg, args...)
}

// With returns a new Logger with the given key-value pairs added to all log entries
func (l *Logger) With(args ...any) *Logger {
	return &Logger{
		slog: l.slog.With(args...),
	}
}

// SanitizeSensitive redacts sensitive information like API keys from log output
// It replaces API keys and other sensitive patterns with [REDACTED]
func SanitizeSensitive(value string) string {
	// Pattern for API keys (common formats)
	// Matches: sk-..., api_key_..., Bearer ..., etc.
	patterns := []string{
		`sk-[a-zA-Z0-9]{20,}`,                    // OpenAI style keys
		`api[_-]?key[_-]?[a-zA-Z0-9]{20,}`,       // Generic API keys
		`Bearer\s+[a-zA-Z0-9\-._~+/]+=*`,         // Bearer tokens
		`[a-zA-Z0-9]{32,}`,                       // Long alphanumeric strings (likely keys)
		`AIza[a-zA-Z0-9\-_]{35}`,                 // Google API keys
		`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`, // UUIDs (often used as keys)
	}

	result := value
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllString(result, "[REDACTED]")
	}

	// Also redact common environment variable patterns
	envVarPattern := regexp.MustCompile(`(?i)(api[_-]?key|token|secret|password)\s*[:=]\s*[^\s,;]+`)
	result = envVarPattern.ReplaceAllStringFunc(result, func(match string) string {
		parts := strings.SplitN(match, ":", 2)
		if len(parts) == 2 {
			return parts[0] + ": [REDACTED]"
		}
		parts = strings.SplitN(match, "=", 2)
		if len(parts) == 2 {
			return parts[0] + "=[REDACTED]"
		}
		return "[REDACTED]"
	})

	return result
}
