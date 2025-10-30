package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pterm/pterm"
	"github.com/sgaunet/perplexity-go/v2"
	"github.com/sgaunet/pplx/pkg/config"
	"github.com/sgaunet/pplx/pkg/console"
	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, _ []string) {
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
			fmt.Fprintf(os.Stderr, "Error: PPLX_API_KEY env var is not set\n")
			os.Exit(1)
		}

		client := perplexity.NewClient(os.Getenv("PPLX_API_KEY"))
		client.SetHTTPTimeout(timeout)

		if userPrompt == "" {
			fmt.Println("Error: user prompt is required")
			_ = cmd.Usage()
			os.Exit(1)
		}
		msg := perplexity.NewMessages(perplexity.WithSystemMessage(systemPrompt))
		_ = msg.AddUserMessage(userPrompt)

		// Build options list
		opts := []perplexity.CompletionRequestOption{
			perplexity.WithMessages(msg.GetMessages()),
			perplexity.WithModel(model),
			perplexity.WithFrequencyPenalty(frequencyPenalty),
			perplexity.WithMaxTokens(maxTokens),
			perplexity.WithPresencePenalty(presencePenalty),
			perplexity.WithTemperature(temperature),
			perplexity.WithTopK(topK),
			perplexity.WithTopP(topP),
		}

		// Add search/web options if provided
		if len(searchDomains) > 0 {
			opts = append(opts, perplexity.WithSearchDomainFilter(searchDomains))
		}
		if searchRecency != "" {
			// Validate search recency
			validRecency := map[string]bool{"day": true, "week": true, "month": true, "year": true, "hour": true}
			if !validRecency[searchRecency] {
				fmt.Printf("Error: Invalid search-recency value '%s'. Must be one of: day, week, month, year, hour\n",
					searchRecency)
				os.Exit(1)
			}
			// Search recency filter is incompatible with images
			if returnImages {
				fmt.Printf("Warning: search-recency filter is incompatible with images, ignoring search-recency\n")
			} else {
				opts = append(opts, perplexity.WithSearchRecencyFilter(searchRecency))
			}
		}
		if locationLat != 0 || locationLon != 0 || locationCountry != "" {
			opts = append(opts, perplexity.WithUserLocation(locationLat, locationLon, locationCountry))
		}

		// Add response enhancement options
		if returnImages {
			opts = append(opts, perplexity.WithReturnImages(returnImages))
			// When images are requested, explicitly disable search recency filter
			// to avoid API incompatibility issues
			opts = append(opts, perplexity.WithSearchRecencyFilter(""))
		}
		if returnRelated {
			opts = append(opts, perplexity.WithReturnRelatedQuestions(returnRelated))
		}
		if stream {
			opts = append(opts, perplexity.WithStream(stream))
		}

		// Add image filtering options
		if len(imageDomains) > 0 {
			opts = append(opts, perplexity.WithImageDomainFilter(imageDomains))
		}
		if len(imageFormats) > 0 {
			// Validate image formats
			validFormats := map[string]bool{
				"jpg": true, "jpeg": true, "png": true, "gif": true,
				"webp": true, "svg": true, "bmp": true,
			}
			for _, format := range imageFormats {
				if !validFormats[format] {
					warnMsg := "Warning: Image format '%s' may not be supported. " +
						"Common formats are: jpg, jpeg, png, gif, webp, svg, bmp\n"
					fmt.Printf(warnMsg, format)
				}
			}
			opts = append(opts, perplexity.WithImageFormatFilter(imageFormats))
		}

		// Add response format options
		if responseFormatJSONSchema != "" && responseFormatRegex != "" {
			fmt.Println("Error: Cannot use both --response-format-json-schema and --response-format-regex")
			os.Exit(1)
		}
		if responseFormatJSONSchema != "" || responseFormatRegex != "" {
			// Validate model supports response formats
			if !strings.HasPrefix(model, "sonar") {
				fmt.Printf("Error: Response formats (JSON schema and regex) are only supported by sonar models\n")
				os.Exit(1)
			}
		}
		if responseFormatJSONSchema != "" {
			// Parse JSON schema
			var schema interface{}
			err := json.Unmarshal([]byte(responseFormatJSONSchema), &schema)
			if err != nil {
				fmt.Printf("Error: Invalid JSON schema: %v\n", err)
				os.Exit(1)
			}
			opts = append(opts, perplexity.WithJSONSchemaResponseFormat(schema))
		}
		if responseFormatRegex != "" {
			opts = append(opts, perplexity.WithRegexResponseFormat(responseFormatRegex))
		}

		// Add search mode options
		if searchMode != "" {
			// Validate search mode
			validModes := map[string]bool{"web": true, "academic": true}
			if !validModes[searchMode] {
				fmt.Printf("Error: Invalid search mode '%s'. Must be one of: web, academic\n", searchMode)
				os.Exit(1)
			}
			opts = append(opts, perplexity.WithSearchMode(searchMode))
		}
		if searchContextSize != "" {
			// Validate search context size
			validSizes := map[string]bool{"low": true, "medium": true, "high": true}
			if !validSizes[searchContextSize] {
				fmt.Printf("Error: Invalid search context size '%s'. Must be one of: low, medium, high\n", searchContextSize)
				os.Exit(1)
			}
			opts = append(opts, perplexity.WithSearchContextSize(searchContextSize))
		}

		// Add date filtering options
		if searchAfterDate != "" {
			date, err := time.Parse("01/02/2006", searchAfterDate)
			if err != nil {
				fmt.Printf("Error: Invalid search-after-date format '%s'. Use MM/DD/YYYY\n", searchAfterDate)
				os.Exit(1)
			}
			opts = append(opts, perplexity.WithSearchAfterDateFilter(date))
		}
		if searchBeforeDate != "" {
			date, err := time.Parse("01/02/2006", searchBeforeDate)
			if err != nil {
				fmt.Printf("Error: Invalid search-before-date format '%s'. Use MM/DD/YYYY\n", searchBeforeDate)
				os.Exit(1)
			}
			opts = append(opts, perplexity.WithSearchBeforeDateFilter(date))
		}
		if lastUpdatedAfter != "" {
			date, err := time.Parse("01/02/2006", lastUpdatedAfter)
			if err != nil {
				fmt.Printf("Error: Invalid last-updated-after format '%s'. Use MM/DD/YYYY\n", lastUpdatedAfter)
				os.Exit(1)
			}
			opts = append(opts, perplexity.WithLastUpdatedAfterFilter(date))
		}
		if lastUpdatedBefore != "" {
			date, err := time.Parse("01/02/2006", lastUpdatedBefore)
			if err != nil {
				fmt.Printf("Error: Invalid last-updated-before format '%s'. Use MM/DD/YYYY\n", lastUpdatedBefore)
				os.Exit(1)
			}
			opts = append(opts, perplexity.WithLastUpdatedBeforeFilter(date))
		}

		// Add deep research options
		if reasoningEffort != "" {
			// Validate reasoning effort
			validEfforts := map[string]bool{"low": true, "medium": true, "high": true}
			if !validEfforts[reasoningEffort] {
				fmt.Printf("Error: Invalid reasoning effort '%s'. Must be one of: low, medium, high\n", reasoningEffort)
				os.Exit(1)
			}
			// Check if the model supports reasoning effort
			if !strings.Contains(model, "deep-research") {
				fmt.Printf("Warning: reasoning-effort is only supported by sonar-deep-research model\n")
			}
			opts = append(opts, perplexity.WithReasoningEffort(reasoningEffort))
		}

		req := perplexity.NewCompletionRequest(opts...)
		err = req.Validate()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if stream {
			// Handle streaming response
			responseChannel := make(chan perplexity.CompletionResponse)
			var wg sync.WaitGroup
			wg.Add(1)

			// Start goroutine to handle streaming responses
			go func() {
				defer wg.Done()
				var lastResponse *perplexity.CompletionResponse

				if outputJSON {
					// For JSON output, just collect the final response
					for response := range responseChannel {
						lastResponse = &response
					}
				} else {
					// For console output, render incrementally
					renderer := console.NewStreamingRenderer(os.Stdout)
					for response := range responseChannel {
						err := renderer.RenderIncremental(&response)
						if err != nil {
							fmt.Printf("Error rendering streaming content: %v\n", err)
						}
						lastResponse = &response
					}
				}

				// After streaming is complete, render output
				if lastResponse != nil {
					if !outputJSON {
						fmt.Println() // Add newline after streaming content
					}
					err := console.RenderResponse(lastResponse, os.Stdout, outputJSON)
					if err != nil {
						fmt.Printf("Error: %v\n", err)
					}
				}
			}()

			// Send the streaming request
			err = client.SendSSEHTTPRequest(&wg, req, responseChannel)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			// Wait for streaming to complete
			wg.Wait()
		} else {
			// Handle non-streaming response
			var spinnerInfo *pterm.SpinnerPrinter
			if !outputJSON {
				spinnerInfo, _ = pterm.DefaultSpinner.Start("Waiting after the response from perplexity...")
			}
			res, err := client.SendCompletionRequest(req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			if !outputJSON {
				spinnerInfo.Success("Response received")
			}

			err = console.RenderResponse(res, os.Stdout, outputJSON)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		}
	},
}
