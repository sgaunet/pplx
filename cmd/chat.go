// Package cmd provides command-line interface commands for the Perplexity API.
package cmd

import (
	"fmt"
	"os"

	"github.com/pterm/pterm"
	"github.com/sgaunet/perplexity-go/v2"
	"github.com/sgaunet/pplx/pkg/chat"
	"github.com/sgaunet/pplx/pkg/console"
	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "chat subcommand is an interactive chat with the Perplexity API",
	Long: `With chat subcommand you can interactively chat with the Perplexity API.
You can ask questions and get answers from the API. As long as you don't enter an empty question,
 the chat will continue.`,
	Run: func(_ *cobra.Command, _ []string) {
		// Check env var PPLX_API_KEY exists
		if os.Getenv("PPLX_API_KEY") == "" {
			fmt.Fprintf(os.Stderr, "Error: PPLX_API_KEY env var is not set\n")
			os.Exit(1)
		}

		client := perplexity.NewClient(os.Getenv("PPLX_API_KEY"))
		client.SetHTTPTimeout(timeout)

		systemMessage, err := console.Input("system message (optional - enter to skip)")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading system message: %v\n", err)
			os.Exit(1)
		}
		// Create chat options
		chatOptions := chat.Options{
			Model:            model,
			FrequencyPenalty: frequencyPenalty,
			MaxTokens:        maxTokens,
			PresencePenalty:  presencePenalty,
			Temperature:      temperature,
			TopK:             topK,
			TopP:             topP,
			SearchDomains:    searchDomains,
			SearchRecency:    searchRecency,
			LocationLat:      locationLat,
			LocationLon:      locationLon,
			LocationCountry:  locationCountry,
			ReturnImages:     returnImages,
			ReturnRelated:    returnRelated,
			Stream:           stream,
			ImageDomains:     imageDomains,
			ImageFormats:     imageFormats,
			// Response format options
			ResponseFormatJSONSchema: responseFormatJSONSchema,
			ResponseFormatRegex:      responseFormatRegex,
			// Search mode options
			SearchMode:        searchMode,
			SearchContextSize: searchContextSize,
			// Date filtering options
			SearchAfterDate:   searchAfterDate,
			SearchBeforeDate:  searchBeforeDate,
			LastUpdatedAfter:  lastUpdatedAfter,
			LastUpdatedBefore: lastUpdatedBefore,
			// Deep research options
			ReasoningEffort: reasoningEffort,
		}
		c := chat.NewChatWithOptions(client, systemMessage, chatOptions)

		// discussion loop
	loop:
		for {
			prompt, err := console.Input("Ask anything (enter to quit)")
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading prompt: %v\n", err)
				os.Exit(1)
			}
			if prompt == "" {
				break loop
			}
			err = c.AddUserMessage(prompt)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error adding user message: %v\n", err)
				os.Exit(1)
			}
			// Print spinner while waiting for the response
			spinnerInfo, _ := pterm.DefaultSpinner.Start("Waiting after the response from perplexity...")
			response, err := c.Run()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error running chat: %v\n", err)
				os.Exit(1)
			}
			spinnerInfo.Success("Response received")

			err = c.AddAgentMessage(response.GetLastContent())
			if err != nil {
				fmt.Fprintf(os.Stderr, "error adding agent message: %v\n", err)
				os.Exit(1)
			}
			err = console.RenderAsMarkdown(response, os.Stdout)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			err = console.RenderCitations(response, os.Stdout)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			err = console.RenderImages(response, os.Stdout)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			err = console.RenderRelatedQuestions(response, os.Stdout)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		}
	},
}
