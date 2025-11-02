package config

import (
	"embed"
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Embed all template files from the templates directory
//
//go:embed templates/*.yaml
var templatesFS embed.FS

// Template names.
const (
	TemplateResearch    = "research"
	TemplateCreative    = "creative"
	TemplateNews        = "news"
	TemplateFullExample = "full-example"
)

var (
	// ErrTemplateNotFound is returned when a template name is not recognized.
	ErrTemplateNotFound = errors.New("template not found")

	// ErrTemplateInvalid is returned when a template file cannot be parsed.
	ErrTemplateInvalid = errors.New("template is invalid")
)

// templateFileMap maps template names to their file paths in the embedded filesystem.
var templateFileMap = map[string]string{
	TemplateResearch:    "templates/research.yaml",
	TemplateCreative:    "templates/creative.yaml",
	TemplateNews:        "templates/news.yaml",
	TemplateFullExample: "templates/full-example.yaml",
}

// TemplateInfo contains metadata about a configuration template.
type TemplateInfo struct {
	// Name is the template identifier used with LoadTemplate
	Name string

	// Description provides a brief overview of the template's purpose
	Description string

	// UseCase describes the primary use case for this template
	UseCase string
}

// LoadTemplate loads a configuration template by name.
// Valid template names are: research, creative, news, full-example.
// Returns a ConfigData struct populated with the template's configuration,
// or an error if the template name is invalid or the template cannot be parsed.
func LoadTemplate(name string) (*ConfigData, error) {
	// Check if template exists
	filePath, exists := templateFileMap[name]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrTemplateNotFound, name)
	}

	// Read the embedded template file
	data, err := templatesFS.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read template %s: %w", ErrTemplateInvalid, name, err)
	}

	// Parse YAML into ConfigData
	var config ConfigData
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("%w: failed to parse template %s: %w", ErrTemplateInvalid, name, err)
	}

	return &config, nil
}

// ListTemplates returns a list of all available configuration templates.
// Each template includes metadata describing its name, purpose, and use case.
func ListTemplates() []TemplateInfo {
	return []TemplateInfo{
		{
			Name:        TemplateResearch,
			Description: "Optimized for academic and scholarly research with authoritative sources",
			UseCase:     "Academic research, literature reviews, and scholarly inquiries requiring peer-reviewed sources",
		},
		{
			Name:        TemplateCreative,
			Description: "High-temperature creative configuration with streaming enabled",
			UseCase:     "Creative writing, brainstorming, content generation, and exploratory queries",
		},
		{
			Name:        TemplateNews,
			Description: "News-focused configuration with reputable news sources and weekly recency filter",
			UseCase:     "Current events tracking, news analysis, and recent developments research",
		},
		{
			Name:        TemplateFullExample,
			Description: "Comprehensive example showing all 29+ configuration options with detailed comments",
			UseCase:     "Learning available options, creating custom configurations, and understanding configuration structure",
		},
	}
}

// GetTemplateDescription returns detailed information about a specific template.
// Returns an error if the template name is not recognized.
func GetTemplateDescription(name string) (*TemplateInfo, error) {
	templates := ListTemplates()
	for i := range templates {
		if templates[i].Name == name {
			return &templates[i], nil
		}
	}
	return nil, fmt.Errorf("%w: %s", ErrTemplateNotFound, name)
}

// IsTemplateNotFoundError checks if an error is a template not found error.
func IsTemplateNotFoundError(err error) bool {
	return errors.Is(err, ErrTemplateNotFound)
}
