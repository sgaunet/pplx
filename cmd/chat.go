package cmd

import (
	"os"

	"github.com/pterm/pterm"
	"github.com/sgaunet/perplexity-go/v2"
	"github.com/sgaunet/pplx/pkg/chat"
	"github.com/sgaunet/pplx/pkg/config"
	"github.com/sgaunet/pplx/pkg/console"
	clerrors "github.com/sgaunet/pplx/pkg/clerrors"
	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "chat subcommand is an interactive chat with the Perplexity API",
	Long: `With chat subcommand you can interactively chat with the Perplexity API.
You can ask questions and get answers from the API. As long as you don't enter an empty question,
 the chat will continue.`,
	RunE: func(cmd *cobra.Command, _ []string) error {
		// Load configuration from file and merge with CLI flags
		cfg, err := config.LoadAndMergeConfig(cmd, configFilePath)
		if err != nil {
			// Non-fatal: continue with CLI flags only
			cfg = config.NewConfigData()
		}

		// Apply configuration to global variables
		config.ApplyToGlobals(cfg,
			&model, &temperature, &maxTokens, &topK, &topP,
			&frequencyPenalty, &presencePenalty, &timeout,
			&searchDomains, &searchRecency, &locationLat, &locationLon, &locationCountry,
			&returnImages, &returnRelated, &stream,
			&searchMode, &searchContextSize,
		)

		// Check env var PPLX_API_KEY exists
		if os.Getenv("PPLX_API_KEY") == "" {
			return clerrors.NewConfigError("PPLX_API_KEY environment variable is not set", nil)
		}

		client := perplexity.NewClient(os.Getenv("PPLX_API_KEY"))
		client.SetHTTPTimeout(timeout)

		systemMessage, err := console.Input("system message (optional - enter to skip)")
		if err != nil {
			return clerrors.NewIOError("failed to read system message", err)
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
				return clerrors.NewIOError("failed to read prompt", err)
			}
			if prompt == "" {
				break loop
			}
			err = c.AddUserMessage(prompt)
			if err != nil {
				return clerrors.NewAPIError("failed to add user message", err)
			}
			// Print spinner while waiting for the response
			spinnerInfo, _ := pterm.DefaultSpinner.Start("Waiting after the response from perplexity...")
			response, err := c.Run()
			if err != nil {
				return clerrors.NewAPIError("failed to run chat", err)
			}
			spinnerInfo.Success("Response received")

			err = c.AddAgentMessage(response.GetLastContent())
			if err != nil {
				return clerrors.NewAPIError("failed to add agent message", err)
			}
			err = console.RenderAsMarkdown(response, os.Stdout)
			if err != nil {
				return clerrors.NewIOError("failed to render markdown", err)
			}
			err = console.RenderCitations(response, os.Stdout)
			if err != nil {
				return clerrors.NewIOError("failed to render citations", err)
			}
			err = console.RenderImages(response, os.Stdout)
			if err != nil {
				return clerrors.NewIOError("failed to render images", err)
			}
			err = console.RenderRelatedQuestions(response, os.Stdout)
			if err != nil {
				return clerrors.NewIOError("failed to render related questions", err)
			}
		}
		return nil
	},
}
