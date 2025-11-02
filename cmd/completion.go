// Package cmd provides command-line interface commands for the Perplexity API.
package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	completionShell      string
	completionInstall    bool
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
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		shell := args[0]

		// Determine output writer
		var out io.Writer = os.Stdout
		if completionOutputFile != "" {
			file, err := os.Create(completionOutputFile)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer file.Close()
			out = file
		}

		// Generate completion script
		return generateCompletion(cmd.Root(), shell, out)
	},
}

// bashCmd represents the bash completion subcommand.
var bashCmd = &cobra.Command{
	Use:   "bash",
	Short: "Generate bash completion script",
	Long:  "Generate the autocompletion script for bash",
	RunE: func(cmd *cobra.Command, _ []string) error {
		var out io.Writer = os.Stdout
		if completionOutputFile != "" {
			file, err := os.Create(completionOutputFile)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer file.Close()
			out = file
		}
		return cmd.Root().GenBashCompletion(out)
	},
}

// zshCmd represents the zsh completion subcommand.
var zshCmd = &cobra.Command{
	Use:   "zsh",
	Short: "Generate zsh completion script",
	Long:  "Generate the autocompletion script for zsh",
	RunE: func(cmd *cobra.Command, _ []string) error {
		var out io.Writer = os.Stdout
		if completionOutputFile != "" {
			file, err := os.Create(completionOutputFile)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer file.Close()
			out = file
		}
		return cmd.Root().GenZshCompletion(out)
	},
}

// fishCmd represents the fish completion subcommand.
var fishCmd = &cobra.Command{
	Use:   "fish",
	Short: "Generate fish completion script",
	Long:  "Generate the autocompletion script for fish",
	RunE: func(cmd *cobra.Command, _ []string) error {
		var out io.Writer = os.Stdout
		if completionOutputFile != "" {
			file, err := os.Create(completionOutputFile)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer file.Close()
			out = file
		}
		return cmd.Root().GenFishCompletion(out, true)
	},
}

// powershellCmd represents the powershell completion subcommand.
var powershellCmd = &cobra.Command{
	Use:   "powershell",
	Short: "Generate PowerShell completion script",
	Long:  "Generate the autocompletion script for PowerShell",
	RunE: func(cmd *cobra.Command, _ []string) error {
		var out io.Writer = os.Stdout
		if completionOutputFile != "" {
			file, err := os.Create(completionOutputFile)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer file.Close()
			out = file
		}
		return cmd.Root().GenPowerShellCompletionWithDesc(out)
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
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
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
	switch shell {
	case "bash":
		return rootCmd.GenBashCompletion(out)
	case "zsh":
		return rootCmd.GenZshCompletion(out)
	case "fish":
		return rootCmd.GenFishCompletion(out, true)
	case "powershell":
		return rootCmd.GenPowerShellCompletionWithDesc(out)
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}
}

// detectShell attempts to detect the user's current shell.
func detectShell() (string, error) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return "", fmt.Errorf("SHELL environment variable not set")
	}

	base := filepath.Base(shell)
	switch {
	case strings.Contains(base, "bash"):
		return "bash", nil
	case strings.Contains(base, "zsh"):
		return "zsh", nil
	case strings.Contains(base, "fish"):
		return "fish", nil
	case strings.Contains(base, "pwsh") || strings.Contains(base, "powershell"):
		return "powershell", nil
	default:
		return "", fmt.Errorf("unsupported shell: %s", base)
	}
}

// installCompletion installs the completion script for the specified shell.
func installCompletion(rootCmd *cobra.Command, shell string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	var targetPath string
	var setupInstructions string

	switch shell {
	case "bash":
		// Try macOS Homebrew location first, then Linux
		brewPrefix := os.Getenv("HOMEBREW_PREFIX")
		if brewPrefix == "" {
			brewPrefix = "/usr/local" // Default Homebrew prefix
		}

		targetPath = filepath.Join(brewPrefix, "etc", "bash_completion.d", "pplx")
		if _, err := os.Stat(filepath.Dir(targetPath)); os.IsNotExist(err) {
			// Fallback to user directory
			targetPath = filepath.Join(homeDir, ".bash_completion.d", "pplx")
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
			setupInstructions = fmt.Sprintf("\nAdd the following to your ~/.bashrc:\n\n  [[ -f %s ]] && source %s\n", targetPath, targetPath)
		}

	case "zsh":
		// Get first element of fpath or use default
		targetPath = filepath.Join(homeDir, ".zsh", "completions", "_pplx")
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
		setupInstructions = fmt.Sprintf("\nAdd the following to your ~/.zshrc:\n\n  fpath=(%s $fpath)\n  autoload -U compinit; compinit\n", filepath.Dir(targetPath))

	case "fish":
		configDir := filepath.Join(homeDir, ".config", "fish", "completions")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
		targetPath = filepath.Join(configDir, "pplx.fish")

	case "powershell":
		// PowerShell profile location varies by platform
		targetPath = filepath.Join(homeDir, "Documents", "PowerShell", "Scripts", "pplx-completion.ps1")
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
		setupInstructions = fmt.Sprintf("\nAdd the following to your PowerShell profile:\n\n  . %s\n", targetPath)

	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}

	// Create the completion file
	file, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create completion file: %w", err)
	}
	defer file.Close()

	// Generate completion script
	if err := generateCompletion(rootCmd, shell, file); err != nil {
		return fmt.Errorf("failed to generate completion: %w", err)
	}

	fmt.Printf("✓ Shell completion installed to: %s\n", targetPath)
	if setupInstructions != "" {
		fmt.Println(setupInstructions)
		fmt.Println("Restart your shell or source the configuration file for changes to take effect.")
	} else {
		fmt.Println("\nRestart your shell for changes to take effect.")
	}

	return nil
}

// uninstallCompletion removes the completion script for the specified shell.
func uninstallCompletion(shell string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	var targetPaths []string

	switch shell {
	case "bash":
		brewPrefix := os.Getenv("HOMEBREW_PREFIX")
		if brewPrefix == "" {
			brewPrefix = "/usr/local"
		}
		targetPaths = []string{
			filepath.Join(brewPrefix, "etc", "bash_completion.d", "pplx"),
			filepath.Join(homeDir, ".bash_completion.d", "pplx"),
		}

	case "zsh":
		targetPaths = []string{
			filepath.Join(homeDir, ".zsh", "completions", "_pplx"),
		}

	case "fish":
		targetPaths = []string{
			filepath.Join(homeDir, ".config", "fish", "completions", "pplx.fish"),
		}

	case "powershell":
		targetPaths = []string{
			filepath.Join(homeDir, "Documents", "PowerShell", "Scripts", "pplx-completion.ps1"),
		}

	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}

	removed := false
	for _, path := range targetPaths {
		if _, err := os.Stat(path); err == nil {
			if err := os.Remove(path); err != nil {
				fmt.Printf("Warning: failed to remove %s: %v\n", path, err)
			} else {
				fmt.Printf("✓ Removed completion file: %s\n", path)
				removed = true
			}
		}
	}

	if !removed {
		fmt.Println("No completion files found to remove.")
	} else {
		fmt.Println("\nYou may need to restart your shell or manually remove any source lines from your shell configuration.")
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
	completionCmd.PersistentFlags().StringVarP(&completionOutputFile, "output", "o", "", "Output file path (default: stdout)")
	installCmd.Flags().BoolVar(&completionUninstall, "uninstall", false, "Uninstall completion instead of installing")
}
