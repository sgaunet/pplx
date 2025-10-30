package config

import (
	"time"
)

// Config interface defines methods for configuration management
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

// ConfigData represents the complete configuration structure
type ConfigData struct {
	// Defaults contains default values for CLI flags
	Defaults DefaultsConfig `mapstructure:"defaults" yaml:"defaults,omitempty"`

	// Search contains search-related preferences
	Search SearchConfig `mapstructure:"search" yaml:"search,omitempty"`

	// Output contains output-related preferences
	Output OutputConfig `mapstructure:"output" yaml:"output,omitempty"`

	// API contains API configuration
	API APIConfig `mapstructure:"api" yaml:"api,omitempty"`

	// Profiles contains named configuration profiles
	Profiles map[string]*Profile `mapstructure:"profiles" yaml:"profiles,omitempty"`

	// ActiveProfile is the name of the currently active profile
	ActiveProfile string `mapstructure:"active_profile" yaml:"active_profile,omitempty"`
}

// DefaultsConfig contains default values for common options
type DefaultsConfig struct {
	Model            string  `mapstructure:"model" yaml:"model,omitempty"`
	Temperature      float64 `mapstructure:"temperature" yaml:"temperature,omitempty"`
	MaxTokens        int     `mapstructure:"max_tokens" yaml:"max_tokens,omitempty"`
	TopK             int     `mapstructure:"top_k" yaml:"top_k,omitempty"`
	TopP             float64 `mapstructure:"top_p" yaml:"top_p,omitempty"`
	FrequencyPenalty float64 `mapstructure:"frequency_penalty" yaml:"frequency_penalty,omitempty"`
	PresencePenalty  float64 `mapstructure:"presence_penalty" yaml:"presence_penalty,omitempty"`
	Timeout          string  `mapstructure:"timeout" yaml:"timeout,omitempty"`
}

// SearchConfig contains search-related preferences
type SearchConfig struct {
	Domains       []string `mapstructure:"domains" yaml:"domains,omitempty"`
	Recency       string   `mapstructure:"recency" yaml:"recency,omitempty"`
	Mode          string   `mapstructure:"mode" yaml:"mode,omitempty"`
	ContextSize   string   `mapstructure:"context_size" yaml:"context_size,omitempty"`

	// Location preferences
	LocationLat     float64 `mapstructure:"location_lat" yaml:"location_lat,omitempty"`
	LocationLon     float64 `mapstructure:"location_lon" yaml:"location_lon,omitempty"`
	LocationCountry string  `mapstructure:"location_country" yaml:"location_country,omitempty"`

	// Date filtering
	AfterDate        string `mapstructure:"after_date" yaml:"after_date,omitempty"`
	BeforeDate       string `mapstructure:"before_date" yaml:"before_date,omitempty"`
	LastUpdatedAfter string `mapstructure:"last_updated_after" yaml:"last_updated_after,omitempty"`
	LastUpdatedBefore string `mapstructure:"last_updated_before" yaml:"last_updated_before,omitempty"`
}

// OutputConfig contains output-related preferences
type OutputConfig struct {
	Stream        bool     `mapstructure:"stream" yaml:"stream,omitempty"`
	ReturnImages  bool     `mapstructure:"return_images" yaml:"return_images,omitempty"`
	ReturnRelated bool     `mapstructure:"return_related" yaml:"return_related,omitempty"`
	JSON          bool     `mapstructure:"json" yaml:"json,omitempty"`

	// Image filtering
	ImageDomains []string `mapstructure:"image_domains" yaml:"image_domains,omitempty"`
	ImageFormats []string `mapstructure:"image_formats" yaml:"image_formats,omitempty"`

	// Response format
	ResponseFormatJSONSchema string `mapstructure:"response_format_json_schema" yaml:"response_format_json_schema,omitempty"`
	ResponseFormatRegex      string `mapstructure:"response_format_regex" yaml:"response_format_regex,omitempty"`

	// Deep research
	ReasoningEffort string `mapstructure:"reasoning_effort" yaml:"reasoning_effort,omitempty"`
}

// APIConfig contains API-related configuration
type APIConfig struct {
	Key     string        `mapstructure:"key" yaml:"key,omitempty"`
	BaseURL string        `mapstructure:"base_url" yaml:"base_url,omitempty"`
	Timeout time.Duration `mapstructure:"timeout" yaml:"timeout,omitempty"`
}

// Profile represents a named configuration profile
type Profile struct {
	Name        string         `mapstructure:"name" yaml:"name"`
	Description string         `mapstructure:"description" yaml:"description,omitempty"`
	Defaults    DefaultsConfig `mapstructure:"defaults" yaml:"defaults,omitempty"`
	Search      SearchConfig   `mapstructure:"search" yaml:"search,omitempty"`
	Output      OutputConfig   `mapstructure:"output" yaml:"output,omitempty"`
}

// NewConfigData creates a new ConfigData with default values
func NewConfigData() *ConfigData {
	return &ConfigData{
		Profiles:      make(map[string]*Profile),
		ActiveProfile: "default",
	}
}
