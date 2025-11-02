package config

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// maxCommentWidth is the maximum width for comment lines before wrapping.
	maxCommentWidth = 78 // 80 - 2 for "# "

	// commentPrefix is the YAML comment prefix.
	commentPrefix = "# "

	// minHeaderWidth is the minimum width for section headers.
	minHeaderWidth = 40

	// boxPadding is the padding around text in box headers.
	boxPadding = 4

	// boxBorderWidth is the width of box borders (left + right).
	boxBorderWidth = 2

	// boxTextOffset is the offset for text positioning in box.
	boxTextOffset = 3

	// commentIndent is the indentation for field comments.
	commentIndent = 2
)

// HeaderStyle defines the style of section headers.
type HeaderStyle string

const (
	// HeaderStyleBox uses Unicode box drawing characters.
	HeaderStyleBox HeaderStyle = "box"

	// HeaderStyleLine uses simple line separators.
	HeaderStyleLine HeaderStyle = "line"

	// HeaderStyleMinimal uses minimal formatting.
	HeaderStyleMinimal HeaderStyle = "minimal"
)

// generateFieldComment creates a formatted YAML comment for a configuration field.
// It includes the field's description, valid values, defaults, and usage tips.
func generateFieldComment(opt *OptionMetadata, indent int) string {
	if opt == nil {
		return ""
	}

	var lines []string

	// Add description
	if opt.Description != "" {
		lines = append(lines, wrapComment(opt.Description, indent)...)
	}

	// Add type information
	if opt.Type != "" {
		lines = append(lines, wrapComment("Type: "+opt.Type, indent)...)
	}

	// Add default value
	if opt.Default != nil {
		defaultStr := formatDefaultForComment(opt.Default)
		lines = append(lines, wrapComment("Default: "+defaultStr, indent)...)
	}

	// Add validation rules
	if len(opt.ValidationRules) > 0 {
		lines = append(lines, wrapComment("Valid values:", indent)...)
		for _, rule := range opt.ValidationRules {
			lines = append(lines, wrapComment("  - "+rule, indent)...)
		}
	}

	// Add example
	if opt.Example != "" {
		lines = append(lines, wrapComment("Example: "+opt.Example, indent)...)
	}

	// Add environment variable
	if opt.EnvVar != "" {
		lines = append(lines, wrapComment("Env var: "+opt.EnvVar, indent)...)
	}

	// Add required indicator
	if opt.Required {
		lines = append(lines, wrapComment("Required: true", indent)...)
	}

	return strings.Join(lines, "\n")
}

// generateSectionHeader creates a formatted section header with the specified style.
// The header width is calculated based on the title length, with a minimum width.
func generateSectionHeader(title string, style HeaderStyle, indent int) string {
	indentStr := strings.Repeat(" ", indent)

	switch style {
	case HeaderStyleBox:
		return generateBoxHeader(title, indentStr)
	case HeaderStyleLine:
		return generateLineHeader(title, indentStr)
	case HeaderStyleMinimal:
		return generateMinimalHeader(title, indentStr)
	default:
		return generateBoxHeader(title, indentStr)
	}
}

// generateBoxHeader creates a box-style header using Unicode box drawing characters.
func generateBoxHeader(title string, indent string) string {
	// Calculate width (title + padding)
	titleLen := len(title)
	width := titleLen + boxPadding
	if width < minHeaderWidth {
		width = minHeaderWidth
	}

	// Create box components
	topLine := indent + commentPrefix + "╭" + strings.Repeat("─", width-boxBorderWidth) + "╮"
	titleLine := indent + commentPrefix + "│ " + title + strings.Repeat(" ", width-titleLen-boxTextOffset) + "│"
	bottomLine := indent + commentPrefix + "╰" + strings.Repeat("─", width-boxBorderWidth) + "╯"

	return topLine + "\n" + titleLine + "\n" + bottomLine
}

// generateLineHeader creates a line-style header with simple separators.
func generateLineHeader(title string, indent string) string {
	width := len(title) + boxPadding
	if width < minHeaderWidth {
		width = minHeaderWidth
	}

	separator := indent + commentPrefix + strings.Repeat("═", width)
	titleLine := indent + commentPrefix + " " + title

	return separator + "\n" + titleLine + "\n" + separator
}

// generateMinimalHeader creates a minimal header with just the title and a simple underline.
func generateMinimalHeader(title string, indent string) string {
	titleLine := indent + commentPrefix + title
	underline := indent + commentPrefix + strings.Repeat("-", len(title))

	return titleLine + "\n" + underline
}

// wrapComment wraps a comment line to maxCommentWidth, preserving indentation.
// Returns a slice of comment lines with proper prefixes and indentation.
func wrapComment(text string, indent int) []string {
	if text == "" {
		return []string{}
	}

	indentStr := strings.Repeat(" ", indent)
	maxWidth := maxCommentWidth - indent

	// If text fits on one line, return it directly
	if len(text) <= maxWidth {
		return []string{indentStr + commentPrefix + text}
	}

	// Wrap text to multiple lines
	words := strings.Fields(text)
	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		// Check if adding this word would exceed the line width
		proposedLength := currentLine.Len()
		if proposedLength > 0 {
			proposedLength++ // for space
		}
		proposedLength += len(word)

		if proposedLength > maxWidth && currentLine.Len() > 0 {
			// Current line is full, start a new line
			lines = append(lines, indentStr+commentPrefix+currentLine.String())
			currentLine.Reset()
			currentLine.WriteString(word)
		} else {
			// Add word to current line
			if currentLine.Len() > 0 {
				currentLine.WriteString(" ")
			}
			currentLine.WriteString(word)
		}
	}

	// Add the last line if not empty
	if currentLine.Len() > 0 {
		lines = append(lines, indentStr+commentPrefix+currentLine.String())
	}

	return lines
}

// formatDefaultForComment formats a default value for display in comments.
// Returns a human-readable string representation without quotes for strings.
//
//nolint:cyclop // Type switch inherently has multiple cases
func formatDefaultForComment(value interface{}) string {
	if value == nil {
		return "none"
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			return "(empty)"
		}
		return v
	case []string:
		if len(v) == 0 {
			return "[]"
		}
		return fmt.Sprintf("[%s]", strings.Join(v, ", "))
	case bool:
		if v {
			return "true"
		}
		return "false"
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%.2f", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// AnnotationOptions controls how the annotated configuration is generated.
type AnnotationOptions struct {
	// IncludeExamples includes example profile configurations in the output.
	IncludeExamples bool

	// HeaderStyle determines the style of section headers.
	HeaderStyle HeaderStyle

	// IncludeDescriptions includes field descriptions in comments.
	IncludeDescriptions bool
}

// DefaultAnnotationOptions returns the default annotation options.
func DefaultAnnotationOptions() AnnotationOptions {
	return AnnotationOptions{
		IncludeExamples:     false,
		HeaderStyle:         HeaderStyleBox,
		IncludeDescriptions: true,
	}
}

// GenerateAnnotatedConfig generates a fully annotated YAML configuration file.
// It includes inline comments explaining all options, section headers, and
// optionally example profiles.
//
//nolint:cyclop,funlen // Configuration generation requires multiple steps
func GenerateAnnotatedConfig(cfg *ConfigData, opts AnnotationOptions) (string, error) {
	if cfg == nil {
		cfg = NewConfigData()
	}

	var output strings.Builder
	registry := NewMetadataRegistry()

	// Add file header
	output.WriteString("# Perplexity CLI Configuration\n")
	output.WriteString("# This file contains all available configuration options with descriptions.\n")
	output.WriteString("#\n\n")

	// Generate Defaults section
	if err := generateSection(&output, "Defaults", SectionDefaults, registry, cfg, opts); err != nil {
		return "", fmt.Errorf("failed to generate defaults section: %w", err)
	}

	// Generate Search section
	if err := generateSection(&output, "Search Options", SectionSearch, registry, cfg, opts); err != nil {
		return "", fmt.Errorf("failed to generate search section: %w", err)
	}

	// Generate Output section
	if err := generateSection(&output, "Output Options", SectionOutput, registry, cfg, opts); err != nil {
		return "", fmt.Errorf("failed to generate output section: %w", err)
	}

	// Generate API section
	if err := generateSection(&output, "API Configuration", SectionAPI, registry, cfg, opts); err != nil {
		return "", fmt.Errorf("failed to generate api section: %w", err)
	}

	// Add profiles section if they exist
	if len(cfg.Profiles) > 0 {
		output.WriteString("\n")
		output.WriteString(generateSectionHeader("Profiles", opts.HeaderStyle, 0))
		output.WriteString("\n")
		output.WriteString("# Named configuration profiles for different use cases\n")
		output.WriteString("\nprofiles:\n")

		for name, profile := range cfg.Profiles {
			output.WriteString(fmt.Sprintf("  %s:\n", name))
			profileYAML, err := yaml.Marshal(profile)
			if err != nil {
				return "", fmt.Errorf("failed to marshal profile %s: %w", name, err)
			}
			// Indent profile YAML
			lines := strings.Split(string(profileYAML), "\n")
			for _, line := range lines {
				if line != "" {
					output.WriteString("    " + line + "\n")
				}
			}
		}
	}

	// Add active profile
	if cfg.ActiveProfile != "" {
		output.WriteString("\n# Currently active profile\n")
		output.WriteString(fmt.Sprintf("active_profile: %s\n", cfg.ActiveProfile))
	}

	// Add example profiles if requested
	if opts.IncludeExamples {
		if err := generateExampleProfiles(&output, opts); err != nil {
			return "", fmt.Errorf("failed to generate example profiles: %w", err)
		}
	}

	return output.String(), nil
}

// generateSection generates an annotated section for a specific configuration category.
func generateSection(
	output *strings.Builder,
	title string,
	section string,
	registry *MetadataRegistry,
	cfg *ConfigData,
	opts AnnotationOptions,
) error {
	options := registry.GetBySection(section)
	if len(options) == 0 {
		return nil
	}

	// Add section header
	output.WriteString(generateSectionHeader(title, opts.HeaderStyle, 0))
	output.WriteString("\n\n")

	// Add section name
	sectionKey := strings.ToLower(section)
	output.WriteString(sectionKey + ":\n")

	// Add each option with comments
	for _, opt := range options {
		if opts.IncludeDescriptions {
			comment := generateFieldComment(opt, commentIndent)
			if comment != "" {
				output.WriteString(comment + "\n")
			}
		}

		// Get the field name (remove section prefix)
		fieldName := strings.TrimPrefix(opt.Name, section+".")

		// Write the field with its value
		_, _ = fmt.Fprintf(output, "  %s: ", fieldName)

		// Get value from config or use default
		value := getConfigValue(cfg, section, fieldName, opt.Default)
		valueYAML, err := yaml.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value for %s: %w", opt.Name, err)
		}

		// Trim the trailing newline from YAML marshal
		output.WriteString(strings.TrimSpace(string(valueYAML)) + "\n")
	}

	output.WriteString("\n")
	return nil
}

// getConfigValue retrieves the value for a field from the config, or returns the default.
//
//nolint:gocognit,cyclop,gocyclo,funlen // Config field mapping requires many cases
func getConfigValue(cfg *ConfigData, section, fieldName string, defaultValue interface{}) interface{} {
	switch section {
	case SectionDefaults:
		switch fieldName {
		case "model":
			if cfg.Defaults.Model != "" {
				return cfg.Defaults.Model
			}
		case "temperature":
			if cfg.Defaults.Temperature != 0 {
				return cfg.Defaults.Temperature
			}
		case "max_tokens":
			if cfg.Defaults.MaxTokens != 0 {
				return cfg.Defaults.MaxTokens
			}
		case "top_k":
			if cfg.Defaults.TopK != 0 {
				return cfg.Defaults.TopK
			}
		case "top_p":
			if cfg.Defaults.TopP != 0 {
				return cfg.Defaults.TopP
			}
		case "frequency_penalty":
			if cfg.Defaults.FrequencyPenalty != 0 {
				return cfg.Defaults.FrequencyPenalty
			}
		case "presence_penalty":
			if cfg.Defaults.PresencePenalty != 0 {
				return cfg.Defaults.PresencePenalty
			}
		case "timeout":
			if cfg.Defaults.Timeout != "" {
				return cfg.Defaults.Timeout
			}
		}
	case SectionSearch:
		switch fieldName {
		case "domains":
			if len(cfg.Search.Domains) > 0 {
				return cfg.Search.Domains
			}
		case "recency":
			if cfg.Search.Recency != "" {
				return cfg.Search.Recency
			}
		case "mode":
			if cfg.Search.Mode != "" {
				return cfg.Search.Mode
			}
		case "context_size":
			if cfg.Search.ContextSize != "" {
				return cfg.Search.ContextSize
			}
		case "location_lat":
			if cfg.Search.LocationLat != 0 {
				return cfg.Search.LocationLat
			}
		case "location_lon":
			if cfg.Search.LocationLon != 0 {
				return cfg.Search.LocationLon
			}
		case "location_country":
			if cfg.Search.LocationCountry != "" {
				return cfg.Search.LocationCountry
			}
		case "after_date":
			if cfg.Search.AfterDate != "" {
				return cfg.Search.AfterDate
			}
		case "before_date":
			if cfg.Search.BeforeDate != "" {
				return cfg.Search.BeforeDate
			}
		case "last_updated_after":
			if cfg.Search.LastUpdatedAfter != "" {
				return cfg.Search.LastUpdatedAfter
			}
		case "last_updated_before":
			if cfg.Search.LastUpdatedBefore != "" {
				return cfg.Search.LastUpdatedBefore
			}
		}
	case SectionOutput:
		switch fieldName {
		case "stream":
			return cfg.Output.Stream
		case "return_images":
			return cfg.Output.ReturnImages
		case "return_related":
			return cfg.Output.ReturnRelated
		case "json":
			return cfg.Output.JSON
		case "image_domains":
			if len(cfg.Output.ImageDomains) > 0 {
				return cfg.Output.ImageDomains
			}
		case "image_formats":
			if len(cfg.Output.ImageFormats) > 0 {
				return cfg.Output.ImageFormats
			}
		case "response_format_json_schema":
			if cfg.Output.ResponseFormatJSONSchema != "" {
				return cfg.Output.ResponseFormatJSONSchema
			}
		case "response_format_regex":
			if cfg.Output.ResponseFormatRegex != "" {
				return cfg.Output.ResponseFormatRegex
			}
		case "reasoning_effort":
			if cfg.Output.ReasoningEffort != "" {
				return cfg.Output.ReasoningEffort
			}
		}
	case SectionAPI:
		switch fieldName {
		case "key":
			if cfg.API.Key != "" {
				return cfg.API.Key
			}
		case "base_url":
			if cfg.API.BaseURL != "" {
				return cfg.API.BaseURL
			}
		case "timeout":
			if cfg.API.Timeout != 0 {
				return cfg.API.Timeout
			}
		}
	}

	return defaultValue
}

// generateExampleProfiles adds example profile configurations to the output.
//
//nolint:funlen // Example profile generation requires comprehensive template loading
func generateExampleProfiles(output *strings.Builder, opts AnnotationOptions) error {
	output.WriteString("\n\n")
	output.WriteString(generateSectionHeader("Example Profiles", opts.HeaderStyle, 0))
	output.WriteString("\n")
	output.WriteString("# Below are example profile configurations for common use cases.\n")
	output.WriteString("# Uncomment and customize as needed.\n\n")

	examples := []struct {
		name        string
		description string
		template    string
	}{
		{
			name:        "research",
			description: "Academic research with authoritative sources",
			template:    TemplateResearch,
		},
		{
			name:        "creative",
			description: "Creative writing and brainstorming",
			template:    TemplateCreative,
		},
		{
			name:        "news",
			description: "Current events and news coverage",
			template:    TemplateNews,
		},
	}

	for _, example := range examples {
		_, _ = fmt.Fprintf(output, "# Example: %s - %s\n", example.name, example.description)
		output.WriteString("# profiles:\n")
		_, _ = fmt.Fprintf(output, "#   %s:\n", example.name)

		// Load template and format it
		template, err := LoadTemplate(example.template)
		if err != nil {
			return fmt.Errorf("failed to load template %s: %w", example.template, err)
		}

		// Marshal the template's defaults/search/output sections
		sections := []struct {
			name  string
			value interface{}
		}{
			{"defaults", template.Defaults},
			{"search", template.Search},
			{"output", template.Output},
		}

		for _, section := range sections {
			yamlData, err := yaml.Marshal(map[string]interface{}{section.name: section.value})
			if err != nil {
				return fmt.Errorf("failed to marshal %s section: %w", section.name, err)
			}

			// Comment out and indent the YAML
			lines := strings.Split(string(yamlData), "\n")
			for _, line := range lines {
				if line != "" {
					output.WriteString("#     " + line + "\n")
				}
			}
		}
		output.WriteString("#\n")
	}

	return nil
}
