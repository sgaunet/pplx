package cmd

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/sgaunet/pplx/pkg/clerrors"
	"github.com/sgaunet/pplx/pkg/config"
)

const (
	filterSeparatorParts = 2
)

// WizardState manages the state and flow of the interactive configuration wizard.
type WizardState struct {
	// Scanner for reading user input
	scanner *bufio.Scanner

	// Configuration being built
	config *config.ConfigData

	// User selections
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
		scanner:        bufio.NewScanner(os.Stdin),
		config:         config.NewConfigData(),
		customSettings: make(map[string]any),
	}
}

// Run executes the wizard flow and returns the configured ConfigData.
func (w *WizardState) Run() (*config.ConfigData, error) {
	// Print welcome banner
	w.printWelcome()

	// Step 1: Use case selection
	if err := w.selectUseCase(); err != nil {
		return nil, fmt.Errorf("use case selection failed: %w", err)
	}

	// Step 2: Model selection
	if err := w.selectModel(); err != nil {
		return nil, fmt.Errorf("model selection failed: %w", err)
	}

	// Step 3: Streaming toggle
	if err := w.configureStreaming(); err != nil {
		return nil, fmt.Errorf("streaming configuration failed: %w", err)
	}

	// Step 4: Search filters
	if err := w.selectSearchFilters(); err != nil {
		return nil, fmt.Errorf("search filter selection failed: %w", err)
	}

	// Step 5: API key
	if err := w.configureAPIKey(); err != nil {
		return nil, fmt.Errorf("API key configuration failed: %w", err)
	}

	// Step 6: Additional customization
	if err := w.offerCustomization(); err != nil {
		return nil, fmt.Errorf("customization failed: %w", err)
	}

	// Step 7: Generate final configuration
	w.buildConfiguration()

	// Print summary
	w.printSummary()

	return w.config, nil
}

// printWelcome displays the wizard welcome banner.
func (w *WizardState) printWelcome() {
	fmt.Println()
	fmt.Println("╔═══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                                                                   ║")
	fmt.Println("║          Welcome to the pplx Configuration Wizard!               ║")
	fmt.Println("║                                                                   ║")
	fmt.Println("║  This wizard will guide you through creating a personalized      ║")
	fmt.Println("║  configuration for your Perplexity AI CLI experience.            ║")
	fmt.Println("║                                                                   ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

// selectUseCase prompts the user to select their primary use case.
func (w *WizardState) selectUseCase() error {
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println("Step 1: Select Your Primary Use Case")
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("Choose the configuration that best matches your needs:")
	fmt.Println()
	fmt.Println("  1. Research     - Academic and scholarly work with authoritative sources")
	fmt.Println("  2. Creative     - Content generation, writing, and brainstorming")
	fmt.Println("  3. News         - Current events tracking with reputable news sources")
	fmt.Println("  4. General      - Balanced configuration for everyday queries")
	fmt.Println("  5. Custom       - Start from scratch with full customization")
	fmt.Println()

	choice, err := w.promptChoice("Enter your choice (1-5)", []string{"1", "2", "3", "4", "5"})
	if err != nil {
		return err
	}

	// Map choice to use case
	useCaseMap := map[string]string{
		"1": config.TemplateResearch,
		"2": config.TemplateCreative,
		"3": config.TemplateNews,
		"4": "general",
		"5": "custom",
	}

	w.useCase = useCaseMap[choice]
	fmt.Printf("\n✓ Selected: %s\n\n", w.getUseCaseName())

	return nil
}

// selectModel prompts the user to select their preferred model.
func (w *WizardState) selectModel() error {
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println("Step 2: Choose Your AI Model")
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("Available models:")
	fmt.Println()
	fmt.Println("  1. sonar                - Balanced performance (recommended)")
	fmt.Println("  2. sonar-pro            - Enhanced capabilities")
	fmt.Println("  3. sonar-reasoning      - Advanced reasoning")
	fmt.Println("  4. sonar-deep-research  - In-depth research and analysis")
	fmt.Println()

	choice, err := w.promptChoice("Enter your choice (1-4)", []string{"1", "2", "3", "4"})
	if err != nil {
		return err
	}

	modelMap := map[string]string{
		"1": "sonar",
		"2": "sonar-pro",
		"3": "sonar-reasoning",
		"4": "sonar-deep-research",
	}

	w.selectedModel = modelMap[choice]
	fmt.Printf("\n✓ Selected model: %s\n\n", w.selectedModel)

	return nil
}

// configureStreaming prompts the user to enable/disable streaming.
func (w *WizardState) configureStreaming() error {
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println("Step 3: Configure Response Streaming")
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("Streaming displays responses as they're generated (real-time output).")
	fmt.Println()

	response, err := w.promptYesNo("Enable streaming?", true)
	if err != nil {
		return err
	}

	w.enableStream = response
	fmt.Printf("\n✓ Streaming: %s\n\n", w.formatBool(response))

	return nil
}

// promptSearchMode prompts for search mode (web/academic).
func (w *WizardState) promptSearchMode() error {
	fmt.Println()
	fmt.Print("Search mode (web/academic) [web]: ")
	if !w.scanner.Scan() {
		return clerrors.ErrFailedToReadInput
	}

	mode := strings.TrimSpace(w.scanner.Text())
	if mode == "academic" {
		w.searchFilters = append(w.searchFilters, "mode:academic")
	}

	return nil
}

// promptRecencyFilter prompts for recency filter.
func (w *WizardState) promptRecencyFilter() error {
	fmt.Print("Recency filter (day/week/month/year) [skip]: ")
	if !w.scanner.Scan() {
		return clerrors.ErrFailedToReadInput
	}

	recency := strings.TrimSpace(w.scanner.Text())
	if recency != "" && isValidRecency(recency) {
		w.searchFilters = append(w.searchFilters, "recency:"+recency)
	}

	return nil
}

// promptContextSize prompts for search context size.
func (w *WizardState) promptContextSize() error {
	fmt.Print("Context size (low/medium/high) [skip]: ")
	if !w.scanner.Scan() {
		return clerrors.ErrFailedToReadInput
	}

	contextSize := strings.TrimSpace(w.scanner.Text())
	if contextSize != "" && isValidContextSize(contextSize) {
		w.searchFilters = append(w.searchFilters, "context:"+contextSize)
	}

	return nil
}

// selectSearchFilters prompts the user to select search filters.
func (w *WizardState) selectSearchFilters() error {
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println("Step 4: Configure Search Preferences")
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("Select search filters (optional, press Enter to skip):")
	fmt.Println()
	fmt.Println("  1. Search mode        - Web or academic search")
	fmt.Println("  2. Recency filter     - Time-based filtering (day, week, month, year)")
	fmt.Println("  3. Domain filtering   - Restrict to specific domains")
	fmt.Println("  4. Context size       - Search context depth (low, medium, high)")
	fmt.Println()

	fmt.Print("Configure search preferences? (y/N): ")
	if !w.scanner.Scan() {
		return clerrors.ErrFailedToReadInput
	}

	response := strings.ToLower(strings.TrimSpace(w.scanner.Text()))
	if response != "y" && response != "yes" {
		fmt.Println("\n✓ Skipped search configuration (using defaults)")
		return nil
	}

	w.searchFilters = []string{}

	if err := w.promptSearchMode(); err != nil {
		return err
	}

	if err := w.promptRecencyFilter(); err != nil {
		return err
	}

	if err := w.promptContextSize(); err != nil {
		return err
	}

	fmt.Printf("\n✓ Configured %d search filters\n\n", len(w.searchFilters))

	return nil
}

// printAPIKeyHeader prints the API key section header.
func (w *WizardState) printAPIKeyHeader() {
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println("Step 5: API Key Configuration")
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println()
}

// checkEnvironmentAPIKey checks for API key in environment and prompts to use it.
// Returns (shouldUseEnv, foundEnv, error).
func (w *WizardState) checkEnvironmentAPIKey() (bool, bool, error) {
	envKey := os.Getenv("PERPLEXITY_API_KEY")
	if envKey == "" {
		return false, false, nil
	}

	fmt.Printf("✓ API key found in environment (PERPLEXITY_API_KEY)\n")
	fmt.Println("  Length:", len(envKey), "characters")
	fmt.Println()

	useEnv, err := w.promptYesNo("Use existing environment variable?", true)
	if err != nil {
		return false, true, err
	}

	return useEnv, true, nil
}

// readAndStoreAPIKey reads and stores an API key from user input.
func (w *WizardState) readAndStoreAPIKey() error {
	fmt.Println()
	fmt.Print("Enter your Perplexity API key: ")
	if !w.scanner.Scan() {
		return clerrors.ErrFailedToReadAPIKey
	}

	w.apiKey = strings.TrimSpace(w.scanner.Text())
	if w.apiKey != "" {
		fmt.Printf("\n✓ API key added (%d characters)\n", len(w.apiKey))
	}

	return nil
}

// configureAPIKey prompts the user to optionally provide an API key.
func (w *WizardState) configureAPIKey() error {
	w.printAPIKeyHeader()

	shouldUseEnv, foundEnv, err := w.checkEnvironmentAPIKey()
	if err != nil {
		return err
	}

	if shouldUseEnv && foundEnv {
		fmt.Println("\n✓ Using PERPLEXITY_API_KEY from environment")
		return nil
	}

	fmt.Println("You can provide your API key now or set it later via environment variable.")
	fmt.Println()

	addKey, err := w.promptYesNo("Add API key to configuration?", false)
	if err != nil {
		return err
	}

	if !addKey {
		fmt.Println("\n✓ Skipped (you can set PERPLEXITY_API_KEY environment variable later)")
		return nil
	}

	return w.readAndStoreAPIKey()
}

// printCustomizationHeader prints the customization section header.
func (w *WizardState) printCustomizationHeader() {
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println("Step 6: Additional Customization (Optional)")
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println()
}

// promptTemperature prompts for temperature customization.
func (w *WizardState) promptTemperature() error {
	const minTemp, maxTemp, defaultTemp = 0.0, 2.0, 0.2

	fmt.Println()
	temp, wasSet, err := w.promptFloat("Temperature (0.0-2.0) [0.2]: ", minTemp, maxTemp, defaultTemp)
	if err != nil {
		return err
	}

	if wasSet {
		w.customSettings["temperature"] = temp
	}

	return nil
}

// promptMaxTokens prompts for max_tokens customization.
func (w *WizardState) promptMaxTokens() error {
	const minTokens, maxTokens, defaultTokens = 1, 100000, 4000

	tokens, wasSet, err := w.promptInt("Max tokens (1-100000) [4000]: ", minTokens, maxTokens, defaultTokens)
	if err != nil {
		return err
	}

	if wasSet {
		w.customSettings["max_tokens"] = tokens
	}

	return nil
}

// offerCustomization asks if the user wants additional customization.
func (w *WizardState) offerCustomization() error {
	w.printCustomizationHeader()

	customize, err := w.promptYesNo("Configure advanced options (temperature, max tokens, etc.)?", false)
	if err != nil {
		return err
	}

	if !customize {
		fmt.Println("\n✓ Using recommended defaults")
		return nil
	}

	if err := w.promptTemperature(); err != nil {
		return err
	}

	if err := w.promptMaxTokens(); err != nil {
		return err
	}

	fmt.Printf("\n✓ Configured %d advanced options\n\n", len(w.customSettings))

	return nil
}

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
		switch key {
		case "temperature":
			if temp, ok := value.(float64); ok {
				w.config.Defaults.Temperature = temp
			}
		case "max_tokens":
			if tokens, ok := value.(int); ok {
				w.config.Defaults.MaxTokens = tokens
			}
		}
	}
}

// buildConfiguration constructs the final ConfigData from wizard selections.
func (w *WizardState) buildConfiguration() {
	w.loadTemplateIfApplicable()

	w.config.Defaults.Model = w.selectedModel
	w.config.Output.Stream = w.enableStream

	w.applySearchFilters()
	w.applyCustomSettings()

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

		key, value := parts[0], parts[1]
		switch key {
		case "mode":
			w.config.Search.Mode = value
		case "recency":
			w.config.Search.Recency = value
		case "context":
			w.config.Search.ContextSize = value
		}
	}
}

// printSummary displays a summary of the configuration.
func (w *WizardState) printSummary() {
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println("Configuration Summary")
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Printf("  Use Case:    %s\n", w.getUseCaseName())
	fmt.Printf("  Model:       %s\n", w.selectedModel)
	fmt.Printf("  Streaming:   %s\n", w.formatBool(w.enableStream))
	if len(w.searchFilters) > 0 {
		fmt.Printf("  Filters:     %d configured\n", len(w.searchFilters))
	}
	if w.apiKey != "" {
		fmt.Println("  API Key:     Configured")
	}
	if len(w.customSettings) > 0 {
		fmt.Printf("  Advanced:    %d custom settings\n", len(w.customSettings))
	}
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println()
}

// promptChoice prompts for a choice from valid options.
func (w *WizardState) promptChoice(prompt string, validChoices []string) (string, error) {
	for {
		fmt.Printf("%s: ", prompt)
		if !w.scanner.Scan() {
			return "", clerrors.ErrFailedToReadInput
		}

		choice := strings.TrimSpace(w.scanner.Text())
		if slices.Contains(validChoices, choice) {
			return choice, nil
		}

		fmt.Printf("Invalid choice. Please select from: %s\n", strings.Join(validChoices, ", "))
	}
}

// promptYesNo prompts for a yes/no response.
func (w *WizardState) promptYesNo(prompt string, defaultYes bool) (bool, error) {
	defaultIndicator := "(y/N)"
	if defaultYes {
		defaultIndicator = "(Y/n)"
	}

	fmt.Printf("%s %s: ", prompt, defaultIndicator)
	if !w.scanner.Scan() {
		return false, clerrors.ErrFailedToReadInput
	}

	response := strings.ToLower(strings.TrimSpace(w.scanner.Text()))

	// Handle empty response (use default)
	if response == "" {
		return defaultYes, nil
	}

	return response == "y" || response == "yes", nil
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

// promptFloat prompts for a floating-point value with range validation.
// Returns (value, wasSet, error) where wasSet indicates non-empty input.
// Loops until valid input or empty string (uses default).
func (w *WizardState) promptFloat(prompt string, minVal, maxVal, defaultVal float64) (float64, bool, error) {
	for {
		fmt.Print(prompt)
		if !w.scanner.Scan() {
			return 0, false, clerrors.ErrFailedToReadInput
		}

		input := strings.TrimSpace(w.scanner.Text())
		if input == "" {
			return defaultVal, false, nil
		}

		value, err := strconv.ParseFloat(input, 64)
		if err != nil {
			fmt.Printf("Invalid number. Please enter a value between %.1f and %.1f\n", minVal, maxVal)
			continue
		}

		if value < minVal || value > maxVal {
			fmt.Printf("Value out of range. Please enter a value between %.1f and %.1f\n", minVal, maxVal)
			continue
		}

		return value, true, nil
	}
}

// promptInt prompts for an integer value with range validation.
// Returns (value, wasSet, error) where wasSet indicates non-empty input.
// Loops until valid input or empty string (uses default).
func (w *WizardState) promptInt(prompt string, minVal, maxVal, defaultVal int) (int, bool, error) {
	for {
		fmt.Print(prompt)
		if !w.scanner.Scan() {
			return 0, false, clerrors.ErrFailedToReadInput
		}

		input := strings.TrimSpace(w.scanner.Text())
		if input == "" {
			return defaultVal, false, nil
		}

		value, err := strconv.Atoi(input)
		if err != nil {
			fmt.Printf("Invalid number. Please enter a value between %d and %d\n", minVal, maxVal)
			continue
		}

		if value < minVal || value > maxVal {
			fmt.Printf("Value out of range. Please enter a value between %d and %d\n", minVal, maxVal)
			continue
		}

		return value, true, nil
	}
}

// isValidRecency checks if a recency value is valid.
func isValidRecency(value string) bool {
	validRecency := []string{"day", "week", "month", "year", "hour"}
	return slices.Contains(validRecency, value)
}

// isValidContextSize checks if a context size value is valid.
func isValidContextSize(value string) bool {
	validSizes := []string{"low", "medium", "high"}
	return slices.Contains(validSizes, value)
}
