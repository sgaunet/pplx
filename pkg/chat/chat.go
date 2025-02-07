package chat

import (
	"fmt"

	"github.com/sgaunet/perplexity-go/v2"
)

type Chat struct {
	Messages         perplexity.Messages
	client           *perplexity.Client
	model            string
	frequencyPenalty float64
	maxTokens        int
	presencePenalty  float64
	temperature      float64
	topK             int
	topP             float64
}

func NewChat(client *perplexity.Client, model string, systemMessage string,
	frequencyPenalty float64, maxTokens int, presencePenalty float64,
	temperature float64, topK int, topP float64,
) *Chat {
	msg := perplexity.NewMessages(perplexity.WithSystemMessage(systemMessage))
	return &Chat{
		Messages:         msg,
		client:           client,
		model:            model,
		frequencyPenalty: frequencyPenalty,
		maxTokens:        maxTokens,
		presencePenalty:  presencePenalty,
		temperature:      temperature,
		topK:             topK,
		topP:             topP,
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
	req := perplexity.NewCompletionRequest(perplexity.WithMessages(c.Messages.GetMessages()),
		perplexity.WithModel(c.model), perplexity.WithFrequencyPenalty(c.frequencyPenalty),
		perplexity.WithMaxTokens(c.maxTokens), perplexity.WithPresencePenalty(c.presencePenalty),
		perplexity.WithTemperature(c.temperature), perplexity.WithTopK(c.topK),
		perplexity.WithTopP(c.topP))
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
