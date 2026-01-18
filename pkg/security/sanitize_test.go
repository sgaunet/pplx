package security

import (
	"errors"
	"testing"
)

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "env var format with API_KEY",
			input:    "PERPLEXITY_API_KEY=pplx-1234567890abcdef",
			expected: "PERPLEXITY_API_KEY=pplx-****-cdef",
		},
		{
			name:     "env var format with api_key",
			input:    "api_key=sk-proj-abcd1234efgh5678",
			expected: "api_key=sk-p-****-5678",
		},
		{
			name:     "standalone pplx key",
			input:    "pplx-1234567890abcdef",
			expected: "pplx-****-cdef",
		},
		{
			name:     "standalone openai key",
			input:    "sk-proj-abcdefghijklmnopqrstuvwxyz",
			expected: "sk-p-****-wxyz",
		},
		{
			name:     "normal text",
			input:    "this is normal text",
			expected: "this is normal text",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "URL",
			input:    "https://api.example.com/v1/chat",
			expected: "https://api.example.com/v1/chat",
		},
		{
			name:     "long alphanumeric (potential key without prefix)",
			input:    "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5",
			expected: "a1b2-****-n4o5",
		},
		{
			name:     "short value",
			input:    "short",
			expected: "short",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeString(tt.input)
			if got != tt.expected {
				t.Errorf("SanitizeString(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSanitizeValue(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		value    any
		validate func(t *testing.T, result any)
	}{
		{
			name:  "api_key field with pplx key",
			key:   "api_key",
			value: "pplx-1234567890abcdef",
			validate: func(t *testing.T, result interface{}) {
				str, ok := result.(string)
				if !ok {
					t.Error("result should be string")
					return
				}
				if str == "pplx-1234567890abcdef" {
					t.Error("API key should be masked")
				}
				if str != "pplx-****-cdef" {
					t.Errorf("got %q, want %q", str, "pplx-****-cdef")
				}
			},
		},
		{
			name:  "token field",
			key:   "token",
			value: "sk-proj-abcdefghijklmnopqrstuvwxyz",
			validate: func(t *testing.T, result interface{}) {
				str, ok := result.(string)
				if !ok {
					t.Error("result should be string")
					return
				}
				if str == "sk-proj-abcdefghijklmnopqrstuvwxyz" {
					t.Error("token should be masked")
				}
			},
		},
		{
			name:  "password field",
			key:   "password",
			value: "mysecretpassword",
			validate: func(t *testing.T, result interface{}) {
				str, ok := result.(string)
				if !ok {
					t.Error("result should be string")
					return
				}
				if str == "mysecretpassword" {
					t.Error("password should be masked")
				}
			},
		},
		{
			name:  "secret field",
			key:   "secret",
			value: "topsecretvalue12",
			validate: func(t *testing.T, result interface{}) {
				str, ok := result.(string)
				if !ok {
					t.Error("result should be string")
					return
				}
				if str == "topsecretvalue12" {
					t.Error("secret should be masked")
				}
			},
		},
		{
			name:  "model field (not sensitive)",
			key:   "model",
			value: "sonar",
			validate: func(t *testing.T, result interface{}) {
				str, ok := result.(string)
				if !ok {
					t.Error("result should be string")
					return
				}
				if str != "sonar" {
					t.Error("non-sensitive field should not change")
				}
			},
		},
		{
			name:  "temperature field (not sensitive)",
			key:   "temperature",
			value: 0.7,
			validate: func(t *testing.T, result interface{}) {
				if result != 0.7 {
					t.Error("non-sensitive numeric field should not change")
				}
			},
		},
		{
			name:  "apikey without underscore",
			key:   "apikey",
			value: "pplx-test123456789abc",
			validate: func(t *testing.T, result interface{}) {
				str, ok := result.(string)
				if !ok {
					t.Error("result should be string")
					return
				}
				if str == "pplx-test123456789abc" {
					t.Error("apikey should be masked")
				}
			},
		},
		{
			name:  "KEY in uppercase",
			key:   "API_KEY",
			value: "key-123456789012345",
			validate: func(t *testing.T, result interface{}) {
				str, ok := result.(string)
				if !ok {
					t.Error("result should be string")
					return
				}
				if str == "key-123456789012345" {
					t.Error("API_KEY should be masked")
				}
			},
		},
		{
			name:  "non-sensitive key with potential API key value",
			key:   "user_input",
			value: "pplx-1234567890abcdef",
			validate: func(t *testing.T, result interface{}) {
				str, ok := result.(string)
				if !ok {
					t.Error("result should be string")
					return
				}
				if str == "pplx-1234567890abcdef" {
					t.Error("value that looks like API key should be masked even with non-sensitive key name")
				}
			},
		},
		{
			name:  "empty value with sensitive key",
			key:   "api_key",
			value: "",
			validate: func(t *testing.T, result interface{}) {
				// Empty values remain empty (not masked)
				if result != "" {
					t.Errorf("empty value should remain empty, got %v", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeValue(tt.key, tt.value)
			tt.validate(t, result)
		})
	}
}

func TestSanitizeError(t *testing.T) {
	tests := []struct {
		name        string
		input       error
		shouldMask  bool
		checkResult func(t *testing.T, result error)
	}{
		{
			name:       "nil error",
			input:      nil,
			shouldMask: false,
			checkResult: func(t *testing.T, result error) {
				if result != nil {
					t.Error("nil error should return nil")
				}
			},
		},
		{
			name:       "error with API key",
			input:      errors.New("API error with key: pplx-1234567890abcdef"),
			shouldMask: true,
			checkResult: func(t *testing.T, result error) {
				if result == nil {
					t.Error("should return non-nil error")
					return
				}
				errMsg := result.Error()
				if contains(errMsg, "pplx-1234567890abcdef") {
					t.Error("API key in error should be masked")
				}
				if !contains(errMsg, "pplx-****-cdef") {
					t.Errorf("error should contain masked key, got: %s", errMsg)
				}
				// Should wrap with ErrSanitized
				if !errors.Is(result, ErrSanitized) {
					t.Error("sanitized error should wrap ErrSanitized")
				}
			},
		},
		{
			name:       "error without API key",
			input:      errors.New("normal error message"),
			shouldMask: false,
			checkResult: func(t *testing.T, result error) {
				if result == nil {
					t.Error("should return non-nil error")
					return
				}
				if result.Error() != "normal error message" {
					t.Errorf("non-sensitive error should be unchanged, got: %q, want: %q", result.Error(), "normal error message")
				}
			},
		},
		{
			name:       "error with sk- prefix key",
			input:      errors.New("authentication failed: sk-proj-abc123def456ghi789"),
			shouldMask: true,
			checkResult: func(t *testing.T, result error) {
				if result == nil {
					t.Error("should return non-nil error")
					return
				}
				errMsg := result.Error()
				if contains(errMsg, "sk-proj-abc123def456ghi789") {
					t.Error("sk- key in error should be masked")
				}
				// Should wrap with ErrSanitized
				if !errors.Is(result, ErrSanitized) {
					t.Error("sanitized error should wrap ErrSanitized")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeError(tt.input)
			tt.checkResult(t, result)
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func BenchmarkSanitizeString(b *testing.B) {
	testCases := []string{
		"PERPLEXITY_API_KEY=pplx-1234567890abcdef",
		"pplx-1234567890abcdef",
		"normal text without secrets",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			_ = SanitizeString(tc)
		}
	}
}

func BenchmarkSanitizeValue(b *testing.B) {
	testCases := []struct {
		key   string
		value any
	}{
		{"api_key", "pplx-1234567890abcdef"},
		{"model", "sonar"},
		{"temperature", 0.7},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			_ = SanitizeValue(tc.key, tc.value)
		}
	}
}
