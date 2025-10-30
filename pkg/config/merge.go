package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Merger handles merging configuration from files with CLI flags.
type Merger struct {
	viper *viper.Viper
	data  *ConfigData
}

// NewMerger creates a new configuration merger.
func NewMerger(data *ConfigData) *Merger {
	v := viper.New()
	return &Merger{
		viper: v,
		data:  data,
	}
}

// BindFlags binds cobra command flags to viper keys.
func (m *Merger) BindFlags(cmd *cobra.Command) error {
	// Bind all persistent flags
	if err := m.viper.BindPFlags(cmd.PersistentFlags()); err != nil {
		return fmt.Errorf("failed to bind persistent flags: %w", err)
	}

	// Bind local flags
	if err := m.viper.BindPFlags(cmd.Flags()); err != nil {
		return fmt.Errorf("failed to bind local flags: %w", err)
	}

	return nil
}

// MergeWithFlags merges configuration data with CLI flags
// Precedence: CLI flags > environment variables > config file > defaults.
//nolint:cyclop,funlen // Function complexity is inherent - it checks 29 different CLI flags.
func (m *Merger) MergeWithFlags(cmd *cobra.Command) *ConfigData {
	// Start with config file data as base
	merged := m.data

	// Check if flags were explicitly set and override config values
	// Defaults section
	if cmd.Flags().Changed("model") {
		merged.Defaults.Model = m.viper.GetString("model")
	}
	if cmd.Flags().Changed("temperature") {
		merged.Defaults.Temperature = m.viper.GetFloat64("temperature")
	}
	if cmd.Flags().Changed("max-tokens") {
		merged.Defaults.MaxTokens = m.viper.GetInt("max-tokens")
	}
	if cmd.Flags().Changed("top-k") {
		merged.Defaults.TopK = m.viper.GetInt("top-k")
	}
	if cmd.Flags().Changed("top-p") {
		merged.Defaults.TopP = m.viper.GetFloat64("top-p")
	}
	if cmd.Flags().Changed("frequency-penalty") {
		merged.Defaults.FrequencyPenalty = m.viper.GetFloat64("frequency-penalty")
	}
	if cmd.Flags().Changed("presence-penalty") {
		merged.Defaults.PresencePenalty = m.viper.GetFloat64("presence-penalty")
	}
	if cmd.Flags().Changed("timeout") {
		merged.Defaults.Timeout = m.viper.GetDuration("timeout").String()
	}

	// Search section
	if cmd.Flags().Changed("search-domains") {
		merged.Search.Domains = m.viper.GetStringSlice("search-domains")
	}
	if cmd.Flags().Changed("search-recency") {
		merged.Search.Recency = m.viper.GetString("search-recency")
	}
	if cmd.Flags().Changed("search-mode") {
		merged.Search.Mode = m.viper.GetString("search-mode")
	}
	if cmd.Flags().Changed("search-context-size") {
		merged.Search.ContextSize = m.viper.GetString("search-context-size")
	}
	if cmd.Flags().Changed("location-lat") {
		merged.Search.LocationLat = m.viper.GetFloat64("location-lat")
	}
	if cmd.Flags().Changed("location-lon") {
		merged.Search.LocationLon = m.viper.GetFloat64("location-lon")
	}
	if cmd.Flags().Changed("location-country") {
		merged.Search.LocationCountry = m.viper.GetString("location-country")
	}
	if cmd.Flags().Changed("search-after-date") {
		merged.Search.AfterDate = m.viper.GetString("search-after-date")
	}
	if cmd.Flags().Changed("search-before-date") {
		merged.Search.BeforeDate = m.viper.GetString("search-before-date")
	}
	if cmd.Flags().Changed("last-updated-after") {
		merged.Search.LastUpdatedAfter = m.viper.GetString("last-updated-after")
	}
	if cmd.Flags().Changed("last-updated-before") {
		merged.Search.LastUpdatedBefore = m.viper.GetString("last-updated-before")
	}

	// Output section
	if cmd.Flags().Changed("stream") {
		merged.Output.Stream = m.viper.GetBool("stream")
	}
	if cmd.Flags().Changed("return-images") {
		merged.Output.ReturnImages = m.viper.GetBool("return-images")
	}
	if cmd.Flags().Changed("return-related") {
		merged.Output.ReturnRelated = m.viper.GetBool("return-related")
	}
	if cmd.Flags().Changed("json") {
		merged.Output.JSON = m.viper.GetBool("json")
	}
	if cmd.Flags().Changed("image-domains") {
		merged.Output.ImageDomains = m.viper.GetStringSlice("image-domains")
	}
	if cmd.Flags().Changed("image-formats") {
		merged.Output.ImageFormats = m.viper.GetStringSlice("image-formats")
	}
	if cmd.Flags().Changed("response-format-json-schema") {
		merged.Output.ResponseFormatJSONSchema = m.viper.GetString("response-format-json-schema")
	}
	if cmd.Flags().Changed("response-format-regex") {
		merged.Output.ResponseFormatRegex = m.viper.GetString("response-format-regex")
	}
	if cmd.Flags().Changed("reasoning-effort") {
		merged.Output.ReasoningEffort = m.viper.GetString("reasoning-effort")
	}

	return merged
}

// ApplyToGlobals applies merged configuration to global variables
// This maintains compatibility with existing code that uses global vars.
//nolint:gocognit,gocyclo,cyclop,funlen // Function complexity is inherent - it applies config to many global variables.
func ApplyToGlobals(cfg *ConfigData,
	model *string,
	temperature *float64,
	maxTokens *int,
	topK *int,
	topP *float64,
	frequencyPenalty *float64,
	presencePenalty *float64,
	timeout *time.Duration,
	searchDomains *[]string,
	searchRecency *string,
	locationLat *float64,
	locationLon *float64,
	locationCountry *string,
	returnImages *bool,
	returnRelated *bool,
	stream *bool,
	searchMode *string,
	searchContextSize *string,
) {
	// Apply defaults
	if cfg.Defaults.Model != "" && *model == "" {
		*model = cfg.Defaults.Model
	}
	if cfg.Defaults.Temperature != 0 && *temperature == 0 {
		*temperature = cfg.Defaults.Temperature
	}
	if cfg.Defaults.MaxTokens != 0 && *maxTokens == 0 {
		*maxTokens = cfg.Defaults.MaxTokens
	}
	if cfg.Defaults.TopK != 0 && *topK == 0 {
		*topK = cfg.Defaults.TopK
	}
	if cfg.Defaults.TopP != 0 && *topP == 0 {
		*topP = cfg.Defaults.TopP
	}
	if cfg.Defaults.FrequencyPenalty != 0 && *frequencyPenalty == 0 {
		*frequencyPenalty = cfg.Defaults.FrequencyPenalty
	}
	if cfg.Defaults.PresencePenalty != 0 && *presencePenalty == 0 {
		*presencePenalty = cfg.Defaults.PresencePenalty
	}
	if cfg.Defaults.Timeout != "" {
		if d, err := time.ParseDuration(cfg.Defaults.Timeout); err == nil && *timeout == 0 {
			*timeout = d
		}
	}

	// Apply search config
	if len(cfg.Search.Domains) > 0 && len(*searchDomains) == 0 {
		*searchDomains = cfg.Search.Domains
	}
	if cfg.Search.Recency != "" && *searchRecency == "" {
		*searchRecency = cfg.Search.Recency
	}
	if cfg.Search.LocationLat != 0 && *locationLat == 0 {
		*locationLat = cfg.Search.LocationLat
	}
	if cfg.Search.LocationLon != 0 && *locationLon == 0 {
		*locationLon = cfg.Search.LocationLon
	}
	if cfg.Search.LocationCountry != "" && *locationCountry == "" {
		*locationCountry = cfg.Search.LocationCountry
	}
	if cfg.Search.Mode != "" && *searchMode == "" {
		*searchMode = cfg.Search.Mode
	}
	if cfg.Search.ContextSize != "" && *searchContextSize == "" {
		*searchContextSize = cfg.Search.ContextSize
	}

	// Apply output config
	if cfg.Output.ReturnImages {
		*returnImages = true
	}
	if cfg.Output.ReturnRelated {
		*returnRelated = true
	}
	if cfg.Output.Stream {
		*stream = true
	}
}

// ExpandEnvVars expands environment variables in configuration values
// Supports ${VAR_NAME} and $VAR_NAME syntax.
func ExpandEnvVars(cfg *ConfigData) {
	// Expand in API config
	cfg.API.Key = expandString(cfg.API.Key)
	cfg.API.BaseURL = expandString(cfg.API.BaseURL)

	// Expand in defaults
	cfg.Defaults.Model = expandString(cfg.Defaults.Model)
	cfg.Defaults.Timeout = expandString(cfg.Defaults.Timeout)

	// Expand in search config
	for i, domain := range cfg.Search.Domains {
		cfg.Search.Domains[i] = expandString(domain)
	}
	cfg.Search.LocationCountry = expandString(cfg.Search.LocationCountry)

	// Expand in output config
	for i, domain := range cfg.Output.ImageDomains {
		cfg.Output.ImageDomains[i] = expandString(domain)
	}
	for i, format := range cfg.Output.ImageFormats {
		cfg.Output.ImageFormats[i] = expandString(format)
	}
}

// expandString expands environment variables in a string.
func expandString(s string) string {
	return os.ExpandEnv(s)
}

// LoadAndMergeConfig loads configuration and merges with CLI flags.
func LoadAndMergeConfig(cmd *cobra.Command, configPath string) (*ConfigData, error) {
	loader := NewLoader()

	// Load config file
	if configPath != "" {
		if err := loader.LoadFrom(configPath); err != nil {
			return nil, err
		}
	} else {
		// Try to load from standard locations (ignore error if not found)
		_ = loader.Load()
	}

	cfg := loader.Data()

	// Expand environment variables
	ExpandEnvVars(cfg)

	// Apply active profile if set
	if cfg.ActiveProfile != "" && cfg.ActiveProfile != "default" {
		pm := NewProfileManager(cfg)
		merged, err := pm.MergeProfile(cfg.ActiveProfile)
		if err == nil {
			cfg = merged
		}
	}

	// Merge with CLI flags
	merger := NewMerger(cfg)
	if err := merger.BindFlags(cmd); err != nil {
		return nil, err
	}
	cfg = merger.MergeWithFlags(cmd)

	return cfg, nil
}
