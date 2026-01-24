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
	origSearchDomains := globalOpts.SearchDomains
	origSearchRecency := globalOpts.SearchRecency
	origReturnImages := globalOpts.ReturnImages
	origLocationLat := globalOpts.LocationLat
	origLocationLon := globalOpts.LocationLon
	origLocationCountry := globalOpts.LocationCountry
	origSearchMode := globalOpts.SearchMode
	origSearchContextSize := globalOpts.SearchContextSize

	// Restore original values after test
	defer func() {
		globalOpts.SearchDomains = origSearchDomains
		globalOpts.SearchRecency = origSearchRecency
		globalOpts.ReturnImages = origReturnImages
		globalOpts.LocationLat = origLocationLat
		globalOpts.LocationLon = origLocationLon
		globalOpts.LocationCountry = origLocationCountry
		globalOpts.SearchMode = origSearchMode
		globalOpts.SearchContextSize = origSearchContextSize
	}()

	t.Run("empty options", func(t *testing.T) {
		// Reset all search-related globals
		globalOpts.SearchDomains = []string{}
		globalOpts.SearchRecency = ""
		globalOpts.ReturnImages = false
		globalOpts.LocationLat = 0
		globalOpts.LocationLon = 0
		globalOpts.LocationCountry = ""
		globalOpts.SearchMode = ""
		globalOpts.SearchContextSize = ""

		opts := buildSearchOptions()
		if len(opts) != 0 {
			t.Errorf("buildSearchOptions() returned %d options, want 0", len(opts))
		}
	})

	t.Run("with search domains", func(t *testing.T) {
		globalOpts.SearchDomains = []string{"example.com", "test.com"}
		globalOpts.SearchRecency = ""
		globalOpts.ReturnImages = false
		globalOpts.LocationLat = 0
		globalOpts.LocationLon = 0
		globalOpts.LocationCountry = ""
		globalOpts.SearchMode = ""
		globalOpts.SearchContextSize = ""

		opts := buildSearchOptions()
		if len(opts) != 1 {
			t.Errorf("buildSearchOptions() returned %d options, want 1", len(opts))
		}
	})

	t.Run("with search recency and no images", func(t *testing.T) {
		globalOpts.SearchDomains = []string{}
		globalOpts.SearchRecency = "week"
		globalOpts.ReturnImages = false
		globalOpts.LocationLat = 0
		globalOpts.LocationLon = 0
		globalOpts.LocationCountry = ""
		globalOpts.SearchMode = ""
		globalOpts.SearchContextSize = ""

		opts := buildSearchOptions()
		// Should have 1 option (search recency)
		if len(opts) != 1 {
			t.Errorf("buildSearchOptions() returned %d options, want 1", len(opts))
		}
	})

	t.Run("with search recency and images - recency should be skipped", func(t *testing.T) {
		globalOpts.SearchDomains = []string{}
		globalOpts.SearchRecency = "week"
		globalOpts.ReturnImages = true
		globalOpts.LocationLat = 0
		globalOpts.LocationLon = 0
		globalOpts.LocationCountry = ""
		globalOpts.SearchMode = ""
		globalOpts.SearchContextSize = ""

		opts := buildSearchOptions()
		// Should have 0 options (recency skipped due to images)
		if len(opts) != 0 {
			t.Errorf("buildSearchOptions() returned %d options, want 0 (recency should be skipped with images)", len(opts))
		}
	})

	t.Run("with location", func(t *testing.T) {
		globalOpts.SearchDomains = []string{}
		globalOpts.SearchRecency = ""
		globalOpts.ReturnImages = false
		globalOpts.LocationLat = 40.7128
		globalOpts.LocationLon = -74.0060
		globalOpts.LocationCountry = "US"
		globalOpts.SearchMode = ""
		globalOpts.SearchContextSize = ""

		opts := buildSearchOptions()
		if len(opts) != 1 {
			t.Errorf("buildSearchOptions() returned %d options, want 1", len(opts))
		}
	})

	t.Run("with search mode and context size", func(t *testing.T) {
		globalOpts.SearchDomains = []string{}
		globalOpts.SearchRecency = ""
		globalOpts.ReturnImages = false
		globalOpts.LocationLat = 0
		globalOpts.LocationLon = 0
		globalOpts.LocationCountry = ""
		globalOpts.SearchMode = "web"
		globalOpts.SearchContextSize = "medium"

		opts := buildSearchOptions()
		// Should have 2 options (mode + context size)
		if len(opts) != 2 {
			t.Errorf("buildSearchOptions() returned %d options, want 2", len(opts))
		}
	})

	t.Run("with all options", func(t *testing.T) {
		globalOpts.SearchDomains = []string{"example.com"}
		globalOpts.SearchRecency = "week"
		globalOpts.ReturnImages = false
		globalOpts.LocationLat = 40.7128
		globalOpts.LocationLon = -74.0060
		globalOpts.LocationCountry = "US"
		globalOpts.SearchMode = "web"
		globalOpts.SearchContextSize = "high"

		opts := buildSearchOptions()
		// Should have 5 options (domains, recency, location, mode, context)
		if len(opts) != 5 {
			t.Errorf("buildSearchOptions() returned %d options, want 5", len(opts))
		}
	})
}

func TestBuildDateFilterOptions(t *testing.T) {
	// Save original values
	origSearchAfterDate := globalOpts.SearchAfterDate
	origSearchBeforeDate := globalOpts.SearchBeforeDate
	origLastUpdatedAfter := globalOpts.LastUpdatedAfter
	origLastUpdatedBefore := globalOpts.LastUpdatedBefore

	// Restore original values after test
	defer func() {
		globalOpts.SearchAfterDate = origSearchAfterDate
		globalOpts.SearchBeforeDate = origSearchBeforeDate
		globalOpts.LastUpdatedAfter = origLastUpdatedAfter
		globalOpts.LastUpdatedBefore = origLastUpdatedBefore
	}()

	t.Run("no date filters", func(t *testing.T) {
		globalOpts.SearchAfterDate = ""
		globalOpts.SearchBeforeDate = ""
		globalOpts.LastUpdatedAfter = ""
		globalOpts.LastUpdatedBefore = ""

		opts, err := buildDateFilterOptions()
		if err != nil {
			t.Errorf("buildDateFilterOptions() unexpected error = %v", err)
		}
		if len(opts) != 0 {
			t.Errorf("buildDateFilterOptions() returned %d options, want 0", len(opts))
		}
	})

	t.Run("with valid search after date", func(t *testing.T) {
		globalOpts.SearchAfterDate = "01/01/2024"
		globalOpts.SearchBeforeDate = ""
		globalOpts.LastUpdatedAfter = ""
		globalOpts.LastUpdatedBefore = ""

		opts, err := buildDateFilterOptions()
		if err != nil {
			t.Errorf("buildDateFilterOptions() unexpected error = %v", err)
		}
		if len(opts) != 1 {
			t.Errorf("buildDateFilterOptions() returned %d options, want 1", len(opts))
		}
	})

	t.Run("with invalid date format", func(t *testing.T) {
		globalOpts.SearchAfterDate = "2024-01-01"
		globalOpts.SearchBeforeDate = ""
		globalOpts.LastUpdatedAfter = ""
		globalOpts.LastUpdatedBefore = ""

		_, err := buildDateFilterOptions()
		if err == nil {
			t.Errorf("buildDateFilterOptions() expected error for invalid date format, got nil")
		}
	})

	t.Run("with all date filters", func(t *testing.T) {
		globalOpts.SearchAfterDate = "01/01/2024"
		globalOpts.SearchBeforeDate = "12/31/2024"
		globalOpts.LastUpdatedAfter = "06/01/2024"
		globalOpts.LastUpdatedBefore = "06/30/2024"

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
	origUserPrompt := globalOpts.UserPrompt
	origSearchRecency := globalOpts.SearchRecency
	origSearchMode := globalOpts.SearchMode
	origSearchContextSize := globalOpts.SearchContextSize
	origReasoningEffort := globalOpts.ReasoningEffort
	origResponseFormatJSONSchema := globalOpts.ResponseFormatJSONSchema
	origResponseFormatRegex := globalOpts.ResponseFormatRegex
	origModel := globalOpts.Model

	// Restore original values after test
	defer func() {
		globalOpts.UserPrompt = origUserPrompt
		globalOpts.SearchRecency = origSearchRecency
		globalOpts.SearchMode = origSearchMode
		globalOpts.SearchContextSize = origSearchContextSize
		globalOpts.ReasoningEffort = origReasoningEffort
		globalOpts.ResponseFormatJSONSchema = origResponseFormatJSONSchema
		globalOpts.ResponseFormatRegex = origResponseFormatRegex
		globalOpts.Model= origModel
	}()

	t.Run("missing user prompt", func(t *testing.T) {
		globalOpts.UserPrompt = ""
		globalOpts.SearchRecency = ""
		globalOpts.SearchMode = ""
		globalOpts.SearchContextSize = ""
		globalOpts.ReasoningEffort = ""
		globalOpts.ResponseFormatJSONSchema = ""
		globalOpts.ResponseFormatRegex = ""

		err := validateInputs()
		if err == nil {
			t.Errorf("validateInputs() expected error for missing user prompt, got nil")
		}
	})

	t.Run("valid inputs", func(t *testing.T) {
		globalOpts.UserPrompt = "test query"
		globalOpts.SearchRecency = "week"
		globalOpts.SearchMode = "web"
		globalOpts.SearchContextSize = "medium"
		globalOpts.ReasoningEffort = "high"
		globalOpts.ResponseFormatJSONSchema = ""
		globalOpts.ResponseFormatRegex = ""
		globalOpts.Model= "sonar-pro"

		err := validateInputs()
		if err != nil {
			t.Errorf("validateInputs() unexpected error = %v", err)
		}
	})

	t.Run("invalid search recency", func(t *testing.T) {
		globalOpts.UserPrompt = "test query"
		globalOpts.SearchRecency = "invalid"
		globalOpts.SearchMode = ""
		globalOpts.SearchContextSize = ""
		globalOpts.ReasoningEffort = ""
		globalOpts.ResponseFormatJSONSchema = ""
		globalOpts.ResponseFormatRegex = ""

		err := validateInputs()
		if err == nil {
			t.Errorf("validateInputs() expected error for invalid search recency, got nil")
		}
	})

	t.Run("invalid search mode", func(t *testing.T) {
		globalOpts.UserPrompt = "test query"
		globalOpts.SearchRecency = ""
		globalOpts.SearchMode = "invalid"
		globalOpts.SearchContextSize = ""
		globalOpts.ReasoningEffort = ""
		globalOpts.ResponseFormatJSONSchema = ""
		globalOpts.ResponseFormatRegex = ""

		err := validateInputs()
		if err == nil {
			t.Errorf("validateInputs() expected error for invalid search mode, got nil")
		}
	})

	t.Run("both JSON schema and regex", func(t *testing.T) {
		globalOpts.UserPrompt = "test query"
		globalOpts.SearchRecency = ""
		globalOpts.SearchMode = ""
		globalOpts.SearchContextSize = ""
		globalOpts.ReasoningEffort = ""
		globalOpts.ResponseFormatJSONSchema = "{}"
		globalOpts.ResponseFormatRegex = "test"
		globalOpts.Model= "sonar-pro"

		err := validateInputs()
		if err == nil {
			t.Errorf("validateInputs() expected error for both JSON schema and regex, got nil")
		}
	})

	t.Run("response format with non-sonar globalOpts.Model", func(t *testing.T) {
		globalOpts.UserPrompt = "test query"
		globalOpts.SearchRecency = ""
		globalOpts.SearchMode = ""
		globalOpts.SearchContextSize = ""
		globalOpts.ReasoningEffort = ""
		globalOpts.ResponseFormatJSONSchema = "{}"
		globalOpts.ResponseFormatRegex = ""
		globalOpts.Model= "llama-3.1-70b"

		err := validateInputs()
		if err == nil {
			t.Errorf("validateInputs() expected error for response format with non-sonar globalOpts.Model, got nil")
		}
	})
}
