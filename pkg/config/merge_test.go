package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// createTestCommand creates a cobra.Command with common flags for testing.
func createTestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use: "test",
		Run: func(cmd *cobra.Command, args []string) {},
	}

	// Add flags that match those in the actual application
	cmd.Flags().String("model", "", "Model to use")
	cmd.Flags().Float64("temperature", 0, "Temperature")
	cmd.Flags().Int("max-tokens", 0, "Max tokens")
	cmd.Flags().Int("top-k", 0, "Top K")
	cmd.Flags().Float64("top-p", 0, "Top P")
	cmd.Flags().Float64("frequency-penalty", 0, "Frequency penalty")
	cmd.Flags().Float64("presence-penalty", 0, "Presence penalty")
	cmd.Flags().Duration("timeout", 0, "Timeout")

	// Search flags
	cmd.Flags().StringSlice("search-domains", nil, "Search domains")
	cmd.Flags().String("search-recency", "", "Search recency")
	cmd.Flags().String("search-mode", "", "Search mode")
	cmd.Flags().String("search-context-size", "", "Search context size")
	cmd.Flags().Float64("location-lat", 0, "Location latitude")
	cmd.Flags().Float64("location-lon", 0, "Location longitude")
	cmd.Flags().String("location-country", "", "Location country")
	cmd.Flags().String("search-after-date", "", "After date")
	cmd.Flags().String("search-before-date", "", "Before date")
	cmd.Flags().String("last-updated-after", "", "Last updated after")
	cmd.Flags().String("last-updated-before", "", "Last updated before")

	// Output flags
	cmd.Flags().Bool("stream", false, "Stream output")
	cmd.Flags().Bool("return-images", false, "Return images")
	cmd.Flags().Bool("return-related", false, "Return related")
	cmd.Flags().Bool("json", false, "JSON output")
	cmd.Flags().StringSlice("image-domains", nil, "Image domains")
	cmd.Flags().StringSlice("image-formats", nil, "Image formats")
	cmd.Flags().String("response-format-json-schema", "", "JSON schema")
	cmd.Flags().String("response-format-regex", "", "Regex format")
	cmd.Flags().String("reasoning-effort", "", "Reasoning effort")

	return cmd
}

// =============================================================================
// Category 1: MergeWithFlags Tests (8 tests)
// =============================================================================

func TestMergeWithFlags_NoFlagsChanged(t *testing.T) {
	// Config file values
	cfg := &ConfigData{
		Defaults: DefaultsConfig{
			Model:       "config-model",
			Temperature: 0.7,
			MaxTokens:   1000,
		},
		Search: SearchConfig{
			Recency: "week",
			Mode:    "web",
		},
		Output: OutputConfig{
			Stream: true,
		},
	}

	cmd := createTestCommand()
	merger := NewMerger(cfg)
	if err := merger.BindFlags(cmd); err != nil {
		t.Fatalf("Failed to bind flags: %v", err)
	}

	merged := merger.MergeWithFlags(cmd)

	// Since no flags were changed, config values should be preserved
	if merged.Defaults.Model != "config-model" {
		t.Errorf("Model should be preserved from config, got '%s'", merged.Defaults.Model)
	}
	if merged.Defaults.Temperature != 0.7 {
		t.Errorf("Temperature should be preserved from config, got %f", merged.Defaults.Temperature)
	}
	if merged.Defaults.MaxTokens != 1000 {
		t.Errorf("MaxTokens should be preserved from config, got %d", merged.Defaults.MaxTokens)
	}
	if merged.Search.Recency != "week" {
		t.Errorf("Recency should be preserved from config, got '%s'", merged.Search.Recency)
	}
	if !merged.Output.Stream {
		t.Error("Stream should be preserved from config")
	}
}

func TestMergeWithFlags_AllFlagsChanged(t *testing.T) {
	// Config file values
	cfg := &ConfigData{
		Defaults: DefaultsConfig{
			Model:       "config-model",
			Temperature: 0.7,
		},
	}

	cmd := createTestCommand()

	// Set flags
	_ = cmd.Flags().Set("model", "flag-model")
	_ = cmd.Flags().Set("temperature", "0.9")

	merger := NewMerger(cfg)
	if err := merger.BindFlags(cmd); err != nil {
		t.Fatalf("Failed to bind flags: %v", err)
	}

	merged := merger.MergeWithFlags(cmd)

	// Flag values should override config
	if merged.Defaults.Model != "flag-model" {
		t.Errorf("Expected flag model 'flag-model', got '%s'", merged.Defaults.Model)
	}
	if merged.Defaults.Temperature != 0.9 {
		t.Errorf("Expected flag temperature 0.9, got %f", merged.Defaults.Temperature)
	}
}

func TestMergeWithFlags_PartialFlagsChanged(t *testing.T) {
	cfg := &ConfigData{
		Defaults: DefaultsConfig{
			Model:       "config-model",
			Temperature: 0.7,
			MaxTokens:   1000,
			TopK:        50,
		},
	}

	cmd := createTestCommand()

	// Only set some flags
	_ = cmd.Flags().Set("model", "flag-model")
	_ = cmd.Flags().Set("max-tokens", "2000")

	merger := NewMerger(cfg)
	if err := merger.BindFlags(cmd); err != nil {
		t.Fatalf("Failed to bind flags: %v", err)
	}

	merged := merger.MergeWithFlags(cmd)

	// Changed flags should override
	if merged.Defaults.Model != "flag-model" {
		t.Errorf("Model should be from flag, got '%s'", merged.Defaults.Model)
	}
	if merged.Defaults.MaxTokens != 2000 {
		t.Errorf("MaxTokens should be from flag, got %d", merged.Defaults.MaxTokens)
	}

	// Unchanged flags should preserve config
	if merged.Defaults.Temperature != 0.7 {
		t.Errorf("Temperature should be from config, got %f", merged.Defaults.Temperature)
	}
	if merged.Defaults.TopK != 50 {
		t.Errorf("TopK should be from config, got %d", merged.Defaults.TopK)
	}
}

func TestMergeWithFlags_ExplicitZeroValues(t *testing.T) {
	// This tests the critical Changed() pattern
	cfg := &ConfigData{
		Defaults: DefaultsConfig{
			Temperature: 0.7,
			TopK:        50,
		},
	}

	cmd := createTestCommand()

	// Explicitly set temperature to 0 (valid value)
	_ = cmd.Flags().Set("temperature", "0")

	merger := NewMerger(cfg)
	if err := merger.BindFlags(cmd); err != nil {
		t.Fatalf("Failed to bind flags: %v", err)
	}

	merged := merger.MergeWithFlags(cmd)

	// Explicit zero should override config
	if merged.Defaults.Temperature != 0 {
		t.Errorf("Explicit zero temperature should override config, got %f", merged.Defaults.Temperature)
	}

	// Non-changed flag should preserve config
	if merged.Defaults.TopK != 50 {
		t.Errorf("TopK should be preserved from config, got %d", merged.Defaults.TopK)
	}
}

func TestMergeWithFlags_BooleanHandling(t *testing.T) {
	cfg := &ConfigData{
		Output: OutputConfig{
			Stream:        true,
			ReturnImages:  false,
			ReturnRelated: true,
		},
	}

	cmd := createTestCommand()

	// Explicitly set stream to false
	_ = cmd.Flags().Set("stream", "false")
	// Explicitly set return-images to true
	_ = cmd.Flags().Set("return-images", "true")
	// Don't touch return-related

	merger := NewMerger(cfg)
	if err := merger.BindFlags(cmd); err != nil {
		t.Fatalf("Failed to bind flags: %v", err)
	}

	merged := merger.MergeWithFlags(cmd)

	// Changed flags should override
	if merged.Output.Stream {
		t.Error("Stream should be false from flag")
	}
	if !merged.Output.ReturnImages {
		t.Error("ReturnImages should be true from flag")
	}

	// Unchanged should preserve config
	if !merged.Output.ReturnRelated {
		t.Error("ReturnRelated should be preserved from config")
	}
}

func TestMergeWithFlags_StringArrays(t *testing.T) {
	cfg := &ConfigData{
		Search: SearchConfig{
			Domains: []string{"config1.com", "config2.com"},
		},
		Output: OutputConfig{
			ImageFormats: []string{"png", "jpg"},
		},
	}

	cmd := createTestCommand()

	// Set search-domains flag
	_ = cmd.Flags().Set("search-domains", "flag1.com,flag2.com,flag3.com")

	merger := NewMerger(cfg)
	if err := merger.BindFlags(cmd); err != nil {
		t.Fatalf("Failed to bind flags: %v", err)
	}

	merged := merger.MergeWithFlags(cmd)

	// Changed array should override
	expectedDomains := []string{"flag1.com", "flag2.com", "flag3.com"}
	if len(merged.Search.Domains) != len(expectedDomains) {
		t.Fatalf("Expected %d domains, got %d", len(expectedDomains), len(merged.Search.Domains))
	}
	for i, exp := range expectedDomains {
		if merged.Search.Domains[i] != exp {
			t.Errorf("Domain[%d]: expected '%s', got '%s'", i, exp, merged.Search.Domains[i])
		}
	}

	// Unchanged array should preserve config
	if len(merged.Output.ImageFormats) != 2 {
		t.Errorf("ImageFormats should be preserved from config")
	}
}

func TestMergeWithFlags_DurationParsing(t *testing.T) {
	cfg := &ConfigData{
		Defaults: DefaultsConfig{
			Timeout: "30s",
		},
	}

	cmd := createTestCommand()

	// Set timeout flag
	_ = cmd.Flags().Set("timeout", "1m30s")

	merger := NewMerger(cfg)
	if err := merger.BindFlags(cmd); err != nil {
		t.Fatalf("Failed to bind flags: %v", err)
	}

	merged := merger.MergeWithFlags(cmd)

	// Timeout should be overridden and formatted correctly
	expected := (1*time.Minute + 30*time.Second).String()
	if merged.Defaults.Timeout != expected {
		t.Errorf("Expected timeout '%s', got '%s'", expected, merged.Defaults.Timeout)
	}
}

func TestMergeWithFlags_FloatBoundaries(t *testing.T) {
	cfg := &ConfigData{
		Defaults: DefaultsConfig{
			Temperature: 0.5,
		},
	}

	testCases := []struct {
		name     string
		value    string
		expected float64
	}{
		{"zero", "0", 0.0},
		{"exact max", "2.0", 2.0},
		{"small value", "0.1", 0.1},
		{"mid value", "1.5", 1.5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := createTestCommand()
			_ = cmd.Flags().Set("temperature", tc.value)

			merger := NewMerger(cfg)
			if err := merger.BindFlags(cmd); err != nil {
				t.Fatalf("Failed to bind flags: %v", err)
			}

			merged := merger.MergeWithFlags(cmd)

			if merged.Defaults.Temperature != tc.expected {
				t.Errorf("Expected temperature %f, got %f", tc.expected, merged.Defaults.Temperature)
			}
		})
	}
}

// =============================================================================
// Category 2: Precedence Chain Tests (5 tests)
// =============================================================================

func TestPrecedence_CLIOverridesConfig(t *testing.T) {
	cfg := &ConfigData{
		Defaults: DefaultsConfig{
			Model: "config-model",
		},
	}

	cmd := createTestCommand()
	_ = cmd.Flags().Set("model", "cli-model")

	merger := NewMerger(cfg)
	if err := merger.BindFlags(cmd); err != nil {
		t.Fatalf("Failed to bind flags: %v", err)
	}

	merged := merger.MergeWithFlags(cmd)

	if merged.Defaults.Model != "cli-model" {
		t.Errorf("CLI flag should override config, got '%s'", merged.Defaults.Model)
	}
}

func TestPrecedence_ConfigOverridesDefaults(t *testing.T) {
	// Config with values set
	cfg := &ConfigData{
		Defaults: DefaultsConfig{
			Model:       "config-model",
			Temperature: 0.8,
			MaxTokens:   2000,
		},
	}

	cmd := createTestCommand()
	// No flags set

	merger := NewMerger(cfg)
	if err := merger.BindFlags(cmd); err != nil {
		t.Fatalf("Failed to bind flags: %v", err)
	}

	merged := merger.MergeWithFlags(cmd)

	// Config values should be used (they override zero defaults)
	if merged.Defaults.Model != "config-model" {
		t.Errorf("Config should override defaults, got '%s'", merged.Defaults.Model)
	}
	if merged.Defaults.Temperature != 0.8 {
		t.Errorf("Config should override defaults, got %f", merged.Defaults.Temperature)
	}
	if merged.Defaults.MaxTokens != 2000 {
		t.Errorf("Config should override defaults, got %d", merged.Defaults.MaxTokens)
	}
}

func TestPrecedence_CLIOverridesEnvVar(t *testing.T) {
	// Set up environment variable
	cleanup := setupEnvTest(t, map[string]string{
		"TEST_MODEL": "env-model",
	})
	defer cleanup()

	cfg := &ConfigData{
		Defaults: DefaultsConfig{
			Model: "${TEST_MODEL}",
		},
	}

	// Expand env vars first
	ExpandEnvVars(cfg)

	// Verify env var was expanded
	if cfg.Defaults.Model != "env-model" {
		t.Fatalf("Env var should be expanded, got '%s'", cfg.Defaults.Model)
	}

	// Now set CLI flag
	cmd := createTestCommand()
	_ = cmd.Flags().Set("model", "cli-model")

	merger := NewMerger(cfg)
	if err := merger.BindFlags(cmd); err != nil {
		t.Fatalf("Failed to bind flags: %v", err)
	}

	merged := merger.MergeWithFlags(cmd)

	// CLI should override env var (which was in config)
	if merged.Defaults.Model != "cli-model" {
		t.Errorf("CLI flag should override env var, got '%s'", merged.Defaults.Model)
	}
}

func TestPrecedence_EnvVarOverridesConfig(t *testing.T) {
	cleanup := setupEnvTest(t, map[string]string{
		"ENV_MODEL": "env-value",
	})
	defer cleanup()

	// Config with env var reference takes precedence over literal value
	cfg := &ConfigData{
		Defaults: DefaultsConfig{
			Model: "$ENV_MODEL",
		},
	}

	ExpandEnvVars(cfg)

	if cfg.Defaults.Model != "env-value" {
		t.Errorf("Env var should expand to 'env-value', got '%s'", cfg.Defaults.Model)
	}
}

func TestPrecedence_FullChainIntegration(t *testing.T) {
	// Set up all three sources
	cleanup := setupEnvTest(t, map[string]string{
		"ENV_TEMP": "0.6",
	})
	defer cleanup()

	// Config file with mix of literal and env var
	cfg := &ConfigData{
		Defaults: DefaultsConfig{
			Model:       "config-model",        // Literal in config
			Temperature: 0.5,                   // Will be overridden by CLI
			MaxTokens:   1000,                  // Will stay from config
			TopK:        50,                    // Will be overridden by CLI
		},
		Search: SearchConfig{
			Recency: "week", // Will stay from config
		},
	}

	// Expand env vars (though none in this test)
	ExpandEnvVars(cfg)

	// Set some CLI flags
	cmd := createTestCommand()
	_ = cmd.Flags().Set("temperature", "0.9")  // CLI override
	_ = cmd.Flags().Set("top-k", "100")        // CLI override

	merger := NewMerger(cfg)
	if err := merger.BindFlags(cmd); err != nil {
		t.Fatalf("Failed to bind flags: %v", err)
	}

	merged := merger.MergeWithFlags(cmd)

	// Verify precedence: CLI > env var > config > defaults
	if merged.Defaults.Temperature != 0.9 {
		t.Errorf("CLI should win for temperature, got %f", merged.Defaults.Temperature)
	}
	if merged.Defaults.TopK != 100 {
		t.Errorf("CLI should win for top-k, got %d", merged.Defaults.TopK)
	}
	if merged.Defaults.Model != "config-model" {
		t.Errorf("Config should be preserved for model, got '%s'", merged.Defaults.Model)
	}
	if merged.Defaults.MaxTokens != 1000 {
		t.Errorf("Config should be preserved for max-tokens, got %d", merged.Defaults.MaxTokens)
	}
	if merged.Search.Recency != "week" {
		t.Errorf("Config should be preserved for recency, got '%s'", merged.Search.Recency)
	}
}

// =============================================================================
// Category 3: LoadAndMergeConfig Integration (6 tests)
// =============================================================================

func TestLoadAndMergeConfig_CompleteWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
defaults:
  model: workflow-model
  temperature: 0.8
  max_tokens: 1500

search:
  recency: month
  mode: academic

output:
  stream: true
  return_images: false
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	cmd := createTestCommand()
	// Don't set any flags

	cfg, err := LoadAndMergeConfig(cmd, configPath)
	if err != nil {
		t.Fatalf("LoadAndMergeConfig failed: %v", err)
	}

	// Verify loaded values
	if cfg.Defaults.Model != "workflow-model" {
		t.Errorf("Expected model 'workflow-model', got '%s'", cfg.Defaults.Model)
	}
	if cfg.Defaults.Temperature != 0.8 {
		t.Errorf("Expected temperature 0.8, got %f", cfg.Defaults.Temperature)
	}
	if cfg.Search.Recency != "month" {
		t.Errorf("Expected recency 'month', got '%s'", cfg.Search.Recency)
	}
}

func TestLoadAndMergeConfig_WithEnvVars(t *testing.T) {
	cleanup := setupEnvTest(t, map[string]string{
		"TEST_API_KEY": "sk-env-key-123",
		"TEST_MODEL":   "env-model",
	})
	defer cleanup()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
api:
  key: $TEST_API_KEY

defaults:
  model: ${TEST_MODEL}
  temperature: 0.7
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	cmd := createTestCommand()

	cfg, err := LoadAndMergeConfig(cmd, configPath)
	if err != nil {
		t.Fatalf("LoadAndMergeConfig failed: %v", err)
	}

	// Verify env vars were expanded
	if cfg.API.Key != "sk-env-key-123" {
		t.Errorf("API key should be expanded from env, got '%s'", cfg.API.Key)
	}
	if cfg.Defaults.Model != "env-model" {
		t.Errorf("Model should be expanded from env, got '%s'", cfg.Defaults.Model)
	}
}

func TestLoadAndMergeConfig_NoConfigFile(t *testing.T) {
	cmd := createTestCommand()

	// Pass empty path
	cfg, err := LoadAndMergeConfig(cmd, "")
	if err != nil {
		t.Fatalf("LoadAndMergeConfig should not fail without config file: %v", err)
	}

	// Should return empty config
	if cfg == nil {
		t.Fatal("Config should not be nil")
	}
}

func TestLoadAndMergeConfig_CLIOverridesAll(t *testing.T) {
	cleanup := setupEnvTest(t, map[string]string{
		"ENV_MODEL": "env-model",
	})
	defer cleanup()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
defaults:
  model: $ENV_MODEL
  temperature: 0.7
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	cmd := createTestCommand()
	// Set CLI flags
	_ = cmd.Flags().Set("model", "cli-model")
	_ = cmd.Flags().Set("temperature", "0.9")

	cfg, err := LoadAndMergeConfig(cmd, configPath)
	if err != nil {
		t.Fatalf("LoadAndMergeConfig failed: %v", err)
	}

	// CLI should override everything
	if cfg.Defaults.Model != "cli-model" {
		t.Errorf("CLI should override config and env, got '%s'", cfg.Defaults.Model)
	}
	if cfg.Defaults.Temperature != 0.9 {
		t.Errorf("CLI should override config, got %f", cfg.Defaults.Temperature)
	}
}

func TestLoadAndMergeConfig_WithActiveProfile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
defaults:
  model: base-model
  temperature: 0.5

active_profile: production

profiles:
  production:
    name: production
    defaults:
      model: prod-model
      temperature: 0.8
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	cmd := createTestCommand()

	cfg, err := LoadAndMergeConfig(cmd, configPath)
	if err != nil {
		t.Fatalf("LoadAndMergeConfig failed: %v", err)
	}

	// Profile values should be merged
	if cfg.Defaults.Model != "prod-model" {
		t.Errorf("Profile should override base config, got '%s'", cfg.Defaults.Model)
	}
	if cfg.Defaults.Temperature != 0.8 {
		t.Errorf("Profile should override base config, got %f", cfg.Defaults.Temperature)
	}
}

func TestLoadAndMergeConfig_InvalidConfigPath(t *testing.T) {
	cmd := createTestCommand()

	// Try to load non-existent file
	_, err := LoadAndMergeConfig(cmd, "/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Expected error for invalid config path")
	}
}
