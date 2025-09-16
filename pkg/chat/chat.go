// Package chat provides chat functionality for interacting with the Perplexity API.
package chat

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sgaunet/perplexity-go/v2"
)

// Error definitions for static error wrapping.
var (
	ErrInvalidSearchRecency         = errors.New("invalid search-recency value")
	ErrConflictingResponseFormats   = errors.New("cannot use both JSON schema and regex response formats")
	ErrResponseFormatNotSupported   = errors.New(
		"response formats (JSON schema and regex) are only supported by sonar models")
	ErrInvalidSearchMode           = errors.New("invalid search mode")
	ErrInvalidSearchContextSize    = errors.New("invalid search context size")
	ErrInvalidSearchAfterDate      = errors.New("invalid search-after-date format")
	ErrInvalidSearchBeforeDate     = errors.New("invalid search-before-date format")
	ErrInvalidLastUpdatedAfter     = errors.New("invalid last-updated-after format")
	ErrInvalidLastUpdatedBefore    = errors.New("invalid last-updated-before format")
	ErrInvalidReasoningEffort      = errors.New("invalid reasoning effort")
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

func (c *Chat) addSearchOptions(opts *[]perplexity.CompletionRequestOption) error {
	if len(c.options.SearchDomains) > 0 {
		*opts = append(*opts, perplexity.WithSearchDomainFilter(c.options.SearchDomains))
	}
	if c.options.SearchRecency != "" {
		validRecency := map[string]bool{"day": true, "week": true, "month": true, "year": true}
		if !validRecency[c.options.SearchRecency] {
			return fmt.Errorf("%w: '%s'. Must be one of: day, week, month, year",
				ErrInvalidSearchRecency, c.options.SearchRecency)
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
		return ErrConflictingResponseFormats
	}
	if c.options.ResponseFormatJSONSchema != "" || c.options.ResponseFormatRegex != "" {
		if !strings.HasPrefix(c.options.Model, "sonar") {
			return ErrResponseFormatNotSupported
		}
	}
	if c.options.ResponseFormatJSONSchema != "" {
		var schema interface{}
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

func (c *Chat) addModeOptions(opts *[]perplexity.CompletionRequestOption) error {
	if c.options.SearchMode != "" {
		validModes := map[string]bool{"web": true, "academic": true}
		if !validModes[c.options.SearchMode] {
			return fmt.Errorf("%w: '%s'. Must be one of: web, academic", ErrInvalidSearchMode, c.options.SearchMode)
		}
		*opts = append(*opts, perplexity.WithSearchMode(c.options.SearchMode))
	}
	if c.options.SearchContextSize != "" {
		validSizes := map[string]bool{"low": true, "medium": true, "high": true}
		if !validSizes[c.options.SearchContextSize] {
			return fmt.Errorf("%w: '%s'. Must be one of: low, medium, high",
				ErrInvalidSearchContextSize, c.options.SearchContextSize)
		}
		*opts = append(*opts, perplexity.WithSearchContextSize(c.options.SearchContextSize))
	}
	return nil
}

func (c *Chat) addDateOptions(opts *[]perplexity.CompletionRequestOption) error {
	if c.options.SearchAfterDate != "" {
		date, err := time.Parse("01/02/2006", c.options.SearchAfterDate)
		if err != nil {
			return fmt.Errorf("%w: '%s'. Use MM/DD/YYYY", ErrInvalidSearchAfterDate, c.options.SearchAfterDate)
		}
		*opts = append(*opts, perplexity.WithSearchAfterDateFilter(date))
	}
	if c.options.SearchBeforeDate != "" {
		date, err := time.Parse("01/02/2006", c.options.SearchBeforeDate)
		if err != nil {
			return fmt.Errorf("%w: '%s'. Use MM/DD/YYYY", ErrInvalidSearchBeforeDate, c.options.SearchBeforeDate)
		}
		*opts = append(*opts, perplexity.WithSearchBeforeDateFilter(date))
	}
	if c.options.LastUpdatedAfter != "" {
		date, err := time.Parse("01/02/2006", c.options.LastUpdatedAfter)
		if err != nil {
			return fmt.Errorf("%w: '%s'. Use MM/DD/YYYY", ErrInvalidLastUpdatedAfter, c.options.LastUpdatedAfter)
		}
		*opts = append(*opts, perplexity.WithLastUpdatedAfterFilter(date))
	}
	if c.options.LastUpdatedBefore != "" {
		date, err := time.Parse("01/02/2006", c.options.LastUpdatedBefore)
		if err != nil {
			return fmt.Errorf("%w: '%s'. Use MM/DD/YYYY", ErrInvalidLastUpdatedBefore, c.options.LastUpdatedBefore)
		}
		*opts = append(*opts, perplexity.WithLastUpdatedBeforeFilter(date))
	}
	return nil
}

func (c *Chat) addResearchOptions(opts *[]perplexity.CompletionRequestOption) error {
	if c.options.ReasoningEffort != "" {
		validEfforts := map[string]bool{"low": true, "medium": true, "high": true}
		if !validEfforts[c.options.ReasoningEffort] {
			return fmt.Errorf("%w: '%s'. Must be one of: low, medium, high",
				ErrInvalidReasoningEffort, c.options.ReasoningEffort)
		}
		if !strings.Contains(c.options.Model, "deep-research") {
			fmt.Printf("Warning: reasoning-effort is only supported by sonar-deep-research model\n")
		}
		*opts = append(*opts, perplexity.WithReasoningEffort(c.options.ReasoningEffort))
	}
	return nil
}
