package clerrors

import (
	"errors"
	"fmt"
	"testing"
)

// TestSentinelErrorsAreUnique verifies that all sentinel error messages are unique.
// This prevents accidental duplicate error messages which could cause confusion.
func TestSentinelErrorsAreUnique(t *testing.T) {
	allErrors := []error{
		// Configuration errors
		ErrNoConfigFound,
		ErrPathIsDirectory,
		ErrConfigFileExists,
		ErrValidationFailed,
		ErrUnknownSection,

		// Profile errors
		ErrProfileNameEmpty,
		ErrProfileNameReserved,
		ErrProfileNotFound,
		ErrProfileAlreadyExists,
		ErrDeleteDefaultProfile,
		ErrUpdateDefaultProfile,
		ErrImportReservedName,

		// Template errors
		ErrTemplateNotFound,
		ErrTemplateInvalid,

		// Metadata errors
		ErrOptionNotFound,
		ErrUnsupportedFormat,

		// Chat errors
		ErrInvalidSearchRecency,
		ErrConflictingResponseFormats,
		ErrResponseFormatNotSupported,
		ErrInvalidSearchMode,
		ErrInvalidSearchContextSize,
		ErrInvalidSearchAfterDate,
		ErrInvalidSearchBeforeDate,
		ErrInvalidLastUpdatedAfter,
		ErrInvalidLastUpdatedBefore,
		ErrInvalidReasoningEffort,

		// Command errors
		ErrInvalidLogLevel,
		ErrInvalidLogFormat,
		ErrFailedToReadInput,
		ErrFailedToReadAPIKey,
		ErrUnsupportedShell,
		ErrNoShellEnv,
	}

	// Check for duplicate error messages
	seen := make(map[string]error)
	for _, err := range allErrors {
		msg := err.Error()
		if existing, found := seen[msg]; found {
			t.Errorf("Duplicate error message found: %q (both %v and %v)", msg, existing, err)
		}
		seen[msg] = err
	}

	// Verify we have all expected errors
	expectedCount := 32
	if len(allErrors) != expectedCount {
		t.Errorf("Expected %d sentinel errors, got %d", expectedCount, len(allErrors))
	}
}

// TestErrorsIsWorks verifies that errors.Is() works correctly with sentinel errors.
// This ensures that sentinel errors can be properly detected even when wrapped.
func TestErrorsIsWorks(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		target    error
		shouldMatch bool
	}{
		{
			name:      "exact match - ErrProfileNotFound",
			err:       ErrProfileNotFound,
			target:    ErrProfileNotFound,
			shouldMatch: true,
		},
		{
			name:      "wrapped match - ErrTemplateNotFound",
			err:       fmt.Errorf("failed to load: %w", ErrTemplateNotFound),
			target:    ErrTemplateNotFound,
			shouldMatch: true,
		},
		{
			name:      "no match - different errors",
			err:       ErrProfileNotFound,
			target:    ErrTemplateNotFound,
			shouldMatch: false,
		},
		{
			name:      "double wrapped - ErrConfigFileExists",
			err:       fmt.Errorf("operation failed: %w", fmt.Errorf("nested: %w", ErrConfigFileExists)),
			target:    ErrConfigFileExists,
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := errors.Is(tt.err, tt.target)
			if result != tt.shouldMatch {
				t.Errorf("errors.Is(%v, %v) = %v, want %v", tt.err, tt.target, result, tt.shouldMatch)
			}
		})
	}
}

// TestConfigurationErrors tests configuration-related sentinel errors.
func TestConfigurationErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrNoConfigFound", ErrNoConfigFound},
		{"ErrPathIsDirectory", ErrPathIsDirectory},
		{"ErrConfigFileExists", ErrConfigFileExists},
		{"ErrValidationFailed", ErrValidationFailed},
		{"ErrUnknownSection", ErrUnknownSection},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Errorf("%s is nil", tt.name)
			}
			if tt.err.Error() == "" {
				t.Errorf("%s has empty error message", tt.name)
			}
		})
	}
}

// TestProfileErrors tests profile-related sentinel errors.
func TestProfileErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrProfileNameEmpty", ErrProfileNameEmpty},
		{"ErrProfileNameReserved", ErrProfileNameReserved},
		{"ErrProfileNotFound", ErrProfileNotFound},
		{"ErrProfileAlreadyExists", ErrProfileAlreadyExists},
		{"ErrDeleteDefaultProfile", ErrDeleteDefaultProfile},
		{"ErrUpdateDefaultProfile", ErrUpdateDefaultProfile},
		{"ErrImportReservedName", ErrImportReservedName},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Errorf("%s is nil", tt.name)
			}
			if tt.err.Error() == "" {
				t.Errorf("%s has empty error message", tt.name)
			}
		})
	}
}

// TestChatErrors tests chat-related sentinel errors.
func TestChatErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"ErrInvalidSearchRecency", ErrInvalidSearchRecency},
		{"ErrConflictingResponseFormats", ErrConflictingResponseFormats},
		{"ErrResponseFormatNotSupported", ErrResponseFormatNotSupported},
		{"ErrInvalidSearchMode", ErrInvalidSearchMode},
		{"ErrInvalidSearchContextSize", ErrInvalidSearchContextSize},
		{"ErrInvalidSearchAfterDate", ErrInvalidSearchAfterDate},
		{"ErrInvalidSearchBeforeDate", ErrInvalidSearchBeforeDate},
		{"ErrInvalidLastUpdatedAfter", ErrInvalidLastUpdatedAfter},
		{"ErrInvalidLastUpdatedBefore", ErrInvalidLastUpdatedBefore},
		{"ErrInvalidReasoningEffort", ErrInvalidReasoningEffort},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Errorf("%s is nil", tt.name)
			}
			if tt.err.Error() == "" {
				t.Errorf("%s has empty error message", tt.name)
			}
		})
	}
}

// TestValidationErrorsCollection tests the ValidationErrors collection type.
func TestValidationErrorsCollection(t *testing.T) {
	t.Run("empty collection", func(t *testing.T) {
		errs := ValidationErrors{}
		if errs.Error() != "" {
			t.Errorf("Empty ValidationErrors should return empty string, got %q", errs.Error())
		}
	})

	t.Run("single error", func(t *testing.T) {
		errs := ValidationErrors{
			*NewValidationError("field1", "value1", "error message"),
		}
		msg := errs.Error()
		if msg == "" {
			t.Error("ValidationErrors with one error should not return empty string")
		}
		if !contains(msg, "field1") {
			t.Errorf("Error message should contain field name, got %q", msg)
		}
	})

	t.Run("multiple errors", func(t *testing.T) {
		errs := ValidationErrors{
			*NewValidationError("field1", "value1", "error 1"),
			*NewValidationError("field2", "value2", "error 2"),
			*NewValidationError("field3", "value3", "error 3"),
		}
		msg := errs.Error()
		if msg == "" {
			t.Error("ValidationErrors with multiple errors should not return empty string")
		}
		// Should contain all field names
		for _, fieldName := range []string{"field1", "field2", "field3"} {
			if !contains(msg, fieldName) {
				t.Errorf("Error message should contain %s, got %q", fieldName, msg)
			}
		}
		// Should join with semicolons
		if !contains(msg, ";") {
			t.Errorf("Multiple errors should be joined with semicolons, got %q", msg)
		}
	})
}

// Helper function to check if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
