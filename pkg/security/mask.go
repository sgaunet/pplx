// Package security provides utilities for masking and sanitizing sensitive data
// such as API keys, tokens, and passwords in logs, error messages, and output.
package security

import "strings"

const (
	// minMaskLength is the minimum length for a key to show partial masking.
	minMaskLength = 12
	// minKeyLength is the minimum length for potential API key detection.
	minKeyLength = 10
	// minLongKeyLength is the minimum length for long alphanumeric key detection.
	minLongKeyLength = 20
	// alphanumThreshold is the minimum percentage of alphanumeric characters.
	alphanumThreshold = 0.9
	// maxKeyValueParts is the maximum parts when splitting KEY=value.
	maxKeyValueParts = 2
)

// MaskAPIKey masks an API key showing first 4 and last 4 characters.
// This function provides a safe way to display API keys in logs, error messages,
// and configuration output without exposing the full sensitive value.
//
// Examples:
//   - "pplx-1234567890abcdef" -> "pplx-****-cdef"
//   - "short" -> "[REDACTED]"
//   - "" -> "[EMPTY]"
func MaskAPIKey(key string) string {
	if key == "" {
		return "[EMPTY]"
	}

	if len(key) < minMaskLength {
		return "[REDACTED]"
	}

	return key[:4] + "-****-" + key[len(key)-4:]
}

// IsPotentialAPIKey checks if a string looks like an API key.
// This function uses heuristics to detect common API key patterns including:
//   - Common prefixes (pplx-, sk-, pk-, api-, key-, Bearer, token-, secret-)
//   - Long alphanumeric strings (20+ characters with 90%+ alphanumeric content)
//
// Returns true if the string matches API key patterns, false otherwise.
func IsPotentialAPIKey(s string) bool {
	s = strings.TrimSpace(s)

	if len(s) < minKeyLength {
		return false
	}

	// Check for common API key prefixes
	if hasAPIKeyPrefix(s) {
		return true
	}

	// Check for long alphanumeric strings
	return isLongAlphanumeric(s)
}

// hasAPIKeyPrefix checks if the string starts with a common API key prefix.
func hasAPIKeyPrefix(s string) bool {
	prefixes := []string{
		"pplx-", "sk-", "pk-", "api-", "key-",
		"Bearer ", "token-", "secret-",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}

	return false
}

// isLongAlphanumeric checks if a string is a long alphanumeric sequence
// that could be an API key (20+ chars, 90%+ alphanumeric, no spaces).
func isLongAlphanumeric(s string) bool {
	// Strings with spaces are not API keys
	if len(s) < minLongKeyLength || strings.Contains(s, " ") {
		return false
	}

	alphanumCount := 0
	for _, c := range s {
		if isAlphanumOrSeparator(c) {
			alphanumCount++
		}
	}

	// If 90% or more is alphanumeric, likely an API key
	return float64(alphanumCount)/float64(len(s)) >= alphanumThreshold
}

// isAlphanumOrSeparator checks if a character is alphanumeric or a separator (- or _).
func isAlphanumOrSeparator(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') || c == '-' || c == '_'
}
