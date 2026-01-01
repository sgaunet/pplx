package mcp

import (
	"time"

	"github.com/sgaunet/perplexity-go/v2"
)

// QueryParams contains all parameters for a Perplexity query.
type QueryParams struct {
	// Required parameter
	UserPrompt string

	// Core parameters
	SystemPrompt     string
	Model            string
	FrequencyPenalty float64
	MaxTokens        int
	PresencePenalty  float64
	Temperature      float64
	TopK             int
	TopP             float64
	Timeout          time.Duration

	// Search/Web options
	SearchDomains   []string
	SearchRecency   string
	LocationLat     float64
	LocationLon     float64
	LocationCountry string

	// Response enhancement options
	ReturnImages  bool
	ReturnRelated bool
	Stream        bool

	// Image filtering options
	ImageDomains []string
	ImageFormats []string

	// Response format options
	ResponseFormatJSONSchema string
	ResponseFormatRegex      string

	// Search mode options
	SearchMode        string
	SearchContextSize string

	// Date filtering options (MM/DD/YYYY format)
	SearchAfterDate   string
	SearchBeforeDate  string
	LastUpdatedAfter  string
	LastUpdatedBefore string

	// Deep research options
	ReasoningEffort string
}

// ParameterExtractor extracts and validates MCP tool parameters.
type ParameterExtractor struct{}

// NewParameterExtractor creates a new parameter extractor.
func NewParameterExtractor() *ParameterExtractor {
	return &ParameterExtractor{}
}

// Extract converts raw MCP arguments to typed QueryParams.
func (e *ParameterExtractor) Extract(args map[string]any) (*QueryParams, error) {
	// Extract required user_prompt
	userPrompt, ok := args["user_prompt"].(string)
	if !ok || userPrompt == "" {
		return nil, NewParameterError("user_prompt", args["user_prompt"], "must be a non-empty string")
	}

	params := &QueryParams{
		UserPrompt: userPrompt,
	}

	// Extract optional parameters
	params.SystemPrompt = e.extractString(args, "system_prompt", "")
	params.Model = e.extractString(args, "model", "")
	params.FrequencyPenalty = e.extractFloat(args, "frequency_penalty", 0)
	params.MaxTokens = e.extractInt(args, "max_tokens", 0)
	params.PresencePenalty = e.extractFloat(args, "presence_penalty", 0)
	params.Temperature = e.extractFloat(args, "temperature", 0)
	params.TopK = e.extractInt(args, "top_k", 0)
	params.TopP = e.extractFloat(args, "top_p", 0)

	// Timeout is special - convert from seconds to Duration
	timeoutSeconds := e.extractFloat(args, "timeout", 0)
	if timeoutSeconds > 0 {
		params.Timeout = time.Duration(timeoutSeconds) * time.Second
	}

	// Search/Web options
	params.SearchDomains = e.extractStringSlice(args, "search_domains")
	params.SearchRecency = e.extractString(args, "search_recency", "")
	params.LocationLat = e.extractFloat(args, "location_lat", 0)
	params.LocationLon = e.extractFloat(args, "location_lon", 0)
	params.LocationCountry = e.extractString(args, "location_country", "")

	// Response enhancement options
	params.ReturnImages = e.extractBool(args, "return_images", false)
	params.ReturnRelated = e.extractBool(args, "return_related", false)
	params.Stream = e.extractBool(args, "stream", false)

	// Image filtering options
	params.ImageDomains = e.extractStringSlice(args, "image_domains")
	params.ImageFormats = e.extractStringSlice(args, "image_formats")

	// Response format options
	params.ResponseFormatJSONSchema = e.extractString(args, "response_format_json_schema", "")
	params.ResponseFormatRegex = e.extractString(args, "response_format_regex", "")

	// Search mode options
	params.SearchMode = e.extractString(args, "search_mode", "")
	params.SearchContextSize = e.extractString(args, "search_context_size", "")

	// Date filtering options
	params.SearchAfterDate = e.extractString(args, "search_after_date", "")
	params.SearchBeforeDate = e.extractString(args, "search_before_date", "")
	params.LastUpdatedAfter = e.extractString(args, "last_updated_after", "")
	params.LastUpdatedBefore = e.extractString(args, "last_updated_before", "")

	// Deep research options
	params.ReasoningEffort = e.extractString(args, "reasoning_effort", "")

	// Apply default values from perplexity-go library
	e.applyDefaults(params)

	return params, nil
}

// extractString safely extracts a string parameter with a default value.
//nolint:unparam // defaultVal is kept for API consistency even though currently always ""
func (e *ParameterExtractor) extractString(args map[string]any, key string, defaultVal string) string {
	if val, ok := args[key].(string); ok {
		return val
	}
	return defaultVal
}

// extractFloat safely extracts a float64 parameter with a default value.
//nolint:unparam // defaultVal is kept for API consistency even though currently always 0
func (e *ParameterExtractor) extractFloat(args map[string]any, key string, defaultVal float64) float64 {
	if val, ok := args[key].(float64); ok {
		return val
	}
	return defaultVal
}

// extractInt safely extracts an int parameter (converting from float64) with a default value.
func (e *ParameterExtractor) extractInt(args map[string]any, key string, defaultVal int) int {
	if val, ok := args[key].(float64); ok {
		return int(val)
	}
	return defaultVal
}

// extractBool safely extracts a boolean parameter with a default value.
func (e *ParameterExtractor) extractBool(args map[string]any, key string, defaultVal bool) bool {
	if val, ok := args[key].(bool); ok {
		return val
	}
	return defaultVal
}

// extractStringSlice safely extracts a string slice parameter.
func (e *ParameterExtractor) extractStringSlice(args map[string]any, key string) []string {
	if val, ok := args[key].([]any); ok {
		result := make([]string, 0, len(val))
		for _, item := range val {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	}
	return nil
}

// applyDefaults applies default values from the perplexity-go library.
func (e *ParameterExtractor) applyDefaults(params *QueryParams) {
	if params.Model == "" {
		params.Model = perplexity.DefaultModel
	}
	if params.FrequencyPenalty == 0 {
		params.FrequencyPenalty = perplexity.DefaultFrequencyPenalty
	}
	if params.MaxTokens == 0 {
		params.MaxTokens = perplexity.DefaultMaxTokens
	}
	if params.PresencePenalty == 0 {
		params.PresencePenalty = perplexity.DefaultPresencePenalty
	}
	if params.Temperature == 0 {
		params.Temperature = perplexity.DefaultTemperature
	}
	if params.TopK == 0 {
		params.TopK = perplexity.DefaultTopK
	}
	if params.TopP == 0 {
		params.TopP = perplexity.DefaultTopP
	}
	if params.Timeout == 0 {
		params.Timeout = perplexity.DefaultTimeout
	}
}
