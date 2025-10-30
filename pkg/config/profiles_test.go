package config

import (
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
