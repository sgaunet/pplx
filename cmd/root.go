package cmd

import (
	"os"

	"github.com/sgaunet/perplexity-go/v2"
	"github.com/spf13/cobra"
)

var (
	systemPrompt     string
	userPrompt       string
	model            string  = perplexity.DefaultModel
	frequencyPenalty float64 = perplexity.DefaultFrequencyPenalty
	maxTokens        int     = perplexity.DefaultMaxTokens
	presencePenalty  float64 = perplexity.DefaultPresencePenalty
	temperature      float64 = perplexity.DefaultTemperature
	topK             int     = perplexity.DefaultTopK
	topP             float64 = perplexity.DefaultTopP
	timeout                  = perplexity.DefaultTimeout
)

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
	chatCmd.PersistentFlags().StringVarP(&model, "model", "m", perplexity.DefaultModel, "List of models: https://docs.perplexity.ai/guides/model-cards")

	chatCmd.PersistentFlags().Float64Var(&frequencyPenalty, "frequency-penalty", frequencyPenalty, "Frequency penalty")
	chatCmd.PersistentFlags().IntVar(&maxTokens, "max-tokens", maxTokens, "Max tokens")
	chatCmd.PersistentFlags().Float64Var(&presencePenalty, "presence-penalty", presencePenalty, "Presence penalty")
	chatCmd.PersistentFlags().Float64Var(&temperature, "temperature", temperature, "Temperature")
	chatCmd.PersistentFlags().IntVar(&topK, "top-k", topK, "Top K")
	chatCmd.PersistentFlags().Float64Var(&topP, "top-p", topP, "Top P")
	chatCmd.PersistentFlags().DurationVar(&timeout, "timeout", timeout, "HTTP timeout")

	rootCmd.AddCommand(queryCmd)
	queryCmd.PersistentFlags().StringVarP(&model, "model", "m", perplexity.DefaultModel, "List of models: https://docs.perplexity.ai/guides/model-cards")
	queryCmd.PersistentFlags().StringVarP(&systemPrompt, "sys-prompt", "s", "", "system prompt")
	queryCmd.PersistentFlags().StringVarP(&userPrompt, "user-prompt", "p", "", "user prompt")

	queryCmd.PersistentFlags().Float64Var(&frequencyPenalty, "frequency-penalty", frequencyPenalty, "Frequency penalty")
	queryCmd.PersistentFlags().IntVar(&maxTokens, "max-tokens", maxTokens, "Max tokens")
	queryCmd.PersistentFlags().Float64Var(&presencePenalty, "presence-penalty", presencePenalty, "Presence penalty")
	queryCmd.PersistentFlags().Float64Var(&temperature, "temperature", temperature, "Temperature")
	queryCmd.PersistentFlags().IntVar(&topK, "top-k", topK, "Top K")
	queryCmd.PersistentFlags().Float64Var(&topP, "top-p", topP, "Top P")
	queryCmd.PersistentFlags().DurationVar(&timeout, "timeout", timeout, "HTTP timeout")
}
