package config

import (
	"encoding/json"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestNewMetadataRegistry tests registry creation.
func TestNewMetadataRegistry(t *testing.T) {
	t.Parallel()

	registry := NewMetadataRegistry()
	if registry == nil {
		t.Fatal("NewMetadataRegistry() returned nil")
	}

	if registry.options == nil {
		t.Fatal("NewMetadataRegistry() created registry with nil options map")
	}

	// Verify options are initialized
	if len(registry.options) == 0 {
		t.Error("NewMetadataRegistry() created registry with no options")
	}
}

// TestMetadataRegistry_Count tests option counting.
func TestMetadataRegistry_Count(t *testing.T) {
	t.Parallel()

	registry := NewMetadataRegistry()
	count := registry.Count()

	// Should have 31 total options (8 defaults + 11 search + 9 output + 3 api)
	expectedCount := 31
	if count != expectedCount {
		t.Errorf("Count() = %d, want %d", count, expectedCount)
	}
}

// TestMetadataRegistry_CountBySection tests section counting.
func TestMetadataRegistry_CountBySection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		section       string
		expectedCount int
	}{
		{SectionDefaults, 8},
		{SectionSearch, 11},
		{SectionOutput, 9},
		{SectionAPI, 3},
	}

	registry := NewMetadataRegistry()

	for _, tt := range tests {
		t.Run(tt.section, func(t *testing.T) {
			count := registry.CountBySection(tt.section)
			if count != tt.expectedCount {
				t.Errorf("CountBySection(%s) = %d, want %d", tt.section, count, tt.expectedCount)
			}
		})
	}
}

// TestMetadataRegistry_GetOption tests retrieving specific options.
func TestMetadataRegistry_GetOption(t *testing.T) {
	t.Parallel()

	registry := NewMetadataRegistry()

	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "get by full key",
			key:     "defaults.model",
			wantErr: false,
		},
		{
			name:    "get by name only",
			key:     "model",
			wantErr: false,
		},
		{
			name:    "get search option",
			key:     "search.domains",
			wantErr: false,
		},
		{
			name:    "get by short name",
			key:     "domains",
			wantErr: false,
		},
		{
			name:    "nonexistent option",
			key:     "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt, err := registry.GetOption(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetOption(%s) error = %v, wantErr %v", tt.key, err, tt.wantErr)
				return
			}

			if !tt.wantErr && opt == nil {
				t.Errorf("GetOption(%s) returned nil option", tt.key)
			}

			if !tt.wantErr && opt != nil {
				if opt.Name == "" {
					t.Error("GetOption() returned option with empty name")
				}
				if opt.Section == "" {
					t.Error("GetOption() returned option with empty section")
				}
				if opt.Type == "" {
					t.Error("GetOption() returned option with empty type")
				}
			}
		})
	}
}

// TestMetadataRegistry_GetBySection tests section filtering.
func TestMetadataRegistry_GetBySection(t *testing.T) {
	t.Parallel()

	registry := NewMetadataRegistry()

	tests := []struct {
		section       string
		expectedCount int
	}{
		{SectionDefaults, 8},
		{SectionSearch, 11},
		{SectionOutput, 9},
		{SectionAPI, 3},
		{"DEFAULTS", 8}, // Case insensitive
		{"Search", 11},  // Case insensitive
	}

	for _, tt := range tests {
		t.Run(tt.section, func(t *testing.T) {
			options := registry.GetBySection(tt.section)
			if len(options) != tt.expectedCount {
				t.Errorf("GetBySection(%s) returned %d options, want %d", tt.section, len(options), tt.expectedCount)
			}

			// Verify all returned options belong to the correct section
			for _, opt := range options {
				if opt.Section != SectionDefaults && opt.Section != SectionSearch &&
					opt.Section != SectionOutput && opt.Section != SectionAPI {
					t.Errorf("GetBySection(%s) returned option with invalid section: %s", tt.section, opt.Section)
				}
			}
		})
	}
}

// TestMetadataRegistry_GetAll tests retrieving all options.
func TestMetadataRegistry_GetAll(t *testing.T) {
	t.Parallel()

	registry := NewMetadataRegistry()
	allOptions := registry.GetAll()

	expectedCount := 31 // 8 + 11 + 9 + 3
	if len(allOptions) != expectedCount {
		t.Errorf("GetAll() returned %d options, want %d", len(allOptions), expectedCount)
	}

	// Verify all options have required fields
	for _, opt := range allOptions {
		if opt.Name == "" {
			t.Error("GetAll() returned option with empty name")
		}
		if opt.Section == "" {
			t.Error("GetAll() returned option with empty section")
		}
		if opt.Type == "" {
			t.Error("GetAll() returned option with empty type")
		}
		if opt.Description == "" {
			t.Errorf("GetAll() returned option %s.%s with empty description", opt.Section, opt.Name)
		}
	}
}

// TestMetadataRegistry_ListSections tests section listing.
func TestMetadataRegistry_ListSections(t *testing.T) {
	t.Parallel()

	registry := NewMetadataRegistry()
	sections := registry.ListSections()

	expectedSections := []string{SectionDefaults, SectionSearch, SectionOutput, SectionAPI}
	if len(sections) != len(expectedSections) {
		t.Errorf("ListSections() returned %d sections, want %d", len(sections), len(expectedSections))
	}

	// Verify all expected sections are present
	sectionMap := make(map[string]bool)
	for _, section := range sections {
		sectionMap[section] = true
	}

	for _, expected := range expectedSections {
		if !sectionMap[expected] {
			t.Errorf("ListSections() missing expected section: %s", expected)
		}
	}
}

// TestOptionMetadata_RequiredFields tests that all options have required metadata fields.
func TestOptionMetadata_RequiredFields(t *testing.T) {
	t.Parallel()

	registry := NewMetadataRegistry()
	allOptions := registry.GetAll()

	for _, opt := range allOptions {
		t.Run(opt.Section+"."+opt.Name, func(t *testing.T) {
			if opt.Name == "" {
				t.Error("Option has empty Name")
			}
			if opt.Section == "" {
				t.Error("Option has empty Section")
			}
			if opt.Type == "" {
				t.Error("Option has empty Type")
			}
			if opt.Description == "" {
				t.Error("Option has empty Description")
			}

			// Verify type is one of the valid types
			validTypes := map[string]bool{
				"string":   true,
				"int":      true,
				"float64":  true,
				"bool":     true,
				"[]string": true,
				"duration": true,
			}
			if !validTypes[opt.Type] {
				t.Errorf("Option has invalid Type: %s", opt.Type)
			}
		})
	}
}

// TestOptionMetadata_KnownOptions tests that all expected options are present.
func TestOptionMetadata_KnownOptions(t *testing.T) {
	t.Parallel()

	registry := NewMetadataRegistry()

	// Known options that must be present
	knownOptions := []string{
		"defaults.model",
		"defaults.temperature",
		"defaults.max_tokens",
		"search.domains",
		"search.mode",
		"search.recency",
		"output.stream",
		"output.json",
		"output.return_images",
		"api.key",
		"api.base_url",
	}

	for _, optKey := range knownOptions {
		t.Run(optKey, func(t *testing.T) {
			opt, err := registry.GetOption(optKey)
			if err != nil {
				t.Errorf("GetOption(%s) error = %v, expected option to exist", optKey, err)
			}
			if opt == nil {
				t.Errorf("GetOption(%s) returned nil", optKey)
			}
		})
	}
}

// TestOptionMetadata_APIKeyRequired tests that API key is marked as required.
func TestOptionMetadata_APIKeyRequired(t *testing.T) {
	t.Parallel()

	registry := NewMetadataRegistry()
	opt, err := registry.GetOption("api.key")
	if err != nil {
		t.Fatalf("GetOption(api.key) error = %v", err)
	}

	if !opt.Required {
		t.Error("API key should be marked as required")
	}

	if opt.EnvVar != "PERPLEXITY_API_KEY" {
		t.Errorf("API key EnvVar = %s, want PERPLEXITY_API_KEY", opt.EnvVar)
	}
}

// TestOptionMetadata_ValidationRules tests that validation rules are present where expected.
func TestOptionMetadata_ValidationRules(t *testing.T) {
	t.Parallel()

	registry := NewMetadataRegistry()

	// Options that should have validation rules
	optionsWithValidation := []string{
		"defaults.temperature",
		"defaults.top_k",
		"defaults.top_p",
		"defaults.frequency_penalty",
		"defaults.presence_penalty",
		"search.mode",
		"search.recency",
		"api.key",
	}

	for _, optKey := range optionsWithValidation {
		t.Run(optKey, func(t *testing.T) {
			opt, err := registry.GetOption(optKey)
			if err != nil {
				t.Fatalf("GetOption(%s) error = %v", optKey, err)
			}

			if len(opt.ValidationRules) == 0 {
				t.Errorf("Option %s should have validation rules", optKey)
			}
		})
	}
}

// BenchmarkGetOption benchmarks option retrieval.
func BenchmarkGetOption(b *testing.B) {
	registry := NewMetadataRegistry()

	b.Run("by_full_key", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = registry.GetOption("defaults.model")
		}
	})

	b.Run("by_name_only", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = registry.GetOption("model")
		}
	})
}

// BenchmarkGetBySection benchmarks section filtering.
func BenchmarkGetBySection(b *testing.B) {
	registry := NewMetadataRegistry()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = registry.GetBySection(SectionDefaults)
	}
}

// TestTableFormatter tests table formatting.
func TestTableFormatter(t *testing.T) {
	t.Parallel()

	formatter := NewTableFormatter()

	t.Run("format_empty_options", func(t *testing.T) {
		result, err := formatter.Format([]*OptionMetadata{})
		if err != nil {
			t.Errorf("Format() error = %v", err)
		}
		if result == "" {
			t.Error("Format() returned empty string for empty options")
		}
	})

	t.Run("format_single_option", func(t *testing.T) {
		options := []*OptionMetadata{
			{
				Section:     "test",
				Name:        "option1",
				Type:        "string",
				Description: "Test option",
				Default:     "value",
			},
		}

		result, err := formatter.Format(options)
		if err != nil {
			t.Errorf("Format() error = %v", err)
		}
		if !containsAll(result, "test", "option1", "string") {
			t.Error("Format() missing expected content")
		}
	})

	t.Run("format_registry_options", func(t *testing.T) {
		registry := NewMetadataRegistry()
		options := registry.GetBySection(SectionDefaults)

		result, err := formatter.Format(options)
		if err != nil {
			t.Errorf("Format() error = %v", err)
		}
		if result == "" {
			t.Error("Format() returned empty string")
		}
	})
}

// TestJSONFormatter tests JSON formatting.
func TestJSONFormatter(t *testing.T) {
	t.Parallel()

	t.Run("format_with_indent", func(t *testing.T) {
		formatter := NewJSONFormatter(true)

		options := []*OptionMetadata{
			{
				Section:     "test",
				Name:        "option1",
				Type:        "string",
				Description: "Test option",
				Default:     "value",
			},
		}

		result, err := formatter.Format(options)
		if err != nil {
			t.Errorf("Format() error = %v", err)
		}

		// Verify it's valid JSON
		var parsed []OptionMetadata
		if err := json.Unmarshal([]byte(result), &parsed); err != nil {
			t.Errorf("Format() produced invalid JSON: %v", err)
		}

		if len(parsed) != 1 {
			t.Errorf("Format() parsed %d options, want 1", len(parsed))
		}
	})

	t.Run("format_without_indent", func(t *testing.T) {
		formatter := NewJSONFormatter(false)

		options := []*OptionMetadata{
			{
				Section: "test",
				Name:    "option1",
				Type:    "string",
			},
		}

		result, err := formatter.Format(options)
		if err != nil {
			t.Errorf("Format() error = %v", err)
		}

		var parsed []OptionMetadata
		if err := json.Unmarshal([]byte(result), &parsed); err != nil {
			t.Errorf("Format() produced invalid JSON: %v", err)
		}
	})

	t.Run("format_empty_options", func(t *testing.T) {
		formatter := NewJSONFormatter(true)

		result, err := formatter.Format([]*OptionMetadata{})
		if err != nil {
			t.Errorf("Format() error = %v", err)
		}

		if result != "[]" {
			t.Errorf("Format() = %q, want '[]'", result)
		}
	})
}

// TestYAMLFormatter tests YAML formatting.
func TestYAMLFormatter(t *testing.T) {
	t.Parallel()

	formatter := NewYAMLFormatter()

	t.Run("format_single_option", func(t *testing.T) {
		options := []*OptionMetadata{
			{
				Section:     "test",
				Name:        "option1",
				Type:        "string",
				Description: "Test option",
				Default:     "value",
			},
		}

		result, err := formatter.Format(options)
		if err != nil {
			t.Errorf("Format() error = %v", err)
		}

		// Verify it's valid YAML
		var parsed []OptionMetadata
		if err := yaml.Unmarshal([]byte(result), &parsed); err != nil {
			t.Errorf("Format() produced invalid YAML: %v", err)
		}

		if len(parsed) != 1 {
			t.Errorf("Format() parsed %d options, want 1", len(parsed))
		}
	})

	t.Run("format_empty_options", func(t *testing.T) {
		result, err := formatter.Format([]*OptionMetadata{})
		if err != nil {
			t.Errorf("Format() error = %v", err)
		}

		if result != "[]" {
			t.Errorf("Format() = %q, want '[]'", result)
		}
	})
}

// TestFormatOptions tests the format dispatcher.
func TestFormatOptions(t *testing.T) {
	t.Parallel()

	registry := NewMetadataRegistry()
	options := registry.GetBySection(SectionDefaults)

	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{
			name:    "table_format",
			format:  "table",
			wantErr: false,
		},
		{
			name:    "json_format",
			format:  "json",
			wantErr: false,
		},
		{
			name:    "yaml_format",
			format:  "yaml",
			wantErr: false,
		},
		{
			name:    "yml_format",
			format:  "yml",
			wantErr: false,
		},
		{
			name:    "case_insensitive",
			format:  "TABLE",
			wantErr: false,
		},
		{
			name:    "invalid_format",
			format:  "xml",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FormatOptions(options, tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("FormatOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result == "" {
				t.Error("FormatOptions() returned empty string")
			}
		})
	}
}

// TestFormatDefault tests the formatDefault helper.
func TestFormatDefault(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value interface{}
		want  string
	}{
		{"nil_value", nil, "(none)"},
		{"empty_string", "", "(empty)"},
		{"non_empty_string", "test", `"test"`},
		{"bool_true", true, "true"},
		{"bool_false", false, "false"},
		{"int_zero", 0, "(unset)"},
		{"int_nonzero", 42, "42"},
		{"float_zero", 0.0, "(unset)"},
		{"float_nonzero", 3.14, "3.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDefault(tt.value)
			if result != tt.want {
				t.Errorf("formatDefault(%v) = %q, want %q", tt.value, result, tt.want)
			}
		})
	}
}

// TestTruncate tests the truncate helper.
func TestTruncate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{"no_truncation", "short", 10, "short"},
		{"exact_length", "exact", 5, "exact"},
		{"truncate_long", "this is a very long string", 10, "this is..."},
		{"truncate_short_max", "test", 3, "tes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncate(tt.input, tt.maxLen)
			if result != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.want)
			}
		})
	}
}

// containsAll checks if a string contains all substrings.
func containsAll(s string, substrings ...string) bool {
	for _, sub := range substrings {
		if !containsString(s, sub) {
			return false
		}
	}
	return true
}

// containsString checks if a string contains a substring.
func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
