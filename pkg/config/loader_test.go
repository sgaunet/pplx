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
	if err.Error() != "no config file found in standard locations" {
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
