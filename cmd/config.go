package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sgaunet/pplx/pkg/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Errors for config operations.
var (
	ErrConfigFileExists = errors.New("configuration file already exists")
	ErrValidationFailed = errors.New("validation failed")
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
	Long:  `Create a new configuration file with default values at ~/.config/pplx/config.yaml`,
	RunE: func(_ *cobra.Command, _ []string) error {
		configPath := config.GetDefaultConfigPath()
		if configFilePath != "" {
			configPath = configFilePath
		}

		// Check if file already exists
		if _, err := os.Stat(configPath); err == nil {
			return fmt.Errorf("%w at %s", ErrConfigFileExists, configPath)
		}

		// Create directory if it doesn't exist
		configDir := filepath.Dir(configPath)
		if err := os.MkdirAll(configDir, configDirPermission); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		// Create default config
		cfg := config.NewConfigData()
		cfg.Defaults.Model = "sonar"
		cfg.Defaults.Temperature = 0.2
		cfg.Defaults.MaxTokens = 4000

		// Marshal to YAML
		data, err := yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		// Write to file
		if err := os.WriteFile(configPath, data, configFilePermission); err != nil {
			return fmt.Errorf("failed to write config file: %w", err)
		}

		fmt.Printf("Configuration file created at %s\n", configPath)
		return nil
	},
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
			fmt.Printf("Warning: Configuration file has syntax errors: %v\n", err)
			return nil
		}

		validator := config.NewValidator()
		if err := validator.Validate(loader.Data()); err != nil {
			fmt.Printf("Warning: Configuration validation failed: %v\n", err)
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
			return fmt.Errorf("unknown section: %s (valid: defaults, search, output, api)", optionsSection)
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
	configCmd.AddCommand(configProfileCmd)

	// Add profile subcommands
	configProfileCmd.AddCommand(configProfileListCmd)
	configProfileCmd.AddCommand(configProfileCreateCmd)
	configProfileCmd.AddCommand(configProfileSwitchCmd)
	configProfileCmd.AddCommand(configProfileDeleteCmd)

	// Flags for config command
	configCmd.PersistentFlags().StringVar(&configFilePath, "config", "", "Path to config file")

	// Flags for show command
	configShowCmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	configShowCmd.Flags().StringVar(&profileName, "profile", "", "Show specific profile")

	// Flags for options command
	configOptionsCmd.Flags().StringVarP(&optionsSection, "section", "s", "", "Filter by section (defaults, search, output, api)")
	configOptionsCmd.Flags().StringVarP(&optionsFormat, "format", "f", "table", "Output format (table, json, yaml)")
	configOptionsCmd.Flags().BoolVarP(&optionsValidation, "validation", "v", false, "Show validation rules")
}
