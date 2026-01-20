package config

import (
	"fmt"

	"github.com/sgaunet/pplx/pkg/clerrors"
)

// ProfileManager manages configuration profiles.
type ProfileManager struct {
	data *ConfigData
}

// NewProfileManager creates a new profile manager.
func NewProfileManager(data *ConfigData) *ProfileManager {
	if data.Profiles == nil {
		data.Profiles = make(map[string]*Profile)
	}
	return &ProfileManager{
		data: data,
	}
}

// CreateProfile creates a new profile.
func (pm *ProfileManager) CreateProfile(name, description string) (*Profile, error) {
	if name == "" {
		return nil, clerrors.ErrProfileNameEmpty
	}

	if name == DefaultProfileName {
		return nil, clerrors.ErrProfileNameReserved
	}

	if _, exists := pm.data.Profiles[name]; exists {
		return nil, fmt.Errorf("%w: '%s'", clerrors.ErrProfileAlreadyExists, name)
	}

	profile := &Profile{
		Name:        name,
		Description: description,
	}

	pm.data.Profiles[name] = profile
	return profile, nil
}

// LoadProfile retrieves a profile by name.
func (pm *ProfileManager) LoadProfile(name string) (*Profile, error) {
	// Handle default profile
	if name == DefaultProfileName || name == "" {
		return pm.getDefaultProfile(), nil
	}

	profile, exists := pm.data.Profiles[name]
	if !exists {
		return nil, fmt.Errorf("%w: '%s'", clerrors.ErrProfileNotFound, name)
	}

	return profile, nil
}

// ListProfiles returns all available profiles.
func (pm *ProfileManager) ListProfiles() []string {
	profiles := []string{DefaultProfileName}

	for name := range pm.data.Profiles {
		if name != DefaultProfileName {
			profiles = append(profiles, name)
		}
	}

	return profiles
}

// DeleteProfile removes a profile.
func (pm *ProfileManager) DeleteProfile(name string) error {
	if name == DefaultProfileName {
		return clerrors.ErrDeleteDefaultProfile
	}

	if _, exists := pm.data.Profiles[name]; !exists {
		return fmt.Errorf("%w: '%s'", clerrors.ErrProfileNotFound, name)
	}

	// If deleting the active profile, switch to default
	if pm.data.ActiveProfile == name {
		pm.data.ActiveProfile = DefaultProfileName
	}

	delete(pm.data.Profiles, name)
	return nil
}

// SetActiveProfile sets the active profile.
func (pm *ProfileManager) SetActiveProfile(name string) error {
	if name != DefaultProfileName {
		if _, exists := pm.data.Profiles[name]; !exists {
			return fmt.Errorf("%w: '%s'", clerrors.ErrProfileNotFound, name)
		}
	}

	pm.data.ActiveProfile = name
	return nil
}

// GetActiveProfile returns the currently active profile.
func (pm *ProfileManager) GetActiveProfile() (*Profile, error) {
	if pm.data.ActiveProfile == "" || pm.data.ActiveProfile == DefaultProfileName {
		return pm.getDefaultProfile(), nil
	}

	return pm.LoadProfile(pm.data.ActiveProfile)
}

// GetActiveProfileName returns the name of the active profile.
func (pm *ProfileManager) GetActiveProfileName() string {
	if pm.data.ActiveProfile == "" {
		return DefaultProfileName
	}
	return pm.data.ActiveProfile
}

// UpdateProfile updates an existing profile.
func (pm *ProfileManager) UpdateProfile(name string, profile *Profile) error {
	if name == DefaultProfileName {
		return clerrors.ErrUpdateDefaultProfile
	}

	if _, exists := pm.data.Profiles[name]; !exists {
		return fmt.Errorf("%w: '%s'", clerrors.ErrProfileNotFound, name)
	}

	// Ensure name consistency
	profile.Name = name
	pm.data.Profiles[name] = profile
	return nil
}

// ExportProfile exports a profile for sharing.
func (pm *ProfileManager) ExportProfile(name string) (*Profile, error) {
	profile, err := pm.LoadProfile(name)
	if err != nil {
		return nil, err
	}

	// Create a copy to avoid modifying the original
	exported := &Profile{
		Name:        profile.Name,
		Description: profile.Description,
		Defaults:    profile.Defaults,
		Search:      profile.Search,
		Output:      profile.Output,
	}

	return exported, nil
}

// ImportProfile imports a profile.
func (pm *ProfileManager) ImportProfile(profile *Profile, overwrite bool) error {
	if profile.Name == "" {
		return clerrors.ErrProfileNameEmpty
	}

	if profile.Name == DefaultProfileName {
		return clerrors.ErrImportReservedName
	}

	if _, exists := pm.data.Profiles[profile.Name]; exists && !overwrite {
		return fmt.Errorf("%w: '%s' (use overwrite=true to replace)", clerrors.ErrProfileAlreadyExists, profile.Name)
	}

	pm.data.Profiles[profile.Name] = profile
	return nil
}

// MergeProfile merges a profile with the base configuration
// Returns a new ConfigData with the profile settings applied.
//nolint:cyclop // Function complexity is inherent - it merges multiple config sections.
func (pm *ProfileManager) MergeProfile(profileName string) (*ConfigData, error) {
	profile, err := pm.LoadProfile(profileName)
	if err != nil {
		return nil, err
	}

	// Start with base config
	merged := &ConfigData{
		Defaults:      pm.data.Defaults,
		Search:        pm.data.Search,
		Output:        pm.data.Output,
		API:           pm.data.API,
		Profiles:      pm.data.Profiles,
		ActiveProfile: pm.data.ActiveProfile,
	}

	// Override with profile settings (only non-zero values)
	if profile.Defaults.Model != "" {
		merged.Defaults.Model = profile.Defaults.Model
	}
	if profile.Defaults.Temperature != 0 {
		merged.Defaults.Temperature = profile.Defaults.Temperature
	}
	if profile.Defaults.MaxTokens != 0 {
		merged.Defaults.MaxTokens = profile.Defaults.MaxTokens
	}
	if profile.Defaults.TopK != 0 {
		merged.Defaults.TopK = profile.Defaults.TopK
	}
	if profile.Defaults.TopP != 0 {
		merged.Defaults.TopP = profile.Defaults.TopP
	}
	if profile.Defaults.FrequencyPenalty != 0 {
		merged.Defaults.FrequencyPenalty = profile.Defaults.FrequencyPenalty
	}
	if profile.Defaults.PresencePenalty != 0 {
		merged.Defaults.PresencePenalty = profile.Defaults.PresencePenalty
	}
	if profile.Defaults.Timeout != "" {
		merged.Defaults.Timeout = profile.Defaults.Timeout
	}

	// Override search settings
	if len(profile.Search.Domains) > 0 {
		merged.Search.Domains = profile.Search.Domains
	}
	if profile.Search.Recency != "" {
		merged.Search.Recency = profile.Search.Recency
	}
	if profile.Search.Mode != "" {
		merged.Search.Mode = profile.Search.Mode
	}
	if profile.Search.ContextSize != "" {
		merged.Search.ContextSize = profile.Search.ContextSize
	}

	// Override output settings
	merged.Output.Stream = profile.Output.Stream || merged.Output.Stream
	merged.Output.ReturnImages = profile.Output.ReturnImages || merged.Output.ReturnImages
	merged.Output.ReturnRelated = profile.Output.ReturnRelated || merged.Output.ReturnRelated

	return merged, nil
}

// getDefaultProfile returns the default profile (constructed from base config).
func (pm *ProfileManager) getDefaultProfile() *Profile {
	return &Profile{
		Name:        DefaultProfileName,
		Description: "Default configuration",
		Defaults:    pm.data.Defaults,
		Search:      pm.data.Search,
		Output:      pm.data.Output,
	}
}
