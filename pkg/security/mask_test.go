package security

import "testing"

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "standard pplx key",
			input:    "pplx-1234567890abcdef",
			expected: "pplx-****-cdef",
		},
		{
			name:     "openai key",
			input:    "sk-proj-1234567890abcdefghijklmn",
			expected: "sk-p-****-klmn",
		},
		{
			name:     "short key",
			input:    "short",
			expected: "[REDACTED]",
		},
		{
			name:     "empty key",
			input:    "",
			expected: "[EMPTY]",
		},
		{
			name:     "minimum length (11 chars)",
			input:    "12345678901",
			expected: "[REDACTED]",
		},
		{
			name:     "exactly 12 chars",
			input:    "123456789012",
			expected: "1234-****-9012",
		},
		{
			name:     "long key",
			input:    "api-key-very-long-1234567890abcdefghijklmnopqrstuvwxyz",
			expected: "api--****-wxyz",
		},
		{
			name:     "bearer token",
			input:    "Bearer abcd1234efgh5678ijkl",
			expected: "Bear-****-ijkl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskAPIKey(tt.input)
			if got != tt.expected {
				t.Errorf("MaskAPIKey(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsPotentialAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "pplx prefix",
			input:    "pplx-abc123def456",
			expected: true,
		},
		{
			name:     "sk prefix",
			input:    "sk-proj-abc123",
			expected: true,
		},
		{
			name:     "pk prefix",
			input:    "pk-test-xyz789",
			expected: true,
		},
		{
			name:     "api- prefix",
			input:    "api-1234567890",
			expected: true,
		},
		{
			name:     "key- prefix",
			input:    "key-abcdefghij",
			expected: true,
		},
		{
			name:     "Bearer prefix",
			input:    "Bearer token123456",
			expected: true,
		},
		{
			name:     "token- prefix",
			input:    "token-xyz123abc",
			expected: true,
		},
		{
			name:     "secret- prefix",
			input:    "secret-password123",
			expected: true,
		},
		{
			name:     "long alphanumeric (30 chars, 100% alphanumeric)",
			input:    "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5",
			expected: true,
		},
		{
			name:     "long alphanumeric with dashes (25 chars, 96% alphanumeric)",
			input:    "abc-123-def-456-ghi-789",
			expected: true,
		},
		{
			name:     "long alphanumeric with underscores",
			input:    "abc_123_def_456_ghi_789_jkl",
			expected: true,
		},
		{
			name:     "short string",
			input:    "short",
			expected: false,
		},
		{
			name:     "normal text",
			input:    "this is normal text",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "9 chars (below threshold)",
			input:    "123456789",
			expected: false,
		},
		{
			name:     "long text with spaces (not alphanumeric enough)",
			input:    "this is a very long sentence with many words",
			expected: false,
		},
		{
			name:     "URL (not alphanumeric enough due to special chars)",
			input:    "https://example.com/path?query=value",
			expected: false,
		},
		{
			name:     "whitespace padded pplx key",
			input:    "  pplx-abc123def456  ",
			expected: true,
		},
		{
			name:     "exactly 20 chars alphanumeric",
			input:    "12345678901234567890",
			expected: true,
		},
		{
			name:     "20 chars but only 80% alphanumeric",
			input:    "abc def ghi jkl mnop",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsPotentialAPIKey(tt.input)
			if got != tt.expected {
				t.Errorf("IsPotentialAPIKey(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestMaskAPIKey_EdgeCases(t *testing.T) {
	// Test with unicode characters
	unicode := "pplx-你好世界1234"
	masked := MaskAPIKey(unicode)
	if len(masked) == 0 {
		t.Error("MaskAPIKey should handle unicode without panicking")
	}

	// Test with very long key
	veryLong := "key-" + string(make([]byte, 1000))
	masked = MaskAPIKey(veryLong)
	if len(masked) == 0 {
		t.Error("MaskAPIKey should handle very long strings")
	}
}

func BenchmarkMaskAPIKey(b *testing.B) {
	key := "pplx-1234567890abcdefghijklmnop"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MaskAPIKey(key)
	}
}

func BenchmarkIsPotentialAPIKey(b *testing.B) {
	testCases := []string{
		"pplx-abc123",
		"normal text",
		"a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, tc := range testCases {
			_ = IsPotentialAPIKey(tc)
		}
	}
}
