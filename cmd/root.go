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

	// Search/Web options.
	searchDomains    []string
	searchRecency    string
	locationLat      float64
	locationLon      float64
	locationCountry  string

	// Response enhancement options.
	returnImages    bool
	returnRelated   bool
	stream          bool

	// Image filtering options.
	imageDomains []string
	imageFormats []string

	// Response format options.
	responseFormatJSONSchema string
	responseFormatRegex      string

	// Search mode options.
	searchMode        string
	searchContextSize string

	// Date filtering options.
	searchAfterDate     string
	searchBeforeDate    string
	lastUpdatedAfter    string
	lastUpdatedBefore   string

	// Deep research options.
	reasoningEffort string

	// Output options.
	outputJSON bool
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "pplx",
	Short: "Program to interact with the Perplexity API",
	Long: `Program to interact with the Perplexity API.
	
	You can use it to chat with the AI or to query it.`,
}

// Execute runs the root command.
func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func addChatFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&model, "model", "m", perplexity.DefaultModel,
		"List of models: https://docs.perplexity.ai/guides/model-cards")
	cmd.PersistentFlags().Float64Var(&frequencyPenalty, "frequency-penalty", frequencyPenalty, "Frequency penalty")
	cmd.PersistentFlags().IntVarP(&maxTokens, "max-tokens", "T", maxTokens, "Max tokens")
	cmd.PersistentFlags().Float64Var(&presencePenalty, "presence-penalty", presencePenalty, "Presence penalty")
	cmd.PersistentFlags().Float64VarP(&temperature, "temperature", "t", temperature, "Temperature")
	cmd.PersistentFlags().IntVarP(&topK, "top-k", "k", topK, "Top K")
	cmd.PersistentFlags().Float64Var(&topP, "top-p", topP, "Top P")
	cmd.PersistentFlags().DurationVar(&timeout, "timeout", timeout, "HTTP timeout")
}

func addSearchFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringSliceVarP(&searchDomains, "search-domains", "d", searchDomains,
		"Filter search results to specific domains")
	cmd.PersistentFlags().StringVarP(&searchRecency, "search-recency", "r", searchRecency,
		"Filter by time: day, week, month, year")
	cmd.PersistentFlags().Float64Var(&locationLat, "location-lat", locationLat, "User location latitude")
	cmd.PersistentFlags().Float64Var(&locationLon, "location-lon", locationLon, "User location longitude")
	cmd.PersistentFlags().StringVar(&locationCountry, "location-country", locationCountry, "User location country code")
}

func addResponseFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&returnImages, "return-images", "i", returnImages, "Include images in response")
	cmd.PersistentFlags().BoolVarP(&returnRelated, "return-related", "q", returnRelated, "Include related questions")
	cmd.PersistentFlags().BoolVarP(&stream, "stream", "S", stream, "Enable streaming responses")
}

func addImageFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringSliceVar(&imageDomains, "image-domains", imageDomains, "Filter images by domains")
	cmd.PersistentFlags().StringSliceVar(&imageFormats, "image-formats", imageFormats,
		"Filter images by formats (jpg, png, etc.)")
}

func addFormatFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&responseFormatJSONSchema, "response-format-json-schema",
		responseFormatJSONSchema, "JSON schema for structured output (sonar model only)")
	cmd.PersistentFlags().StringVar(&responseFormatRegex, "response-format-regex",
		responseFormatRegex, "Regex pattern for structured output (sonar model only)")
	cmd.PersistentFlags().StringVarP(&searchMode, "search-mode", "a", searchMode, "Search mode: web (default) or academic")
	cmd.PersistentFlags().StringVarP(&searchContextSize, "search-context-size", "c", searchContextSize,
		"Search context size: low, medium, or high")
}

func addDateFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&searchAfterDate, "search-after-date", searchAfterDate,
		"Filter results published after date (MM/DD/YYYY)")
	cmd.PersistentFlags().StringVar(&searchBeforeDate, "search-before-date", searchBeforeDate,
		"Filter results published before date (MM/DD/YYYY)")
	cmd.PersistentFlags().StringVar(&lastUpdatedAfter, "last-updated-after", lastUpdatedAfter,
		"Filter results last updated after date (MM/DD/YYYY)")
	cmd.PersistentFlags().StringVar(&lastUpdatedBefore, "last-updated-before", lastUpdatedBefore,
		"Filter results last updated before date (MM/DD/YYYY)")
}

func addResearchFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&reasoningEffort, "reasoning-effort", reasoningEffort,
		"Reasoning effort for sonar-deep-research: low, medium, or high")
}

func addOutputFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVar(&outputJSON, "json", outputJSON, "Output response in JSON format")
}

func init() {
	rootCmd.AddCommand(chatCmd)
	addChatFlags(chatCmd)
	addSearchFlags(chatCmd)
	addResponseFlags(chatCmd)
	addImageFlags(chatCmd)
	addFormatFlags(chatCmd)
	addDateFlags(chatCmd)
	addResearchFlags(chatCmd)

	rootCmd.AddCommand(queryCmd)
	queryCmd.PersistentFlags().StringVarP(&systemPrompt, "sys-prompt", "s", "", "system prompt")
	queryCmd.PersistentFlags().StringVarP(&userPrompt, "user-prompt", "p", "", "user prompt")
	addChatFlags(queryCmd)
	addSearchFlags(queryCmd)
	addResponseFlags(queryCmd)
	addImageFlags(queryCmd)
	addFormatFlags(queryCmd)
	addDateFlags(queryCmd)
	addResearchFlags(queryCmd)
	addOutputFlags(queryCmd)

	rootCmd.AddCommand(mcpStdioCmd)
}
