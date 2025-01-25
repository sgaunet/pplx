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
	Short: "Program to interact with the Perplexity API",
	Long: `Program to interact with the Perplexity API.
	
	You can use it to chat with the AI or to query it.`,
}

func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(chatCmd)
	chatCmd.PersistentFlags().StringVarP(&model, "model", "m", "basic", "Online model to use: basic pro")

	rootCmd.AddCommand(queryCmd)
	queryCmd.PersistentFlags().StringVarP(&model, "model", "m", "basic", "Online model to use: basic pro")
	queryCmd.PersistentFlags().StringVarP(&systemPrompt, "sys-prompt", "s", "", "system prompt")
	queryCmd.PersistentFlags().StringVarP(&userPrompt, "user-prompt", "p", "", "user prompt")

	// Check env var PPLX_API_KEY exists
	if os.Getenv("PPLX_API_KEY") == "" {
		fmt.Fprintf(os.Stderr, "Error: PPLX_API_KEY env var is not set\n")
		os.Exit(1)
	}
	apiKey = os.Getenv("PPLX_API_KEY")
}
