// Package logger provides structured logging using log/slog.
package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/sgaunet/pplx/pkg/security"
)

// Level represents the supported log levels.
type Level string

const (
	// LevelDebug enables debug-level logging.
	LevelDebug Level = "debug"
	// LevelInfo enables info-level logging.
	LevelInfo Level = "info"
	// LevelWarn enables warning-level logging.
	LevelWarn Level = "warn"
	// LevelError enables error-level logging.
	LevelError Level = "error"
)

// Format represents the supported log output formats.
type Format string

const (
	// FormatText outputs logs in human-readable text format.
	FormatText Format = "text"
	// FormatJSON outputs logs in JSON format.
	FormatJSON Format = "json"
)

var (
	// defaultLogger is the package-level logger instance.
	defaultLogger *slog.Logger
)

func init() {
	// Initialize with default settings (info level, text format)
	defaultLogger = New(LevelInfo, FormatText, os.Stderr)
}

// New creates a new slog.Logger with the specified level, format, and output.
func New(level Level, format Format, output io.Writer) *slog.Logger {
	// Convert string level to slog.Level
	var slogLevel slog.Level
	switch level {
	case LevelDebug:
		slogLevel = slog.LevelDebug
	case LevelInfo:
		slogLevel = slog.LevelInfo
	case LevelWarn:
		slogLevel = slog.LevelWarn
	case LevelError:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	// Create handler options
	opts := &slog.HandlerOptions{
		Level: slogLevel,
	}

	// Create handler based on format
	var handler slog.Handler
	if format == FormatJSON {
		handler = slog.NewJSONHandler(output, opts)
	} else {
		handler = slog.NewTextHandler(output, opts)
	}

	return slog.New(handler)
}

// Init initializes the default logger with the specified configuration.
func Init(level Level, format Format, output io.Writer) {
	defaultLogger = New(level, format, output)
	slog.SetDefault(defaultLogger)
}

// ParseLevel converts a string to a Level.
func ParseLevel(s string) (Level, bool) {
	switch strings.ToLower(s) {
	case "debug":
		return LevelDebug, true
	case "info":
		return LevelInfo, true
	case "warn", "warning":
		return LevelWarn, true
	case "error":
		return LevelError, true
	default:
		return LevelInfo, false
	}
}

// ParseFormat converts a string to a Format.
func ParseFormat(s string) (Format, bool) {
	switch strings.ToLower(s) {
	case "text":
		return FormatText, true
	case "json":
		return FormatJSON, true
	default:
		return FormatText, false
	}
}

// ValidLevels returns a slice of valid log level strings.
func ValidLevels() []string {
	return []string{"debug", "info", "warn", "error"}
}

// ValidFormats returns a slice of valid log format strings.
func ValidFormats() []string {
	return []string{"text", "json"}
}

// Debug logs a debug message using the default logger.
func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

// Info logs an info message using the default logger.
func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

// Warn logs a warning message using the default logger.
func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

// Error logs an error message using the default logger.
func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

// GetDefault returns the default logger instance.
func GetDefault() *slog.Logger {
	return defaultLogger
}

// SafeAttr creates a sanitized slog attribute.
// Use this for any attribute that might contain sensitive data such as API keys,
// tokens, passwords, or other secrets.
//
// Example:
//
//	logger.Info("API request", logger.SafeAttr("api_key", apiKey))
func SafeAttr(key string, value any) slog.Attr {
	strValue := fmt.Sprintf("%v", value)
	sanitized := security.SanitizeValue(key, strValue)
	return slog.Any(key, sanitized)
}

// InfoSafe logs an info message with automatic sanitization of all attributes.
// This function inspects attribute keys and values, masking any that look like
// sensitive data (API keys, tokens, passwords, etc.).
//
// Example:
//
//	logger.InfoSafe("API request", "api_key", apiKey, "model", "sonar")
func InfoSafe(msg string, args ...any) {
	safeArgs := sanitizeArgs(args)
	defaultLogger.Info(msg, safeArgs...)
}

// WarnSafe logs a warning message with automatic sanitization of all attributes.
func WarnSafe(msg string, args ...any) {
	safeArgs := sanitizeArgs(args)
	defaultLogger.Warn(msg, safeArgs...)
}

// ErrorSafe logs an error message with automatic sanitization of all attributes.
func ErrorSafe(msg string, args ...any) {
	safeArgs := sanitizeArgs(args)
	defaultLogger.Error(msg, safeArgs...)
}

// DebugSafe logs a debug message with automatic sanitization of all attributes.
func DebugSafe(msg string, args ...any) {
	safeArgs := sanitizeArgs(args)
	defaultLogger.Debug(msg, safeArgs...)
}

// sanitizeArgs sanitizes slog-style alternating key-value arguments.
func sanitizeArgs(args []any) []any {
	if len(args) == 0 {
		return args
	}

	sanitized := make([]any, len(args))
	copy(sanitized, args)

	// Process key-value pairs (slog format: key1, value1, key2, value2, ...)
	for i := 0; i < len(sanitized)-1; i += 2 {
		key, ok := sanitized[i].(string)
		if !ok {
			continue
		}

		sanitized[i+1] = security.SanitizeValue(key, sanitized[i+1])
	}

	return sanitized
}
