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
		config.ApplyToGlobals(cfg, globalOpts)

		// Step 2: Initialize API client
		// API key checked here (not in config load) because it's required at runtime,
		// but config file is optional. This provides fast feedback if key is missing.
		// Fail fast principle: better to error immediately than during expensive API call.
		if os.Getenv("PPLX_API_KEY") == "" {
			return clerrors.NewConfigError("PPLX_API_KEY environment variable is not set", nil)
		}

		client := perplexity.NewClient(os.Getenv("PPLX_API_KEY"))
		client.SetHTTPTimeout(globalOpts.Timeout)

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
		if globalOpts.Stream {
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
	msg := perplexity.NewMessages(perplexity.WithSystemMessage(globalOpts.SystemPrompt))
	if err := msg.AddUserMessage(globalOpts.UserPrompt); err != nil {
		return nil, fmt.Errorf("failed to add user message to request: %w", err)
	}

	return []perplexity.CompletionRequestOption{
		perplexity.WithMessages(msg.GetMessages()),
		perplexity.WithModel(globalOpts.Model),
		perplexity.WithFrequencyPenalty(globalOpts.FrequencyPenalty),
		perplexity.WithMaxTokens(globalOpts.MaxTokens),
		perplexity.WithPresencePenalty(globalOpts.PresencePenalty),
		perplexity.WithTemperature(globalOpts.Temperature),
		perplexity.WithTopK(globalOpts.TopK),
		perplexity.WithTopP(globalOpts.TopP),
	}, nil
}

// buildSearchOptions creates search-related options for the completion request.
// Returns options based on search flags (domains, recency, location).
func buildSearchOptions() []perplexity.CompletionRequestOption {
	var opts []perplexity.CompletionRequestOption

	if len(globalOpts.SearchDomains) > 0 {
		opts = append(opts, perplexity.WithSearchDomainFilter(globalOpts.SearchDomains))
	}

	// Handle search recency - incompatible with images
	if globalOpts.SearchRecency != "" {
		if globalOpts.ReturnImages {
			// User-facing notification (not a log message)
			fmt.Printf("Note: When using --return-images, search-recency is automatically disabled\nProceeding with image search...\n")
		} else {
			opts = append(opts, perplexity.WithSearchRecencyFilter(globalOpts.SearchRecency))
		}
	}

	if globalOpts.LocationLat != 0 || globalOpts.LocationLon != 0 || globalOpts.LocationCountry != "" {
		opts = append(opts, perplexity.WithUserLocation(globalOpts.LocationLat, globalOpts.LocationLon, globalOpts.LocationCountry))
	}

	if globalOpts.SearchMode != "" {
		opts = append(opts, perplexity.WithSearchMode(globalOpts.SearchMode))
	}

	if globalOpts.SearchContextSize != "" {
		opts = append(opts, perplexity.WithSearchContextSize(globalOpts.SearchContextSize))
	}

	return opts
}

// buildResponseOptions creates response enhancement options.
// Handles streaming, images, and related questions.
func buildResponseOptions() []perplexity.CompletionRequestOption {
	var opts []perplexity.CompletionRequestOption

	if globalOpts.ReturnImages {
		opts = append(opts, perplexity.WithReturnImages(globalOpts.ReturnImages))
		// When images are requested, explicitly disable search recency
		opts = append(opts, perplexity.WithSearchRecencyFilter(""))
	}

	if globalOpts.ReturnRelated {
		opts = append(opts, perplexity.WithReturnRelatedQuestions(globalOpts.ReturnRelated))
	}

	if globalOpts.Stream {
		opts = append(opts, perplexity.WithStream(globalOpts.Stream))
	}

	return opts
}

// buildImageOptions creates image filtering options.
// Validates image formats and warns about unsupported formats.
func buildImageOptions() []perplexity.CompletionRequestOption {
	var opts []perplexity.CompletionRequestOption

	if len(globalOpts.ImageDomains) > 0 {
		opts = append(opts, perplexity.WithImageDomainFilter(globalOpts.ImageDomains))
	}

	if len(globalOpts.ImageFormats) > 0 {
		// Validate image formats
		for _, format := range globalOpts.ImageFormats {
			if !config.ValidImageFormats[format] {
				logger.Warn("image format may not be supported",
					"format", format,
					"supported", "jpg, jpeg, png, gif, webp, svg, bmp")
			}
		}
		opts = append(opts, perplexity.WithImageFormatFilter(globalOpts.ImageFormats))
	}

	return opts
}

// buildResponseFormatOptions creates structured output format options.
// Handles JSON schema and regex response formats (sonar models only).
func buildResponseFormatOptions() ([]perplexity.CompletionRequestOption, error) {
	var opts []perplexity.CompletionRequestOption

	if globalOpts.ResponseFormatJSONSchema != "" {
		var schema any
		err := json.Unmarshal([]byte(globalOpts.ResponseFormatJSONSchema), &schema)
		if err != nil {
			return nil, clerrors.NewValidationError("response-format-json-schema",
				globalOpts.ResponseFormatJSONSchema, fmt.Sprintf("invalid JSON schema: %v", err))
		}
		opts = append(opts, perplexity.WithJSONSchemaResponseFormat(schema))
	}

	if globalOpts.ResponseFormatRegex != "" {
		opts = append(opts, perplexity.WithRegexResponseFormat(globalOpts.ResponseFormatRegex))
	}

	return opts, nil
}

// buildDateFilterOptions creates date-based filtering options.
// Parses and validates all date filter flags.
func buildDateFilterOptions() ([]perplexity.CompletionRequestOption, error) {
	var opts []perplexity.CompletionRequestOption

	if globalOpts.SearchAfterDate != "" {
		date, err := parseDateFilter("search-after-date", globalOpts.SearchAfterDate)
		if err != nil {
			return nil, err
		}
		opts = append(opts, perplexity.WithPublishedAfter(date))
	}

	if globalOpts.SearchBeforeDate != "" {
		date, err := parseDateFilter("search-before-date", globalOpts.SearchBeforeDate)
		if err != nil {
			return nil, err
		}
		opts = append(opts, perplexity.WithPublishedBefore(date))
	}

	if globalOpts.LastUpdatedAfter != "" {
		date, err := parseDateFilter("last-updated-after", globalOpts.LastUpdatedAfter)
		if err != nil {
			return nil, err
		}
		opts = append(opts, perplexity.WithLastUpdatedAfterFilter(date))
	}

	if globalOpts.LastUpdatedBefore != "" {
		date, err := parseDateFilter("last-updated-before", globalOpts.LastUpdatedBefore)
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

	if globalOpts.ReasoningEffort != "" {
		// Warn if model doesn't support reasoning effort
		if !strings.Contains(globalOpts.Model, "deep-research") {
			logger.Warn("reasoning-effort only supported by sonar-deep-research model",
				"current_model", globalOpts.Model)
		}
		opts = append(opts, perplexity.WithReasoningEffort(globalOpts.ReasoningEffort))
	}

	return opts
}

// validateInputs validates user inputs before building the request.
// This centralizes all pre-request validation logic.
func validateInputs() error {
	if globalOpts.UserPrompt == "" {
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
	if err := validateStringEnum("search-recency", globalOpts.SearchRecency,
		config.ValidSearchRecency, "day, week, month, year, hour"); err != nil {
		return err
	}

	// Validate search mode
	if err := validateStringEnum("search-mode", globalOpts.SearchMode,
		config.ValidSearchModes, "web, academic"); err != nil {
		return err
	}

	// Validate search context size
	if err := validateStringEnum("search-context-size", globalOpts.SearchContextSize,
		config.ValidContextSizes, "low, medium, high"); err != nil {
		return err
	}

	// Validate reasoning effort
	return validateStringEnum("reasoning-effort", globalOpts.ReasoningEffort,
		config.ValidReasoningEfforts, "low, medium, high")
}

// validateResponseFormats validates response format options and model compatibility.
func validateResponseFormats() error {
	// Validate response format mutual exclusivity
	if globalOpts.ResponseFormatJSONSchema != "" && globalOpts.ResponseFormatRegex != "" {
		return clerrors.NewValidationError("response-format", "",
			"cannot use both --response-format-json-schema and --response-format-regex")
	}

	// Validate model for response formats
	hasResponseFormat := globalOpts.ResponseFormatJSONSchema != "" || globalOpts.ResponseFormatRegex != ""
	if hasResponseFormat && !strings.HasPrefix(globalOpts.Model, "sonar") {
		return clerrors.NewValidationError("model", globalOpts.Model,
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

		if globalOpts.OutputJSON {
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
			if !globalOpts.OutputJSON {
				// Visual separation between streaming content and metadata sections
				fmt.Println()
			}
			err := console.RenderResponse(lastResponse, os.Stdout, globalOpts.OutputJSON)
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
	if !globalOpts.OutputJSON {
		spinnerInfo, _ = pterm.DefaultSpinner.Start("Waiting for response from perplexity...")
	}

	res, err := client.SendCompletionRequest(req)
	if err != nil {
		return clerrors.NewAPIError("failed to send completion request", err)
	}

	if !globalOpts.OutputJSON {
		spinnerInfo.Success("Response received")
	}

	err = console.RenderResponse(res, os.Stdout, globalOpts.OutputJSON)
	if err != nil {
		return clerrors.NewIOError("failed to render response", err)
	}

	return nil
}
