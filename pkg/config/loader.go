package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// ErrNoConfigFound is returned when no config file is found in standard locations.
var ErrNoConfigFound = errors.New("no config file found in standard locations")

// ConfigPaths defines the standard locations where config files are searched.
var ConfigPaths = []string{
	".",                  // Current directory
	"$HOME/.config/pplx", // User config directory
	"/etc/pplx",          // System-wide config directory
}

// Loader handles loading configuration from files.
type Loader struct {
	viper *viper.Viper
	data  *ConfigData
}

// NewLoader creates a new configuration loader.
func NewLoader() *Loader {
	return &Loader{
		viper: viper.New(),
		data:  NewConfigData(),
	}
}

// Load loads configuration from standard locations
// Searches in order: ./pplx.yaml, ~/.config/pplx/config.yaml, /etc/pplx/config.yaml.
func (l *Loader) Load() error {
	l.viper.SetConfigName("pplx")
	l.viper.SetConfigType("yaml")

	// Add all search paths
	for _, path := range ConfigPaths {
		expandedPath := os.ExpandEnv(path)
		l.viper.AddConfigPath(expandedPath)
	}

	// Also support config.yaml as filename
	l.viper.SetConfigName("config")
	for _, path := range ConfigPaths {
		expandedPath := os.ExpandEnv(path)
		l.viper.AddConfigPath(expandedPath)
	}

	// Reset to pplx as primary name
	l.viper.SetConfigName("pplx")

	// Try to read config file
	if err := l.viper.ReadInConfig(); err != nil {
		// It's okay if config file doesn't exist
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return fmt.Errorf("error reading config file: %w", err)
		}
	}

	// Unmarshal config into data structure
	if err := l.viper.Unmarshal(l.data); err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
	}

	return nil
}

// LoadFrom loads configuration from a specific file path.
func (l *Loader) LoadFrom(path string) error {
	l.viper.SetConfigFile(path)

	if err := l.viper.ReadInConfig(); err != nil {
		return fmt.Errorf("error reading config file %s: %w", path, err)
	}

	if err := l.viper.Unmarshal(l.data); err != nil {
		return fmt.Errorf("error unmarshaling config: %w", err)
	}

	return nil
}

// Data returns the loaded configuration data.
func (l *Loader) Data() *ConfigData {
	return l.data
}

// Viper returns the underlying viper instance.
func (l *Loader) Viper() *viper.Viper {
	return l.viper
}

// FindConfigFile searches for a config file in standard locations.
func FindConfigFile() (string, error) {
	for _, basePath := range ConfigPaths {
		expandedPath := os.ExpandEnv(basePath)

		// Try pplx.yaml
		configPath := filepath.Join(expandedPath, "pplx.yaml")
		if fileExists(configPath) {
			return configPath, nil
		}

		// Try config.yaml
		configPath = filepath.Join(expandedPath, "config.yaml")
		if fileExists(configPath) {
			return configPath, nil
		}

		// Try pplx.yml
		configPath = filepath.Join(expandedPath, "pplx.yml")
		if fileExists(configPath) {
			return configPath, nil
		}

		// Try config.yml
		configPath = filepath.Join(expandedPath, "config.yml")
		if fileExists(configPath) {
			return configPath, nil
		}
	}

	return "", ErrNoConfigFound
}

// fileExists checks if a file exists.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// GetDefaultConfigPath returns the default path where config should be created.
func GetDefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "./pplx.yaml"
	}

	configDir := filepath.Join(home, ".config", "pplx")
	return filepath.Join(configDir, "config.yaml")
}
