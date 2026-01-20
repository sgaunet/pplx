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

func getBashInstallTarget(homeDir string) (*shellInstallTarget, error) {
	// Try macOS Homebrew location first, then Linux
	brewPrefix := os.Getenv("HOMEBREW_PREFIX")
	if brewPrefix == "" {
		brewPrefix = "/usr/local" // Default Homebrew prefix
	}

	targetPath := filepath.Join(brewPrefix, "etc", "bash_completion.d", "pplx")
	if _, err := os.Stat(filepath.Dir(targetPath)); os.IsNotExist(err) {
		// Fallback to user directory
		targetPath = filepath.Join(homeDir, ".bash_completion.d", "pplx")
		if err := os.MkdirAll(filepath.Dir(targetPath), dirPerms); err != nil { // #nosec G301
			return nil, fmt.Errorf("failed to create directory %s: %w", filepath.Dir(targetPath), err)
		}
		return &shellInstallTarget{
			targetPath: targetPath,
			setupInstructions: fmt.Sprintf(
				"\nAdd the following to your ~/.bashrc:\n\n  [[ -f %s ]] && source %s\n",
				targetPath, targetPath),
		}, nil
	}
	return &shellInstallTarget{targetPath: targetPath}, nil
}

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

func getFishInstallTarget(homeDir string) (*shellInstallTarget, error) {
	configDir := filepath.Join(homeDir, ".config", "fish", "completions")
	if err := os.MkdirAll(configDir, dirPerms); err != nil { // #nosec G301
		return nil, fmt.Errorf("failed to create directory %s: %w", configDir, err)
	}
	return &shellInstallTarget{
		targetPath: filepath.Join(configDir, "pplx.fish"),
	}, nil
}

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
		if _, err := os.Stat(path); err == nil {
			if err := os.Remove(path); err != nil {
				logger.Warn("failed to remove completion file", "path", path, "error", err)
			} else {
				fmt.Printf("✓ Removed completion file: %s\n", path)
				removed = true
			}
		}
	}

	if !removed {
		fmt.Println("No completion files found to remove.")
	} else {
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
