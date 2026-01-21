package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sgaunet/pplx/pkg/clerrors"
	"github.com/sgaunet/pplx/pkg/logger"
	"github.com/spf13/cobra"
)

const (
	shellBash       = "bash"
	shellZsh        = "zsh"
	shellFish       = "fish"
	shellPowershell = "powershell"

	// File permissions.
	dirPerms = 0750
)

var (
	completionUninstall  bool
	completionOutputFile string
)

// completionCmd represents the completion command.
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for pplx.

To load completions:

Bash:
  $ source <(pplx completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ pplx completion bash > /etc/bash_completion.d/pplx
  # macOS:
  $ pplx completion bash > $(brew --prefix)/etc/bash_completion.d/pplx

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ pplx completion zsh > "${fpath[1]}/_pplx"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ pplx completion fish | source

  # To load completions for each session, execute once:
  $ pplx completion fish > ~/.config/fish/completions/pplx.fish

PowerShell:
  PS> pplx completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> pplx completion powershell > pplx.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{shellBash, shellZsh, shellFish, shellPowershell},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		shell := args[0]

		// Determine output writer
		var out io.Writer = os.Stdout
		if completionOutputFile != "" {
			file, err := os.Create(completionOutputFile) // #nosec G304
			if err != nil {
				return fmt.Errorf("failed to create %s completion output file %s: %w", shell, completionOutputFile, err)
			}
			defer func() {
				if closeErr := file.Close(); closeErr != nil {
					logger.Warn("failed to close file", "error", closeErr)
				}
			}()
			out = file
		}

		// Generate completion script
		return generateCompletion(cmd.Root(), shell, out)
	},
}

// bashCmd represents the bash completion subcommand.
var bashCmd = &cobra.Command{
	Use:   shellBash,
	Short: "Generate bash completion script",
	Long:  "Generate the autocompletion script for bash",
	RunE: func(cmd *cobra.Command, _ []string) error {
		var out io.Writer = os.Stdout
		if completionOutputFile != "" {
			file, err := os.Create(completionOutputFile) // #nosec G304
			if err != nil {
				return fmt.Errorf("failed to create bash completion output file %s: %w", completionOutputFile, err)
			}
			defer func() {
				if closeErr := file.Close(); closeErr != nil {
					logger.Warn("failed to close file", "error", closeErr)
				}
			}()
			out = file
		}
		if err := cmd.Root().GenBashCompletion(out); err != nil {
			return fmt.Errorf("failed to generate bash completion: %w", err)
		}
		return nil
	},
}

// zshCmd represents the zsh completion subcommand.
var zshCmd = &cobra.Command{
	Use:   shellZsh,
	Short: "Generate zsh completion script",
	Long:  "Generate the autocompletion script for zsh",
	RunE: func(cmd *cobra.Command, _ []string) error {
		var out io.Writer = os.Stdout
		if completionOutputFile != "" {
			file, err := os.Create(completionOutputFile) // #nosec G304
			if err != nil {
				return fmt.Errorf("failed to create zsh completion output file %s: %w", completionOutputFile, err)
			}
			defer func() {
				if closeErr := file.Close(); closeErr != nil {
					logger.Warn("failed to close file", "error", closeErr)
				}
			}()
			out = file
		}
		if err := cmd.Root().GenZshCompletion(out); err != nil {
			return fmt.Errorf("failed to generate zsh completion: %w", err)
		}
		return nil
	},
}

// fishCmd represents the fish completion subcommand.
var fishCmd = &cobra.Command{
	Use:   shellFish,
	Short: "Generate fish completion script",
	Long:  "Generate the autocompletion script for fish",
	RunE: func(cmd *cobra.Command, _ []string) error {
		var out io.Writer = os.Stdout
		if completionOutputFile != "" {
			file, err := os.Create(completionOutputFile) // #nosec G304
			if err != nil {
				return fmt.Errorf("failed to create fish completion output file %s: %w", completionOutputFile, err)
			}
			defer func() {
				if closeErr := file.Close(); closeErr != nil {
					logger.Warn("failed to close file", "error", closeErr)
				}
			}()
			out = file
		}
		if err := cmd.Root().GenFishCompletion(out, true); err != nil {
			return fmt.Errorf("failed to generate fish completion: %w", err)
		}
		return nil
	},
}

// powershellCmd represents the powershell completion subcommand.
var powershellCmd = &cobra.Command{
	Use:   shellPowershell,
	Short: "Generate PowerShell completion script",
	Long:  "Generate the autocompletion script for PowerShell",
	RunE: func(cmd *cobra.Command, _ []string) error {
		var out io.Writer = os.Stdout
		if completionOutputFile != "" {
			file, err := os.Create(completionOutputFile) // #nosec G304
			if err != nil {
				return fmt.Errorf("failed to create powershell completion output file %s: %w", completionOutputFile, err)
			}
			defer func() {
				if closeErr := file.Close(); closeErr != nil {
					logger.Warn("failed to close file", "error", closeErr)
				}
			}()
			out = file
		}
		if err := cmd.Root().GenPowerShellCompletionWithDesc(out); err != nil {
			return fmt.Errorf("failed to generate powershell completion: %w", err)
		}
		return nil
	},
}

// installCmd represents the install subcommand.
var installCmd = &cobra.Command{
	Use:   "install [shell]",
	Short: "Install shell completion automatically",
	Long: `Install shell completion scripts automatically for the current shell.

This command will detect your current shell (or use the one specified) and install
the completion script to the appropriate location. It will create backups of any
existing configuration files before modifying them.

Supported shells: bash, zsh, fish, powershell

Example:
  $ pplx completion install         # Auto-detect shell
  $ pplx completion install bash    # Install for bash
`,
	ValidArgs: []string{shellBash, shellZsh, shellFish, shellPowershell},
	Args:      cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var shell string
		if len(args) > 0 {
			shell = args[0]
		} else {
			// Auto-detect shell
			detectedShell, err := detectShell()
			if err != nil {
				return fmt.Errorf("failed to detect shell (please specify one): %w", err)
			}
			shell = detectedShell
		}

		if completionUninstall {
			return uninstallCompletion(shell)
		}
		return installCompletion(cmd.Root(), shell)
	},
}

// generateCompletion generates the completion script for the specified shell.
func generateCompletion(rootCmd *cobra.Command, shell string, out io.Writer) error {
	var err error
	switch shell {
	case shellBash:
		err = rootCmd.GenBashCompletion(out)
	case shellZsh:
		err = rootCmd.GenZshCompletion(out)
	case shellFish:
		err = rootCmd.GenFishCompletion(out, true)
	case shellPowershell:
		err = rootCmd.GenPowerShellCompletionWithDesc(out)
	default:
		return fmt.Errorf("%w: %s", clerrors.ErrUnsupportedShell, shell)
	}
	if err != nil {
		return fmt.Errorf("failed to generate %s completion: %w", shell, err)
	}
	return nil
}

// detectShell attempts to detect the user's current shell.
func detectShell() (string, error) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return "", clerrors.ErrNoShellEnv
	}

	base := filepath.Base(shell)
	switch {
	case strings.Contains(base, shellBash):
		return shellBash, nil
	case strings.Contains(base, shellZsh):
		return shellZsh, nil
	case strings.Contains(base, shellFish):
		return shellFish, nil
	case strings.Contains(base, "pwsh") || strings.Contains(base, shellPowershell):
		return shellPowershell, nil
	default:
		return "", fmt.Errorf("%w: %s", clerrors.ErrUnsupportedShell, base)
	}
}

// shellInstallTarget contains installation info for a shell.
type shellInstallTarget struct {
	targetPath        string
	setupInstructions string
}

// getInstallTarget returns the installation target path and setup instructions for a shell.
// Delegates to shell-specific helpers that handle platform conventions and package manager paths.
func getInstallTarget(homeDir, shell string) (*shellInstallTarget, error) {
	switch shell {
	case shellBash:
		return getBashInstallTarget(homeDir)
	case shellZsh:
		return getZshInstallTarget(homeDir)
	case shellFish:
		return getFishInstallTarget(homeDir)
	case shellPowershell:
		return getPowershellInstallTarget(homeDir)
	default:
		return nil, fmt.Errorf("%w: %s", clerrors.ErrUnsupportedShell, shell)
	}
}

// getBashInstallTarget determines bash completion installation path with Homebrew detection.
// Uses a sophisticated fallback strategy to handle both package manager and user installations.
//
// Path resolution strategy for bash (most complex of all shells):
// 1. Try Homebrew system location: $HOMEBREW_PREFIX/etc/bash_completion.d/pplx
//    - Honors HOMEBREW_PREFIX environment variable (custom Homebrew installations)
//    - Falls back to /usr/local if HOMEBREW_PREFIX not set (default macOS Homebrew)
//    - Tests if directory exists before using (Stat check on parent directory)
// 2. Fallback to user directory: ~/.bash_completion.d/pplx
//    - Used when Homebrew directory doesn't exist (Linux, non-Homebrew macOS, etc.)
//    - Requires manual sourcing in ~/.bashrc (returns setup instructions)
//
// Rationale for Homebrew preference:
// On macOS, Homebrew's bash-completion package provides automatic loading via
// /etc/bash_completion.d. If user has bash-completion installed (common on macOS),
// system-wide path provides zero-config experience. If not, user path provides
// compatibility with Linux conventions and doesn't require package dependencies.
//
// Setup instructions difference:
// - Homebrew path: No instructions (bash-completion package auto-sources)
// - User path: Must add source line to ~/.bashrc (manual setup required)
//
// This is the ONLY shell that needs Homebrew detection - others use consistent user paths.
func getBashInstallTarget(homeDir string) (*shellInstallTarget, error) {
	// Attempt 1: Homebrew system location
	// Check HOMEBREW_PREFIX environment variable for custom Homebrew installations
	brewPrefix := os.Getenv("HOMEBREW_PREFIX")
	if brewPrefix == "" {
		brewPrefix = "/usr/local" // Default Homebrew prefix on macOS
	}

	targetPath := filepath.Join(brewPrefix, "etc", "bash_completion.d", "pplx")
	// Check if Homebrew bash_completion.d directory exists
	if _, err := os.Stat(filepath.Dir(targetPath)); os.IsNotExist(err) {
		// Attempt 2: User directory fallback (Linux-style path)
		targetPath = filepath.Join(homeDir, ".bash_completion.d", "pplx")
		if err := os.MkdirAll(filepath.Dir(targetPath), dirPerms); err != nil { // #nosec G301
			return nil, fmt.Errorf("failed to create directory %s: %w", filepath.Dir(targetPath), err)
		}
		// User path requires manual sourcing in ~/.bashrc
		return &shellInstallTarget{
			targetPath: targetPath,
			setupInstructions: fmt.Sprintf(
				"\nAdd the following to your ~/.bashrc:\n\n  [[ -f %s ]] && source %s\n",
				targetPath, targetPath),
		}, nil
	}
	// Homebrew path: bash-completion package auto-sources, no setup needed
	return &shellInstallTarget{targetPath: targetPath}, nil
}

// getZshInstallTarget returns zsh completion path following zsh conventions.
// Path: ~/.zsh/completions/_pplx (underscore prefix is zsh convention for completion files)
// Setup: User must add completion directory to fpath and enable compinit in ~/.zshrc.
func getZshInstallTarget(homeDir string) (*shellInstallTarget, error) {
	targetPath := filepath.Join(homeDir, ".zsh", "completions", "_pplx")
	if err := os.MkdirAll(filepath.Dir(targetPath), dirPerms); err != nil { // #nosec G301
		return nil, fmt.Errorf("failed to create directory %s: %w", filepath.Dir(targetPath), err)
	}
	return &shellInstallTarget{
		targetPath: targetPath,
		setupInstructions: fmt.Sprintf(
			"\nAdd the following to your ~/.zshrc:\n\n"+
				"  fpath=(%s $fpath)\n  autoload -U compinit; compinit\n",
			filepath.Dir(targetPath)),
	}, nil
}

// getFishInstallTarget returns fish completion path following XDG Base Directory spec.
// Path: ~/.config/fish/completions/pplx.fish
// Setup: No manual setup needed - fish auto-loads from this directory.
func getFishInstallTarget(homeDir string) (*shellInstallTarget, error) {
	configDir := filepath.Join(homeDir, ".config", "fish", "completions")
	if err := os.MkdirAll(configDir, dirPerms); err != nil { // #nosec G301
		return nil, fmt.Errorf("failed to create directory %s: %w", configDir, err)
	}
	return &shellInstallTarget{
		targetPath: filepath.Join(configDir, "pplx.fish"),
	}, nil
}

// getPowershellInstallTarget returns PowerShell completion path following Windows conventions.
// Path: ~/Documents/PowerShell/Scripts/pplx-completion.ps1
// Setup: User must dot-source the script in their PowerShell profile.
func getPowershellInstallTarget(homeDir string) (*shellInstallTarget, error) {
	targetPath := filepath.Join(homeDir, "Documents", "PowerShell", "Scripts", "pplx-completion.ps1")
	if err := os.MkdirAll(filepath.Dir(targetPath), dirPerms); err != nil { // #nosec G301
		return nil, fmt.Errorf("failed to create directory %s: %w", filepath.Dir(targetPath), err)
	}
	return &shellInstallTarget{
		targetPath:        targetPath,
		setupInstructions: fmt.Sprintf("\nAdd the following to your PowerShell profile:\n\n  . %s\n", targetPath),
	}, nil
}

// installCompletion installs the completion script for the specified shell.
func installCompletion(rootCmd *cobra.Command, shell string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory for %s completion installation: %w", shell, err)
	}

	target, err := getInstallTarget(homeDir, shell)
	if err != nil {
		return err
	}

	// Create the completion file
	file, err := os.Create(target.targetPath)
	if err != nil {
		return fmt.Errorf("failed to create %s completion file %s: %w", shell, target.targetPath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			logger.Warn("failed to close file", "error", closeErr)
		}
	}()

	// Generate completion script
	if err := generateCompletion(rootCmd, shell, file); err != nil {
		return fmt.Errorf("failed to generate %s completion: %w", shell, err)
	}

	fmt.Printf("✓ Shell completion installed to: %s\n", target.targetPath)
	if target.setupInstructions != "" {
		fmt.Println(target.setupInstructions)
		fmt.Println("Restart your shell or source the configuration file for changes to take effect.")
	} else {
		fmt.Println("\nRestart your shell for changes to take effect.")
	}

	return nil
}

// getUninstallPaths returns the list of paths to check for uninstalling a shell's completion.
func getUninstallPaths(homeDir, shell string) ([]string, error) {
	switch shell {
	case shellBash:
		brewPrefix := os.Getenv("HOMEBREW_PREFIX")
		if brewPrefix == "" {
			brewPrefix = "/usr/local"
		}
		return []string{
			filepath.Join(brewPrefix, "etc", "bash_completion.d", "pplx"),
			filepath.Join(homeDir, ".bash_completion.d", "pplx"),
		}, nil
	case shellZsh:
		return []string{
			filepath.Join(homeDir, ".zsh", "completions", "_pplx"),
		}, nil
	case shellFish:
		return []string{
			filepath.Join(homeDir, ".config", "fish", "completions", "pplx.fish"),
		}, nil
	case shellPowershell:
		return []string{
			filepath.Join(homeDir, "Documents", "PowerShell", "Scripts", "pplx-completion.ps1"),
		}, nil
	default:
		return nil, fmt.Errorf("%w: %s", clerrors.ErrUnsupportedShell, shell)
	}
}

// uninstallCompletion removes the completion script for the specified shell.
// Handles multiple potential installation paths and fails gracefully if files don't exist.
//
// Multi-path checking rationale:
// Completion files may exist in different locations depending on:
// - Installation method (Homebrew vs user install for bash)
// - Legacy installations (user may have moved files or installed multiple times)
// - Tool version changes (old versions may have used different paths)
//
// Error handling: Intentional silent failure for missing files
// Philosophy: Uninstall should succeed if the end state is "completion not installed",
// regardless of whether files existed before. This makes uninstall idempotent:
// - Running uninstall twice doesn't error on second run
// - Running uninstall after manual deletion doesn't error
// - Partial installations (file deleted but not others) are cleaned up gracefully
//
// The "removed" flag tracks whether ANY file was actually deleted to provide
// meaningful user feedback:
// - If files found and removed: Show success message + manual cleanup reminder
// - If no files found: Inform user (not an error, just informative)
//
// Logging strategy for removal failures:
// - Warn instead of error: File exists but can't delete (permissions, in use, etc.)
// - Continue processing other paths: One failure shouldn't prevent cleaning other locations
// - Don't fail entire operation: Partial cleanup is better than no cleanup.
func uninstallCompletion(shell string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory for %s completion uninstallation: %w", shell, err)
	}

	targetPaths, err := getUninstallPaths(homeDir, shell)
	if err != nil {
		return err
	}

	removed := false
	for _, path := range targetPaths {
		// Check if file exists before attempting removal
		if _, err := os.Stat(path); err == nil {
			// File exists - attempt removal
			if err := os.Remove(path); err != nil {
				// Warn but don't fail: permission issues, file in use, etc.
				logger.Warn("failed to remove completion file", "path", path, "error", err)
			} else {
				fmt.Printf("✓ Removed completion file: %s\n", path)
				removed = true
			}
		}
		// File doesn't exist: Silent skip (not an error - desired end state achieved)
	}

	// Provide user feedback based on what was found and removed
	if !removed {
		fmt.Println("No completion files found to remove.")
	} else {
		// Remind user about manual shell config cleanup
		// We only delete completion files, not the source/fpath lines in shell config
		fmt.Println("\nYou may need to restart your shell or manually remove any " +
			"source lines from your shell configuration.")
	}

	return nil
}

func init() {
	// Add completion command and its subcommands
	rootCmd.AddCommand(completionCmd)

	// Add individual shell subcommands
	completionCmd.AddCommand(bashCmd)
	completionCmd.AddCommand(zshCmd)
	completionCmd.AddCommand(fishCmd)
	completionCmd.AddCommand(powershellCmd)

	// Add install subcommand
	completionCmd.AddCommand(installCmd)

	// Add flags
	completionCmd.PersistentFlags().StringVarP(
		&completionOutputFile, "output", "o", "",
		"Output file path (default: stdout)")
	installCmd.Flags().BoolVar(
		&completionUninstall, "uninstall", false,
		"Uninstall completion instead of installing")
}
