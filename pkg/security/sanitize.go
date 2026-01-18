package security

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrSanitized indicates that an error message was sanitized to remove sensitive data.
	ErrSanitized = errors.New("error message sanitized")
)

// SanitizeString replaces potential API keys with masked versions.
// This function handles both KEY=value format and standalone API keys.
// It also detects API keys embedded within larger text.
//
// Examples:
//   - "PERPLEXITY_API_KEY=pplx-abc123" -> "PERPLEXITY_API_KEY=pplx-****-123"
//   - "pplx-abc123def456" -> "pplx-****-456"
//   - "Error: pplx-abc123 is invalid" -> "Error: pplx-****-123 is invalid"
//   - "normal text" -> "normal text" (unchanged)
func SanitizeString(s string) string {
	// Handle KEY=value format
	if strings.Contains(s, "API_KEY=") || strings.Contains(s, "api_key=") {
		parts := strings.SplitN(s, "=", maxKeyValueParts)
		if len(parts) == maxKeyValueParts && IsPotentialAPIKey(parts[1]) {
			return parts[0] + "=" + MaskAPIKey(parts[1])
		}
	}

	// If entire string looks like API key, mask it
	if IsPotentialAPIKey(s) {
		return MaskAPIKey(s)
	}

	// Search for API keys embedded in text (word-by-word)
	words := strings.Fields(s)
	modified := false
	for i, word := range words {
		// Remove trailing punctuation for detection
		cleanWord := strings.TrimRight(word, ".,;:!?")
		if IsPotentialAPIKey(cleanWord) {
			// Preserve trailing punctuation
			suffix := word[len(cleanWord):]
			words[i] = MaskAPIKey(cleanWord) + suffix
			modified = true
		}
	}
	if modified {
		return strings.Join(words, " ")
	}

	return s
}

// SanitizeValue sanitizes a value based on its key name and content.
// Used for structured logging and error contexts where key-value pairs are common.
//
// This function identifies sensitive keys by name (api_key, token, secret, password)
// and masks their values. It also detects values that look like API keys regardless
// of the key name.
//
// Returns the masked value if sensitive, otherwise returns the original value.
func SanitizeValue(key string, value any) any {
	lowerKey := strings.ToLower(key)
	isSensitiveKey := strings.Contains(lowerKey, "key") ||
		strings.Contains(lowerKey, "token") ||
		strings.Contains(lowerKey, "secret") ||
		strings.Contains(lowerKey, "password") ||
		strings.Contains(lowerKey, "apikey") ||
		strings.Contains(lowerKey, "api_key")

	strValue := fmt.Sprintf("%v", value)

	if isSensitiveKey && strValue != "" {
		return MaskAPIKey(strValue)
	}

	if IsPotentialAPIKey(strValue) {
		return MaskAPIKey(strValue)
	}

	return value
}

// SanitizeError creates a sanitized error message.
// This function removes potential API keys from error messages.
//
// Note: This intentionally breaks the error wrapping chain for security.
// If the error message contains sensitive data, a new error is returned with
// the sanitized message. Otherwise, the original error is returned unchanged.
func SanitizeError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	sanitized := SanitizeString(msg)

	if sanitized == msg {
		return err
	}

	// Return wrapped error with sanitized message
	// Intentionally breaks error chain for security
	return fmt.Errorf("%w: %s", ErrSanitized, sanitized)
}
