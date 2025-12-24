package clerrors

import (
	"errors"
	"fmt"
	"testing"
)

func TestValidationError(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		value    string
		message  string
		expected string
	}{
		{
			name:     "validation error with value",
			field:    "user-prompt",
			value:    "invalid",
			message:  "must not be empty",
			expected: "validation failed for user-prompt=invalid: must not be empty",
		},
		{
			name:     "validation error without value",
			field:    "api-key",
			value:    "",
			message:  "is required",
			expected: "validation failed for api-key: is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewValidationError(tt.field, tt.value, tt.message)
			if err.Error() != tt.expected {
				t.Errorf("ValidationError.Error() = %q, want %q", err.Error(), tt.expected)
			}
			if err.Field != tt.field {
				t.Errorf("ValidationError.Field = %q, want %q", err.Field, tt.field)
			}
			if err.Value != tt.value {
				t.Errorf("ValidationError.Value = %q, want %q", err.Value, tt.value)
			}
			if err.Message != tt.message {
				t.Errorf("ValidationError.Message = %q, want %q", err.Message, tt.message)
			}
		})
	}
}

func TestAPIError(t *testing.T) {
	tests := []struct {
		name       string
		message    string
		wrappedErr error
		statusCode int
		expected   string
	}{
		{
			name:       "API error with wrapped error",
			message:    "request failed",
			wrappedErr: fmt.Errorf("network timeout"),
			expected:   "API error: request failed: network timeout",
		},
		{
			name:       "API error without wrapped error",
			message:    "invalid response",
			wrappedErr: nil,
			expected:   "API error: invalid response",
		},
		{
			name:       "API error with status code",
			message:    "server error",
			statusCode: 500,
			expected:   "API error (status 500): server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewAPIError(tt.message, tt.wrappedErr)
			if tt.statusCode > 0 {
				err.StatusCode = tt.statusCode
			}
			if err.Error() != tt.expected {
				t.Errorf("APIError.Error() = %q, want %q", err.Error(), tt.expected)
			}
			if tt.wrappedErr != nil {
				if !errors.Is(err, tt.wrappedErr) {
					t.Errorf("APIError should wrap the original error")
				}
			}
		})
	}
}

func TestAPIErrorUnwrap(t *testing.T) {
	wrappedErr := fmt.Errorf("original error")
	apiErr := NewAPIError("API failed", wrappedErr)

	unwrapped := errors.Unwrap(apiErr)
	if unwrapped != wrappedErr {
		t.Errorf("APIError.Unwrap() = %v, want %v", unwrapped, wrappedErr)
	}
}

func TestConfigError(t *testing.T) {
	tests := []struct {
		name       string
		message    string
		wrappedErr error
		expected   string
	}{
		{
			name:       "config error with wrapped error",
			message:    "failed to load config",
			wrappedErr: fmt.Errorf("file not found"),
			expected:   "configuration error: failed to load config: file not found",
		},
		{
			name:       "config error without wrapped error",
			message:    "invalid configuration",
			wrappedErr: nil,
			expected:   "configuration error: invalid configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewConfigError(tt.message, tt.wrappedErr)
			if err.Error() != tt.expected {
				t.Errorf("ConfigError.Error() = %q, want %q", err.Error(), tt.expected)
			}
			if tt.wrappedErr != nil {
				if !errors.Is(err, tt.wrappedErr) {
					t.Errorf("ConfigError should wrap the original error")
				}
			}
		})
	}
}

func TestConfigErrorUnwrap(t *testing.T) {
	wrappedErr := fmt.Errorf("original error")
	configErr := NewConfigError("config failed", wrappedErr)

	unwrapped := errors.Unwrap(configErr)
	if unwrapped != wrappedErr {
		t.Errorf("ConfigError.Unwrap() = %v, want %v", unwrapped, wrappedErr)
	}
}

func TestIOError(t *testing.T) {
	tests := []struct {
		name       string
		message    string
		wrappedErr error
		expected   string
	}{
		{
			name:       "I/O error with wrapped error",
			message:    "failed to read file",
			wrappedErr: fmt.Errorf("permission denied"),
			expected:   "I/O error: failed to read file: permission denied",
		},
		{
			name:       "I/O error without wrapped error",
			message:    "write failed",
			wrappedErr: nil,
			expected:   "I/O error: write failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewIOError(tt.message, tt.wrappedErr)
			if err.Error() != tt.expected {
				t.Errorf("IOError.Error() = %q, want %q", err.Error(), tt.expected)
			}
			if tt.wrappedErr != nil {
				if !errors.Is(err, tt.wrappedErr) {
					t.Errorf("IOError should wrap the original error")
				}
			}
		})
	}
}

func TestIOErrorUnwrap(t *testing.T) {
	wrappedErr := fmt.Errorf("original error")
	ioErr := NewIOError("I/O failed", wrappedErr)

	unwrapped := errors.Unwrap(ioErr)
	if unwrapped != wrappedErr {
		t.Errorf("IOError.Unwrap() = %v, want %v", unwrapped, wrappedErr)
	}
}

func TestErrorsAs(t *testing.T) {
	// Test that errors.As works correctly with our custom error types
	t.Run("ValidationError with errors.As", func(t *testing.T) {
		err := NewValidationError("field", "value", "message")
		var validationErr *ValidationError
		if !errors.As(err, &validationErr) {
			t.Error("errors.As should recognize ValidationError")
		}
		if validationErr.Field != "field" {
			t.Errorf("ValidationError.Field = %q, want %q", validationErr.Field, "field")
		}
	})

	t.Run("APIError with errors.As", func(t *testing.T) {
		err := NewAPIError("message", nil)
		var apiErr *APIError
		if !errors.As(err, &apiErr) {
			t.Error("errors.As should recognize APIError")
		}
		if apiErr.Message != "message" {
			t.Errorf("APIError.Message = %q, want %q", apiErr.Message, "message")
		}
	})

	t.Run("ConfigError with errors.As", func(t *testing.T) {
		err := NewConfigError("message", nil)
		var configErr *ConfigError
		if !errors.As(err, &configErr) {
			t.Error("errors.As should recognize ConfigError")
		}
		if configErr.Message != "message" {
			t.Errorf("ConfigError.Message = %q, want %q", configErr.Message, "message")
		}
	})

	t.Run("IOError with errors.As", func(t *testing.T) {
		err := NewIOError("message", nil)
		var ioErr *IOError
		if !errors.As(err, &ioErr) {
			t.Error("errors.As should recognize IOError")
		}
		if ioErr.Message != "message" {
			t.Errorf("IOError.Message = %q, want %q", ioErr.Message, "message")
		}
	})
}

func TestWrappedErrorsAs(t *testing.T) {
	// Test that errors.As works with wrapped custom errors
	originalErr := fmt.Errorf("original error")
	apiErr := NewAPIError("API failed", originalErr)
	wrappedErr := fmt.Errorf("wrapper: %w", apiErr)

	var targetErr *APIError
	if !errors.As(wrappedErr, &targetErr) {
		t.Error("errors.As should find APIError in wrapped error chain")
	}
	if targetErr.Message != "API failed" {
		t.Errorf("APIError.Message = %q, want %q", targetErr.Message, "API failed")
	}

	// Test that we can also get the original error
	if !errors.Is(wrappedErr, originalErr) {
		t.Error("errors.Is should find original error in wrapped chain")
	}
}
