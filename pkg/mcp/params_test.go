package mcp

import (
	"errors"
	"testing"
	"time"

	"github.com/sgaunet/perplexity-go/v2"
)

func TestParameterExtractor_Extract(t *testing.T) {
	tests := []struct {
		name      string
		args      map[string]any
		expected  *QueryParams
		shouldErr bool
		errMsg    string
	}{
		{
			name: "all parameters provided",
			args: map[string]any{
				"user_prompt":                  "test query",
				"system_prompt":                "system",
				"model":                        "sonar",
				"frequency_penalty":            0.5,
				"max_tokens":                   float64(100),
				"presence_penalty":             0.3,
				"temperature":                  0.7,
				"top_k":                        float64(50),
				"top_p":                        0.9,
				"timeout":                      float64(60),
				"search_domains":               []any{"example.com", "test.com"},
				"search_recency":               "week",
				"location_lat":                 40.7128,
				"location_lon":                 -74.0060,
				"location_country":             "US",
				"return_images":                true,
				"return_related":               true,
				"stream":                       true,
				"image_domains":                []any{"img.example.com"},
				"image_formats":                []any{"jpg", "png"},
				"response_format_json_schema":  `{"type": "object"}`,
				"response_format_regex":        "",
				"search_mode":                  "web",
				"search_context_size":          "high",
				"search_after_date":            "01/01/2024",
				"search_before_date":           "12/31/2024",
				"last_updated_after":           "06/01/2024",
				"last_updated_before":          "06/30/2024",
				"reasoning_effort":             "high",
			},
			expected: &QueryParams{
				UserPrompt:                  "test query",
				SystemPrompt:                "system",
				Model:                       "sonar",
				FrequencyPenalty:            0.5,
				MaxTokens:                   100,
				PresencePenalty:             0.3,
				Temperature:                 0.7,
				TopK:                        50,
				TopP:                        0.9,
				Timeout:                     60 * time.Second,
				SearchDomains:               []string{"example.com", "test.com"},
				SearchRecency:               "week",
				LocationLat:                 40.7128,
				LocationLon:                 -74.0060,
				LocationCountry:             "US",
				ReturnImages:                true,
				ReturnRelated:               true,
				Stream:                      true,
				ImageDomains:                []string{"img.example.com"},
				ImageFormats:                []string{"jpg", "png"},
				ResponseFormatJSONSchema:    `{"type": "object"}`,
				ResponseFormatRegex:         "",
				SearchMode:                  "web",
				SearchContextSize:           "high",
				SearchAfterDate:             "01/01/2024",
				SearchBeforeDate:            "12/31/2024",
				LastUpdatedAfter:            "06/01/2024",
				LastUpdatedBefore:           "06/30/2024",
				ReasoningEffort:             "high",
			},
			shouldErr: false,
		},
		{
			name: "minimal parameters with defaults",
			args: map[string]any{
				"user_prompt": "test query",
			},
			expected: &QueryParams{
				UserPrompt:       "test query",
				Model:            perplexity.DefaultModel,
				FrequencyPenalty: perplexity.DefaultFrequencyPenalty,
				MaxTokens:        perplexity.DefaultMaxTokens,
				PresencePenalty:  perplexity.DefaultPresencePenalty,
				Temperature:      perplexity.DefaultTemperature,
				TopK:             perplexity.DefaultTopK,
				TopP:             perplexity.DefaultTopP,
				Timeout:          perplexity.DefaultTimeout,
			},
			shouldErr: false,
		},
		{
			name:      "missing user_prompt",
			args:      map[string]any{},
			shouldErr: true,
			errMsg:    "must be a non-empty string",
		},
		{
			name: "empty user_prompt",
			args: map[string]any{
				"user_prompt": "",
			},
			shouldErr: true,
			errMsg:    "must be a non-empty string",
		},
		{
			name: "user_prompt wrong type",
			args: map[string]any{
				"user_prompt": 123,
			},
			shouldErr: true,
			errMsg:    "must be a non-empty string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := NewParameterExtractor()
			result, err := extractor.Extract(tt.args)

			if tt.shouldErr {
				if err == nil {
					t.Error("Expected error but got nil")
					return
				}
				if tt.errMsg != "" {
					var paramErr *ParameterError
					if !errors.As(err, &paramErr) {
						t.Errorf("Expected ParameterError, got %T", err)
					}
					if !contains(err.Error(), tt.errMsg) {
						t.Errorf("Expected error containing %q, got %q", tt.errMsg, err.Error())
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Verify all fields
			if result.UserPrompt != tt.expected.UserPrompt {
				t.Errorf("UserPrompt: expected %q, got %q", tt.expected.UserPrompt, result.UserPrompt)
			}
			if result.SystemPrompt != tt.expected.SystemPrompt {
				t.Errorf("SystemPrompt: expected %q, got %q", tt.expected.SystemPrompt, result.SystemPrompt)
			}
			if result.Model != tt.expected.Model {
				t.Errorf("Model: expected %q, got %q", tt.expected.Model, result.Model)
			}
			if result.FrequencyPenalty != tt.expected.FrequencyPenalty {
				t.Errorf("FrequencyPenalty: expected %v, got %v", tt.expected.FrequencyPenalty, result.FrequencyPenalty)
			}
			if result.MaxTokens != tt.expected.MaxTokens {
				t.Errorf("MaxTokens: expected %d, got %d", tt.expected.MaxTokens, result.MaxTokens)
			}
			if result.Temperature != tt.expected.Temperature {
				t.Errorf("Temperature: expected %v, got %v", tt.expected.Temperature, result.Temperature)
			}
			if result.TopK != tt.expected.TopK {
				t.Errorf("TopK: expected %d, got %d", tt.expected.TopK, result.TopK)
			}
			if result.Timeout != tt.expected.Timeout {
				t.Errorf("Timeout: expected %v, got %v", tt.expected.Timeout, result.Timeout)
			}
			if result.ReturnImages != tt.expected.ReturnImages {
				t.Errorf("ReturnImages: expected %v, got %v", tt.expected.ReturnImages, result.ReturnImages)
			}
			if result.Stream != tt.expected.Stream {
				t.Errorf("Stream: expected %v, got %v", tt.expected.Stream, result.Stream)
			}
		})
	}
}

func TestParameterExtractor_ExtractString(t *testing.T) {
	extractor := NewParameterExtractor()

	tests := []struct {
		name     string
		args     map[string]any
		key      string
		defVal   string
		expected string
	}{
		{
			name:     "string present",
			args:     map[string]any{"model": "sonar"},
			key:      "model",
			defVal:   "default",
			expected: "sonar",
		},
		{
			name:     "string missing",
			args:     map[string]any{},
			key:      "model",
			defVal:   "default",
			expected: "default",
		},
		{
			name:     "wrong type",
			args:     map[string]any{"model": 123},
			key:      "model",
			defVal:   "default",
			expected: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.extractString(tt.args, tt.key, tt.defVal)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestParameterExtractor_ExtractFloat(t *testing.T) {
	extractor := NewParameterExtractor()

	tests := []struct {
		name     string
		args     map[string]any
		key      string
		defVal   float64
		expected float64
	}{
		{
			name:     "float present",
			args:     map[string]any{"temperature": 0.7},
			key:      "temperature",
			defVal:   1.0,
			expected: 0.7,
		},
		{
			name:     "float missing",
			args:     map[string]any{},
			key:      "temperature",
			defVal:   1.0,
			expected: 1.0,
		},
		{
			name:     "wrong type",
			args:     map[string]any{"temperature": "not a float"},
			key:      "temperature",
			defVal:   1.0,
			expected: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.extractFloat(tt.args, tt.key, tt.defVal)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestParameterExtractor_ExtractInt(t *testing.T) {
	extractor := NewParameterExtractor()

	tests := []struct {
		name     string
		args     map[string]any
		key      string
		defVal   int
		expected int
	}{
		{
			name:     "float converted to int",
			args:     map[string]any{"max_tokens": float64(100)},
			key:      "max_tokens",
			defVal:   50,
			expected: 100,
		},
		{
			name:     "int missing",
			args:     map[string]any{},
			key:      "max_tokens",
			defVal:   50,
			expected: 50,
		},
		{
			name:     "wrong type",
			args:     map[string]any{"max_tokens": "not a number"},
			key:      "max_tokens",
			defVal:   50,
			expected: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.extractInt(tt.args, tt.key, tt.defVal)
			if result != tt.expected {
				t.Errorf("Expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestParameterExtractor_ExtractBool(t *testing.T) {
	extractor := NewParameterExtractor()

	tests := []struct {
		name     string
		args     map[string]any
		key      string
		defVal   bool
		expected bool
	}{
		{
			name:     "bool true",
			args:     map[string]any{"stream": true},
			key:      "stream",
			defVal:   false,
			expected: true,
		},
		{
			name:     "bool false",
			args:     map[string]any{"stream": false},
			key:      "stream",
			defVal:   true,
			expected: false,
		},
		{
			name:     "bool missing",
			args:     map[string]any{},
			key:      "stream",
			defVal:   true,
			expected: true,
		},
		{
			name:     "wrong type",
			args:     map[string]any{"stream": "not a bool"},
			key:      "stream",
			defVal:   false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.extractBool(tt.args, tt.key, tt.defVal)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestParameterExtractor_ExtractStringSlice(t *testing.T) {
	extractor := NewParameterExtractor()

	tests := []struct {
		name     string
		args     map[string]any
		key      string
		expected []string
	}{
		{
			name:     "string slice present",
			args:     map[string]any{"domains": []any{"a.com", "b.com"}},
			key:      "domains",
			expected: []string{"a.com", "b.com"},
		},
		{
			name:     "empty slice",
			args:     map[string]any{"domains": []any{}},
			key:      "domains",
			expected: []string{},
		},
		{
			name:     "slice missing",
			args:     map[string]any{},
			key:      "domains",
			expected: nil,
		},
		{
			name:     "wrong type",
			args:     map[string]any{"domains": "not a slice"},
			key:      "domains",
			expected: nil,
		},
		{
			name:     "slice with mixed types",
			args:     map[string]any{"domains": []any{"a.com", 123, "b.com"}},
			key:      "domains",
			expected: []string{"a.com", "b.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.extractStringSlice(tt.args, tt.key)
			if !stringSliceEqual(result, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestParameterExtractor_ApplyDefaults(t *testing.T) {
	extractor := NewParameterExtractor()

	t.Run("applies all defaults", func(t *testing.T) {
		params := &QueryParams{
			UserPrompt: "test",
		}

		extractor.applyDefaults(params)

		if params.Model != perplexity.DefaultModel {
			t.Errorf("Expected model %q, got %q", perplexity.DefaultModel, params.Model)
		}
		if params.FrequencyPenalty != perplexity.DefaultFrequencyPenalty {
			t.Errorf("Expected frequency_penalty %v, got %v", perplexity.DefaultFrequencyPenalty, params.FrequencyPenalty)
		}
		if params.MaxTokens != perplexity.DefaultMaxTokens {
			t.Errorf("Expected max_tokens %d, got %d", perplexity.DefaultMaxTokens, params.MaxTokens)
		}
		if params.PresencePenalty != perplexity.DefaultPresencePenalty {
			t.Errorf("Expected presence_penalty %v, got %v", perplexity.DefaultPresencePenalty, params.PresencePenalty)
		}
		if params.Temperature != perplexity.DefaultTemperature {
			t.Errorf("Expected temperature %v, got %v", perplexity.DefaultTemperature, params.Temperature)
		}
		if params.TopK != perplexity.DefaultTopK {
			t.Errorf("Expected top_k %d, got %d", perplexity.DefaultTopK, params.TopK)
		}
		if params.TopP != perplexity.DefaultTopP {
			t.Errorf("Expected top_p %v, got %v", perplexity.DefaultTopP, params.TopP)
		}
		if params.Timeout != perplexity.DefaultTimeout {
			t.Errorf("Expected timeout %v, got %v", perplexity.DefaultTimeout, params.Timeout)
		}
	})

	t.Run("preserves non-zero values", func(t *testing.T) {
		params := &QueryParams{
			UserPrompt:       "test",
			Model:            "custom-model",
			FrequencyPenalty: 0.5,
			MaxTokens:        200,
			Temperature:      0.9,
		}

		extractor.applyDefaults(params)

		if params.Model != "custom-model" {
			t.Error("Model should not be overwritten")
		}
		if params.FrequencyPenalty != 0.5 {
			t.Error("FrequencyPenalty should not be overwritten")
		}
		if params.MaxTokens != 200 {
			t.Error("MaxTokens should not be overwritten")
		}
		if params.Temperature != 0.9 {
			t.Error("Temperature should not be overwritten")
		}
	})
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func stringSliceEqual(a, b []string) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
