// Package cmd provides the CLI commands for pplx.
package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sgaunet/pplx/pkg/config"
)

// Wizard errors.
var (
	ErrFailedToReadInput = errors.New("failed to read input")
	ErrFailedToReadAPIKey = errors.New("failed to read API key")
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
	customSettings map[string]interface{}
}

// NewWizardState creates a new wizard state with default values.
func NewWizardState() *WizardState {
	return &WizardState{
		scanner:        bufio.NewScanner(os.Stdin),
		config:         config.NewConfigData(),
		customSettings: make(map[string]interface{}),
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

// selectSearchFilters prompts the user to select search filters.
//
//nolint:cyclop // Interactive wizard requires sequential prompt handling
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
		return ErrFailedToReadInput
	}

	response := strings.ToLower(strings.TrimSpace(w.scanner.Text()))
	if response != "y" && response != "yes" {
		fmt.Println("\n✓ Skipped search configuration (using defaults)")
		return nil
	}

	w.searchFilters = []string{}

	// Search mode
	fmt.Println()
	fmt.Print("Search mode (web/academic) [web]: ")
	if w.scanner.Scan() {
		mode := strings.TrimSpace(w.scanner.Text())
		if mode == "academic" {
			w.searchFilters = append(w.searchFilters, "mode:academic")
		}
	}

	// Recency filter
	fmt.Print("Recency filter (day/week/month/year) [skip]: ")
	if w.scanner.Scan() {
		recency := strings.TrimSpace(w.scanner.Text())
		if recency != "" && isValidRecency(recency) {
			w.searchFilters = append(w.searchFilters, "recency:"+recency)
		}
	}

	// Context size
	fmt.Print("Context size (low/medium/high) [skip]: ")
	if w.scanner.Scan() {
		contextSize := strings.TrimSpace(w.scanner.Text())
		if contextSize != "" && isValidContextSize(contextSize) {
			w.searchFilters = append(w.searchFilters, "context:"+contextSize)
		}
	}

	fmt.Printf("\n✓ Configured %d search filters\n\n", len(w.searchFilters))

	return nil
}

// configureAPIKey prompts the user to optionally provide an API key.
func (w *WizardState) configureAPIKey() error {
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println("Step 5: API Key Configuration")
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println()

	// Check if API key exists in environment
	envKey := os.Getenv("PERPLEXITY_API_KEY")
	if envKey != "" {
		fmt.Printf("✓ API key found in environment (PERPLEXITY_API_KEY)\n")
		fmt.Println("  Length:", len(envKey), "characters")
		fmt.Println()

		useEnv, err := w.promptYesNo("Use existing environment variable?", true)
		if err != nil {
			return err
		}

		if useEnv {
			fmt.Println("\n✓ Using PERPLEXITY_API_KEY from environment")
			return nil
		}
	}

	fmt.Println("You can provide your API key now or set it later via environment variable.")
	fmt.Println()

	addKey, err := w.promptYesNo("Add API key to configuration?", false)
	if err != nil {
		return err
	}

	if addKey {
		fmt.Println()
		fmt.Print("Enter your Perplexity API key: ")
		if !w.scanner.Scan() {
			return ErrFailedToReadAPIKey
		}

		w.apiKey = strings.TrimSpace(w.scanner.Text())
		if w.apiKey != "" {
			fmt.Printf("\n✓ API key added (%d characters)\n", len(w.apiKey))
		}
	} else {
		fmt.Println("\n✓ Skipped (you can set PERPLEXITY_API_KEY environment variable later)")
	}

	return nil
}

// offerCustomization asks if the user wants additional customization.
//
//nolint:cyclop // Interactive wizard requires sequential prompt handling
func (w *WizardState) offerCustomization() error {
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println("Step 6: Additional Customization (Optional)")
	fmt.Println("═══════════════════════════════════════════════════════════════════")
	fmt.Println()

	customize, err := w.promptYesNo("Configure advanced options (temperature, max tokens, etc.)?", false)
	if err != nil {
		return err
	}

	if !customize {
		fmt.Println("\n✓ Using recommended defaults")
		return nil
	}

	// Temperature
	fmt.Println()
	fmt.Print("Temperature (0.0-2.0) [0.2]: ")
	if w.scanner.Scan() {
		tempStr := strings.TrimSpace(w.scanner.Text())
		if tempStr != "" {
			if temp, err := strconv.ParseFloat(tempStr, 64); err == nil && temp >= 0 && temp <= 2 {
				w.customSettings["temperature"] = temp
			}
		}
	}

	// Max tokens
	fmt.Print("Max tokens (1-100000) [4000]: ")
	if w.scanner.Scan() {
		tokensStr := strings.TrimSpace(w.scanner.Text())
		if tokensStr != "" {
			if tokens, err := strconv.Atoi(tokensStr); err == nil && tokens > 0 && tokens <= 100000 {
				w.customSettings["max_tokens"] = tokens
			}
		}
	}

	fmt.Printf("\n✓ Configured %d advanced options\n\n", len(w.customSettings))

	return nil
}

// buildConfiguration constructs the final ConfigData from wizard selections.
func (w *WizardState) buildConfiguration() {
	// Load template if not custom
	if w.useCase != "custom" && w.useCase != "general" {
		templateCfg, err := config.LoadTemplate(w.useCase)
		if err != nil {
			// Fall back to default config if template fails
			w.config = config.NewConfigData()
		} else {
			w.config = templateCfg
		}
	}

	// Apply model selection
	w.config.Defaults.Model = w.selectedModel

	// Apply streaming preference
	w.config.Output.Stream = w.enableStream

	// Apply search filters
	w.applySearchFilters()

	// Apply custom settings
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

	// Apply API key if provided
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
			return "", ErrFailedToReadInput
		}

		choice := strings.TrimSpace(w.scanner.Text())
		for _, valid := range validChoices {
			if choice == valid {
				return choice, nil
			}
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
		return false, ErrFailedToReadInput
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

// isValidRecency checks if a recency value is valid.
func isValidRecency(value string) bool {
	validRecency := []string{"day", "week", "month", "year", "hour"}
	for _, valid := range validRecency {
		if value == valid {
			return true
		}
	}
	return false
}

// isValidContextSize checks if a context size value is valid.
func isValidContextSize(value string) bool {
	validSizes := []string{"low", "medium", "high"}
	for _, valid := range validSizes {
		if value == valid {
			return true
		}
	}
	return false
}
