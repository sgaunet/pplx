package chat

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sgaunet/perplexity-go/v2"
)

type ChatOptions struct {
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

type Chat struct {
	Messages perplexity.Messages
	client   *perplexity.Client
	options  ChatOptions
}

func NewChat(client *perplexity.Client, model string, systemMessage string,
	frequencyPenalty float64, maxTokens int, presencePenalty float64,
	temperature float64, topK int, topP float64,
) *Chat {
	// Create options from individual parameters for backward compatibility
	options := ChatOptions{
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

func NewChatWithOptions(client *perplexity.Client, systemMessage string, options ChatOptions) *Chat {
	msg := perplexity.NewMessages(perplexity.WithSystemMessage(systemMessage))
	return &Chat{
		Messages: msg,
		client:   client,
		options:  options,
	}
}

func (c *Chat) AddUserMessage(message string) error {
	err := c.Messages.AddUserMessage(message)
	if err != nil {
		return fmt.Errorf("error adding user message: %w", err)
	}
	return nil
}

func (c *Chat) AddAgentMessage(message string) error {
	err := c.Messages.AddAgentMessage(message)
	if err != nil {
		return fmt.Errorf("error adding agent message: %w", err)
	}
	return nil
}

func (c *Chat) Run() (*perplexity.CompletionResponse, error) {
	// Build options list
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

	// Add search/web options if provided
	if len(c.options.SearchDomains) > 0 {
		opts = append(opts, perplexity.WithSearchDomainFilter(c.options.SearchDomains))
	}
	if c.options.SearchRecency != "" {
		// Validate search recency
		validRecency := map[string]bool{"day": true, "week": true, "month": true, "year": true}
		if !validRecency[c.options.SearchRecency] {
			return nil, fmt.Errorf("invalid search-recency value '%s'. Must be one of: day, week, month, year", c.options.SearchRecency)
		}
		opts = append(opts, perplexity.WithSearchRecencyFilter(c.options.SearchRecency))
	}
	if c.options.LocationLat != 0 || c.options.LocationLon != 0 || c.options.LocationCountry != "" {
		opts = append(opts, perplexity.WithUserLocation(c.options.LocationLat, c.options.LocationLon, c.options.LocationCountry))
	}

	// Add response enhancement options
	if c.options.ReturnImages {
		opts = append(opts, perplexity.WithReturnImages(c.options.ReturnImages))
	}
	if c.options.ReturnRelated {
		opts = append(opts, perplexity.WithReturnRelatedQuestions(c.options.ReturnRelated))
	}
	// Note: Stream option is handled separately in the chat command

	// Add image filtering options
	if len(c.options.ImageDomains) > 0 {
		opts = append(opts, perplexity.WithImageDomainFilter(c.options.ImageDomains))
	}
	if len(c.options.ImageFormats) > 0 {
		opts = append(opts, perplexity.WithImageFormatFilter(c.options.ImageFormats))
	}

	// Add response format options
	if c.options.ResponseFormatJSONSchema != "" && c.options.ResponseFormatRegex != "" {
		return nil, fmt.Errorf("cannot use both JSON schema and regex response formats")
	}
	if c.options.ResponseFormatJSONSchema != "" || c.options.ResponseFormatRegex != "" {
		// Validate model supports response formats
		if !strings.HasPrefix(c.options.Model, "sonar") {
			return nil, fmt.Errorf("response formats (JSON schema and regex) are only supported by sonar models")
		}
	}
	if c.options.ResponseFormatJSONSchema != "" {
		// Parse JSON schema
		var schema interface{}
		err := json.Unmarshal([]byte(c.options.ResponseFormatJSONSchema), &schema)
		if err != nil {
			return nil, fmt.Errorf("invalid JSON schema: %w", err)
		}
		opts = append(opts, perplexity.WithJSONSchemaResponseFormat(schema))
	}
	if c.options.ResponseFormatRegex != "" {
		opts = append(opts, perplexity.WithRegexResponseFormat(c.options.ResponseFormatRegex))
	}

	// Add search mode options
	if c.options.SearchMode != "" {
		// Validate search mode
		validModes := map[string]bool{"web": true, "academic": true}
		if !validModes[c.options.SearchMode] {
			return nil, fmt.Errorf("invalid search mode '%s'. Must be one of: web, academic", c.options.SearchMode)
		}
		opts = append(opts, perplexity.WithSearchMode(c.options.SearchMode))
	}
	if c.options.SearchContextSize != "" {
		// Validate search context size
		validSizes := map[string]bool{"low": true, "medium": true, "high": true}
		if !validSizes[c.options.SearchContextSize] {
			return nil, fmt.Errorf("invalid search context size '%s'. Must be one of: low, medium, high", c.options.SearchContextSize)
		}
		opts = append(opts, perplexity.WithSearchContextSize(c.options.SearchContextSize))
	}

	// Add date filtering options
	if c.options.SearchAfterDate != "" {
		date, err := time.Parse("01/02/2006", c.options.SearchAfterDate)
		if err != nil {
			return nil, fmt.Errorf("invalid search-after-date format '%s'. Use MM/DD/YYYY", c.options.SearchAfterDate)
		}
		opts = append(opts, perplexity.WithSearchAfterDateFilter(date))
	}
	if c.options.SearchBeforeDate != "" {
		date, err := time.Parse("01/02/2006", c.options.SearchBeforeDate)
		if err != nil {
			return nil, fmt.Errorf("invalid search-before-date format '%s'. Use MM/DD/YYYY", c.options.SearchBeforeDate)
		}
		opts = append(opts, perplexity.WithSearchBeforeDateFilter(date))
	}
	if c.options.LastUpdatedAfter != "" {
		date, err := time.Parse("01/02/2006", c.options.LastUpdatedAfter)
		if err != nil {
			return nil, fmt.Errorf("invalid last-updated-after format '%s'. Use MM/DD/YYYY", c.options.LastUpdatedAfter)
		}
		opts = append(opts, perplexity.WithLastUpdatedAfterFilter(date))
	}
	if c.options.LastUpdatedBefore != "" {
		date, err := time.Parse("01/02/2006", c.options.LastUpdatedBefore)
		if err != nil {
			return nil, fmt.Errorf("invalid last-updated-before format '%s'. Use MM/DD/YYYY", c.options.LastUpdatedBefore)
		}
		opts = append(opts, perplexity.WithLastUpdatedBeforeFilter(date))
	}

	// Add deep research options
	if c.options.ReasoningEffort != "" {
		// Validate reasoning effort
		validEfforts := map[string]bool{"low": true, "medium": true, "high": true}
		if !validEfforts[c.options.ReasoningEffort] {
			return nil, fmt.Errorf("invalid reasoning effort '%s'. Must be one of: low, medium, high", c.options.ReasoningEffort)
		}
		// Check if the model supports reasoning effort
		if !strings.Contains(c.options.Model, "deep-research") {
			// Just a warning in chat mode, not an error
			fmt.Printf("Warning: reasoning-effort is only supported by sonar-deep-research model\n")
		}
		opts = append(opts, perplexity.WithReasoningEffort(c.options.ReasoningEffort))
	}

	req := perplexity.NewCompletionRequest(opts...)
	err := req.Validate()
	if err != nil {
		return nil, fmt.Errorf("error validating completion request: %w", err)
	}

	res, err := c.client.SendCompletionRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error sending completion request: %w", err)
	}
	return res, nil
}
