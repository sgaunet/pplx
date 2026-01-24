// Package cmd provides command-line interface commands for the Perplexity API.
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/sgaunet/pplx/pkg/clerrors"
	"github.com/sgaunet/pplx/pkg/completion"
	"github.com/sgaunet/pplx/pkg/config"
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
	// globalOpts contains all global flag values for the application.
	globalOpts = config.NewGlobalOptions()
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
	level, ok := logger.ParseLevel(globalOpts.LogLevel)
	if !ok {
		return fmt.Errorf("%w: %q, must be one of: %v", clerrors.ErrInvalidLogLevel, globalOpts.LogLevel, logger.ValidLevels())
	}

	// Parse log format
	format, ok := logger.ParseFormat(globalOpts.LogFormat)
	if !ok {
		return fmt.Errorf("%w: %q, must be one of: %v", clerrors.ErrInvalidLogFormat, globalOpts.LogFormat, logger.ValidFormats())
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
	cmd.PersistentFlags().StringVarP(&globalOpts.Model, "model", "m", globalOpts.Model,
		"List of models: https://docs.perplexity.ai/guides/model-cards")
	cmd.PersistentFlags().Float64Var(&globalOpts.FrequencyPenalty, "frequency-penalty", globalOpts.FrequencyPenalty, "Frequency penalty")
	cmd.PersistentFlags().IntVarP(&globalOpts.MaxTokens, "max-tokens", "T", globalOpts.MaxTokens, "Max tokens")
	cmd.PersistentFlags().Float64Var(&globalOpts.PresencePenalty, "presence-penalty", globalOpts.PresencePenalty, "Presence penalty")
	cmd.PersistentFlags().Float64VarP(&globalOpts.Temperature, "temperature", "t", globalOpts.Temperature, "Temperature")
	cmd.PersistentFlags().IntVarP(&globalOpts.TopK, "top-k", "k", globalOpts.TopK, "Top K")
	cmd.PersistentFlags().Float64Var(&globalOpts.TopP, "top-p", globalOpts.TopP, "Top P")
	cmd.PersistentFlags().DurationVar(&globalOpts.Timeout, "timeout", globalOpts.Timeout, "HTTP timeout")
}

func addSearchFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringSliceVarP(&globalOpts.SearchDomains, "search-domains", "d", globalOpts.SearchDomains,
		"Filter search results to specific domains")
	cmd.PersistentFlags().StringVarP(&globalOpts.SearchRecency, "search-recency", "r", globalOpts.SearchRecency,
		"Filter by time: day, week, month, year")
	cmd.PersistentFlags().Float64Var(&globalOpts.LocationLat, "location-lat", globalOpts.LocationLat, "User location latitude")
	cmd.PersistentFlags().Float64Var(&globalOpts.LocationLon, "location-lon", globalOpts.LocationLon, "User location longitude")
	cmd.PersistentFlags().StringVar(&globalOpts.LocationCountry, "location-country", globalOpts.LocationCountry, "User location country code")
}

func addResponseFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&globalOpts.ReturnImages, "return-images", "i", globalOpts.ReturnImages, "Include images in response (note: automatically disables --search-recency)")
	cmd.PersistentFlags().BoolVarP(&globalOpts.ReturnRelated, "return-related", "q", globalOpts.ReturnRelated, "Include related questions")
	cmd.PersistentFlags().BoolVarP(&globalOpts.Stream, "stream", "S", globalOpts.Stream, "Enable streaming responses")
}

func addImageFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringSliceVar(&globalOpts.ImageDomains, "image-domains", globalOpts.ImageDomains, "Filter images by domains")
	cmd.PersistentFlags().StringSliceVar(&globalOpts.ImageFormats, "image-formats", globalOpts.ImageFormats,
		"Filter images by formats (jpg, png, etc.)")
}

func addFormatFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&globalOpts.ResponseFormatJSONSchema, "response-format-json-schema",
		globalOpts.ResponseFormatJSONSchema, "JSON schema for structured output (sonar model only)")
	cmd.PersistentFlags().StringVar(&globalOpts.ResponseFormatRegex, "response-format-regex",
		globalOpts.ResponseFormatRegex, "Regex pattern for structured output (sonar model only)")
	cmd.PersistentFlags().StringVarP(&globalOpts.SearchMode, "search-mode", "a", globalOpts.SearchMode, "Search mode: web (default) or academic")
	cmd.PersistentFlags().StringVarP(&globalOpts.SearchContextSize, "search-context-size", "c", globalOpts.SearchContextSize,
		"Search context size: low, medium, or high")
}

func addDateFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&globalOpts.SearchAfterDate, "search-after-date", globalOpts.SearchAfterDate,
		"Filter results published after date (MM/DD/YYYY)")
	cmd.PersistentFlags().StringVar(&globalOpts.SearchBeforeDate, "search-before-date", globalOpts.SearchBeforeDate,
		"Filter results published before date (MM/DD/YYYY)")
	cmd.PersistentFlags().StringVar(&globalOpts.LastUpdatedAfter, "last-updated-after", globalOpts.LastUpdatedAfter,
		"Filter results last updated after date (MM/DD/YYYY)")
	cmd.PersistentFlags().StringVar(&globalOpts.LastUpdatedBefore, "last-updated-before", globalOpts.LastUpdatedBefore,
		"Filter results last updated before date (MM/DD/YYYY)")
}

func addResearchFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&globalOpts.ReasoningEffort, "reasoning-effort", globalOpts.ReasoningEffort,
		"Reasoning effort for sonar-deep-research: low, medium, or high")
}

func addOutputFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVar(&globalOpts.OutputJSON, "json", globalOpts.OutputJSON, "Output response in JSON format")
}

func addLoggingFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&globalOpts.LogLevel, "log-level", globalOpts.LogLevel,
		"Log level (debug, info, warn, error)")
	cmd.PersistentFlags().StringVar(&globalOpts.LogFormat, "log-format", globalOpts.LogFormat,
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
	queryCmd.PersistentFlags().StringVarP(&globalOpts.SystemPrompt, "sys-prompt", "s", "", "system prompt")
	queryCmd.PersistentFlags().StringVarP(&globalOpts.UserPrompt, "user-prompt", "p", "", "user prompt")
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
