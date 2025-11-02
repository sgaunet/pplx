package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// ErrNoConfigFound is returned when no config file is found in the config directory.
var ErrNoConfigFound = errors.New("no config file found in ~/.config/pplx/")

// ErrPathIsDirectory is returned when a file path points to a directory instead of a file.
var ErrPathIsDirectory = errors.New("path is a directory, not a file")

// ConfigPaths defines the standard location where config files are searched.
var ConfigPaths = []string{
	"$HOME/.config/pplx", // User config directory
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

// Load loads configuration from the standard location ~/.config/pplx/
// Searches for pplx.yaml, config.yaml, pplx.yml, or config.yml in that directory.
func (l *Loader) Load() error {
	l.viper.SetConfigName("pplx")
	l.viper.SetConfigType("yaml")

	// Add config directory search path
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

// FindConfigFile searches for a config file in ~/.config/pplx/.
// Files are checked in precedence order: config.yaml, pplx.yaml, config.yml, pplx.yml.
func FindConfigFile() (string, error) {
	for _, basePath := range ConfigPaths {
		expandedPath := os.ExpandEnv(basePath)

		// Try config.yaml (highest precedence)
		configPath := filepath.Join(expandedPath, "config.yaml")
		if fileExists(configPath) {
			return configPath, nil
		}

		// Try pplx.yaml
		configPath = filepath.Join(expandedPath, "pplx.yaml")
		if fileExists(configPath) {
			return configPath, nil
		}

		// Try config.yml
		configPath = filepath.Join(expandedPath, "config.yml")
		if fileExists(configPath) {
			return configPath, nil
		}

		// Try pplx.yml
		configPath = filepath.Join(expandedPath, "pplx.yml")
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

// ListConfigFiles returns information about all configuration files in ~/.config/pplx/.
func ListConfigFiles() ([]ConfigFileInfo, error) {
	// Get config directory
	configDir := os.ExpandEnv(ConfigPaths[0])

	// Check if directory exists
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return []ConfigFileInfo{}, nil // Empty list if directory doesn't exist
	}

	// Read directory contents
	entries, err := os.ReadDir(configDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read config directory: %w", err)
	}

	files := make([]ConfigFileInfo, 0, len(entries))

	// Find the active config file based on precedence
	activeFile, _ := FindConfigFile()

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		// Only include .yaml and .yml files
		if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}

		path := filepath.Join(configDir, name)
		info, err := entry.Info()
		if err != nil {
			continue // Skip files we can't stat
		}

		fileInfo := ConfigFileInfo{
			Name:     name,
			Path:     path,
			Size:     info.Size(),
			ModTime:  info.ModTime(),
			IsActive: path == activeFile,
		}

		// Try to validate and count profiles
		loader := NewLoader()
		if err := loader.LoadFrom(path); err == nil {
			fileInfo.IsValid = true
			fileInfo.ProfileCount = len(loader.Data().Profiles)
		} else {
			fileInfo.IsValid = false
			fileInfo.ProfileCount = 0
		}

		files = append(files, fileInfo)
	}

	// Sort files by precedence: config.yaml, pplx.yaml, then alphabetically
	sortConfigFiles(files)

	return files, nil
}

// sortConfigFiles sorts config files by precedence.
// Standard names (config.yaml, pplx.yaml, config.yml, pplx.yml) come first in that order,
// followed by other files sorted alphabetically.
func sortConfigFiles(files []ConfigFileInfo) {
	const (
		precedenceConfig     = 1
		precedencePplx       = 2
		precedenceConfigYml  = 3
		precedencePplxYml    = 4
		precedenceNonStandard = 100
	)

	precedence := map[string]int{
		"config.yaml": precedenceConfig,
		"pplx.yaml":   precedencePplx,
		"config.yml":  precedenceConfigYml,
		"pplx.yml":    precedencePplxYml,
	}

	getPrecedence := func(name string) int {
		if p, ok := precedence[name]; ok {
			return p
		}
		return precedenceNonStandard
	}

	// Simple bubble sort by precedence, then alphabetically
	for i := 0; i < len(files); i++ {
		for j := i + 1; j < len(files); j++ {
			pi := getPrecedence(files[i].Name)
			pj := getPrecedence(files[j].Name)

			var shouldSwap bool
			if pi == pj {
				// Same precedence level, sort alphabetically
				shouldSwap = files[i].Name > files[j].Name
			} else {
				// Different precedence levels
				shouldSwap = pi > pj
			}

			if shouldSwap {
				files[i], files[j] = files[j], files[i]
			}
		}
	}
}

// ValidateYAMLFile checks if a file contains valid YAML syntax.
// Returns true if the file is valid YAML, false otherwise.
func ValidateYAMLFile(path string) (bool, error) {
	loader := NewLoader()
	if err := loader.LoadFrom(path); err != nil {
		return false, err
	}
	return true, nil
}

// CountProfiles counts the number of profiles in a config file.
// Returns 0 if the file is invalid or has no profiles.
func CountProfiles(path string) (int, error) {
	loader := NewLoader()
	if err := loader.LoadFrom(path); err != nil {
		return 0, err
	}
	return len(loader.Data().Profiles), nil
}

// DetermineActiveConfig determines which config file is currently active
// based on precedence rules (config.yaml > pplx.yaml > alphabetical).
func DetermineActiveConfig(configFiles []ConfigFileInfo) *ConfigFileInfo {
	if len(configFiles) == 0 {
		return nil
	}

	// Find the active file using FindConfigFile
	activePath, err := FindConfigFile()
	if err != nil {
		return nil
	}

	// Find and return the matching ConfigFileInfo
	for i := range configFiles {
		if configFiles[i].Path == activePath {
			return &configFiles[i]
		}
	}

	return nil
}

// FormatFileSize converts bytes to a human-readable format (B, KB, MB).
func FormatFileSize(bytes int64) string {
	const (
		kb = 1024
		mb = kb * 1024
	)

	switch {
	case bytes >= mb:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// FormatTimestamp formats a timestamp for consistent display.
func FormatTimestamp(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// GetConfigFileInfo gathers all metadata for a single config file.
func GetConfigFileInfo(path string) (*ConfigFileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	if info.IsDir() {
		return nil, ErrPathIsDirectory
	}

	fileInfo := &ConfigFileInfo{
		Name:    filepath.Base(path),
		Path:    path,
		Size:    info.Size(),
		ModTime: info.ModTime(),
	}

	// Try to validate and count profiles
	loader := NewLoader()
	if err := loader.LoadFrom(path); err == nil {
		fileInfo.IsValid = true
		fileInfo.ProfileCount = len(loader.Data().Profiles)
	} else {
		fileInfo.IsValid = false
		fileInfo.ProfileCount = 0
	}

	// Determine if this is the active config
	activeFile, _ := FindConfigFile()
	fileInfo.IsActive = path == activeFile

	return fileInfo, nil
}

// AnalyzeConfigFile analyzes a config file and returns all metadata.
// This function provides complete information about a config file including
// validation status, profile count, and whether it's the active config.
// Returns partial information even if validation fails.
func AnalyzeConfigFile(path string) (*ConfigFileInfo, error) {
	// Use GetConfigFileInfo which already implements all the analysis
	return GetConfigFileInfo(path)
}
