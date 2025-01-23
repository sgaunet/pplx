package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var (
	model        string
	systemPrompt string
	userPrompt   string
	apiKey       string
)

const DefaultTimeout = 30 * time.Second

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pplx",
	Short: "",
	Long:  ``,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(chatCmd)
	chatCmd.PersistentFlags().StringVar(&model, "m", "small", "Online model to use: small, large, huge")

	rootCmd.AddCommand(queryCmd)
	queryCmd.PersistentFlags().StringVar(&model, "m", "small", "Online model to use: small, large, huge")
	queryCmd.PersistentFlags().StringVar(&systemPrompt, "sys", "", "system prompt")
	queryCmd.PersistentFlags().StringVar(&userPrompt, "u", "", "user prompt")

	// Check env var PPLX_API_KEY exists
	if os.Getenv("PPLX_API_KEY") == "" {
		fmt.Fprintf(os.Stderr, "Error: PPLX_API_KEY env var is not set\n")
		os.Exit(1)
	}
	apiKey = os.Getenv("PPLX_API_KEY")
}
