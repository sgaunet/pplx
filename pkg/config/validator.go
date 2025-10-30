package config

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	maxTemperature = 2.0
	maxPenalty     = 2.0
)

// ValidationError represents a configuration validation error.
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

// Error implements the error interface.
func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	messages := make([]string, 0, len(e))
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// Validator validates configuration data.
type Validator struct {
	errors ValidationErrors
}

// NewValidator creates a new validator.
func NewValidator() *Validator {
	return &Validator{
		errors: make(ValidationErrors, 0),
	}
}

// Validate validates the configuration data.
func (v *Validator) Validate(data *ConfigData) error {
	v.errors = make(ValidationErrors, 0)

	// Validate defaults
	v.validateDefaults(&data.Defaults)

	// Validate search config
	v.validateSearch(&data.Search)

	// Validate output config
	v.validateOutput(&data.Output)

	// Validate API config
	v.validateAPI(&data.API)

	// Validate profiles
	v.validateProfiles(data.Profiles)

	// Validate active profile exists
	if data.ActiveProfile != "" {
		if _, ok := data.Profiles[data.ActiveProfile]; !ok && data.ActiveProfile != DefaultProfileName {
			v.addError("active_profile", fmt.Sprintf("profile '%s' does not exist", data.ActiveProfile))
		}
	}

	if len(v.errors) > 0 {
		return v.errors
	}

	return nil
}

// Errors returns all validation errors.
func (v *Validator) Errors() ValidationErrors {
	return v.errors
}

// validateRange checks if a value is within a range (0 to maxVal) and adds an error if not.
func (v *Validator) validateRange(field string, value, maxVal float64) {
	if value < 0 || value > maxVal {
		v.addError(field, fmt.Sprintf("must be between 0.0 and %.1f", maxVal))
	}
}

// validatePositive checks if an integer value is positive and adds an error if not.
func (v *Validator) validatePositive(field string, value int) {
	if value < 0 {
		v.addError(field, "must be a positive integer")
	}
}

// validateDefaults validates default configuration values.
func (v *Validator) validateDefaults(defaults *DefaultsConfig) {
	v.validateRange("defaults.temperature", defaults.Temperature, maxTemperature)
	v.validatePositive("defaults.max_tokens", defaults.MaxTokens)
	v.validatePositive("defaults.top_k", defaults.TopK)
	v.validateRange("defaults.top_p", defaults.TopP, 1.0)
	v.validateRange("defaults.frequency_penalty", defaults.FrequencyPenalty, maxPenalty)
	v.validateRange("defaults.presence_penalty", defaults.PresencePenalty, maxPenalty)
}

// validateSearch validates search configuration.
func (v *Validator) validateSearch(search *SearchConfig) {
	v.validateSearchRecency(search.Recency)
	v.validateSearchMode(search.Mode)
	v.validateSearchContextSize(search.ContextSize)
	v.validateCoordinates(search.LocationLat, search.LocationLon)
	v.validateSearchDates(search)
}

// validateSearchRecency validates the recency field.
func (v *Validator) validateSearchRecency(recency string) {
	if recency != "" {
		validRecency := map[string]bool{
			"hour": true, "day": true, "week": true, "month": true, "year": true,
		}
		if !validRecency[recency] {
			v.addError("search.recency", "must be one of: hour, day, week, month, year")
		}
	}
}

// validateSearchMode validates the mode field.
func (v *Validator) validateSearchMode(mode string) {
	if mode != "" && mode != "web" && mode != "academic" {
		v.addError("search.mode", "must be 'web' or 'academic'")
	}
}

// validateSearchContextSize validates the context_size field.
func (v *Validator) validateSearchContextSize(contextSize string) {
	if contextSize != "" {
		if contextSize != "low" && contextSize != "medium" && contextSize != "high" {
			v.addError("search.context_size", "must be 'low', 'medium', or 'high'")
		}
	}
}

// validateCoordinates validates location coordinates.
func (v *Validator) validateCoordinates(lat, lon float64) {
	if lat < -90 || lat > 90 {
		v.addError("search.location_lat", "must be between -90 and 90")
	}
	if lon < -180 || lon > 180 {
		v.addError("search.location_lon", "must be between -180 and 180")
	}
}

// validateSearchDates validates date format fields.
func (v *Validator) validateSearchDates(search *SearchConfig) {
	datePattern := regexp.MustCompile(`^\d{2}/\d{2}/\d{4}$`)

	if search.AfterDate != "" && !datePattern.MatchString(search.AfterDate) {
		v.addError("search.after_date", "must be in MM/DD/YYYY format")
	}
	if search.BeforeDate != "" && !datePattern.MatchString(search.BeforeDate) {
		v.addError("search.before_date", "must be in MM/DD/YYYY format")
	}
	if search.LastUpdatedAfter != "" && !datePattern.MatchString(search.LastUpdatedAfter) {
		v.addError("search.last_updated_after", "must be in MM/DD/YYYY format")
	}
	if search.LastUpdatedBefore != "" && !datePattern.MatchString(search.LastUpdatedBefore) {
		v.addError("search.last_updated_before", "must be in MM/DD/YYYY format")
	}
}

// validateOutput validates output configuration.
func (v *Validator) validateOutput(output *OutputConfig) {
	// Validate reasoning effort
	if output.ReasoningEffort != "" {
		if output.ReasoningEffort != "low" && output.ReasoningEffort != "medium" && output.ReasoningEffort != "high" {
			v.addError("output.reasoning_effort", "must be 'low', 'medium', or 'high'")
		}
	}
}

// validateAPI validates API configuration.
func (v *Validator) validateAPI(api *APIConfig) {
	// Validate base URL format if provided
	if api.BaseURL != "" {
		if !strings.HasPrefix(api.BaseURL, "http://") && !strings.HasPrefix(api.BaseURL, "https://") {
			v.addError("api.base_url", "must start with http:// or https://")
		}
	}
}

// validateProfiles validates all profiles.
func (v *Validator) validateProfiles(profiles map[string]*Profile) {
	profileNamePattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	for name, profile := range profiles {
		// Validate profile name
		if !profileNamePattern.MatchString(name) {
			v.addError("profiles."+name,
				"profile name must contain only alphanumeric characters, hyphens, and underscores")
		}

		// Validate profile fields
		if profile.Name != name {
			v.addError(fmt.Sprintf("profiles.%s.name", name),
				fmt.Sprintf("profile name mismatch: key is '%s' but name field is '%s'", name, profile.Name))
		}

		// Validate nested configurations
		v.validateDefaults(&profile.Defaults)
		v.validateSearch(&profile.Search)
		v.validateOutput(&profile.Output)
	}
}

// addError adds a validation error.
func (v *Validator) addError(field, message string) {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Message: message,
	})
}
