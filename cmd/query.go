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
		if model == "pro" {
			model = perplexity.ProModel
		} else {
			model = perplexity.DefaultModel
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

		req := perplexity.NewCompletionRequest(perplexity.WithMessages(msg.GetMessages()), perplexity.WithModel(model))
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
