package cmd

import (
	"testing"
	"time"
)

func TestParseDateFilter(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		dateStr   string
		wantErr   bool
	}{
		{
			name:      "valid date",
			fieldName: "test-date",
			dateStr:   "01/15/2024",
			wantErr:   false,
		},
		{
			name:      "invalid format - ISO 8601",
			fieldName: "test-date",
			dateStr:   "2024-01-15",
			wantErr:   true,
		},
		{
			name:      "invalid format - European",
			fieldName:      "test-date",
			dateStr:   "15/01/2024",
			wantErr:   true,
		},
		{
			name:      "invalid date - month out of range",
			fieldName: "test-date",
			dateStr:   "13/32/2024",
			wantErr:   true,
		},
		{
			name:      "empty string",
			fieldName: "test-date",
			dateStr:   "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDateFilter(tt.fieldName, tt.dateStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseDateFilter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Verify we can parse it back
				if got.IsZero() {
					t.Errorf("parseDateFilter() returned zero time for valid input")
				}
				// Verify the date components match
				expected, _ := time.Parse("01/02/2006", tt.dateStr)
				if !got.Equal(expected) {
					t.Errorf("parseDateFilter() = %v, want %v", got, expected)
				}
			}
		})
	}
}

func TestValidateStringEnum(t *testing.T) {
	validValues := map[string]bool{"foo": true, "bar": true, "baz": true}

	tests := []struct {
		name      string
		fieldName string
		value     string
		wantErr   bool
	}{
		{
			name:      "valid value - foo",
			fieldName: "test-field",
			value:     "foo",
			wantErr:   false,
		},
		{
			name:      "valid value - bar",
			fieldName: "test-field",
			value:     "bar",
			wantErr:   false,
		},
		{
			name:      "valid value - baz",
			fieldName: "test-field",
			value:     "baz",
			wantErr:   false,
		},
		{
			name:      "empty value - should be valid",
			fieldName: "test-field",
			value:     "",
			wantErr:   false,
		},
		{
			name:      "invalid value",
			fieldName: "test-field",
			value:     "invalid",
			wantErr:   true,
		},
		{
			name:      "case sensitive - FOO should fail",
			fieldName: "test-field",
			value:     "FOO",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStringEnum(tt.fieldName, tt.value, validValues, "foo, bar, baz")
			if (err != nil) != tt.wantErr {
				t.Errorf("validateStringEnum() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuildSearchOptions(t *testing.T) {
	// Save original values
	origSearchDomains := searchDomains
	origSearchRecency := searchRecency
	origReturnImages := returnImages
	origLocationLat := locationLat
	origLocationLon := locationLon
	origLocationCountry := locationCountry
	origSearchMode := searchMode
	origSearchContextSize := searchContextSize

	// Restore original values after test
	defer func() {
		searchDomains = origSearchDomains
		searchRecency = origSearchRecency
		returnImages = origReturnImages
		locationLat = origLocationLat
		locationLon = origLocationLon
		locationCountry = origLocationCountry
		searchMode = origSearchMode
		searchContextSize = origSearchContextSize
	}()

	t.Run("empty options", func(t *testing.T) {
		// Reset all search-related globals
		searchDomains = []string{}
		searchRecency = ""
		returnImages = false
		locationLat = 0
		locationLon = 0
		locationCountry = ""
		searchMode = ""
		searchContextSize = ""

		opts := buildSearchOptions()
		if len(opts) != 0 {
			t.Errorf("buildSearchOptions() returned %d options, want 0", len(opts))
		}
	})

	t.Run("with search domains", func(t *testing.T) {
		searchDomains = []string{"example.com", "test.com"}
		searchRecency = ""
		returnImages = false
		locationLat = 0
		locationLon = 0
		locationCountry = ""
		searchMode = ""
		searchContextSize = ""

		opts := buildSearchOptions()
		if len(opts) != 1 {
			t.Errorf("buildSearchOptions() returned %d options, want 1", len(opts))
		}
	})

	t.Run("with search recency and no images", func(t *testing.T) {
		searchDomains = []string{}
		searchRecency = "week"
		returnImages = false
		locationLat = 0
		locationLon = 0
		locationCountry = ""
		searchMode = ""
		searchContextSize = ""

		opts := buildSearchOptions()
		// Should have 1 option (search recency)
		if len(opts) != 1 {
			t.Errorf("buildSearchOptions() returned %d options, want 1", len(opts))
		}
	})

	t.Run("with search recency and images - recency should be skipped", func(t *testing.T) {
		searchDomains = []string{}
		searchRecency = "week"
		returnImages = true
		locationLat = 0
		locationLon = 0
		locationCountry = ""
		searchMode = ""
		searchContextSize = ""

		opts := buildSearchOptions()
		// Should have 0 options (recency skipped due to images)
		if len(opts) != 0 {
			t.Errorf("buildSearchOptions() returned %d options, want 0 (recency should be skipped with images)", len(opts))
		}
	})

	t.Run("with location", func(t *testing.T) {
		searchDomains = []string{}
		searchRecency = ""
		returnImages = false
		locationLat = 40.7128
		locationLon = -74.0060
		locationCountry = "US"
		searchMode = ""
		searchContextSize = ""

		opts := buildSearchOptions()
		if len(opts) != 1 {
			t.Errorf("buildSearchOptions() returned %d options, want 1", len(opts))
		}
	})

	t.Run("with search mode and context size", func(t *testing.T) {
		searchDomains = []string{}
		searchRecency = ""
		returnImages = false
		locationLat = 0
		locationLon = 0
		locationCountry = ""
		searchMode = "web"
		searchContextSize = "medium"

		opts := buildSearchOptions()
		// Should have 2 options (mode + context size)
		if len(opts) != 2 {
			t.Errorf("buildSearchOptions() returned %d options, want 2", len(opts))
		}
	})

	t.Run("with all options", func(t *testing.T) {
		searchDomains = []string{"example.com"}
		searchRecency = "week"
		returnImages = false
		locationLat = 40.7128
		locationLon = -74.0060
		locationCountry = "US"
		searchMode = "web"
		searchContextSize = "high"

		opts := buildSearchOptions()
		// Should have 5 options (domains, recency, location, mode, context)
		if len(opts) != 5 {
			t.Errorf("buildSearchOptions() returned %d options, want 5", len(opts))
		}
	})
}

func TestBuildDateFilterOptions(t *testing.T) {
	// Save original values
	origSearchAfterDate := searchAfterDate
	origSearchBeforeDate := searchBeforeDate
	origLastUpdatedAfter := lastUpdatedAfter
	origLastUpdatedBefore := lastUpdatedBefore

	// Restore original values after test
	defer func() {
		searchAfterDate = origSearchAfterDate
		searchBeforeDate = origSearchBeforeDate
		lastUpdatedAfter = origLastUpdatedAfter
		lastUpdatedBefore = origLastUpdatedBefore
	}()

	t.Run("no date filters", func(t *testing.T) {
		searchAfterDate = ""
		searchBeforeDate = ""
		lastUpdatedAfter = ""
		lastUpdatedBefore = ""

		opts, err := buildDateFilterOptions()
		if err != nil {
			t.Errorf("buildDateFilterOptions() unexpected error = %v", err)
		}
		if len(opts) != 0 {
			t.Errorf("buildDateFilterOptions() returned %d options, want 0", len(opts))
		}
	})

	t.Run("with valid search after date", func(t *testing.T) {
		searchAfterDate = "01/01/2024"
		searchBeforeDate = ""
		lastUpdatedAfter = ""
		lastUpdatedBefore = ""

		opts, err := buildDateFilterOptions()
		if err != nil {
			t.Errorf("buildDateFilterOptions() unexpected error = %v", err)
		}
		if len(opts) != 1 {
			t.Errorf("buildDateFilterOptions() returned %d options, want 1", len(opts))
		}
	})

	t.Run("with invalid date format", func(t *testing.T) {
		searchAfterDate = "2024-01-01"
		searchBeforeDate = ""
		lastUpdatedAfter = ""
		lastUpdatedBefore = ""

		_, err := buildDateFilterOptions()
		if err == nil {
			t.Errorf("buildDateFilterOptions() expected error for invalid date format, got nil")
		}
	})

	t.Run("with all date filters", func(t *testing.T) {
		searchAfterDate = "01/01/2024"
		searchBeforeDate = "12/31/2024"
		lastUpdatedAfter = "06/01/2024"
		lastUpdatedBefore = "06/30/2024"

		opts, err := buildDateFilterOptions()
		if err != nil {
			t.Errorf("buildDateFilterOptions() unexpected error = %v", err)
		}
		if len(opts) != 4 {
			t.Errorf("buildDateFilterOptions() returned %d options, want 4", len(opts))
		}
	})
}

func TestValidateInputs(t *testing.T) {
	// Save original values
	origUserPrompt := userPrompt
	origSearchRecency := searchRecency
	origSearchMode := searchMode
	origSearchContextSize := searchContextSize
	origReasoningEffort := reasoningEffort
	origResponseFormatJSONSchema := responseFormatJSONSchema
	origResponseFormatRegex := responseFormatRegex
	origModel := model

	// Restore original values after test
	defer func() {
		userPrompt = origUserPrompt
		searchRecency = origSearchRecency
		searchMode = origSearchMode
		searchContextSize = origSearchContextSize
		reasoningEffort = origReasoningEffort
		responseFormatJSONSchema = origResponseFormatJSONSchema
		responseFormatRegex = origResponseFormatRegex
		model = origModel
	}()

	t.Run("missing user prompt", func(t *testing.T) {
		userPrompt = ""
		searchRecency = ""
		searchMode = ""
		searchContextSize = ""
		reasoningEffort = ""
		responseFormatJSONSchema = ""
		responseFormatRegex = ""

		err := validateInputs()
		if err == nil {
			t.Errorf("validateInputs() expected error for missing user prompt, got nil")
		}
	})

	t.Run("valid inputs", func(t *testing.T) {
		userPrompt = "test query"
		searchRecency = "week"
		searchMode = "web"
		searchContextSize = "medium"
		reasoningEffort = "high"
		responseFormatJSONSchema = ""
		responseFormatRegex = ""
		model = "sonar-pro"

		err := validateInputs()
		if err != nil {
			t.Errorf("validateInputs() unexpected error = %v", err)
		}
	})

	t.Run("invalid search recency", func(t *testing.T) {
		userPrompt = "test query"
		searchRecency = "invalid"
		searchMode = ""
		searchContextSize = ""
		reasoningEffort = ""
		responseFormatJSONSchema = ""
		responseFormatRegex = ""

		err := validateInputs()
		if err == nil {
			t.Errorf("validateInputs() expected error for invalid search recency, got nil")
		}
	})

	t.Run("invalid search mode", func(t *testing.T) {
		userPrompt = "test query"
		searchRecency = ""
		searchMode = "invalid"
		searchContextSize = ""
		reasoningEffort = ""
		responseFormatJSONSchema = ""
		responseFormatRegex = ""

		err := validateInputs()
		if err == nil {
			t.Errorf("validateInputs() expected error for invalid search mode, got nil")
		}
	})

	t.Run("both JSON schema and regex", func(t *testing.T) {
		userPrompt = "test query"
		searchRecency = ""
		searchMode = ""
		searchContextSize = ""
		reasoningEffort = ""
		responseFormatJSONSchema = "{}"
		responseFormatRegex = "test"
		model = "sonar-pro"

		err := validateInputs()
		if err == nil {
			t.Errorf("validateInputs() expected error for both JSON schema and regex, got nil")
		}
	})

	t.Run("response format with non-sonar model", func(t *testing.T) {
		userPrompt = "test query"
		searchRecency = ""
		searchMode = ""
		searchContextSize = ""
		reasoningEffort = ""
		responseFormatJSONSchema = "{}"
		responseFormatRegex = ""
		model = "llama-3.1-70b"

		err := validateInputs()
		if err == nil {
			t.Errorf("validateInputs() expected error for response format with non-sonar model, got nil")
		}
	})
}
