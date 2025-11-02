package cmd

import (
	"bufio"
	"strings"
	"testing"

	"github.com/sgaunet/pplx/pkg/config"
)

// mockScanner creates a bufio.Scanner from a string for testing.
func mockScanner(input string) *bufio.Scanner {
	return bufio.NewScanner(strings.NewReader(input))
}

// TestSelectUseCase tests the use case selection step.
func TestSelectUseCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		input         string
		expectedUse   string
		expectedError bool
	}{
		{
			name:          "select research",
			input:         "1\n",
			expectedUse:   config.TemplateResearch,
			expectedError: false,
		},
		{
			name:          "select creative",
			input:         "2\n",
			expectedUse:   config.TemplateCreative,
			expectedError: false,
		},
		{
			name:          "select news",
			input:         "3\n",
			expectedUse:   config.TemplateNews,
			expectedError: false,
		},
		{
			name:          "select general",
			input:         "4\n",
			expectedUse:   "general",
			expectedError: false,
		},
		{
			name:          "select custom",
			input:         "5\n",
			expectedUse:   "custom",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := NewWizardState()
			w.scanner = mockScanner(tt.input)

			err := w.selectUseCase()

			if (err != nil) != tt.expectedError {
				t.Errorf("selectUseCase() error = %v, expectedError %v", err, tt.expectedError)
			}

			if !tt.expectedError && w.useCase != tt.expectedUse {
				t.Errorf("selectUseCase() useCase = %v, want %v", w.useCase, tt.expectedUse)
			}
		})
	}
}

// TestSelectModel tests the model selection step.
func TestSelectModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		input         string
		expectedModel string
		expectedError bool
	}{
		{
			name:          "select sonar",
			input:         "1\n",
			expectedModel: "sonar",
			expectedError: false,
		},
		{
			name:          "select sonar-pro",
			input:         "2\n",
			expectedModel: "sonar-pro",
			expectedError: false,
		},
		{
			name:          "select sonar-reasoning",
			input:         "3\n",
			expectedModel: "sonar-reasoning",
			expectedError: false,
		},
		{
			name:          "select sonar-deep-research",
			input:         "4\n",
			expectedModel: "sonar-deep-research",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := NewWizardState()
			w.scanner = mockScanner(tt.input)

			err := w.selectModel()

			if (err != nil) != tt.expectedError {
				t.Errorf("selectModel() error = %v, expectedError %v", err, tt.expectedError)
			}

			if !tt.expectedError && w.selectedModel != tt.expectedModel {
				t.Errorf("selectModel() selectedModel = %v, want %v", w.selectedModel, tt.expectedModel)
			}
		})
	}
}

// TestConfigureStreaming tests the streaming configuration step.
func TestConfigureStreaming(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		input          string
		expectedStream bool
		expectedError  bool
	}{
		{
			name:           "enable streaming yes",
			input:          "y\n",
			expectedStream: true,
			expectedError:  false,
		},
		{
			name:           "enable streaming yes full",
			input:          "yes\n",
			expectedStream: true,
			expectedError:  false,
		},
		{
			name:           "disable streaming no",
			input:          "n\n",
			expectedStream: false,
			expectedError:  false,
		},
		{
			name:           "disable streaming no full",
			input:          "no\n",
			expectedStream: false,
			expectedError:  false,
		},
		{
			name:           "default yes with empty input",
			input:          "\n",
			expectedStream: true,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := NewWizardState()
			w.scanner = mockScanner(tt.input)

			err := w.configureStreaming()

			if (err != nil) != tt.expectedError {
				t.Errorf("configureStreaming() error = %v, expectedError %v", err, tt.expectedError)
			}

			if !tt.expectedError && w.enableStream != tt.expectedStream {
				t.Errorf("configureStreaming() enableStream = %v, want %v", w.enableStream, tt.expectedStream)
			}
		})
	}
}

// TestPromptChoice tests the choice prompt utility.
func TestPromptChoice(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		input          string
		validChoices   []string
		expectedChoice string
		expectedError  bool
	}{
		{
			name:           "valid choice 1",
			input:          "1\n",
			validChoices:   []string{"1", "2", "3"},
			expectedChoice: "1",
			expectedError:  false,
		},
		{
			name:           "valid choice 3",
			input:          "3\n",
			validChoices:   []string{"1", "2", "3"},
			expectedChoice: "3",
			expectedError:  false,
		},
		{
			name:           "retry on invalid then valid",
			input:          "99\n1\n",
			validChoices:   []string{"1", "2", "3"},
			expectedChoice: "1",
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := NewWizardState()
			w.scanner = mockScanner(tt.input)

			choice, err := w.promptChoice("Test prompt", tt.validChoices)

			if (err != nil) != tt.expectedError {
				t.Errorf("promptChoice() error = %v, expectedError %v", err, tt.expectedError)
			}

			if !tt.expectedError && choice != tt.expectedChoice {
				t.Errorf("promptChoice() choice = %v, want %v", choice, tt.expectedChoice)
			}
		})
	}
}

// TestPromptYesNo tests the yes/no prompt utility.
func TestPromptYesNo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		input          string
		defaultYes     bool
		expectedResult bool
		expectedError  bool
	}{
		{
			name:           "yes response",
			input:          "y\n",
			defaultYes:     false,
			expectedResult: true,
			expectedError:  false,
		},
		{
			name:           "no response",
			input:          "n\n",
			defaultYes:     true,
			expectedResult: false,
			expectedError:  false,
		},
		{
			name:           "default yes with empty",
			input:          "\n",
			defaultYes:     true,
			expectedResult: true,
			expectedError:  false,
		},
		{
			name:           "default no with empty",
			input:          "\n",
			defaultYes:     false,
			expectedResult: false,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := NewWizardState()
			w.scanner = mockScanner(tt.input)

			result, err := w.promptYesNo("Test prompt", tt.defaultYes)

			if (err != nil) != tt.expectedError {
				t.Errorf("promptYesNo() error = %v, expectedError %v", err, tt.expectedError)
			}

			if !tt.expectedError && result != tt.expectedResult {
				t.Errorf("promptYesNo() result = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

// TestBuildConfiguration tests the configuration building logic.
func TestBuildConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		useCase        string
		selectedModel  string
		enableStream   bool
		apiKey         string
		customSettings map[string]interface{}
		expectedModel  string
		expectedStream bool
	}{
		{
			name:           "research template with overrides",
			useCase:        config.TemplateResearch,
			selectedModel:  "sonar-pro",
			enableStream:   true,
			apiKey:         "test-key",
			customSettings: map[string]interface{}{"temperature": 0.5},
			expectedModel:  "sonar-pro",
			expectedStream: true,
		},
		{
			name:           "custom config from scratch",
			useCase:        "custom",
			selectedModel:  "sonar",
			enableStream:   false,
			apiKey:         "",
			customSettings: map[string]interface{}{},
			expectedModel:  "sonar",
			expectedStream: false,
		},
		{
			name:           "general config",
			useCase:        "general",
			selectedModel:  "sonar-reasoning",
			enableStream:   true,
			apiKey:         "",
			customSettings: map[string]interface{}{},
			expectedModel:  "sonar-reasoning",
			expectedStream: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := NewWizardState()
			w.useCase = tt.useCase
			w.selectedModel = tt.selectedModel
			w.enableStream = tt.enableStream
			w.apiKey = tt.apiKey
			w.customSettings = tt.customSettings

			w.buildConfiguration()

			if w.config == nil {
				t.Fatal("buildConfiguration() returned nil config")
			}

			if w.config.Defaults.Model != tt.expectedModel {
				t.Errorf("buildConfiguration() model = %v, want %v", w.config.Defaults.Model, tt.expectedModel)
			}

			if w.config.Output.Stream != tt.expectedStream {
				t.Errorf("buildConfiguration() stream = %v, want %v", w.config.Output.Stream, tt.expectedStream)
			}

			if tt.apiKey != "" && w.config.API.Key != tt.apiKey {
				t.Errorf("buildConfiguration() apiKey = %v, want %v", w.config.API.Key, tt.apiKey)
			}

			// Test custom settings application
			if temp, ok := tt.customSettings["temperature"]; ok {
				if w.config.Defaults.Temperature != temp.(float64) {
					t.Errorf("buildConfiguration() temperature = %v, want %v", w.config.Defaults.Temperature, temp)
				}
			}
		})
	}
}

// TestApplySearchFilters tests the search filter application logic.
func TestApplySearchFilters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		searchFilters  []string
		expectedMode   string
		expectedRecency string
		expectedContext string
	}{
		{
			name:           "academic mode filter",
			searchFilters:  []string{"mode:academic"},
			expectedMode:   "academic",
			expectedRecency: "",
			expectedContext: "",
		},
		{
			name:           "recency filter",
			searchFilters:  []string{"recency:week"},
			expectedMode:   "",
			expectedRecency: "week",
			expectedContext: "",
		},
		{
			name:           "context size filter",
			searchFilters:  []string{"context:high"},
			expectedMode:   "",
			expectedRecency: "",
			expectedContext: "high",
		},
		{
			name:           "multiple filters",
			searchFilters:  []string{"mode:academic", "recency:month", "context:medium"},
			expectedMode:   "academic",
			expectedRecency: "month",
			expectedContext: "medium",
		},
		{
			name:           "no filters",
			searchFilters:  []string{},
			expectedMode:   "",
			expectedRecency: "",
			expectedContext: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := NewWizardState()
			w.searchFilters = tt.searchFilters

			w.applySearchFilters()

			if w.config.Search.Mode != tt.expectedMode {
				t.Errorf("applySearchFilters() mode = %v, want %v", w.config.Search.Mode, tt.expectedMode)
			}

			if w.config.Search.Recency != tt.expectedRecency {
				t.Errorf("applySearchFilters() recency = %v, want %v", w.config.Search.Recency, tt.expectedRecency)
			}

			if w.config.Search.ContextSize != tt.expectedContext {
				t.Errorf("applySearchFilters() contextSize = %v, want %v", w.config.Search.ContextSize, tt.expectedContext)
			}
		})
	}
}

// TestIsValidRecency tests the recency validation function.
func TestIsValidRecency(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{name: "valid day", value: "day", expected: true},
		{name: "valid week", value: "week", expected: true},
		{name: "valid month", value: "month", expected: true},
		{name: "valid year", value: "year", expected: true},
		{name: "valid hour", value: "hour", expected: true},
		{name: "invalid value", value: "invalid", expected: false},
		{name: "empty string", value: "", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := isValidRecency(tt.value)
			if result != tt.expected {
				t.Errorf("isValidRecency(%q) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

// TestIsValidContextSize tests the context size validation function.
func TestIsValidContextSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{name: "valid low", value: "low", expected: true},
		{name: "valid medium", value: "medium", expected: true},
		{name: "valid high", value: "high", expected: true},
		{name: "invalid value", value: "invalid", expected: false},
		{name: "empty string", value: "", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := isValidContextSize(tt.value)
			if result != tt.expected {
				t.Errorf("isValidContextSize(%q) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

// TestGetUseCaseName tests the use case name mapping.
func TestGetUseCaseName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		useCase      string
		expectedName string
	}{
		{name: "research", useCase: config.TemplateResearch, expectedName: "Research"},
		{name: "creative", useCase: config.TemplateCreative, expectedName: "Creative"},
		{name: "news", useCase: config.TemplateNews, expectedName: "News"},
		{name: "general", useCase: "general", expectedName: "General"},
		{name: "custom", useCase: "custom", expectedName: "Custom"},
		{name: "unknown", useCase: "unknown", expectedName: "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := NewWizardState()
			w.useCase = tt.useCase

			result := w.getUseCaseName()
			if result != tt.expectedName {
				t.Errorf("getUseCaseName() = %v, want %v", result, tt.expectedName)
			}
		})
	}
}

// TestFormatBool tests the boolean formatting function.
func TestFormatBool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    bool
		expected string
	}{
		{name: "true", value: true, expected: "Enabled"},
		{name: "false", value: false, expected: "Disabled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := NewWizardState()
			result := w.formatBool(tt.value)
			if result != tt.expected {
				t.Errorf("formatBool(%v) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

// TestNewWizardState tests the wizard state initialization.
func TestNewWizardState(t *testing.T) {
	t.Parallel()

	w := NewWizardState()

	if w == nil {
		t.Fatal("NewWizardState() returned nil")
	}

	if w.scanner == nil {
		t.Error("NewWizardState() scanner is nil")
	}

	if w.config == nil {
		t.Error("NewWizardState() config is nil")
	}

	if w.customSettings == nil {
		t.Error("NewWizardState() customSettings is nil")
	}
}

// TestConfigureAPIKey tests API key configuration with environment variable.
func TestConfigureAPIKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		input         string
		envKey        string
		expectedKey   string
		expectedError bool
	}{
		{
			name:          "add API key",
			input:         "y\ntest-api-key-123\n",
			envKey:        "",
			expectedKey:   "test-api-key-123",
			expectedError: false,
		},
		{
			name:          "skip API key",
			input:         "n\n",
			envKey:        "",
			expectedKey:   "",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Set environment variable if specified
			if tt.envKey != "" {
				t.Setenv("PERPLEXITY_API_KEY", tt.envKey)
			}

			w := NewWizardState()
			w.scanner = mockScanner(tt.input)

			err := w.configureAPIKey()

			if (err != nil) != tt.expectedError {
				t.Errorf("configureAPIKey() error = %v, expectedError %v", err, tt.expectedError)
			}

			if !tt.expectedError && w.apiKey != tt.expectedKey {
				t.Errorf("configureAPIKey() apiKey = %v, want %v", w.apiKey, tt.expectedKey)
			}
		})
	}
}

// TestOfferCustomization tests the customization step.
func TestOfferCustomization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		input              string
		expectedTemp       float64
		expectedMaxTokens  int
		expectedError      bool
	}{
		{
			name:               "skip customization",
			input:              "n\n",
			expectedTemp:       0,
			expectedMaxTokens:  0,
			expectedError:      false,
		},
		{
			name:               "add temperature",
			input:              "y\n0.8\n\n",
			expectedTemp:       0.8,
			expectedMaxTokens:  0,
			expectedError:      false,
		},
		{
			name:               "add max tokens",
			input:              "y\n\n5000\n",
			expectedTemp:       0,
			expectedMaxTokens:  5000,
			expectedError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := NewWizardState()
			w.scanner = mockScanner(tt.input)

			err := w.offerCustomization()

			if (err != nil) != tt.expectedError {
				t.Errorf("offerCustomization() error = %v, expectedError %v", err, tt.expectedError)
			}

			if !tt.expectedError {
				if temp, ok := w.customSettings["temperature"]; ok {
					if temp.(float64) != tt.expectedTemp {
						t.Errorf("offerCustomization() temperature = %v, want %v", temp, tt.expectedTemp)
					}
				}

				if tokens, ok := w.customSettings["max_tokens"]; ok {
					if tokens.(int) != tt.expectedMaxTokens {
						t.Errorf("offerCustomization() max_tokens = %v, want %v", tokens, tt.expectedMaxTokens)
					}
				}
			}
		})
	}
}

// TestTemplateLoadingFallback tests that wizard falls back to default config when template loading fails.
func TestTemplateLoadingFallback(t *testing.T) {
	t.Parallel()

	w := NewWizardState()
	w.useCase = "invalid-template"
	w.selectedModel = "sonar"
	w.enableStream = true

	w.buildConfiguration()

	// Should have created a default config even though template loading failed
	if w.config == nil {
		t.Fatal("buildConfiguration() with invalid template returned nil config")
	}

	// Model should still be applied
	if w.config.Defaults.Model != "sonar" {
		t.Errorf("buildConfiguration() fallback model = %v, want sonar", w.config.Defaults.Model)
	}

	// Stream should still be applied
	if w.config.Output.Stream != true {
		t.Error("buildConfiguration() fallback did not apply streaming preference")
	}
}
