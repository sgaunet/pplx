package chat

import (
	"fmt"

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
