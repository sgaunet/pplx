// Package mcp provides MCP (Model Context Protocol) server functionality
// for exposing Perplexity API through a standard protocol.
package mcp

import (
	"fmt"

	"github.com/sgaunet/pplx/pkg/security"
)

// ParameterError represents an error that occurred during parameter extraction.
type ParameterError struct {
	Parameter string
	Value     any
	Reason    string
}

// NewParameterError creates a new parameter error.
func NewParameterError(param string, value any, reason string) *ParameterError {
	return &ParameterError{
		Parameter: param,
		Value:     value,
		Reason:    reason,
	}
}

func (e *ParameterError) Error() string {
	if e.Value == nil {
		return fmt.Sprintf("parameter error for %s: %s", e.Parameter, e.Reason)
	}

	// Sanitize value before including in error
	strValue := fmt.Sprintf("%v", e.Value)
	safeValue := security.SanitizeString(strValue)

	return fmt.Sprintf("parameter error for %s=%s: %s", e.Parameter, safeValue, e.Reason)
}

// StreamError represents an error that occurred during streaming execution.
type StreamError struct {
	Message string
	Err     error
}

// NewStreamError creates a new stream error.
func NewStreamError(message string, err error) *StreamError {
	return &StreamError{
		Message: message,
		Err:     err,
	}
}

func (e *StreamError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("stream error: %s: %v", e.Message, e.Err)
	}
	return "stream error: " + e.Message
}

// Unwrap returns the wrapped error.
func (e *StreamError) Unwrap() error {
	return e.Err
}

// ValidationError represents a validation error for query parameters.
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

func (e *ValidationError) Error() string {
	// Sanitize value in error output
	safeValue := security.SanitizeString(e.Value)

	if safeValue == "" {
		return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation failed for %s=%s: %s", e.Field, safeValue, e.Message)
}
