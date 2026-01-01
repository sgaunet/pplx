package mcp

import (
	"testing"
)

func TestBuildQueryTool(t *testing.T) {
	tool := BuildQueryTool()

	if tool == nil {
		t.Fatal("BuildQueryTool returned nil")
	}

	t.Run("has correct name", func(t *testing.T) {
		if tool.Name != "query" {
			t.Errorf("Expected tool name %q, got %q", "query", tool.Name)
		}
	})

	t.Run("has description", func(t *testing.T) {
		if tool.Description == "" {
			t.Error("Tool description should not be empty")
		}
		expectedDesc := "Query Perplexity AI with extensive search and filtering options"
		if tool.Description != expectedDesc {
			t.Errorf("Expected description %q, got %q", expectedDesc, tool.Description)
		}
	})

	t.Run("has all required parameters in schema", func(t *testing.T) {
		schema := tool.InputSchema
		if schema.Properties == nil {
			t.Fatal("Input schema properties should not be nil")
		}

		// user_prompt should be present
		if _, ok := schema.Properties["user_prompt"]; !ok {
			t.Error("Missing required parameter: user_prompt")
		}

		// user_prompt should be in required list
		found := false
		for _, req := range schema.Required {
			if req == "user_prompt" {
				found = true
				break
			}
		}
		if !found {
			t.Error("user_prompt should be marked as required")
		}
	})

	t.Run("has all optional parameters", func(t *testing.T) {
		schema := tool.InputSchema
		allParams := []string{
			"user_prompt", // required
			// Core parameters
			"system_prompt",
			"model",
			"frequency_penalty",
			"max_tokens",
			"presence_penalty",
			"temperature",
			"top_k",
			"top_p",
			"timeout",
			// Search/Web options
			"search_domains",
			"search_recency",
			"location_lat",
			"location_lon",
			"location_country",
			// Response enhancement
			"return_images",
			"return_related",
			"stream",
			// Image filtering
			"image_domains",
			"image_formats",
			// Response format
			"response_format_json_schema",
			"response_format_regex",
			// Search mode
			"search_mode",
			"search_context_size",
			// Date filtering
			"search_after_date",
			"search_before_date",
			"last_updated_after",
			"last_updated_before",
			// Deep research
			"reasoning_effort",
		}

		for _, param := range allParams {
			if _, ok := schema.Properties[param]; !ok {
				t.Errorf("Missing parameter: %s", param)
			}
		}
	})

	t.Run("has correct parameter count", func(t *testing.T) {
		schema := tool.InputSchema
		// We expect all 30 parameters from the list above to be present
		actualCount := len(schema.Properties)
		// Minimum expected is 29 (all parameters we defined)
		if actualCount < 29 {
			t.Errorf("Expected at least 29 parameters, got %d", actualCount)
		}
	})
}
