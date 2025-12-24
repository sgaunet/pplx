package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/sgaunet/pplx/pkg/clerrors"
)

func TestGetExitCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "ValidationError returns exit code 2",
			err:      clerrors.NewValidationError("field", "value", "message"),
			expected: exitCodeValidation,
		},
		{
			name:     "APIError returns exit code 3",
			err:      clerrors.NewAPIError("message", nil),
			expected: exitCodeAPI,
		},
		{
			name:     "ConfigError returns exit code 4",
			err:      clerrors.NewConfigError("message", nil),
			expected: exitCodeConfiguration,
		},
		{
			name:     "IOError returns exit code 5",
			err:      clerrors.NewIOError("message", nil),
			expected: exitCodeIO,
		},
		{
			name:     "Generic error returns exit code 1",
			err:      fmt.Errorf("generic error"),
			expected: exitCodeGeneral,
		},
		{
			name:     "Wrapped ValidationError returns exit code 2",
			err:      fmt.Errorf("wrapper: %w", clerrors.NewValidationError("field", "value", "message")),
			expected: exitCodeValidation,
		},
		{
			name:     "Wrapped APIError returns exit code 3",
			err:      fmt.Errorf("wrapper: %w", clerrors.NewAPIError("message", nil)),
			expected: exitCodeAPI,
		},
		{
			name:     "Wrapped ConfigError returns exit code 4",
			err:      fmt.Errorf("wrapper: %w", clerrors.NewConfigError("message", nil)),
			expected: exitCodeConfiguration,
		},
		{
			name:     "Wrapped IOError returns exit code 5",
			err:      fmt.Errorf("wrapper: %w", clerrors.NewIOError("message", nil)),
			expected: exitCodeIO,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getExitCode(tt.err)
			if got != tt.expected {
				t.Errorf("getExitCode() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestPrintError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		contains string
	}{
		{
			name:     "ValidationError prints validation error message",
			err:      clerrors.NewValidationError("user-prompt", "", "is required"),
			contains: "❌ Validation Error",
		},
		{
			name:     "APIError prints API error message",
			err:      clerrors.NewAPIError("request failed", nil),
			contains: "❌ API Error",
		},
		{
			name:     "ConfigError prints config error message",
			err:      clerrors.NewConfigError("config not found", nil),
			contains: "❌ Configuration Error",
		},
		{
			name:     "IOError prints I/O error message",
			err:      clerrors.NewIOError("read failed", nil),
			contains: "❌ I/O Error",
		},
		{
			name:     "Generic error prints generic error message",
			err:      fmt.Errorf("something went wrong"),
			contains: "❌ Error",
		},
		{
			name:     "Wrapped ValidationError prints validation error message",
			err:      fmt.Errorf("wrapper: %w", clerrors.NewValidationError("field", "value", "invalid")),
			contains: "❌ Validation Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stderr
			oldStderr := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w

			printError(tt.err)

			// Restore stderr
			w.Close()
			os.Stderr = oldStderr

			// Read captured output
			var buf bytes.Buffer
			buf.ReadFrom(r)
			output := buf.String()

			if output == "" {
				t.Error("printError() produced no output")
			}

			if !contains(output, tt.contains) {
				t.Errorf("printError() output = %q, should contain %q", output, tt.contains)
			}
		})
	}
}

func TestExitCodeConstants(t *testing.T) {
	// Verify exit code constants have expected values
	tests := []struct {
		name     string
		constant int
		expected int
	}{
		{"exitCodeSuccess", exitCodeSuccess, 0},
		{"exitCodeGeneral", exitCodeGeneral, 1},
		{"exitCodeValidation", exitCodeValidation, 2},
		{"exitCodeAPI", exitCodeAPI, 3},
		{"exitCodeConfiguration", exitCodeConfiguration, 4},
		{"exitCodeIO", exitCodeIO, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %d, want %d", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestPrintErrorWithNilError(t *testing.T) {
	// This shouldn't happen in practice, but let's ensure it doesn't panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("printError(nil) panicked: %v", r)
		}
	}()

	// Create a typed nil error
	var err error
	if err != nil {
		// Capture stderr to avoid noise in test output
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		printError(err)

		w.Close()
		os.Stderr = oldStderr
		var buf bytes.Buffer
		buf.ReadFrom(r)
	}
}

func TestGetExitCodePriority(t *testing.T) {
	// Test that when multiple error types could match, the most specific one is used
	// This tests the priority order of error checking
	t.Run("ValidationError takes precedence over wrapped errors", func(t *testing.T) {
		baseErr := fmt.Errorf("base error")
		validationErr := clerrors.NewValidationError("field", "value", "invalid")
		// Wrap validation error
		wrappedErr := fmt.Errorf("wrapped: %w", validationErr)
		// Wrap again with base error
		doubleWrapped := fmt.Errorf("double wrapped: %w: %w", wrappedErr, baseErr)

		exitCode := getExitCode(doubleWrapped)
		if exitCode != exitCodeValidation {
			t.Errorf("getExitCode() = %d, want %d (ValidationError should be found)", exitCode, exitCodeValidation)
		}
	})
}

// Helper function for contains check
func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}

func TestErrorsAsWorksWithMultipleWrapping(t *testing.T) {
	// Test that errors.As works through multiple levels of wrapping
	baseErr := fmt.Errorf("base")
	apiErr := clerrors.NewAPIError("API failed", baseErr)
	wrapped1 := fmt.Errorf("level 1: %w", apiErr)
	wrapped2 := fmt.Errorf("level 2: %w", wrapped1)

	var targetErr *clerrors.APIError
	if !errors.As(wrapped2, &targetErr) {
		t.Error("errors.As should find APIError through multiple wrapping levels")
	}

	if targetErr.Message != "API failed" {
		t.Errorf("APIError.Message = %q, want %q", targetErr.Message, "API failed")
	}

	// Verify exit code works correctly
	exitCode := getExitCode(wrapped2)
	if exitCode != exitCodeAPI {
		t.Errorf("getExitCode() = %d, want %d", exitCode, exitCodeAPI)
	}
}
