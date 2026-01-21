package config

import (
	"os"
	"strings"
	"testing"
)

// setupEnvTest sets up environment variables for a test and returns cleanup function.
func setupEnvTest(t *testing.T, vars map[string]string) func() {
	t.Helper()

	// Store original values
	originals := make(map[string]string)
	for key := range vars {
		if val, exists := os.LookupEnv(key); exists {
			originals[key] = val
		}
	}

	// Set new values
	for key, val := range vars {
		if err := os.Setenv(key, val); err != nil {
			t.Fatalf("Failed to set env var %s: %v", key, err)
		}
	}

	// Return cleanup function
	return func() {
		for key := range vars {
			if orig, had := originals[key]; had {
				_ = os.Setenv(key, orig)
			} else {
				_ = os.Unsetenv(key)
			}
		}
	}
}

// createTestConfig creates a ConfigData with env var placeholders for testing.
func createTestConfig() *ConfigData {
	return &ConfigData{
		API: APIConfig{
			Key:     "$API_KEY",
			BaseURL: "${BASE_URL}",
		},
		Defaults: DefaultsConfig{
			Model:   "${MODEL}",
			Timeout: "$TIMEOUT",
		},
		Search: SearchConfig{
			Domains:         []string{"$DOMAIN1", "${DOMAIN2}", "literal.com"},
			LocationCountry: "${COUNTRY}",
		},
		Output: OutputConfig{
			ImageDomains: []string{"${IMG_DOMAIN1}", "$IMG_DOMAIN2"},
			ImageFormats: []string{"$FORMAT1", "png"},
		},
	}
}

// =============================================================================
// Category 1: Basic Syntax Variants (4 tests)
// =============================================================================

func TestExpandEnvVars_SimpleDollarSyntax(t *testing.T) {
	cleanup := setupEnvTest(t, map[string]string{
		"TEST_VAR": "test-value",
	})
	defer cleanup()

	cfg := &ConfigData{
		API: APIConfig{
			Key: "$TEST_VAR",
		},
	}

	ExpandEnvVars(cfg)

	if cfg.API.Key != "test-value" {
		t.Errorf("Expected 'test-value', got '%s'", cfg.API.Key)
	}
}

func TestExpandEnvVars_BracedSyntax(t *testing.T) {
	cleanup := setupEnvTest(t, map[string]string{
		"TEST_VAR": "braced-value",
	})
	defer cleanup()

	cfg := &ConfigData{
		API: APIConfig{
			BaseURL: "${TEST_VAR}",
		},
	}

	ExpandEnvVars(cfg)

	if cfg.API.BaseURL != "braced-value" {
		t.Errorf("Expected 'braced-value', got '%s'", cfg.API.BaseURL)
	}
}

func TestExpandEnvVars_MixedSyntax(t *testing.T) {
	cleanup := setupEnvTest(t, map[string]string{
		"VAR1": "first",
		"VAR2": "second",
	})
	defer cleanup()

	cfg := &ConfigData{
		Defaults: DefaultsConfig{
			Model: "$VAR1 and ${VAR2}",
		},
	}

	ExpandEnvVars(cfg)

	expected := "first and second"
	if cfg.Defaults.Model != expected {
		t.Errorf("Expected '%s', got '%s'", expected, cfg.Defaults.Model)
	}
}

func TestExpandEnvVars_MultipleVarsInString(t *testing.T) {
	cleanup := setupEnvTest(t, map[string]string{
		"HOST": "example.com",
		"PORT": "8080",
		"PATH": "/api/v1",
	})
	defer cleanup()

	cfg := &ConfigData{
		API: APIConfig{
			BaseURL: "https://$HOST:$PORT$PATH",
		},
	}

	ExpandEnvVars(cfg)

	expected := "https://example.com:8080/api/v1"
	if cfg.API.BaseURL != expected {
		t.Errorf("Expected '%s', got '%s'", expected, cfg.API.BaseURL)
	}
}

// =============================================================================
// Category 2: Undefined/Empty Variables (3 tests)
// =============================================================================

func TestExpandEnvVars_UndefinedVariable(t *testing.T) {
	// Ensure variable doesn't exist
	_ = os.Unsetenv("UNDEFINED_VAR_12345")

	cfg := &ConfigData{
		API: APIConfig{
			Key: "$UNDEFINED_VAR_12345",
		},
	}

	ExpandEnvVars(cfg)

	// os.ExpandEnv replaces undefined vars with empty string
	if cfg.API.Key != "" {
		t.Errorf("Expected empty string for undefined var, got '%s'", cfg.API.Key)
	}
}

func TestExpandEnvVars_EmptyVariable(t *testing.T) {
	cleanup := setupEnvTest(t, map[string]string{
		"EMPTY_VAR": "",
	})
	defer cleanup()

	cfg := &ConfigData{
		Defaults: DefaultsConfig{
			Model: "${EMPTY_VAR}",
		},
	}

	ExpandEnvVars(cfg)

	if cfg.Defaults.Model != "" {
		t.Errorf("Expected empty string, got '%s'", cfg.Defaults.Model)
	}
}

func TestExpandEnvVars_WhitespaceOnlyValue(t *testing.T) {
	cleanup := setupEnvTest(t, map[string]string{
		"WHITESPACE_VAR": "   ",
	})
	defer cleanup()

	cfg := &ConfigData{
		Search: SearchConfig{
			LocationCountry: "$WHITESPACE_VAR",
		},
	}

	ExpandEnvVars(cfg)

	if cfg.Search.LocationCountry != "   " {
		t.Errorf("Expected '   ', got '%s'", cfg.Search.LocationCountry)
	}
}

// =============================================================================
// Category 3: Special Characters in Values (4 tests)
// =============================================================================

func TestExpandEnvVars_URLsWithProtocols(t *testing.T) {
	cleanup := setupEnvTest(t, map[string]string{
		"API_URL": "https://api.perplexity.ai/v1",
	})
	defer cleanup()

	cfg := &ConfigData{
		API: APIConfig{
			BaseURL: "${API_URL}",
		},
	}

	ExpandEnvVars(cfg)

	expected := "https://api.perplexity.ai/v1"
	if cfg.API.BaseURL != expected {
		t.Errorf("Expected '%s', got '%s'", expected, cfg.API.BaseURL)
	}
}

func TestExpandEnvVars_SpecialCharacters(t *testing.T) {
	cleanup := setupEnvTest(t, map[string]string{
		"SPECIAL_CHARS": "test@#%&*!()[]",
	})
	defer cleanup()

	cfg := &ConfigData{
		API: APIConfig{
			Key: "$SPECIAL_CHARS",
		},
	}

	ExpandEnvVars(cfg)

	expected := "test@#%&*!()[]"
	if cfg.API.Key != expected {
		t.Errorf("Expected '%s', got '%s'", expected, cfg.API.Key)
	}
}

func TestExpandEnvVars_QuotesInValues(t *testing.T) {
	testCases := []struct {
		name     string
		envValue string
	}{
		{"single quotes", "test'value'here"},
		{"double quotes", `test"value"here`},
		{"mixed quotes", `test'mixed"quotes`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cleanup := setupEnvTest(t, map[string]string{
				"QUOTE_VAR": tc.envValue,
			})
			defer cleanup()

			cfg := &ConfigData{
				Defaults: DefaultsConfig{
					Model: "${QUOTE_VAR}",
				},
			}

			ExpandEnvVars(cfg)

			if cfg.Defaults.Model != tc.envValue {
				t.Errorf("Expected '%s', got '%s'", tc.envValue, cfg.Defaults.Model)
			}
		})
	}
}

func TestExpandEnvVars_UnicodeCharacters(t *testing.T) {
	testCases := []struct {
		name     string
		envValue string
	}{
		{"emoji", "test-🚀-value"},
		{"accented", "café-résumé"},
		{"chinese", "测试值"},
		{"mixed", "test-café-🚀-测试"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cleanup := setupEnvTest(t, map[string]string{
				"UNICODE_VAR": tc.envValue,
			})
			defer cleanup()

			cfg := &ConfigData{
				Search: SearchConfig{
					LocationCountry: "$UNICODE_VAR",
				},
			}

			ExpandEnvVars(cfg)

			if cfg.Search.LocationCountry != tc.envValue {
				t.Errorf("Expected '%s', got '%s'", tc.envValue, cfg.Search.LocationCountry)
			}
		})
	}
}

// =============================================================================
// Category 4: Nested Structures (4 tests)
// =============================================================================

func TestExpandEnvVars_InStringArrays(t *testing.T) {
	cleanup := setupEnvTest(t, map[string]string{
		"DOMAIN1": "example.com",
		"DOMAIN2": "test.org",
	})
	defer cleanup()

	cfg := &ConfigData{
		Search: SearchConfig{
			Domains: []string{"$DOMAIN1", "${DOMAIN2}", "literal.com"},
		},
	}

	ExpandEnvVars(cfg)

	expected := []string{"example.com", "test.org", "literal.com"}
	if len(cfg.Search.Domains) != len(expected) {
		t.Fatalf("Expected %d domains, got %d", len(expected), len(cfg.Search.Domains))
	}

	for i, exp := range expected {
		if cfg.Search.Domains[i] != exp {
			t.Errorf("Domain[%d]: expected '%s', got '%s'", i, exp, cfg.Search.Domains[i])
		}
	}
}

func TestExpandEnvVars_InAllConfigSections(t *testing.T) {
	cleanup := setupEnvTest(t, map[string]string{
		"API_KEY":    "sk-test-123",
		"BASE_URL":   "https://api.test.com",
		"MODEL":      "sonar-pro",
		"TIMEOUT":    "30s",
		"DOMAIN":     "example.com",
		"COUNTRY":    "US",
		"IMG_DOMAIN": "images.test.com",
		"FORMAT":     "jpg",
	})
	defer cleanup()

	cfg := &ConfigData{
		API: APIConfig{
			Key:     "$API_KEY",
			BaseURL: "${BASE_URL}",
		},
		Defaults: DefaultsConfig{
			Model:   "$MODEL",
			Timeout: "${TIMEOUT}",
		},
		Search: SearchConfig{
			Domains:         []string{"$DOMAIN"},
			LocationCountry: "$COUNTRY",
		},
		Output: OutputConfig{
			ImageDomains: []string{"${IMG_DOMAIN}"},
			ImageFormats: []string{"$FORMAT"},
		},
	}

	ExpandEnvVars(cfg)

	// Verify API section
	if cfg.API.Key != "sk-test-123" {
		t.Errorf("API.Key: expected 'sk-test-123', got '%s'", cfg.API.Key)
	}
	if cfg.API.BaseURL != "https://api.test.com" {
		t.Errorf("API.BaseURL: expected 'https://api.test.com', got '%s'", cfg.API.BaseURL)
	}

	// Verify Defaults section
	if cfg.Defaults.Model != "sonar-pro" {
		t.Errorf("Defaults.Model: expected 'sonar-pro', got '%s'", cfg.Defaults.Model)
	}
	if cfg.Defaults.Timeout != "30s" {
		t.Errorf("Defaults.Timeout: expected '30s', got '%s'", cfg.Defaults.Timeout)
	}

	// Verify Search section
	if len(cfg.Search.Domains) != 1 || cfg.Search.Domains[0] != "example.com" {
		t.Errorf("Search.Domains: expected ['example.com'], got %v", cfg.Search.Domains)
	}
	if cfg.Search.LocationCountry != "US" {
		t.Errorf("Search.LocationCountry: expected 'US', got '%s'", cfg.Search.LocationCountry)
	}

	// Verify Output section
	if len(cfg.Output.ImageDomains) != 1 || cfg.Output.ImageDomains[0] != "images.test.com" {
		t.Errorf("Output.ImageDomains: expected ['images.test.com'], got %v", cfg.Output.ImageDomains)
	}
	if len(cfg.Output.ImageFormats) != 1 || cfg.Output.ImageFormats[0] != "jpg" {
		t.Errorf("Output.ImageFormats: expected ['jpg'], got %v", cfg.Output.ImageFormats)
	}
}

func TestExpandEnvVars_PreservesNonEnvValues(t *testing.T) {
	cleanup := setupEnvTest(t, map[string]string{
		"TEST_VAR": "replaced",
	})
	defer cleanup()

	cfg := &ConfigData{
		API: APIConfig{
			Key:     "$TEST_VAR",
			BaseURL: "https://literal.com",
		},
		Defaults: DefaultsConfig{
			Model:   "literal-model",
			Timeout: "30s",
		},
	}

	ExpandEnvVars(cfg)

	// Verify env var was expanded
	if cfg.API.Key != "replaced" {
		t.Errorf("API.Key should be expanded, got '%s'", cfg.API.Key)
	}

	// Verify literals were preserved
	if cfg.API.BaseURL != "https://literal.com" {
		t.Errorf("API.BaseURL should be preserved, got '%s'", cfg.API.BaseURL)
	}
	if cfg.Defaults.Model != "literal-model" {
		t.Errorf("Defaults.Model should be preserved, got '%s'", cfg.Defaults.Model)
	}
	if cfg.Defaults.Timeout != "30s" {
		t.Errorf("Defaults.Timeout should be preserved, got '%s'", cfg.Defaults.Timeout)
	}
}

func TestExpandEnvVars_MixedArrayContent(t *testing.T) {
	cleanup := setupEnvTest(t, map[string]string{
		"DYNAMIC1": "dynamic-first.com",
		"DYNAMIC2": "dynamic-second.com",
	})
	defer cleanup()

	cfg := &ConfigData{
		Search: SearchConfig{
			Domains: []string{
				"$DYNAMIC1",
				"literal.com",
				"${DYNAMIC2}",
				"another-literal.org",
			},
		},
	}

	ExpandEnvVars(cfg)

	expected := []string{
		"dynamic-first.com",
		"literal.com",
		"dynamic-second.com",
		"another-literal.org",
	}

	if len(cfg.Search.Domains) != len(expected) {
		t.Fatalf("Expected %d domains, got %d", len(expected), len(cfg.Search.Domains))
	}

	for i, exp := range expected {
		if cfg.Search.Domains[i] != exp {
			t.Errorf("Domain[%d]: expected '%s', got '%s'", i, exp, cfg.Search.Domains[i])
		}
	}
}

// =============================================================================
// Category 5: Edge Cases (5 tests)
// =============================================================================

func TestExpandEnvVars_MalformedSyntax(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty braces", "${}", ""},                   // os.ExpandEnv expands ${} to empty string
		{"unclosed brace", "${VAR", "VAR"},            // os.ExpandEnv treats as literal after $
		{"lone dollar", "test $ value", "test $ value"}, // Lone $ preserved
		{"dollar at end", "test$", "test$"},           // $ at end preserved
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &ConfigData{
				API: APIConfig{
					Key: tc.input,
				},
			}

			ExpandEnvVars(cfg)

			if cfg.API.Key != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, cfg.API.Key)
			}
		})
	}
}

func TestExpandEnvVars_EscapedDollarSigns(t *testing.T) {
	// Note: os.ExpandEnv behavior with $$ varies by context
	// In practice, $$VAR becomes "VAR" (the $$ escapes the variable expansion)
	cfg := &ConfigData{
		API: APIConfig{
			Key: "$$VAR",
		},
	}

	ExpandEnvVars(cfg)

	// os.ExpandEnv behavior: $$VAR → VAR (escaped expansion)
	expected := "VAR"
	if cfg.API.Key != expected {
		t.Errorf("Expected '%s', got '%s'", expected, cfg.API.Key)
	}
}

func TestExpandEnvVars_VeryLongValues(t *testing.T) {
	// Create a >1KB value
	longValue := strings.Repeat("x", 2000)

	cleanup := setupEnvTest(t, map[string]string{
		"LONG_VAR": longValue,
	})
	defer cleanup()

	cfg := &ConfigData{
		API: APIConfig{
			Key: "${LONG_VAR}",
		},
	}

	ExpandEnvVars(cfg)

	if len(cfg.API.Key) != 2000 {
		t.Errorf("Expected length 2000, got %d", len(cfg.API.Key))
	}
	if cfg.API.Key != longValue {
		t.Error("Long value was not preserved correctly")
	}
}

func TestExpandEnvVars_VarNamesWithNumbers(t *testing.T) {
	cleanup := setupEnvTest(t, map[string]string{
		"VAR1":      "first",
		"VAR2":      "second",
		"API_KEY_2": "key-value",
	})
	defer cleanup()

	cfg := &ConfigData{
		API: APIConfig{
			Key:     "$VAR1-$VAR2",
			BaseURL: "${API_KEY_2}",
		},
	}

	ExpandEnvVars(cfg)

	if cfg.API.Key != "first-second" {
		t.Errorf("API.Key: expected 'first-second', got '%s'", cfg.API.Key)
	}
	if cfg.API.BaseURL != "key-value" {
		t.Errorf("API.BaseURL: expected 'key-value', got '%s'", cfg.API.BaseURL)
	}
}

func TestExpandEnvVars_VarNamesWithUnderscores(t *testing.T) {
	cleanup := setupEnvTest(t, map[string]string{
		"MY_API_KEY":    "secret123",
		"DB_CONNECTION": "postgres://localhost",
	})
	defer cleanup()

	cfg := &ConfigData{
		API: APIConfig{
			Key:     "${MY_API_KEY}",
			BaseURL: "$DB_CONNECTION",
		},
	}

	ExpandEnvVars(cfg)

	if cfg.API.Key != "secret123" {
		t.Errorf("API.Key: expected 'secret123', got '%s'", cfg.API.Key)
	}
	if cfg.API.BaseURL != "postgres://localhost" {
		t.Errorf("API.BaseURL: expected 'postgres://localhost', got '%s'", cfg.API.BaseURL)
	}
}

// =============================================================================
// Category 6: Integration (2 tests)
// =============================================================================

func TestExpandEnvVars_FullConfigExpansion(t *testing.T) {
	cleanup := setupEnvTest(t, map[string]string{
		"API_KEY":     "sk-production-key",
		"BASE_URL":    "https://prod.api.com",
		"MODEL":       "sonar-professional",
		"TIMEOUT":     "60s",
		"DOMAIN1":     "prod1.example.com",
		"DOMAIN2":     "prod2.example.com",
		"COUNTRY":     "US",
		"IMG_DOMAIN1": "images1.cdn.com",
		"IMG_DOMAIN2": "images2.cdn.com",
		"FORMAT1":     "webp",
	})
	defer cleanup()

	cfg := createTestConfig()

	// Verify config has placeholders before expansion
	if !strings.Contains(cfg.API.Key, "$") {
		t.Error("Config should have $ placeholders before expansion")
	}

	ExpandEnvVars(cfg)

	// Verify all sections expanded correctly
	if cfg.API.Key != "sk-production-key" {
		t.Errorf("API.Key not expanded correctly: %s", cfg.API.Key)
	}
	if cfg.API.BaseURL != "https://prod.api.com" {
		t.Errorf("API.BaseURL not expanded correctly: %s", cfg.API.BaseURL)
	}
	if cfg.Defaults.Model != "sonar-professional" {
		t.Errorf("Defaults.Model not expanded correctly: %s", cfg.Defaults.Model)
	}
	if cfg.Defaults.Timeout != "60s" {
		t.Errorf("Defaults.Timeout not expanded correctly: %s", cfg.Defaults.Timeout)
	}

	// Verify arrays expanded correctly
	expectedDomains := []string{"prod1.example.com", "prod2.example.com", "literal.com"}
	if len(cfg.Search.Domains) != 3 {
		t.Errorf("Expected 3 domains, got %d", len(cfg.Search.Domains))
	}
	for i, exp := range expectedDomains {
		if cfg.Search.Domains[i] != exp {
			t.Errorf("Domain[%d]: expected '%s', got '%s'", i, exp, cfg.Search.Domains[i])
		}
	}

	if cfg.Search.LocationCountry != "US" {
		t.Errorf("LocationCountry not expanded correctly: %s", cfg.Search.LocationCountry)
	}

	// Verify no $ symbols remain (except in literal.com)
	if strings.Contains(cfg.API.Key, "$") ||
		strings.Contains(cfg.API.BaseURL, "$") ||
		strings.Contains(cfg.Defaults.Model, "$") ||
		strings.Contains(cfg.Defaults.Timeout, "$") {
		t.Error("Found unexpanded $ symbols in config")
	}
}

func TestExpandEnvVars_EmptyConfigData(t *testing.T) {
	// Test that ExpandEnvVars doesn't crash on empty/nil values
	cfg := &ConfigData{}

	// Should not panic
	ExpandEnvVars(cfg)

	// Verify config remains empty
	if cfg.API.Key != "" {
		t.Error("Empty config should remain empty")
	}
	if cfg.Defaults.Model != "" {
		t.Error("Empty config should remain empty")
	}
	if len(cfg.Search.Domains) != 0 {
		t.Error("Empty arrays should remain empty")
	}
	if len(cfg.Output.ImageDomains) != 0 {
		t.Error("Empty arrays should remain empty")
	}
}
