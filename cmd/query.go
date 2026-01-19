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
	clerrors "github.com/sgaunet/pplx/pkg/clerrors"
	"github.com/sgaunet/pplx/pkg/logger"
	"github.com/spf13/cobra"
)

// Validation maps for query command options
var (
	validSearchRecency = map[string]bool{
		"day": true, "week": true, "month": true,
		"year": true, "hour": true,
	}
	validSearchModes = map[string]bool{
		"web": true, "academic": true,
	}
	validContextSizes = map[string]bool{
		"low": true, "medium": true, "high": true,
	}
	validReasoningEfforts = map[string]bool{
		"low": true, "medium": true, "high": true,
	}
	validImageFormats = map[string]bool{
		"jpg": true, "jpeg": true, "png": true, "gif": true,
		"webp": true, "svg": true, "bmp": true,
	}
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "",
	Long:  ``,
	RunE: func(cmd *cobra.Command, _ []string) error {
		// 1. Load and apply configuration
		cfg, err := config.LoadAndMergeConfig(cmd, configFilePath)
		if err != nil {
			// Non-fatal: continue with CLI flags only
			cfg = config.NewConfigData()
		}

		config.ApplyToGlobals(cfg,
			&model, &temperature, &maxTokens, &topK, &topP,
			&frequencyPenalty, &presencePenalty, &timeout,
			&searchDomains, &searchRecency, &locationLat, &locationLon, &locationCountry,
			&returnImages, &returnRelated, &stream,
			&searchMode, &searchContextSize,
		)

		// 2. Initialize client
		if os.Getenv("PPLX_API_KEY") == "" {
			return clerrors.NewConfigError("PPLX_API_KEY environment variable is not set", nil)
		}

		client := perplexity.NewClient(os.Getenv("PPLX_API_KEY"))
		client.SetHTTPTimeout(timeout)

		// 3. Validate inputs
		if err := validateInputs(); err != nil {
			return err
		}

		// 4. Build request
		req, err := buildAllOptions()
		if err != nil {
			return err
		}

		// 5. Execute request (streaming or non-streaming)
		if stream {
			return handleStreamingResponse(client, req)
		}
		return handleNonStreamingResponse(client, req)
	},
}

// parseDateFilter parses a date string in MM/DD/YYYY format for API filters.
// Returns the parsed time and an error if parsing fails.
func parseDateFilter(fieldName, dateStr string) (time.Time, error) {
	date, err := time.Parse("01/02/2006", dateStr)
	if err != nil {
		return time.Time{}, clerrors.NewValidationError(
			fieldName, dateStr, "invalid date format, use MM/DD/YYYY",
		)
	}
	return date, nil
}

// validateStringEnum validates that a value is in the allowed set.
// Returns nil if value is empty or valid, error otherwise.
func validateStringEnum(fieldName, value string, validValues map[string]bool, validList string) error {
	if value == "" {
		return nil
	}
	if !validValues[value] {
		return clerrors.NewValidationError(fieldName, value,
			fmt.Sprintf("must be one of: %s", validList))
	}
	return nil
}

// buildBaseOptions creates the base completion request options.
// These options are always included in every request.
func buildBaseOptions() ([]perplexity.CompletionRequestOption, error) {
	msg := perplexity.NewMessages(perplexity.WithSystemMessage(systemPrompt))
	if err := msg.AddUserMessage(userPrompt); err != nil {
		return nil, fmt.Errorf("failed to add user message: %w", err)
	}

	return []perplexity.CompletionRequestOption{
		perplexity.WithMessages(msg.GetMessages()),
		perplexity.WithModel(model),
		perplexity.WithFrequencyPenalty(frequencyPenalty),
		perplexity.WithMaxTokens(maxTokens),
		perplexity.WithPresencePenalty(presencePenalty),
		perplexity.WithTemperature(temperature),
		perplexity.WithTopK(topK),
		perplexity.WithTopP(topP),
	}, nil
}

// buildSearchOptions creates search-related options for the completion request.
// Returns options based on search flags (domains, recency, location).
func buildSearchOptions() []perplexity.CompletionRequestOption {
	var opts []perplexity.CompletionRequestOption

	if len(searchDomains) > 0 {
		opts = append(opts, perplexity.WithSearchDomainFilter(searchDomains))
	}

	// Handle search recency - incompatible with images
	if searchRecency != "" {
		if returnImages {
			// User-facing notification (not a log message)
			fmt.Printf("Note: When using --return-images, search-recency is automatically disabled\nProceeding with image search...\n")
		} else {
			opts = append(opts, perplexity.WithSearchRecencyFilter(searchRecency))
		}
	}

	if locationLat != 0 || locationLon != 0 || locationCountry != "" {
		opts = append(opts, perplexity.WithUserLocation(locationLat, locationLon, locationCountry))
	}

	if searchMode != "" {
		opts = append(opts, perplexity.WithSearchMode(searchMode))
	}

	if searchContextSize != "" {
		opts = append(opts, perplexity.WithSearchContextSize(searchContextSize))
	}

	return opts
}

// buildResponseOptions creates response enhancement options.
// Handles streaming, images, and related questions.
func buildResponseOptions() []perplexity.CompletionRequestOption {
	var opts []perplexity.CompletionRequestOption

	if returnImages {
		opts = append(opts, perplexity.WithReturnImages(returnImages))
		// When images are requested, explicitly disable search recency
		opts = append(opts, perplexity.WithSearchRecencyFilter(""))
	}

	if returnRelated {
		opts = append(opts, perplexity.WithReturnRelatedQuestions(returnRelated))
	}

	if stream {
		opts = append(opts, perplexity.WithStream(stream))
	}

	return opts
}

// buildImageOptions creates image filtering options.
// Validates image formats and warns about unsupported formats.
func buildImageOptions() []perplexity.CompletionRequestOption {
	var opts []perplexity.CompletionRequestOption

	if len(imageDomains) > 0 {
		opts = append(opts, perplexity.WithImageDomainFilter(imageDomains))
	}

	if len(imageFormats) > 0 {
		// Validate image formats
		for _, format := range imageFormats {
			if !validImageFormats[format] {
				logger.Warn("image format may not be supported",
					"format", format,
					"supported", "jpg, jpeg, png, gif, webp, svg, bmp")
			}
		}
		opts = append(opts, perplexity.WithImageFormatFilter(imageFormats))
	}

	return opts
}

// buildResponseFormatOptions creates structured output format options.
// Handles JSON schema and regex response formats (sonar models only).
func buildResponseFormatOptions() ([]perplexity.CompletionRequestOption, error) {
	var opts []perplexity.CompletionRequestOption

	if responseFormatJSONSchema != "" {
		var schema any
		err := json.Unmarshal([]byte(responseFormatJSONSchema), &schema)
		if err != nil {
			return nil, clerrors.NewValidationError("response-format-json-schema",
				responseFormatJSONSchema, fmt.Sprintf("invalid JSON schema: %v", err))
		}
		opts = append(opts, perplexity.WithJSONSchemaResponseFormat(schema))
	}

	if responseFormatRegex != "" {
		opts = append(opts, perplexity.WithRegexResponseFormat(responseFormatRegex))
	}

	return opts, nil
}

// buildDateFilterOptions creates date-based filtering options.
// Parses and validates all date filter flags.
func buildDateFilterOptions() ([]perplexity.CompletionRequestOption, error) {
	var opts []perplexity.CompletionRequestOption

	if searchAfterDate != "" {
		date, err := parseDateFilter("search-after-date", searchAfterDate)
		if err != nil {
			return nil, err
		}
		opts = append(opts, perplexity.WithSearchAfterDateFilter(date))
	}

	if searchBeforeDate != "" {
		date, err := parseDateFilter("search-before-date", searchBeforeDate)
		if err != nil {
			return nil, err
		}
		opts = append(opts, perplexity.WithSearchBeforeDateFilter(date))
	}

	if lastUpdatedAfter != "" {
		date, err := parseDateFilter("last-updated-after", lastUpdatedAfter)
		if err != nil {
			return nil, err
		}
		opts = append(opts, perplexity.WithLastUpdatedAfterFilter(date))
	}

	if lastUpdatedBefore != "" {
		date, err := parseDateFilter("last-updated-before", lastUpdatedBefore)
		if err != nil {
			return nil, err
		}
		opts = append(opts, perplexity.WithLastUpdatedBeforeFilter(date))
	}

	return opts, nil
}

// buildDeepResearchOptions creates deep research options.
// Handles reasoning effort for sonar-deep-research models.
func buildDeepResearchOptions() []perplexity.CompletionRequestOption {
	var opts []perplexity.CompletionRequestOption

	if reasoningEffort != "" {
		// Warn if model doesn't support reasoning effort
		if !strings.Contains(model, "deep-research") {
			logger.Warn("reasoning-effort only supported by sonar-deep-research model",
				"current_model", model)
		}
		opts = append(opts, perplexity.WithReasoningEffort(reasoningEffort))
	}

	return opts
}

// validateInputs validates user inputs before building the request.
// This centralizes all pre-request validation logic.
func validateInputs() error {
	if userPrompt == "" {
		return clerrors.NewValidationError("user-prompt", "", "user prompt is required")
	}

	// Validate search recency
	if err := validateStringEnum("search-recency", searchRecency,
		validSearchRecency, "day, week, month, year, hour"); err != nil {
		return err
	}

	// Validate search mode
	if err := validateStringEnum("search-mode", searchMode,
		validSearchModes, "web, academic"); err != nil {
		return err
	}

	// Validate search context size
	if err := validateStringEnum("search-context-size", searchContextSize,
		validContextSizes, "low, medium, high"); err != nil {
		return err
	}

	// Validate reasoning effort
	if err := validateStringEnum("reasoning-effort", reasoningEffort,
		validReasoningEfforts, "low, medium, high"); err != nil {
		return err
	}

	// Validate response format mutual exclusivity
	if responseFormatJSONSchema != "" && responseFormatRegex != "" {
		return clerrors.NewValidationError("response-format", "",
			"cannot use both --response-format-json-schema and --response-format-regex")
	}

	// Validate model for response formats
	if (responseFormatJSONSchema != "" || responseFormatRegex != "") &&
		!strings.HasPrefix(model, "sonar") {
		return clerrors.NewValidationError("model", model,
			"response formats (JSON schema and regex) are only supported by sonar models")
	}

	return nil
}

// buildAllOptions builds all completion request options.
// Combines options from all builders and validates the final request.
func buildAllOptions() (*perplexity.CompletionRequest, error) {
	// Build base options
	baseOpts, err := buildBaseOptions()
	if err != nil {
		return nil, err
	}

	// Append all option groups
	opts := baseOpts
	opts = append(opts, buildSearchOptions()...)
	opts = append(opts, buildResponseOptions()...)
	opts = append(opts, buildImageOptions()...)

	// These can return errors, so handle them
	formatOpts, err := buildResponseFormatOptions()
	if err != nil {
		return nil, err
	}
	opts = append(opts, formatOpts...)

	dateOpts, err := buildDateFilterOptions()
	if err != nil {
		return nil, err
	}
	opts = append(opts, dateOpts...)

	opts = append(opts, buildDeepResearchOptions()...)

	// Create and validate request
	req := perplexity.NewCompletionRequest(opts...)
	if err := req.Validate(); err != nil {
		return nil, clerrors.NewValidationError("request", "", err.Error())
	}

	return req, nil
}

// handleStreamingResponse processes a streaming completion request.
// Manages goroutines for incremental rendering and final output.
func handleStreamingResponse(client *perplexity.Client, req *perplexity.CompletionRequest) error {
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
					logger.Error("failed to render streaming content", "error", err)
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
				logger.Error("failed to render response", "error", err)
			}
		}
	}()

	// Send the streaming request
	err := client.SendSSEHTTPRequest(&wg, req, responseChannel)
	if err != nil {
		return clerrors.NewAPIError("failed to send streaming request", err)
	}

	// Wait for streaming to complete
	wg.Wait()
	return nil
}

// handleNonStreamingResponse processes a standard (non-streaming) completion request.
// Shows a spinner while waiting for the response (unless JSON output is requested).
func handleNonStreamingResponse(client *perplexity.Client, req *perplexity.CompletionRequest) error {
	var spinnerInfo *pterm.SpinnerPrinter
	if !outputJSON {
		spinnerInfo, _ = pterm.DefaultSpinner.Start("Waiting for response from perplexity...")
	}

	res, err := client.SendCompletionRequest(req)
	if err != nil {
		return clerrors.NewAPIError("failed to send completion request", err)
	}

	if !outputJSON {
		spinnerInfo.Success("Response received")
	}

	err = console.RenderResponse(res, os.Stdout, outputJSON)
	if err != nil {
		return clerrors.NewIOError("failed to render response", err)
	}

	return nil
}
