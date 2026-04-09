package config

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/sgaunet/pplx/pkg/clerrors"
)

const (
	maxTemperature = 2.0
	maxPenalty     = 2.0
	// enumSuggestMaxDistance is the maximum edit distance for enum suggestions.
	enumSuggestMaxDistance = 2
)

// Validator validates configuration data.
type Validator struct {
	errors clerrors.ValidationErrors
}

// NewValidator creates a new validator.
func NewValidator() *Validator {
	return &Validator{
		errors: make(clerrors.ValidationErrors, 0),
	}
}

// Validate validates the configuration data.
func (v *Validator) Validate(data *ConfigData) error {
	v.errors = make(clerrors.ValidationErrors, 0)

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
func (v *Validator) Errors() clerrors.ValidationErrors {
	return v.errors
}

// validateRange checks if a value is within a range (0 to maxVal) and adds an error if not.
func (v *Validator) validateRange(field string, value, maxVal float64) {
	if value < 0 || value > maxVal {
		v.addError(field, fmt.Sprintf("%g is out of range (must be between 0.0 and %.1f)", value, maxVal))
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
	if recency == "" {
		return
	}
	valid := GetValidSearchRecencyValues()
	if !IsValidSearchRecency(recency) {
		msg := fmt.Sprintf("%q is not valid (must be one of: %s)", recency, strings.Join(valid, ", "))
		if suggestion := SuggestEnum(recency, valid, enumSuggestMaxDistance); suggestion != "" {
			msg += fmt.Sprintf(`. Did you mean %q?`, suggestion)
		}
		v.addError("search.recency", msg)
	}
}

// validateSearchMode validates the mode field.
func (v *Validator) validateSearchMode(mode string) {
	if mode == "" {
		return
	}
	valid := GetValidSearchModeValues()
	if !IsValidSearchMode(mode) {
		msg := fmt.Sprintf("%q is not valid (must be one of: %s)", mode, strings.Join(valid, ", "))
		if suggestion := SuggestEnum(mode, valid, enumSuggestMaxDistance); suggestion != "" {
			msg += fmt.Sprintf(`. Did you mean %q?`, suggestion)
		}
		v.addError("search.mode", msg)
	}
}

// validateSearchContextSize validates the context_size field.
func (v *Validator) validateSearchContextSize(contextSize string) {
	if contextSize == "" {
		return
	}
	valid := GetValidContextSizeValues()
	if !IsValidContextSize(contextSize) {
		msg := fmt.Sprintf("%q is not valid (must be one of: %s)", contextSize, strings.Join(valid, ", "))
		if suggestion := SuggestEnum(contextSize, valid, enumSuggestMaxDistance); suggestion != "" {
			msg += fmt.Sprintf(`. Did you mean %q?`, suggestion)
		}
		v.addError("search.context_size", msg)
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

// isValidDate reports whether s is a valid date in either YYYY-MM-DD (ISO 8601)
// or MM/DD/YYYY format.
func isValidDate(s string) bool {
	if _, err := time.Parse("2006-01-02", s); err == nil {
		return true
	}
	_, err := time.Parse("01/02/2006", s)
	return err == nil
}

// validateSearchDates validates date format fields.
// Accepted formats: YYYY-MM-DD (ISO 8601) and MM/DD/YYYY.
func (v *Validator) validateSearchDates(search *SearchConfig) {
	const dateErrMsg = "must be in YYYY-MM-DD or MM/DD/YYYY format"

	if search.AfterDate != "" && !isValidDate(search.AfterDate) {
		v.addError("search.after_date", dateErrMsg)
	}
	if search.BeforeDate != "" && !isValidDate(search.BeforeDate) {
		v.addError("search.before_date", dateErrMsg)
	}
	if search.LastUpdatedAfter != "" && !isValidDate(search.LastUpdatedAfter) {
		v.addError("search.last_updated_after", dateErrMsg)
	}
	if search.LastUpdatedBefore != "" && !isValidDate(search.LastUpdatedBefore) {
		v.addError("search.last_updated_before", dateErrMsg)
	}
}

// validateOutput validates output configuration.
func (v *Validator) validateOutput(output *OutputConfig) {
	// Validate reasoning effort
	if output.ReasoningEffort == "" {
		return
	}
	valid := GetValidReasoningEffortValues()
	if !IsValidReasoningEffort(output.ReasoningEffort) {
		msg := fmt.Sprintf("%q is not valid (must be one of: %s)", output.ReasoningEffort, strings.Join(valid, ", "))
		if suggestion := SuggestEnum(output.ReasoningEffort, valid, enumSuggestMaxDistance); suggestion != "" {
			msg += fmt.Sprintf(`. Did you mean %q?`, suggestion)
		}
		v.addError("output.reasoning_effort", msg)
	}
}

// validateAPI validates API configuration.
func (v *Validator) validateAPI(api *APIConfig) {
	// Validate base URL format if provided
	if api.BaseURL != "" {
		u, err := url.Parse(api.BaseURL)
		if err != nil {
			v.addError("api.base_url", fmt.Sprintf("invalid URL format: %v", err))
			return
		}
		if u.Scheme != "http" && u.Scheme != "https" {
			v.addError("api.base_url", "must use http or https scheme")
			return
		}
		if u.Host == "" {
			v.addError("api.base_url", "must specify a host")
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

		// Profile field values use pointer-based types (ProfileDefaults, ProfileSearch,
		// ProfileOutput) which are incompatible with the concrete config validators.
		// Field-level validation happens after the profile is merged into ConfigData
		// via MergeProfile, so we skip redundant per-profile field validation here.
	}
}

// addError adds a validation error.
func (v *Validator) addError(field, message string) {
	v.errors = append(v.errors, clerrors.ValidationError{
		Field:   field,
		Value:   "", // Config validation doesn't use Value field
		Message: message,
	})
}
