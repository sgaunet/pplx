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

// Validation maps for query command options.
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
		// Step 1: Load and merge configuration
		// Graceful degradation: If config load fails, continue with CLI flags only.
		// Rationale: User may not have a config file yet, but CLI should still work.
		// This allows the tool to be used immediately after installation without setup.
		cfg, err := config.LoadAndMergeConfig(cmd, configFilePath)
		if err != nil {
			// Non-fatal: continue with CLI flags only
			cfg = config.NewConfigData()
		}

		// Apply merged config to global variables
		// Design note: Uses globals for cobra flag compatibility - flags are bound to globals,
		// and ApplyToGlobals ensures config values only apply when flags aren't set.
		config.ApplyToGlobals(cfg,
			&model, &temperature, &maxTokens, &topK, &topP,
			&frequencyPenalty, &presencePenalty, &timeout,
			&searchDomains, &searchRecency, &locationLat, &locationLon, &locationCountry,
			&returnImages, &returnRelated, &stream,
			&searchMode, &searchContextSize,
		)

		// Step 2: Initialize API client
		// API key checked here (not in config load) because it's required at runtime,
		// but config file is optional. This provides fast feedback if key is missing.
		// Fail fast principle: better to error immediately than during expensive API call.
		if os.Getenv("PPLX_API_KEY") == "" {
			return clerrors.NewConfigError("PPLX_API_KEY environment variable is not set", nil)
		}

		client := perplexity.NewClient(os.Getenv("PPLX_API_KEY"))
		client.SetHTTPTimeout(timeout)

		// Step 3: Validate inputs
		// Early validation before expensive API call provides fast feedback on errors.
		// Catches malformed options (invalid dates, conflicting flags) before network call.
		if err := validateInputs(); err != nil {
			return err
		}

		// Step 4: Build request with all options
		// Separated into dedicated function for testability and reusability.
		// Allows testing request building logic independently from API calls.
		req, err := buildAllOptions()
		if err != nil {
			return err
		}

		// Step 5: Execute request (streaming or non-streaming)
		// Different code paths because streaming requires goroutine coordination
		// while non-streaming uses synchronous request-response pattern.
		// Streaming: incremental rendering with channels and goroutines
		// Non-streaming: spinner while waiting, then render complete response
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
			"must be one of: "+validList)
	}
	return nil
}

// buildBaseOptions creates the base completion request options.
// These options are always included in every request.
func buildBaseOptions() ([]perplexity.CompletionRequestOption, error) {
	msg := perplexity.NewMessages(perplexity.WithSystemMessage(systemPrompt))
	if err := msg.AddUserMessage(userPrompt); err != nil {
		return nil, fmt.Errorf("failed to add user message to request: %w", err)
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

	if err := validateEnumFields(); err != nil {
		return err
	}

	return validateResponseFormats()
}

// validateEnumFields validates all enum-based configuration options.
func validateEnumFields() error {
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
	return validateStringEnum("reasoning-effort", reasoningEffort,
		validReasoningEfforts, "low, medium, high")
}

// validateResponseFormats validates response format options and model compatibility.
func validateResponseFormats() error {
	// Validate response format mutual exclusivity
	if responseFormatJSONSchema != "" && responseFormatRegex != "" {
		return clerrors.NewValidationError("response-format", "",
			"cannot use both --response-format-json-schema and --response-format-regex")
	}

	// Validate model for response formats
	hasResponseFormat := responseFormatJSONSchema != "" || responseFormatRegex != ""
	if hasResponseFormat && !strings.HasPrefix(model, "sonar") {
		return clerrors.NewValidationError("model", model,
			"response formats (JSON schema and regex) are only supported by sonar models")
	}

	return nil
}

// buildAllOptions builds all completion request options using the builder aggregation pattern.
// Each builder contributes a subset of options for its concern (search, format, dates, etc.),
// making individual builders testable and allowing conditional inclusion based on flags.
//
// Design rationale: Separating builders by concern (rather than one monolithic function)
// improves maintainability and testability. Each builder can be tested independently,
// and new option categories can be added without modifying existing builders.
func buildAllOptions() (*perplexity.CompletionRequest, error) {
	// Build base options first - always required, no conditionals
	// These establish the request foundation: model, messages, temperature, etc.
	// Must come first as other builders may depend on model selection
	baseOpts, err := buildBaseOptions()
	if err != nil {
		return nil, err
	}

	// Append-only builders: these never fail, safe to chain unconditionally
	// Each builder checks its relevant flags and adds options if set
	// Order doesn't matter for these - they're independent concerns
	opts := baseOpts
	opts = append(opts, buildSearchOptions()...)      // Domain filters, recency, location
	opts = append(opts, buildResponseOptions()...)    // Return images, related questions
	opts = append(opts, buildImageOptions()...)       // Image domain/format filters

	// Error-returning builders: handle validation during option creation
	// These validate input format/syntax before API call to provide fast feedback

	// Format options: validates JSON schema syntax and checks mutual exclusivity
	// (can't have both JSON schema and regex at same time)
	formatOpts, err := buildResponseFormatOptions()
	if err != nil {
		return nil, err
	}
	opts = append(opts, formatOpts...)

	// Date options: validates MM/DD/YYYY format before sending to API
	// Catches malformed dates early rather than getting API error
	dateOpts, err := buildDateFilterOptions()
	if err != nil {
		return nil, err
	}
	opts = append(opts, dateOpts...)

	// Deep research options: appended last for clarity (order doesn't matter functionally)
	// Keeps research-specific options isolated from standard query options
	opts = append(opts, buildDeepResearchOptions()...)

	// Create request and run final validation
	// This checks the complete option set for conflicts that span multiple builders
	// Example: response format requires sonar model (checked across base + format builders)
	req := perplexity.NewCompletionRequest(opts...)
	if err := req.Validate(); err != nil {
		return nil, clerrors.NewValidationError("request", "", err.Error())
	}

	return req, nil
}

// handleStreamingResponse processes a streaming completion request.
// Uses goroutine-based concurrency to handle incremental token rendering while
// maintaining the final response for complete metadata (citations, images, related questions).
// The goroutine consumes from responseChannel while the main thread produces via SendSSEHTTPRequest.
func handleStreamingResponse(client *perplexity.Client, req *perplexity.CompletionRequest) error {
	// Unbuffered channel ensures backpressure: server won't send next chunk until we process current one
	// This prevents memory buildup if rendering is slower than server response generation
	responseChannel := make(chan perplexity.CompletionResponse)
	var wg sync.WaitGroup
	wg.Add(1)

	// Start consumer goroutine - must be running before SendSSEHTTPRequest starts producing
	// Goroutine lifecycle: spawn -> consume from channel -> defer wg.Done() -> terminate when channel closes
	go func() {
		defer wg.Done()
		var lastResponse *perplexity.CompletionResponse

		if outputJSON {
			// JSON mode: skip incremental rendering, just collect final response
			// Rationale: JSON clients expect complete, valid JSON - not streaming fragments
			// The lastResponse will contain the full completion with all metadata
			for response := range responseChannel {
				lastResponse = &response
			}
		} else {
			// Console mode: render tokens incrementally for better user experience
			// Shows progress as the model generates output (similar to ChatGPT interface)
			renderer := console.NewStreamingRenderer(os.Stdout)
			for response := range responseChannel {
				err := renderer.RenderIncremental(&response)
				if err != nil {
					logger.Error("failed to render streaming content", "error", err)
				}
				// Preserve last response for complete metadata (citations, images, related questions)
				// Only the final chunk contains these - intermediate chunks have partial text only
				lastResponse = &response
			}
		}

		// After channel closes (streaming complete), render final metadata
		// This includes citations, images, and related questions that only arrive in the final chunk
		if lastResponse != nil {
			if !outputJSON {
				// Visual separation between streaming content and metadata sections
				fmt.Println()
			}
			err := console.RenderResponse(lastResponse, os.Stdout, outputJSON)
			if err != nil {
				logger.Error("failed to render response", "error", err)
			}
		}
	}()

	// Producer: sends SSE request, server writes to responseChannel as tokens arrive
	// Must happen after goroutine spawn to ensure consumer is ready to receive
	// If called before goroutine starts, channel would block and deadlock
	err := client.SendSSEHTTPRequest(&wg, req, responseChannel)
	if err != nil {
		return clerrors.NewAPIError("failed to send streaming request", err)
	}

	// Wait for consumer goroutine to finish processing all responses
	// Critical: prevents goroutine leak and ensures all output is written before function returns
	// Without this, main function might exit while goroutine is still rendering
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
