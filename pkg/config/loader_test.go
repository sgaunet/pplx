package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewLoader(t *testing.T) {
	loader := NewLoader()
	if loader == nil {
		t.Fatal("NewLoader returned nil")
	}

	if loader.viper == nil {
		t.Error("Loader viper instance is nil")
	}

	if loader.data == nil {
		t.Error("Loader data is nil")
	}
}

func TestLoadNonExistentFile(t *testing.T) {
	loader := NewLoader()

	// Loading non-existent config should not error (it's optional)
	err := loader.Load()
	if err != nil {
		t.Errorf("Load() should not error for non-existent config: %v", err)
	}
}

func TestLoadFrom(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test-config.yaml")

	configContent := `
defaults:
  model: test-model
  temperature: 0.5
  max_tokens: 1000

search:
  recency: week
  mode: web

output:
  stream: true
  return_images: false
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	loader := NewLoader()
	if err := loader.LoadFrom(configPath); err != nil {
		t.Fatalf("LoadFrom() failed: %v", err)
	}

	data := loader.Data()

	// Verify loaded values
	if data.Defaults.Model != "test-model" {
		t.Errorf("Expected model 'test-model', got '%s'", data.Defaults.Model)
	}

	if data.Defaults.Temperature != 0.5 {
		t.Errorf("Expected temperature 0.5, got %f", data.Defaults.Temperature)
	}

	if data.Defaults.MaxTokens != 1000 {
		t.Errorf("Expected max_tokens 1000, got %d", data.Defaults.MaxTokens)
	}

	if data.Search.Recency != "week" {
		t.Errorf("Expected recency 'week', got '%s'", data.Search.Recency)
	}

	if data.Search.Mode != "web" {
		t.Errorf("Expected mode 'web', got '%s'", data.Search.Mode)
	}

	if !data.Output.Stream {
		t.Error("Expected stream to be true")
	}

	if data.Output.ReturnImages {
		t.Error("Expected return_images to be false")
	}
}

func TestFindConfigFile(t *testing.T) {
	// Test that FindConfigFile doesn't crash when no config exists
	_, err := FindConfigFile()
	if err == nil {
		// It's okay if a config file is found (user might have one)
		// Just make sure the function doesn't crash
		return
	}

	// Expected error when no config file exists
	if err.Error() != "no config file found in ~/.config/pplx/" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestGetDefaultConfigPath(t *testing.T) {
	path := GetDefaultConfigPath()
	if path == "" {
		t.Error("GetDefaultConfigPath returned empty string")
	}

	// Should contain .config/pplx or be ./pplx.yaml
	if path != "./pplx.yaml" && !filepath.IsAbs(path) {
		t.Errorf("GetDefaultConfigPath should return absolute path or ./pplx.yaml, got: %s", path)
	}
}

func TestFileExists(t *testing.T) {
	// Test with non-existent file
	if fileExists("/path/that/does/not/exist/config.yaml") {
		t.Error("fileExists returned true for non-existent file")
	}

	// Test with existing file (this test file itself)
	if !fileExists("loader_test.go") {
		t.Error("fileExists returned false for existing file")
	}
}

func TestListConfigFiles(t *testing.T) {
	// Test with non-existent directory (should return empty list)
	// Temporarily override ConfigPaths
	oldPaths := ConfigPaths
	ConfigPaths = []string{"/path/that/does/not/exist"}
	defer func() { ConfigPaths = oldPaths }()

	files, err := ListConfigFiles()
	if err != nil {
		t.Errorf("ListConfigFiles returned error for non-existent dir: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("Expected empty list for non-existent dir, got %d files", len(files))
	}
}

func TestListConfigFilesWithFiles(t *testing.T) {
	// Create a temporary directory with config files
	tmpDir := t.TempDir()

	// Override ConfigPaths
	oldPaths := ConfigPaths
	ConfigPaths = []string{tmpDir}
	defer func() { ConfigPaths = oldPaths }()

	// Create test config files
	validConfig := `
defaults:
  model: test-model

profiles:
  dev:
    defaults:
      model: dev-model
`
	if err := os.WriteFile(filepath.Join(tmpDir, "config.yaml"), []byte(validConfig), 0644); err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "pplx.yaml"), []byte("defaults:\n  model: pplx-model\n"), 0644); err != nil {
		t.Fatalf("Failed to create test pplx file: %v", err)
	}

	// Create an invalid YAML file
	if err := os.WriteFile(filepath.Join(tmpDir, "config.yml"), []byte("invalid: yaml: content: [\n"), 0644); err != nil {
		t.Fatalf("Failed to create invalid config file: %v", err)
	}

	// Create a custom named config file
	if err := os.WriteFile(filepath.Join(tmpDir, "dev.yaml"), []byte("defaults:\n  model: dev-model\n"), 0644); err != nil {
		t.Fatalf("Failed to create dev config file: %v", err)
	}

	// Create a non-config file (should be ignored)
	if err := os.WriteFile(filepath.Join(tmpDir, "readme.txt"), []byte("ignore me"), 0644); err != nil {
		t.Fatalf("Failed to create readme file: %v", err)
	}

	files, err := ListConfigFiles()
	if err != nil {
		t.Fatalf("ListConfigFiles returned error: %v", err)
	}

	// Should find 4 config files (config.yaml, pplx.yaml, config.yml, dev.yaml)
	if len(files) != 4 {
		t.Errorf("Expected 4 config files, got %d", len(files))
		for i, f := range files {
			t.Logf("  [%d] %s", i, f.Name)
		}
	}

	// Check file ordering (config.yaml should be first)
	if len(files) > 0 && files[0].Name != "config.yaml" {
		t.Errorf("Expected config.yaml to be first, got %s", files[0].Name)
	}

	// Check that valid files are marked as valid
	foundValid := false
	foundInvalid := false
	foundDev := false
	for _, file := range files {
		if file.Name == "config.yaml" {
			if !file.IsValid {
				t.Error("config.yaml should be valid")
			}
			if file.ProfileCount != 1 {
				t.Errorf("config.yaml should have 1 profile, got %d", file.ProfileCount)
			}
			// First file in precedence order should be active
			if files[0].Name == "config.yaml" && !file.IsActive {
				// Get the active file for debugging
				activeFile, _ := FindConfigFile()
				t.Errorf("config.yaml should be active (highest precedence), path=%s, activeFile=%s", file.Path, activeFile)
			}
			foundValid = true
		}
		if file.Name == "config.yml" {
			if file.IsValid {
				t.Error("config.yml should be invalid (bad YAML)")
			}
			foundInvalid = true
		}
		if file.Name == "dev.yaml" {
			if !file.IsValid {
				t.Error("dev.yaml should be valid")
			}
			foundDev = true
		}
	}

	if !foundValid {
		t.Error("Did not find config.yaml in results")
	}
	if !foundInvalid {
		t.Error("Did not find config.yml in results")
	}
	if !foundDev {
		t.Error("Did not find dev.yaml in results")
	}

	// Verify ordering: standard files first, then custom files alphabetically
	// Expected order: config.yaml, pplx.yaml, config.yml, dev.yaml
	expectedOrder := []string{"config.yaml", "pplx.yaml", "config.yml", "dev.yaml"}
	for i, expected := range expectedOrder {
		if i < len(files) && files[i].Name != expected {
			t.Errorf("Expected file at position %d to be %s, got %s", i, expected, files[i].Name)
		}
	}
}
