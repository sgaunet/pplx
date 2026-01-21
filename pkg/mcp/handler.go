package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/sgaunet/perplexity-go/v2"
)

// QueryHandler handles Perplexity query execution.
type QueryHandler struct {
	clientFactory func(apiKey string) *perplexity.Client
}

// NewQueryHandler creates a new query handler.
func NewQueryHandler() *QueryHandler {
	return &QueryHandler{
		clientFactory: perplexity.NewClient,
	}
}

// Handle processes a query tool request.
func (h *QueryHandler) Handle(
	_ context.Context,
	apiKey string,
	params QueryParams,
) (*perplexity.CompletionResponse, error) {
	// Create Perplexity client
	client := h.clientFactory(apiKey)
	client.SetHTTPTimeout(params.Timeout)

	// Build messages
	msg := perplexity.NewMessages(perplexity.WithSystemMessage(params.SystemPrompt))
	if err := msg.AddUserMessage(params.UserPrompt); err != nil {
		return nil, fmt.Errorf("failed to add user message: %w", err)
	}

	// Build request options
	opts, err := h.buildRequestOptions(params, msg)
	if err != nil {
		return nil, err
	}

	// Create and validate request
	req := perplexity.NewCompletionRequest(opts...)
	err = req.Validate()
	if err != nil {
		return nil, fmt.Errorf("request validation failed: %w", err)
	}

	// Execute request (streaming or non-streaming)
	var response *perplexity.CompletionResponse
	if params.Stream {
		response, err = h.executeStreaming(client, req)
	} else {
		response, err = h.executeNonStreaming(client, req)
	}

	if err != nil {
		return nil, err
	}

	if response == nil {
		return nil, NewStreamError("no response received", nil)
	}

	return response, nil
}

// buildRequestOptions converts QueryParams to Perplexity request options.
// This function handles the complex task of translating MCP tool parameters into the format
// expected by the Perplexity API client, with parameter validation and compatibility handling.
//
// Design rationale: Centralized option building with validation
// Rather than building options directly in the Handle method, this separation allows for:
// - Early parameter validation before any API calls
// - Explicit handling of parameter incompatibilities (e.g., search recency vs images)
// - Date parsing with user-friendly MM/DD/YYYY format conversion
// - Warning-level validation for unsupported features vs hard errors for invalid input
//
// Complexity sources (148 lines, cyclomatic 20+):
// - 30+ optional parameters, each with conditional inclusion logic
// - 4 date fields requiring MM/DD/YYYY parsing
// - Image format validation with warning emission
// - JSON schema parsing and validation
// - Parameter incompatibility handling (images + search recency conflict)
// - Model compatibility warnings (response formats, reasoning effort)
//
// Parameter incompatibility patterns:
// - search_recency + return_images: API constraint - when images are requested,
//   search recency filter must be explicitly disabled (empty string) to avoid API error.
//   Rationale: Image search uses different indexing that doesn't support time filtering.
// - response_format_json_schema + response_format_regex: Logical conflict - can only
//   constrain output format with one schema type at a time.
// - response formats + non-sonar models: API constraint - structured output formats
//   only work with sonar model family.
//
// Warning vs error strategy:
// - Hard errors: Invalid enum values, conflicting parameters, invalid JSON schema, date parse failures
// - Warnings: Unsupported image formats (user might know better than our list),
//   reasoning effort on non-deep-research models (degrades gracefully)
//
//nolint:gocognit,cyclop,funlen // Complexity inherent to building 30+ request options with validation
func (h *QueryHandler) buildRequestOptions(
	params QueryParams,
	msg perplexity.Messages,
) ([]perplexity.CompletionRequestOption, error) {
	// Validate parameters first to fail fast before building options
	// This catches invalid enum values and parameter conflicts early
	if err := h.validateParameters(params); err != nil {
		return nil, err
	}

	opts := []perplexity.CompletionRequestOption{
		perplexity.WithMessages(msg.GetMessages()),
		perplexity.WithModel(params.Model),
		perplexity.WithFrequencyPenalty(params.FrequencyPenalty),
		perplexity.WithMaxTokens(params.MaxTokens),
		perplexity.WithPresencePenalty(params.PresencePenalty),
		perplexity.WithTemperature(params.Temperature),
		perplexity.WithTopK(params.TopK),
		perplexity.WithTopP(params.TopP),
	}

	// Add search/web options
	if len(params.SearchDomains) > 0 {
		opts = append(opts, perplexity.WithSearchDomainFilter(params.SearchDomains))
	}

	if params.SearchRecency != "" {
		// API incompatibility: Search recency filter conflicts with image search
		// Image indexing uses different backend that doesn't support temporal filtering
		if params.ReturnImages {
			// When images are requested, explicitly disable search recency filter (empty string)
			// to prevent API error. User's time filter preference is ignored in this case.
			opts = append(opts, perplexity.WithSearchRecencyFilter(""))
		} else {
			opts = append(opts, perplexity.WithSearchRecencyFilter(params.SearchRecency))
		}
	}

	if params.LocationLat != 0 || params.LocationLon != 0 || params.LocationCountry != "" {
		opts = append(opts, perplexity.WithUserLocation(params.LocationLat, params.LocationLon, params.LocationCountry))
	}

	// Add response enhancement options
	if params.ReturnImages {
		opts = append(opts, perplexity.WithReturnImages(params.ReturnImages))
		// Defensive: Explicitly disable search recency filter when images requested
		// This ensures no time filter is applied even if one was set earlier,
		// preventing API incompatibility. Redundant with check above but provides
		// safety if parameter checking order changes in future refactoring.
		opts = append(opts, perplexity.WithSearchRecencyFilter(""))
	}

	if params.ReturnRelated {
		opts = append(opts, perplexity.WithReturnRelatedQuestions(params.ReturnRelated))
	}

	if params.Stream {
		opts = append(opts, perplexity.WithStream(params.Stream))
	}

	// Add image filtering options
	if len(params.ImageDomains) > 0 {
		opts = append(opts, perplexity.WithImageDomainFilter(params.ImageDomains))
	}

	if len(params.ImageFormats) > 0 {
		// Validate image formats with warnings, not errors
		// Rationale: API format support may expand over time, and user may know about
		// newly supported formats we haven't updated. Warning allows experimentation
		// while guiding users toward known-good formats. This is a "be liberal in
		// what you accept" strategy - let the API be the final validator.
		validFormats := map[string]bool{
			"jpg": true, "jpeg": true, "png": true, "gif": true,
			"webp": true, "svg": true, "bmp": true,
		}
		for _, format := range params.ImageFormats {
			if !validFormats[format] {
				log.Printf("Warning: Image format '%s' may not be supported. "+
					"Common formats are: jpg, jpeg, png, gif, webp, svg, bmp", format)
			}
		}
		opts = append(opts, perplexity.WithImageFormatFilter(params.ImageFormats))
	}

	// Add response format options
	if params.ResponseFormatJSONSchema != "" {
		// Parse JSON schema
		var schema any
		err := json.Unmarshal([]byte(params.ResponseFormatJSONSchema), &schema)
		if err != nil {
			return nil, NewValidationError("response_format_json_schema", params.ResponseFormatJSONSchema,
				fmt.Sprintf("invalid JSON schema: %v", err))
		}
		opts = append(opts, perplexity.WithJSONSchemaResponseFormat(schema))
	}

	if params.ResponseFormatRegex != "" {
		opts = append(opts, perplexity.WithRegexResponseFormat(params.ResponseFormatRegex))
	}

	// Add search mode options
	if params.SearchMode != "" {
		opts = append(opts, perplexity.WithSearchMode(params.SearchMode))
	}

	if params.SearchContextSize != "" {
		opts = append(opts, perplexity.WithSearchContextSize(params.SearchContextSize))
	}

	// Add date filtering options
	// Date format: MM/DD/YYYY chosen for user-friendliness
	// Rationale: While ISO 8601 (YYYY-MM-DD) is technically superior, MM/DD/YYYY
	// is more familiar to US users and matches common date picker formats.
	// The API accepts time.Time so the string format is purely for UX.
	if params.SearchAfterDate != "" {
		date, err := time.Parse("01/02/2006", params.SearchAfterDate)
		if err != nil {
			return nil, NewValidationError("search_after_date", params.SearchAfterDate,
				"invalid format, use MM/DD/YYYY")
		}
		opts = append(opts, perplexity.WithSearchAfterDateFilter(date))
	}

	if params.SearchBeforeDate != "" {
		date, err := time.Parse("01/02/2006", params.SearchBeforeDate)
		if err != nil {
			return nil, NewValidationError("search_before_date", params.SearchBeforeDate,
				"invalid format, use MM/DD/YYYY")
		}
		opts = append(opts, perplexity.WithSearchBeforeDateFilter(date))
	}

	if params.LastUpdatedAfter != "" {
		date, err := time.Parse("01/02/2006", params.LastUpdatedAfter)
		if err != nil {
			return nil, NewValidationError("last_updated_after", params.LastUpdatedAfter,
				"invalid format, use MM/DD/YYYY")
		}
		opts = append(opts, perplexity.WithLastUpdatedAfterFilter(date))
	}

	if params.LastUpdatedBefore != "" {
		date, err := time.Parse("01/02/2006", params.LastUpdatedBefore)
		if err != nil {
			return nil, NewValidationError("last_updated_before", params.LastUpdatedBefore,
				"invalid format, use MM/DD/YYYY")
		}
		opts = append(opts, perplexity.WithLastUpdatedBeforeFilter(date))
	}

	// Add deep research options
	if params.ReasoningEffort != "" {
		// Check if the model supports reasoning effort
		if !strings.Contains(params.Model, "deep-research") {
			log.Printf("Warning: reasoning-effort is only supported by sonar-deep-research model")
		}
		opts = append(opts, perplexity.WithReasoningEffort(params.ReasoningEffort))
	}

	return opts, nil
}

// validateParameters validates query parameters before option building.
// Performs early validation of 5 distinct categories to fail fast with clear error messages.
//
// Validation categories and their relationships:
// 1. Search recency enum validation (day, week, month, year, hour)
//    - Independent validation, no cross-parameter dependencies
//
// 2. Response format conflict detection (json_schema vs regex)
//    - Mutual exclusivity constraint: only one structured format type allowed
//    - Rationale: API can't simultaneously validate against JSON schema AND regex pattern
//
// 3. Response format model compatibility (requires sonar models)
//    - Dependency: response formats depend on model capability
//    - Rationale: Only sonar model family implements structured output parsing
//
// 4. Search mode enum validation (web, academic)
//    - Independent validation, affects search backend selection
//
// 5. Search context size and reasoning effort enum validation (low, medium, high)
//    - Independent validations, control resource allocation for query processing
//
// Design choice: Map-based enum validation vs switch statements
// Maps provide O(1) lookup and make valid values self-documenting in code.
// Alternative switch/case approach would be more verbose and harder to maintain
// as valid values change (requires modifying multiple case labels).
//
//nolint:cyclop // Complexity inherent to validating multiple parameter constraints
func (h *QueryHandler) validateParameters(params QueryParams) error {
	// Category 1: Search recency enum validation
	if params.SearchRecency != "" {
		validRecency := map[string]bool{"day": true, "week": true, "month": true, "year": true, "hour": true}
		if !validRecency[params.SearchRecency] {
			return NewValidationError("search_recency", params.SearchRecency,
				"must be one of: day, week, month, year, hour")
		}
	}

	// Category 2: Response format conflict detection
	// Ensure user didn't specify both json_schema AND regex (mutual exclusivity)
	if params.ResponseFormatJSONSchema != "" && params.ResponseFormatRegex != "" {
		return NewValidationError("response_format", "",
			"cannot use both json_schema and regex")
	}

	// Category 3: Response format model compatibility
	// Structured output formats only work with sonar model family
	if (params.ResponseFormatJSONSchema != "" || params.ResponseFormatRegex != "") &&
		!strings.HasPrefix(params.Model, "sonar") {
		return NewValidationError("response_format", "",
			"only supported by sonar models")
	}

	// Category 4: Search mode enum validation
	// Controls whether search uses web or academic (scholarly) backend
	if params.SearchMode != "" {
		validModes := map[string]bool{"web": true, "academic": true}
		if !validModes[params.SearchMode] {
			return NewValidationError("search_mode", params.SearchMode,
				"must be one of: web, academic")
		}
	}

	// Category 5a: Search context size enum validation
	// Controls how much context from search results is included in the query
	// (low = less context/faster, high = more context/slower but potentially better answers)
	if params.SearchContextSize != "" {
		validSizes := map[string]bool{"low": true, "medium": true, "high": true}
		if !validSizes[params.SearchContextSize] {
			return NewValidationError("search_context_size", params.SearchContextSize,
				"must be one of: low, medium, high")
		}
	}

	// Category 5b: Reasoning effort enum validation
	// Controls computational resources for deep-research model
	// (low = faster/cheaper, high = slower/more thorough reasoning)
	if params.ReasoningEffort != "" {
		validEfforts := map[string]bool{"low": true, "medium": true, "high": true}
		if !validEfforts[params.ReasoningEffort] {
			return NewValidationError("reasoning_effort", params.ReasoningEffort,
				"must be one of: low, medium, high")
		}
	}

	return nil
}

// executeStreaming handles streaming response execution.
// Uses goroutine-based concurrency to consume the stream while the main thread produces
// Server-Sent Events (SSE) via the HTTP request.
//
// Design rationale: Return last response, not concatenated content
// Perplexity's SSE stream sends cumulative responses where each event contains the
// COMPLETE response so far, not just incremental deltas. Therefore:
// - We don't need to concatenate content from multiple events
// - The last response contains the final, complete answer
// - Earlier responses are just intermediate states (useful for UI rendering but not final result)
// - Returning lastResponse gives us the complete final answer plus all metadata (citations, etc.)
//
// Goroutine coordination pattern:
// - Main thread: Calls SendSSEHTTPRequest which produces events to responseChannel
// - Consumer goroutine: Reads from responseChannel, keeping only the last response
// - WaitGroup: Ensures consumer goroutine finishes before we return
// - Unbuffered channel: Provides natural backpressure (server waits if consumer is slow)
//
// Alternative designs considered:
// - Return first response: Would give incomplete answer
// - Concatenate all responses: Would duplicate content since responses are cumulative
// - Stream to caller: Would require caller to handle channel/goroutine complexity
//
// Error handling:
// - If SendSSEHTTPRequest fails, return immediately (consumer goroutine will finish naturally)
// - If stream completes but no responses received, return error (likely network issue).
func (h *QueryHandler) executeStreaming(
	client *perplexity.Client,
	req *perplexity.CompletionRequest,
) (*perplexity.CompletionResponse, error) {
	responseChannel := make(chan perplexity.CompletionResponse)
	var wg sync.WaitGroup
	wg.Add(1)

	// Consumer goroutine: Drain channel and keep only the last response
	// Rationale: Last response contains complete cumulative content
	var lastResponse *perplexity.CompletionResponse
	go func() {
		defer wg.Done()
		for res := range responseChannel {
			// Overwrite with each new response; only last one matters for final result
			lastResponse = &res
		}
	}()

	// Producer: Main thread sends SSE request which populates responseChannel
	err := client.SendSSEHTTPRequest(&wg, req, responseChannel)
	if err != nil {
		return nil, NewStreamError("streaming request failed", err)
	}

	// Wait for consumer goroutine to finish draining channel
	wg.Wait()

	if lastResponse == nil {
		return nil, NewStreamError("no response received from stream", nil)
	}

	return lastResponse, nil
}

// executeNonStreaming handles non-streaming response execution.
func (h *QueryHandler) executeNonStreaming(
	client *perplexity.Client,
	req *perplexity.CompletionRequest,
) (*perplexity.CompletionResponse, error) {
	response, err := client.SendCompletionRequest(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	return response, nil
}
