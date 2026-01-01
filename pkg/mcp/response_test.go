package mcp

import (
	"errors"
	"testing"

	"github.com/sgaunet/perplexity-go/v2"
)

func TestResponseFormatter_Format(t *testing.T) {
	formatter := NewResponseFormatter()

	t.Run("formats complete response", func(t *testing.T) {
		searchResults := []perplexity.SearchResult{{}}
		images := []perplexity.Image{{}}
		relatedQuestions := []string{"What is AI?"}

		response := &perplexity.CompletionResponse{
			Choices: []perplexity.Choice{
				{Message: perplexity.Message{Content: "test response"}},
			},
			Model:            "sonar",
			Usage:            perplexity.Usage{TotalTokens: 100},
			SearchResults:    &searchResults,
			Images:           &images,
			RelatedQuestions: &relatedQuestions,
		}

		result, err := formatter.Format(response)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.IsError {
			t.Error("Expected success result, got error")
		}
	})

	t.Run("formats minimal response", func(t *testing.T) {
		response := &perplexity.CompletionResponse{
			Choices: []perplexity.Choice{
				{Message: perplexity.Message{Content: "minimal response"}},
			},
			Model: "sonar",
			Usage: perplexity.Usage{TotalTokens: 50},
		}

		result, err := formatter.Format(response)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.IsError {
			t.Error("Expected success result, got error")
		}
	})

	t.Run("uses citations fallback when search_results empty", func(t *testing.T) {
		citations := []string{"citation1", "citation2"}
		response := &perplexity.CompletionResponse{
			Choices: []perplexity.Choice{
				{Message: perplexity.Message{Content: "response"}},
			},
			Model:     "sonar",
			Citations: &citations,
		}

		result, err := formatter.Format(response)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.IsError {
			t.Error("Expected success result, got error")
		}
	})

	t.Run("prefers search_results over citations", func(t *testing.T) {
		searchResults := []perplexity.SearchResult{{}}
		citations := []string{"citation1"}
		response := &perplexity.CompletionResponse{
			Choices: []perplexity.Choice{
				{Message: perplexity.Message{Content: "response"}},
			},
			Model:         "sonar",
			SearchResults: &searchResults,
			Citations:     &citations,
		}

		result, err := formatter.Format(response)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if result.IsError {
			t.Error("Expected success result, got error")
		}
	})

	t.Run("handles nil response", func(t *testing.T) {
		result, err := formatter.Format(nil)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error result for nil response")
		}
	})

	t.Run("handles empty choices", func(t *testing.T) {
		response := &perplexity.CompletionResponse{
			Choices: []perplexity.Choice{},
			Model:   "sonar",
		}

		result, err := formatter.Format(response)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error result for empty choices")
		}
	})
}

func TestResponseFormatter_BuildResponse(t *testing.T) {
	formatter := NewResponseFormatter()

	t.Run("builds response with all fields", func(t *testing.T) {
		searchResults := []perplexity.SearchResult{{}}
		images := []perplexity.Image{{}}
		relatedQuestions := []string{"Question 1"}

		response := &perplexity.CompletionResponse{
			Choices: []perplexity.Choice{
				{Message: perplexity.Message{Content: "content"}},
			},
			Model:            "sonar",
			Usage:            perplexity.Usage{TotalTokens: 100},
			SearchResults:    &searchResults,
			Images:           &images,
			RelatedQuestions: &relatedQuestions,
		}

		result := formatter.buildResponse(response)

		if result["content"] != "content" {
			t.Errorf("Expected content %q, got %v", "content", result["content"])
		}
		if result["model"] != "sonar" {
			t.Errorf("Expected model %q, got %v", "sonar", result["model"])
		}
		if result["usage"] == nil {
			t.Error("Expected usage to be present")
		}
		if result["search_results"] == nil {
			t.Error("Expected search_results to be present")
		}
		if result["images"] == nil {
			t.Error("Expected images to be present")
		}
		if result["related_questions"] == nil {
			t.Error("Expected related_questions to be present")
		}
	})

	t.Run("builds minimal response", func(t *testing.T) {
		response := &perplexity.CompletionResponse{
			Choices: []perplexity.Choice{
				{Message: perplexity.Message{Content: "content"}},
			},
			Model: "sonar",
			Usage: perplexity.Usage{TotalTokens: 50},
		}

		result := formatter.buildResponse(response)

		if len(result) != 3 {
			t.Errorf("Expected 3 fields (content, model, usage), got %d", len(result))
		}
	})

	t.Run("includes empty slices as nil", func(t *testing.T) {
		emptySearchResults := []perplexity.SearchResult{}
		response := &perplexity.CompletionResponse{
			Choices: []perplexity.Choice{
				{Message: perplexity.Message{Content: "content"}},
			},
			Model:         "sonar",
			SearchResults: &emptySearchResults,
		}

		result := formatter.buildResponse(response)

		if result["search_results"] != nil {
			t.Error("Expected search_results to be nil for empty slice")
		}
	})
}

func TestFormatError(t *testing.T) {
	t.Run("formats error correctly", func(t *testing.T) {
		err := errors.New("test error")
		result := FormatError(err)

		if !result.IsError {
			t.Error("Expected IsError to be true")
		}
	})
}
