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

// =============================================================================
// File System Edge Case Tests
// =============================================================================

func TestLoadFrom_FilePermissionDenied(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create config file
	configContent := "defaults:\n  model: test\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Make file unreadable (chmod 000)
	if err := os.Chmod(configPath, 0000); err != nil {
		t.Fatalf("Failed to chmod file: %v", err)
	}

	// Restore permissions for cleanup
	defer func() {
		_ = os.Chmod(configPath, 0644)
	}()

	loader := NewLoader()
	err := loader.LoadFrom(configPath)

	// Should get permission error
	if err == nil {
		t.Error("Expected error for permission denied file")
	}
}

func TestLoadFrom_CorruptedYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create file with invalid structure (valid YAML syntax, but missing expected fields)
	configContent := `
this_is_valid_yaml: true
but:
  - not
  - the
  - expected
  - structure
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewLoader()
	err := loader.LoadFrom(configPath)

	// May not error (depends on implementation), but should not crash
	if err != nil {
		t.Logf("Corrupted YAML returned error (expected): %v", err)
	}

	// Verify loader doesn't crash
	data := loader.Data()
	if data == nil {
		t.Error("Data should not be nil even with corrupted YAML")
	}
}

func TestLoadFrom_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create empty file (0 bytes)
	if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	loader := NewLoader()
	err := loader.LoadFrom(configPath)

	// Empty file might be considered valid or invalid depending on implementation
	if err != nil {
		t.Logf("Empty file returned error: %v", err)
	}

	data := loader.Data()
	if data == nil {
		t.Error("Data should not be nil for empty file")
	}
}

func TestLoadFrom_OnlyWhitespace(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create file with only whitespace
	configContent := "   \n\n   \n\t\t\n   "
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewLoader()
	err := loader.LoadFrom(configPath)

	// Should handle whitespace gracefully
	if err != nil {
		t.Logf("Whitespace-only file returned error: %v", err)
	}

	data := loader.Data()
	if data == nil {
		t.Error("Data should not be nil for whitespace-only file")
	}
}

func TestLoadFrom_VeryLargeFile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large file test in short mode")
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "large-config.yaml")

	// Create a ~1MB config file with many profiles
	var configContent string
	configContent += "defaults:\n  model: test\n\nprofiles:\n"

	for i := 0; i < 1000; i++ {
		configContent += "  profile" + string(rune('0'+i%10)) + string(rune('0'+(i/10)%10)) + string(rune('0'+(i/100)%10)) + ":\n"
		configContent += "    name: profile" + string(rune('0'+i%10)) + string(rune('0'+(i/10)%10)) + string(rune('0'+(i/100)%10)) + "\n"
		configContent += "    defaults:\n"
		configContent += "      model: model" + string(rune('0'+i%10)) + "\n"
		configContent += "      temperature: 0.7\n"
	}

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create large file: %v", err)
	}

	loader := NewLoader()
	err := loader.LoadFrom(configPath)

	// Should handle large file without crashing
	if err != nil {
		t.Logf("Large file load error: %v", err)
	} else {
		data := loader.Data()
		if data == nil {
			t.Error("Data should not be nil for large file")
		}
		t.Logf("Successfully loaded large config file")
	}
}

func TestLoad_SymbolicLinks(t *testing.T) {
	tmpDir := t.TempDir()
	realPath := filepath.Join(tmpDir, "real-config.yaml")
	symlinkPath := filepath.Join(tmpDir, "config.yaml")

	// Create real config file
	configContent := "defaults:\n  model: symlink-test\n"
	if err := os.WriteFile(realPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create real file: %v", err)
	}

	// Create symlink
	if err := os.Symlink(realPath, symlinkPath); err != nil {
		t.Skipf("Cannot create symlink (may not be supported on this platform): %v", err)
	}

	loader := NewLoader()
	err := loader.LoadFrom(symlinkPath)

	// Should follow symlink and load successfully
	if err != nil {
		t.Errorf("Failed to load config through symlink: %v", err)
	}

	data := loader.Data()
	if data.Defaults.Model != "symlink-test" {
		t.Errorf("Expected model 'symlink-test', got '%s'", data.Defaults.Model)
	}
}

func TestLoad_PathWithSpaces(t *testing.T) {
	tmpDir := t.TempDir()
	spacedDir := filepath.Join(tmpDir, "path with spaces")
	if err := os.Mkdir(spacedDir, 0755); err != nil {
		t.Fatalf("Failed to create directory with spaces: %v", err)
	}

	configPath := filepath.Join(spacedDir, "config.yaml")
	configContent := "defaults:\n  model: spaced-path\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config in spaced path: %v", err)
	}

	loader := NewLoader()
	err := loader.LoadFrom(configPath)

	// Should handle paths with spaces
	if err != nil {
		t.Errorf("Failed to load config from path with spaces: %v", err)
	}

	data := loader.Data()
	if data.Defaults.Model != "spaced-path" {
		t.Errorf("Expected model 'spaced-path', got '%s'", data.Defaults.Model)
	}
}

func TestLoad_PathWithUnicode(t *testing.T) {
	tmpDir := t.TempDir()
	unicodeDir := filepath.Join(tmpDir, "路径-café")
	if err := os.Mkdir(unicodeDir, 0755); err != nil {
		t.Skipf("Cannot create directory with unicode (may not be supported): %v", err)
	}

	configPath := filepath.Join(unicodeDir, "config.yaml")
	configContent := "defaults:\n  model: unicode-path\n"
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to create config in unicode path: %v", err)
	}

	loader := NewLoader()
	err := loader.LoadFrom(configPath)

	// Should handle unicode paths
	if err != nil {
		t.Errorf("Failed to load config from unicode path: %v", err)
	}

	data := loader.Data()
	if data.Defaults.Model != "unicode-path" {
		t.Errorf("Expected model 'unicode-path', got '%s'", data.Defaults.Model)
	}
}
