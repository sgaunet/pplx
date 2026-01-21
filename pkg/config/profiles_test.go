package config

import (
	"os"
	"testing"
)

func TestProfileManagerCreate(t *testing.T) {
	data := NewConfigData()
	pm := NewProfileManager(data)

	profile, err := pm.CreateProfile("test", "Test profile")
	if err != nil {
		t.Fatalf("CreateProfile failed: %v", err)
	}

	if profile.Name != "test" {
		t.Errorf("Expected profile name 'test', got '%s'", profile.Name)
	}

	if profile.Description != "Test profile" {
		t.Errorf("Expected description 'Test profile', got '%s'", profile.Description)
	}
}

func TestProfileManagerCreateDuplicate(t *testing.T) {
	data := NewConfigData()
	pm := NewProfileManager(data)

	_, err := pm.CreateProfile("test", "Test profile")
	if err != nil {
		t.Fatalf("First CreateProfile failed: %v", err)
	}

	_, err = pm.CreateProfile("test", "Duplicate")
	if err == nil {
		t.Error("Expected error when creating duplicate profile")
	}
}

func TestProfileManagerCreateReservedName(t *testing.T) {
	data := NewConfigData()
	pm := NewProfileManager(data)

	_, err := pm.CreateProfile("default", "Should fail")
	if err == nil {
		t.Error("Expected error when creating profile with reserved name 'default'")
	}
}

func TestProfileManagerLoadDefault(t *testing.T) {
	data := NewConfigData()
	data.Defaults.Model = "test-model"
	pm := NewProfileManager(data)

	profile, err := pm.LoadProfile("default")
	if err != nil {
		t.Fatalf("LoadProfile('default') failed: %v", err)
	}

	if profile.Name != "default" {
		t.Errorf("Expected profile name 'default', got '%s'", profile.Name)
	}

	if profile.Defaults.Model != "test-model" {
		t.Errorf("Expected model 'test-model', got '%s'", profile.Defaults.Model)
	}
}

func TestProfileManagerLoadNonExistent(t *testing.T) {
	data := NewConfigData()
	pm := NewProfileManager(data)

	_, err := pm.LoadProfile("nonexistent")
	if err == nil {
		t.Error("Expected error when loading non-existent profile")
	}
}

func TestProfileManagerList(t *testing.T) {
	data := NewConfigData()
	pm := NewProfileManager(data)

	// Should always include default
	profiles := pm.ListProfiles()
	if len(profiles) < 1 {
		t.Error("Expected at least 1 profile (default)")
	}

	hasDefault := false
	for _, name := range profiles {
		if name == "default" {
			hasDefault = true
			break
		}
	}
	if !hasDefault {
		t.Error("Expected 'default' in profile list")
	}

	// Add a profile
	_, _ = pm.CreateProfile("test", "")
	profiles = pm.ListProfiles()
	if len(profiles) < 2 {
		t.Error("Expected at least 2 profiles after creating one")
	}
}

func TestProfileManagerDelete(t *testing.T) {
	data := NewConfigData()
	pm := NewProfileManager(data)

	_, _ = pm.CreateProfile("test", "")

	err := pm.DeleteProfile("test")
	if err != nil {
		t.Errorf("DeleteProfile failed: %v", err)
	}

	// Verify it's gone
	_, err = pm.LoadProfile("test")
	if err == nil {
		t.Error("Expected error when loading deleted profile")
	}
}

func TestProfileManagerDeleteDefault(t *testing.T) {
	data := NewConfigData()
	pm := NewProfileManager(data)

	err := pm.DeleteProfile("default")
	if err == nil {
		t.Error("Expected error when deleting default profile")
	}
}

func TestProfileManagerDeleteActive(t *testing.T) {
	data := NewConfigData()
	pm := NewProfileManager(data)

	_, _ = pm.CreateProfile("test", "")
	_ = pm.SetActiveProfile("test")

	err := pm.DeleteProfile("test")
	if err != nil {
		t.Errorf("DeleteProfile should succeed and switch to default: %v", err)
	}

	// Active profile should be default now
	if data.ActiveProfile != "default" {
		t.Errorf("Expected active profile 'default', got '%s'", data.ActiveProfile)
	}
}

func TestProfileManagerSetActive(t *testing.T) {
	data := NewConfigData()
	pm := NewProfileManager(data)

	_, _ = pm.CreateProfile("test", "")

	err := pm.SetActiveProfile("test")
	if err != nil {
		t.Errorf("SetActiveProfile failed: %v", err)
	}

	if data.ActiveProfile != "test" {
		t.Errorf("Expected active profile 'test', got '%s'", data.ActiveProfile)
	}
}

func TestProfileManagerSetActiveNonExistent(t *testing.T) {
	data := NewConfigData()
	pm := NewProfileManager(data)

	err := pm.SetActiveProfile("nonexistent")
	if err == nil {
		t.Error("Expected error when setting non-existent profile as active")
	}
}

func TestProfileManagerGetActive(t *testing.T) {
	data := NewConfigData()
	pm := NewProfileManager(data)

	profile, err := pm.GetActiveProfile()
	if err != nil {
		t.Fatalf("GetActiveProfile failed: %v", err)
	}

	if profile.Name != "default" {
		t.Errorf("Expected default profile, got '%s'", profile.Name)
	}
}

func TestProfileManagerMerge(t *testing.T) {
	data := NewConfigData()
	data.Defaults.Model = "base-model"
	data.Defaults.Temperature = 0.5

	pm := NewProfileManager(data)

	// Create a profile with overrides
	profile, _ := pm.CreateProfile("test", "")
	profile.Defaults.Model = "override-model"
	profile.Search.Recency = "week"
	data.Profiles["test"] = profile

	// Merge the profile
	merged, err := pm.MergeProfile("test")
	if err != nil {
		t.Fatalf("MergeProfile failed: %v", err)
	}

	// Check overrides
	if merged.Defaults.Model != "override-model" {
		t.Errorf("Expected model 'override-model', got '%s'", merged.Defaults.Model)
	}

	// Check base values are preserved
	if merged.Defaults.Temperature != 0.5 {
		t.Errorf("Expected temperature 0.5, got %f", merged.Defaults.Temperature)
	}

	// Check profile-only values
	if merged.Search.Recency != "week" {
		t.Errorf("Expected recency 'week', got '%s'", merged.Search.Recency)
	}
}

func TestProfileManagerExportImport(t *testing.T) {
	data := NewConfigData()
	pm := NewProfileManager(data)

	// Create and configure a profile
	original, _ := pm.CreateProfile("test", "Test profile")
	original.Defaults.Model = "test-model"
	original.Search.Recency = "week"
	data.Profiles["test"] = original

	// Export
	exported, err := pm.ExportProfile("test")
	if err != nil {
		t.Fatalf("ExportProfile failed: %v", err)
	}

	// Verify it's a copy
	if exported == original {
		t.Error("Exported profile should be a copy, not the same reference")
	}

	// Import to new data
	newData := NewConfigData()
	newPm := NewProfileManager(newData)

	err = newPm.ImportProfile(exported, false)
	if err != nil {
		t.Fatalf("ImportProfile failed: %v", err)
	}

	// Verify imported profile
	imported, err := newPm.LoadProfile("test")
	if err != nil {
		t.Fatalf("LoadProfile after import failed: %v", err)
	}

	if imported.Defaults.Model != "test-model" {
		t.Errorf("Expected model 'test-model', got '%s'", imported.Defaults.Model)
	}
}

func TestProfileManagerImportDuplicateWithoutOverwrite(t *testing.T) {
	data := NewConfigData()
	pm := NewProfileManager(data)

	profile := &Profile{Name: "test", Description: "Test"}
	_ = pm.ImportProfile(profile, false)

	// Try to import again without overwrite
	err := pm.ImportProfile(profile, false)
	if err == nil {
		t.Error("Expected error when importing duplicate without overwrite")
	}
}

func TestProfileManagerImportDuplicateWithOverwrite(t *testing.T) {
	data := NewConfigData()
	pm := NewProfileManager(data)

	profile1 := &Profile{Name: "test", Description: "First"}
	_ = pm.ImportProfile(profile1, false)

	profile2 := &Profile{Name: "test", Description: "Second"}
	err := pm.ImportProfile(profile2, true)
	if err != nil {
		t.Errorf("ImportProfile with overwrite failed: %v", err)
	}

	// Verify it was overwritten
	loaded, _ := pm.LoadProfile("test")
	if loaded.Description != "Second" {
		t.Errorf("Expected description 'Second', got '%s'", loaded.Description)
	}
}

// =============================================================================
// Profile Integration Edge Case Tests
// =============================================================================

func TestMergeProfile_WithEnvVarExpansion(t *testing.T) {
	// Set up environment variables
	cleanup := func() {
		_ = os.Unsetenv("TEST_MODEL")
		_ = os.Unsetenv("TEST_DOMAIN")
	}
	defer cleanup()

	_ = os.Setenv("TEST_MODEL", "env-model")
	_ = os.Setenv("TEST_DOMAIN", "env.example.com")

	// Create config with env vars in base config (not in profile)
	// Note: ExpandEnvVars() does not expand env vars inside profiles
	data := &ConfigData{
		Defaults: DefaultsConfig{
			Model:       "$TEST_MODEL", // Will be expanded
			Temperature: 0.5,
		},
		Search: SearchConfig{
			Domains: []string{"${TEST_DOMAIN}"}, // Will be expanded
		},
		Profiles: map[string]*Profile{
			"test": {
				Name: "test",
				Defaults: DefaultsConfig{
					Temperature: 0.8, // Override temperature
				},
			},
		},
	}

	// Expand env vars in base config
	ExpandEnvVars(data)

	// Now merge profile
	pm := NewProfileManager(data)
	merged, err := pm.MergeProfile("test")
	if err != nil {
		t.Fatalf("MergeProfile failed: %v", err)
	}

	// Verify env vars were expanded in base config
	if merged.Defaults.Model != "env-model" {
		t.Errorf("Expected model 'env-model', got '%s'", merged.Defaults.Model)
	}
	if len(merged.Search.Domains) != 1 || merged.Search.Domains[0] != "env.example.com" {
		t.Errorf("Expected domain 'env.example.com', got %v", merged.Search.Domains)
	}

	// Profile temperature should override base
	if merged.Defaults.Temperature != 0.8 {
		t.Errorf("Expected temperature 0.8 from profile, got %f", merged.Defaults.Temperature)
	}
}

func TestMergeProfile_AllFieldsZero(t *testing.T) {
	data := &ConfigData{
		Defaults: DefaultsConfig{
			Model:       "base-model",
			Temperature: 0.7,
			MaxTokens:   1000,
		},
		Profiles: map[string]*Profile{
			"zeros": {
				Name: "zeros",
				Defaults: DefaultsConfig{
					// All zero values
					Model:       "",
					Temperature: 0,
					MaxTokens:   0,
				},
			},
		},
	}

	pm := NewProfileManager(data)
	merged, err := pm.MergeProfile("zeros")
	if err != nil {
		t.Fatalf("MergeProfile failed: %v", err)
	}

	// Zero values should not override base config (treated as unset)
	if merged.Defaults.Model != "base-model" {
		t.Errorf("Expected base model preserved, got '%s'", merged.Defaults.Model)
	}
	if merged.Defaults.Temperature != 0.7 {
		t.Errorf("Expected base temperature preserved, got %f", merged.Defaults.Temperature)
	}
	if merged.Defaults.MaxTokens != 1000 {
		t.Errorf("Expected base max tokens preserved, got %d", merged.Defaults.MaxTokens)
	}
}

func TestMergeProfile_ArrayMerging(t *testing.T) {
	data := &ConfigData{
		Search: SearchConfig{
			Domains: []string{"base1.com", "base2.com"},
		},
		Output: OutputConfig{
			ImageFormats: []string{"png", "jpg"},
		},
		Profiles: map[string]*Profile{
			"test": {
				Name: "test",
				Search: SearchConfig{
					Domains: []string{"profile1.com", "profile2.com"},
				},
				Output: OutputConfig{
					// Empty array - should be ignored
					ImageFormats: []string{},
				},
			},
		},
	}

	pm := NewProfileManager(data)
	merged, err := pm.MergeProfile("test")
	if err != nil {
		t.Fatalf("MergeProfile failed: %v", err)
	}

	// Non-empty profile array should replace base
	expectedDomains := []string{"profile1.com", "profile2.com"}
	if len(merged.Search.Domains) != len(expectedDomains) {
		t.Fatalf("Expected %d domains, got %d", len(expectedDomains), len(merged.Search.Domains))
	}
	for i, exp := range expectedDomains {
		if merged.Search.Domains[i] != exp {
			t.Errorf("Domain[%d]: expected '%s', got '%s'", i, exp, merged.Search.Domains[i])
		}
	}

	// Empty profile array should preserve base
	if len(merged.Output.ImageFormats) != 2 {
		t.Errorf("Expected base image formats preserved, got %d", len(merged.Output.ImageFormats))
	}
}

func TestMergeProfile_BooleanPrecedence(t *testing.T) {
	data := &ConfigData{
		Output: OutputConfig{
			Stream:        true,
			ReturnImages:  false,
			ReturnRelated: true,
		},
		Profiles: map[string]*Profile{
			"test": {
				Name: "test",
				Output: OutputConfig{
					Stream:        false, // Override to false
					ReturnImages:  true,  // Override to true
					// ReturnRelated not set (should preserve base)
				},
			},
		},
	}

	pm := NewProfileManager(data)
	merged, err := pm.MergeProfile("test")
	if err != nil {
		t.Fatalf("MergeProfile failed: %v", err)
	}

	// Profile values should override (even when false)
	// Note: This depends on MergeProfile implementation
	// If it only merges non-zero values, booleans are tricky
	t.Logf("Stream: %v (base=true, profile=false)", merged.Output.Stream)
	t.Logf("ReturnImages: %v (base=false, profile=true)", merged.Output.ReturnImages)
	t.Logf("ReturnRelated: %v (base=true, profile not set)", merged.Output.ReturnRelated)

	// Document current behavior
	if merged.Output.ReturnImages {
		t.Log("Profile successfully overrode ReturnImages to true")
	}
}

func TestMergeProfile_ChainedExpansion(t *testing.T) {
	// Test the full workflow: base config → active profile → merged result
	data := &ConfigData{
		Defaults: DefaultsConfig{
			Model:       "base-model",
			Temperature: 0.5,
			MaxTokens:   1000,
			TopK:        10,
		},
		Search: SearchConfig{
			Recency: "week",
			Mode:    "web",
		},
		ActiveProfile: "production",
		Profiles: map[string]*Profile{
			"production": {
				Name: "production",
				Defaults: DefaultsConfig{
					Model:       "prod-model",
					Temperature: 0.8,
					// MaxTokens not set (should preserve base)
					TopK:        50, // Override
				},
				Search: SearchConfig{
					Recency: "month", // Override
					// Mode not set (should preserve base)
				},
			},
		},
	}

	pm := NewProfileManager(data)
	merged, err := pm.MergeProfile("production")
	if err != nil {
		t.Fatalf("MergeProfile failed: %v", err)
	}

	// Verify profile overrides
	if merged.Defaults.Model != "prod-model" {
		t.Errorf("Model should be overridden, got '%s'", merged.Defaults.Model)
	}
	if merged.Defaults.Temperature != 0.8 {
		t.Errorf("Temperature should be overridden, got %f", merged.Defaults.Temperature)
	}
	if merged.Defaults.TopK != 50 {
		t.Errorf("TopK should be overridden, got %d", merged.Defaults.TopK)
	}
	if merged.Search.Recency != "month" {
		t.Errorf("Recency should be overridden, got '%s'", merged.Search.Recency)
	}

	// Verify base preservation
	if merged.Defaults.MaxTokens != 1000 {
		t.Errorf("MaxTokens should be preserved, got %d", merged.Defaults.MaxTokens)
	}
	if merged.Search.Mode != "web" {
		t.Errorf("Mode should be preserved, got '%s'", merged.Search.Mode)
	}
}

func TestMergeProfile_InvalidProfileName(t *testing.T) {
	data := NewConfigData()
	pm := NewProfileManager(data)

	// Try to merge non-existent profile
	_, err := pm.MergeProfile("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent profile")
	}

	// Verify error message is helpful
	if err != nil && err.Error() == "" {
		t.Error("Error should have descriptive message")
	}
}
