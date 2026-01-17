package config

import (
	"testing"
)

func TestValidatorValidConfig(t *testing.T) {
	cfg := &ConfigData{
		Defaults: DefaultsConfig{
			Model:            "sonar",
			Temperature:      0.5,
			MaxTokens:        1000,
			TopK:             10,
			TopP:             0.9,
			FrequencyPenalty: 0.5,
			PresencePenalty:  0.5,
			Timeout:          "30s",
		},
		Search: SearchConfig{
			Recency:     "week",
			Mode:        "web",
			ContextSize: "medium",
			LocationLat: 40.7128,
			LocationLon: -74.0060,
			AfterDate:   "01/01/2024",
			BeforeDate:  "12/31/2024",
		},
		Output: OutputConfig{
			ReasoningEffort: "medium",
		},
	}

	validator := NewValidator()
	err := validator.Validate(cfg)
	if err != nil {
		t.Errorf("Valid config failed validation: %v", err)
	}
}

func TestValidatorInvalidTemperature(t *testing.T) {
	cfg := &ConfigData{
		Defaults: DefaultsConfig{
			Temperature: 3.0, // Invalid: > 2.0
		},
	}

	validator := NewValidator()
	err := validator.Validate(cfg)
	if err == nil {
		t.Error("Expected validation error for invalid temperature")
	}
}

func TestValidatorInvalidRecency(t *testing.T) {
	cfg := &ConfigData{
		Search: SearchConfig{
			Recency: "invalid",
		},
	}

	validator := NewValidator()
	err := validator.Validate(cfg)
	if err == nil {
		t.Error("Expected validation error for invalid recency")
	}
}

func TestValidatorInvalidMode(t *testing.T) {
	cfg := &ConfigData{
		Search: SearchConfig{
			Mode: "invalid",
		},
	}

	validator := NewValidator()
	err := validator.Validate(cfg)
	if err == nil {
		t.Error("Expected validation error for invalid mode")
	}
}

func TestValidatorInvalidCoordinates(t *testing.T) {
	cfg := &ConfigData{
		Search: SearchConfig{
			LocationLat: 100.0, // Invalid: > 90
			LocationLon: -200.0, // Invalid: < -180
		},
	}

	validator := NewValidator()
	err := validator.Validate(cfg)
	if err == nil {
		t.Error("Expected validation error for invalid coordinates")
	}
}

func TestValidatorInvalidDateFormat(t *testing.T) {
	cfg := &ConfigData{
		Search: SearchConfig{
			AfterDate: "2024-01-01", // Invalid format (should be MM/DD/YYYY)
		},
	}

	validator := NewValidator()
	err := validator.Validate(cfg)
	if err == nil {
		t.Error("Expected validation error for invalid date format")
	}
}

func TestValidatorProfileNameValidation(t *testing.T) {
	cfg := &ConfigData{
		Profiles: map[string]*Profile{
			"invalid name": { // Invalid: contains space
				Name: "invalid name",
			},
		},
	}

	validator := NewValidator()
	err := validator.Validate(cfg)
	if err == nil {
		t.Error("Expected validation error for invalid profile name")
	}
}

func TestValidatorMultipleErrors(t *testing.T) {
	cfg := &ConfigData{
		Defaults: DefaultsConfig{
			Temperature: 3.0,  // Invalid
			TopP:        1.5,  // Invalid
		},
		Search: SearchConfig{
			Recency: "invalid", // Invalid
		},
	}

	validator := NewValidator()
	err := validator.Validate(cfg)
	if err == nil {
		t.Error("Expected validation errors")
	}

	// Check that we got multiple errors
	if verrs, ok := err.(ValidationErrors); ok {
		if len(verrs) < 3 {
			t.Errorf("Expected at least 3 validation errors, got %d", len(verrs))
		}
	}
}

func TestValidatorInvalidAPIURL(t *testing.T) {
	cfg := &ConfigData{
		API: APIConfig{
			BaseURL: "invalid-url", // Missing http:// or https://
		},
	}

	validator := NewValidator()
	err := validator.Validate(cfg)
	if err == nil {
		t.Error("Expected validation error for invalid API URL")
	}
}

func TestValidatorAPIURLWithoutHost(t *testing.T) {
	cfg := &ConfigData{
		API: APIConfig{
			BaseURL: "http://", // Missing host
		},
	}

	validator := NewValidator()
	err := validator.Validate(cfg)
	if err == nil {
		t.Error("Expected validation error for URL without host")
	}
}

func TestValidatorAPIURLWithInvalidScheme(t *testing.T) {
	cfg := &ConfigData{
		API: APIConfig{
			BaseURL: "ftp://example.com", // Invalid scheme
		},
	}

	validator := NewValidator()
	err := validator.Validate(cfg)
	if err == nil {
		t.Error("Expected validation error for URL with invalid scheme")
	}
}

func TestValidatorAPIURLMalformed(t *testing.T) {
	testCases := []struct {
		name    string
		baseURL string
	}{
		{"invalid characters", "http://exa mple.com"},
		{"missing scheme separator", "httpexample.com"},
		{"incomplete URL", "http:/"},
		{"just scheme", "https://"},
		{"invalid format", "not a url at all"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &ConfigData{
				API: APIConfig{
					BaseURL: tc.baseURL,
				},
			}

			validator := NewValidator()
			err := validator.Validate(cfg)
			if err == nil {
				t.Errorf("Expected validation error for malformed URL: %s", tc.baseURL)
			}
		})
	}
}

func TestValidatorAPIURLValid(t *testing.T) {
	testCases := []struct {
		name    string
		baseURL string
	}{
		{"http URL", "http://example.com"},
		{"https URL", "https://example.com"},
		{"with port", "https://example.com:8080"},
		{"with path", "https://example.com/api"},
		{"with query", "https://example.com/api?key=value"},
		{"localhost", "http://localhost:3000"},
		{"IP address", "http://192.168.1.1"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &ConfigData{
				API: APIConfig{
					BaseURL: tc.baseURL,
				},
			}

			validator := NewValidator()
			err := validator.Validate(cfg)
			if err != nil {
				t.Errorf("Valid URL should pass validation: %s, error: %v", tc.baseURL, err)
			}
		})
	}
}

func TestValidatorNonExistentActiveProfile(t *testing.T) {
	cfg := &ConfigData{
		ActiveProfile: "nonexistent",
		Profiles:      make(map[string]*Profile),
	}

	validator := NewValidator()
	err := validator.Validate(cfg)
	if err == nil {
		t.Error("Expected validation error for non-existent active profile")
	}
}

func TestValidatorDefaultProfileIsValid(t *testing.T) {
	cfg := &ConfigData{
		ActiveProfile: "default",
		Profiles:      make(map[string]*Profile),
	}

	validator := NewValidator()
	err := validator.Validate(cfg)
	if err != nil {
		t.Errorf("Default profile should be valid: %v", err)
	}
}
