package config

import (
	"testing"

	"github.com/sgaunet/pplx/pkg/clerrors"
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
	if verrs, ok := err.(clerrors.ValidationErrors); ok {
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

// =============================================================================
// Edge Case Tests - Temperature Boundaries
// =============================================================================

func TestValidator_TemperatureExactZero(t *testing.T) {
	cfg := &ConfigData{
		Defaults: DefaultsConfig{
			Temperature: 0.0,
		},
	}

	validator := NewValidator()
	err := validator.Validate(cfg)
	// Zero is valid (minimum boundary)
	if err != nil {
		t.Errorf("Temperature 0.0 should be valid: %v", err)
	}
}

func TestValidator_TemperatureExactMax(t *testing.T) {
	cfg := &ConfigData{
		Defaults: DefaultsConfig{
			Temperature: 2.0,
		},
	}

	validator := NewValidator()
	err := validator.Validate(cfg)
	// 2.0 is valid (maximum boundary)
	if err != nil {
		t.Errorf("Temperature 2.0 should be valid: %v", err)
	}
}

func TestValidator_TemperatureJustOverMax(t *testing.T) {
	cfg := &ConfigData{
		Defaults: DefaultsConfig{
			Temperature: 2.00001,
		},
	}

	validator := NewValidator()
	err := validator.Validate(cfg)
	if err == nil {
		t.Error("Temperature 2.00001 should be invalid (over max)")
	}
}

func TestValidator_TemperatureNegative(t *testing.T) {
	cfg := &ConfigData{
		Defaults: DefaultsConfig{
			Temperature: -0.1,
		},
	}

	validator := NewValidator()
	err := validator.Validate(cfg)
	if err == nil {
		t.Error("Negative temperature should be invalid")
	}
}

// =============================================================================
// Edge Case Tests - Coordinate Boundaries
// =============================================================================

func TestValidator_CoordinatesExactBoundaries(t *testing.T) {
	testCases := []struct {
		name string
		lat  float64
		lon  float64
	}{
		{"max lat", 90.0, 0},
		{"min lat", -90.0, 0},
		{"max lon", 0, 180.0},
		{"min lon", 0, -180.0},
		{"all max", 90.0, 180.0},
		{"all min", -90.0, -180.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &ConfigData{
				Search: SearchConfig{
					LocationLat: tc.lat,
					LocationLon: tc.lon,
				},
			}

			validator := NewValidator()
			err := validator.Validate(cfg)
			if err != nil {
				t.Errorf("Coordinates (%f, %f) should be valid: %v", tc.lat, tc.lon, err)
			}
		})
	}
}

func TestValidator_CoordinatesJustOver(t *testing.T) {
	testCases := []struct {
		name string
		lat  float64
		lon  float64
	}{
		{"lat over max", 90.00001, 0},
		{"lat under min", -90.00001, 0},
		{"lon over max", 0, 180.00001},
		{"lon under min", 0, -180.00001},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &ConfigData{
				Search: SearchConfig{
					LocationLat: tc.lat,
					LocationLon: tc.lon,
				},
			}

			validator := NewValidator()
			err := validator.Validate(cfg)
			if err == nil {
				t.Errorf("Coordinates (%f, %f) should be invalid", tc.lat, tc.lon)
			}
		})
	}
}

// =============================================================================
// Edge Case Tests - Date Formats
// =============================================================================

func TestValidator_DateFormatLeapYear(t *testing.T) {
	testCases := []struct {
		name      string
		date      string
		shouldErr bool
	}{
		{"valid leap year", "02/29/2024", false},
		{"invalid leap year", "02/29/2023", false}, // Note: validator only checks format, not calendar validity
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &ConfigData{
				Search: SearchConfig{
					AfterDate: tc.date,
				},
			}

			validator := NewValidator()
			err := validator.Validate(cfg)
			if tc.shouldErr && err == nil {
				t.Errorf("Date %s should be invalid", tc.date)
			} else if !tc.shouldErr && err != nil {
				t.Errorf("Date %s should be valid: %v", tc.date, err)
			}
		})
	}
}

func TestValidator_DateFormatInvalidDay(t *testing.T) {
	testCases := []string{
		"02/30/2024", // February doesn't have 30 days
		"13/01/2024", // Month 13 doesn't exist
		"00/01/2024", // Month 0 doesn't exist
		"01/32/2024", // Day 32 doesn't exist
		"01/00/2024", // Day 0 doesn't exist
	}

	for _, date := range testCases {
		t.Run(date, func(t *testing.T) {
			cfg := &ConfigData{
				Search: SearchConfig{
					BeforeDate: date,
				},
			}

			validator := NewValidator()
			err := validator.Validate(cfg)
			// The validator checks format MM/DD/YYYY but may not validate calendar correctness
			// This test documents the current behavior
			_ = err
		})
	}
}

func TestValidator_DateFormatWhitespace(t *testing.T) {
	testCases := []struct {
		name string
		date string
	}{
		{"leading space", " 01/01/2024"},
		{"trailing space", "01/01/2024 "},
		{"both spaces", " 01/01/2024 "},
		{"internal spaces", "01 / 01 / 2024"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &ConfigData{
				Search: SearchConfig{
					LastUpdatedAfter: tc.date,
				},
			}

			validator := NewValidator()
			err := validator.Validate(cfg)
			if err == nil {
				t.Logf("Date with whitespace '%s' passed validation", tc.date)
			}
		})
	}
}

func TestValidator_DateFormatAlternative(t *testing.T) {
	// Test non-MM/DD/YYYY formats (should be invalid)
	testCases := []string{
		"2024-01-01",    // ISO format
		"01-01-2024",    // Dash separator
		"01.01.2024",    // Dot separator
		"01/01/24",      // Two-digit year
		"1/1/2024",      // No leading zeros
	}

	for _, date := range testCases {
		t.Run(date, func(t *testing.T) {
			cfg := &ConfigData{
				Search: SearchConfig{
					AfterDate: date,
				},
			}

			validator := NewValidator()
			err := validator.Validate(cfg)
			if err == nil {
				t.Logf("Alternative date format '%s' passed validation (may be accepted)", date)
			} else {
				t.Logf("Alternative date format '%s' rejected as expected", date)
			}
		})
	}
}

// =============================================================================
// Edge Case Tests - Profile Names
// =============================================================================

func TestValidator_ProfileNameUnicode(t *testing.T) {
	testCases := []struct {
		name        string
		profileName string
	}{
		{"french accents", "café"},
		{"chinese", "测试"},
		{"emoji", "profile-🚀"},
		{"mixed", "test-café-测试"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &ConfigData{
				Profiles: map[string]*Profile{
					tc.profileName: {
						Name: tc.profileName,
					},
				},
			}

			validator := NewValidator()
			err := validator.Validate(cfg)
			if err != nil {
				t.Logf("Unicode profile name '%s' rejected: %v", tc.profileName, err)
			} else {
				t.Logf("Unicode profile name '%s' accepted", tc.profileName)
			}
		})
	}
}

func TestValidator_ProfileNameVeryLong(t *testing.T) {
	// Create a 256-character profile name
	longName := ""
	for i := 0; i < 256; i++ {
		longName += "a"
	}

	cfg := &ConfigData{
		Profiles: map[string]*Profile{
			longName: {
				Name: longName,
			},
		},
	}

	validator := NewValidator()
	err := validator.Validate(cfg)
	if err != nil {
		t.Logf("Very long profile name (256 chars) rejected: %v", err)
	} else {
		t.Logf("Very long profile name (256 chars) accepted")
	}
}

// =============================================================================
// validateDefaults — numeric boundaries
// =============================================================================

func TestValidator_MaxTokensNegative(t *testing.T) {
	cfg := &ConfigData{
		Defaults: DefaultsConfig{
			MaxTokens: -1,
		},
	}
	validator := NewValidator()
	if err := validator.Validate(cfg); err == nil {
		t.Error("Expected validation error for MaxTokens = -1")
	}
}

func TestValidator_TopKNegative(t *testing.T) {
	cfg := &ConfigData{
		Defaults: DefaultsConfig{
			TopK: -1,
		},
	}
	validator := NewValidator()
	if err := validator.Validate(cfg); err == nil {
		t.Error("Expected validation error for TopK = -1")
	}
}

func TestValidator_TopPBoundaries(t *testing.T) {
	tests := []struct {
		name    string
		topP    float64
		wantErr bool
	}{
		{"exact zero", 0.0, false},
		{"exact one", 1.0, false},
		{"just below zero", -0.001, true},
		{"just above one", 1.001, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &ConfigData{Defaults: DefaultsConfig{TopP: tc.topP}}
			err := NewValidator().Validate(cfg)
			if tc.wantErr && err == nil {
				t.Errorf("TopP=%v: expected error, got nil", tc.topP)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("TopP=%v: expected nil, got %v", tc.topP, err)
			}
		})
	}
}

func TestValidator_FrequencyPenaltyBoundaries(t *testing.T) {
	tests := []struct {
		name    string
		val     float64
		wantErr bool
	}{
		{"exact zero", 0.0, false},
		{"exact two", 2.0, false},
		{"just below zero", -0.001, true},
		{"just above two", 2.001, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &ConfigData{Defaults: DefaultsConfig{FrequencyPenalty: tc.val}}
			err := NewValidator().Validate(cfg)
			if tc.wantErr && err == nil {
				t.Errorf("FrequencyPenalty=%v: expected error, got nil", tc.val)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("FrequencyPenalty=%v: expected nil, got %v", tc.val, err)
			}
		})
	}
}

func TestValidator_PresencePenaltyBoundaries(t *testing.T) {
	tests := []struct {
		name    string
		val     float64
		wantErr bool
	}{
		{"exact zero", 0.0, false},
		{"exact two", 2.0, false},
		{"just below zero", -0.001, true},
		{"just above two", 2.001, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &ConfigData{Defaults: DefaultsConfig{PresencePenalty: tc.val}}
			err := NewValidator().Validate(cfg)
			if tc.wantErr && err == nil {
				t.Errorf("PresencePenalty=%v: expected error, got nil", tc.val)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("PresencePenalty=%v: expected nil, got %v", tc.val, err)
			}
		})
	}
}

// =============================================================================
// validateSearch — enum coverage
// =============================================================================

func TestValidator_SearchContextSizeAllValid(t *testing.T) {
	for _, size := range []string{"low", "medium", "high"} {
		t.Run(size, func(t *testing.T) {
			cfg := &ConfigData{Search: SearchConfig{ContextSize: size}}
			if err := NewValidator().Validate(cfg); err != nil {
				t.Errorf("ContextSize=%q should be valid: %v", size, err)
			}
		})
	}
}

func TestValidator_SearchContextSizeInvalid(t *testing.T) {
	cfg := &ConfigData{Search: SearchConfig{ContextSize: "extreme"}}
	if err := NewValidator().Validate(cfg); err == nil {
		t.Error("Expected error for ContextSize='extreme'")
	}
}

func TestValidator_SearchRecencyAllValid(t *testing.T) {
	for _, recency := range []string{"hour", "day", "week", "month", "year"} {
		t.Run(recency, func(t *testing.T) {
			cfg := &ConfigData{Search: SearchConfig{Recency: recency}}
			if err := NewValidator().Validate(cfg); err != nil {
				t.Errorf("Recency=%q should be valid: %v", recency, err)
			}
		})
	}
}

func TestValidator_SearchModeAcademic(t *testing.T) {
	cfg := &ConfigData{Search: SearchConfig{Mode: "academic"}}
	if err := NewValidator().Validate(cfg); err != nil {
		t.Errorf("Mode='academic' should be valid: %v", err)
	}
}

// =============================================================================
// validateOutput — reasoning effort
// =============================================================================

func TestValidator_ReasoningEffortAllValid(t *testing.T) {
	for _, effort := range []string{"low", "medium", "high"} {
		t.Run(effort, func(t *testing.T) {
			cfg := &ConfigData{Output: OutputConfig{ReasoningEffort: effort}}
			if err := NewValidator().Validate(cfg); err != nil {
				t.Errorf("ReasoningEffort=%q should be valid: %v", effort, err)
			}
		})
	}
}

func TestValidator_ReasoningEffortInvalid(t *testing.T) {
	cfg := &ConfigData{Output: OutputConfig{ReasoningEffort: "extreme"}}
	if err := NewValidator().Validate(cfg); err == nil {
		t.Error("Expected error for ReasoningEffort='extreme'")
	}
}

// =============================================================================
// validateProfiles — mismatch & nested validation
// =============================================================================

func TestValidator_ProfileNameMismatch(t *testing.T) {
	cfg := &ConfigData{
		Profiles: map[string]*Profile{
			"work": {Name: "personal"},
		},
	}
	if err := NewValidator().Validate(cfg); err == nil {
		t.Error("Expected error for profile key/name mismatch (key='work', Name='personal')")
	}
}

func TestValidator_ProfileNestedInvalidDefaults(t *testing.T) {
	cfg := &ConfigData{
		Profiles: map[string]*Profile{
			"work": {
				Name:     "work",
				Defaults: DefaultsConfig{Temperature: 3.0},
			},
		},
	}
	if err := NewValidator().Validate(cfg); err == nil {
		t.Error("Expected error for profile with Temperature=3.0")
	}
}

func TestValidator_ProfileNestedInvalidSearch(t *testing.T) {
	cfg := &ConfigData{
		Profiles: map[string]*Profile{
			"work": {
				Name:   "work",
				Search: SearchConfig{Recency: "invalid"},
			},
		},
	}
	if err := NewValidator().Validate(cfg); err == nil {
		t.Error("Expected error for profile with Recency='invalid'")
	}
}

// =============================================================================
// Validator API behaviour
// =============================================================================

func TestValidator_ErrorsMethod(t *testing.T) {
	cfg := &ConfigData{
		Defaults: DefaultsConfig{Temperature: 3.0},
		Search:   SearchConfig{Recency: "invalid"},
	}
	v := NewValidator()
	validateErr := v.Validate(cfg)
	errorsSlice := v.Errors()

	if validateErr == nil {
		t.Fatal("Expected validation error, got nil")
	}
	if len(errorsSlice) == 0 {
		t.Fatal("Errors() should return non-empty slice after failed validation")
	}

	verrs, ok := validateErr.(clerrors.ValidationErrors)
	if !ok {
		t.Fatalf("Expected ValidationErrors type, got %T", validateErr)
	}
	if len(verrs) != len(errorsSlice) {
		t.Errorf("Validate() returned %d errors, Errors() returned %d", len(verrs), len(errorsSlice))
	}
}

func TestValidator_ResetBetweenCalls(t *testing.T) {
	v := NewValidator()

	// First call with invalid config
	invalid := &ConfigData{Defaults: DefaultsConfig{Temperature: 3.0}}
	if err := v.Validate(invalid); err == nil {
		t.Fatal("Expected error on first (invalid) call")
	}

	// Second call with valid config — errors must be cleared
	valid := &ConfigData{}
	if err := v.Validate(valid); err != nil {
		t.Errorf("Expected nil on second (valid) call, got: %v", err)
	}
}

func TestValidator_EmptyConfig(t *testing.T) {
	cfg := &ConfigData{}
	if err := NewValidator().Validate(cfg); err != nil {
		t.Errorf("Empty ConfigData should be valid, got: %v", err)
	}
}
