package config

import (
	"fmt"
	"os"
)

// CheckStatus represents the result of a health check.
type CheckStatus int

const (
	// CheckPass indicates the check passed successfully.
	CheckPass CheckStatus = iota
	// CheckFail indicates the check failed.
	CheckFail
	// CheckWarn indicates the check passed with a warning.
	CheckWarn
)

const (
	// expectedHealthChecks is the number of health checks performed by RunHealthChecks.
	expectedHealthChecks = 7
	// expectedFilePermissions is the expected file permissions for the config file.
	expectedFilePermissions = 0o600
)

// HealthCheck represents a single diagnostic check result.
type HealthCheck struct {
	Name   string      `json:"name"`
	Status CheckStatus `json:"status"`
	Detail string      `json:"detail"`
}

// RunHealthChecks executes all configuration health checks for the given config
// file path and returns the results. If configPath is empty, it attempts to find
// the config file automatically.
func RunHealthChecks(configPath string) []HealthCheck {
	checks := make([]HealthCheck, 0, expectedHealthChecks)

	// Check 1: Config file existence.
	path, existCheck := checkConfigFileExists(configPath)
	checks = append(checks, existCheck)

	if existCheck.Status == CheckFail {
		// Remaining checks are meaningless without a file.
		checks = append(checks,
			HealthCheck{Name: "File Permissions", Status: CheckFail, Detail: "skipped: no config file found"},
			HealthCheck{Name: "YAML Syntax", Status: CheckFail, Detail: "skipped: no config file found"},
			HealthCheck{Name: "Field Validation", Status: CheckFail, Detail: "skipped: no config file found"},
			HealthCheck{Name: "Profile Integrity", Status: CheckFail, Detail: "skipped: no config file found"},
			HealthCheck{Name: "API Key", Status: checkAPIKey("", nil).Status, Detail: checkAPIKey("", nil).Detail},
			HealthCheck{Name: "Config Version", Status: CheckFail, Detail: "skipped: no config file found"},
		)
		return checks
	}

	// Check 2: File permissions (warn if not 0600).
	checks = append(checks, checkFilePermissions(path))

	// Check 3: YAML syntax.
	yamlCheck := checkYAMLSyntax(path)
	checks = append(checks, yamlCheck)

	// Load the config for remaining checks.
	var data *ConfigData
	if yamlCheck.Status != CheckFail {
		loader := NewLoader()
		if err := loader.LoadFrom(path); err == nil {
			data = loader.Data()
		}
	}

	// Check 4: Field validation.
	checks = append(checks, checkFieldValidation(data))

	// Check 5: Profile integrity.
	checks = append(checks, checkProfileIntegrity(data))

	// Check 6: API key availability.
	var apiKey string
	if data != nil {
		apiKey = data.API.Key
	}
	checks = append(checks, checkAPIKey(apiKey, data))

	// Check 7: Config version.
	checks = append(checks, checkConfigVersion(data))

	return checks
}

// checkConfigFileExists verifies the config file is present.
// It returns both the resolved path and the check result.
func checkConfigFileExists(configPath string) (string, HealthCheck) {
	name := "Config File"

	// If a specific path was given, check it directly.
	if configPath != "" {
		if _, err := os.Stat(configPath); err != nil {
			return "", HealthCheck{
				Name:   name,
				Status: CheckFail,
				Detail: "not found at " + configPath,
			}
		}
		return configPath, HealthCheck{
			Name:   name,
			Status: CheckPass,
			Detail: "found at " + configPath,
		}
	}

	// Auto-discover.
	found, err := FindConfigFile()
	if err != nil {
		return "", HealthCheck{
			Name:   name,
			Status: CheckFail,
			Detail: "no config file found in ~/.config/pplx/ (run: pplx config init)",
		}
	}
	return found, HealthCheck{
		Name:   name,
		Status: CheckPass,
		Detail: "found at " + found,
	}
}

// checkFilePermissions warns when the config file is not restricted to 0600.
func checkFilePermissions(path string) HealthCheck {
	name := "File Permissions"

	info, err := os.Stat(path)
	if err != nil {
		return HealthCheck{
			Name:   name,
			Status: CheckFail,
			Detail: fmt.Sprintf("cannot stat file: %v", err),
		}
	}

	mode := info.Mode().Perm()
	if mode != expectedFilePermissions {
		return HealthCheck{
			Name:   name,
			Status: CheckWarn,
			Detail: fmt.Sprintf("%04o (should be 0600 — run: chmod 600 %s)", mode, path),
		}
	}

	return HealthCheck{Name: name, Status: CheckPass, Detail: "0600"}
}

// checkYAMLSyntax validates the config file contains parseable YAML.
func checkYAMLSyntax(path string) HealthCheck {
	name := "YAML Syntax"

	if _, err := ValidateYAMLFile(path); err != nil {
		return HealthCheck{
			Name:   name,
			Status: CheckFail,
			Detail: fmt.Sprintf("invalid YAML: %v", err),
		}
	}

	return HealthCheck{Name: name, Status: CheckPass, Detail: "valid"}
}

// checkFieldValidation runs the config validator against the loaded data.
func checkFieldValidation(data *ConfigData) HealthCheck {
	name := "Field Validation"

	if data == nil {
		return HealthCheck{Name: name, Status: CheckFail, Detail: "skipped: config could not be loaded"}
	}

	v := NewValidator()
	if err := v.Validate(data); err != nil {
		return HealthCheck{
			Name:   name,
			Status: CheckFail,
			Detail: fmt.Sprintf("validation errors: %v", err),
		}
	}

	return HealthCheck{Name: name, Status: CheckPass, Detail: "all fields valid"}
}

// checkProfileIntegrity verifies the active profile exists in the profiles map.
func checkProfileIntegrity(data *ConfigData) HealthCheck {
	name := "Profile Integrity"

	if data == nil {
		return HealthCheck{Name: name, Status: CheckFail, Detail: "skipped: config could not be loaded"}
	}

	active := data.ActiveProfile
	if active == "" || active == DefaultProfileName {
		return HealthCheck{Name: name, Status: CheckPass, Detail: "using built-in default profile"}
	}

	if _, ok := data.Profiles[active]; !ok {
		return HealthCheck{
			Name:   name,
			Status: CheckFail,
			Detail: fmt.Sprintf("active profile %q not found in profiles map", active),
		}
	}

	return HealthCheck{
		Name:   name,
		Status: CheckPass,
		Detail: fmt.Sprintf("active profile %q exists", active),
	}
}

// checkAPIKey verifies an API key is available from the environment or config.
func checkAPIKey(configKey string, _ *ConfigData) HealthCheck {
	name := "API Key"

	// Environment variables take highest precedence.
	// PPLX_API_KEY is the primary env var used by the CLI.
	if envKey := os.Getenv("PPLX_API_KEY"); envKey != "" {
		return HealthCheck{
			Name:   name,
			Status: CheckPass,
			Detail: "set via PPLX_API_KEY environment variable",
		}
	}

	if envKey := os.Getenv("PERPLEXITY_API_KEY"); envKey != "" {
		return HealthCheck{
			Name:   name,
			Status: CheckPass,
			Detail: "set via PERPLEXITY_API_KEY environment variable",
		}
	}

	if configKey != "" {
		return HealthCheck{
			Name:   name,
			Status: CheckPass,
			Detail: "set via config api.key",
		}
	}

	return HealthCheck{
		Name:   name,
		Status: CheckFail,
		Detail: "not found (set PPLX_API_KEY or PERPLEXITY_API_KEY env var, or api.key in config)",
	}
}

// checkConfigVersion warns when the Version field is absent or zero.
func checkConfigVersion(data *ConfigData) HealthCheck {
	name := "Config Version"

	if data == nil {
		return HealthCheck{Name: name, Status: CheckFail, Detail: "skipped: config could not be loaded"}
	}

	if data.Version <= 0 {
		return HealthCheck{
			Name:   name,
			Status: CheckWarn,
			Detail: "version field is 0 or missing (add 'version: 1' to your config)",
		}
	}

	return HealthCheck{
		Name:   name,
		Status: CheckPass,
		Detail: fmt.Sprintf("version %d", data.Version),
	}
}
