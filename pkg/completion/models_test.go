package completion

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestKnownModels(t *testing.T) {
	models := KnownModels()

	if len(models) == 0 {
		t.Error("KnownModels() returned empty slice")
	}

	// Check that default model is in the list
	found := false
	for _, m := range models {
		if m == "sonar" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Default model 'sonar' not found in KnownModels()")
	}
}

func TestSearchModes(t *testing.T) {
	modes := SearchModes()

	expected := []string{"web", "academic"}
	if len(modes) != len(expected) {
		t.Errorf("Expected %d search modes, got %d", len(expected), len(modes))
	}

	for i, mode := range expected {
		if modes[i] != mode {
			t.Errorf("Expected mode %s, got %s", mode, modes[i])
		}
	}
}

func TestRecencyValues(t *testing.T) {
	values := RecencyValues()

	expected := []string{"hour", "day", "week", "month", "year"}
	if len(values) != len(expected) {
		t.Errorf("Expected %d recency values, got %d", len(expected), len(values))
	}

	for i, value := range expected {
		if values[i] != value {
			t.Errorf("Expected value %s, got %s", value, values[i])
		}
	}
}

func TestContextSizes(t *testing.T) {
	sizes := ContextSizes()

	expected := []string{"low", "medium", "high"}
	if len(sizes) != len(expected) {
		t.Errorf("Expected %d context sizes, got %d", len(expected), len(sizes))
	}

	for i, size := range expected {
		if sizes[i] != size {
			t.Errorf("Expected size %s, got %s", size, sizes[i])
		}
	}
}

func TestReasoningEfforts(t *testing.T) {
	efforts := ReasoningEfforts()

	expected := []string{"low", "medium", "high"}
	if len(efforts) != len(expected) {
		t.Errorf("Expected %d reasoning efforts, got %d", len(expected), len(efforts))
	}

	for i, effort := range expected {
		if efforts[i] != effort {
			t.Errorf("Expected effort %s, got %s", effort, efforts[i])
		}
	}
}

func TestImageFormats(t *testing.T) {
	formats := ImageFormats()

	if len(formats) == 0 {
		t.Error("ImageFormats() returned empty slice")
	}

	// Check for common formats
	commonFormats := []string{"jpg", "png", "gif"}
	for _, format := range commonFormats {
		found := false
		for _, f := range formats {
			if f == format {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Common format %s not found in ImageFormats()", format)
		}
	}
}

func TestCommonDomains(t *testing.T) {
	domains := CommonDomains()

	if len(domains) == 0 {
		t.Error("CommonDomains() returned empty slice")
	}

	// Check for some common domains
	expectedDomains := []string{"github.com", "stackoverflow.com"}
	for _, domain := range expectedDomains {
		found := false
		for _, d := range domains {
			if d == domain {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected domain %s not found in CommonDomains()", domain)
		}
	}
}

func TestGetCacheDir(t *testing.T) {
	cacheDir, err := GetCacheDir()
	if err != nil {
		t.Fatalf("GetCacheDir() returned error: %v", err)
	}

	if cacheDir == "" {
		t.Error("GetCacheDir() returned empty string")
	}

	// Check that directory exists (GetCacheDir creates it)
	info, err := os.Stat(cacheDir)
	if err != nil {
		t.Errorf("Cache directory does not exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("Cache path is not a directory")
	}
}

func TestSaveAndGetCachedModels(t *testing.T) {
	// Create a temporary cache directory
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")

	// Set HOME to temp directory for testing
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	testModels := []string{"model1", "model2", "model3"}

	// Save models to cache
	err := SaveModelsToCache(testModels)
	if err != nil {
		t.Fatalf("SaveModelsToCache() returned error: %v", err)
	}

	// Retrieve cached models
	cached, err := GetCachedModels()
	if err != nil {
		t.Fatalf("GetCachedModels() returned error: %v", err)
	}

	// Verify models match
	if len(cached) != len(testModels) {
		t.Errorf("Expected %d cached models, got %d", len(testModels), len(cached))
	}

	for i, model := range testModels {
		if cached[i] != model {
			t.Errorf("Expected model %s at index %d, got %s", model, i, cached[i])
		}
	}
}

func TestGetCachedModelsExpired(t *testing.T) {
	// Create a temporary cache directory
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")

	// Set HOME to temp directory for testing
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	// Create an expired cache
	cacheDir := filepath.Join(tempDir, ".cache", "pplx")
	err := os.MkdirAll(cacheDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create cache directory: %v", err)
	}

	cachePath := filepath.Join(cacheDir, "models.json")

	// Create a cache with an old timestamp
	cache := ModelCache{
		Models:    []string{"old-model"},
		UpdatedAt: time.Now().Add(-25 * time.Hour), // Expired (older than CacheTTL)
	}

	data := `{"models":["old-model"],"updated_at":"` + cache.UpdatedAt.Format(time.RFC3339) + `"}`
	err = os.WriteFile(cachePath, []byte(data), 0644)
	if err != nil {
		t.Fatalf("Failed to write cache file: %v", err)
	}

	// Get cached models should return known models when cache is expired
	cached, err := GetCachedModels()
	if err != nil {
		t.Fatalf("GetCachedModels() returned error: %v", err)
	}

	// Should return KnownModels instead of expired cache
	knownModels := KnownModels()
	if len(cached) != len(knownModels) {
		t.Errorf("Expected %d models (from KnownModels), got %d", len(knownModels), len(cached))
	}
}

func TestGetCachedModelsNoCache(t *testing.T) {
	// Create a temporary cache directory
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")

	// Set HOME to temp directory for testing
	defer os.Setenv("HOME", originalHome)
	os.Setenv("HOME", tempDir)

	// Get cached models when no cache exists should return known models
	cached, err := GetCachedModels()
	if err != nil {
		t.Fatalf("GetCachedModels() returned error: %v", err)
	}

	knownModels := KnownModels()
	if len(cached) != len(knownModels) {
		t.Errorf("Expected %d models (from KnownModels), got %d", len(knownModels), len(cached))
	}
}

func TestGetModels(t *testing.T) {
	models := GetModels()

	if len(models) == 0 {
		t.Error("GetModels() returned empty slice")
	}

	// Should at least return known models
	knownModels := KnownModels()
	if len(models) < len(knownModels) {
		t.Error("GetModels() returned fewer models than KnownModels()")
	}
}

func TestTemplateNames(t *testing.T) {
	t.Parallel()

	names := TemplateNames()
	if len(names) != 4 {
		t.Errorf("TemplateNames() returned %d names, want 4", len(names))
	}

	expected := []string{"research", "creative", "news", "full-example"}
	for i, name := range expected {
		if names[i] != name {
			t.Errorf("TemplateNames()[%d] = %v, want %v", i, names[i], name)
		}
	}
}

func TestConfigSections(t *testing.T) {
	t.Parallel()

	sections := ConfigSections()
	if len(sections) != 4 {
		t.Errorf("ConfigSections() returned %d sections, want 4", len(sections))
	}

	expected := []string{"defaults", "search", "output", "api"}
	for i, section := range expected {
		if sections[i] != section {
			t.Errorf("ConfigSections()[%d] = %v, want %v", i, sections[i], section)
		}
	}
}

func TestOutputFormats(t *testing.T) {
	t.Parallel()

	formats := OutputFormats()
	if len(formats) != 3 {
		t.Errorf("OutputFormats() returned %d formats, want 3", len(formats))
	}

	expected := []string{"table", "json", "yaml"}
	for i, format := range expected {
		if formats[i] != format {
			t.Errorf("OutputFormats()[%d] = %v, want %v", i, formats[i], format)
		}
	}
}
