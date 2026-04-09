package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	huh "charm.land/huh/v2"
	"github.com/sgaunet/pplx/pkg/config"
)

const (
	filterSeparatorParts = 2

	// Default selection for skippable menus.
	choiceSkip = "skip"

	// Validation range constants.
	minTemperature     = 0.0
	maxTemperature     = 2.0
	minTokens          = 1
	maxTokens          = 100000
	minLatitude        = -90.0
	maxLatitude        = 90.0
	minLongitude       = -180.0
	maxLongitude       = 180.0
	minTopK            = 0
	maxTopK            = 2048
	minTopP            = 0.0
	maxTopP            = 1.0
	minPenalty         = 0.0
	maxPenalty         = 2.0
)

// Validation error sentinels.
var (
	errMustBeNumber  = errors.New("must be a number")
	errMustBeInteger = errors.New("must be an integer")
	errInvalidDate   = errors.New("must be YYYY-MM-DD format (e.g. 2024-01-15)")
	errWizardAborted = errors.New("wizard cancelled by user")
)

// WizardState manages the state and flow of the interactive configuration wizard.
type WizardState struct {
	// I/O for form rendering and testing.
	input      io.Reader
	output     io.Writer
	accessible bool

	// Configuration being built.
	config *config.ConfigData

	// Existing config loaded when --update is active (nil otherwise).
	existingConfig *config.ConfigData

	// User selections.
	useCase        string
	selectedModel  string
	enableStream   bool
	searchFilters  []string
	apiKey         string
	customSettings map[string]any
}

// NewWizardState creates a new wizard state with default values.
func NewWizardState() *WizardState {
	return &WizardState{
		input:          os.Stdin,
		output:         os.Stdout,
		accessible:     os.Getenv("ACCESSIBLE") != "",
		config:         config.NewConfigData(),
		customSettings: make(map[string]any),
	}
}

// NewWizardStateWithExisting creates a wizard state seeded with an existing config
// for --update mode. The wizard will default each prompt to the current value.
func NewWizardStateWithExisting(existing *config.ConfigData) *WizardState {
	w := NewWizardState()
	w.existingConfig = existing
	// Seed wizard selections from the existing config so prompts can show them.
	w.selectedModel = existing.Defaults.Model
	w.enableStream = existing.Output.Stream
	return w
}

// Run executes the wizard flow and returns the configured ConfigData.
// Guides users through 7 sequential steps to build a personalized configuration.
//
// Wizard philosophy: Progressive disclosure with smart defaults
// Rather than overwhelming users with all ~30 configuration options at once, the wizard:
// - Asks 5 essential questions (use case, model, streaming, filters, API key)
// - Offers optional advanced customization (step 6)
// - Provides template-based defaults based on use case selection
// - Builds final config by merging user selections with template defaults
//
// Step sequence rationale:
// 1. Use case: Determines template selection (research vs general vs creative)
// 2. Model: Most important technical choice, influences available features
// 3. Streaming: Simple boolean, affects UX significantly
// 4. Search filters: Optional refinements (mode, recency, context size)
// 5. API key: Required credential (user may skip if already set in env)
// 6. Customization: Advanced users can tweak temperature, max_tokens, etc.
// 7. Build config: Merges all selections into final ConfigData structure.
func (w *WizardState) Run() (*config.ConfigData, error) {
	// Step 1: Use case selection.
	if err := w.selectUseCase(); err != nil {
		return nil, fmt.Errorf("use case selection failed: %w", err)
	}

	// Step 2: Model selection.
	if err := w.selectModel(); err != nil {
		return nil, fmt.Errorf("model selection failed: %w", err)
	}

	// Step 3: Streaming toggle.
	if err := w.configureStreaming(); err != nil {
		return nil, fmt.Errorf("streaming configuration failed: %w", err)
	}

	// Step 4: Search filters.
	if err := w.selectSearchFilters(); err != nil {
		return nil, fmt.Errorf("search filter selection failed: %w", err)
	}

	// Step 5: API key.
	if err := w.configureAPIKey(); err != nil {
		return nil, fmt.Errorf("API key configuration failed: %w", err)
	}

	// Step 6: Additional customization.
	if err := w.offerCustomization(); err != nil {
		return nil, fmt.Errorf("customization failed: %w", err)
	}

	// Step 7: Advanced category-based customization.
	if err := w.offerAdvancedCustomization(); err != nil {
		return nil, fmt.Errorf("advanced customization failed: %w", err)
	}

	// Step 8: Generate final configuration.
	w.buildConfiguration()

	// Print summary.
	w.printSummary()

	return w.config, nil
}

// selectUseCase prompts the user to select their primary use case.
func (w *WizardState) selectUseCase() error {
	return w.runForm(huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select Your Primary Use Case").
				Description("Choose the configuration that best matches your needs.").
				Options(
					huh.NewOption("Research  - Academic and scholarly work with authoritative sources", config.TemplateResearch),
					huh.NewOption("Creative  - Content generation, writing, and brainstorming", config.TemplateCreative),
					huh.NewOption("News      - Current events tracking with reputable news sources", config.TemplateNews),
					huh.NewOption("General   - Balanced configuration for everyday queries", "general"),
					huh.NewOption("Custom    - Start from scratch with full customization", "custom"),
				).
				Value(&w.useCase),
		),
	))
}

// selectModel prompts the user to select their preferred model.
func (w *WizardState) selectModel() error {
	if w.selectedModel == "" {
		w.selectedModel = "sonar"
	}

	return w.runForm(huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Choose Your AI Model").
				Options(
					huh.NewOption("sonar               - Balanced performance (recommended)", "sonar"),
					huh.NewOption("sonar-pro           - Enhanced capabilities", "sonar-pro"),
					huh.NewOption("sonar-reasoning     - Advanced reasoning", "sonar-reasoning"),
					huh.NewOption("sonar-deep-research - In-depth research and analysis", "sonar-deep-research"),
				).
				Value(&w.selectedModel),
		),
	))
}

// configureStreaming prompts the user to enable/disable streaming.
func (w *WizardState) configureStreaming() error {
	if w.existingConfig == nil {
		w.enableStream = true // default to true for new configs
	}

	return w.runForm(huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Enable Response Streaming?").
				Description("Streaming displays responses as they're generated in real-time.").
				Affirmative("Yes").
				Negative("No").
				Value(&w.enableStream),
		),
	))
}

// selectSearchFilters prompts the user to configure search preferences.
func (w *WizardState) selectSearchFilters() error {
	configureSearch := w.existingConfig == nil
	prompt := "Configure search preferences?"
	if w.existingConfig != nil {
		prompt = "Change search preferences?"
	}

	if err := w.runForm(huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(prompt).
				Affirmative("Yes").
				Negative("No").
				Value(&configureSearch),
		),
	)); err != nil {
		return err
	}

	if !configureSearch {
		return nil
	}

	w.searchFilters = []string{}

	if err := w.collectSearchCoreFilters(); err != nil {
		return err
	}

	if err := w.collectSearchDomains(); err != nil {
		return err
	}

	if err := w.promptLocation(); err != nil {
		return err
	}

	return w.promptDateRange()
}

// collectSearchCoreFilters collects mode, recency, and context size selections.
func (w *WizardState) collectSearchCoreFilters() error {
	var searchMode, recency, contextSize string
	if err := w.runForm(huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Search Mode").
				Options(
					huh.NewOption("Web (default)", ""),
					huh.NewOption("Academic - Scholarly sources only", "academic"),
				).
				Value(&searchMode),
			huh.NewSelect[string]().
				Title("Recency Filter").
				Options(
					huh.NewOption("No filter (default)", ""),
					huh.NewOption("Hour", "hour"),
					huh.NewOption("Day", "day"),
					huh.NewOption("Week", "week"),
					huh.NewOption("Month", "month"),
					huh.NewOption("Year", "year"),
				).
				Value(&recency),
			huh.NewSelect[string]().
				Title("Context Size").
				Options(
					huh.NewOption("No preference (default)", ""),
					huh.NewOption("Low - Minimal context", "low"),
					huh.NewOption("Medium - Balanced", "medium"),
					huh.NewOption("High - Maximum context", "high"),
				).
				Value(&contextSize),
		).Title("Search Filters"),
	)); err != nil {
		return err
	}

	if searchMode == "academic" {
		w.searchFilters = append(w.searchFilters, "mode:academic")
	}
	if recency != "" {
		w.searchFilters = append(w.searchFilters, "recency:"+recency)
	}
	if contextSize != "" {
		w.searchFilters = append(w.searchFilters, "context:"+contextSize)
	}

	return nil
}

// collectSearchDomains collects domain filter input.
func (w *WizardState) collectSearchDomains() error {
	var domains string
	if err := w.runForm(huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Domain Filter").
				Description("Comma-separated domains (e.g. example.com,news.org). Leave empty to skip.").
				Placeholder("example.com,news.org").
				Value(&domains),
		),
	)); err != nil {
		return err
	}

	if domains != "" {
		w.searchFilters = append(w.searchFilters, "domains:"+domains)
	}

	return nil
}

// promptLocation groups location prompts (country, lat, lon) behind a single gate.
func (w *WizardState) promptLocation() error {
	var setLoc bool
	if err := w.runForm(huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Set location preferences (country/lat/lon)?").
				Affirmative("Yes").
				Negative("No").
				Value(&setLoc),
		),
	)); err != nil {
		return err
	}

	if !setLoc {
		return nil
	}

	var country, latStr, lonStr string
	if err := w.runForm(huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Country Code").
				Description("2-letter ISO code (e.g. US, FR, DE). Leave empty to skip.").
				Placeholder("US").
				Value(&country),
			huh.NewInput().
				Title("Latitude").
				Description("Value between -90 and 90. Leave empty to skip.").
				Placeholder("37.7749").
				Validate(validateOptionalFloat(minLatitude, maxLatitude)).
				Value(&latStr),
			huh.NewInput().
				Title("Longitude").
				Description("Value between -180 and 180. Leave empty to skip.").
				Placeholder("-122.4194").
				Validate(validateOptionalFloat(minLongitude, maxLongitude)).
				Value(&lonStr),
		).Title("Location Preferences"),
	)); err != nil {
		return err
	}

	country = strings.ToUpper(strings.TrimSpace(country))
	if country != "" {
		w.searchFilters = append(w.searchFilters, "location_country:"+country)
	}
	if latStr != "" {
		w.searchFilters = append(w.searchFilters, "location_lat:"+latStr)
	}
	if lonStr != "" {
		w.searchFilters = append(w.searchFilters, "location_lon:"+lonStr)
	}

	return nil
}

// promptDateRange groups after/before date prompts behind a single gate.
func (w *WizardState) promptDateRange() error {
	var setDates bool
	if err := w.runForm(huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Set date range filters?").
				Affirmative("Yes").
				Negative("No").
				Value(&setDates),
		),
	)); err != nil {
		return err
	}

	if !setDates {
		return nil
	}

	var afterDate, beforeDate string
	if err := w.runForm(huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Results After Date").
				Description("YYYY-MM-DD format. Leave empty to skip.").
				Placeholder("2024-01-15").
				Validate(validateOptionalDate).
				Value(&afterDate),
			huh.NewInput().
				Title("Results Before Date").
				Description("YYYY-MM-DD format. Leave empty to skip.").
				Placeholder("2024-12-31").
				Validate(validateOptionalDate).
				Value(&beforeDate),
		).Title("Date Range Filters"),
	)); err != nil {
		return err
	}

	if afterDate != "" {
		w.searchFilters = append(w.searchFilters, "after_date:"+afterDate)
	}
	if beforeDate != "" {
		w.searchFilters = append(w.searchFilters, "before_date:"+beforeDate)
	}

	return nil
}

// configureAPIKey prompts the user to optionally provide an API key.
func (w *WizardState) configureAPIKey() error {
	envKey := os.Getenv("PERPLEXITY_API_KEY")
	if envKey != "" {
		useEnv := true
		if err := w.runForm(huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(fmt.Sprintf("API key found in environment (%d chars). Use it?", len(envKey))).
					Affirmative("Yes").
					Negative("No").
					Value(&useEnv),
			),
		)); err != nil {
			return err
		}

		if useEnv {
			return nil
		}
	}

	var addKey bool
	if err := w.runForm(huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Add API key to configuration?").
				Description("You can also set PERPLEXITY_API_KEY environment variable later.").
				Affirmative("Yes").
				Negative("No").
				Value(&addKey),
		),
	)); err != nil {
		return err
	}

	if !addKey {
		return nil
	}

	apiKeyInput := huh.NewInput().
		Title("Perplexity API Key").
		Value(&w.apiKey)
	// Only mask password in TUI mode; accessible mode requires plain text.
	if !w.accessible {
		apiKeyInput = apiKeyInput.EchoMode(huh.EchoModePassword)
	}

	return w.runForm(huh.NewForm(
		huh.NewGroup(apiKeyInput),
	))
}

// offerCustomization asks if the user wants basic additional customization.
func (w *WizardState) offerCustomization() error {
	var customize bool
	if err := w.runForm(huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Configure advanced options (temperature, max tokens)?").
				Affirmative("Yes").
				Negative("No").
				Value(&customize),
		),
	)); err != nil {
		return err
	}

	if !customize {
		return nil
	}

	var tempStr, tokensStr string
	if err := w.runForm(huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Temperature (0.0-2.0)").
				Description("Controls randomness. Lower = more focused, higher = more creative.").
				Placeholder("0.2").
				Validate(validateOptionalFloat(minTemperature, maxTemperature)).
				Value(&tempStr),
			huh.NewInput().
				Title("Max Tokens (1-100000)").
				Description("Maximum number of tokens in the response.").
				Placeholder("4000").
				Validate(validateOptionalInt(minTokens, maxTokens)).
				Value(&tokensStr),
		).Title("Basic Customization"),
	)); err != nil {
		return err
	}

	if tempStr != "" {
		if v, err := strconv.ParseFloat(tempStr, 64); err == nil {
			w.customSettings["temperature"] = v
		}
	}
	if tokensStr != "" {
		if v, err := strconv.Atoi(tokensStr); err == nil {
			w.customSettings["max_tokens"] = v
		}
	}

	return nil
}

// offerAdvancedCustomization presents a category menu for extended options.
func (w *WizardState) offerAdvancedCustomization() error {
	choice := choiceSkip
	if err := w.runForm(huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Extended Options (Optional)").
				Description("Configure additional advanced options?").
				Options(
					huh.NewOption("Sampling parameters   (top-k, top-p, penalties)", "sampling"),
					huh.NewOption("Output format         (return images, related questions)", "output"),
					huh.NewOption("Response format       (JSON schema, regex constraint)", "response"),
					huh.NewOption("Reasoning effort      (low / medium / high)", "reasoning"),
					huh.NewOption("All of the above", "all"),
					huh.NewOption("Skip", choiceSkip),
				).
				Value(&choice),
		),
	)); err != nil {
		return err
	}

	switch choice {
	case "sampling":
		return w.configureSamplingParams()
	case "output":
		return w.configureOutputFormat()
	case "response":
		return w.configureResponseFormat()
	case "reasoning":
		return w.configureReasoningEffort()
	case "all":
		return w.configureAllAdvanced()
	default:
		return nil
	}
}

// configureSamplingParams collects top-k, top-p, frequency penalty, presence penalty.
func (w *WizardState) configureSamplingParams() error {
	var topKStr, topPStr, freqPenStr, presPenStr string

	if err := w.runForm(huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Top-K (0-2048, 0=disabled)").
				Placeholder("0").
				Validate(validateOptionalInt(minTopK, maxTopK)).
				Value(&topKStr),
			huh.NewInput().
				Title("Top-P (0.0-1.0)").
				Placeholder("0.9").
				Validate(validateOptionalFloat(minTopP, maxTopP)).
				Value(&topPStr),
			huh.NewInput().
				Title("Frequency Penalty (0.0-2.0)").
				Placeholder("0.0").
				Validate(validateOptionalFloat(minPenalty, maxPenalty)).
				Value(&freqPenStr),
			huh.NewInput().
				Title("Presence Penalty (0.0-2.0)").
				Placeholder("0.0").
				Validate(validateOptionalFloat(minPenalty, maxPenalty)).
				Value(&presPenStr),
		).Title("Sampling Parameters"),
	)); err != nil {
		return err
	}

	if topKStr != "" {
		if v, err := strconv.Atoi(topKStr); err == nil {
			w.customSettings["top_k"] = v
		}
	}
	if topPStr != "" {
		if v, err := strconv.ParseFloat(topPStr, 64); err == nil {
			w.customSettings["top_p"] = v
		}
	}
	if freqPenStr != "" {
		if v, err := strconv.ParseFloat(freqPenStr, 64); err == nil {
			w.customSettings["frequency_penalty"] = v
		}
	}
	if presPenStr != "" {
		if v, err := strconv.ParseFloat(presPenStr, 64); err == nil {
			w.customSettings["presence_penalty"] = v
		}
	}

	return nil
}

// configureOutputFormat collects return_images and return_related settings.
func (w *WizardState) configureOutputFormat() error {
	returnImages := false
	returnRelated := false
	if w.existingConfig != nil {
		returnImages = w.existingConfig.Output.ReturnImages
		returnRelated = w.existingConfig.Output.ReturnRelated
	}

	if err := w.runForm(huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Return images in responses?").
				Affirmative("Yes").
				Negative("No").
				Value(&returnImages),
			huh.NewConfirm().
				Title("Return related questions?").
				Affirmative("Yes").
				Negative("No").
				Value(&returnRelated),
		).Title("Output Format"),
	)); err != nil {
		return err
	}

	w.customSettings["return_images"] = returnImages
	w.customSettings["return_related"] = returnRelated

	return nil
}

// configureResponseFormat collects JSON schema and regex constraints.
func (w *WizardState) configureResponseFormat() error {
	var jsonSchema, regex string

	jsonPlaceholder := ""
	regexPlaceholder := ""
	if w.existingConfig != nil {
		jsonPlaceholder = w.existingConfig.Output.ResponseFormatJSONSchema
		regexPlaceholder = w.existingConfig.Output.ResponseFormatRegex
	}

	if err := w.runForm(huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Response Format JSON Schema").
				Description("Leave empty to skip.").
				Placeholder(jsonPlaceholder).
				Value(&jsonSchema),
			huh.NewInput().
				Title("Response Format Regex").
				Description("Leave empty to skip.").
				Placeholder(regexPlaceholder).
				Value(&regex),
		).Title("Response Format Constraints"),
	)); err != nil {
		return err
	}

	if jsonSchema != "" {
		w.customSettings["response_format_json_schema"] = jsonSchema
	}
	if regex != "" {
		w.customSettings["response_format_regex"] = regex
	}

	return nil
}

// configureReasoningEffort collects the reasoning effort level.
func (w *WizardState) configureReasoningEffort() error {
	effort := choiceSkip

	if w.existingConfig != nil && w.existingConfig.Output.ReasoningEffort != "" {
		effort = w.existingConfig.Output.ReasoningEffort
	}

	if err := w.runForm(huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Reasoning Effort Level").
				Options(
					huh.NewOption("Low    - Faster, less thorough", "low"),
					huh.NewOption("Medium - Balanced (default)", "medium"),
					huh.NewOption("High   - Slower, maximum depth", "high"),
					huh.NewOption("Skip", choiceSkip),
				).
				Value(&effort),
		),
	)); err != nil {
		return err
	}

	if effort != choiceSkip {
		w.customSettings["reasoning_effort"] = effort
	}

	return nil
}

// configureAllAdvanced runs all four advanced configuration categories.
func (w *WizardState) configureAllAdvanced() error {
	if err := w.configureSamplingParams(); err != nil {
		return err
	}
	if err := w.configureOutputFormat(); err != nil {
		return err
	}
	if err := w.configureResponseFormat(); err != nil {
		return err
	}
	return w.configureReasoningEffort()
}

// runForm applies common form options and runs the form.
func (w *WizardState) runForm(f *huh.Form) error {
	err := f.
		WithAccessible(w.accessible).
		WithInput(w.input).
		WithOutput(w.output).
		WithTheme(huh.ThemeFunc(huh.ThemeCatppuccin)).
		Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return errWizardAborted
		}
		return fmt.Errorf("form failed: %w", err)
	}
	return nil
}

// --- Validation helpers for huh forms ---

// validateOptionalFloat returns a validation function for optional float input within [lo, hi].
func validateOptionalFloat(lo, hi float64) func(string) error {
	return func(s string) error {
		if s == "" {
			return nil
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return errMustBeNumber
		}
		if v < lo || v > hi {
			return fmt.Errorf("%w: must be between %.1f and %.1f", errMustBeNumber, lo, hi)
		}
		return nil
	}
}

// validateOptionalInt returns a validation function for optional integer input within [lo, hi].
func validateOptionalInt(lo, hi int) func(string) error {
	return func(s string) error {
		if s == "" {
			return nil
		}
		v, err := strconv.Atoi(s)
		if err != nil {
			return errMustBeInteger
		}
		if v < lo || v > hi {
			return fmt.Errorf("%w: must be between %d and %d", errMustBeInteger, lo, hi)
		}
		return nil
	}
}

// validateOptionalDate validates an optional YYYY-MM-DD date string.
func validateOptionalDate(s string) error {
	if s == "" {
		return nil
	}
	if !isValidDateFormat(s) {
		return errInvalidDate
	}
	return nil
}

// --- Config building logic (unchanged) ---

// loadTemplateIfApplicable loads a template config if useCase requires it.
// Falls back to default config on error.
func (w *WizardState) loadTemplateIfApplicable() {
	if w.useCase == "custom" || w.useCase == "general" {
		return
	}

	templateCfg, err := config.LoadTemplate(w.useCase)
	if err != nil {
		w.config = config.NewConfigData()
	} else {
		w.config = templateCfg
	}
}

// applyCustomSettings applies custom settings from w.customSettings to config.
func (w *WizardState) applyCustomSettings() {
	for key, value := range w.customSettings {
		w.applyCustomSetting(key, value)
	}
}

// applyCustomSetting applies a single custom setting by key to the wizard config.
func (w *WizardState) applyCustomSetting(key string, value any) {
	switch key {
	case "temperature", "top_p", "frequency_penalty", "presence_penalty":
		w.applyCustomFloat64Default(key, value)
	case "max_tokens", "top_k":
		w.applyCustomIntDefault(key, value)
	case "return_images", "return_related":
		w.applyCustomBoolOutput(key, value)
	case "response_format_json_schema", "response_format_regex", "reasoning_effort":
		w.applyCustomStringOutput(key, value)
	}
}

// applyCustomFloat64Default applies a float64 default setting.
func (w *WizardState) applyCustomFloat64Default(key string, value any) {
	v, ok := value.(float64)
	if !ok {
		return
	}
	switch key {
	case "temperature":
		w.config.Defaults.Temperature = v
	case "top_p":
		w.config.Defaults.TopP = v
	case "frequency_penalty":
		w.config.Defaults.FrequencyPenalty = v
	case "presence_penalty":
		w.config.Defaults.PresencePenalty = v
	}
}

// applyCustomIntDefault applies an int default setting.
func (w *WizardState) applyCustomIntDefault(key string, value any) {
	v, ok := value.(int)
	if !ok {
		return
	}
	switch key {
	case "max_tokens":
		w.config.Defaults.MaxTokens = v
	case "top_k":
		w.config.Defaults.TopK = v
	}
}

// applyCustomBoolOutput applies a bool output setting.
func (w *WizardState) applyCustomBoolOutput(key string, value any) {
	v, ok := value.(bool)
	if !ok {
		return
	}
	switch key {
	case "return_images":
		w.config.Output.ReturnImages = v
	case "return_related":
		w.config.Output.ReturnRelated = v
	}
}

// applyCustomStringOutput applies a string output setting.
func (w *WizardState) applyCustomStringOutput(key string, value any) {
	v, ok := value.(string)
	if !ok {
		return
	}
	switch key {
	case "response_format_json_schema":
		w.config.Output.ResponseFormatJSONSchema = v
	case "response_format_regex":
		w.config.Output.ResponseFormatRegex = v
	case "reasoning_effort":
		w.config.Output.ReasoningEffort = v
	}
}

// buildConfiguration constructs the final ConfigData from wizard selections.
// Merges template defaults with user selections using a layered override strategy.
//
// Configuration precedence hierarchy (lowest to highest priority):
// 1. Empty ConfigData defaults (zero values)
// 2. Existing config (if --update mode), or template defaults
// 3. User selections from wizard (model, streaming, filters)
// 4. Custom settings (if user chose advanced customization)
// 5. API key (if provided during wizard, not inherited from environment).
func (w *WizardState) buildConfiguration() {
	if w.existingConfig != nil {
		// --update mode: start from existing config so unchanged fields are preserved.
		w.config = w.existingConfig
	} else {
		// Layer 1-2: Start with template defaults (if applicable).
		w.loadTemplateIfApplicable()
	}

	// Layer 3: Apply user selections from wizard steps.
	w.config.Defaults.Model = w.selectedModel
	w.config.Output.Stream = w.enableStream

	// Layer 3 continued: Apply search filters.
	if w.searchFilters != nil {
		w.config.Search = config.SearchConfig{}
	}
	w.applySearchFilters()

	// Layer 4: Apply custom settings (highest priority).
	w.applyCustomSettings()

	// Layer 5: Apply API key if provided.
	if w.apiKey != "" {
		w.config.API.Key = w.apiKey
	}
}

// applySearchFilters applies the selected search filters to the configuration.
func (w *WizardState) applySearchFilters() {
	for _, filter := range w.searchFilters {
		parts := strings.SplitN(filter, ":", filterSeparatorParts)
		if len(parts) != filterSeparatorParts {
			continue
		}
		w.applySearchFilter(parts[0], parts[1])
	}
}

// applySearchFilter applies a single parsed search filter key/value to the config.
func (w *WizardState) applySearchFilter(key, value string) {
	switch key {
	case "mode":
		w.config.Search.Mode = value
	case "recency":
		w.config.Search.Recency = value
	case "context":
		w.config.Search.ContextSize = value
	case "domains":
		w.config.Search.Domains = splitDomains(value)
	case "location_country":
		w.config.Search.LocationCountry = value
	case "location_lat", "location_lon":
		w.applyLocationCoord(key, value)
	case "after_date":
		w.config.Search.AfterDate = value
	case "before_date":
		w.config.Search.BeforeDate = value
	}
}

// applyLocationCoord parses a float64 coordinate value and applies it to the search config.
func (w *WizardState) applyLocationCoord(key, value string) {
	coord, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return
	}
	switch key {
	case "location_lat":
		w.config.Search.LocationLat = coord
	case "location_lon":
		w.config.Search.LocationLon = coord
	}
}

// splitDomains splits a comma-separated domain list, trimming whitespace and dropping empties.
func splitDomains(value string) []string {
	var domains []string
	for d := range strings.SplitSeq(value, ",") {
		d = strings.TrimSpace(d)
		if d != "" {
			domains = append(domains, d)
		}
	}
	return domains
}

// printSummary displays a summary of the configuration.
func (w *WizardState) printSummary() {
	_, _ = fmt.Fprintln(w.output)
	_, _ = fmt.Fprintln(w.output, "═══════════════════════════════════════════════════════════════════")
	_, _ = fmt.Fprintln(w.output, "Configuration Summary")
	_, _ = fmt.Fprintln(w.output, "═══════════════════════════════════════════════════════════════════")
	_, _ = fmt.Fprintln(w.output)
	_, _ = fmt.Fprintf(w.output, "  Use Case:    %s\n", w.getUseCaseName())
	_, _ = fmt.Fprintf(w.output, "  Model:       %s\n", w.selectedModel)
	_, _ = fmt.Fprintf(w.output, "  Streaming:   %s\n", w.formatBool(w.enableStream))
	if len(w.searchFilters) > 0 {
		_, _ = fmt.Fprintf(w.output, "  Filters:     %d configured\n", len(w.searchFilters))
	}
	if w.apiKey != "" {
		_, _ = fmt.Fprintln(w.output, "  API Key:     Configured")
	}
	if len(w.customSettings) > 0 {
		_, _ = fmt.Fprintf(w.output, "  Advanced:    %d custom settings\n", len(w.customSettings))
	}
	_, _ = fmt.Fprintln(w.output)
	_, _ = fmt.Fprintln(w.output, "═══════════════════════════════════════════════════════════════════")
	_, _ = fmt.Fprintln(w.output)
}

// getUseCaseName returns a human-readable name for the current use case.
func (w *WizardState) getUseCaseName() string {
	useCaseNames := map[string]string{
		config.TemplateResearch: "Research",
		config.TemplateCreative: "Creative",
		config.TemplateNews:     "News",
		"general":               "General",
		"custom":                "Custom",
	}

	if name, ok := useCaseNames[w.useCase]; ok {
		return name
	}
	return "Unknown"
}

// formatBool formats a boolean value for display.
func (w *WizardState) formatBool(value bool) string {
	if value {
		return "Enabled"
	}
	return "Disabled"
}

// --- Date validation helpers (unchanged) ---

// isValidDateFormat validates a date string in YYYY-MM-DD format.
func isValidDateFormat(s string) bool {
	const dateLen = 10
	if len(s) != dateLen || s[4] != '-' || s[7] != '-' {
		return false
	}
	if !dateHasOnlyDigitsAtNonSeparators(s) {
		return false
	}
	return dateComponentsInRange(s)
}

// dateHasOnlyDigitsAtNonSeparators checks that non-separator characters are all ASCII digits.
func dateHasOnlyDigitsAtNonSeparators(s string) bool {
	for i, ch := range s {
		if i == 4 || i == 7 {
			continue
		}
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

// dateComponentsInRange parses and validates the YYYY-MM-DD date components.
func dateComponentsInRange(s string) bool {
	year, err1 := strconv.Atoi(s[0:4])
	month, err2 := strconv.Atoi(s[5:7])
	day, err3 := strconv.Atoi(s[8:10])
	if err1 != nil || err2 != nil || err3 != nil {
		return false
	}
	const (
		minMonth = 1
		maxMonth = 12
		minDay   = 1
		maxDay   = 31
	)
	return year > 0 && month >= minMonth && month <= maxMonth && day >= minDay && day <= maxDay
}
