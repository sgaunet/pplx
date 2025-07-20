package cmd

import (
	"os"

	"github.com/sgaunet/perplexity-go/v2"
	"github.com/spf13/cobra"
)

var (
	systemPrompt     string
	userPrompt       string
	model            string  = perplexity.DefaultModel
	frequencyPenalty float64 = perplexity.DefaultFrequencyPenalty
	maxTokens        int     = perplexity.DefaultMaxTokens
	presencePenalty  float64 = perplexity.DefaultPresencePenalty
	temperature      float64 = perplexity.DefaultTemperature
	topK             int     = perplexity.DefaultTopK
	topP             float64 = perplexity.DefaultTopP
	timeout                  = perplexity.DefaultTimeout

	// Search/Web options
	searchDomains    []string
	searchRecency    string
	locationLat      float64
	locationLon      float64
	locationCountry  string

	// Response enhancement options
	returnImages    bool
	returnRelated   bool
	stream          bool

	// Image filtering options
	imageDomains []string
	imageFormats []string

	// Response format options
	responseFormatJSONSchema string
	responseFormatRegex      string

	// Search mode options
	searchMode        string
	searchContextSize string

	// Date filtering options
	searchAfterDate     string
	searchBeforeDate    string
	lastUpdatedAfter    string
	lastUpdatedBefore   string

	// Deep research options
	reasoningEffort string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pplx",
	Short: "Program to interact with the Perplexity API",
	Long: `Program to interact with the Perplexity API.
	
	You can use it to chat with the AI or to query it.`,
}

func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(chatCmd)
	chatCmd.PersistentFlags().StringVarP(&model, "model", "m", perplexity.DefaultModel, "List of models: https://docs.perplexity.ai/guides/model-cards")

	chatCmd.PersistentFlags().Float64Var(&frequencyPenalty, "frequency-penalty", frequencyPenalty, "Frequency penalty")
	chatCmd.PersistentFlags().IntVar(&maxTokens, "max-tokens", maxTokens, "Max tokens")
	chatCmd.PersistentFlags().Float64Var(&presencePenalty, "presence-penalty", presencePenalty, "Presence penalty")
	chatCmd.PersistentFlags().Float64Var(&temperature, "temperature", temperature, "Temperature")
	chatCmd.PersistentFlags().IntVar(&topK, "top-k", topK, "Top K")
	chatCmd.PersistentFlags().Float64Var(&topP, "top-p", topP, "Top P")
	chatCmd.PersistentFlags().DurationVar(&timeout, "timeout", timeout, "HTTP timeout")

	// Search/Web options
	chatCmd.PersistentFlags().StringSliceVar(&searchDomains, "search-domains", searchDomains, "Filter search results to specific domains")
	chatCmd.PersistentFlags().StringVar(&searchRecency, "search-recency", searchRecency, "Filter by time: day, week, month, year")
	chatCmd.PersistentFlags().Float64Var(&locationLat, "location-lat", locationLat, "User location latitude")
	chatCmd.PersistentFlags().Float64Var(&locationLon, "location-lon", locationLon, "User location longitude")
	chatCmd.PersistentFlags().StringVar(&locationCountry, "location-country", locationCountry, "User location country code")

	// Response enhancement options
	chatCmd.PersistentFlags().BoolVar(&returnImages, "return-images", returnImages, "Include images in response")
	chatCmd.PersistentFlags().BoolVar(&returnRelated, "return-related", returnRelated, "Include related questions")
	chatCmd.PersistentFlags().BoolVar(&stream, "stream", stream, "Enable streaming responses")

	// Image filtering options
	chatCmd.PersistentFlags().StringSliceVar(&imageDomains, "image-domains", imageDomains, "Filter images by domains")
	chatCmd.PersistentFlags().StringSliceVar(&imageFormats, "image-formats", imageFormats, "Filter images by formats (jpg, png, etc.)")

	// Response format options
	chatCmd.PersistentFlags().StringVar(&responseFormatJSONSchema, "response-format-json-schema", responseFormatJSONSchema, "JSON schema for structured output (sonar model only)")
	chatCmd.PersistentFlags().StringVar(&responseFormatRegex, "response-format-regex", responseFormatRegex, "Regex pattern for structured output (sonar model only)")

	// Search mode options
	chatCmd.PersistentFlags().StringVar(&searchMode, "search-mode", searchMode, "Search mode: web (default) or academic")
	chatCmd.PersistentFlags().StringVar(&searchContextSize, "search-context-size", searchContextSize, "Search context size: low, medium, or high")

	// Date filtering options
	chatCmd.PersistentFlags().StringVar(&searchAfterDate, "search-after-date", searchAfterDate, "Filter results published after date (MM/DD/YYYY)")
	chatCmd.PersistentFlags().StringVar(&searchBeforeDate, "search-before-date", searchBeforeDate, "Filter results published before date (MM/DD/YYYY)")
	chatCmd.PersistentFlags().StringVar(&lastUpdatedAfter, "last-updated-after", lastUpdatedAfter, "Filter results last updated after date (MM/DD/YYYY)")
	chatCmd.PersistentFlags().StringVar(&lastUpdatedBefore, "last-updated-before", lastUpdatedBefore, "Filter results last updated before date (MM/DD/YYYY)")

	// Deep research options
	chatCmd.PersistentFlags().StringVar(&reasoningEffort, "reasoning-effort", reasoningEffort, "Reasoning effort for sonar-deep-research: low, medium, or high")

	rootCmd.AddCommand(queryCmd)
	queryCmd.PersistentFlags().StringVarP(&model, "model", "m", perplexity.DefaultModel, "List of models: https://docs.perplexity.ai/guides/model-cards")
	queryCmd.PersistentFlags().StringVarP(&systemPrompt, "sys-prompt", "s", "", "system prompt")
	queryCmd.PersistentFlags().StringVarP(&userPrompt, "user-prompt", "p", "", "user prompt")

	queryCmd.PersistentFlags().Float64Var(&frequencyPenalty, "frequency-penalty", frequencyPenalty, "Frequency penalty")
	queryCmd.PersistentFlags().IntVar(&maxTokens, "max-tokens", maxTokens, "Max tokens")
	queryCmd.PersistentFlags().Float64Var(&presencePenalty, "presence-penalty", presencePenalty, "Presence penalty")
	queryCmd.PersistentFlags().Float64Var(&temperature, "temperature", temperature, "Temperature")
	queryCmd.PersistentFlags().IntVar(&topK, "top-k", topK, "Top K")
	queryCmd.PersistentFlags().Float64Var(&topP, "top-p", topP, "Top P")
	queryCmd.PersistentFlags().DurationVar(&timeout, "timeout", timeout, "HTTP timeout")

	// Search/Web options
	queryCmd.PersistentFlags().StringSliceVar(&searchDomains, "search-domains", searchDomains, "Filter search results to specific domains")
	queryCmd.PersistentFlags().StringVar(&searchRecency, "search-recency", searchRecency, "Filter by time: day, week, month, year")
	queryCmd.PersistentFlags().Float64Var(&locationLat, "location-lat", locationLat, "User location latitude")
	queryCmd.PersistentFlags().Float64Var(&locationLon, "location-lon", locationLon, "User location longitude")
	queryCmd.PersistentFlags().StringVar(&locationCountry, "location-country", locationCountry, "User location country code")

	// Response enhancement options
	queryCmd.PersistentFlags().BoolVar(&returnImages, "return-images", returnImages, "Include images in response")
	queryCmd.PersistentFlags().BoolVar(&returnRelated, "return-related", returnRelated, "Include related questions")
	queryCmd.PersistentFlags().BoolVar(&stream, "stream", stream, "Enable streaming responses")

	// Image filtering options
	queryCmd.PersistentFlags().StringSliceVar(&imageDomains, "image-domains", imageDomains, "Filter images by domains")
	queryCmd.PersistentFlags().StringSliceVar(&imageFormats, "image-formats", imageFormats, "Filter images by formats (jpg, png, etc.)")

	// Response format options
	queryCmd.PersistentFlags().StringVar(&responseFormatJSONSchema, "response-format-json-schema", responseFormatJSONSchema, "JSON schema for structured output (sonar model only)")
	queryCmd.PersistentFlags().StringVar(&responseFormatRegex, "response-format-regex", responseFormatRegex, "Regex pattern for structured output (sonar model only)")

	// Search mode options
	queryCmd.PersistentFlags().StringVar(&searchMode, "search-mode", searchMode, "Search mode: web (default) or academic")
	queryCmd.PersistentFlags().StringVar(&searchContextSize, "search-context-size", searchContextSize, "Search context size: low, medium, or high")

	// Date filtering options
	queryCmd.PersistentFlags().StringVar(&searchAfterDate, "search-after-date", searchAfterDate, "Filter results published after date (MM/DD/YYYY)")
	queryCmd.PersistentFlags().StringVar(&searchBeforeDate, "search-before-date", searchBeforeDate, "Filter results published before date (MM/DD/YYYY)")
	queryCmd.PersistentFlags().StringVar(&lastUpdatedAfter, "last-updated-after", lastUpdatedAfter, "Filter results last updated after date (MM/DD/YYYY)")
	queryCmd.PersistentFlags().StringVar(&lastUpdatedBefore, "last-updated-before", lastUpdatedBefore, "Filter results last updated before date (MM/DD/YYYY)")

	// Deep research options
	queryCmd.PersistentFlags().StringVar(&reasoningEffort, "reasoning-effort", reasoningEffort, "Reasoning effort for sonar-deep-research: low, medium, or high")

	rootCmd.AddCommand(mcpStdioCmd)
}
