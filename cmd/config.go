package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sgaunet/pplx/pkg/completion"
	"github.com/sgaunet/pplx/pkg/config"
	"github.com/sgaunet/pplx/pkg/logger"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Errors for config operations.
var (
	ErrConfigFileExists = errors.New("configuration file already exists")
	ErrValidationFailed = errors.New("validation failed")
	ErrUnknownSection   = errors.New("unknown section")
)

const (
	// File permission constants.
	configFilePermission = 0600
	configDirPermission  = 0750
)

var (
	configFilePath    string
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
)

// saveConfigData saves configuration data to a file.
func saveConfigData(data *config.ConfigData) error {
	configPath, err := config.FindConfigFile()
	if err != nil {
		configPath = config.GetDefaultConfigPath()
	}

	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, yamlData, configFilePermission); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
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
  pplx config init --template creative --with-examples --check-env`,
	RunE: runConfigInit,
}

// loadOrCreateConfig loads configuration based on init flags.
func loadOrCreateConfig() (*config.ConfigData, error) {
	switch {
	case initInteractive:
		wizard := NewWizardState()
		cfg, err := wizard.Run()
		if err != nil {
			return nil, fmt.Errorf("wizard failed: %w", err)
		}
		return cfg, nil

	case initTemplate != "":
		cfg, err := config.LoadTemplate(initTemplate)
		if err != nil {
			return nil, fmt.Errorf("failed to load template: %w", err)
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
	configPath := config.GetDefaultConfigPath()
	if configFilePath != "" {
		configPath = configFilePath
	}

	// Check if file already exists
	if _, err := os.Stat(configPath); err == nil {
		if !initForce {
			return fmt.Errorf("%w at %s (use --force to overwrite)", ErrConfigFileExists, configPath)
		}
		fmt.Printf("Overwriting existing configuration at %s\n", configPath)
	}

	// Create directory if it doesn't exist
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, configDirPermission); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check environment if requested
	if initCheckEnv {
		checkEnvironment()
	}

	// Load or create configuration
	cfg, err := loadOrCreateConfig()
	if err != nil {
		return err
	}

	// Generate YAML content
	yamlContent, err := generateYAMLContent(cfg)
	if err != nil {
		return err
	}

	// Write to file
	if err := os.WriteFile(configPath, []byte(yamlContent), configFilePermission); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Configuration file created at %s\n", configPath)

	// Verify permissions after creation
	if err := verifyConfigPermissions(configPath); err != nil {
		// Don't fail the init operation, just log the warning
		logger.Warn("config file permissions check failed", "error", err)
	}

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
		return fmt.Errorf("failed to check file permissions: %w", err)
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
				return fmt.Errorf("failed to load config: %w", err)
			}
		} else {
			if err := loader.Load(); err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
		}

		cfg := loader.Data()

		// If profile specified, show only that profile
		if profileName != "" {
			pm := config.NewProfileManager(cfg)
			profile, err := pm.LoadProfile(profileName)
			if err != nil {
				return fmt.Errorf("failed to load profile: %w", err)
			}

			if jsonOutput {
				data, err := json.MarshalIndent(profile, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal profile: %w", err)
				}
				fmt.Println(string(data))
			} else {
				data, err := yaml.Marshal(profile)
				if err != nil {
					return fmt.Errorf("failed to marshal profile: %w", err)
				}
				fmt.Print(string(data))
			}
			return nil
		}

		// Show full config
		if jsonOutput {
			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal config: %w", err)
			}
			fmt.Println(string(data))
		} else {
			data, err := yaml.Marshal(cfg)
			if err != nil {
				return fmt.Errorf("failed to marshal config: %w", err)
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
				return fmt.Errorf("failed to load config: %w", err)
			}
		} else {
			if err := loader.Load(); err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
		}

		cfg := loader.Data()

		// Validate configuration
		validator := config.NewValidator()
		if err := validator.Validate(cfg); err != nil {
			fmt.Println("Configuration validation failed:")
			fmt.Println(err.Error())
			return ErrValidationFailed
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
			return fmt.Errorf("failed to open editor: %w", err)
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
			return fmt.Errorf("%w: %s (valid: defaults, search, output, api)", ErrUnknownSection, optionsSection)
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
		return fmt.Errorf("failed to format options: %w", err)
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
			return fmt.Errorf("failed to load config: %w", err)
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
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		profileName := args[0]

		loader := config.NewLoader()
		if err := loader.Load(); err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		pm := config.NewProfileManager(loader.Data())
		if _, err := pm.CreateProfile(profileName, ""); err != nil {
			return fmt.Errorf("failed to create profile: %w", err)
		}

		// Save config
		configPath, err := config.FindConfigFile()
		if err != nil {
			configPath = config.GetDefaultConfigPath()
		}

		data, err := yaml.Marshal(loader.Data())
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		if err := os.WriteFile(configPath, data, configFilePermission); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("Profile '%s' created successfully\n", profileName)
		return nil
	},
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
			return fmt.Errorf("failed to load config: %w", err)
		}

		pm := config.NewProfileManager(loader.Data())
		if err := pm.SetActiveProfile(profileName); err != nil {
			return fmt.Errorf("failed to switch profile: %w", err)
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
		profileName := args[0]

		loader := config.NewLoader()
		if err := loader.Load(); err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		pm := config.NewProfileManager(loader.Data())
		if err := pm.DeleteProfile(profileName); err != nil {
			return fmt.Errorf("failed to delete profile: %w", err)
		}

		// Save config
		if err := saveConfigData(loader.Data()); err != nil {
			return err
		}

		fmt.Printf("Profile '%s' deleted successfully\n", profileName)
		return nil
	},
}

func init() {
	// Add config command to root
	rootCmd.AddCommand(configCmd)

	// Add subcommands
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configValidateCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configOptionsCmd)
	configCmd.AddCommand(configPathCmd)
	configCmd.AddCommand(configProfileCmd)

	// Add profile subcommands
	configProfileCmd.AddCommand(configProfileListCmd)
	configProfileCmd.AddCommand(configProfileCreateCmd)
	configProfileCmd.AddCommand(configProfileSwitchCmd)
	configProfileCmd.AddCommand(configProfileDeleteCmd)

	// Flags for config command
	configCmd.PersistentFlags().StringVar(&configFilePath, "config", "", "Path to config file")

	// Flags for init command
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

	// Flags for path command
	configPathCmd.Flags().BoolVarP(
		&pathCheckFlag, "check", "c", false,
		"Validate configuration and show details")

	// Flags for show command
	configShowCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	configShowCmd.Flags().StringVar(&profileName, "profile", "", "Show specific profile")

	// Flags for options command
	configOptionsCmd.Flags().StringVarP(
		&optionsSection, "section", "s", "",
		"Filter by section (defaults, search, output, api)")
	configOptionsCmd.Flags().StringVarP(
		&optionsFormat, "format", "f", "table",
		"Output format (table, json, yaml)")
	configOptionsCmd.Flags().BoolVarP(
		&optionsValidation, "validation", "v", false,
		"Show validation rules")

	// Register flag completions
	registerConfigFlagCompletions()
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
}
