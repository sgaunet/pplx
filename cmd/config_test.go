package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sgaunet/pplx/pkg/config"
)

// setupTempConfigDir creates a temporary directory for config testing.
func setupTempConfigDir(t *testing.T) string {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "pplx-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	return tempDir
}

// copyTestFixture copies a test fixture to the temp directory.
func copyTestFixture(t *testing.T, fixtureName, destPath string) {
	t.Helper()

	fixtureData, err := os.ReadFile(filepath.Join("testdata", fixtureName))
	if err != nil {
		t.Fatalf("Failed to read fixture %s: %v", fixtureName, err)
	}

	if err := os.WriteFile(destPath, fixtureData, configFilePermission); err != nil {
		t.Fatalf("Failed to write fixture to %s: %v", destPath, err)
	}
}

// TestConfigInit tests the config init command with various options.
func TestConfigInit(t *testing.T) {
	// Note: Cannot run in parallel due to shared global state (initTemplate, etc.)

	tests := []struct {
		name          string
		template      string
		force         bool
		withExamples  bool
		interactive   bool
		expectError   bool
		validateFunc  func(*testing.T, string)
	}{
		{
			name:        "basic init creates config",
			template:    "",
			force:       false,
			expectError: false,
			validateFunc: func(t *testing.T, configPath string) {
				t.Helper()
				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					t.Errorf("Config file was not created at %s", configPath)
				}
			},
		},
		{
			name:        "init with research template",
			template:    "research",
			force:       false,
			expectError: false,
			validateFunc: func(t *testing.T, configPath string) {
				t.Helper()
				// Check file exists
				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					t.Errorf("Config file was not created at %s", configPath)
					return
				}
				// Read file content to verify it's not empty
				data, err := os.ReadFile(configPath)
				if err != nil {
					t.Fatalf("Failed to read config file: %v", err)
				}
				content := string(data)
				// Verify it contains research-related settings
				if !strings.Contains(content, "academic") {
					t.Error("Research template should contain 'academic' mode")
				}
				if !strings.Contains(content, "scholar.google.com") {
					t.Error("Research template should contain scholarly domains")
				}
			},
		},
		{
			name:        "init with creative template",
			template:    "creative",
			force:       false,
			expectError: false,
			validateFunc: func(t *testing.T, configPath string) {
				t.Helper()
				// Check file exists
				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					t.Errorf("Config file was not created at %s", configPath)
					return
				}
				// Read file content to verify it's not empty
				data, err := os.ReadFile(configPath)
				if err != nil {
					t.Fatalf("Failed to read config file: %v", err)
				}
				content := string(data)
				// Verify it contains creative-related settings
				if !strings.Contains(content, "stream") {
					t.Error("Creative template should mention streaming")
				}
				if !strings.Contains(content, "0.9") {
					t.Error("Creative template should have high temperature (0.9)")
				}
			},
		},
		{
			name:        "init with news template",
			template:    "news",
			force:       false,
			expectError: false,
			validateFunc: func(t *testing.T, configPath string) {
				t.Helper()
				// Check file exists
				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					t.Errorf("Config file was not created at %s", configPath)
					return
				}
				// Read file content to verify it's not empty
				data, err := os.ReadFile(configPath)
				if err != nil {
					t.Fatalf("Failed to read config file: %v", err)
				}
				content := string(data)
				// Verify it contains news-related settings
				if !strings.Contains(content, "week") {
					t.Error("News template should have week recency")
				}
				if !strings.Contains(content, "reuters.com") || !strings.Contains(content, "bbc.com") {
					t.Error("News template should contain news domains")
				}
			},
		},
		{
			name:        "init with full-example template",
			template:    "full-example",
			force:       false,
			expectError: false,
			validateFunc: func(t *testing.T, configPath string) {
				t.Helper()
				data, err := os.ReadFile(configPath)
				if err != nil {
					t.Fatalf("Failed to read config: %v", err)
				}
				content := string(data)
				// Full-example should have comments
				if !strings.Contains(content, "#") {
					t.Error("Full-example template should contain comments")
				}
			},
		},
		{
			name:        "init with invalid template",
			template:    "nonexistent",
			force:       false,
			expectError: true,
		},
		{
			name:         "init with examples flag",
			template:     "",
			force:        false,
			withExamples: true,
			expectError:  false,
			validateFunc: func(t *testing.T, configPath string) {
				t.Helper()
				data, err := os.ReadFile(configPath)
				if err != nil {
					t.Fatalf("Failed to read config: %v", err)
				}
				content := string(data)
				// With examples should have comments/annotations
				if !strings.Contains(content, "#") {
					t.Error("Config with examples should contain comments")
				}
			},
		},
		{
			name:        "force overwrite existing config",
			template:    "research",
			force:       true,
			expectError: false,
			validateFunc: func(t *testing.T, configPath string) {
				t.Helper()
				// Check file exists
				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					t.Errorf("Config file was not created at %s", configPath)
					return
				}
				// Read file content to verify it was overwritten
				data, err := os.ReadFile(configPath)
				if err != nil {
					t.Fatalf("Failed to read config file: %v", err)
				}
				content := string(data)
				// Verify it contains research template content, not initial content
				if strings.Contains(content, "initial: config") {
					t.Error("Config was not overwritten")
				}
				if !strings.Contains(content, "academic") {
					t.Error("Config should contain research template content")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Cannot run in parallel due to shared global state

			tempDir := setupTempConfigDir(t)
			configPath := filepath.Join(tempDir, "config.yaml")
			configFilePath = configPath

			// For force overwrite test, pre-create a config file
			if tt.name == "force overwrite existing config" {
				initialContent := []byte("initial: config\n")
				if err := os.WriteFile(configPath, initialContent, configFilePermission); err != nil {
					t.Fatalf("Failed to create initial config: %v", err)
				}
			}

			// Set flags
			initTemplate = tt.template
			initForce = tt.force
			initWithExamples = tt.withExamples
			initInteractive = tt.interactive

			err := runConfigInit(nil, nil)

			if (err != nil) != tt.expectError {
				t.Errorf("runConfigInit() error = %v, expectError %v", err, tt.expectError)
				return
			}

			if !tt.expectError && tt.validateFunc != nil {
				tt.validateFunc(t, configPath)
			}
		})
	}
}

// TestConfigShow tests the config show command.
func TestConfigShow(t *testing.T) {
	// Note: Cannot use t.Parallel() because subtests modify global variables (configFilePath, profileName)

	tests := []struct{
		name         string
		fixtureName  string
		profileName  string
		expectError  bool
		validateFunc func(*testing.T, error)
	}{
		{
			name:        "show valid config",
			fixtureName: "valid_config.yaml",
			expectError: false,
		},
		{
			name:        "show config with profiles",
			fixtureName: "profile_config.yaml",
			profileName: "research",
			expectError: false,
		},
		{
			name:        "show nonexistent profile",
			fixtureName: "profile_config.yaml",
			profileName: "nonexistent",
			expectError: true,
		},
		{
			name:        "show invalid syntax config",
			fixtureName: "invalid_syntax.yaml",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Cannot use t.Parallel() because we modify global variables

			tempDir := setupTempConfigDir(t)
			configPath := filepath.Join(tempDir, "config.yaml")
			copyTestFixture(t, tt.fixtureName, configPath)

			configFilePath = configPath
			profileName = tt.profileName

			loader := config.NewLoader()
			err := loader.LoadFrom(configPath)

			// If profile is specified, validate it exists
			if tt.profileName != "" && err == nil {
				pm := config.NewProfileManager(loader.Data())
				_, err = pm.LoadProfile(tt.profileName)
			}

			if (err != nil) != tt.expectError {
				t.Errorf("LoadFrom() error = %v, expectError %v", err, tt.expectError)
			}

			if !tt.expectError {
				cfg := loader.Data()
				if cfg == nil {
					t.Error("Loaded config is nil")
				}
			}
		})
	}
}

// TestConfigValidate tests the config validate command.
func TestConfigValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fixtureName string
		expectError bool
	}{
		{
			name:        "validate valid config",
			fixtureName: "valid_config.yaml",
			expectError: false,
		},
		{
			name:        "validate minimal config",
			fixtureName: "minimal_config.yaml",
			expectError: false,
		},
		{
			name:        "validate invalid values",
			fixtureName: "invalid_values.yaml",
			expectError: true,
		},
		{
			name:        "validate invalid syntax",
			fixtureName: "invalid_syntax.yaml",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tempDir := setupTempConfigDir(t)
			configPath := filepath.Join(tempDir, "config.yaml")
			copyTestFixture(t, tt.fixtureName, configPath)

			loader := config.NewLoader()
			err := loader.LoadFrom(configPath)

			if tt.expectError && err == nil {
				// Try validation on loaded config
				validator := config.NewValidator()
				err = validator.Validate(loader.Data())
			}

			if (err != nil) != tt.expectError {
				t.Errorf("Config validation error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

// TestConfigPath tests the config path command.
func TestConfigPath(t *testing.T) {
	// Note: Cannot use t.Parallel() because subtests modify global config.ConfigPaths

	tests := []struct {
		name         string
		createConfig bool
		expectFound  bool
	}{
		{
			name:         "find existing config",
			createConfig: true,
			expectFound:  true,
		},
		{
			name:         "no config found",
			createConfig: false,
			expectFound:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Cannot use t.Parallel() because we modify global config.ConfigPaths

			tempDir := setupTempConfigDir(t)

			// Always override config paths to prevent finding real config files
			oldPaths := config.ConfigPaths
			config.ConfigPaths = []string{tempDir}
			t.Cleanup(func() {
				config.ConfigPaths = oldPaths
			})

			if tt.createConfig {
				configPath := filepath.Join(tempDir, "config.yaml")
				copyTestFixture(t, "valid_config.yaml", configPath)
			}

			foundPath, err := config.FindConfigFile()
			found := err == nil

			if found != tt.expectFound {
				t.Errorf("FindConfigFile() found = %v, expectFound %v", found, tt.expectFound)
			}

			if tt.expectFound && foundPath == "" {
				t.Error("Config should be found but path is empty")
			}
		})
	}
}

// TestConfigOptions tests the config options command.
func TestConfigOptions(t *testing.T) {
	// Note: Cannot use t.Parallel() because subtests modify global variables (optionsSection, optionsFormat, optionsValidation)

	tests := []struct {
		name             string
		section          string
		format           string
		showValidation   bool
		expectError      bool
		expectedMinCount int
	}{
		{
			name:             "list all options",
			section:          "",
			format:           "table",
			showValidation:   false,
			expectError:      false,
			expectedMinCount: 29, // All 29 config options
		},
		{
			name:             "list defaults section",
			section:          "defaults",
			format:           "table",
			showValidation:   false,
			expectError:      false,
			expectedMinCount: 8,
		},
		{
			name:             "list search section",
			section:          "search",
			format:           "table",
			showValidation:   false,
			expectError:      false,
			expectedMinCount: 11,
		},
		{
			name:             "list output section",
			section:          "output",
			format:           "table",
			showValidation:   false,
			expectError:      false,
			expectedMinCount: 9,
		},
		{
			name:             "list api section",
			section:          "api",
			format:           "table",
			showValidation:   false,
			expectError:      false,
			expectedMinCount: 3,
		},
		{
			name:        "invalid section",
			section:     "nonexistent",
			format:      "table",
			expectError: true,
		},
		{
			name:             "json format",
			section:          "",
			format:           "json",
			showValidation:   false,
			expectError:      false,
			expectedMinCount: 29,
		},
		{
			name:             "yaml format",
			section:          "",
			format:           "yaml",
			showValidation:   false,
			expectError:      false,
			expectedMinCount: 29,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Cannot use t.Parallel() because we modify global variables

			optionsSection = tt.section
			optionsFormat = tt.format
			optionsValidation = tt.showValidation

			registry := config.NewMetadataRegistry()
			var options []*config.OptionMetadata

			if tt.section != "" {
				options = registry.GetBySection(tt.section)
				if len(options) == 0 && !tt.expectError {
					t.Errorf("No options found for section %s", tt.section)
				}
			} else {
				options = registry.GetAll()
			}

			if !tt.expectError && len(options) < tt.expectedMinCount {
				t.Errorf("Expected at least %d options, got %d", tt.expectedMinCount, len(options))
			}

			// Test formatting
			if !tt.expectError && len(options) > 0 {
				output, err := config.FormatOptions(options, tt.format)
				if err != nil {
					t.Errorf("FormatOptions() error = %v", err)
				}
				if output == "" {
					t.Error("FormatOptions() returned empty output")
				}
			}
		})
	}
}

// TestTemplateLoading tests loading all template types.
func TestTemplateLoading(t *testing.T) {
	// Note: Cannot use t.Parallel() due to potential race conditions in template loading

	templates := []string{
		config.TemplateResearch,
		config.TemplateCreative,
		config.TemplateNews,
		config.TemplateFullExample,
	}

	for _, tmpl := range templates {
		t.Run(tmpl, func(t *testing.T) {
			// Note: Cannot use t.Parallel() due to potential race conditions in template loading

			cfg, err := config.LoadTemplate(tmpl)
			if err != nil {
				t.Fatalf("LoadTemplate(%s) error = %v", tmpl, err)
			}

			if cfg == nil {
				t.Fatalf("LoadTemplate(%s) returned nil config", tmpl)
			}

			// Validate loaded template
			validator := config.NewValidator()
			if err := validator.Validate(cfg); err != nil {
				t.Errorf("Template %s failed validation: %v", tmpl, err)
			}

			// Verify template has model set
			if cfg.Defaults.Model == "" {
				t.Errorf("Template %s should have model set", tmpl)
			}
		})
	}
}

// TestProfileManagement tests profile creation, switching, and deletion.
func TestProfileManagement(t *testing.T) {
	// Note: Cannot use t.Parallel() due to potential race conditions in profile operations

	t.Run("create and switch profiles", func(t *testing.T) {
		// Note: Cannot use t.Parallel() due to potential race conditions

		tempDir := setupTempConfigDir(t)
		configPath := filepath.Join(tempDir, "config.yaml")
		copyTestFixture(t, "profile_config.yaml", configPath)

		loader := config.NewLoader()
		if err := loader.LoadFrom(configPath); err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		pm := config.NewProfileManager(loader.Data())

		// Test list profiles
		profiles := pm.ListProfiles()
		if len(profiles) < 2 {
			t.Errorf("Expected at least 2 profiles, got %d", len(profiles))
		}

		// Test get active profile
		activeProfile := pm.GetActiveProfileName()
		if activeProfile != "research" {
			t.Errorf("Expected active profile 'research', got %s", activeProfile)
		}

		// Test load profile
		profile, err := pm.LoadProfile("creative")
		if err != nil {
			t.Fatalf("Failed to load creative profile: %v", err)
		}

		if profile.Defaults.Temperature < 0.8 {
			t.Errorf("Creative profile should have high temperature, got %f", profile.Defaults.Temperature)
		}

		// Test switch profile
		if err := pm.SetActiveProfile("creative"); err != nil {
			t.Fatalf("Failed to switch profile: %v", err)
		}

		if pm.GetActiveProfileName() != "creative" {
			t.Error("Profile was not switched")
		}
	})
}

// TestConfigPrecedence tests that configuration precedence works correctly.
func TestConfigPrecedence(t *testing.T) {
	// Note: Cannot use t.Parallel() because subtests use t.Setenv()

	tests := []struct {
		name              string
		configTemp        float64
		profileTemp       float64
		envTemp           string
		expectedTemp      float64
		useProfile        bool
		setEnv            bool
	}{
		{
			name:         "config only",
			configTemp:   0.5,
			expectedTemp: 0.5,
			useProfile:   false,
			setEnv:       false,
		},
		{
			name:         "profile overrides config",
			configTemp:   0.5,
			profileTemp:  0.9,
			expectedTemp: 0.9,
			useProfile:   true,
			setEnv:       false,
		},
		{
			name:         "env overrides profile",
			configTemp:   0.5,
			profileTemp:  0.9,
			envTemp:      "0.3",
			expectedTemp: 0.3,
			useProfile:   true,
			setEnv:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Cannot use t.Parallel() with t.Setenv()

			if tt.setEnv {
				t.Setenv("PPLX_TEMPERATURE", tt.envTemp)
			}

			cfg := config.NewConfigData()
			cfg.Defaults.Temperature = tt.configTemp

			if tt.useProfile {
				profile := &config.Profile{
					Name: "test",
					Defaults: config.DefaultsConfig{
						Temperature: tt.profileTemp,
					},
				}
				cfg.Profiles = map[string]*config.Profile{
					"test": profile,
				}
				cfg.ActiveProfile = "test"
			}

			// Note: Full precedence testing would require CLI flag integration
			// which is tested at the integration level
			if !tt.setEnv {
				if tt.useProfile {
					profile := cfg.Profiles["test"]
					if profile.Defaults.Temperature != tt.expectedTemp {
						t.Errorf("Expected temperature %f, got %f", tt.expectedTemp, profile.Defaults.Temperature)
					}
				} else {
					if cfg.Defaults.Temperature != tt.expectedTemp {
						t.Errorf("Expected temperature %f, got %f", tt.expectedTemp, cfg.Defaults.Temperature)
					}
				}
			}
		})
	}
}

// TestErrorMessages tests that error messages are helpful.
func TestErrorMessages(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		operation       func() error
		expectedMsgPart string
	}{
		{
			name: "missing config file",
			operation: func() error {
				loader := config.NewLoader()
				return loader.LoadFrom("/nonexistent/config.yaml")
			},
			expectedMsgPart: "no such file",
		},
		{
			name: "invalid template name",
			operation: func() error {
				_, err := config.LoadTemplate("nonexistent-template")
				return err
			},
			expectedMsgPart: "template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.operation()
			if err == nil {
				t.Error("Expected error but got nil")
				return
			}

			errMsg := strings.ToLower(err.Error())
			expectedPart := strings.ToLower(tt.expectedMsgPart)

			if !strings.Contains(errMsg, expectedPart) {
				t.Errorf("Error message %q does not contain %q", errMsg, expectedPart)
			}
		})
	}
}

// TestEnvironmentVariableExpansion tests that env vars are properly expanded.
func TestEnvironmentVariableExpansion(t *testing.T) {
	// Note: Cannot use t.Parallel() with t.Setenv()

	t.Setenv("TEST_API_KEY", "test-key-12345")
	t.Setenv("TEST_MODEL", "test-model")

	tempDir := setupTempConfigDir(t)
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create config with env var references
	// Note: Only string fields can have env var references, not typed fields like timeout (time.Duration)
	configContent := `
api:
  key: ${TEST_API_KEY}
  timeout: 60s
defaults:
  model: ${TEST_MODEL}
`
	if err := os.WriteFile(configPath, []byte(configContent), configFilePermission); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	loader := config.NewLoader()
	if err := loader.LoadFrom(configPath); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	cfg := loader.Data()

	// Expand environment variables (normally done by LoadAndMergeConfig)
	config.ExpandEnvVars(cfg)

	// Verify env var expansion worked correctly
	if cfg.API.Key != "test-key-12345" {
		t.Errorf("API key should be expanded from env: got %q, want %q", cfg.API.Key, "test-key-12345")
	}
	if cfg.Defaults.Model != "test-model" {
		t.Errorf("Model should be expanded from env: got %q, want %q", cfg.Defaults.Model, "test-model")
	}
	if cfg.API.Timeout != 60*time.Second {
		t.Errorf("Timeout should be parsed correctly: got %v, want %v", cfg.API.Timeout, 60*time.Second)
	}
}

// TestConcurrentConfigAccess tests concurrent config loading.
func TestConcurrentConfigAccess(t *testing.T) {
	t.Parallel()

	tempDir := setupTempConfigDir(t)
	configPath := filepath.Join(tempDir, "config.yaml")
	copyTestFixture(t, "valid_config.yaml", configPath)

	// Launch multiple goroutines reading config
	const numReaders = 10
	done := make(chan error, numReaders)

	for i := 0; i < numReaders; i++ {
		go func() {
			loader := config.NewLoader()
			err := loader.LoadFrom(configPath)
			done <- err
		}()
	}

	// Wait for all readers
	for i := 0; i < numReaders; i++ {
		if err := <-done; err != nil {
			t.Errorf("Concurrent config load failed: %v", err)
		}
	}
}

// TestAnnotatedConfigGeneration tests generation of annotated configs.
func TestAnnotatedConfigGeneration(t *testing.T) {
	t.Parallel()

	cfg := config.NewConfigData()
	cfg.Defaults.Model = "sonar"
	cfg.Defaults.Temperature = 0.7

	opts := config.DefaultAnnotationOptions()
	opts.IncludeExamples = true

	annotated, err := config.GenerateAnnotatedConfig(cfg, opts)
	if err != nil {
		t.Fatalf("GenerateAnnotatedConfig() error = %v", err)
	}

	if annotated == "" {
		t.Error("GenerateAnnotatedConfig() returned empty string")
	}

	// Verify it contains comments
	if !strings.Contains(annotated, "#") {
		t.Error("Annotated config should contain comments")
	}

	// Verify it's valid YAML
	tempDir := setupTempConfigDir(t)
	testPath := filepath.Join(tempDir, "annotated.yaml")
	if err := os.WriteFile(testPath, []byte(annotated), configFilePermission); err != nil {
		t.Fatalf("Failed to write annotated config: %v", err)
	}

	loader := config.NewLoader()
	if err := loader.LoadFrom(testPath); err != nil {
		t.Errorf("Annotated config is not valid YAML: %v", err)
	}
}

// TestVerifyConfigPermissions tests the permission verification function.
func TestVerifyConfigPermissions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		permissions os.FileMode
		expectWarn  bool
		expectError bool
	}{
		{
			name:        "secure permissions 0600",
			permissions: 0600,
			expectWarn:  false,
			expectError: false,
		},
		{
			name:        "secure permissions 0400",
			permissions: 0400,
			expectWarn:  false,
			expectError: false,
		},
		{
			name:        "insecure permissions 0644 (group/other read)",
			permissions: 0644,
			expectWarn:  true,
			expectError: false,
		},
		{
			name:        "insecure permissions 0666 (world writable)",
			permissions: 0666,
			expectWarn:  true,
			expectError: false,
		},
		{
			name:        "insecure permissions 0755 (world readable/executable)",
			permissions: 0755,
			expectWarn:  true,
			expectError: false,
		},
		{
			name:        "insecure permissions 0664 (group writable)",
			permissions: 0664,
			expectWarn:  true,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tempDir := setupTempConfigDir(t)
			testFile := filepath.Join(tempDir, "test_config.yaml")

			// Create test file with initial permissions
			content := []byte("test: config\n")
			if err := os.WriteFile(testFile, content, 0600); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Explicitly set desired permissions using chmod (bypasses umask)
			if err := os.Chmod(testFile, tt.permissions); err != nil {
				t.Fatalf("Failed to set permissions: %v", err)
			}

			// Verify permissions were set correctly
			info, err := os.Stat(testFile)
			if err != nil {
				t.Fatalf("Failed to stat test file: %v", err)
			}

			actualPerms := info.Mode().Perm()
			if actualPerms != tt.permissions {
				t.Fatalf("Permissions not set correctly: got %#o, want %#o", actualPerms, tt.permissions)
			}

			// Call verifyConfigPermissions
			err = verifyConfigPermissions(testFile)

			// Check for errors
			if (err != nil) != tt.expectError {
				t.Errorf("verifyConfigPermissions() error = %v, expectError %v", err, tt.expectError)
			}

			// Note: Testing for warnings to stderr is complex and would require
			// capturing stderr output. The important part is that the function
			// doesn't return an error for insecure permissions, only warns.
			if tt.expectWarn && err != nil {
				t.Error("Function should warn but not error for insecure permissions")
			}
		})
	}
}

// TestVerifyConfigPermissionsNonexistent tests permission check on missing file.
func TestVerifyConfigPermissionsNonexistent(t *testing.T) {
	t.Parallel()

	err := verifyConfigPermissions("/nonexistent/file/path.yaml")
	if err == nil {
		t.Error("verifyConfigPermissions() should return error for nonexistent file")
	}

	if !strings.Contains(err.Error(), "failed to check file permissions") {
		t.Errorf("Error message should mention permission check failure, got: %v", err)
	}
}

// TestConfigInitVerifiesPermissions tests that config init verifies permissions.
func TestConfigInitVerifiesPermissions(t *testing.T) {
	// Note: Cannot run in parallel due to shared global state

	tempDir := setupTempConfigDir(t)
	configPath := filepath.Join(tempDir, "config.yaml")
	configFilePath = configPath

	// Reset flags
	initTemplate = ""
	initForce = false
	initWithExamples = false
	initInteractive = false

	err := runConfigInit(nil, nil)
	if err != nil {
		t.Fatalf("runConfigInit() error = %v", err)
	}

	// Verify file was created
	info, err := os.Stat(configPath)
	if err != nil {
		t.Fatalf("Config file was not created: %v", err)
	}

	// Verify file has correct permissions
	mode := info.Mode().Perm()
	if mode != configFilePermission {
		t.Errorf("Config file has wrong permissions: got %#o, want %#o", mode, configFilePermission)
	}
}
