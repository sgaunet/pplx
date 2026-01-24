package config

import (
	"reflect"
	"testing"
)

// TestIsValidSearchRecency tests the search recency validation helper.
func TestIsValidSearchRecency(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"valid hour", "hour", true},
		{"valid day", "day", true},
		{"valid week", "week", true},
		{"valid month", "month", true},
		{"valid year", "year", true},
		{"invalid value", "invalid", false},
		{"empty string", "", false},
		{"uppercase", "HOUR", false},
		{"mixed case", "Day", false},
		{"partial match", "hours", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidSearchRecency(tt.value)
			if got != tt.expected {
				t.Errorf("IsValidSearchRecency(%q) = %v, want %v", tt.value, got, tt.expected)
			}
		})
	}
}

// TestIsValidSearchMode tests the search mode validation helper.
func TestIsValidSearchMode(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"valid web", "web", true},
		{"valid academic", "academic", true},
		{"invalid value", "invalid", false},
		{"empty string", "", false},
		{"uppercase", "WEB", false},
		{"mixed case", "Academic", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidSearchMode(tt.value)
			if got != tt.expected {
				t.Errorf("IsValidSearchMode(%q) = %v, want %v", tt.value, got, tt.expected)
			}
		})
	}
}

// TestIsValidContextSize tests the context size validation helper.
func TestIsValidContextSize(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"valid low", "low", true},
		{"valid medium", "medium", true},
		{"valid high", "high", true},
		{"invalid value", "invalid", false},
		{"empty string", "", false},
		{"uppercase", "LOW", false},
		{"mixed case", "Medium", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidContextSize(tt.value)
			if got != tt.expected {
				t.Errorf("IsValidContextSize(%q) = %v, want %v", tt.value, got, tt.expected)
			}
		})
	}
}

// TestIsValidReasoningEffort tests the reasoning effort validation helper.
func TestIsValidReasoningEffort(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"valid low", "low", true},
		{"valid medium", "medium", true},
		{"valid high", "high", true},
		{"invalid value", "invalid", false},
		{"empty string", "", false},
		{"uppercase", "HIGH", false},
		{"mixed case", "Low", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidReasoningEffort(tt.value)
			if got != tt.expected {
				t.Errorf("IsValidReasoningEffort(%q) = %v, want %v", tt.value, got, tt.expected)
			}
		})
	}
}

// TestIsValidImageFormat tests the image format validation helper.
func TestIsValidImageFormat(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"valid jpg", "jpg", true},
		{"valid jpeg", "jpeg", true},
		{"valid png", "png", true},
		{"valid gif", "gif", true},
		{"valid webp", "webp", true},
		{"valid svg", "svg", true},
		{"valid bmp", "bmp", true},
		{"invalid format", "invalid", false},
		{"empty string", "", false},
		{"uppercase", "JPG", false},
		{"mixed case", "Png", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidImageFormat(tt.value)
			if got != tt.expected {
				t.Errorf("IsValidImageFormat(%q) = %v, want %v", tt.value, got, tt.expected)
			}
		})
	}
}

// TestGetValidSearchRecencyValues tests the search recency values getter.
func TestGetValidSearchRecencyValues(t *testing.T) {
	expected := []string{"hour", "day", "week", "month", "year"}
	got := GetValidSearchRecencyValues()

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("GetValidSearchRecencyValues() = %v, want %v", got, expected)
	}

	// Verify all returned values are valid
	for _, value := range got {
		if !IsValidSearchRecency(value) {
			t.Errorf("GetValidSearchRecencyValues() returned invalid value: %q", value)
		}
	}
}

// TestGetValidSearchModeValues tests the search mode values getter.
func TestGetValidSearchModeValues(t *testing.T) {
	expected := []string{"web", "academic"}
	got := GetValidSearchModeValues()

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("GetValidSearchModeValues() = %v, want %v", got, expected)
	}

	// Verify all returned values are valid
	for _, value := range got {
		if !IsValidSearchMode(value) {
			t.Errorf("GetValidSearchModeValues() returned invalid value: %q", value)
		}
	}
}

// TestGetValidContextSizeValues tests the context size values getter.
func TestGetValidContextSizeValues(t *testing.T) {
	expected := []string{"low", "medium", "high"}
	got := GetValidContextSizeValues()

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("GetValidContextSizeValues() = %v, want %v", got, expected)
	}

	// Verify all returned values are valid
	for _, value := range got {
		if !IsValidContextSize(value) {
			t.Errorf("GetValidContextSizeValues() returned invalid value: %q", value)
		}
	}
}

// TestGetValidReasoningEffortValues tests the reasoning effort values getter.
func TestGetValidReasoningEffortValues(t *testing.T) {
	expected := []string{"low", "medium", "high"}
	got := GetValidReasoningEffortValues()

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("GetValidReasoningEffortValues() = %v, want %v", got, expected)
	}

	// Verify all returned values are valid
	for _, value := range got {
		if !IsValidReasoningEffort(value) {
			t.Errorf("GetValidReasoningEffortValues() returned invalid value: %q", value)
		}
	}
}

// TestGetValidImageFormatValues tests the image format values getter.
func TestGetValidImageFormatValues(t *testing.T) {
	expected := []string{"jpg", "jpeg", "png", "gif", "webp", "svg", "bmp"}
	got := GetValidImageFormatValues()

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("GetValidImageFormatValues() = %v, want %v", got, expected)
	}

	// Verify all returned values are valid
	for _, value := range got {
		if !IsValidImageFormat(value) {
			t.Errorf("GetValidImageFormatValues() returned invalid value: %q", value)
		}
	}
}

// TestValidationMapsConsistency tests that validation maps and getters are consistent.
func TestValidationMapsConsistency(t *testing.T) {
	t.Run("search recency map consistency", func(t *testing.T) {
		values := GetValidSearchRecencyValues()
		if len(values) != len(ValidSearchRecency) {
			t.Errorf("GetValidSearchRecencyValues() length = %d, ValidSearchRecency length = %d",
				len(values), len(ValidSearchRecency))
		}
		for _, v := range values {
			if !ValidSearchRecency[v] {
				t.Errorf("Value %q from GetValidSearchRecencyValues() not in ValidSearchRecency", v)
			}
		}
	})

	t.Run("search mode map consistency", func(t *testing.T) {
		values := GetValidSearchModeValues()
		if len(values) != len(ValidSearchModes) {
			t.Errorf("GetValidSearchModeValues() length = %d, ValidSearchModes length = %d",
				len(values), len(ValidSearchModes))
		}
		for _, v := range values {
			if !ValidSearchModes[v] {
				t.Errorf("Value %q from GetValidSearchModeValues() not in ValidSearchModes", v)
			}
		}
	})

	t.Run("context size map consistency", func(t *testing.T) {
		values := GetValidContextSizeValues()
		if len(values) != len(ValidContextSizes) {
			t.Errorf("GetValidContextSizeValues() length = %d, ValidContextSizes length = %d",
				len(values), len(ValidContextSizes))
		}
		for _, v := range values {
			if !ValidContextSizes[v] {
				t.Errorf("Value %q from GetValidContextSizeValues() not in ValidContextSizes", v)
			}
		}
	})

	t.Run("reasoning effort map consistency", func(t *testing.T) {
		values := GetValidReasoningEffortValues()
		if len(values) != len(ValidReasoningEfforts) {
			t.Errorf("GetValidReasoningEffortValues() length = %d, ValidReasoningEfforts length = %d",
				len(values), len(ValidReasoningEfforts))
		}
		for _, v := range values {
			if !ValidReasoningEfforts[v] {
				t.Errorf("Value %q from GetValidReasoningEffortValues() not in ValidReasoningEfforts", v)
			}
		}
	})

	t.Run("image format map consistency", func(t *testing.T) {
		values := GetValidImageFormatValues()
		if len(values) != len(ValidImageFormats) {
			t.Errorf("GetValidImageFormatValues() length = %d, ValidImageFormats length = %d",
				len(values), len(ValidImageFormats))
		}
		for _, v := range values {
			if !ValidImageFormats[v] {
				t.Errorf("Value %q from GetValidImageFormatValues() not in ValidImageFormats", v)
			}
		}
	})
}

// TestValidationMapsContent tests that validation maps contain exactly the expected keys.
func TestValidationMapsContent(t *testing.T) {
	t.Run("ValidSearchRecency contains exactly expected keys", func(t *testing.T) {
		expectedKeys := map[string]bool{
			"hour": true, "day": true, "week": true, "month": true, "year": true,
		}
		if !reflect.DeepEqual(ValidSearchRecency, expectedKeys) {
			t.Errorf("ValidSearchRecency = %v, want %v", ValidSearchRecency, expectedKeys)
		}
	})

	t.Run("ValidSearchModes contains exactly expected keys", func(t *testing.T) {
		expectedKeys := map[string]bool{
			"web": true, "academic": true,
		}
		if !reflect.DeepEqual(ValidSearchModes, expectedKeys) {
			t.Errorf("ValidSearchModes = %v, want %v", ValidSearchModes, expectedKeys)
		}
	})

	t.Run("ValidContextSizes contains exactly expected keys", func(t *testing.T) {
		expectedKeys := map[string]bool{
			"low": true, "medium": true, "high": true,
		}
		if !reflect.DeepEqual(ValidContextSizes, expectedKeys) {
			t.Errorf("ValidContextSizes = %v, want %v", ValidContextSizes, expectedKeys)
		}
	})

	t.Run("ValidReasoningEfforts contains exactly expected keys", func(t *testing.T) {
		expectedKeys := map[string]bool{
			"low": true, "medium": true, "high": true,
		}
		if !reflect.DeepEqual(ValidReasoningEfforts, expectedKeys) {
			t.Errorf("ValidReasoningEfforts = %v, want %v", ValidReasoningEfforts, expectedKeys)
		}
	})

	t.Run("ValidImageFormats contains exactly expected keys", func(t *testing.T) {
		expectedKeys := map[string]bool{
			"jpg": true, "jpeg": true, "png": true, "gif": true,
			"webp": true, "svg": true, "bmp": true,
		}
		if !reflect.DeepEqual(ValidImageFormats, expectedKeys) {
			t.Errorf("ValidImageFormats = %v, want %v", ValidImageFormats, expectedKeys)
		}
	})
}
