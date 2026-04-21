package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/pterm/pterm"
	"github.com/sgaunet/perplexity-go/v2"
	"github.com/sgaunet/pplx/pkg/config"
	"github.com/sgaunet/pplx/pkg/console"
	clerrors "github.com/sgaunet/pplx/pkg/clerrors"
	"github.com/sgaunet/pplx/pkg/logger"
	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "",
	Long:  ``,
	RunE: func(cmd *cobra.Command, _ []string) error {
		// Step 1: Load and merge configuration
		// Graceful degradation: If config load fails, continue with CLI flags only.
		// Rationale: User may not have a config file yet, but CLI should still work.
		// This allows the tool to be used immediately after installation without setup.
		cfg, err := config.LoadAndMergeConfig(cmd, configFilePath, runtimeProfile)
		if err != nil {
			// Non-fatal: continue with CLI flags only
			cfg = config.NewConfigData()
		}

		// Apply merged config to global variables
		// Design note: Uses globals for cobra flag compatibility - flags are bound to globals,
		// and ApplyToGlobals ensures config values only apply when flags aren't set.
		config.ApplyToGlobals(cfg, globalOpts)

		// Step 2: Initialize API client
		// API key checked here (not in config load) because it's required at runtime,
		// but config file is optional. This provides fast feedback if key is missing.
		// Fail fast principle: better to error immediately than during expensive API call.
		if os.Getenv("PPLX_API_KEY") == "" {
			return clerrors.NewConfigError("PPLX_API_KEY environment variable is not set", nil)
		}

		client := perplexity.NewClient(os.Getenv("PPLX_API_KEY"))
		client.SetHTTPTimeout(globalOpts.Timeout)

		// Step 3: Validate inputs
		// Early validation before expensive API call provides fast feedback on errors.
		// Catches malformed options (invalid dates, conflicting flags) before network call.
		if err := validateInputs(); err != nil {
			return err
		}

		// Step 4: Build request with all options
		// Separated into dedicated function for testability and reusability.
		// Allows testing request building logic independently from API calls.
		req, err := buildAllOptions()
		if err != nil {
			return err
		}

		// Step 5: Execute request (streaming or non-streaming)
		// Different code paths because streaming requires goroutine coordination
		// while non-streaming uses synchronous request-response pattern.
		// Streaming: incremental rendering with channels and goroutines
		// Non-streaming: spinner while waiting, then render complete response
		if globalOpts.Stream {
			return handleStreamingResponse(client, req)
		}
		return handleNonStreamingResponse(client, req)
	},
}

// parseDateFilter parses a date string in either YYYY-MM-DD (ISO 8601) or MM/DD/YYYY format.
// ISO 8601 is tried first; MM/DD/YYYY is the fallback.
// Returns the parsed time and an error if neither format matches.
func parseDateFilter(fieldName, dateStr string) (time.Time, error) {
	if date, err := time.Parse("2006-01-02", dateStr); err == nil {
		return date, nil
	}
	if date, err := time.Parse("01/02/2006", dateStr); err == nil {
		return date, nil
	}
	return time.Time{}, clerrors.NewValidationError(
		fieldName, dateStr, "invalid date format, use YYYY-MM-DD or MM/DD/YYYY",
	)
}

// validateStringEnum validates that a value is in the allowed set.
// Returns nil if value is empty or valid, error otherwise.
func validateStringEnum(fieldName, value string, validValues map[string]bool, validList string) error {
	if value == "" {
		return nil
	}
	if !validValues[value] {
		return clerrors.NewValidationError(fieldName, value,
			"must be one of: "+validList)
	}
	return nil
}

// buildBaseOptions creates the base completion request options.
// These options are always included in every request.
func buildBaseOptions() ([]perplexity.CompletionRequestOption, error) {
	msg := perplexity.NewMessages(perplexity.WithSystemMessage(globalOpts.SystemPrompt))
	if err := addUserMessage(&msg); err != nil {
		return nil, err
	}

	return []perplexity.CompletionRequestOption{
		perplexity.WithMessagesFromMessages(&msg),
		perplexity.WithModel(globalOpts.Model),
		perplexity.WithFrequencyPenalty(globalOpts.FrequencyPenalty),
		perplexity.WithMaxTokens(globalOpts.MaxTokens),
		perplexity.WithPresencePenalty(globalOpts.PresencePenalty),
		perplexity.WithTemperature(globalOpts.Temperature),
		perplexity.WithTopK(globalOpts.TopK),
		perplexity.WithTopP(globalOpts.TopP),
	}, nil
}

// addUserMessage appends the user prompt (and any file attachments) to msg.
// Without --file, it falls back to a plain text message.
// With --file, it builds a multimodal message combining the prompt text with
// each attachment routed to image or file content based on extension.
func addUserMessage(msg *perplexity.Messages) error {
	if len(globalOpts.Files) == 0 {
		if err := msg.AddUserMessage(globalOpts.UserPrompt); err != nil {
			return fmt.Errorf("failed to add user message to request: %w", err)
		}
		return nil
	}

	contents := []perplexity.Content{perplexity.NewTextContent(globalOpts.UserPrompt)}
	for _, entry := range globalOpts.Files {
		content, err := buildAttachmentContent(entry)
		if err != nil {
			return err
		}
		contents = append(contents, content)
	}

	if err := msg.AddMultimodalUserMessage(contents); err != nil {
		return fmt.Errorf("failed to add multimodal user message: %w", err)
	}
	return nil
}

// buildAttachmentContent turns a local path or HTTPS URL into a perplexity.Content.
// Library sentinel errors (size, format, missing file) are mapped to pplx typed errors.
func buildAttachmentContent(entry string) (perplexity.Content, error) {
	isImage, isURL, err := classifyAttachment(entry)
	if err != nil {
		return perplexity.Content{}, err
	}

	if isURL {
		if isImage {
			return perplexity.NewImageURLContent(entry), nil
		}
		return perplexity.NewFileURLContent(entry, path.Base(entry)), nil
	}

	if isImage {
		content, libErr := perplexity.NewImageFileContent(entry)
		if libErr != nil {
			return perplexity.Content{}, mapAttachmentError("file", entry, libErr)
		}
		return content, nil
	}
	content, libErr := perplexity.NewFileFileContent(entry)
	if libErr != nil {
		return perplexity.Content{}, mapAttachmentError("file", entry, libErr)
	}
	return content, nil
}

// classifyAttachment inspects a --file entry and reports whether it is an image
// or a document, and whether it is an https URL or a local path.
// Returns a validation error when the entry is empty, the URL is not https, or
// the extension is not supported by the Perplexity API.
func classifyAttachment(entry string) (bool, bool, error) {
	if entry == "" {
		return false, false, clerrors.NewValidationError("file", entry, "file path or URL is required")
	}

	var (
		ext   string
		isURL bool
	)
	if strings.Contains(entry, "://") {
		parsed, perr := url.Parse(entry)
		if perr != nil || parsed.Host == "" {
			return false, false, clerrors.NewValidationError("file", entry, "invalid URL")
		}
		if parsed.Scheme != "https" {
			return false, false, clerrors.NewValidationError("file", entry, "URLs must use https://")
		}
		isURL = true
		ext = strings.TrimPrefix(strings.ToLower(path.Ext(parsed.Path)), ".")
	} else {
		ext = strings.TrimPrefix(strings.ToLower(filepath.Ext(entry)), ".")
	}

	if ext == "" {
		return false, isURL, clerrors.NewValidationError("file", entry,
			"missing file extension; supported: "+supportedExtensionsList())
	}
	if slices.Contains(perplexity.SupportedImageFormats, ext) {
		return true, isURL, nil
	}
	if slices.Contains(perplexity.SupportedFileFormats, ext) {
		return false, isURL, nil
	}
	return false, isURL, clerrors.NewValidationError("file", entry,
		"unsupported extension ."+ext+"; supported: "+supportedExtensionsList())
}

// supportedExtensionsList returns a human-readable comma-joined list of
// every extension the Perplexity API accepts as an attachment.
func supportedExtensionsList() string {
	all := append([]string{}, perplexity.SupportedImageFormats...)
	all = append(all, perplexity.SupportedFileFormats...)
	return strings.Join(all, ", ")
}

// mapAttachmentError converts perplexity-go sentinel errors into pplx typed errors
// so attachment issues map to the correct exit code (validation vs IO).
func mapAttachmentError(field, value string, err error) error {
	switch {
	case errors.Is(err, perplexity.ErrImageFileNotFound), errors.Is(err, perplexity.ErrFileNotFound):
		return clerrors.NewValidationError(field, value, "file not found")
	case errors.Is(err, perplexity.ErrImageTooLarge), errors.Is(err, perplexity.ErrFileTooLarge):
		return clerrors.NewValidationError(field, value, "file exceeds 50MB limit")
	case errors.Is(err, perplexity.ErrImageFormatNotSupported),
		errors.Is(err, perplexity.ErrFileFormatNotSupported):
		return clerrors.NewValidationError(field, value,
			"unsupported format; supported: "+supportedExtensionsList())
	case errors.Is(err, perplexity.ErrImageReadFailed), errors.Is(err, perplexity.ErrFileReadFailed):
		return clerrors.NewIOError("failed to read "+value, err)
	}
	return clerrors.NewIOError("failed to process "+value, err)
}

// buildSearchOptions creates search-related options for the completion request.
// Returns options based on search flags (domains, recency, location).
func buildSearchOptions() []perplexity.CompletionRequestOption {
	var opts []perplexity.CompletionRequestOption

	if len(globalOpts.SearchDomains) > 0 {
		opts = append(opts, perplexity.WithSearchDomainFilter(globalOpts.SearchDomains))
	}

	// Handle search recency - incompatible with images
	if globalOpts.SearchRecency != "" {
		if globalOpts.ReturnImages {
			// User-facing notification (not a log message)
			fmt.Printf("Note: When using --return-images, search-recency is automatically disabled\nProceeding with image search...\n")
		} else {
			opts = append(opts, perplexity.WithSearchRecencyFilter(globalOpts.SearchRecency))
		}
	}

	if globalOpts.LocationLat != 0 || globalOpts.LocationLon != 0 || globalOpts.LocationCountry != "" {
		opts = append(opts, perplexity.WithUserLocation(globalOpts.LocationLat, globalOpts.LocationLon, globalOpts.LocationCountry))
	}

	if globalOpts.SearchMode != "" {
		opts = append(opts, perplexity.WithSearchMode(globalOpts.SearchMode))
	}

	if globalOpts.SearchContextSize != "" {
		opts = append(opts, perplexity.WithSearchContextSize(globalOpts.SearchContextSize))
	}

	return opts
}

// buildResponseOptions creates response enhancement options.
// Handles streaming, images, and related questions.
func buildResponseOptions() []perplexity.CompletionRequestOption {
	var opts []perplexity.CompletionRequestOption

	if globalOpts.ReturnImages {
		opts = append(opts, perplexity.WithReturnImages(globalOpts.ReturnImages))
		// When images are requested, explicitly disable search recency
		opts = append(opts, perplexity.WithSearchRecencyFilter(""))
	}

	if globalOpts.ReturnRelated {
		opts = append(opts, perplexity.WithReturnRelatedQuestions(globalOpts.ReturnRelated))
	}

	if globalOpts.Stream {
		opts = append(opts, perplexity.WithStream(globalOpts.Stream))
	}

	return opts
}

// buildImageOptions creates image filtering options.
// Validates image formats and warns about unsupported formats.
func buildImageOptions() []perplexity.CompletionRequestOption {
	var opts []perplexity.CompletionRequestOption

	if len(globalOpts.ImageDomains) > 0 {
		opts = append(opts, perplexity.WithImageDomainFilter(globalOpts.ImageDomains))
	}

	if len(globalOpts.ImageFormats) > 0 {
		// Validate image formats
		for _, format := range globalOpts.ImageFormats {
			if !config.ValidImageFormats[format] {
				logger.Warn("image format may not be supported",
					"format", format,
					"supported", "jpg, jpeg, png, gif, webp, svg, bmp")
			}
		}
		opts = append(opts, perplexity.WithImageFormatFilter(globalOpts.ImageFormats))
	}

	return opts
}

// buildResponseFormatOptions creates structured output format options.
// Handles JSON schema and regex response formats (sonar models only).
func buildResponseFormatOptions() ([]perplexity.CompletionRequestOption, error) {
	var opts []perplexity.CompletionRequestOption

	if globalOpts.ResponseFormatJSONSchema != "" {
		var schema any
		err := json.Unmarshal([]byte(globalOpts.ResponseFormatJSONSchema), &schema)
		if err != nil {
			return nil, clerrors.NewValidationError("response-format-json-schema",
				globalOpts.ResponseFormatJSONSchema, fmt.Sprintf("invalid JSON schema: %v", err))
		}
		opts = append(opts, perplexity.WithJSONSchemaResponseFormat(schema))
	}

	if globalOpts.ResponseFormatRegex != "" {
		opts = append(opts, perplexity.WithRegexResponseFormat(globalOpts.ResponseFormatRegex))
	}

	return opts, nil
}

// buildDateFilterOptions creates date-based filtering options.
// Parses and validates all date filter flags.
func buildDateFilterOptions() ([]perplexity.CompletionRequestOption, error) {
	var opts []perplexity.CompletionRequestOption

	if globalOpts.SearchAfterDate != "" {
		date, err := parseDateFilter("search-after-date", globalOpts.SearchAfterDate)
		if err != nil {
			return nil, err
		}
		opts = append(opts, perplexity.WithPublishedAfter(date))
	}

	if globalOpts.SearchBeforeDate != "" {
		date, err := parseDateFilter("search-before-date", globalOpts.SearchBeforeDate)
		if err != nil {
			return nil, err
		}
		opts = append(opts, perplexity.WithPublishedBefore(date))
	}

	if globalOpts.LastUpdatedAfter != "" {
		date, err := parseDateFilter("last-updated-after", globalOpts.LastUpdatedAfter)
		if err != nil {
			return nil, err
		}
		opts = append(opts, perplexity.WithLastUpdatedAfterFilter(date))
	}

	if globalOpts.LastUpdatedBefore != "" {
		date, err := parseDateFilter("last-updated-before", globalOpts.LastUpdatedBefore)
		if err != nil {
			return nil, err
		}
		opts = append(opts, perplexity.WithLastUpdatedBeforeFilter(date))
	}

	return opts, nil
}

// buildDeepResearchOptions creates deep research options.
// Handles reasoning effort for sonar-deep-research models.
func buildDeepResearchOptions() []perplexity.CompletionRequestOption {
	var opts []perplexity.CompletionRequestOption

	if globalOpts.ReasoningEffort != "" {
		// Warn if model doesn't support reasoning effort
		if !strings.Contains(globalOpts.Model, "deep-research") {
			logger.Warn("reasoning-effort only supported by sonar-deep-research model",
				"current_model", globalOpts.Model)
		}
		opts = append(opts, perplexity.WithReasoningEffort(globalOpts.ReasoningEffort))
	}

	return opts
}

// validateInputs validates user inputs before building the request.
// This centralizes all pre-request validation logic.
func validateInputs() error {
	if globalOpts.UserPrompt == "" {
		return clerrors.NewValidationError("user-prompt", "", "user prompt is required")
	}

	if err := validateEnumFields(); err != nil {
		return err
	}

	if err := validateFiles(); err != nil {
		return err
	}

	return validateResponseFormats()
}

// validateFiles checks every --file entry up front: that URLs are https, local
// paths exist, and extensions are supported. The library re-validates size and
// format during encoding, but fast-failing here keeps errors consistent with
// the rest of the pre-request validation flow.
func validateFiles() error {
	for _, entry := range globalOpts.Files {
		_, isURL, err := classifyAttachment(entry)
		if err != nil {
			return err
		}
		if !isURL {
			if _, err := os.Stat(entry); err != nil {
				if os.IsNotExist(err) {
					return clerrors.NewValidationError("file", entry, "file not found")
				}
				return clerrors.NewIOError("cannot access "+entry, err)
			}
		}
	}
	return nil
}

// validateEnumFields validates all enum-based configuration options.
func validateEnumFields() error {
	// Validate search recency
	if err := validateStringEnum("search-recency", globalOpts.SearchRecency,
		config.ValidSearchRecency, "day, week, month, year, hour"); err != nil {
		return err
	}

	// Validate search mode
	if err := validateStringEnum("search-mode", globalOpts.SearchMode,
		config.ValidSearchModes, "web, academic"); err != nil {
		return err
	}

	// Validate search context size
	if err := validateStringEnum("search-context-size", globalOpts.SearchContextSize,
		config.ValidContextSizes, "low, medium, high"); err != nil {
		return err
	}

	// Validate reasoning effort
	return validateStringEnum("reasoning-effort", globalOpts.ReasoningEffort,
		config.ValidReasoningEfforts, "low, medium, high")
}

// validateResponseFormats validates response format options and model compatibility.
func validateResponseFormats() error {
	// Validate response format mutual exclusivity
	if globalOpts.ResponseFormatJSONSchema != "" && globalOpts.ResponseFormatRegex != "" {
		return clerrors.NewValidationError("response-format", "",
			"cannot use both --response-format-json-schema and --response-format-regex")
	}

	// Validate model for response formats
	hasResponseFormat := globalOpts.ResponseFormatJSONSchema != "" || globalOpts.ResponseFormatRegex != ""
	if hasResponseFormat && !strings.HasPrefix(globalOpts.Model, "sonar") {
		return clerrors.NewValidationError("model", globalOpts.Model,
			"response formats (JSON schema and regex) are only supported by sonar models")
	}

	return nil
}

// buildAllOptions builds all completion request options using the builder aggregation pattern.
// Each builder contributes a subset of options for its concern (search, format, dates, etc.),
// making individual builders testable and allowing conditional inclusion based on flags.
//
// Design rationale: Separating builders by concern (rather than one monolithic function)
// improves maintainability and testability. Each builder can be tested independently,
// and new option categories can be added without modifying existing builders.
func buildAllOptions() (*perplexity.CompletionRequest, error) {
	// Build base options first - always required, no conditionals
	// These establish the request foundation: model, messages, temperature, etc.
	// Must come first as other builders may depend on model selection
	baseOpts, err := buildBaseOptions()
	if err != nil {
		return nil, err
	}

	// Append-only builders: these never fail, safe to chain unconditionally
	// Each builder checks its relevant flags and adds options if set
	// Order doesn't matter for these - they're independent concerns
	opts := baseOpts
	opts = append(opts, buildSearchOptions()...)      // Domain filters, recency, location
	opts = append(opts, buildResponseOptions()...)    // Return images, related questions
	opts = append(opts, buildImageOptions()...)       // Image domain/format filters

	// Error-returning builders: handle validation during option creation
	// These validate input format/syntax before API call to provide fast feedback

	// Format options: validates JSON schema syntax and checks mutual exclusivity
	// (can't have both JSON schema and regex at same time)
	formatOpts, err := buildResponseFormatOptions()
	if err != nil {
		return nil, err
	}
	opts = append(opts, formatOpts...)

	// Date options: validates MM/DD/YYYY format before sending to API
	// Catches malformed dates early rather than getting API error
	dateOpts, err := buildDateFilterOptions()
	if err != nil {
		return nil, err
	}
	opts = append(opts, dateOpts...)

	// Deep research options: appended last for clarity (order doesn't matter functionally)
	// Keeps research-specific options isolated from standard query options
	opts = append(opts, buildDeepResearchOptions()...)

	// Create request and run final validation
	// This checks the complete option set for conflicts that span multiple builders
	// Example: response format requires sonar model (checked across base + format builders)
	req := perplexity.NewCompletionRequest(opts...)
	if err := req.Validate(); err != nil {
		return nil, clerrors.NewValidationError("request", "", err.Error())
	}

	return req, nil
}

// handleStreamingResponse processes a streaming completion request.
// Uses goroutine-based concurrency to handle incremental token rendering while
// maintaining the final response for complete metadata (citations, images, related questions).
// The goroutine consumes from responseChannel while the main thread produces via SendSSEHTTPRequest.
func handleStreamingResponse(client *perplexity.Client, req *perplexity.CompletionRequest) error {
	// Unbuffered channel ensures backpressure: server won't send next chunk until we process current one
	// This prevents memory buildup if rendering is slower than server response generation
	responseChannel := make(chan perplexity.CompletionResponse)
	var wg sync.WaitGroup

	// Start consumer goroutine - must be running before SendSSEHTTPRequest starts producing
	// Goroutine lifecycle: spawn -> consume from channel -> terminate when channel closes
	wg.Go(func() {
		var lastResponse *perplexity.CompletionResponse

		if globalOpts.OutputJSON {
			// JSON mode: skip incremental rendering, just collect final response
			// Rationale: JSON clients expect complete, valid JSON - not streaming fragments
			// The lastResponse will contain the full completion with all metadata
			for response := range responseChannel {
				lastResponse = &response
			}
		} else {
			// Console mode: render tokens incrementally for better user experience
			// Shows progress as the model generates output (similar to ChatGPT interface)
			renderer := console.NewStreamingRenderer(os.Stdout)
			for response := range responseChannel {
				err := renderer.RenderIncremental(&response)
				if err != nil {
					logger.Error("failed to render streaming content", "error", err)
				}
				// Preserve last response for complete metadata (citations, images, related questions)
				// Only the final chunk contains these - intermediate chunks have partial text only
				lastResponse = &response
			}
		}

		// After channel closes (streaming complete), render final metadata
		// This includes citations, images, and related questions that only arrive in the final chunk
		if lastResponse != nil {
			if !globalOpts.OutputJSON {
				// Visual separation between streaming content and metadata sections
				fmt.Println()
			}
			err := console.RenderResponse(lastResponse, os.Stdout, globalOpts.OutputJSON)
			if err != nil {
				logger.Error("failed to render response", "error", err)
			}
		}
	})

	// Producer: sends SSE request, server writes to responseChannel as tokens arrive
	// Must happen after goroutine spawn to ensure consumer is ready to receive
	// If called before goroutine starts, channel would block and deadlock
	err := client.SendSSEHTTPRequest(&wg, req, responseChannel)
	if err != nil {
		return clerrors.NewAPIError("failed to send streaming request", err)
	}

	// Wait for consumer goroutine to finish processing all responses
	// Critical: prevents goroutine leak and ensures all output is written before function returns
	// Without this, main function might exit while goroutine is still rendering
	wg.Wait()
	return nil
}

// handleNonStreamingResponse processes a standard (non-streaming) completion request.
// Shows a spinner while waiting for the response (unless JSON output is requested).
func handleNonStreamingResponse(client *perplexity.Client, req *perplexity.CompletionRequest) error {
	var spinnerInfo *pterm.SpinnerPrinter
	if !globalOpts.OutputJSON {
		spinnerInfo, _ = pterm.DefaultSpinner.Start("Waiting for response from perplexity...")
	}

	res, err := client.SendCompletionRequest(req)
	if err != nil {
		return clerrors.NewAPIError("failed to send completion request", err)
	}

	if !globalOpts.OutputJSON {
		spinnerInfo.Success("Response received")
	}

	err = console.RenderResponse(res, os.Stdout, globalOpts.OutputJSON)
	if err != nil {
		return clerrors.NewIOError("failed to render response", err)
	}

	return nil
}
