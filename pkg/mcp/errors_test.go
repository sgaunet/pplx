package mcp

import (
	"errors"
	"testing"

	"github.com/sgaunet/pplx/pkg/clerrors"
)

func TestParameterError(t *testing.T) {
	t.Run("with value", func(t *testing.T) {
		err := NewParameterError("temperature", 2.5, "value out of range")
		expected := "parameter error for temperature=2.5: value out of range"
		if err.Error() != expected {
			t.Errorf("Expected error message %q, got %q", expected, err.Error())
		}
	})

	t.Run("without value", func(t *testing.T) {
		err := NewParameterError("model", nil, "required parameter missing")
		expected := "parameter error for model: required parameter missing"
		if err.Error() != expected {
			t.Errorf("Expected error message %q, got %q", expected, err.Error())
		}
	})

	t.Run("errors.As compatibility", func(t *testing.T) {
		err := NewParameterError("test", "value", "reason")
		var paramErr *ParameterError
		if !errors.As(err, &paramErr) {
			t.Error("Expected errors.As to work with ParameterError")
		}
		if paramErr.Parameter != "test" {
			t.Errorf("Expected parameter %q, got %q", "test", paramErr.Parameter)
		}
	})
}

func TestStreamError(t *testing.T) {
	t.Run("with wrapped error", func(t *testing.T) {
		wrappedErr := errors.New("connection timeout")
		err := NewStreamError("failed to read stream", wrappedErr)
		expected := "stream error: failed to read stream: connection timeout"
		if err.Error() != expected {
			t.Errorf("Expected error message %q, got %q", expected, err.Error())
		}
	})

	t.Run("without wrapped error", func(t *testing.T) {
		err := NewStreamError("no response received", nil)
		expected := "stream error: no response received"
		if err.Error() != expected {
			t.Errorf("Expected error message %q, got %q", expected, err.Error())
		}
	})

	t.Run("Unwrap returns wrapped error", func(t *testing.T) {
		wrappedErr := errors.New("original error")
		err := NewStreamError("stream failed", wrappedErr)
		if !errors.Is(err, wrappedErr) {
			t.Error("Expected errors.Is to find wrapped error")
		}
		if err.Unwrap() != wrappedErr {
			t.Error("Expected Unwrap to return original error")
		}
	})

	t.Run("errors.As compatibility", func(t *testing.T) {
		err := NewStreamError("test", nil)
		var streamErr *StreamError
		if !errors.As(err, &streamErr) {
			t.Error("Expected errors.As to work with StreamError")
		}
		if streamErr.Message != "test" {
			t.Errorf("Expected message %q, got %q", "test", streamErr.Message)
		}
	})
}

func TestValidationError(t *testing.T) {
	t.Run("with value", func(t *testing.T) {
		err := NewValidationError("search_recency", "invalid", "must be one of: day, week, month")
		expected := "validation failed for search_recency=invalid: must be one of: day, week, month"
		if err.Error() != expected {
			t.Errorf("Expected error message %q, got %q", expected, err.Error())
		}
	})

	t.Run("without value", func(t *testing.T) {
		err := NewValidationError("response_format", "", "cannot use both json_schema and regex")
		expected := "validation failed for response_format: cannot use both json_schema and regex"
		if err.Error() != expected {
			t.Errorf("Expected error message %q, got %q", expected, err.Error())
		}
	})

	t.Run("errors.As compatibility", func(t *testing.T) {
		err := NewValidationError("field", "value", "message")
		var valErr *clerrors.ValidationError
		if !errors.As(err, &valErr) {
			t.Error("Expected errors.As to work with ValidationError")
		}
		if valErr.Field != "field" {
			t.Errorf("Expected field %q, got %q", "field", valErr.Field)
		}
	})
}

func TestErrorWrapping(t *testing.T) {
	t.Run("StreamError wraps errors correctly", func(t *testing.T) {
		originalErr := errors.New("network error")
		streamErr := NewStreamError("streaming failed", originalErr)

		var target *StreamError
		if !errors.As(streamErr, &target) {
			t.Error("Expected errors.As to work")
		}

		if !errors.Is(streamErr, originalErr) {
			t.Error("Expected errors.Is to find original error")
		}

		unwrapped := errors.Unwrap(streamErr)
		if unwrapped != originalErr {
			t.Error("Expected Unwrap to return original error")
		}
	})
}
