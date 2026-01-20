// Package cmd provides command-line interface commands for the Perplexity API.
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/sgaunet/perplexity-go/v2"
	"github.com/sgaunet/pplx/pkg/clerrors"
	"github.com/sgaunet/pplx/pkg/completion"
	"github.com/sgaunet/pplx/pkg/logger"
	"github.com/spf13/cobra"
)

const (
	// Exit codes for different error types.
	exitCodeSuccess         = 0
	exitCodeGeneral         = 1
	exitCodeValidation      = 2
	exitCodeAPI             = 3
	exitCodeConfiguration   = 4
	exitCodeIO              = 5
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

	// Logging options.
	logLevel  string
	logFormat string
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

	// Initialize logger before command execution
	if err := initLogger(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize logger: %v\n", err)
	}

	err := rootCmd.Execute()
	if err != nil {
		printError(err)
		exitCode := getExitCode(err)
		os.Exit(exitCode)
	}
}

// initLogger initializes the logger with the configured level and format.
func initLogger() error {
	// Parse log level
	level, ok := logger.ParseLevel(logLevel)
	if !ok {
		return fmt.Errorf("%w: %q, must be one of: %v", clerrors.ErrInvalidLogLevel, logLevel, logger.ValidLevels())
	}

	// Parse log format
	format, ok := logger.ParseFormat(logFormat)
	if !ok {
		return fmt.Errorf("%w: %q, must be one of: %v", clerrors.ErrInvalidLogFormat, logFormat, logger.ValidFormats())
	}

	// Initialize logger
	logger.Init(level, format, os.Stderr)
	return nil
}

// printError prints error messages with appropriate formatting based on error type.
func printError(err error) {
	var validationErr *clerrors.ValidationError
	var apiErr *clerrors.APIError
	var configErr *clerrors.ConfigError
	var ioErr *clerrors.IOError

	//nolint:gocritic // errors.As requires if-else chain, cannot use switch
	if errors.As(err, &validationErr) {
		fmt.Fprintf(os.Stderr, "❌ Validation Error: %v\n", validationErr)
	} else if errors.As(err, &apiErr) {
		fmt.Fprintf(os.Stderr, "❌ API Error: %v\n", apiErr)
	} else if errors.As(err, &configErr) {
		fmt.Fprintf(os.Stderr, "❌ Configuration Error: %v\n", configErr)
	} else if errors.As(err, &ioErr) {
		fmt.Fprintf(os.Stderr, "❌ I/O Error: %v\n", ioErr)
	} else {
		fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
	}
}

// getExitCode maps error types to exit codes.
func getExitCode(err error) int {
	var validationErr *clerrors.ValidationError
	var apiErr *clerrors.APIError
	var configErr *clerrors.ConfigError
	var ioErr *clerrors.IOError

	//nolint:gocritic // errors.As requires if-else chain, cannot use switch
	if errors.As(err, &validationErr) {
		return exitCodeValidation
	} else if errors.As(err, &apiErr) {
		return exitCodeAPI
	} else if errors.As(err, &configErr) {
		return exitCodeConfiguration
	} else if errors.As(err, &ioErr) {
		return exitCodeIO
	}
	return exitCodeGeneral
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
	cmd.PersistentFlags().BoolVarP(&returnImages, "return-images", "i", returnImages, "Include images in response (note: automatically disables --search-recency)")
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

func addLoggingFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&logLevel, "log-level", "info",
		"Log level (debug, info, warn, error)")
	cmd.PersistentFlags().StringVar(&logFormat, "log-format", "text",
		"Log format (text, json)")
}

func registerFlagCompletions(cmd *cobra.Command) {
	// Model completion
	if err := cmd.RegisterFlagCompletionFunc("model",
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return completion.GetModels(), cobra.ShellCompDirectiveNoFileComp
		}); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to register completion for 'model' flag: %v\n", err)
	}

	// Search mode completion
	if err := cmd.RegisterFlagCompletionFunc("search-mode",
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return completion.SearchModes(), cobra.ShellCompDirectiveNoFileComp
		}); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to register completion for 'search-mode' flag: %v\n", err)
	}

	// Search recency completion
	if err := cmd.RegisterFlagCompletionFunc("search-recency",
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return completion.RecencyValues(), cobra.ShellCompDirectiveNoFileComp
		}); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to register completion for 'search-recency' flag: %v\n", err)
	}

	// Search context size completion
	if err := cmd.RegisterFlagCompletionFunc("search-context-size",
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return completion.ContextSizes(), cobra.ShellCompDirectiveNoFileComp
		}); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to register completion for 'search-context-size' flag: %v\n", err)
	}

	// Reasoning effort completion
	if err := cmd.RegisterFlagCompletionFunc("reasoning-effort",
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return completion.ReasoningEfforts(), cobra.ShellCompDirectiveNoFileComp
		}); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to register completion for 'reasoning-effort' flag: %v\n", err)
	}

	// Image formats completion
	if err := cmd.RegisterFlagCompletionFunc("image-formats",
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return completion.ImageFormats(), cobra.ShellCompDirectiveNoFileComp
		}); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to register completion for 'image-formats' flag: %v\n", err)
	}

	// Domain suggestions (for both search-domains and image-domains)
	domainCompletion := func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return completion.CommonDomains(), cobra.ShellCompDirectiveNoFileComp
	}
	if err := cmd.RegisterFlagCompletionFunc("search-domains", domainCompletion); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to register completion for 'search-domains' flag: %v\n", err)
	}
	if err := cmd.RegisterFlagCompletionFunc("image-domains", domainCompletion); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to register completion for 'image-domains' flag: %v\n", err)
	}
}

func registerLoggingFlagCompletions(cmd *cobra.Command) {
	// Log level completion
	if err := cmd.RegisterFlagCompletionFunc("log-level",
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return logger.ValidLevels(), cobra.ShellCompDirectiveNoFileComp
		}); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to register completion for 'log-level' flag: %v\n", err)
	}

	// Log format completion
	if err := cmd.RegisterFlagCompletionFunc("log-format",
		func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
			return logger.ValidFormats(), cobra.ShellCompDirectiveNoFileComp
		}); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to register completion for 'log-format' flag: %v\n", err)
	}
}

func init() {
	// Add logging flags to root command
	addLoggingFlags(rootCmd)
	registerLoggingFlagCompletions(rootCmd)

	rootCmd.AddCommand(chatCmd)
	addChatFlags(chatCmd)
	addSearchFlags(chatCmd)
	addResponseFlags(chatCmd)
	addImageFlags(chatCmd)
	addFormatFlags(chatCmd)
	addDateFlags(chatCmd)
	addResearchFlags(chatCmd)
	registerFlagCompletions(chatCmd)

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
	registerFlagCompletions(queryCmd)

	rootCmd.AddCommand(mcpStdioCmd)
}
