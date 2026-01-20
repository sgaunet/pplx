package clerrors

import (
	"fmt"
	"strings"

	"github.com/sgaunet/pplx/pkg/security"
)

// ValidationError represents a validation error for command inputs.
type ValidationError struct {
	Field   string
	Value   string
	Message string
}

// NewValidationError creates a new validation error.
func NewValidationError(field, value, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

// NewValidationErrorSafe creates a validation error with automatic value sanitization.
// This function masks potential API keys and other sensitive data in the value field.
// Use this constructor when the value might contain sensitive information.
func NewValidationErrorSafe(field, value, message string) *ValidationError {
	sanitizedValue := security.SanitizeString(value)

	return &ValidationError{
		Field:   field,
		Value:   sanitizedValue,
		Message: message,
	}
}

func (e *ValidationError) Error() string {
	// Sanitize value in error output as defense-in-depth
	safeValue := security.SanitizeString(e.Value)

	if safeValue == "" {
		return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation failed for %s=%s: %s", e.Field, safeValue, e.Message)
}

// APIError represents an error from the Perplexity API.
type APIError struct {
	StatusCode int
	Message    string
	Err        error
}

// NewAPIError creates a new API error.
func NewAPIError(message string, err error) *APIError {
	return &APIError{
		Message: message,
		Err:     err,
	}
}

func (e *APIError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
	}
	if e.Err != nil {
		return fmt.Sprintf("API error: %s: %v", e.Message, e.Err)
	}
	return "API error: " + e.Message
}

// Unwrap returns the wrapped error.
func (e *APIError) Unwrap() error {
	return e.Err
}

// ConfigError represents a configuration error.
type ConfigError struct {
	Message string
	Err     error
}

// NewConfigError creates a new configuration error.
func NewConfigError(message string, err error) *ConfigError {
	return &ConfigError{
		Message: message,
		Err:     err,
	}
}

func (e *ConfigError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("configuration error: %s: %v", e.Message, e.Err)
	}
	return "configuration error: " + e.Message
}

// Unwrap returns the wrapped error.
func (e *ConfigError) Unwrap() error {
	return e.Err
}

// IOError represents an I/O error.
type IOError struct {
	Message string
	Err     error
}

// NewIOError creates a new I/O error.
func NewIOError(message string, err error) *IOError {
	return &IOError{
		Message: message,
		Err:     err,
	}
}

func (e *IOError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("I/O error: %s: %v", e.Message, e.Err)
	}
	return "I/O error: " + e.Message
}

// Unwrap returns the wrapped error.
func (e *IOError) Unwrap() error {
	return e.Err
}

// ValidationErrors is a collection of validation errors.
// It allows multiple validation errors to be collected and returned together,
// providing comprehensive validation feedback in a single error.
type ValidationErrors []ValidationError

// Error implements the error interface by joining all validation error messages.
func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	messages := make([]string, 0, len(e))
	for _, err := range e {
		messages = append(messages, err.Error())
	}

	// Join all error messages with semicolons for readability
	return strings.Join(messages, "; ")
}
