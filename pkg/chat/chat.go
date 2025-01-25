package chat

import (
	"fmt"

	"github.com/sgaunet/perplexity-go/v2"
)

type Chat struct {
	Messages perplexity.Messages
	client   *perplexity.Client
	model    string
}

func NewChat(client *perplexity.Client, model string, systemMessage string) *Chat {
	if model == "pro" {
		model = perplexity.ProModel
	} else {
		model = perplexity.DefaultModel
	}

	msg := perplexity.NewMessages(perplexity.WithSystemMessage(systemMessage))
	return &Chat{
		Messages: msg,
		client:   client,
		model:    model,
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
		perplexity.WithModel(c.model))
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
