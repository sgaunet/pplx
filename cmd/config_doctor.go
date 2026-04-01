package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/sgaunet/pplx/pkg/clerrors"
	"github.com/sgaunet/pplx/pkg/config"
	"github.com/spf13/cobra"
)

const (
	// doctorSymbolPass is the symbol shown for a passing check.
	doctorSymbolPass = "\u2713" // ✓
	// doctorSymbolFail is the symbol shown for a failing check.
	doctorSymbolFail = "\u2717" // ✗
	// doctorSymbolWarn is the symbol shown for a warning check.
	doctorSymbolWarn = "\u26a0" // ⚠
	// configFilePermMode is the expected permission mode for the config file.
	configFilePermMode = 0o600
)

var (
	doctorJSON bool
	doctorFix  bool
)

var configDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check configuration health",
	Long: `Run a series of health checks on your pplx configuration.

Checks performed:
  - Config file existence
  - File permissions (warns if not 0600)
  - YAML syntax validity
  - Field validation
  - Profile integrity (active profile exists)
  - API key availability
  - Config version field

Examples:
  pplx config doctor
  pplx config doctor --json
  pplx config doctor --fix`,
	RunE: runConfigDoctor,
}

func runConfigDoctor(_ *cobra.Command, _ []string) error {
	// Resolve config path: prefer --config flag, then auto-discover.
	path := configFilePath // set by the persistent --config flag on configCmd

	checks := config.RunHealthChecks(path)

	if doctorJSON {
		return printDoctorJSON(checks)
	}

	return printDoctorTable(checks, path)
}

// printDoctorJSON serialises the health checks as a JSON array.
func printDoctorJSON(checks []config.HealthCheck) error {
	type jsonCheck struct {
		Name   string `json:"name"`
		Status string `json:"status"`
		Detail string `json:"detail"`
	}

	out := make([]jsonCheck, len(checks))
	for i, c := range checks {
		out[i] = jsonCheck{
			Name:   c.Name,
			Status: statusString(c.Status),
			Detail: c.Detail,
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		return fmt.Errorf("encoding health checks as JSON: %w", err)
	}
	return nil
}

// printDoctorTable renders the health checks in human-readable table format.
func printDoctorTable(checks []config.HealthCheck, resolvedPath string) error {
	fmt.Println("Configuration Health Check")
	fmt.Println()

	// Apply --fix before printing results so the permission status reflects reality.
	if doctorFix {
		applyFixes(checks, resolvedPath)
		// Re-run checks so the output reflects any fixes.
		checks = config.RunHealthChecks(resolvedPath)
	}

	// Determine the longest check name for alignment.
	maxLen := 0
	for _, c := range checks {
		if len(c.Name) > maxLen {
			maxLen = len(c.Name)
		}
	}

	passed, failed, warned := 0, 0, 0
	for _, c := range checks {
		sym := symbolFor(c.Status)
		// Left-pad the name so details align.
		label := fmt.Sprintf("  %-*s", maxLen, c.Name+":")
		fmt.Printf("%s %s %s\n", label, sym, c.Detail)

		switch c.Status {
		case config.CheckPass:
			passed++
		case config.CheckFail:
			failed++
		case config.CheckWarn:
			warned++
		}
	}

	fmt.Println()
	total := len(checks)
	fmt.Printf("%d/%d checks passed", passed, total)
	if warned > 0 {
		fmt.Printf(", %d warning(s)", warned)
	}
	fmt.Println(".")

	if failed > 0 {
		return fmt.Errorf("%w: %d check(s) failed", clerrors.ErrHealthChecksFailed, failed)
	}
	return nil
}

// applyFixes attempts to fix auto-correctable issues.
// Currently handles file permission correction (chmod 0600).
func applyFixes(checks []config.HealthCheck, configPath string) {
	for _, c := range checks {
		if c.Name == "File Permissions" && c.Status == config.CheckWarn {
			fixFilePermissions(configPath)
		}
	}
}

// fixFilePermissions sets the config file permissions to 0600.
// If configPath is empty it auto-discovers the config file first.
func fixFilePermissions(configPath string) {
	path, err := resolveConfigPath(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fix: cannot locate config file: %v\n", err)
		return
	}
	if err := os.Chmod(path, configFilePermMode); err != nil {
		fmt.Fprintf(os.Stderr, "fix: failed to chmod %s: %v\n", path, err)
	} else {
		fmt.Printf("fix: set permissions to 0600 on %s\n", path)
	}
}

// resolveConfigPath returns configPath as-is if non-empty, or auto-discovers it.
func resolveConfigPath(configPath string) (string, error) {
	if configPath != "" {
		return configPath, nil
	}
	found, err := config.FindConfigFile()
	if err != nil {
		return "", fmt.Errorf("locating config file: %w", err)
	}
	return found, nil
}

// symbolFor returns the display symbol for a check status.
func symbolFor(s config.CheckStatus) string {
	switch s {
	case config.CheckPass:
		return doctorSymbolPass
	case config.CheckFail:
		return doctorSymbolFail
	case config.CheckWarn:
		return doctorSymbolWarn
	default:
		return "?"
	}
}

// statusString converts a CheckStatus to a JSON-friendly string.
func statusString(s config.CheckStatus) string {
	switch s {
	case config.CheckPass:
		return "pass"
	case config.CheckFail:
		return "fail"
	case config.CheckWarn:
		return "warn"
	default:
		return "unknown"
	}
}
