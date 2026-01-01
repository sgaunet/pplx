package mcp

import (
	"context"
	"errors"
	"testing"

	"github.com/sgaunet/perplexity-go/v2"
)

func TestQueryHandler_ValidateParameters(t *testing.T) {
	handler := NewQueryHandler()

	tests := []struct {
		name      string
		params    QueryParams
		shouldErr bool
		errField  string
	}{
		{
			name: "valid parameters",
			params: QueryParams{
				UserPrompt:        "test",
				Model:             "sonar",
				SearchRecency:     "week",
				SearchMode:        "web",
				SearchContextSize: "high",
				ReasoningEffort:   "medium",
			},
			shouldErr: false,
		},
		{
			name: "invalid search_recency",
			params: QueryParams{
				UserPrompt:    "test",
				SearchRecency: "invalid",
			},
			shouldErr: true,
			errField:  "search_recency",
		},
		{
			name: "valid search_recency values",
			params: QueryParams{
				UserPrompt:    "test",
				SearchRecency: "day",
			},
			shouldErr: false,
		},
		{
			name: "conflicting response formats",
			params: QueryParams{
				UserPrompt:               "test",
				ResponseFormatJSONSchema: "{}",
				ResponseFormatRegex:      ".*",
			},
			shouldErr: true,
			errField:  "response_format",
		},
		{
			name: "response format with non-sonar model",
			params: QueryParams{
				UserPrompt:               "test",
				Model:                    "gpt-4",
				ResponseFormatJSONSchema: "{}",
			},
			shouldErr: true,
			errField:  "response_format",
		},
		{
			name: "response format with sonar model",
			params: QueryParams{
				UserPrompt:               "test",
				Model:                    "sonar-pro",
				ResponseFormatJSONSchema: "{}",
			},
			shouldErr: false,
		},
		{
			name: "invalid search_mode",
			params: QueryParams{
				UserPrompt: "test",
				SearchMode: "invalid",
			},
			shouldErr: true,
			errField:  "search_mode",
		},
		{
			name: "valid search_mode web",
			params: QueryParams{
				UserPrompt: "test",
				SearchMode: "web",
			},
			shouldErr: false,
		},
		{
			name: "valid search_mode academic",
			params: QueryParams{
				UserPrompt: "test",
				SearchMode: "academic",
			},
			shouldErr: false,
		},
		{
			name: "invalid search_context_size",
			params: QueryParams{
				UserPrompt:        "test",
				SearchContextSize: "invalid",
			},
			shouldErr: true,
			errField:  "search_context_size",
		},
		{
			name: "valid search_context_size",
			params: QueryParams{
				UserPrompt:        "test",
				SearchContextSize: "low",
			},
			shouldErr: false,
		},
		{
			name: "invalid reasoning_effort",
			params: QueryParams{
				UserPrompt:      "test",
				ReasoningEffort: "invalid",
			},
			shouldErr: true,
			errField:  "reasoning_effort",
		},
		{
			name: "valid reasoning_effort",
			params: QueryParams{
				UserPrompt:      "test",
				ReasoningEffort: "high",
			},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.validateParameters(tt.params)

			if tt.shouldErr {
				if err == nil {
					t.Error("Expected validation error, got nil")
					return
				}
				var valErr *ValidationError
				if errors.As(err, &valErr) {
					if valErr.Field != tt.errField {
						t.Errorf("Expected error field %q, got %q", tt.errField, valErr.Field)
					}
				} else {
					t.Errorf("Expected ValidationError, got %T: %v", err, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestQueryHandler_BuildRequestOptions(t *testing.T) {
	handler := NewQueryHandler()

	t.Run("builds basic options", func(t *testing.T) {
		params := QueryParams{
			UserPrompt:       "test",
			SystemPrompt:     "system",
			Model:            "sonar",
			Temperature:      0.7,
			MaxTokens:        100,
			FrequencyPenalty: 0.5,
		}

		msg := perplexity.NewMessages(perplexity.WithSystemMessage(params.SystemPrompt))
		_ = msg.AddUserMessage(params.UserPrompt)

		opts, err := handler.buildRequestOptions(params, msg)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(opts) < 8 {
			t.Errorf("Expected at least 8 basic options, got %d", len(opts))
		}
	})

	t.Run("builds options with search domains", func(t *testing.T) {
		params := QueryParams{
			UserPrompt:    "test",
			Model:         "sonar",
			SearchDomains: []string{"example.com", "test.com"},
		}

		msg := perplexity.NewMessages()
		_ = msg.AddUserMessage(params.UserPrompt)

		opts, err := handler.buildRequestOptions(params, msg)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Should have basic opts + search domains
		if len(opts) < 9 {
			t.Errorf("Expected at least 9 options with search domains, got %d", len(opts))
		}
	})

	t.Run("handles return_images with search_recency conflict", func(t *testing.T) {
		params := QueryParams{
			UserPrompt:    "test",
			Model:         "sonar",
			ReturnImages:  true,
			SearchRecency: "week",
		}

		msg := perplexity.NewMessages()
		_ = msg.AddUserMessage(params.UserPrompt)

		opts, err := handler.buildRequestOptions(params, msg)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Should handle the conflict by disabling search recency
		if len(opts) == 0 {
			t.Error("Expected options to be built")
		}
	})

	t.Run("builds options with image filters", func(t *testing.T) {
		params := QueryParams{
			UserPrompt:   "test",
			Model:        "sonar",
			ImageDomains: []string{"img.example.com"},
			ImageFormats: []string{"jpg", "png", "unsupported-format"},
		}

		msg := perplexity.NewMessages()
		_ = msg.AddUserMessage(params.UserPrompt)

		opts, err := handler.buildRequestOptions(params, msg)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(opts) == 0 {
			t.Error("Expected options to be built")
		}
	})

	t.Run("validates date formats", func(t *testing.T) {
		testCases := []struct {
			name      string
			dateField string
			dateValue string
		}{
			{"search_after_date", "SearchAfterDate", "invalid-date"},
			{"search_before_date", "SearchBeforeDate", "2024-01-01"},
			{"last_updated_after", "LastUpdatedAfter", "01/32/2024"},
			{"last_updated_before", "LastUpdatedBefore", "13/01/2024"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				params := QueryParams{
					UserPrompt: "test",
					Model:      "sonar",
				}

				switch tc.dateField {
				case "SearchAfterDate":
					params.SearchAfterDate = tc.dateValue
				case "SearchBeforeDate":
					params.SearchBeforeDate = tc.dateValue
				case "LastUpdatedAfter":
					params.LastUpdatedAfter = tc.dateValue
				case "LastUpdatedBefore":
					params.LastUpdatedBefore = tc.dateValue
				}

				msg := perplexity.NewMessages()
				_ = msg.AddUserMessage(params.UserPrompt)

				_, err := handler.buildRequestOptions(params, msg)
				if err == nil {
					t.Error("Expected date validation error")
				}

				var valErr *ValidationError
				if !errors.As(err, &valErr) {
					t.Errorf("Expected ValidationError, got %T", err)
				}
			})
		}
	})

	t.Run("parses valid dates", func(t *testing.T) {
		params := QueryParams{
			UserPrompt:        "test",
			Model:             "sonar",
			SearchAfterDate:   "01/15/2024",
			SearchBeforeDate:  "12/31/2024",
			LastUpdatedAfter:  "06/01/2024",
			LastUpdatedBefore: "06/30/2024",
		}

		msg := perplexity.NewMessages()
		_ = msg.AddUserMessage(params.UserPrompt)

		opts, err := handler.buildRequestOptions(params, msg)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Should have basic opts + 4 date filters
		if len(opts) < 12 {
			t.Errorf("Expected at least 12 options with date filters, got %d", len(opts))
		}
	})

	t.Run("validates JSON schema", func(t *testing.T) {
		params := QueryParams{
			UserPrompt:               "test",
			Model:                    "sonar",
			ResponseFormatJSONSchema: "invalid json",
		}

		msg := perplexity.NewMessages()
		_ = msg.AddUserMessage(params.UserPrompt)

		_, err := handler.buildRequestOptions(params, msg)
		if err == nil {
			t.Error("Expected JSON schema validation error")
		}

		var valErr *ValidationError
		if !errors.As(err, &valErr) {
			t.Errorf("Expected ValidationError, got %T", err)
		}
	})

	t.Run("accepts valid JSON schema", func(t *testing.T) {
		params := QueryParams{
			UserPrompt:               "test",
			Model:                    "sonar",
			ResponseFormatJSONSchema: `{"type": "object", "properties": {"name": {"type": "string"}}}`,
		}

		msg := perplexity.NewMessages()
		_ = msg.AddUserMessage(params.UserPrompt)

		opts, err := handler.buildRequestOptions(params, msg)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(opts) == 0 {
			t.Error("Expected options to be built")
		}
	})

	t.Run("builds options with regex response format", func(t *testing.T) {
		params := QueryParams{
			UserPrompt:          "test",
			Model:               "sonar",
			ResponseFormatRegex: `\d{3}-\d{3}-\d{4}`,
		}

		msg := perplexity.NewMessages()
		_ = msg.AddUserMessage(params.UserPrompt)

		opts, err := handler.buildRequestOptions(params, msg)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(opts) == 0 {
			t.Error("Expected options to be built")
		}
	})

	t.Run("builds options with all parameters", func(t *testing.T) {
		params := QueryParams{
			UserPrompt:        "test",
			SystemPrompt:      "system",
			Model:             "sonar-deep-research",
			Temperature:       0.7,
			MaxTokens:         100,
			SearchDomains:     []string{"example.com"},
			SearchRecency:     "week",
			ReturnImages:      false,
			ReturnRelated:     true,
			SearchMode:        "web",
			SearchContextSize: "high",
			ReasoningEffort:   "medium",
			LocationLat:       40.7128,
			LocationLon:       -74.0060,
			LocationCountry:   "US",
		}

		msg := perplexity.NewMessages(perplexity.WithSystemMessage(params.SystemPrompt))
		_ = msg.AddUserMessage(params.UserPrompt)

		opts, err := handler.buildRequestOptions(params, msg)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Should have many options with all parameters
		if len(opts) < 15 {
			t.Errorf("Expected at least 15 options with all parameters, got %d", len(opts))
		}
	})

	t.Run("validates parameters before building", func(t *testing.T) {
		params := QueryParams{
			UserPrompt:    "test",
			Model:         "sonar",
			SearchRecency: "invalid-recency",
		}

		msg := perplexity.NewMessages()
		_ = msg.AddUserMessage(params.UserPrompt)

		_, err := handler.buildRequestOptions(params, msg)
		if err == nil {
			t.Error("Expected validation error")
		}

		var valErr *ValidationError
		if !errors.As(err, &valErr) {
			t.Errorf("Expected ValidationError, got %T", err)
		}
	})
}

func TestNewQueryHandler(t *testing.T) {
	handler := NewQueryHandler()

	if handler == nil {
		t.Fatal("Expected handler, got nil")
	}

	if handler.clientFactory == nil {
		t.Error("Expected clientFactory to be set")
	}

	// Test that default factory creates a client
	client := handler.clientFactory("test-key")
	if client == nil {
		t.Error("Expected clientFactory to create a client")
	}
}

func TestQueryHandler_Handle_ValidationFlow(t *testing.T) {
	handler := NewQueryHandler()

	t.Run("validation error stops execution", func(t *testing.T) {
		params := QueryParams{
			UserPrompt:    "test",
			Model:         "sonar",
			SearchRecency: "invalid",
		}

		_, err := handler.Handle(context.Background(), "test-api-key", params)
		if err == nil {
			t.Fatal("Expected validation error, got nil")
		}

		var valErr *ValidationError
		if !errors.As(err, &valErr) {
			t.Errorf("Expected ValidationError, got %T", err)
		}
	})
}
