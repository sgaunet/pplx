// Package chat provides chat functionality for interacting with the Perplexity API.
package chat

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sgaunet/perplexity-go/v2"
	"github.com/sgaunet/pplx/pkg/clerrors"
	"github.com/sgaunet/pplx/pkg/logger"
)

// Options contains all configuration options for a chat session.
type Options struct {
	Model            string
	FrequencyPenalty float64
	MaxTokens        int
	PresencePenalty  float64
	Temperature      float64
	TopK             int
	TopP             float64
	SearchDomains    []string
	SearchRecency    string
	LocationLat      float64
	LocationLon      float64
	LocationCountry  string
	ReturnImages     bool
	ReturnRelated    bool
	Stream           bool
	ImageDomains     []string
	ImageFormats     []string
	
	// Response format options
	ResponseFormatJSONSchema string
	ResponseFormatRegex      string
	
	// Search mode options
	SearchMode        string
	SearchContextSize string
	
	// Date filtering options
	SearchAfterDate   string
	SearchBeforeDate  string
	LastUpdatedAfter  string
	LastUpdatedBefore string
	
	// Deep research options
	ReasoningEffort string
}

// Chat represents a chat session with the Perplexity API.
type Chat struct {
	Messages perplexity.Messages
	client   *perplexity.Client
	options  Options
}

// NewChat creates a new chat instance with individual parameters for backward compatibility.
func NewChat(client *perplexity.Client, model string, systemMessage string,
	frequencyPenalty float64, maxTokens int, presencePenalty float64,
	temperature float64, topK int, topP float64,
) *Chat {
	// Create options from individual parameters for backward compatibility
	options := Options{
		Model:            model,
		FrequencyPenalty: frequencyPenalty,
		MaxTokens:        maxTokens,
		PresencePenalty:  presencePenalty,
		Temperature:      temperature,
		TopK:             topK,
		TopP:             topP,
	}
	return NewChatWithOptions(client, systemMessage, options)
}

// NewChatWithOptions creates a new chat instance with structured options.
func NewChatWithOptions(client *perplexity.Client, systemMessage string, options Options) *Chat {
	msg := perplexity.NewMessages(perplexity.WithSystemMessage(systemMessage))
	return &Chat{
		Messages: msg,
		client:   client,
		options:  options,
	}
}

// AddUserMessage adds a user message to the chat conversation.
func (c *Chat) AddUserMessage(message string) error {
	err := c.Messages.AddUserMessage(message)
	if err != nil {
		return fmt.Errorf("error adding user message: %w", err)
	}
	return nil
}

// AddAgentMessage adds an agent message to the chat conversation.
func (c *Chat) AddAgentMessage(message string) error {
	err := c.Messages.AddAgentMessage(message)
	if err != nil {
		return fmt.Errorf("error adding agent message: %w", err)
	}
	return nil
}

// Run executes the chat request with the configured options.
func (c *Chat) Run() (*perplexity.CompletionResponse, error) {
	opts, err := c.buildRequestOptions()
	if err != nil {
		return nil, err
	}

	req := perplexity.NewCompletionRequest(opts...)
	err = req.Validate()
	if err != nil {
		return nil, fmt.Errorf("error validating completion request: %w", err)
	}

	res, err := c.client.SendCompletionRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error sending completion request: %w", err)
	}
	return res, nil
}

func (c *Chat) buildRequestOptions() ([]perplexity.CompletionRequestOption, error) {
	opts := []perplexity.CompletionRequestOption{
		perplexity.WithMessages(c.Messages.GetMessages()),
		perplexity.WithModel(c.options.Model),
		perplexity.WithFrequencyPenalty(c.options.FrequencyPenalty),
		perplexity.WithMaxTokens(c.options.MaxTokens),
		perplexity.WithPresencePenalty(c.options.PresencePenalty),
		perplexity.WithTemperature(c.options.Temperature),
		perplexity.WithTopK(c.options.TopK),
		perplexity.WithTopP(c.options.TopP),
	}

	if err := c.addSearchOptions(&opts); err != nil {
		return nil, err
	}
	c.addResponseOptions(&opts)
	c.addImageOptions(&opts)
	if err := c.addFormatOptions(&opts); err != nil {
		return nil, err
	}
	if err := c.addModeOptions(&opts); err != nil {
		return nil, err
	}
	if err := c.addDateOptions(&opts); err != nil {
		return nil, err
	}
	if err := c.addResearchOptions(&opts); err != nil {
		return nil, err
	}

	return opts, nil
}

// addSearchOptions adds search-related options to the completion request.
// Validates search recency against API-supported time windows.
func (c *Chat) addSearchOptions(opts *[]perplexity.CompletionRequestOption) error {
	if len(c.options.SearchDomains) > 0 {
		*opts = append(*opts, perplexity.WithSearchDomainFilter(c.options.SearchDomains))
	}
	if c.options.SearchRecency != "" {
		// Validation map: API only accepts these exact time window strings
		// Rationale: Centralized here rather than in validator package because
		// these constraints come from the Perplexity API specification, not our domain logic.
		// Keeping validation close to API call makes it easier to update when API changes.
		validRecency := map[string]bool{"day": true, "week": true, "month": true, "year": true, "hour": true}
		if !validRecency[c.options.SearchRecency] {
			return fmt.Errorf("%w: '%s'. Must be one of: day, week, month, year, hour",
				clerrors.ErrInvalidSearchRecency, c.options.SearchRecency)
		}
		*opts = append(*opts, perplexity.WithSearchRecencyFilter(c.options.SearchRecency))
	}
	if c.options.LocationLat != 0 || c.options.LocationLon != 0 || c.options.LocationCountry != "" {
		*opts = append(*opts, perplexity.WithUserLocation(
			c.options.LocationLat, c.options.LocationLon, c.options.LocationCountry))
	}
	return nil
}

func (c *Chat) addResponseOptions(opts *[]perplexity.CompletionRequestOption) {
	if c.options.ReturnImages {
		*opts = append(*opts, perplexity.WithReturnImages(c.options.ReturnImages))
	}
	if c.options.ReturnRelated {
		*opts = append(*opts, perplexity.WithReturnRelatedQuestions(c.options.ReturnRelated))
	}
}

func (c *Chat) addImageOptions(opts *[]perplexity.CompletionRequestOption) {
	if len(c.options.ImageDomains) > 0 {
		*opts = append(*opts, perplexity.WithImageDomainFilter(c.options.ImageDomains))
	}
	if len(c.options.ImageFormats) > 0 {
		*opts = append(*opts, perplexity.WithImageFormatFilter(c.options.ImageFormats))
	}
}

func (c *Chat) addFormatOptions(opts *[]perplexity.CompletionRequestOption) error {
	if c.options.ResponseFormatJSONSchema != "" && c.options.ResponseFormatRegex != "" {
		return clerrors.ErrConflictingResponseFormats
	}
	if c.options.ResponseFormatJSONSchema != "" || c.options.ResponseFormatRegex != "" {
		if !strings.HasPrefix(c.options.Model, "sonar") {
			return clerrors.ErrResponseFormatNotSupported
		}
	}
	if c.options.ResponseFormatJSONSchema != "" {
		var schema any
		err := json.Unmarshal([]byte(c.options.ResponseFormatJSONSchema), &schema)
		if err != nil {
			return fmt.Errorf("invalid JSON schema: %w", err)
		}
		*opts = append(*opts, perplexity.WithJSONSchemaResponseFormat(schema))
	}
	if c.options.ResponseFormatRegex != "" {
		*opts = append(*opts, perplexity.WithRegexResponseFormat(c.options.ResponseFormatRegex))
	}
	return nil
}

// addModeOptions adds search mode and context size options.
// Validates against API-supported mode and context size values.
func (c *Chat) addModeOptions(opts *[]perplexity.CompletionRequestOption) error {
	if c.options.SearchMode != "" {
		// Search mode validation: API supports two distinct search behaviors
		// "web" = general internet search (default, broader results)
		// "academic" = scholarly sources only (research papers, journals)
		validModes := map[string]bool{"web": true, "academic": true}
		if !validModes[c.options.SearchMode] {
			return fmt.Errorf("%w: '%s'. Must be one of: web, academic", clerrors.ErrInvalidSearchMode, c.options.SearchMode)
		}
		*opts = append(*opts, perplexity.WithSearchMode(c.options.SearchMode))
	}
	if c.options.SearchContextSize != "" {
		// Context size validation: controls how much search context to include in the model prompt
		// "low" = minimal context (faster, cheaper)
		// "medium" = balanced context (default)
		// "high" = maximum context (slower, more comprehensive)
		validSizes := map[string]bool{"low": true, "medium": true, "high": true}
		if !validSizes[c.options.SearchContextSize] {
			return fmt.Errorf("%w: '%s'. Must be one of: low, medium, high",
				clerrors.ErrInvalidSearchContextSize, c.options.SearchContextSize)
		}
		*opts = append(*opts, perplexity.WithSearchContextSize(c.options.SearchContextSize))
	}
	return nil
}

// addDateOptions adds date filter options for search results.
// Parses and validates dates in MM/DD/YYYY format (required by Perplexity API).
func (c *Chat) addDateOptions(opts *[]perplexity.CompletionRequestOption) error {
	// Date format: MM/DD/YYYY is required by Perplexity API
	// Using Go reference time "01/02/2006" (January 2, 2006 at 3:04:05 PM MST)
	// Examples: "12/31/2024", "01/01/2025"
	if c.options.SearchAfterDate != "" {
		date, err := time.Parse("01/02/2006", c.options.SearchAfterDate)
		if err != nil {
			return fmt.Errorf("%w: '%s'. Use MM/DD/YYYY", clerrors.ErrInvalidSearchAfterDate, c.options.SearchAfterDate)
		}
		*opts = append(*opts, perplexity.WithSearchAfterDateFilter(date))
	}
	if c.options.SearchBeforeDate != "" {
		date, err := time.Parse("01/02/2006", c.options.SearchBeforeDate)
		if err != nil {
			return fmt.Errorf("%w: '%s'. Use MM/DD/YYYY", clerrors.ErrInvalidSearchBeforeDate, c.options.SearchBeforeDate)
		}
		*opts = append(*opts, perplexity.WithSearchBeforeDateFilter(date))
	}
	if c.options.LastUpdatedAfter != "" {
		date, err := time.Parse("01/02/2006", c.options.LastUpdatedAfter)
		if err != nil {
			return fmt.Errorf("%w: '%s'. Use MM/DD/YYYY", clerrors.ErrInvalidLastUpdatedAfter, c.options.LastUpdatedAfter)
		}
		*opts = append(*opts, perplexity.WithLastUpdatedAfterFilter(date))
	}
	if c.options.LastUpdatedBefore != "" {
		date, err := time.Parse("01/02/2006", c.options.LastUpdatedBefore)
		if err != nil {
			return fmt.Errorf("%w: '%s'. Use MM/DD/YYYY", clerrors.ErrInvalidLastUpdatedBefore, c.options.LastUpdatedBefore)
		}
		*opts = append(*opts, perplexity.WithLastUpdatedBeforeFilter(date))
	}
	return nil
}

// addResearchOptions adds deep research model options.
// Validates reasoning effort level and warns if used with incompatible model.
func (c *Chat) addResearchOptions(opts *[]perplexity.CompletionRequestOption) error {
	if c.options.ReasoningEffort != "" {
		// Reasoning effort validation: controls depth of analysis for deep-research models
		// "low" = faster, less thorough (suitable for simple queries)
		// "medium" = balanced analysis (default)
		// "high" = maximum depth (slower, comprehensive for complex research tasks)
		validEfforts := map[string]bool{"low": true, "medium": true, "high": true}
		if !validEfforts[c.options.ReasoningEffort] {
			return fmt.Errorf("%w: '%s'. Must be one of: low, medium, high",
				clerrors.ErrInvalidReasoningEffort, c.options.ReasoningEffort)
		}
		// Warning instead of error: allows user to set reasoning-effort in config/flags
		// before switching models. Feature degrades gracefully (parameter ignored by API)
		// rather than failing hard. This improves UX when experimenting with different models.
		if !strings.Contains(c.options.Model, "deep-research") {
			logger.Warn("reasoning-effort only supported by sonar-deep-research model",
				"current_model", c.options.Model)
		}
		*opts = append(*opts, perplexity.WithReasoningEffort(c.options.ReasoningEffort))
	}
	return nil
}
