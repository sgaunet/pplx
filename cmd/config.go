package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sgaunet/pplx/pkg/clerrors"
	"github.com/sgaunet/pplx/pkg/completion"
	"github.com/sgaunet/pplx/pkg/config"
	"github.com/sgaunet/pplx/pkg/logger"
	"github.com/sgaunet/pplx/pkg/security"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	// File permission constants.
	configFilePermission = 0600
	configDirPermission  = 0750

	// configSetArgCount is the number of arguments required by the config set command.
	configSetArgCount = 2

	// enumSuggestMaxDistance is the maximum Levenshtein distance for enum typo suggestions.
	enumSuggestMaxDistance = 2

	// profileDiffMaxArgs is the maximum number of profiles that can be compared in one diff.
	profileDiffMaxArgs = 2
)

var (
	configFilePath    string
	runtimeProfile    string
	jsonOutput        bool
	profileName       string
	optionsSection    string
	optionsFormat     string
	optionsValidation bool
	// Config init flags.
	initTemplate     string
	initWithExamples bool
	initWithProfiles bool
	initForce        bool
	initCheckEnv     bool
	initInteractive  bool
	initDryRun       bool
	initUpdate       bool
	// Config get flags.
	getUnmask   bool
	getJSON     bool
	getProfile  string
	// Config set flags.
	setProfile    string
	setNoValidate bool
	// Config unset flags.
	unsetProfile string
	// Config reset flags.
	resetForce bool
	// Profile create flags.
	createFromTemplate string
	createCopyFrom     string
	createDescription  string
	// Profile delete flags.
	deleteForceFlag bool
	// Profile diff flags.
	profileDiffJSON bool
	// Profile show flags.
	profileShowJSON bool
)

// saveConfigData saves configuration data to a file.
func saveConfigData(data *config.ConfigData) error {
	configPath, err := config.FindConfigFile()
	if err != nil {
		configPath = config.GetDefaultConfigPath()
	}

	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal config for %s: %w", configPath, err)
	}

	if err := os.WriteFile(configPath, yamlData, configFilePermission); err != nil {
		return fmt.Errorf("failed to save config to %s: %w", configPath, err)
	}

	// Verify permissions after writing
	if err := verifyConfigPermissions(configPath); err != nil {
		// Don't fail the save operation, just log the warning
		logger.Warn("config file permissions check failed", "error", err)
	}

	return nil
}

// configCmd represents the config command.
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage pplx configuration",
	Long: `Manage pplx configuration files and profiles.

Configuration files are stored in ~/.config/pplx/
Supported filenames: config.yaml, pplx.yaml, config.yml, pplx.yml

Use subcommands to initialize, view, validate, or edit configuration.`,
}

// configInitCmd initializes a new configuration file.
var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new configuration file",
	Long: `Create a new configuration file at ~/.config/pplx/config.yaml

Examples:
  # Create a minimal config with defaults
  pplx config init

  # Launch interactive wizard
  pplx config init --interactive

  # Create config from a template
  pplx config init --template research

  # Create annotated config with examples
  pplx config init --with-examples

  # Force overwrite existing config
  pplx config init --force

  # Check environment and include API key
  pplx config init --check-env

  # Combine options
  pplx config init --template creative --with-examples --check-env

  # Update an existing config via wizard (keep unchanged values)
  pplx config init --interactive --update

  # Preview what the wizard would produce without writing the file
  pplx config init --interactive --dry-run`,
	RunE: runConfigInit,
}

// loadOrCreateConfig loads configuration based on init flags.
func loadOrCreateConfig() (*config.ConfigData, error) {
	switch {
	case initInteractive:
		return loadOrCreateConfigInteractive()

	case initTemplate != "":
		cfg, err := config.LoadTemplate(initTemplate)
		if err != nil {
			return nil, fmt.Errorf("failed to load template %q: %w", initTemplate, err)
		}
		fmt.Printf("Loaded %q template configuration\n", initTemplate)
		return cfg, nil

	default:
		// Create default minimal config for backward compatibility
		cfg := config.NewConfigData()
		cfg.Defaults.Model = "sonar"
		cfg.Defaults.Temperature = 0.2
		cfg.Defaults.MaxTokens = 4000
		return cfg, nil
	}
}

// loadOrCreateConfigInteractive handles the interactive wizard path, including --update mode.
func loadOrCreateConfigInteractive() (*config.ConfigData, error) {
	if initUpdate {
		existing, err := loadExistingConfigForUpdate()
		if err != nil {
			return nil, err
		}
		wizard := NewWizardStateWithExisting(existing)
		cfg, err := wizard.Run()
		if err != nil {
			return nil, fmt.Errorf("wizard failed: %w", err)
		}
		return cfg, nil
	}

	wizard := NewWizardState()
	cfg, err := wizard.Run()
	if err != nil {
		return nil, fmt.Errorf("wizard failed: %w", err)
	}
	return cfg, nil
}

// loadExistingConfigForUpdate loads the existing config file to seed the --update wizard.
func loadExistingConfigForUpdate() (*config.ConfigData, error) {
	configPath := config.GetDefaultConfigPath()
	if configFilePath != "" {
		configPath = configFilePath
	}
	loader := config.NewLoader()
	if err := loader.LoadFrom(configPath); err != nil {
		return nil, fmt.Errorf("--update requires an existing config file, but loading %s failed: %w", configPath, err)
	}
	return loader.Data(), nil
}

// generateYAMLContent generates YAML content from config.
func generateYAMLContent(cfg *config.ConfigData) (string, error) {
	if initWithExamples || initWithProfiles || initTemplate != "" || initInteractive {
		opts := config.DefaultAnnotationOptions()
		opts.IncludeExamples = initWithExamples

		annotated, err := config.GenerateAnnotatedConfig(cfg, opts)
		if err != nil {
			return "", fmt.Errorf("failed to generate annotated config: %w", err)
		}
		if !initInteractive {
			fmt.Println("Generated annotated configuration with descriptions")
		}
		return annotated, nil
	}

	// Generate minimal YAML for backward compatibility
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}
	return string(data), nil
}

// runConfigInit implements the config init command logic.
func runConfigInit(_ *cobra.Command, _ []string) error {
	configPath := resolveInitConfigPath()

	// --dry-run skips all filesystem checks and just prints the generated YAML.
	if initDryRun {
		return runConfigInitDryRun()
	}

	if err := checkExistingConfigFile(configPath); err != nil {
		return err
	}

	if err := ensureConfigDir(configPath); err != nil {
		return err
	}

	if initCheckEnv {
		checkEnvironment()
	}

	return writeInitConfig(configPath)
}

// resolveInitConfigPath returns the config path from flag or default.
func resolveInitConfigPath() string {
	if configFilePath != "" {
		return configFilePath
	}
	return config.GetDefaultConfigPath()
}

// checkExistingConfigFile validates overwrite/update flags when the file already exists.
func checkExistingConfigFile(configPath string) error {
	if _, err := os.Stat(configPath); err != nil {
		// File does not exist — nothing to check.
		return nil //nolint:nilerr // os.Stat error means file not found; intentionally ignored.
	}
	if !initForce && !initUpdate {
		return fmt.Errorf("%w at %s (use --force to overwrite or --update to modify)", clerrors.ErrConfigFileExists, configPath)
	}
	if !initUpdate {
		fmt.Printf("Overwriting existing configuration at %s\n", configPath)
	}
	return nil
}

// ensureConfigDir creates the config directory if it does not exist.
func ensureConfigDir(configPath string) error {
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, configDirPermission); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", configDir, err)
	}
	return nil
}

// writeInitConfig loads/creates the config, generates YAML and writes it to disk.
func writeInitConfig(configPath string) error {
	cfg, err := loadOrCreateConfig()
	if err != nil {
		return err
	}

	yamlContent, err := generateYAMLContent(cfg)
	if err != nil {
		return err
	}

	if err := os.WriteFile(configPath, []byte(yamlContent), configFilePermission); err != nil {
		return fmt.Errorf("failed to write config file to %s: %w", configPath, err)
	}

	fmt.Printf("Configuration file created at %s\n", configPath)

	if err := verifyConfigPermissions(configPath); err != nil {
		logger.Warn("config file permissions check failed", "error", err)
	}

	return nil
}

// runConfigInitDryRun handles the --dry-run path: generate config and print to stdout
// without writing any files. Works with --interactive (wizard) and non-interactive modes.
func runConfigInitDryRun() error {
	// Check environment if requested (informational only in dry-run)
	if initCheckEnv {
		checkEnvironment()
	}

	cfg, err := loadOrCreateConfig()
	if err != nil {
		return err
	}

	yamlContent, err := generateYAMLContent(cfg)
	if err != nil {
		return err
	}

	fmt.Println("--- dry-run: generated configuration (not written to disk) ---")
	fmt.Print(yamlContent)
	fmt.Println("--- end dry-run ---")
	return nil
}

// checkEnvironment checks for API keys and other environment variables.
func checkEnvironment() {
	fmt.Println("\nChecking environment variables...")

	apiKey := os.Getenv("PERPLEXITY_API_KEY")
	if apiKey != "" {
		fmt.Printf("✓ PERPLEXITY_API_KEY found (length: %d)\n", len(apiKey))
	} else {
		fmt.Println("✗ PERPLEXITY_API_KEY not found")
		fmt.Println("  Set it with: export PERPLEXITY_API_KEY=your-api-key")
	}

	// Check for other optional env vars
	baseURL := os.Getenv("PERPLEXITY_BASE_URL")
	if baseURL != "" {
		fmt.Printf("✓ PERPLEXITY_BASE_URL found: %s\n", baseURL)
	}

	fmt.Println()
}

// verifyConfigPermissions checks file permissions and warns if they are too permissive.
// It returns an error only if the file cannot be accessed.
func verifyConfigPermissions(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to check file permissions for %s: %w", path, err)
	}

	mode := info.Mode().Perm()
	// Check if group or other have any permissions (0077 = group/other rwx bits)
	if mode&0077 != 0 {
		logger.Warn("config file has insecure permissions",
			"path", path,
			"current_permissions", fmt.Sprintf("%#o", mode),
			"recommended_permissions", "0600",
			"fix_command", "chmod 600 "+path,
			"reason", "file may contain API keys")
	}

	return nil
}

// configShowCmd displays the current configuration.
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  `Display the current configuration, either from file or defaults.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		loader := config.NewLoader()

		if configFilePath != "" {
			if err := loader.LoadFrom(configFilePath); err != nil {
				return fmt.Errorf("failed to load config from %s: %w", configFilePath, err)
			}
		} else {
			if err := loader.Load(); err != nil {
				return fmt.Errorf("failed to load config from default locations: %w", err)
			}
		}

		cfg := loader.Data()

		// If profile specified, show only that profile
		if profileName != "" {
			pm := config.NewProfileManager(cfg)
			profile, err := pm.LoadProfile(profileName)
			if err != nil {
				return fmt.Errorf("failed to load profile %q: %w", profileName, err)
			}

			if jsonOutput {
				data, err := json.MarshalIndent(profile, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal profile %q to JSON: %w", profileName, err)
				}
				fmt.Println(string(data))
			} else {
				data, err := yaml.Marshal(profile)
				if err != nil {
					return fmt.Errorf("failed to marshal profile %q to YAML: %w", profileName, err)
				}
				fmt.Print(string(data))
			}
			return nil
		}

		// Show full config - MASK API KEY
		cfgCopy := *cfg

		// Mask API key if present
		if cfgCopy.API.Key != "" {
			cfgCopy.API.Key = security.MaskAPIKey(cfgCopy.API.Key)
		}

		if jsonOutput {
			data, err := json.MarshalIndent(cfgCopy, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal config to JSON: %w", err)
			}
			fmt.Println(string(data))
		} else {
			data, err := yaml.Marshal(cfgCopy)
			if err != nil {
				return fmt.Errorf("failed to marshal config to YAML: %w", err)
			}
			fmt.Print(string(data))
		}

		return nil
	},
}

// configValidateCmd validates the configuration file.
var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration file",
	Long:  `Check the configuration file for syntax errors and invalid values.`,
	RunE: func(_ *cobra.Command, _ []string) error {
		loader := config.NewLoader()

		if configFilePath != "" {
			if err := loader.LoadFrom(configFilePath); err != nil {
				return fmt.Errorf("failed to load config from %s: %w", configFilePath, err)
			}
		} else {
			if err := loader.Load(); err != nil {
				return fmt.Errorf("failed to load config from default locations: %w", err)
			}
		}

		cfg := loader.Data()

		// Validate configuration
		validator := config.NewValidator()
		if err := validator.Validate(cfg); err != nil {
			fmt.Println("Configuration validation failed:")
			fmt.Println(err.Error())
			return clerrors.ErrValidationFailed
		}

		fmt.Println("Configuration is valid ✓")
		return nil
	},
}

// configEditCmd opens the configuration file in an editor.
var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit configuration file",
	Long:  `Open the configuration file in your default editor ($EDITOR).`,
	RunE: func(_ *cobra.Command, _ []string) error {
		// Find config file
		configPath := configFilePath
		if configPath == "" {
			var err error
			configPath, err = config.FindConfigFile()
			if err != nil {
				// If no config exists, use default path
				configPath = config.GetDefaultConfigPath()
			}
		}

		// Get editor from environment
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi" // Fallback to vi
		}

		// Open editor
		editorCmd := exec.Command(editor, configPath) //nolint:gosec // Editor command from $EDITOR is intentional.
		editorCmd.Stdin = os.Stdin
		editorCmd.Stdout = os.Stdout
		editorCmd.Stderr = os.Stderr

		if err := editorCmd.Run(); err != nil {
			return fmt.Errorf("failed to open editor %s: %w", editor, err)
		}

		// Validate after editing
		loader := config.NewLoader()
		if err := loader.LoadFrom(configPath); err != nil {
			logger.Warn("configuration file has syntax errors", "error", err)
			return nil
		}

		validator := config.NewValidator()
		if err := validator.Validate(loader.Data()); err != nil {
			logger.Warn("configuration validation failed", "error", err)
			return nil
		}

		fmt.Println("Configuration updated and validated successfully ✓")
		return nil
	},
}

// configOptionsCmd lists all available configuration options.
var configOptionsCmd = &cobra.Command{
	Use:   "options",
	Short: "List all configuration options",
	Long: `Display all available configuration options with their metadata.

Options can be filtered by section (defaults, search, output, api) and
formatted as a table (default), JSON, or YAML.

Examples:
  # List all options in table format
  pplx config options

  # List options in JSON format
  pplx config options --format json

  # List only search options
  pplx config options --section search

  # Show validation rules
  pplx config options --validation

  # Combine filters
  pplx config options --section defaults --format yaml --validation`,
	RunE: runConfigOptions,
}

// runConfigOptions implements the config options command.
func runConfigOptions(_ *cobra.Command, _ []string) error {
	// Create metadata registry
	registry := config.NewMetadataRegistry()

	// Get options (filtered by section if specified)
	var options []*config.OptionMetadata
	if optionsSection != "" {
		options = registry.GetBySection(optionsSection)
		if len(options) == 0 {
			return fmt.Errorf("%w: %s (valid: defaults, search, output, api)", clerrors.ErrUnknownSection, optionsSection)
		}
	} else {
		options = registry.GetAll()
	}

	// If validation flag not set, clear validation rules to reduce output
	if !optionsValidation {
		for _, opt := range options {
			opt.ValidationRules = nil
		}
	}

	// Format output
	output, err := config.FormatOptions(options, optionsFormat)
	if err != nil {
		return fmt.Errorf("failed to format options as %s: %w", optionsFormat, err)
	}

	fmt.Print(output)
	return nil
}

// configPathCmd shows the active config file location and search order.
var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show configuration file location and search paths",
	Long: `Display the active configuration file path and the full search order.

The search order shows all locations where pplx looks for configuration files,
with status indicators for each location.

Examples:
  # Show active config and search paths
  pplx config path

  # Validate configuration and show details
  pplx config path --check`,
	RunE: runConfigPath,
}

var pathCheckFlag bool

// runConfigPath implements the config path command logic.
func runConfigPath(_ *cobra.Command, _ []string) error {
	// Find active config file
	activeConfig, err := config.FindConfigFile()
	hasConfig := err == nil

	fmt.Println("Configuration File Search Order:")
	fmt.Println()

	// Get possible filenames
	filenames := []string{"config.yaml", "pplx.yaml", "config.yml", "pplx.yml"}

	// Iterate through search paths
	for _, basePath := range config.ConfigPaths {
		expandedPath := os.ExpandEnv(basePath)
		fmt.Printf("📁 %s\n", expandedPath)

		for _, filename := range filenames {
			fullPath := filepath.Join(expandedPath, filename)
			status := getPathStatus(fullPath, activeConfig)
			fmt.Printf("   %s %s\n", status, filename)
		}
		fmt.Println()
	}

	// Show active configuration summary
	if hasConfig {
		fmt.Printf("✓ Active configuration: %s\n", activeConfig)
	} else {
		fmt.Println("✗ No configuration file found")
		fmt.Println()
		fmt.Println("Create one with: pplx config init")
		return nil
	}

	// If --check flag is set, validate the configuration
	if pathCheckFlag {
		fmt.Println()
		fmt.Println("Configuration Validation:")
		fmt.Println()

		loader := config.NewLoader()
		if err := loader.LoadFrom(activeConfig); err != nil {
			fmt.Printf("✗ Failed to load config: %v\n", err)
			return nil
		}

		cfg := loader.Data()

		// Count profiles
		profileCount := len(cfg.Profiles)
		fmt.Printf("  Profiles: %d\n", profileCount)
		if profileCount > 0 {
			fmt.Printf("  Active:   %s\n", cfg.ActiveProfile)
		}

		// Validate configuration
		validator := config.NewValidator()
		if err := validator.Validate(cfg); err != nil {
			fmt.Printf("  Status:   ✗ INVALID\n")
			fmt.Println()
			fmt.Println("Validation errors:")
			fmt.Printf("  %v\n", err)
		} else {
			fmt.Printf("  Status:   ✓ VALID\n")
		}
	}

	return nil
}

// getPathStatus returns a status indicator for a config file path.
func getPathStatus(path, activeConfig string) string {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return "⚪" // not found
	}

	if info.IsDir() {
		return "⚠️ " // is directory (error)
	}

	if path == activeConfig {
		return "✓ " // active config
	}

	return "○ " // exists but not used
}

// configProfileCmd manages profiles.
var configProfileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage configuration profiles",
	Long:  `Create, list, switch, and delete configuration profiles.`,
}

// configProfileListCmd lists all profiles.
var configProfileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all profiles",
	RunE: func(_ *cobra.Command, _ []string) error {
		loader := config.NewLoader()
		if err := loader.Load(); err != nil {
			return fmt.Errorf("failed to load config for profile list: %w", err)
		}

		pm := config.NewProfileManager(loader.Data())
		profiles := pm.ListProfiles()

		activeProfile := pm.GetActiveProfileName()

		fmt.Println("Available profiles:")
		for _, name := range profiles {
			if name == activeProfile {
				fmt.Printf("  * %s (active)\n", name)
			} else {
				fmt.Printf("    %s\n", name)
			}
		}

		return nil
	},
}

// configProfileCreateCmd creates a new profile.
var configProfileCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new profile",
	Long: `Create a new configuration profile.

Profiles can be created empty, from a built-in template, or copied from an
existing profile. An optional description can be attached.

Examples:
  pplx config profile create myprofile
  pplx config profile create myprofile --description "My research setup"
  pplx config profile create myprofile --from-template research
  pplx config profile create myprofile --copy-from existing`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		name := args[0]

		loader := config.NewLoader()
		if err := loader.Load(); err != nil {
			return fmt.Errorf("failed to load config for profile creation: %w", err)
		}

		pm := config.NewProfileManager(loader.Data())

		switch {
		case createCopyFrom != "":
			if _, err := pm.CloneProfile(name, createCopyFrom); err != nil {
				return fmt.Errorf("failed to copy profile %q to %q: %w", createCopyFrom, name, err)
			}
			// Apply description override after clone.
			if createDescription != "" {
				profile, loadErr := pm.LoadProfile(name)
				if loadErr != nil {
					return fmt.Errorf("failed to load cloned profile %q: %w", name, loadErr)
				}
				profile.Description = createDescription
			}

		case createFromTemplate != "":
			tplCfg, err := config.LoadTemplate(createFromTemplate)
			if err != nil {
				return fmt.Errorf("failed to load template %q: %w", createFromTemplate, err)
			}
			profile, err := pm.CreateProfile(name, createDescription)
			if err != nil {
				return fmt.Errorf("failed to create profile %q: %w", name, err)
			}
			// Populate profile from template config defaults/search/output.
			profile.Defaults = buildProfileDefaultsFromConfig(tplCfg)
			profile.Search = buildProfileSearchFromConfig(tplCfg)
			profile.Output = buildProfileOutputFromConfig(tplCfg)

		default:
			if _, err := pm.CreateProfile(name, createDescription); err != nil {
				return fmt.Errorf("failed to create profile %q: %w", name, err)
			}
		}

		if err := saveConfigData(loader.Data()); err != nil {
			return err
		}

		fmt.Printf("Profile '%s' created successfully\n", name)
		return nil
	},
}

// buildProfileDefaultsFromConfig converts DefaultsConfig fields to ProfileDefaults pointers.
// Only non-zero values are converted to avoid overriding base config with zero values.
func buildProfileDefaultsFromConfig(cfg *config.ConfigData) config.ProfileDefaults {
	pd := config.ProfileDefaults{}
	if cfg.Defaults.Model != "" {
		v := cfg.Defaults.Model
		pd.Model = &v
	}
	if cfg.Defaults.Temperature != 0 {
		v := cfg.Defaults.Temperature
		pd.Temperature = &v
	}
	if cfg.Defaults.MaxTokens != 0 {
		v := cfg.Defaults.MaxTokens
		pd.MaxTokens = &v
	}
	if cfg.Defaults.TopK != 0 {
		v := cfg.Defaults.TopK
		pd.TopK = &v
	}
	if cfg.Defaults.TopP != 0 {
		v := cfg.Defaults.TopP
		pd.TopP = &v
	}
	if cfg.Defaults.FrequencyPenalty != 0 {
		v := cfg.Defaults.FrequencyPenalty
		pd.FrequencyPenalty = &v
	}
	if cfg.Defaults.PresencePenalty != 0 {
		v := cfg.Defaults.PresencePenalty
		pd.PresencePenalty = &v
	}
	if cfg.Defaults.Timeout != "" {
		v := cfg.Defaults.Timeout
		pd.Timeout = &v
	}
	return pd
}

// buildProfileSearchFromConfig converts SearchConfig fields to ProfileSearch pointers.
func buildProfileSearchFromConfig(cfg *config.ConfigData) config.ProfileSearch {
	ps := config.ProfileSearch{}
	if len(cfg.Search.Domains) > 0 {
		v := make([]string, len(cfg.Search.Domains))
		copy(v, cfg.Search.Domains)
		ps.Domains = &v
	}
	if cfg.Search.Recency != "" {
		v := cfg.Search.Recency
		ps.Recency = &v
	}
	if cfg.Search.Mode != "" {
		v := cfg.Search.Mode
		ps.Mode = &v
	}
	if cfg.Search.ContextSize != "" {
		v := cfg.Search.ContextSize
		ps.ContextSize = &v
	}
	buildProfileSearchLocation(&ps, cfg)
	buildProfileSearchDates(&ps, cfg)
	return ps
}

// buildProfileSearchLocation sets location-related fields on ps from cfg.
func buildProfileSearchLocation(ps *config.ProfileSearch, cfg *config.ConfigData) {
	if cfg.Search.LocationLat != 0 {
		v := cfg.Search.LocationLat
		ps.LocationLat = &v
	}
	if cfg.Search.LocationLon != 0 {
		v := cfg.Search.LocationLon
		ps.LocationLon = &v
	}
	if cfg.Search.LocationCountry != "" {
		v := cfg.Search.LocationCountry
		ps.LocationCountry = &v
	}
}

// buildProfileSearchDates sets date-filtering fields on ps from cfg.
func buildProfileSearchDates(ps *config.ProfileSearch, cfg *config.ConfigData) {
	if cfg.Search.AfterDate != "" {
		v := cfg.Search.AfterDate
		ps.AfterDate = &v
	}
	if cfg.Search.BeforeDate != "" {
		v := cfg.Search.BeforeDate
		ps.BeforeDate = &v
	}
	if cfg.Search.LastUpdatedAfter != "" {
		v := cfg.Search.LastUpdatedAfter
		ps.LastUpdatedAfter = &v
	}
	if cfg.Search.LastUpdatedBefore != "" {
		v := cfg.Search.LastUpdatedBefore
		ps.LastUpdatedBefore = &v
	}
}

// buildProfileOutputFromConfig converts OutputConfig fields to ProfileOutput pointers.
// Boolean flags are only promoted to pointers when they are true, matching the
// "not set = nil" convention used throughout the profile system.
func buildProfileOutputFromConfig(cfg *config.ConfigData) config.ProfileOutput {
	po := config.ProfileOutput{}
	if cfg.Output.Stream {
		v := cfg.Output.Stream
		po.Stream = &v
	}
	if cfg.Output.ReturnImages {
		v := cfg.Output.ReturnImages
		po.ReturnImages = &v
	}
	if cfg.Output.ReturnRelated {
		v := cfg.Output.ReturnRelated
		po.ReturnRelated = &v
	}
	if cfg.Output.JSON {
		v := cfg.Output.JSON
		po.JSON = &v
	}
	if len(cfg.Output.ImageDomains) > 0 {
		v := make([]string, len(cfg.Output.ImageDomains))
		copy(v, cfg.Output.ImageDomains)
		po.ImageDomains = &v
	}
	if len(cfg.Output.ImageFormats) > 0 {
		v := make([]string, len(cfg.Output.ImageFormats))
		copy(v, cfg.Output.ImageFormats)
		po.ImageFormats = &v
	}
	if cfg.Output.ResponseFormatJSONSchema != "" {
		v := cfg.Output.ResponseFormatJSONSchema
		po.ResponseFormatJSONSchema = &v
	}
	if cfg.Output.ResponseFormatRegex != "" {
		v := cfg.Output.ResponseFormatRegex
		po.ResponseFormatRegex = &v
	}
	if cfg.Output.ReasoningEffort != "" {
		v := cfg.Output.ReasoningEffort
		po.ReasoningEffort = &v
	}
	return po
}

// configProfileSwitchCmd switches the active profile.
var configProfileSwitchCmd = &cobra.Command{
	Use:   "switch [name]",
	Short: "Switch to a different profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		profileName := args[0]

		loader := config.NewLoader()
		if err := loader.Load(); err != nil {
			return fmt.Errorf("failed to load config for profile switch: %w", err)
		}

		pm := config.NewProfileManager(loader.Data())
		if err := pm.SetActiveProfile(profileName); err != nil {
			return fmt.Errorf("failed to switch to profile %q: %w", profileName, err)
		}

		// Save config
		if err := saveConfigData(loader.Data()); err != nil {
			return err
		}

		fmt.Printf("Switched to profile '%s'\n", profileName)
		return nil
	},
}

// configProfileDeleteCmd deletes a profile.
var configProfileDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		name := args[0]

		loader := config.NewLoader()
		if err := loader.Load(); err != nil {
			return fmt.Errorf("failed to load config for profile deletion: %w", err)
		}

		data := loader.Data()
		pm := config.NewProfileManager(data)

		profile, err := pm.LoadProfile(name)
		if err != nil {
			return fmt.Errorf("failed to load profile %q: %w", name, err)
		}

		isActive := data.ActiveProfile == name

		if !deleteForceFlag {
			// Print a brief summary so the user knows what they are deleting.
			fmt.Printf("Profile: %s\n", profile.Name)
			if profile.Description != "" {
				fmt.Printf("Description: %s\n", profile.Description)
			}
			if isActive {
				fmt.Println("Warning: this is the active profile. Deleting it will switch to 'default'.")
			}
			fmt.Print("Confirm (y/N): ")

			scanner := bufio.NewScanner(os.Stdin)
			answer := ""
			if scanner.Scan() {
				answer = strings.TrimSpace(scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				return fmt.Errorf("failed to read confirmation: %w", err)
			}

			if !strings.EqualFold(answer, "y") && !strings.EqualFold(answer, "yes") {
				fmt.Println("Aborted.")
				return nil
			}
		}

		if err := pm.DeleteProfile(name); err != nil {
			return fmt.Errorf("failed to delete profile %q: %w", name, err)
		}

		if err := saveConfigData(data); err != nil {
			return err
		}

		if isActive {
			fmt.Printf("Active profile deleted; switched to '%s'.\n", config.DefaultProfileName)
		}
		fmt.Printf("Profile '%s' deleted successfully\n", name)
		return nil
	},
}

// configProfileDiffCmd compares two profiles or a profile against the base config.
var configProfileDiffCmd = &cobra.Command{
	Use:   "diff <profile1> [profile2]",
	Short: "Compare two profiles or a profile against base config",
	Long: `Show the differences between two profiles, or between a profile and the base config.

When one argument is given, the profile is compared against the base configuration.
When two arguments are given, the two profiles are compared against each other.

Examples:
  pplx config profile diff research
  pplx config profile diff research creative
  pplx config profile diff research creative --json`,
	Args: cobra.RangeArgs(1, profileDiffMaxArgs),
	RunE: func(_ *cobra.Command, args []string) error {
		data, err := loadConfigData(configFilePath)
		if err != nil {
			return err
		}

		pm := config.NewProfileManager(data)

		var left, right *config.ConfigData

		if len(args) == profileDiffMaxArgs {
			// Compare two profiles: merge each with base, then diff.
			left, err = pm.MergeProfile(args[0])
			if err != nil {
				return fmt.Errorf("failed to merge profile %q: %w", args[0], err)
			}
			right, err = pm.MergeProfile(args[1])
			if err != nil {
				return fmt.Errorf("failed to merge profile %q: %w", args[1], err)
			}
		} else {
			// Compare one profile against the base config.
			left = data
			right, err = pm.MergeProfile(args[0])
			if err != nil {
				return fmt.Errorf("failed to merge profile %q: %w", args[0], err)
			}
		}

		entries := config.DiffConfigs(left, right)

		format := "table"
		if profileDiffJSON {
			format = "json"
		}
		fmt.Println(config.FormatDiff(entries, format))
		return nil
	},
}

// configProfileShowCmd displays all settings in a profile.
var configProfileShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show all settings in a profile",
	Long: `Display the full configuration for a named profile.

Examples:
  pplx config profile show research
  pplx config profile show research --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		name := args[0]

		data, err := loadConfigData(configFilePath)
		if err != nil {
			return err
		}

		pm := config.NewProfileManager(data)
		profile, err := pm.LoadProfile(name)
		if err != nil {
			return fmt.Errorf("failed to load profile %q: %w", name, err)
		}

		if profileShowJSON {
			out, marshalErr := json.MarshalIndent(profile, "", "  ")
			if marshalErr != nil {
				return fmt.Errorf("failed to marshal profile %q to JSON: %w", name, marshalErr)
			}
			fmt.Println(string(out))
			return nil
		}

		out, marshalErr := yaml.Marshal(profile)
		if marshalErr != nil {
			return fmt.Errorf("failed to marshal profile %q to YAML: %w", name, marshalErr)
		}
		fmt.Print(string(out))
		return nil
	},
}

// loadConfigData loads configuration from a specific path or the default location.
// When path is empty the standard search order is used.
func loadConfigData(path string) (*config.ConfigData, error) {
	loader := config.NewLoader()
	if path != "" {
		if err := loader.LoadFrom(path); err != nil {
			return nil, fmt.Errorf("failed to load config from %s: %w", path, err)
		}
	} else {
		if err := loader.Load(); err != nil {
			return nil, fmt.Errorf("failed to load config from default locations: %w", err)
		}
	}
	return loader.Data(), nil
}

// enumSuggestionsForKey returns the slice of valid enum values for keys that
// accept a constrained set of values, or nil for keys without enum constraints.
func enumSuggestionsForKey(key string) []string {
	switch key {
	case "search.recency":
		return config.GetValidSearchRecencyValues()
	case "search.mode":
		return config.GetValidSearchModeValues()
	case "search.context_size":
		return config.GetValidContextSizeValues()
	case "output.reasoning_effort":
		return config.GetValidReasoningEffortValues()
	default:
		return nil
	}
}

// configGetCmd retrieves a single configuration value by dot-notation key.
var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Long: `Print the current value of a configuration option.

Keys use dot-notation: <section>.<name> (e.g. api.key, defaults.model).
A section-only key (e.g. "defaults") prints all options in that section as YAML.

Examples:
  pplx config get defaults.model
  pplx config get api.key
  pplx config get api.key --unmask
  pplx config get search --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		key := args[0]

		cfg, err := loadConfigData(configFilePath)
		if err != nil {
			return err
		}

		// If a specific profile is requested, apply its overrides.
		if getProfile != "" {
			pm := config.NewProfileManager(cfg)
			merged, mergeErr := pm.MergeProfile(getProfile)
			if mergeErr != nil {
				return fmt.Errorf("failed to load profile %q: %w", getProfile, mergeErr)
			}
			cfg = merged
		}

		val, err := config.GetValue(cfg, key)
		if err != nil {
			return clerrors.NewConfigError(fmt.Sprintf("unknown key %q", key), err)
		}

		// Mask the API key unless the caller explicitly opted out.
		if key == "api.key" && !getUnmask {
			if strVal, ok := val.(string); ok {
				val = security.MaskAPIKey(strVal)
			}
		}

		if getJSON {
			data, jsonErr := json.MarshalIndent(val, "", "  ")
			if jsonErr != nil {
				return fmt.Errorf("failed to marshal value to JSON: %w", jsonErr)
			}
			fmt.Println(string(data))
			return nil
		}

		// Section-level queries return map[string]any — render as YAML for readability.
		if m, ok := val.(map[string]any); ok {
			data, yamlErr := yaml.Marshal(m)
			if yamlErr != nil {
				return fmt.Errorf("failed to marshal section to YAML: %w", yamlErr)
			}
			fmt.Print(string(data))
			return nil
		}

		fmt.Printf("%v\n", val)
		return nil
	},
}

// configSetCmd sets a configuration value and persists it to disk.
var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration option and save it to the config file.

The value is validated unless --no-validate is provided. For enum fields a
"Did you mean?" suggestion is shown when the value is close to a valid one.

Examples:
  pplx config set defaults.model sonar-pro
  pplx config set defaults.temperature 0.7
  pplx config set search.recency week
  pplx config set api.key pplx-my-secret-key`,
	Args: cobra.ExactArgs(configSetArgCount),
	RunE: func(_ *cobra.Command, args []string) error {
		key, value := args[0], args[1]

		cfg, err := loadConfigData(configFilePath)
		if err != nil {
			// No config file yet — start with an empty one.
			cfg = config.NewConfigData()
		}

		// Profile-scoped set is not yet implemented at the SetValue level
		// (profiles use pointer-based types). Surface a clear error.
		if setProfile != "" {
			return clerrors.NewConfigError(
				"setting values in a specific profile via --profile is not yet supported; "+
					"edit the profile section directly with: pplx config edit",
				nil,
			)
		}

		if err := config.SetValue(cfg, key, value); err != nil {
			return clerrors.NewConfigError(fmt.Sprintf("cannot set %q", key), err)
		}

		if !setNoValidate {
			validator := config.NewValidator()
			if valErr := validator.Validate(cfg); valErr != nil {
				// For enum-constrained fields, offer a typo suggestion.
				if valid := enumSuggestionsForKey(key); valid != nil {
					if suggestion := config.SuggestEnum(value, valid, enumSuggestMaxDistance); suggestion != "" {
						return fmt.Errorf("%w\nDid you mean: %q?", valErr, suggestion)
					}
				}
				return fmt.Errorf("validation failed after setting %s: %w", key, valErr)
			}
		}

		// Ensure the config directory and file exist before saving.
		configPath, findErr := config.FindConfigFile()
		if findErr != nil {
			// No file found — create one at the default path.
			configPath = config.GetDefaultConfigPath()
			if err := os.MkdirAll(filepath.Dir(configPath), configDirPermission); err != nil {
				return fmt.Errorf("failed to create config directory: %w", err)
			}
			if err := os.WriteFile(configPath, []byte{}, configFilePermission); err != nil {
				return fmt.Errorf("failed to create config file at %s: %w", configPath, err)
			}
		}
		_ = configPath // saveConfigData resolves the path itself via FindConfigFile.

		if err := saveConfigData(cfg); err != nil {
			return err
		}

		fmt.Printf("Set %s = %s\n", key, value)
		return nil
	},
}

// configUnsetCmd removes a value from the configuration by resetting it to the zero value.
var configUnsetCmd = &cobra.Command{
	Use:   "unset <key>",
	Short: "Remove a configuration value",
	Long: `Reset a configuration option to its zero value (empty string, 0, false, etc.).

Required fields (such as api.key) cannot be unset.

Examples:
  pplx config unset defaults.model
  pplx config unset search.recency`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		key := args[0]

		// Refuse to unset required fields.
		reg := config.NewMetadataRegistry()
		if meta, err := reg.GetOption(key); err == nil && meta.Required {
			return clerrors.NewConfigError(
				fmt.Sprintf("cannot unset required field %q", key), nil)
		}

		cfg, err := loadConfigData(configFilePath)
		if err != nil {
			return err
		}

		// Profile-scoped unset follows the same limitation as set.
		if unsetProfile != "" {
			return clerrors.NewConfigError(
				"unsetting values in a specific profile via --profile is not yet supported",
				nil,
			)
		}

		// Capture the previous value for display before clearing it.
		prev, getErr := config.GetValue(cfg, key)
		if getErr != nil {
			return clerrors.NewConfigError(fmt.Sprintf("unknown key %q", key), getErr)
		}

		// Mask the API key in display output.
		displayPrev := fmt.Sprintf("%v", prev)
		if key == "api.key" {
			if strVal, ok := prev.(string); ok {
				displayPrev = security.MaskAPIKey(strVal)
			}
		}

		if err := config.SetValue(cfg, key, ""); err != nil {
			return clerrors.NewConfigError(fmt.Sprintf("cannot unset %q", key), err)
		}

		if err := saveConfigData(cfg); err != nil {
			return err
		}

		fmt.Printf("Unset %s (was: %s)\n", key, displayPrev)
		return nil
	},
}

// configResetCmd resets one or all configuration values to their defaults.
var configResetCmd = &cobra.Command{
	Use:   "reset [key]",
	Short: "Reset configuration to defaults",
	Long: `Reset a specific key or the entire configuration to default values.

When a key is provided only that option is reset. Without a key the entire
configuration is reset — a confirmation prompt is shown unless --force is used.

Examples:
  pplx config reset defaults.model
  pplx config reset
  pplx config reset --force`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfigData(configFilePath)
		if err != nil {
			// If there is no file yet start from an empty config.
			cfg = config.NewConfigData()
		}

		if len(args) == 1 {
			key := args[0]

			reg := config.NewMetadataRegistry()
			meta, metaErr := reg.GetOption(key)
			if metaErr != nil {
				return clerrors.NewConfigError(fmt.Sprintf("unknown key %q", key), metaErr)
			}

			// Determine the string representation of the default value.
			defaultStr := ""
			if meta.Default != nil {
				defaultStr = fmt.Sprintf("%v", meta.Default)
			}

			if setErr := config.SetValue(cfg, key, defaultStr); setErr != nil {
				return clerrors.NewConfigError(
					fmt.Sprintf("cannot reset %q to default %q", key, defaultStr), setErr)
			}

			if saveErr := saveConfigData(cfg); saveErr != nil {
				return saveErr
			}

			fmt.Printf("Reset %s to default (%s)\n", key, defaultStr)
			return nil
		}

		// Full reset.
		if !resetForce {
			fmt.Print("This will reset ALL configuration values to defaults. Continue? [y/N] ")
			var answer string
			if _, scanErr := fmt.Fscan(cmd.InOrStdin(), &answer); scanErr != nil {
				return fmt.Errorf("failed to read confirmation: %w", scanErr)
			}
			if answer != "y" && answer != "Y" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		fresh := config.NewConfigData()
		if saveErr := saveConfigData(fresh); saveErr != nil {
			return saveErr
		}

		fmt.Println("Configuration reset to defaults.")
		return nil
	},
}

// export flags.
var exportWithoutSecrets bool

// configExportCmd exports the current configuration to stdout.
var configExportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export configuration to stdout",
	Long: `Print the current configuration as YAML to stdout.

Use --without-secrets to omit sensitive values such as the API key before exporting.

Examples:
  pplx config export
  pplx config export --without-secrets
  pplx config export > backup.yaml`,
	RunE: func(_ *cobra.Command, _ []string) error {
		cfg, err := loadConfigData(configFilePath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if exportWithoutSecrets {
			cfg.API.Key = ""
		}

		out, err := yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		fmt.Print(string(out))
		return nil
	},
}

// configImportCmd imports configuration from a file.
var configImportCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Import configuration from file",
	Long: `Replace the active configuration with the contents of the given YAML file.

The file is validated before writing. On success the previous configuration is
overwritten at the default config location.

Examples:
  pplx config import backup.yaml
  pplx config import ~/exports/pplx-config.yaml`,
	Args: cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		importPath := args[0]

		data, err := os.ReadFile(importPath) // #nosec G304
		if err != nil {
			return fmt.Errorf("failed to read import file %s: %w", importPath, err)
		}

		var cfg config.ConfigData
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("failed to parse import file %s: %w", importPath, err)
		}

		v := config.NewValidator()
		if err := v.Validate(&cfg); err != nil {
			return fmt.Errorf("import file %s failed validation: %w", importPath, err)
		}

		if err := saveConfigData(&cfg); err != nil {
			return err
		}

		fmt.Printf("Configuration imported from %s\n", importPath)
		return nil
	},
}

// configMigrateCmd migrates the configuration to the latest schema version.
var configMigrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate configuration to latest version",
	Long: `Check the configuration schema version and apply any pending migrations.

If the configuration is already at the latest version, no changes are made.

Examples:
  pplx config migrate`,
	RunE: func(_ *cobra.Command, _ []string) error {
		cfg, err := loadConfigData(configFilePath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		migrated, summary, err := config.MigrateConfig(cfg)
		if err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}

		if !migrated {
			fmt.Println("Configuration is already at the latest version.")
			return nil
		}

		if err := saveConfigData(cfg); err != nil {
			return err
		}

		fmt.Printf("Migration complete: %s\n", summary)
		return nil
	},
}

func init() {
	// Add config command to root
	rootCmd.AddCommand(configCmd)
	registerConfigSubcommands()
	registerConfigFlags()
	registerGetSetUnsetResetFlags()
	registerProfileFlags()
	registerDoctorFlags()
	registerConfigFlagCompletions()
}

// registerConfigSubcommands adds all config subcommands to the command tree.
func registerConfigSubcommands() {
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configOptionsCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configProfileCmd)

	configProfileCmd.AddCommand(configProfileListCmd)
	configProfileCmd.AddCommand(configProfileCreateCmd)
	configProfileCmd.AddCommand(configProfileSwitchCmd)
	configProfileCmd.AddCommand(configProfileDeleteCmd)
	configProfileCmd.AddCommand(configProfileDiffCmd)
	configProfileCmd.AddCommand(configProfileShowCmd)
	configProfileCmd.AddCommand(configProfileEditCmd)

	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configUnsetCmd)
	configCmd.AddCommand(configResetCmd)

	configCmd.AddCommand(configExportCmd)
	configCmd.AddCommand(configImportCmd)
	configCmd.AddCommand(configMigrateCmd)
	configCmd.AddCommand(configDoctorCmd)
}

// registerConfigFlags registers flags for the existing config subcommands.
func registerConfigFlags() {
	configCmd.PersistentFlags().StringVar(&configFilePath, "config", "", "Path to config file")

	configInitCmd.Flags().StringVarP(
		&initTemplate, "template", "t", "",
		"Template to use (research, creative, news, full-example)")
	configInitCmd.Flags().BoolVar(
		&initWithExamples, "with-examples", false,
		"Include example profiles in configuration")
	configInitCmd.Flags().BoolVar(
		&initWithProfiles, "with-profiles", false,
		"Include profile configurations")
	configInitCmd.Flags().BoolVarP(
		&initForce, "force", "f", false,
		"Force overwrite existing configuration")
	configInitCmd.Flags().BoolVar(
		&initCheckEnv, "check-env", false,
		"Check environment variables (API keys)")
	configInitCmd.Flags().BoolVarP(
		&initInteractive, "interactive", "i", false,
		"Launch interactive configuration wizard")
	configInitCmd.Flags().BoolVar(
		&initDryRun, "dry-run", false,
		"Print generated YAML to stdout without writing to disk")
	configInitCmd.Flags().BoolVar(
		&initUpdate, "update", false,
		"Update existing config: load current values and only change what you specify")

	configPathCmd.Flags().BoolVarP(
		&pathCheckFlag, "check", "c", false,
		"Validate configuration and show details")

	configShowCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	configShowCmd.Flags().StringVar(&profileName, "profile", "", "Show specific profile")

	configOptionsCmd.Flags().StringVarP(
		&optionsSection, "section", "s", "",
		"Filter by section (defaults, search, output, api)")
	configOptionsCmd.Flags().StringVarP(
		&optionsFormat, "format", "f", "table",
		"Output format (table, json, yaml)")
	configOptionsCmd.Flags().BoolVarP(
		&optionsValidation, "validation", "v", false,
		"Show validation rules")

	configExportCmd.Flags().BoolVar(
		&exportWithoutSecrets, "without-secrets", false,
		"Omit sensitive values (e.g. API key) from the exported YAML")
}

// registerGetSetUnsetResetFlags registers flags for the get, set, unset, and reset subcommands.
func registerGetSetUnsetResetFlags() {
	// Flags for get command
	configGetCmd.Flags().BoolVar(&getUnmask, "unmask", false, "Show the full API key without masking")
	configGetCmd.Flags().BoolVar(&getJSON, "json", false, "Output value as JSON")
	configGetCmd.Flags().StringVar(&getProfile, "profile", "", "Read value from specific profile (merged)")

	// Flags for set command
	configSetCmd.Flags().StringVar(&setProfile, "profile", "", "Set value in specific profile")
	configSetCmd.Flags().BoolVar(&setNoValidate, "no-validate", false, "Skip validation after setting value")

	// Flags for unset command
	configUnsetCmd.Flags().StringVar(&unsetProfile, "profile", "", "Unset value in specific profile")

	// Flags for reset command
	configResetCmd.Flags().BoolVarP(&resetForce, "force", "f", false, "Skip confirmation prompt")
}

// registerProfileFlags registers flags for the profile subcommands added in T026-T031.
func registerProfileFlags() {
	// Flags for profile create command.
	configProfileCreateCmd.Flags().StringVar(
		&createFromTemplate, "from-template", "",
		"Create profile from a built-in template (research, creative, news, full-example)")
	configProfileCreateCmd.Flags().StringVar(
		&createCopyFrom, "copy-from", "",
		"Copy an existing profile as the basis for the new profile")
	configProfileCreateCmd.Flags().StringVar(
		&createDescription, "description", "",
		"Optional description for the new profile")

	// Flags for profile delete command.
	configProfileDeleteCmd.Flags().BoolVarP(
		&deleteForceFlag, "force", "f", false,
		"Skip confirmation prompt")

	// Flags for profile diff command.
	configProfileDiffCmd.Flags().BoolVar(
		&profileDiffJSON, "json", false,
		"Output diff as JSON")

	// Flags for profile show command.
	configProfileShowCmd.Flags().BoolVar(
		&profileShowJSON, "json", false,
		"Output profile as JSON")
}

// registerDoctorFlags registers flags for the config doctor subcommand.
func registerDoctorFlags() {
	configDoctorCmd.Flags().BoolVar(&doctorJSON, "json", false, "Output results as JSON")
	configDoctorCmd.Flags().BoolVar(&doctorFix, "fix", false, "Attempt to fix auto-correctable issues (e.g. file permissions)")
}

// registerConfigFlagCompletions registers completion functions for config command flags.
func registerConfigFlagCompletions() {
	// Template name completion for config init --template
	if err := configInitCmd.RegisterFlagCompletionFunc("template",
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return completion.TemplateNames(), cobra.ShellCompDirectiveNoFileComp
		}); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to register completion for 'template' flag: %v\n", err)
	}

	// Section name completion for config options --section
	if err := configOptionsCmd.RegisterFlagCompletionFunc("section",
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return completion.ConfigSections(), cobra.ShellCompDirectiveNoFileComp
		}); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to register completion for 'section' flag: %v\n", err)
	}

	// Format completion for config options --format
	if err := configOptionsCmd.RegisterFlagCompletionFunc("format",
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return completion.OutputFormats(), cobra.ShellCompDirectiveNoFileComp
		}); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to register completion for 'format' flag: %v\n", err)
	}

	registerArgCompletions()
}

// registerArgCompletions wires ValidArgsFunction for commands that complete positional args.
// Delegates to registerKeyArgCompletions and registerProfileArgCompletions to stay within
// cyclomatic complexity limits.
func registerArgCompletions() {
	registerKeyArgCompletions()
	registerProfileArgCompletions()
}

// registerKeyArgCompletions sets ValidArgsFunction on key-accepting config commands.
func registerKeyArgCompletions() {
	// config get <key>
	configGetCmd.ValidArgsFunction = func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completion.ConfigKeys(), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// config set <key> [value]
	configSetCmd.ValidArgsFunction = func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completion.ConfigKeys(), cobra.ShellCompDirectiveNoFileComp
		}
		if len(args) == 1 {
			if values := completion.ConfigValues(args[0]); values != nil {
				return values, cobra.ShellCompDirectiveNoFileComp
			}
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// config unset <key>
	configUnsetCmd.ValidArgsFunction = func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completion.ConfigKeys(), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// config reset [key]
	configResetCmd.ValidArgsFunction = func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completion.ConfigKeys(), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

// registerProfileArgCompletions sets ValidArgsFunction on profile subcommands.
func registerProfileArgCompletions() {
	// config profile switch <name>
	configProfileSwitchCmd.ValidArgsFunction = func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completion.ProfileNames(), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// config profile delete <name>
	configProfileDeleteCmd.ValidArgsFunction = func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completion.ProfileNames(), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// config profile show <name>
	configProfileShowCmd.ValidArgsFunction = func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completion.ProfileNames(), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// config profile diff <profile1> [profile2]
	configProfileDiffCmd.ValidArgsFunction = func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		if len(args) < profileDiffMaxArgs {
			return completion.ProfileNames(), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// config profile edit <name>
	configProfileEditCmd.ValidArgsFunction = func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		if len(args) == 0 {
			return completion.ProfileNames(), cobra.ShellCompDirectiveNoFileComp
		}
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}
