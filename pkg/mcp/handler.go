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
//nolint:gocognit,cyclop,funlen // Complexity inherent to building 30+ request options with validation
func (h *QueryHandler) buildRequestOptions(
	params QueryParams,
	msg perplexity.Messages,
) ([]perplexity.CompletionRequestOption, error) {
	// Validate parameters first
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
		// Search recency filter is incompatible with images
		if params.ReturnImages {
			// When images are requested, explicitly disable search recency filter
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
		// When images are requested, explicitly disable search recency filter
		// to avoid API incompatibility issues
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
		// Validate image formats (just warn, don't error)
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

// validateParameters validates query parameters.
//nolint:cyclop // Complexity inherent to validating multiple parameter constraints
func (h *QueryHandler) validateParameters(params QueryParams) error {
	// Validate search recency
	if params.SearchRecency != "" {
		validRecency := map[string]bool{"day": true, "week": true, "month": true, "year": true, "hour": true}
		if !validRecency[params.SearchRecency] {
			return NewValidationError("search_recency", params.SearchRecency,
				"must be one of: day, week, month, year, hour")
		}
	}

	// Validate response format conflicts
	if params.ResponseFormatJSONSchema != "" && params.ResponseFormatRegex != "" {
		return NewValidationError("response_format", "",
			"cannot use both json_schema and regex")
	}

	// Validate response format model compatibility
	if (params.ResponseFormatJSONSchema != "" || params.ResponseFormatRegex != "") &&
		!strings.HasPrefix(params.Model, "sonar") {
		return NewValidationError("response_format", "",
			"only supported by sonar models")
	}

	// Validate search mode
	if params.SearchMode != "" {
		validModes := map[string]bool{"web": true, "academic": true}
		if !validModes[params.SearchMode] {
			return NewValidationError("search_mode", params.SearchMode,
				"must be one of: web, academic")
		}
	}

	// Validate search context size
	if params.SearchContextSize != "" {
		validSizes := map[string]bool{"low": true, "medium": true, "high": true}
		if !validSizes[params.SearchContextSize] {
			return NewValidationError("search_context_size", params.SearchContextSize,
				"must be one of: low, medium, high")
		}
	}

	// Validate reasoning effort
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
func (h *QueryHandler) executeStreaming(
	client *perplexity.Client,
	req *perplexity.CompletionRequest,
) (*perplexity.CompletionResponse, error) {
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

	err := client.SendSSEHTTPRequest(&wg, req, responseChannel)
	if err != nil {
		return nil, NewStreamError("streaming request failed", err)
	}

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
