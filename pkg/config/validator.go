package config

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// Validator validates configuration data
type Validator struct {
	errors ValidationErrors
}

// NewValidator creates a new validator
func NewValidator() *Validator {
	return &Validator{
		errors: make(ValidationErrors, 0),
	}
}

// Validate validates the configuration data
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
		if _, ok := data.Profiles[data.ActiveProfile]; !ok && data.ActiveProfile != "default" {
			v.addError("active_profile", fmt.Sprintf("profile '%s' does not exist", data.ActiveProfile))
		}
	}

	if len(v.errors) > 0 {
		return v.errors
	}

	return nil
}

// validateDefaults validates default configuration values
func (v *Validator) validateDefaults(defaults *DefaultsConfig) {
	// Validate temperature range
	if defaults.Temperature < 0 || defaults.Temperature > 2.0 {
		v.addError("defaults.temperature", "must be between 0 and 2.0")
	}

	// Validate max_tokens
	if defaults.MaxTokens < 0 {
		v.addError("defaults.max_tokens", "must be a positive integer")
	}

	// Validate top_k
	if defaults.TopK < 0 {
		v.addError("defaults.top_k", "must be a positive integer")
	}

	// Validate top_p range
	if defaults.TopP < 0 || defaults.TopP > 1.0 {
		v.addError("defaults.top_p", "must be between 0 and 1.0")
	}

	// Validate frequency_penalty range
	if defaults.FrequencyPenalty < 0 || defaults.FrequencyPenalty > 2.0 {
		v.addError("defaults.frequency_penalty", "must be between 0 and 2.0")
	}

	// Validate presence_penalty range
	if defaults.PresencePenalty < 0 || defaults.PresencePenalty > 2.0 {
		v.addError("defaults.presence_penalty", "must be between 0 and 2.0")
	}
}

// validateSearch validates search configuration
func (v *Validator) validateSearch(search *SearchConfig) {
	// Validate recency
	if search.Recency != "" {
		validRecency := map[string]bool{
			"hour": true, "day": true, "week": true, "month": true, "year": true,
		}
		if !validRecency[search.Recency] {
			v.addError("search.recency", "must be one of: hour, day, week, month, year")
		}
	}

	// Validate mode
	if search.Mode != "" {
		if search.Mode != "web" && search.Mode != "academic" {
			v.addError("search.mode", "must be 'web' or 'academic'")
		}
	}

	// Validate context size
	if search.ContextSize != "" {
		if search.ContextSize != "low" && search.ContextSize != "medium" && search.ContextSize != "high" {
			v.addError("search.context_size", "must be 'low', 'medium', or 'high'")
		}
	}

	// Validate location coordinates
	if search.LocationLat < -90 || search.LocationLat > 90 {
		v.addError("search.location_lat", "must be between -90 and 90")
	}

	if search.LocationLon < -180 || search.LocationLon > 180 {
		v.addError("search.location_lon", "must be between -180 and 180")
	}

	// Validate date formats (MM/DD/YYYY)
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

// validateOutput validates output configuration
func (v *Validator) validateOutput(output *OutputConfig) {
	// Validate reasoning effort
	if output.ReasoningEffort != "" {
		if output.ReasoningEffort != "low" && output.ReasoningEffort != "medium" && output.ReasoningEffort != "high" {
			v.addError("output.reasoning_effort", "must be 'low', 'medium', or 'high'")
		}
	}
}

// validateAPI validates API configuration
func (v *Validator) validateAPI(api *APIConfig) {
	// Validate base URL format if provided
	if api.BaseURL != "" {
		if !strings.HasPrefix(api.BaseURL, "http://") && !strings.HasPrefix(api.BaseURL, "https://") {
			v.addError("api.base_url", "must start with http:// or https://")
		}
	}
}

// validateProfiles validates all profiles
func (v *Validator) validateProfiles(profiles map[string]*Profile) {
	profileNamePattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

	for name, profile := range profiles {
		// Validate profile name
		if !profileNamePattern.MatchString(name) {
			v.addError(fmt.Sprintf("profiles.%s", name),
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

// addError adds a validation error
func (v *Validator) addError(field, message string) {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// Errors returns all validation errors
func (v *Validator) Errors() ValidationErrors {
	return v.errors
}
