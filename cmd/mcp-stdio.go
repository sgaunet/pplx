package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sgaunet/perplexity-go/v2"
	clerrors "github.com/sgaunet/pplx/pkg/clerrors"
	"github.com/spf13/cobra"
)

var mcpStdioCmd = &cobra.Command{
	Use:   "mcp-stdio",
	Short: "Start MCP server in stdio mode",
	Long:  `Start an MCP (Model Context Protocol) server that exposes Perplexity query functionality`,
	RunE: func(_ *cobra.Command, _ []string) error {
		// Check env var PPLX_API_KEY exists
		if os.Getenv("PPLX_API_KEY") == "" {
			return clerrors.NewConfigError("PPLX_API_KEY environment variable is not set", nil)
		}

		// Create MCP server
		s := server.NewMCPServer(
			"Perplexity MCP Server",
			version,
			server.WithToolCapabilities(true),
			server.WithResourceCapabilities(false, false),
		)

		// Create query tool
		queryTool := mcp.NewTool("query",
			mcp.WithDescription("Query Perplexity AI with extensive search and filtering options"),
			// Required parameters
			mcp.WithString("user_prompt",
				mcp.Required(),
				mcp.Description("The user query/prompt"),
			),
			// Optional core parameters
			mcp.WithString("system_prompt",
				mcp.Description("System prompt to guide the AI response"),
			),
			mcp.WithString("model",
				mcp.Description("Model to use (default: "+perplexity.DefaultModel+")"),
			),
			mcp.WithNumber("frequency_penalty",
				mcp.Description("Frequency penalty for response generation"),
			),
			mcp.WithNumber("max_tokens",
				mcp.Description("Maximum number of tokens in response"),
			),
			mcp.WithNumber("presence_penalty",
				mcp.Description("Presence penalty for response generation"),
			),
			mcp.WithNumber("temperature",
				mcp.Description("Temperature for response generation"),
			),
			mcp.WithNumber("top_k",
				mcp.Description("Top-K sampling parameter"),
			),
			mcp.WithNumber("top_p",
				mcp.Description("Top-P sampling parameter"),
			),
			mcp.WithNumber("timeout",
				mcp.Description("HTTP timeout in seconds"),
			),
			// Search/Web options
			mcp.WithArray("search_domains",
				mcp.Description("Filter search results to specific domains"),
			),
			mcp.WithString("search_recency",
				mcp.Description("Filter by time: day, week, month, year, hour"),
			),
			mcp.WithNumber("location_lat",
				mcp.Description("User location latitude"),
			),
			mcp.WithNumber("location_lon",
				mcp.Description("User location longitude"),
			),
			mcp.WithString("location_country",
				mcp.Description("User location country code"),
			),
			// Response enhancement options
			mcp.WithBoolean("return_images",
				mcp.Description("Include images in response"),
			),
			mcp.WithBoolean("return_related",
				mcp.Description("Include related questions"),
			),
			mcp.WithBoolean("stream",
				mcp.Description("Enable streaming responses (will be collected and returned as complete response)"),
			),
			// Image filtering options
			mcp.WithArray("image_domains",
				mcp.Description("Filter images by domains"),
			),
			mcp.WithArray("image_formats",
				mcp.Description("Filter images by formats (jpg, png, etc.)"),
			),
			// Response format options
			mcp.WithString("response_format_json_schema",
				mcp.Description("JSON schema for structured output (sonar model only)"),
			),
			mcp.WithString("response_format_regex",
				mcp.Description("Regex pattern for structured output (sonar model only)"),
			),
			// Search mode options
			mcp.WithString("search_mode",
				mcp.Description("Search mode: web (default) or academic"),
			),
			mcp.WithString("search_context_size",
				mcp.Description("Search context size: low, medium, or high"),
			),
			// Date filtering options
			mcp.WithString("search_after_date",
				mcp.Description("Filter results published after date (MM/DD/YYYY)"),
			),
			mcp.WithString("search_before_date",
				mcp.Description("Filter results published before date (MM/DD/YYYY)"),
			),
			mcp.WithString("last_updated_after",
				mcp.Description("Filter results last updated after date (MM/DD/YYYY)"),
			),
			mcp.WithString("last_updated_before",
				mcp.Description("Filter results last updated before date (MM/DD/YYYY)"),
			),
			// Deep research options
			mcp.WithString("reasoning_effort",
				mcp.Description("Reasoning effort for sonar-deep-research: low, medium, or high"),
			),
		)

		// Add query tool handler
		s.AddTool(queryTool, func(_ context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := request.GetArguments()

			// Extract required user_prompt
			userPrompt, ok := args["user_prompt"].(string)
			if !ok || userPrompt == "" {
				return mcp.NewToolResultError("user_prompt must be a non-empty string"), nil
			}

			// Extract optional parameters with defaults
			systemPrompt := ""
			if sp, ok := args["system_prompt"].(string); ok {
				systemPrompt = sp
			}

			model := perplexity.DefaultModel
			if m, ok := args["model"].(string); ok && m != "" {
				model = m
			}

			frequencyPenalty := perplexity.DefaultFrequencyPenalty
			if fp, ok := args["frequency_penalty"].(float64); ok {
				frequencyPenalty = fp
			}

			maxTokens := perplexity.DefaultMaxTokens
			if mt, ok := args["max_tokens"].(float64); ok {
				maxTokens = int(mt)
			}

			presencePenalty := perplexity.DefaultPresencePenalty
			if pp, ok := args["presence_penalty"].(float64); ok {
				presencePenalty = pp
			}

			temperature := perplexity.DefaultTemperature
			if t, ok := args["temperature"].(float64); ok {
				temperature = t
			}

			topK := perplexity.DefaultTopK
			if tk, ok := args["top_k"].(float64); ok {
				topK = int(tk)
			}

			topP := perplexity.DefaultTopP
			if tp, ok := args["top_p"].(float64); ok {
				topP = tp
			}

			timeout := perplexity.DefaultTimeout
			if to, ok := args["timeout"].(float64); ok {
				timeout = time.Duration(to) * time.Second
			}

			// Search/Web options
			var searchDomains []string
			if sd, ok := args["search_domains"].([]any); ok {
				for _, domain := range sd {
					if domainStr, ok := domain.(string); ok {
						searchDomains = append(searchDomains, domainStr)
					}
				}
			}

			searchRecency := ""
			if sr, ok := args["search_recency"].(string); ok {
				searchRecency = sr
			}

			locationLat := float64(0)
			if lat, ok := args["location_lat"].(float64); ok {
				locationLat = lat
			}

			locationLon := float64(0)
			if lon, ok := args["location_lon"].(float64); ok {
				locationLon = lon
			}

			locationCountry := ""
			if lc, ok := args["location_country"].(string); ok {
				locationCountry = lc
			}

			// Response enhancement options
			returnImages := false
			if ri, ok := args["return_images"].(bool); ok {
				returnImages = ri
			}

			returnRelated := false
			if rr, ok := args["return_related"].(bool); ok {
				returnRelated = rr
			}

			stream := false
			if s, ok := args["stream"].(bool); ok {
				stream = s
			}

			// Image filtering options
			var imageDomains []string
			if id, ok := args["image_domains"].([]any); ok {
				for _, domain := range id {
					if domainStr, ok := domain.(string); ok {
						imageDomains = append(imageDomains, domainStr)
					}
				}
			}

			var imageFormats []string
			if imageFormatsSlice, ok := args["image_formats"].([]any); ok {
				for _, format := range imageFormatsSlice {
					if formatStr, ok := format.(string); ok {
						imageFormats = append(imageFormats, formatStr)
					}
				}
			}

			// Response format options
			responseFormatJSONSchema := ""
			if rjfs, ok := args["response_format_json_schema"].(string); ok {
				responseFormatJSONSchema = rjfs
			}

			responseFormatRegex := ""
			if rfr, ok := args["response_format_regex"].(string); ok {
				responseFormatRegex = rfr
			}

			// Search mode options
			searchMode := ""
			if sm, ok := args["search_mode"].(string); ok {
				searchMode = sm
			}

			searchContextSize := ""
			if scs, ok := args["search_context_size"].(string); ok {
				searchContextSize = scs
			}

			// Date filtering options
			searchAfterDate := ""
			if sad, ok := args["search_after_date"].(string); ok {
				searchAfterDate = sad
			}

			searchBeforeDate := ""
			if sbd, ok := args["search_before_date"].(string); ok {
				searchBeforeDate = sbd
			}

			lastUpdatedAfter := ""
			if lua, ok := args["last_updated_after"].(string); ok {
				lastUpdatedAfter = lua
			}

			lastUpdatedBefore := ""
			if lub, ok := args["last_updated_before"].(string); ok {
				lastUpdatedBefore = lub
			}

			// Deep research options
			reasoningEffort := ""
			if re, ok := args["reasoning_effort"].(string); ok {
				reasoningEffort = re
			}

			// Create perplexity client
			client := perplexity.NewClient(os.Getenv("PPLX_API_KEY"))
			client.SetHTTPTimeout(timeout)

			// Build messages
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
					errMsg := fmt.Sprintf("Invalid search-recency value '%s'. Must be one of: day, week, month, year, hour",
						searchRecency)
					return mcp.NewToolResultError(errMsg), nil
				}
				// Search recency filter is incompatible with images
				if returnImages {
					// When images are requested, explicitly disable search recency filter
					opts = append(opts, perplexity.WithSearchRecencyFilter(""))
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
						// Just warn, don't error
						warnMsg := "Warning: Image format '%s' may not be supported. " +
							"Common formats are: jpg, jpeg, png, gif, webp, svg, bmp"
						log.Printf(warnMsg, format)
					}
				}
				opts = append(opts, perplexity.WithImageFormatFilter(imageFormats))
			}

			// Add response format options
			if responseFormatJSONSchema != "" && responseFormatRegex != "" {
				return mcp.NewToolResultError("Cannot use both response_format_json_schema and response_format_regex"), nil
			}
			if responseFormatJSONSchema != "" || responseFormatRegex != "" {
				// Validate model supports response formats
				if !strings.HasPrefix(model, "sonar") {
					return mcp.NewToolResultError("Response formats (JSON schema and regex) are only supported by sonar models"), nil
				}
			}
			if responseFormatJSONSchema != "" {
				// Parse JSON schema
				var schema any
				err := json.Unmarshal([]byte(responseFormatJSONSchema), &schema)
				if err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("Invalid JSON schema: %v", err)), nil
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
					errMsg := fmt.Sprintf("Invalid search mode '%s'. Must be one of: web, academic", searchMode)
					return mcp.NewToolResultError(errMsg), nil
				}
				opts = append(opts, perplexity.WithSearchMode(searchMode))
			}
			if searchContextSize != "" {
				// Validate search context size
				validSizes := map[string]bool{"low": true, "medium": true, "high": true}
				if !validSizes[searchContextSize] {
					errMsg := fmt.Sprintf("Invalid search context size '%s'. Must be one of: low, medium, high", searchContextSize)
					return mcp.NewToolResultError(errMsg), nil
				}
				opts = append(opts, perplexity.WithSearchContextSize(searchContextSize))
			}

			// Add date filtering options
			if searchAfterDate != "" {
				date, err := time.Parse("01/02/2006", searchAfterDate)
				if err != nil {
					errMsg := fmt.Sprintf("Invalid search-after-date format '%s'. Use MM/DD/YYYY", searchAfterDate)
					return mcp.NewToolResultError(errMsg), nil
				}
				opts = append(opts, perplexity.WithSearchAfterDateFilter(date))
			}
			if searchBeforeDate != "" {
				date, err := time.Parse("01/02/2006", searchBeforeDate)
				if err != nil {
					errMsg := fmt.Sprintf("Invalid search-before-date format '%s'. Use MM/DD/YYYY", searchBeforeDate)
					return mcp.NewToolResultError(errMsg), nil
				}
				opts = append(opts, perplexity.WithSearchBeforeDateFilter(date))
			}
			if lastUpdatedAfter != "" {
				date, err := time.Parse("01/02/2006", lastUpdatedAfter)
				if err != nil {
					errMsg := fmt.Sprintf("Invalid last-updated-after format '%s'. Use MM/DD/YYYY", lastUpdatedAfter)
					return mcp.NewToolResultError(errMsg), nil
				}
				opts = append(opts, perplexity.WithLastUpdatedAfterFilter(date))
			}
			if lastUpdatedBefore != "" {
				date, err := time.Parse("01/02/2006", lastUpdatedBefore)
				if err != nil {
					errMsg := fmt.Sprintf("Invalid last-updated-before format '%s'. Use MM/DD/YYYY", lastUpdatedBefore)
					return mcp.NewToolResultError(errMsg), nil
				}
				opts = append(opts, perplexity.WithLastUpdatedBeforeFilter(date))
			}

			// Add deep research options
			if reasoningEffort != "" {
				// Validate reasoning effort
				validEfforts := map[string]bool{"low": true, "medium": true, "high": true}
				if !validEfforts[reasoningEffort] {
					errMsg := fmt.Sprintf("Invalid reasoning effort '%s'. Must be one of: low, medium, high", reasoningEffort)
					return mcp.NewToolResultError(errMsg), nil
				}
				// Check if the model supports reasoning effort
				if !strings.Contains(model, "deep-research") {
					log.Printf("Warning: reasoning-effort is only supported by sonar-deep-research model")
				}
				opts = append(opts, perplexity.WithReasoningEffort(reasoningEffort))
			}

			// Create and validate request
			req := perplexity.NewCompletionRequest(opts...)
			err := req.Validate()
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Request validation failed: %v", err)), nil
			}

			// Execute request
			var response *perplexity.CompletionResponse
			if stream {
				// Handle streaming response by collecting all responses
				responseChannel := make(chan perplexity.CompletionResponse)
				var wg sync.WaitGroup
				wg.Add(1)

				var lastResponse *perplexity.CompletionResponse
				go func() {
					defer wg.Done()
					for res := range responseChannel {
						lastResponse = &res
					}
				}()

				err = client.SendSSEHTTPRequest(&wg, req, responseChannel)
				if err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("Streaming request failed: %v", err)), nil
				}

				wg.Wait()
				response = lastResponse
			} else {
				// Handle non-streaming response
				res, err := client.SendCompletionRequest(req)
				if err != nil {
					return mcp.NewToolResultError(fmt.Sprintf("Request failed: %v", err)), nil
				}
				response = res
			}

			if response == nil {
				return mcp.NewToolResultError("No response received"), nil
			}

			// Create response object
			result := map[string]any{
				"content": response.Choices[0].Message.Content,
				"model":   response.Model,
				"usage":   response.Usage,
			}

			// Add search results if available (preferred over deprecated Citations)
			if response.SearchResults != nil && len(*response.SearchResults) > 0 {
				result["search_results"] = *response.SearchResults
			} else if response.Citations != nil && len(*response.Citations) > 0 { //nolint:staticcheck // fallback
				// Fallback to citations for backwards compatibility
				result["citations"] = *response.Citations //nolint:staticcheck // fallback for compatibility
			}

			// Add images if available
			if response.Images != nil && len(*response.Images) > 0 {
				result["images"] = *response.Images
			}

			// Add related questions if available
			if response.RelatedQuestions != nil && len(*response.RelatedQuestions) > 0 {
				result["related_questions"] = *response.RelatedQuestions
			}

			// Convert to JSON
			jsonData, err := json.Marshal(result)
			if err != nil {
				return mcp.NewToolResultError("Failed to format response"), nil
			}

			return mcp.NewToolResultText(string(jsonData)), nil
		})

		// Start the stdio server
		if err := server.ServeStdio(s); err != nil {
			return clerrors.NewAPIError("MCP server error", err)
		}
		return nil
	},
}