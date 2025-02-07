package cmd

import (
	"fmt"
	"os"

	"github.com/pterm/pterm"
	"github.com/sgaunet/perplexity-go/v2"
	"github.com/sgaunet/pplx/pkg/console"
	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// Check env var PPLX_API_KEY exists
		if os.Getenv("PPLX_API_KEY") == "" {
			fmt.Fprintf(os.Stderr, "Error: PPLX_API_KEY env var is not set\n")
			os.Exit(1)
		}

		client := perplexity.NewClient(os.Getenv("PPLX_API_KEY"))
		client.SetHTTPTimeout(DefaultTimeout)

		if userPrompt == "" {
			fmt.Println("Error: user prompt is required")
			cmd.Usage()
			os.Exit(1)
		}
		msg := perplexity.NewMessages(perplexity.WithSystemMessage(systemPrompt))
		msg.AddUserMessage(userPrompt)

		req := perplexity.NewCompletionRequest(perplexity.WithMessages(msg.GetMessages()),
			perplexity.WithModel(model),
			perplexity.WithFrequencyPenalty(frequencyPenalty),
			perplexity.WithMaxTokens(maxTokens),
			perplexity.WithPresencePenalty(presencePenalty),
			perplexity.WithTemperature(temperature),
			perplexity.WithTopK(topK),
			perplexity.WithTopP(topP),
		)
		err := req.Validate()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		spinnerInfo, _ := pterm.DefaultSpinner.Start("Waiting after the response from perplexity...")
		res, err := client.SendCompletionRequest(req)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		spinnerInfo.Success("Response received")
		err = console.RenderAsMarkdown(res, os.Stdout)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		err = console.RenderCitations(res, os.Stdout)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	},
}
