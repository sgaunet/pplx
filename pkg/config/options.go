package config

import (
	"time"

	"github.com/sgaunet/perplexity-go/v2"
)

// GlobalOptions contains all global flag values for the application.
// These fields are bound to Cobra flags and must remain addressable.
type GlobalOptions struct {
	// Model parameters
	Model            string
	FrequencyPenalty float64
	MaxTokens        int
	PresencePenalty  float64
	Temperature      float64
	TopK             int
	TopP             float64
	Timeout          time.Duration

	// Prompts (query command only)
	SystemPrompt string
	UserPrompt   string

	// Search options
	SearchDomains   []string
	SearchRecency   string
	LocationLat     float64
	LocationLon     float64
	LocationCountry string

	// Response enhancement options
	ReturnImages  bool
	ReturnRelated bool
	Stream        bool

	// Image filtering options
	ImageDomains []string
	ImageFormats []string

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

	// Output options
	OutputJSON bool

	// Logging options
	LogLevel  string
	LogFormat string
}

// NewGlobalOptions creates a new GlobalOptions instance with default values.
func NewGlobalOptions() *GlobalOptions {
	return &GlobalOptions{
		Model:            perplexity.DefaultModel,
		FrequencyPenalty: perplexity.DefaultFrequencyPenalty,
		MaxTokens:        perplexity.DefaultMaxTokens,
		PresencePenalty:  perplexity.DefaultPresencePenalty,
		Temperature:      perplexity.DefaultTemperature,
		TopK:             perplexity.DefaultTopK,
		TopP:             perplexity.DefaultTopP,
		Timeout:          perplexity.DefaultTimeout,
		LogLevel:         "info",
		LogFormat:        "text",
	}
}
