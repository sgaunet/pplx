package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

// Section constants for configuration categorization.
const (
	SectionDefaults = "defaults"
	SectionSearch   = "search"
	SectionOutput   = "output"
	SectionAPI      = "api"

	// Default table formatting constants.
	defaultMaxDescLength = 60
	defaultTabPadding    = 2
	defaultMinTruncate   = 3
)

var (
	// ErrOptionNotFound is returned when an option is not found.
	ErrOptionNotFound = errors.New("option not found")

	// ErrUnsupportedFormat is returned when an unsupported format is requested.
	ErrUnsupportedFormat = errors.New("unsupported format")
)

// OptionMetadata represents metadata for a single configuration option.
type OptionMetadata struct {
	// Section is the configuration section (defaults, search, output, api)
	Section string `json:"section" yaml:"section"`

	// Name is the configuration option name
	Name string `json:"name" yaml:"name"`

	// Type is the data type (string, int, float64, bool, []string, duration)
	Type string `json:"type" yaml:"type"`

	// Description explains what this option does
	Description string `json:"description" yaml:"description"`

	// Default is the default value for this option (may be nil)
	Default any `json:"default,omitempty" yaml:"default,omitempty"`

	// ValidationRules contains human-readable validation constraints
	ValidationRules []string `json:"validation_rules,omitempty" yaml:"validation_rules,omitempty"`

	// Example provides an example value
	Example string `json:"example,omitempty" yaml:"example,omitempty"`

	// EnvVar is the environment variable name (if applicable)
	EnvVar string `json:"env_var,omitempty" yaml:"env_var,omitempty"`

	// Required indicates if this option must be set
	Required bool `json:"required" yaml:"required"`
}

// MetadataRegistry holds all configuration option metadata.
type MetadataRegistry struct {
	options map[string]*OptionMetadata // key is "section.name"
}

// NewMetadataRegistry creates a new metadata registry with all options.
func NewMetadataRegistry() *MetadataRegistry {
	registry := &MetadataRegistry{
		options: make(map[string]*OptionMetadata),
	}
	registry.initialize()
	return registry
}

// initialize populates the registry with all configuration options.
//
//nolint:funcorder,funlen,maintidx // Keep initialization near constructor; metadata registry initialization
func (r *MetadataRegistry) initialize() {
	// Defaults section (8 options)
	r.addOption(&OptionMetadata{
		Section:     SectionDefaults,
		Name:        "model",
		Type:        "string",
		Description: "Model to use for queries",
		Default:     "",
		Example:     "sonar",
		ValidationRules: []string{
			"Valid model IDs: sonar, sonar-pro, sonar-deep-research",
		},
	})

	r.addOption(&OptionMetadata{
		Section:     SectionDefaults,
		Name:        "temperature",
		Type:        "float64",
		Description: "Controls randomness in responses (0.0 = deterministic, 1.0+ = creative)",
		Default:     0.0,
		Example:     "0.7",
		ValidationRules: []string{
			"Must be between 0.0 and 2.0",
		},
	})

	r.addOption(&OptionMetadata{
		Section:     SectionDefaults,
		Name:        "max_tokens",
		Type:        "int",
		Description: "Maximum number of tokens in response",
		Default:     0,
		Example:     "4096",
		ValidationRules: []string{
			"Must be positive integer",
			"Model-dependent maximum (typically 4096-16384)",
		},
	})

	r.addOption(&OptionMetadata{
		Section:     SectionDefaults,
		Name:        "top_k",
		Type:        "int",
		Description: "Top-K sampling: limit to K highest probability tokens",
		Default:     0,
		Example:     "50",
		ValidationRules: []string{
			"Must be between 0 and 100",
			"0 disables top-K sampling",
		},
	})

	r.addOption(&OptionMetadata{
		Section:     SectionDefaults,
		Name:        "top_p",
		Type:        "float64",
		Description: "Top-P (nucleus) sampling: cumulative probability threshold",
		Default:     0.0,
		Example:     "0.9",
		ValidationRules: []string{
			"Must be between 0.0 and 1.0",
		},
	})

	r.addOption(&OptionMetadata{
		Section:     SectionDefaults,
		Name:        "frequency_penalty",
		Type:        "float64",
		Description: "Reduce repetition of frequent tokens",
		Default:     0.0,
		Example:     "0.5",
		ValidationRules: []string{
			"Must be between 0.0 and 2.0",
		},
	})

	r.addOption(&OptionMetadata{
		Section:     SectionDefaults,
		Name:        "presence_penalty",
		Type:        "float64",
		Description: "Encourage discussing new topics",
		Default:     0.0,
		Example:     "0.3",
		ValidationRules: []string{
			"Must be between 0.0 and 2.0",
		},
	})

	r.addOption(&OptionMetadata{
		Section:     SectionDefaults,
		Name:        "timeout",
		Type:        "string",
		Description: "Timeout for API requests",
		Default:     "",
		Example:     "30s",
		ValidationRules: []string{
			"Format: duration string (e.g., '30s', '5m')",
		},
	})

	// Search section (11 options)
	r.addOption(&OptionMetadata{
		Section:     SectionSearch,
		Name:        "domains",
		Type:        "[]string",
		Description: "Limit search to specific domains",
		Default:     nil,
		Example:     "wikipedia.org,github.com",
		ValidationRules: []string{
			"List of domain names",
		},
	})

	r.addOption(&OptionMetadata{
		Section:     SectionSearch,
		Name:        "recency",
		Type:        "string",
		Description: "Time-based filtering of search results",
		Default:     "",
		Example:     "week",
		ValidationRules: []string{
			"Valid values: hour, day, week, month, year",
		},
	})

	r.addOption(&OptionMetadata{
		Section:     SectionSearch,
		Name:        "mode",
		Type:        "string",
		Description: "Search mode",
		Default:     "web",
		Example:     "academic",
		ValidationRules: []string{
			"Valid values: web, academic",
		},
	})

	r.addOption(&OptionMetadata{
		Section:     SectionSearch,
		Name:        "context_size",
		Type:        "string",
		Description: "Amount of context to use from search results",
		Default:     "",
		Example:     "medium",
		ValidationRules: []string{
			"Valid values: low, medium, high",
		},
	})

	r.addOption(&OptionMetadata{
		Section:     SectionSearch,
		Name:        "location_lat",
		Type:        "float64",
		Description: "Geographic location latitude",
		Default:     0.0,
		Example:     "37.7749",
		ValidationRules: []string{
			"Must be between -90.0 and 90.0",
		},
	})

	r.addOption(&OptionMetadata{
		Section:     SectionSearch,
		Name:        "location_lon",
		Type:        "float64",
		Description: "Geographic location longitude",
		Default:     0.0,
		Example:     "-122.4194",
		ValidationRules: []string{
			"Must be between -180.0 and 180.0",
		},
	})

	r.addOption(&OptionMetadata{
		Section:     SectionSearch,
		Name:        "location_country",
		Type:        "string",
		Description: "Country code for location-based results",
		Default:     "",
		Example:     "US",
		ValidationRules: []string{
			"Format: ISO 3166-1 alpha-2 code",
		},
	})

	r.addOption(&OptionMetadata{
		Section:     SectionSearch,
		Name:        "after_date",
		Type:        "string",
		Description: "Filter results published after this date",
		Default:     "",
		Example:     "01/01/2024",
		ValidationRules: []string{
			"Format: MM/DD/YYYY",
		},
	})

	r.addOption(&OptionMetadata{
		Section:     SectionSearch,
		Name:        "before_date",
		Type:        "string",
		Description: "Filter results published before this date",
		Default:     "",
		Example:     "12/31/2024",
		ValidationRules: []string{
			"Format: MM/DD/YYYY",
		},
	})

	r.addOption(&OptionMetadata{
		Section:     SectionSearch,
		Name:        "last_updated_after",
		Type:        "string",
		Description: "Filter results last updated after this date",
		Default:     "",
		Example:     "01/01/2024",
		ValidationRules: []string{
			"Format: MM/DD/YYYY",
		},
	})

	r.addOption(&OptionMetadata{
		Section:     SectionSearch,
		Name:        "last_updated_before",
		Type:        "string",
		Description: "Filter results last updated before this date",
		Default:     "",
		Example:     "12/31/2024",
		ValidationRules: []string{
			"Format: MM/DD/YYYY",
		},
	})

	// Output section (9 options)
	r.addOption(&OptionMetadata{
		Section:     SectionOutput,
		Name:        "stream",
		Type:        "bool",
		Description: "Enable streaming responses (output tokens as generated)",
		Default:     false,
		Example:     "true",
	})

	r.addOption(&OptionMetadata{
		Section:     SectionOutput,
		Name:        "return_images",
		Type:        "bool",
		Description: "Include images in the response",
		Default:     false,
		Example:     "true",
	})

	r.addOption(&OptionMetadata{
		Section:     SectionOutput,
		Name:        "return_related",
		Type:        "bool",
		Description: "Include related questions in the response",
		Default:     false,
		Example:     "true",
	})

	r.addOption(&OptionMetadata{
		Section:     SectionOutput,
		Name:        "json",
		Type:        "bool",
		Description: "Output response as JSON instead of formatted text",
		Default:     false,
		Example:     "true",
	})

	r.addOption(&OptionMetadata{
		Section:     SectionOutput,
		Name:        "image_domains",
		Type:        "[]string",
		Description: "Filter images by domain",
		Default:     nil,
		Example:     "unsplash.com,imgur.com",
	})

	r.addOption(&OptionMetadata{
		Section:     SectionOutput,
		Name:        "image_formats",
		Type:        "[]string",
		Description: "Filter images by format",
		Default:     nil,
		Example:     "jpg,png",
		ValidationRules: []string{
			"Valid values: jpg, png, gif, webp",
		},
	})

	r.addOption(&OptionMetadata{
		Section:     SectionOutput,
		Name:        "response_format_json_schema",
		Type:        "string",
		Description: "JSON schema for structured output (sonar model only)",
		Default:     "",
		Example:     `{"type":"object","properties":{"answer":{"type":"string"}}}`,
	})

	r.addOption(&OptionMetadata{
		Section:     SectionOutput,
		Name:        "response_format_regex",
		Type:        "string",
		Description: "Regex pattern for structured output (sonar model only)",
		Default:     "",
		Example:     `^[A-Z][a-z]+$`,
	})

	r.addOption(&OptionMetadata{
		Section:     SectionOutput,
		Name:        "reasoning_effort",
		Type:        "string",
		Description: "Reasoning effort for sonar-deep-research model",
		Default:     "",
		Example:     "medium",
		ValidationRules: []string{
			"Valid values: low, medium, high",
		},
	})

	// API section (3 options)
	r.addOption(&OptionMetadata{
		Section:     SectionAPI,
		Name:        "key",
		Type:        "string",
		Description: "API key for authentication",
		Default:     "",
		EnvVar:      "PERPLEXITY_API_KEY",
		Required:    true,
		ValidationRules: []string{
			"Required for API access",
		},
	})

	r.addOption(&OptionMetadata{
		Section:     SectionAPI,
		Name:        "base_url",
		Type:        "string",
		Description: "Custom API base URL (if using a proxy or custom endpoint)",
		Default:     "",
		Example:     "https://api.perplexity.ai",
	})

	r.addOption(&OptionMetadata{
		Section:     SectionAPI,
		Name:        "timeout",
		Type:        "duration",
		Description: "API request timeout",
		Default:     nil,
		Example:     "30s",
		ValidationRules: []string{
			"Format: duration (e.g., 30s, 2m)",
		},
	})
}

// addOption adds an option to the registry.
//
//nolint:funcorder // Keep helper near initialization for readability
func (r *MetadataRegistry) addOption(opt *OptionMetadata) {
	key := fmt.Sprintf("%s.%s", opt.Section, opt.Name)
	r.options[key] = opt
}

// GetOption retrieves metadata for a specific option.
// The name can be in the form "section.name" or just "name".
// If only the name is provided, it searches across all sections.
func (r *MetadataRegistry) GetOption(name string) (*OptionMetadata, error) {
	// Try direct lookup first
	if opt, exists := r.options[name]; exists {
		return opt, nil
	}

	// Try searching by name only
	for key, opt := range r.options {
		if strings.HasSuffix(key, "."+name) {
			return opt, nil
		}
	}

	return nil, fmt.Errorf("%w: %s", ErrOptionNotFound, name)
}

// GetBySection returns all options for a specific section.
func (r *MetadataRegistry) GetBySection(section string) []*OptionMetadata {
	var result []*OptionMetadata

	// Case-insensitive section matching
	section = strings.ToLower(section)

	for key, opt := range r.options {
		if strings.HasPrefix(key, section+".") {
			result = append(result, opt)
		}
	}

	return result
}

// GetAll returns all registered option metadata.
func (r *MetadataRegistry) GetAll() []*OptionMetadata {
	result := make([]*OptionMetadata, 0, len(r.options))
	for _, opt := range r.options {
		result = append(result, opt)
	}
	return result
}

// ListSections returns all available section names.
func (r *MetadataRegistry) ListSections() []string {
	return []string{
		SectionDefaults,
		SectionSearch,
		SectionOutput,
		SectionAPI,
	}
}

// Count returns the total number of registered options.
func (r *MetadataRegistry) Count() int {
	return len(r.options)
}

// CountBySection returns the number of options in a specific section.
func (r *MetadataRegistry) CountBySection(section string) int {
	return len(r.GetBySection(section))
}

// Formatter defines the interface for formatting option metadata.
type Formatter interface {
	Format(options []*OptionMetadata) (string, error)
}

// TableFormatter formats options as a table.
type TableFormatter struct {
	maxDescLength int
}

// NewTableFormatter creates a new table formatter.
func NewTableFormatter() *TableFormatter {
	return &TableFormatter{
		maxDescLength: defaultMaxDescLength,
	}
}

// Format formats options as a table with aligned columns.
func (f *TableFormatter) Format(options []*OptionMetadata) (string, error) {
	if len(options) == 0 {
		return "No options found.\n", nil
	}

	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, defaultTabPadding, ' ', 0)

	// Write header
	if _, err := fmt.Fprintf(w, "SECTION\tNAME\tTYPE\tDEFAULT\tDESCRIPTION\n"); err != nil {
		return "", fmt.Errorf("failed to write header: %w", err)
	}
	if _, err := fmt.Fprintf(w, "-------\t----\t----\t-------\t-----------\n"); err != nil {
		return "", fmt.Errorf("failed to write separator: %w", err)
	}

	// Write each option
	for _, opt := range options {
		section := opt.Section
		name := opt.Name
		typ := opt.Type
		def := formatDefault(opt.Default)
		desc := truncate(opt.Description, f.maxDescLength)

		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", section, name, typ, def, desc); err != nil {
			return "", fmt.Errorf("failed to write row: %w", err)
		}
	}

	if err := w.Flush(); err != nil {
		return "", fmt.Errorf("failed to flush writer: %w", err)
	}
	return buf.String(), nil
}

// JSONFormatter formats options as JSON.
type JSONFormatter struct {
	indent bool
}

// NewJSONFormatter creates a new JSON formatter.
func NewJSONFormatter(indent bool) *JSONFormatter {
	return &JSONFormatter{indent: indent}
}

// Format formats options as JSON.
func (f *JSONFormatter) Format(options []*OptionMetadata) (string, error) {
	if len(options) == 0 {
		return "[]", nil
	}

	var data []byte
	var err error

	if f.indent {
		data, err = json.MarshalIndent(options, "", "  ")
	} else {
		data, err = json.Marshal(options)
	}

	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(data), nil
}

// YAMLFormatter formats options as YAML.
type YAMLFormatter struct{}

// NewYAMLFormatter creates a new YAML formatter.
func NewYAMLFormatter() *YAMLFormatter {
	return &YAMLFormatter{}
}

// Format formats options as YAML.
func (f *YAMLFormatter) Format(options []*OptionMetadata) (string, error) {
	if len(options) == 0 {
		return "[]", nil
	}

	data, err := yaml.Marshal(options)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return string(data), nil
}

// FormatOptions formats option metadata using the specified format.
// Supported formats: table, json, yaml.
func FormatOptions(options []*OptionMetadata, format string) (string, error) {
	var formatter Formatter

	switch strings.ToLower(format) {
	case "table":
		formatter = NewTableFormatter()
	case "json":
		formatter = NewJSONFormatter(true)
	case "yaml", "yml":
		formatter = NewYAMLFormatter()
	default:
		return "", fmt.Errorf("%w: %s (supported: table, json, yaml)", ErrUnsupportedFormat, format)
	}

	result, err := formatter.Format(options)
	if err != nil {
		return "", fmt.Errorf("format failed: %w", err)
	}
	return result, nil
}

// formatDefault formats a default value for display.
func formatDefault(val any) string {
	if val == nil {
		return "(none)"
	}

	switch v := val.(type) {
	case string:
		if v == "" {
			return "(empty)"
		}
		return fmt.Sprintf("%q", v)
	case bool:
		return strconv.FormatBool(v)
	case int, int32, int64:
		if v == 0 {
			return "(unset)"
		}
		return fmt.Sprintf("%v", v)
	case float64:
		if v == 0.0 {
			return "(unset)"
		}
		return fmt.Sprintf("%.1f", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// truncate truncates a string to the specified length, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= defaultMinTruncate {
		return s[:maxLen]
	}
	return s[:maxLen-defaultMinTruncate] + "..."
}
