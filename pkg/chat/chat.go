package chat

import (
	"fmt"
	"os"
	"time"

	"github.com/sgaunet/perplexity-go/v2"
)

type Chat struct {
	Messages perplexity.Messages
	client   *perplexity.Client
}

func NewChat(systemMessage string) *Chat {
	msg := perplexity.NewMessages(perplexity.WithSystemMessage(systemMessage))
	client := perplexity.NewClient(os.Getenv("PPLX_API_KEY"))
	client.SetHTTPTimeout(30 * time.Second)
	return &Chat{
		Messages: msg,
		client:   client,
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
	req := perplexity.NewCompletionRequest(perplexity.WithMessages(c.Messages.GetMessages()))
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
