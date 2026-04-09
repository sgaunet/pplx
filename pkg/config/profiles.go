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

	mergeProfileDefaults(&merged.Defaults, &profile.Defaults)
	mergeProfileSearch(&merged.Search, &profile.Search)
	mergeProfileOutput(&merged.Output, &profile.Output)

	return merged, nil
}

// mergeProfileDefaults applies non-nil ProfileDefaults fields onto a DefaultsConfig.
func mergeProfileDefaults(dst *DefaultsConfig, src *ProfileDefaults) {
	if src.Model != nil {
		dst.Model = *src.Model
	}
	if src.Temperature != nil {
		dst.Temperature = *src.Temperature
	}
	if src.MaxTokens != nil {
		dst.MaxTokens = *src.MaxTokens
	}
	if src.TopK != nil {
		dst.TopK = *src.TopK
	}
	if src.TopP != nil {
		dst.TopP = *src.TopP
	}
	if src.FrequencyPenalty != nil {
		dst.FrequencyPenalty = *src.FrequencyPenalty
	}
	if src.PresencePenalty != nil {
		dst.PresencePenalty = *src.PresencePenalty
	}
	if src.Timeout != nil {
		dst.Timeout = *src.Timeout
	}
}

// mergeProfileSearch applies non-nil ProfileSearch fields onto a SearchConfig.
func mergeProfileSearch(dst *SearchConfig, src *ProfileSearch) {
	if src.Domains != nil {
		dst.Domains = *src.Domains
	}
	if src.Recency != nil {
		dst.Recency = *src.Recency
	}
	if src.Mode != nil {
		dst.Mode = *src.Mode
	}
	if src.ContextSize != nil {
		dst.ContextSize = *src.ContextSize
	}
	mergeProfileSearchLocation(dst, src)
	mergeProfileSearchDates(dst, src)
}

// mergeProfileSearchLocation applies non-nil location fields onto a SearchConfig.
func mergeProfileSearchLocation(dst *SearchConfig, src *ProfileSearch) {
	if src.LocationLat != nil {
		dst.LocationLat = *src.LocationLat
	}
	if src.LocationLon != nil {
		dst.LocationLon = *src.LocationLon
	}
	if src.LocationCountry != nil {
		dst.LocationCountry = *src.LocationCountry
	}
}

// mergeProfileSearchDates applies non-nil date fields onto a SearchConfig.
func mergeProfileSearchDates(dst *SearchConfig, src *ProfileSearch) {
	if src.AfterDate != nil {
		dst.AfterDate = *src.AfterDate
	}
	if src.BeforeDate != nil {
		dst.BeforeDate = *src.BeforeDate
	}
	if src.LastUpdatedAfter != nil {
		dst.LastUpdatedAfter = *src.LastUpdatedAfter
	}
	if src.LastUpdatedBefore != nil {
		dst.LastUpdatedBefore = *src.LastUpdatedBefore
	}
}

// mergeProfileOutput applies non-nil ProfileOutput fields onto an OutputConfig.
// Nil means "not set"; non-nil overrides base including false for booleans.
func mergeProfileOutput(dst *OutputConfig, src *ProfileOutput) {
	if src.Stream != nil {
		dst.Stream = *src.Stream
	}
	if src.ReturnImages != nil {
		dst.ReturnImages = *src.ReturnImages
	}
	if src.ReturnRelated != nil {
		dst.ReturnRelated = *src.ReturnRelated
	}
	if src.JSON != nil {
		dst.JSON = *src.JSON
	}
	if src.ImageDomains != nil {
		dst.ImageDomains = *src.ImageDomains
	}
	if src.ImageFormats != nil {
		dst.ImageFormats = *src.ImageFormats
	}
	if src.ResponseFormatJSONSchema != nil {
		dst.ResponseFormatJSONSchema = *src.ResponseFormatJSONSchema
	}
	if src.ResponseFormatRegex != nil {
		dst.ResponseFormatRegex = *src.ResponseFormatRegex
	}
	if src.ReasoningEffort != nil {
		dst.ReasoningEffort = *src.ReasoningEffort
	}
}

// CloneProfile creates a new profile as a deep copy of an existing profile.
func (pm *ProfileManager) CloneProfile(name, sourceName string) (*Profile, error) {
	if name == "" {
		return nil, clerrors.ErrProfileNameEmpty
	}

	if name == DefaultProfileName {
		return nil, clerrors.ErrProfileNameReserved
	}

	if _, exists := pm.data.Profiles[name]; exists {
		return nil, fmt.Errorf("%w: '%s'", clerrors.ErrProfileAlreadyExists, name)
	}

	src, err := pm.LoadProfile(sourceName)
	if err != nil {
		return nil, err
	}

	clone := &Profile{
		Name:        name,
		Description: src.Description,
		Defaults: ProfileDefaults{
			Model:            copyStringPtr(src.Defaults.Model),
			Temperature:      copyFloat64Ptr(src.Defaults.Temperature),
			MaxTokens:        copyIntPtr(src.Defaults.MaxTokens),
			TopK:             copyIntPtr(src.Defaults.TopK),
			TopP:             copyFloat64Ptr(src.Defaults.TopP),
			FrequencyPenalty: copyFloat64Ptr(src.Defaults.FrequencyPenalty),
			PresencePenalty:  copyFloat64Ptr(src.Defaults.PresencePenalty),
			Timeout:          copyStringPtr(src.Defaults.Timeout),
		},
		Search: ProfileSearch{
			Domains:           copyStringSlicePtr(src.Search.Domains),
			Recency:           copyStringPtr(src.Search.Recency),
			Mode:              copyStringPtr(src.Search.Mode),
			ContextSize:       copyStringPtr(src.Search.ContextSize),
			LocationLat:       copyFloat64Ptr(src.Search.LocationLat),
			LocationLon:       copyFloat64Ptr(src.Search.LocationLon),
			LocationCountry:   copyStringPtr(src.Search.LocationCountry),
			AfterDate:         copyStringPtr(src.Search.AfterDate),
			BeforeDate:        copyStringPtr(src.Search.BeforeDate),
			LastUpdatedAfter:  copyStringPtr(src.Search.LastUpdatedAfter),
			LastUpdatedBefore: copyStringPtr(src.Search.LastUpdatedBefore),
		},
		Output: ProfileOutput{
			Stream:                   copyBoolPtr(src.Output.Stream),
			ReturnImages:             copyBoolPtr(src.Output.ReturnImages),
			ReturnRelated:            copyBoolPtr(src.Output.ReturnRelated),
			JSON:                     copyBoolPtr(src.Output.JSON),
			ImageDomains:             copyStringSlicePtr(src.Output.ImageDomains),
			ImageFormats:             copyStringSlicePtr(src.Output.ImageFormats),
			ResponseFormatJSONSchema: copyStringPtr(src.Output.ResponseFormatJSONSchema),
			ResponseFormatRegex:      copyStringPtr(src.Output.ResponseFormatRegex),
			ReasoningEffort:          copyStringPtr(src.Output.ReasoningEffort),
		},
	}

	pm.data.Profiles[name] = clone
	return clone, nil
}

// getDefaultProfile returns the default profile.
// The default profile carries no pointer overrides; merging it with the base
// config simply returns the base config unchanged.
func (pm *ProfileManager) getDefaultProfile() *Profile {
	return &Profile{
		Name:        DefaultProfileName,
		Description: "Default configuration",
	}
}

// copyStringPtr returns a new *string with the same value, or nil if p is nil.
func copyStringPtr(p *string) *string {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}

// copyFloat64Ptr returns a new *float64 with the same value, or nil if p is nil.
func copyFloat64Ptr(p *float64) *float64 {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}

// copyIntPtr returns a new *int with the same value, or nil if p is nil.
func copyIntPtr(p *int) *int {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}

// copyBoolPtr returns a new *bool with the same value, or nil if p is nil.
func copyBoolPtr(p *bool) *bool {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}

// copyStringSlicePtr returns a new *[]string backed by a fresh slice copy, or nil if p is nil.
func copyStringSlicePtr(p *[]string) *[]string {
	if p == nil {
		return nil
	}
	cp := make([]string, len(*p))
	copy(cp, *p)
	return &cp
}
