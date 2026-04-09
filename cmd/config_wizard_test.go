package cmd

import (
	"io"
	"strings"
	"testing"
	"testing/iotest"

	"github.com/sgaunet/pplx/pkg/config"
)

// newTestWizard creates a WizardState configured for accessible mode with injected I/O.
// In accessible mode, huh forms read from the provided reader:
// - Select: expects 1-based number (e.g. "1\n" for first option)
// - Confirm: expects "y\n" or "n\n" (empty line keeps pre-set value)
// - Input: expects text followed by newline (empty line = empty string)
//
// We wrap with iotest.OneByteReader to prevent bufio.Scanner from buffering ahead.
// Without this, each form's scanner would consume the entire reader in one Read call,
// and subsequent forms would see EOF.
func newTestWizard(input string) *WizardState {
	return &WizardState{
		input:          iotest.OneByteReader(strings.NewReader(input)),
		output:         io.Discard,
		accessible:     true,
		config:         config.NewConfigData(),
		customSettings: make(map[string]any),
	}
}

// TestSelectUseCase tests the use case selection step.
func TestSelectUseCase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		expectedUse string
	}{
		{name: "select research", input: "1\n", expectedUse: config.TemplateResearch},
		{name: "select creative", input: "2\n", expectedUse: config.TemplateCreative},
		{name: "select news", input: "3\n", expectedUse: config.TemplateNews},
		{name: "select general", input: "4\n", expectedUse: "general"},
		{name: "select custom", input: "5\n", expectedUse: "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := newTestWizard(tt.input)
			err := w.selectUseCase()

			if err != nil {
				t.Fatalf("selectUseCase() error = %v", err)
			}

			if w.useCase != tt.expectedUse {
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
	}{
		{name: "select sonar", input: "1\n", expectedModel: "sonar"},
		{name: "select sonar-pro", input: "2\n", expectedModel: "sonar-pro"},
		{name: "select sonar-reasoning", input: "3\n", expectedModel: "sonar-reasoning"},
		{name: "select sonar-deep-research", input: "4\n", expectedModel: "sonar-deep-research"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := newTestWizard(tt.input)
			err := w.selectModel()

			if err != nil {
				t.Fatalf("selectModel() error = %v", err)
			}

			if w.selectedModel != tt.expectedModel {
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
	}{
		{name: "enable streaming yes", input: "y\n", expectedStream: true},
		{name: "disable streaming no", input: "n\n", expectedStream: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := newTestWizard(tt.input)
			err := w.configureStreaming()

			if err != nil {
				t.Fatalf("configureStreaming() error = %v", err)
			}

			if w.enableStream != tt.expectedStream {
				t.Errorf("configureStreaming() enableStream = %v, want %v", w.enableStream, tt.expectedStream)
			}
		})
	}
}

// TestSelectSearchFiltersSkip tests skipping search configuration.
func TestSelectSearchFiltersSkip(t *testing.T) {
	t.Parallel()

	// "n\n" to skip search config
	w := newTestWizard("n\n")
	err := w.selectSearchFilters()

	if err != nil {
		t.Fatalf("selectSearchFilters() error = %v", err)
	}

	if w.searchFilters != nil {
		t.Errorf("selectSearchFilters() searchFilters = %v, want nil", w.searchFilters)
	}
}

// TestSelectSearchFiltersWithOptions tests configuring search filters.
func TestSelectSearchFiltersWithOptions(t *testing.T) {
	t.Parallel()

	// Input sequence:
	// 1. Confirm search config: "y\n"
	// 2. Search mode: "2\n" (Academic - option 2)
	// 3. Recency: "4\n" (Week - option 4: No filter, Hour, Day, Week)
	// 4. Context: "1\n" (No preference/default - option 1)
	// 5. Domains input: "\n" (skip)
	// 6. Location gate: "n\n" (skip)
	// 7. Date range gate: "n\n" (skip)
	w := newTestWizard("y\n2\n4\n1\n\nn\nn\n")
	err := w.selectSearchFilters()

	if err != nil {
		t.Fatalf("selectSearchFilters() error = %v", err)
	}

	if len(w.searchFilters) != 2 {
		t.Fatalf("selectSearchFilters() searchFilters count = %d, want 2, got: %v", len(w.searchFilters), w.searchFilters)
	}

	expectedFilters := []string{"mode:academic", "recency:week"}
	for i, expected := range expectedFilters {
		if w.searchFilters[i] != expected {
			t.Errorf("searchFilters[%d] = %q, want %q", i, w.searchFilters[i], expected)
		}
	}
}

// TestConfigureAPIKeySkip tests skipping API key configuration.
func TestConfigureAPIKeySkip(t *testing.T) {
	t.Parallel()

	// "n\n" to skip adding API key
	w := newTestWizard("n\n")
	err := w.configureAPIKey()

	if err != nil {
		t.Fatalf("configureAPIKey() error = %v", err)
	}

	if w.apiKey != "" {
		t.Errorf("configureAPIKey() apiKey = %v, want empty", w.apiKey)
	}
}

// TestConfigureAPIKeyAdd tests adding an API key.
func TestConfigureAPIKeyAdd(t *testing.T) {
	t.Parallel()

	// "y\n" to add key, then "test-api-key-123\n" as the key
	w := newTestWizard("y\ntest-api-key-123\n")
	err := w.configureAPIKey()

	if err != nil {
		t.Fatalf("configureAPIKey() error = %v", err)
	}

	if w.apiKey != "test-api-key-123" {
		t.Errorf("configureAPIKey() apiKey = %v, want test-api-key-123", w.apiKey)
	}
}

// TestOfferCustomizationSkip tests skipping customization.
func TestOfferCustomizationSkip(t *testing.T) {
	t.Parallel()

	w := newTestWizard("n\n")
	err := w.offerCustomization()

	if err != nil {
		t.Fatalf("offerCustomization() error = %v", err)
	}

	if len(w.customSettings) != 0 {
		t.Errorf("offerCustomization() customSettings = %v, want empty", w.customSettings)
	}
}

// TestOfferCustomizationWithValues tests customization with temperature and max tokens.
func TestOfferCustomizationWithValues(t *testing.T) {
	t.Parallel()

	// "y\n" to customize, "0.8\n" for temperature, "5000\n" for max tokens
	w := newTestWizard("y\n0.8\n5000\n")
	err := w.offerCustomization()

	if err != nil {
		t.Fatalf("offerCustomization() error = %v", err)
	}

	if temp, ok := w.customSettings["temperature"]; ok {
		if temp.(float64) != 0.8 {
			t.Errorf("offerCustomization() temperature = %v, want 0.8", temp)
		}
	} else {
		t.Error("offerCustomization() temperature not set")
	}

	if tokens, ok := w.customSettings["max_tokens"]; ok {
		if tokens.(int) != 5000 {
			t.Errorf("offerCustomization() max_tokens = %v, want 5000", tokens)
		}
	} else {
		t.Error("offerCustomization() max_tokens not set")
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
		customSettings map[string]any
		expectedModel  string
		expectedStream bool
	}{
		{
			name:           "research template with overrides",
			useCase:        config.TemplateResearch,
			selectedModel:  "sonar-pro",
			enableStream:   true,
			apiKey:         "test-key",
			customSettings: map[string]any{"temperature": 0.5},
			expectedModel:  "sonar-pro",
			expectedStream: true,
		},
		{
			name:           "custom config from scratch",
			useCase:        "custom",
			selectedModel:  "sonar",
			enableStream:   false,
			apiKey:         "",
			customSettings: map[string]any{},
			expectedModel:  "sonar",
			expectedStream: false,
		},
		{
			name:           "general config",
			useCase:        "general",
			selectedModel:  "sonar-reasoning",
			enableStream:   true,
			apiKey:         "",
			customSettings: map[string]any{},
			expectedModel:  "sonar-reasoning",
			expectedStream: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := newTestWizard("")
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
		name            string
		searchFilters   []string
		expectedMode    string
		expectedRecency string
		expectedContext string
	}{
		{
			name:            "academic mode filter",
			searchFilters:   []string{"mode:academic"},
			expectedMode:    "academic",
			expectedRecency: "",
			expectedContext: "",
		},
		{
			name:            "recency filter",
			searchFilters:   []string{"recency:week"},
			expectedMode:    "",
			expectedRecency: "week",
			expectedContext: "",
		},
		{
			name:            "context size filter",
			searchFilters:   []string{"context:high"},
			expectedMode:    "",
			expectedRecency: "",
			expectedContext: "high",
		},
		{
			name:            "multiple filters",
			searchFilters:   []string{"mode:academic", "recency:month", "context:medium"},
			expectedMode:    "academic",
			expectedRecency: "month",
			expectedContext: "medium",
		},
		{
			name:            "no filters",
			searchFilters:   []string{},
			expectedMode:    "",
			expectedRecency: "",
			expectedContext: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			w := newTestWizard("")
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

			result := config.IsValidSearchRecency(tt.value)
			if result != tt.expected {
				t.Errorf("config.IsValidSearchRecency(%q) = %v, want %v", tt.value, result, tt.expected)
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

			result := config.IsValidContextSize(tt.value)
			if result != tt.expected {
				t.Errorf("config.IsValidContextSize(%q) = %v, want %v", tt.value, result, tt.expected)
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

			w := newTestWizard("")
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

			w := newTestWizard("")
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

	if w.input == nil {
		t.Error("NewWizardState() input is nil")
	}

	if w.output == nil {
		t.Error("NewWizardState() output is nil")
	}

	if w.config == nil {
		t.Error("NewWizardState() config is nil")
	}

	if w.customSettings == nil {
		t.Error("NewWizardState() customSettings is nil")
	}
}

// TestTemplateLoadingFallback tests that wizard falls back to default config when template loading fails.
func TestTemplateLoadingFallback(t *testing.T) {
	t.Parallel()

	w := newTestWizard("")
	w.useCase = "invalid-template"
	w.selectedModel = "sonar"
	w.enableStream = true

	w.buildConfiguration()

	if w.config == nil {
		t.Fatal("buildConfiguration() with invalid template returned nil config")
	}

	if w.config.Defaults.Model != "sonar" {
		t.Errorf("buildConfiguration() fallback model = %v, want sonar", w.config.Defaults.Model)
	}

	if w.config.Output.Stream != true {
		t.Error("buildConfiguration() fallback did not apply streaming preference")
	}
}

// TestAdvancedCustomizationSkip tests skipping advanced customization.
func TestAdvancedCustomizationSkip(t *testing.T) {
	t.Parallel()

	// Select "Skip" (option 6)
	w := newTestWizard("6\n")
	err := w.offerAdvancedCustomization()

	if err != nil {
		t.Fatalf("offerAdvancedCustomization() error = %v", err)
	}

	if len(w.customSettings) != 0 {
		t.Errorf("offerAdvancedCustomization() customSettings = %v, want empty", w.customSettings)
	}
}

// TestIsValidDateFormat tests date format validation.
func TestIsValidDateFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{name: "valid date", value: "2024-01-15", expected: true},
		{name: "valid date end of year", value: "2024-12-31", expected: true},
		{name: "invalid format", value: "01-15-2024", expected: false},
		{name: "invalid month", value: "2024-13-01", expected: false},
		{name: "invalid day", value: "2024-01-32", expected: false},
		{name: "too short", value: "2024-01", expected: false},
		{name: "empty string", value: "", expected: false},
		{name: "letters", value: "abcd-ef-gh", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := isValidDateFormat(tt.value)
			if result != tt.expected {
				t.Errorf("isValidDateFormat(%q) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

// TestValidateOptionalFloat tests the float validation helper.
func TestValidateOptionalFloat(t *testing.T) {
	t.Parallel()

	validate := validateOptionalFloat(0, 2)

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "empty is ok", input: "", wantErr: false},
		{name: "valid float", input: "0.5", wantErr: false},
		{name: "min boundary", input: "0", wantErr: false},
		{name: "max boundary", input: "2", wantErr: false},
		{name: "too high", input: "2.1", wantErr: true},
		{name: "too low", input: "-0.1", wantErr: true},
		{name: "not a number", input: "abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateOptionalFloat()(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

// TestValidateOptionalInt tests the int validation helper.
func TestValidateOptionalInt(t *testing.T) {
	t.Parallel()

	validate := validateOptionalInt(1, 100000)

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "empty is ok", input: "", wantErr: false},
		{name: "valid int", input: "4000", wantErr: false},
		{name: "min boundary", input: "1", wantErr: false},
		{name: "max boundary", input: "100000", wantErr: false},
		{name: "too high", input: "100001", wantErr: true},
		{name: "too low", input: "0", wantErr: true},
		{name: "not a number", input: "abc", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateOptionalInt()(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}
