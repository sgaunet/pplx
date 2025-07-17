package cmd

import (
	"fmt"
	"os"
	"sync"

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
		client.SetHTTPTimeout(timeout)

		if userPrompt == "" {
			fmt.Println("Error: user prompt is required")
			cmd.Usage()
			os.Exit(1)
		}
		msg := perplexity.NewMessages(perplexity.WithSystemMessage(systemPrompt))
		msg.AddUserMessage(userPrompt)

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
				fmt.Printf("Error: Invalid search-recency value '%s'. Must be one of: day, week, month, year, hour\n", searchRecency)
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
			validFormats := map[string]bool{"jpg": true, "jpeg": true, "png": true, "gif": true, "webp": true, "svg": true, "bmp": true}
			for _, format := range imageFormats {
				if !validFormats[format] {
					fmt.Printf("Warning: Image format '%s' may not be supported. Common formats are: jpg, jpeg, png, gif, webp, svg, bmp\n", format)
				}
			}
			opts = append(opts, perplexity.WithImageFormatFilter(imageFormats))
		}

		req := perplexity.NewCompletionRequest(opts...)
		err := req.Validate()
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
				renderer := console.NewStreamingRenderer(os.Stdout)
				
				for response := range responseChannel {
					// Render the streaming content incrementally
					err := renderer.RenderIncremental(&response)
					if err != nil {
						fmt.Printf("Error rendering streaming content: %v\n", err)
					}
					lastResponse = &response
				}
				
				// After streaming is complete, render citations and other metadata
				if lastResponse != nil {
					fmt.Println() // Add newline after streaming content
					err := console.RenderCitations(lastResponse, os.Stdout)
					if err != nil {
						fmt.Printf("Error: %v\n", err)
					}
					err = console.RenderImages(lastResponse, os.Stdout)
					if err != nil {
						fmt.Printf("Error: %v\n", err)
					}
					err = console.RenderRelatedQuestions(lastResponse, os.Stdout)
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
			err = console.RenderImages(res, os.Stdout)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			err = console.RenderRelatedQuestions(res, os.Stdout)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		}
	},
}
