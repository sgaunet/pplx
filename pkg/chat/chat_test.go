package chat

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sgaunet/perplexity-go/v2"
	"github.com/sgaunet/pplx/pkg/clerrors"
)

// mockCompletionResponseJSON returns a minimal valid completion response JSON.
func mockCompletionResponseJSON() string {
	return `{
		"id": "test-id",
		"model": "sonar",
		"created": 1234567890,
		"object": "chat.completion",
		"usage": {"prompt_tokens": 1, "completion_tokens": 1, "total_tokens": 2},
		"choices": [{"index": 0, "finish_reason": "stop",
			"message": {"role": "assistant", "content": "Hello"},
			"delta": {"role": "", "content": ""}}]
	}`
}

func TestNewChat(t *testing.T) {
	client := perplexity.NewClient("test-key")
	c := NewChat(client, "sonar", "system msg", 0.5, 100, 0.3, 0.7, 5, 0.9)

	if c.options.Model != "sonar" {
		t.Errorf("expected model 'sonar', got %q", c.options.Model)
	}
	if c.options.FrequencyPenalty != 0.5 {
		t.Errorf("expected frequency penalty 0.5, got %f", c.options.FrequencyPenalty)
	}
	if c.options.MaxTokens != 100 {
		t.Errorf("expected max tokens 100, got %d", c.options.MaxTokens)
	}
	if c.options.PresencePenalty != 0.3 {
		t.Errorf("expected presence penalty 0.3, got %f", c.options.PresencePenalty)
	}
	if c.options.Temperature != 0.7 {
		t.Errorf("expected temperature 0.7, got %f", c.options.Temperature)
	}
	if c.options.TopK != 5 {
		t.Errorf("expected top k 5, got %d", c.options.TopK)
	}
	if c.options.TopP != 0.9 {
		t.Errorf("expected top p 0.9, got %f", c.options.TopP)
	}
	if c.client == nil {
		t.Error("expected non-nil client")
	}
}

func TestNewChatWithOptions(t *testing.T) {
	client := perplexity.NewClient("test-key")
	opts := Options{
		Model:       "sonar",
		MaxTokens:   200,
		Temperature: 0.5,
	}
	c := NewChatWithOptions(client, "system msg", opts)

	if c.options.Model != "sonar" {
		t.Errorf("expected model 'sonar', got %q", c.options.Model)
	}
	if c.options.MaxTokens != 200 {
		t.Errorf("expected max tokens 200, got %d", c.options.MaxTokens)
	}
	if c.Messages.GetMessages() == nil {
		t.Error("expected messages to be initialized")
	}
}

func TestAddUserMessage(t *testing.T) {
	client := perplexity.NewClient("test-key")
	c := NewChatWithOptions(client, "", Options{Model: "sonar"})

	err := c.AddUserMessage("hello")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestAddAgentMessage(t *testing.T) {
	client := perplexity.NewClient("test-key")
	c := NewChatWithOptions(client, "", Options{Model: "sonar"})

	// Must add a user message first before agent message
	_ = c.AddUserMessage("hello")
	err := c.AddAgentMessage("response")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestRun_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockCompletionResponseJSON()))
	}))
	defer srv.Close()

	client := perplexity.NewClient("test-key")
	client.SetEndpoint(srv.URL)

	c := NewChatWithOptions(client, "system", Options{
		Model:            "sonar",
		MaxTokens:        100,
		TopP:             0.9,
		FrequencyPenalty: 1.0,
		Temperature:      0.7,
	})
	_ = c.AddUserMessage("test question")

	resp, err := c.Run()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.GetLastContent() != "Hello" {
		t.Errorf("expected content 'Hello', got %q", resp.GetLastContent())
	}
}

func TestRun_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := perplexity.NewClient("bad-key")
	client.SetEndpoint(srv.URL)

	c := NewChatWithOptions(client, "", Options{
		Model:            "sonar",
		MaxTokens:        100,
		TopP:             0.9,
		FrequencyPenalty: 1.0,
		Temperature:      0.7,
	})
	_ = c.AddUserMessage("test")

	_, err := c.Run()
	if err == nil {
		t.Fatal("expected error for 401 response, got nil")
	}
}

func TestAddSearchOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr bool
		errIs   error
	}{
		{
			name:    "valid recency day",
			opts:    Options{Model: "sonar", SearchRecency: "day"},
			wantErr: false,
		},
		{
			name:    "valid recency week",
			opts:    Options{Model: "sonar", SearchRecency: "week"},
			wantErr: false,
		},
		{
			name:    "valid recency month",
			opts:    Options{Model: "sonar", SearchRecency: "month"},
			wantErr: false,
		},
		{
			name:    "valid recency year",
			opts:    Options{Model: "sonar", SearchRecency: "year"},
			wantErr: false,
		},
		{
			name:    "valid recency hour",
			opts:    Options{Model: "sonar", SearchRecency: "hour"},
			wantErr: false,
		},
		{
			name:    "invalid recency",
			opts:    Options{Model: "sonar", SearchRecency: "century"},
			wantErr: true,
			errIs:   clerrors.ErrInvalidSearchRecency,
		},
		{
			name:    "search domains",
			opts:    Options{Model: "sonar", SearchDomains: []string{"example.com", "test.org"}},
			wantErr: false,
		},
		{
			name:    "location",
			opts:    Options{Model: "sonar", LocationLat: 48.8, LocationLon: 2.3, LocationCountry: "FR"},
			wantErr: false,
		},
		{
			name:    "empty recency is valid",
			opts:    Options{Model: "sonar", SearchRecency: ""},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := perplexity.NewClient("test-key")
			c := NewChatWithOptions(client, "", tt.opts)

			var opts []perplexity.CompletionRequestOption
			err := c.addSearchOptions(&opts)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errIs != nil && !errors.Is(err, tt.errIs) {
					t.Errorf("expected error wrapping %v, got %v", tt.errIs, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}
		})
	}
}

func TestAddFormatOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr bool
		errIs   error
	}{
		{
			name: "valid JSON schema on sonar model",
			opts: Options{
				Model:                    "sonar",
				ResponseFormatJSONSchema: `{"type":"object","properties":{"name":{"type":"string"}}}`,
			},
			wantErr: false,
		},
		{
			name: "valid regex on sonar model",
			opts: Options{
				Model:               "sonar",
				ResponseFormatRegex: `\d{3}-\d{4}`,
			},
			wantErr: false,
		},
		{
			name: "both JSON schema and regex conflict",
			opts: Options{
				Model:                    "sonar",
				ResponseFormatJSONSchema: `{"type":"object"}`,
				ResponseFormatRegex:      `\d+`,
			},
			wantErr: true,
			errIs:   clerrors.ErrConflictingResponseFormats,
		},
		{
			name: "JSON schema on non-sonar model",
			opts: Options{
				Model:                    "gpt-4",
				ResponseFormatJSONSchema: `{"type":"object"}`,
			},
			wantErr: true,
			errIs:   clerrors.ErrResponseFormatNotSupported,
		},
		{
			name: "regex on non-sonar model",
			opts: Options{
				Model:               "gpt-4",
				ResponseFormatRegex: `\d+`,
			},
			wantErr: true,
			errIs:   clerrors.ErrResponseFormatNotSupported,
		},
		{
			name: "invalid JSON schema",
			opts: Options{
				Model:                    "sonar",
				ResponseFormatJSONSchema: `{not valid json`,
			},
			wantErr: true,
		},
		{
			name:    "no format options",
			opts:    Options{Model: "sonar"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := perplexity.NewClient("test-key")
			c := NewChatWithOptions(client, "", tt.opts)

			var opts []perplexity.CompletionRequestOption
			err := c.addFormatOptions(&opts)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errIs != nil && !errors.Is(err, tt.errIs) {
					t.Errorf("expected error wrapping %v, got %v", tt.errIs, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}
		})
	}
}

func TestAddModeOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr bool
		errIs   error
	}{
		{
			name:    "valid search mode web",
			opts:    Options{Model: "sonar", SearchMode: "web"},
			wantErr: false,
		},
		{
			name:    "valid search mode academic",
			opts:    Options{Model: "sonar", SearchMode: "academic"},
			wantErr: false,
		},
		{
			name:    "invalid search mode",
			opts:    Options{Model: "sonar", SearchMode: "images"},
			wantErr: true,
			errIs:   clerrors.ErrInvalidSearchMode,
		},
		{
			name:    "valid context size low",
			opts:    Options{Model: "sonar", SearchContextSize: "low"},
			wantErr: false,
		},
		{
			name:    "valid context size medium",
			opts:    Options{Model: "sonar", SearchContextSize: "medium"},
			wantErr: false,
		},
		{
			name:    "valid context size high",
			opts:    Options{Model: "sonar", SearchContextSize: "high"},
			wantErr: false,
		},
		{
			name:    "invalid context size",
			opts:    Options{Model: "sonar", SearchContextSize: "huge"},
			wantErr: true,
			errIs:   clerrors.ErrInvalidSearchContextSize,
		},
		{
			name:    "empty mode and context",
			opts:    Options{Model: "sonar"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := perplexity.NewClient("test-key")
			c := NewChatWithOptions(client, "", tt.opts)

			var opts []perplexity.CompletionRequestOption
			err := c.addModeOptions(&opts)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errIs != nil && !errors.Is(err, tt.errIs) {
					t.Errorf("expected error wrapping %v, got %v", tt.errIs, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}
		})
	}
}

func TestAddDateOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr bool
		errIs   error
	}{
		{
			name:    "valid search after date",
			opts:    Options{Model: "sonar", SearchAfterDate: "01/15/2024"},
			wantErr: false,
		},
		{
			name:    "valid search before date",
			opts:    Options{Model: "sonar", SearchBeforeDate: "12/31/2024"},
			wantErr: false,
		},
		{
			name:    "valid last updated after",
			opts:    Options{Model: "sonar", LastUpdatedAfter: "06/01/2024"},
			wantErr: false,
		},
		{
			name:    "valid last updated before",
			opts:    Options{Model: "sonar", LastUpdatedBefore: "09/30/2024"},
			wantErr: false,
		},
		{
			name:    "invalid search after date format",
			opts:    Options{Model: "sonar", SearchAfterDate: "2024-01-15"},
			wantErr: true,
			errIs:   clerrors.ErrInvalidSearchAfterDate,
		},
		{
			name:    "invalid search before date format",
			opts:    Options{Model: "sonar", SearchBeforeDate: "not-a-date"},
			wantErr: true,
			errIs:   clerrors.ErrInvalidSearchBeforeDate,
		},
		{
			name:    "invalid last updated after format",
			opts:    Options{Model: "sonar", LastUpdatedAfter: "15/01/2024"},
			wantErr: true,
			errIs:   clerrors.ErrInvalidLastUpdatedAfter,
		},
		{
			name:    "invalid last updated before format",
			opts:    Options{Model: "sonar", LastUpdatedBefore: "2024/01/15"},
			wantErr: true,
			errIs:   clerrors.ErrInvalidLastUpdatedBefore,
		},
		{
			name:    "all dates valid",
			opts:    Options{Model: "sonar", SearchAfterDate: "01/01/2024", SearchBeforeDate: "12/31/2024", LastUpdatedAfter: "03/01/2024", LastUpdatedBefore: "06/30/2024"},
			wantErr: false,
		},
		{
			name:    "no dates",
			opts:    Options{Model: "sonar"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := perplexity.NewClient("test-key")
			c := NewChatWithOptions(client, "", tt.opts)

			var opts []perplexity.CompletionRequestOption
			err := c.addDateOptions(&opts)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errIs != nil && !errors.Is(err, tt.errIs) {
					t.Errorf("expected error wrapping %v, got %v", tt.errIs, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}
		})
	}
}

func TestAddResearchOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		wantErr bool
		errIs   error
	}{
		{
			name:    "valid effort low",
			opts:    Options{Model: "sonar-deep-research", ReasoningEffort: "low"},
			wantErr: false,
		},
		{
			name:    "valid effort medium",
			opts:    Options{Model: "sonar-deep-research", ReasoningEffort: "medium"},
			wantErr: false,
		},
		{
			name:    "valid effort high",
			opts:    Options{Model: "sonar-deep-research", ReasoningEffort: "high"},
			wantErr: false,
		},
		{
			name:    "invalid effort",
			opts:    Options{Model: "sonar-deep-research", ReasoningEffort: "extreme"},
			wantErr: true,
			errIs:   clerrors.ErrInvalidReasoningEffort,
		},
		{
			name:    "non-deep-research model warns but succeeds",
			opts:    Options{Model: "sonar", ReasoningEffort: "high"},
			wantErr: false,
		},
		{
			name:    "empty effort",
			opts:    Options{Model: "sonar"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := perplexity.NewClient("test-key")
			c := NewChatWithOptions(client, "", tt.opts)

			var opts []perplexity.CompletionRequestOption
			err := c.addResearchOptions(&opts)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errIs != nil && !errors.Is(err, tt.errIs) {
					t.Errorf("expected error wrapping %v, got %v", tt.errIs, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			}
		})
	}
}

func TestAddResponseOptions(t *testing.T) {
	t.Run("return images true", func(t *testing.T) {
		client := perplexity.NewClient("test-key")
		c := NewChatWithOptions(client, "", Options{Model: "sonar", ReturnImages: true})
		var opts []perplexity.CompletionRequestOption
		c.addResponseOptions(&opts)
		if len(opts) != 1 {
			t.Errorf("expected 1 option for ReturnImages=true, got %d", len(opts))
		}
	})

	t.Run("return related true", func(t *testing.T) {
		client := perplexity.NewClient("test-key")
		c := NewChatWithOptions(client, "", Options{Model: "sonar", ReturnRelated: true})
		var opts []perplexity.CompletionRequestOption
		c.addResponseOptions(&opts)
		if len(opts) != 1 {
			t.Errorf("expected 1 option for ReturnRelated=true, got %d", len(opts))
		}
	})

	t.Run("both false", func(t *testing.T) {
		client := perplexity.NewClient("test-key")
		c := NewChatWithOptions(client, "", Options{Model: "sonar"})
		var opts []perplexity.CompletionRequestOption
		c.addResponseOptions(&opts)
		if len(opts) != 0 {
			t.Errorf("expected 0 options for defaults, got %d", len(opts))
		}
	})
}

func TestAddImageOptions(t *testing.T) {
	t.Run("image domains", func(t *testing.T) {
		client := perplexity.NewClient("test-key")
		c := NewChatWithOptions(client, "", Options{
			Model:        "sonar",
			ImageDomains: []string{"example.com"},
		})
		var opts []perplexity.CompletionRequestOption
		c.addImageOptions(&opts)
		if len(opts) != 1 {
			t.Errorf("expected 1 option for ImageDomains, got %d", len(opts))
		}
	})

	t.Run("image formats", func(t *testing.T) {
		client := perplexity.NewClient("test-key")
		c := NewChatWithOptions(client, "", Options{
			Model:        "sonar",
			ImageFormats: []string{"png", "jpg"},
		})
		var opts []perplexity.CompletionRequestOption
		c.addImageOptions(&opts)
		if len(opts) != 1 {
			t.Errorf("expected 1 option for ImageFormats, got %d", len(opts))
		}
	})

	t.Run("empty", func(t *testing.T) {
		client := perplexity.NewClient("test-key")
		c := NewChatWithOptions(client, "", Options{Model: "sonar"})
		var opts []perplexity.CompletionRequestOption
		c.addImageOptions(&opts)
		if len(opts) != 0 {
			t.Errorf("expected 0 options for empty image opts, got %d", len(opts))
		}
	})
}

func TestBuildRequestOptions_Defaults(t *testing.T) {
	client := perplexity.NewClient("test-key")
	c := NewChatWithOptions(client, "sys", Options{Model: "sonar"})
	_ = c.AddUserMessage("test")

	opts, err := c.buildRequestOptions()
	if err != nil {
		t.Fatalf("expected no error with defaults, got %v", err)
	}
	if len(opts) == 0 {
		t.Error("expected non-empty options slice")
	}
}

func TestBuildRequestOptions_AllOptions(t *testing.T) {
	client := perplexity.NewClient("test-key")
	opts := Options{
		Model:             "sonar",
		FrequencyPenalty:  0.5,
		MaxTokens:         100,
		PresencePenalty:   0.3,
		Temperature:       0.7,
		TopK:              5,
		TopP:              0.9,
		SearchDomains:     []string{"example.com"},
		SearchRecency:     "week",
		LocationLat:       48.8,
		LocationLon:       2.3,
		LocationCountry:   "FR",
		ReturnImages:      true,
		ReturnRelated:     true,
		ImageDomains:      []string{"img.example.com"},
		ImageFormats:      []string{"png"},
		SearchMode:        "web",
		SearchContextSize: "medium",
		SearchAfterDate:   "01/01/2024",
		SearchBeforeDate:  "12/31/2024",
		LastUpdatedAfter:  "03/01/2024",
		LastUpdatedBefore: "06/30/2024",
	}
	c := NewChatWithOptions(client, "system", opts)
	_ = c.AddUserMessage("test")

	result, err := c.buildRequestOptions()
	if err != nil {
		t.Fatalf("expected no error with all options, got %v", err)
	}
	if len(result) == 0 {
		t.Error("expected non-empty options slice")
	}
}

func TestBuildRequestOptions_InvalidRecency(t *testing.T) {
	client := perplexity.NewClient("test-key")
	c := NewChatWithOptions(client, "", Options{
		Model:         "sonar",
		SearchRecency: "invalid",
	})
	_ = c.AddUserMessage("test")

	_, err := c.buildRequestOptions()
	if err == nil {
		t.Fatal("expected error for invalid recency")
	}
	if !errors.Is(err, clerrors.ErrInvalidSearchRecency) {
		t.Errorf("expected ErrInvalidSearchRecency, got %v", err)
	}
}

func TestBuildRequestOptions_InvalidSearchMode(t *testing.T) {
	client := perplexity.NewClient("test-key")
	c := NewChatWithOptions(client, "", Options{
		Model:      "sonar",
		SearchMode: "invalid",
	})
	_ = c.AddUserMessage("test")

	_, err := c.buildRequestOptions()
	if err == nil {
		t.Fatal("expected error for invalid search mode")
	}
	if !errors.Is(err, clerrors.ErrInvalidSearchMode) {
		t.Errorf("expected ErrInvalidSearchMode, got %v", err)
	}
}

func TestBuildRequestOptions_InvalidDate(t *testing.T) {
	client := perplexity.NewClient("test-key")
	c := NewChatWithOptions(client, "", Options{
		Model:           "sonar",
		SearchAfterDate: "not-a-date",
	})
	_ = c.AddUserMessage("test")

	_, err := c.buildRequestOptions()
	if err == nil {
		t.Fatal("expected error for invalid date")
	}
	if !errors.Is(err, clerrors.ErrInvalidSearchAfterDate) {
		t.Errorf("expected ErrInvalidSearchAfterDate, got %v", err)
	}
}

func TestBuildRequestOptions_ConflictingFormats(t *testing.T) {
	client := perplexity.NewClient("test-key")
	c := NewChatWithOptions(client, "", Options{
		Model:                    "sonar",
		ResponseFormatJSONSchema: `{"type":"object"}`,
		ResponseFormatRegex:      `\d+`,
	})
	_ = c.AddUserMessage("test")

	_, err := c.buildRequestOptions()
	if err == nil {
		t.Fatal("expected error for conflicting formats")
	}
	if !errors.Is(err, clerrors.ErrConflictingResponseFormats) {
		t.Errorf("expected ErrConflictingResponseFormats, got %v", err)
	}
}
