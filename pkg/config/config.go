// Package config provides configuration management for the pplx application.
// It supports YAML configuration files, profiles, and CLI flag integration.
package config

import (
	"time"
)

// DefaultProfileName is the name of the default profile.
const DefaultProfileName = "default"

// Config interface defines methods for configuration management.
type Config interface {
	// Get retrieves a configuration value by key
	Get(key string) interface{}

	// Set sets a configuration value
	Set(key string, value interface{}) error

	// Save persists the configuration to disk
	Save() error

	// Load loads configuration from disk
	Load() error

	// Validate checks if the configuration is valid
	Validate() error

	// GetProfile retrieves a profile by name
	GetProfile(name string) (*Profile, error)

	// SetActiveProfile sets the active profile
	SetActiveProfile(name string) error
}

// ConfigData represents the complete configuration structure.
//nolint:revive // ConfigData name is part of public API; renaming would be a breaking change.
type ConfigData struct {
	// Defaults contains default values for CLI flags
	Defaults DefaultsConfig `json:"defaults,omitempty" mapstructure:"defaults" yaml:"defaults,omitempty"`

	// Search contains search-related preferences
	Search SearchConfig `json:"search,omitempty" mapstructure:"search" yaml:"search,omitempty"`

	// Output contains output-related preferences
	Output OutputConfig `json:"output,omitempty" mapstructure:"output" yaml:"output,omitempty"`

	// API contains API configuration
	API APIConfig `json:"api,omitempty" mapstructure:"api" yaml:"api,omitempty"`

	// Profiles contains named configuration profiles
	Profiles map[string]*Profile `json:"profiles,omitempty" mapstructure:"profiles" yaml:"profiles,omitempty"`

	// ActiveProfile is the name of the currently active profile
	ActiveProfile string `json:"active_profile,omitempty" mapstructure:"active_profile" yaml:"active_profile,omitempty"`
}

// DefaultsConfig contains default values for common options.
type DefaultsConfig struct {
	Model            string  `json:"model,omitempty"             mapstructure:"model"             yaml:"model,omitempty"`
	Temperature      float64 `json:"temperature,omitempty"       mapstructure:"temperature"       yaml:"temperature,omitempty"`       //nolint:lll
	MaxTokens        int     `json:"max_tokens,omitempty"        mapstructure:"max_tokens"        yaml:"max_tokens,omitempty"`        //nolint:lll
	TopK             int     `json:"top_k,omitempty"             mapstructure:"top_k"             yaml:"top_k,omitempty"`
	TopP             float64 `json:"top_p,omitempty"             mapstructure:"top_p"             yaml:"top_p,omitempty"`
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty" mapstructure:"frequency_penalty" yaml:"frequency_penalty,omitempty"` //nolint:lll
	PresencePenalty  float64 `json:"presence_penalty,omitempty"  mapstructure:"presence_penalty"  yaml:"presence_penalty,omitempty"`  //nolint:lll
	Timeout          string  `json:"timeout,omitempty"           mapstructure:"timeout"           yaml:"timeout,omitempty"`
}

// SearchConfig contains search-related preferences.
type SearchConfig struct {
	Domains     []string `json:"domains,omitempty"      mapstructure:"domains"      yaml:"domains,omitempty"`
	Recency     string   `json:"recency,omitempty"      mapstructure:"recency"      yaml:"recency,omitempty"`
	Mode        string   `json:"mode,omitempty"         mapstructure:"mode"         yaml:"mode,omitempty"`
	ContextSize string   `json:"context_size,omitempty" mapstructure:"context_size" yaml:"context_size,omitempty"`

	// Location preferences
	LocationLat     float64 `json:"location_lat,omitempty"     mapstructure:"location_lat"     yaml:"location_lat,omitempty"`     //nolint:lll
	LocationLon     float64 `json:"location_lon,omitempty"     mapstructure:"location_lon"     yaml:"location_lon,omitempty"`     //nolint:lll
	LocationCountry string  `json:"location_country,omitempty" mapstructure:"location_country" yaml:"location_country,omitempty"` //nolint:lll

	// Date filtering
	AfterDate         string `json:"after_date,omitempty"          mapstructure:"after_date"          yaml:"after_date,omitempty"`          //nolint:lll
	BeforeDate        string `json:"before_date,omitempty"         mapstructure:"before_date"         yaml:"before_date,omitempty"`         //nolint:lll
	LastUpdatedAfter  string `json:"last_updated_after,omitempty"  mapstructure:"last_updated_after"  yaml:"last_updated_after,omitempty"`  //nolint:lll
	LastUpdatedBefore string `json:"last_updated_before,omitempty" mapstructure:"last_updated_before" yaml:"last_updated_before,omitempty"` //nolint:lll
}

// OutputConfig contains output-related preferences.
type OutputConfig struct {
	Stream        bool `json:"stream,omitempty"         mapstructure:"stream"         yaml:"stream,omitempty"`
	ReturnImages  bool `json:"return_images,omitempty"  mapstructure:"return_images"  yaml:"return_images,omitempty"`
	ReturnRelated bool `json:"return_related,omitempty" mapstructure:"return_related" yaml:"return_related,omitempty"`
	JSON          bool `json:"json,omitempty"           mapstructure:"json"           yaml:"json,omitempty"`

	// Image filtering
	ImageDomains []string `json:"image_domains,omitempty" mapstructure:"image_domains" yaml:"image_domains,omitempty"`
	ImageFormats []string `json:"image_formats,omitempty" mapstructure:"image_formats" yaml:"image_formats,omitempty"`

	// Response format
	ResponseFormatJSONSchema string `json:"response_format_json_schema,omitempty" mapstructure:"response_format_json_schema" yaml:"response_format_json_schema,omitempty"` //nolint:lll
	ResponseFormatRegex      string `json:"response_format_regex,omitempty"       mapstructure:"response_format_regex"       yaml:"response_format_regex,omitempty"`       //nolint:lll

	// Deep research
	ReasoningEffort string `json:"reasoning_effort,omitempty" mapstructure:"reasoning_effort" yaml:"reasoning_effort,omitempty"` //nolint:lll
}

// APIConfig contains API-related configuration.
type APIConfig struct {
	Key     string        `json:"key,omitempty"      mapstructure:"key"      yaml:"key,omitempty"`
	BaseURL string        `json:"base_url,omitempty" mapstructure:"base_url" yaml:"base_url,omitempty"`
	Timeout time.Duration `json:"timeout,omitempty"  mapstructure:"timeout"  yaml:"timeout,omitempty"`
}

// Profile represents a named configuration profile.
type Profile struct {
	Name        string         `json:"name"                  mapstructure:"name"        yaml:"name"`
	Description string         `json:"description,omitempty" mapstructure:"description" yaml:"description,omitempty"`
	Defaults    DefaultsConfig `json:"defaults,omitempty"    mapstructure:"defaults"    yaml:"defaults,omitempty"`
	Search      SearchConfig   `json:"search,omitempty"      mapstructure:"search"      yaml:"search,omitempty"`
	Output      OutputConfig   `json:"output,omitempty"      mapstructure:"output"      yaml:"output,omitempty"`
}

// ConfigFileInfo represents metadata about a configuration file.
//nolint:revive // ConfigFileInfo name is part of public API; renaming would be a breaking change.
type ConfigFileInfo struct {
	Name         string    `json:"name"`          // File name (e.g., "config.yaml")
	Path         string    `json:"path"`          // Full absolute path
	Size         int64     `json:"size"`          // File size in bytes
	ModTime      time.Time `json:"modified"`      // Last modification time
	IsActive     bool      `json:"is_active"`     // Whether this is the active config
	IsValid      bool      `json:"is_valid"`      // Whether file has valid YAML syntax
	ProfileCount int       `json:"profile_count"` // Number of profiles in file (if valid)
}

// NewConfigData creates a new ConfigData with default values.
func NewConfigData() *ConfigData {
	return &ConfigData{
		Profiles:      make(map[string]*Profile),
		ActiveProfile: DefaultProfileName,
	}
}
