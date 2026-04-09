package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/sgaunet/pplx/pkg/logger"
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

// MergeWithFlags merges configuration data with CLI flags using Changed() checks.
// Precedence hierarchy: CLI flags > environment variables > config file > defaults.
//
// Changed() pattern: Only override config value if flag was explicitly set by user.
// This critical pattern distinguishes between:
//  1. "flag not provided" (keep config value)
//  2. "flag provided with explicit value, even if zero" (use flag value)
//
// Example with temperature flag (config.yaml sets temperature: 0.7):
//   - User runs: `pplx query "test"`
//     → Changed("temperature") = false, keep config value 0.7
//   - User runs: `pplx query "test" --temperature=0.5`
//     → Changed("temperature") = true, use flag value 0.5
//   - User runs: `pplx query "test" --temperature=0`
//     → Changed("temperature") = true, use flag value 0 (explicit zero is valid)
//
// Without Changed() check, we couldn't distinguish case 1 from case 3, since both
// result in flag variable being 0. Changed() tracks whether user actually provided the flag.
//
// Design note: This function works with merged config (already includes env vars),
// so precedence is: CLI flags > (env vars + config file) merged by viper.
//
//nolint:cyclop,funlen // Function complexity is inherent - checks 29 different CLI flags for explicit user input.
func (m *Merger) MergeWithFlags(cmd *cobra.Command) *ConfigData {
	// Start with config file data as base (already merged with env vars by viper)
	merged := m.data

	// Defaults section: Model parameters and execution settings
	// Pattern: Check if flag was explicitly set (Changed), then override config value
	// This ensures user intent: "I want to override config" vs "I didn't specify, use config"
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

	// Search section: Search behavior and filtering options
	// Same Changed() pattern ensures CLI flags override config only when explicitly provided
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

	// Output section: Response format and presentation options
	// Changed() pattern applies to booleans too - distinguishes "not provided" from "explicitly false"
	// Example: config sets stream: true, user runs without --stream flag → keep true
	//          user runs with --no-stream → Changed=true, set to false
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

// ApplyToGlobals applies merged configuration to global command variables.
// Maintains compatibility with the cobra flag architecture where flags are bound to global variables.
//
// Precedence logic: Config values only apply if the global variable is at its zero value,
// meaning CLI flags (which set globals directly) have already won precedence.
// This ensures the hierarchy: CLI flags > environment variables > config file > defaults.
//
// Exception: Boolean output flags always override when config sets true (true is never zero-valued).
//
// Design rationale: Uses pointer parameters instead of returning a struct to allow selective
// application - caller controls which globals to update. This maintains backward compatibility
// with existing cobra flag binding architecture.
//
func ApplyToGlobals(cfg *ConfigData, opts *GlobalOptions) {
	applyDefaults(cfg, opts)
	applySearchOptions(cfg, opts)
	applyOutputOptions(cfg, opts)
}

// applyDefaults applies default configuration values to GlobalOptions.
// Only applies non-zero config values to avoid overwriting SDK defaults in GlobalOptions
// with zero values from an absent config field. The profile merge bug (zero values ignored)
// is fixed in MergeProfile (using pointer types), not here.
func applyDefaults(cfg *ConfigData, opts *GlobalOptions) {
	if cfg.Defaults.Model != "" {
		opts.Model = cfg.Defaults.Model
	}
	if cfg.Defaults.Temperature != 0 {
		opts.Temperature = cfg.Defaults.Temperature
	}
	if cfg.Defaults.MaxTokens != 0 {
		opts.MaxTokens = cfg.Defaults.MaxTokens
	}
	if cfg.Defaults.TopK != 0 {
		opts.TopK = cfg.Defaults.TopK
	}
	if cfg.Defaults.TopP != 0 {
		opts.TopP = cfg.Defaults.TopP
	}
	if cfg.Defaults.FrequencyPenalty != 0 {
		opts.FrequencyPenalty = cfg.Defaults.FrequencyPenalty
	}
	if cfg.Defaults.PresencePenalty != 0 {
		opts.PresencePenalty = cfg.Defaults.PresencePenalty
	}
	if cfg.Defaults.Timeout != "" {
		if d, err := time.ParseDuration(cfg.Defaults.Timeout); err == nil {
			opts.Timeout = d
		}
	}
}

// applySearchOptions applies search configuration values to GlobalOptions.
// Numeric fields assigned unconditionally (MergeWithFlags handles Changed()).
// String/slice fields: only overwrite when config has a value.
func applySearchOptions(cfg *ConfigData, opts *GlobalOptions) {
	if len(cfg.Search.Domains) > 0 {
		opts.SearchDomains = cfg.Search.Domains
	}
	if cfg.Search.Recency != "" {
		opts.SearchRecency = cfg.Search.Recency
	}
	opts.LocationLat = cfg.Search.LocationLat
	opts.LocationLon = cfg.Search.LocationLon
	if cfg.Search.LocationCountry != "" {
		opts.LocationCountry = cfg.Search.LocationCountry
	}
	if cfg.Search.Mode != "" {
		opts.SearchMode = cfg.Search.Mode
	}
	if cfg.Search.ContextSize != "" {
		opts.SearchContextSize = cfg.Search.ContextSize
	}
}

// applyOutputOptions applies output configuration values to GlobalOptions.
// For booleans: only set to true (enable). This is safe because:
//   - GlobalOptions defaults are false (from NewGlobalOptions)
//   - If config enables a feature (true), we apply it
//   - If profile disables a feature (false via MergeProfile), the merged cfg is false,
//     so we don't touch globalOpts, which stays at its default (false) — correct result
//   - CLI flags override via Changed() in MergeWithFlags, not here
func applyOutputOptions(cfg *ConfigData, opts *GlobalOptions) {
	if cfg.Output.ReturnImages {
		opts.ReturnImages = true
	}
	if cfg.Output.ReturnRelated {
		opts.ReturnRelated = true
	}
	if cfg.Output.Stream {
		opts.Stream = true
	}
}

// ExpandEnvVars expands environment variables in configuration values
// Supports ${VAR_NAME} and $VAR_NAME syntax.
func ExpandEnvVars(cfg *ConfigData) {
	if cfg == nil {
		return
	}

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

// loadConfig loads configuration from an explicit path or auto-discovers it.
func loadConfig(loader *Loader, configPath string) error {
	if configPath != "" {
		// Explicit path: hard error on failure.
		return loader.LoadFrom(configPath)
	}
	// Default locations: warn on parse/syntax errors, ignore "not found".
	if err := loader.Load(); err != nil {
		var configNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &configNotFound) {
			logger.Warn("config file has errors, using defaults",
				"error", err)
		}
	}
	return nil
}

// LoadAndMergeConfig loads configuration and merges with CLI flags.
// If profileOverride is non-empty, it takes precedence over the active_profile in the config file.
func LoadAndMergeConfig(cmd *cobra.Command, configPath, profileOverride string) (*ConfigData, error) {
	loader := NewLoader()

	if err := loadConfig(loader, configPath); err != nil {
		return nil, err
	}

	cfg := loader.Data()

	// Expand environment variables
	ExpandEnvVars(cfg)

	// Determine which profile to apply: CLI flag > config file active_profile.
	activeProfile := cfg.ActiveProfile
	if profileOverride != "" {
		activeProfile = profileOverride
	}

	// Apply profile if set
	if activeProfile != "" && activeProfile != DefaultProfileName {
		pm := NewProfileManager(cfg)
		merged, err := pm.MergeProfile(activeProfile)
		if err != nil {
			if profileOverride != "" {
				// Explicit --profile flag: hard error if profile doesn't exist.
				return nil, fmt.Errorf("failed to apply profile %q: %w", activeProfile, err)
			}
			logger.Warn("failed to apply profile, using base config",
				"profile", activeProfile, "error", err)
		} else {
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
