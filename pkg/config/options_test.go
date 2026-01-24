package config

import (
	"testing"

	"github.com/sgaunet/perplexity-go/v2"
)

func TestNewGlobalOptions(t *testing.T) {
	opts := NewGlobalOptions()

	// Test model parameters have correct defaults
	if opts.Model != perplexity.DefaultModel {
		t.Errorf("Expected default model %s, got %s", perplexity.DefaultModel, opts.Model)
	}
	if opts.FrequencyPenalty != perplexity.DefaultFrequencyPenalty {
		t.Errorf("Expected default frequency penalty %f, got %f", perplexity.DefaultFrequencyPenalty, opts.FrequencyPenalty)
	}
	if opts.MaxTokens != perplexity.DefaultMaxTokens {
		t.Errorf("Expected default max tokens %d, got %d", perplexity.DefaultMaxTokens, opts.MaxTokens)
	}
	if opts.PresencePenalty != perplexity.DefaultPresencePenalty {
		t.Errorf("Expected default presence penalty %f, got %f", perplexity.DefaultPresencePenalty, opts.PresencePenalty)
	}
	if opts.Temperature != perplexity.DefaultTemperature {
		t.Errorf("Expected default temperature %f, got %f", perplexity.DefaultTemperature, opts.Temperature)
	}
	if opts.TopK != perplexity.DefaultTopK {
		t.Errorf("Expected default topK %d, got %d", perplexity.DefaultTopK, opts.TopK)
	}
	if opts.TopP != perplexity.DefaultTopP {
		t.Errorf("Expected default topP %f, got %f", perplexity.DefaultTopP, opts.TopP)
	}
	if opts.Timeout != perplexity.DefaultTimeout {
		t.Errorf("Expected default timeout %v, got %v", perplexity.DefaultTimeout, opts.Timeout)
	}

	// Test logging parameters have defaults
	if opts.LogLevel != "info" {
		t.Errorf("Expected default log level 'info', got %s", opts.LogLevel)
	}
	if opts.LogFormat != "text" {
		t.Errorf("Expected default log format 'text', got %s", opts.LogFormat)
	}

	// Test prompt parameters start empty
	if opts.SystemPrompt != "" {
		t.Errorf("Expected empty system prompt, got %s", opts.SystemPrompt)
	}
	if opts.UserPrompt != "" {
		t.Errorf("Expected empty user prompt, got %s", opts.UserPrompt)
	}

	// Test search options start with zero values
	if len(opts.SearchDomains) != 0 {
		t.Errorf("Expected empty search domains, got %v", opts.SearchDomains)
	}
	if opts.SearchRecency != "" {
		t.Errorf("Expected empty search recency, got %s", opts.SearchRecency)
	}

	// Test boolean options start false
	if opts.ReturnImages {
		t.Error("Expected ReturnImages to be false")
	}
	if opts.ReturnRelated {
		t.Error("Expected ReturnRelated to be false")
	}
	if opts.Stream {
		t.Error("Expected Stream to be false")
	}
	if opts.OutputJSON {
		t.Error("Expected OutputJSON to be false")
	}
}

func TestGlobalOptionsAddressability(t *testing.T) {
	opts := NewGlobalOptions()

	// Verify we can take addresses of all fields (required for Cobra binding)
	// This test will fail at compile time if fields are not addressable

	// Model parameters
	_ = &opts.Model
	_ = &opts.FrequencyPenalty
	_ = &opts.MaxTokens
	_ = &opts.PresencePenalty
	_ = &opts.Temperature
	_ = &opts.TopK
	_ = &opts.TopP
	_ = &opts.Timeout

	// Prompts
	_ = &opts.SystemPrompt
	_ = &opts.UserPrompt

	// Search options
	_ = &opts.SearchDomains
	_ = &opts.SearchRecency
	_ = &opts.LocationLat
	_ = &opts.LocationLon
	_ = &opts.LocationCountry

	// Response options
	_ = &opts.ReturnImages
	_ = &opts.ReturnRelated
	_ = &opts.Stream

	// Image filtering
	_ = &opts.ImageDomains
	_ = &opts.ImageFormats

	// Format options
	_ = &opts.ResponseFormatJSONSchema
	_ = &opts.ResponseFormatRegex
	_ = &opts.SearchMode
	_ = &opts.SearchContextSize

	// Date filtering
	_ = &opts.SearchAfterDate
	_ = &opts.SearchBeforeDate
	_ = &opts.LastUpdatedAfter
	_ = &opts.LastUpdatedBefore

	// Research options
	_ = &opts.ReasoningEffort

	// Output options
	_ = &opts.OutputJSON

	// Logging options
	_ = &opts.LogLevel
	_ = &opts.LogFormat
}
