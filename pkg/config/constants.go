package config

// Validation maps for Perplexity API enum values.
// These define the valid parameter values accepted by the API.
// Centralized here to eliminate duplication across validator, chat, mcp, and cmd packages.

// ValidSearchRecency contains valid search recency time windows supported by the Perplexity API.
// These values control how recent search results must be.
// Valid values: "hour", "day", "week", "month", "year"
// Used by: validator, chat builder, MCP handler, config wizard.
var ValidSearchRecency = map[string]bool{
	"hour": true, "day": true, "week": true, "month": true, "year": true,
}

// ValidSearchModes contains valid search modes supported by the Perplexity API.
// Valid values:
//   - "web": general internet search (default, broader results)
//   - "academic": scholarly sources only (research papers, journals)
//
// Used by: chat builder, MCP handler.
var ValidSearchModes = map[string]bool{
	"web": true, "academic": true,
}

// ValidContextSizes contains valid search context sizes supported by the Perplexity API.
// These control how much search context to include in the model prompt.
// Valid values:
//   - "low": minimal context (faster, cheaper)
//   - "medium": balanced context (default)
//   - "high": maximum context (slower, more comprehensive)
//
// Used by: validator, chat builder, MCP handler, config wizard.
var ValidContextSizes = map[string]bool{
	"low": true, "medium": true, "high": true,
}

// ValidReasoningEfforts contains valid reasoning effort levels for deep-research models.
// These control the depth of analysis for sonar-deep-research models.
// Valid values:
//   - "low": faster, less thorough (suitable for simple queries)
//   - "medium": balanced analysis (default)
//   - "high": maximum depth (slower, comprehensive for complex research tasks)
//
// Used by: validator, chat builder, MCP handler.
var ValidReasoningEfforts = map[string]bool{
	"low": true, "medium": true, "high": true,
}

// ValidImageFormats contains valid image format filters supported by the Perplexity API.
// Valid formats: "jpg", "jpeg", "png", "gif", "webp", "svg", "bmp"
// Used by: MCP handler (for warnings).
var ValidImageFormats = map[string]bool{
	"jpg": true, "jpeg": true, "png": true, "gif": true,
	"webp": true, "svg": true, "bmp": true,
}

// Helper functions for validation

// IsValidSearchRecency validates a search recency value against the Perplexity API specification.
// Returns true if the value is one of: hour, day, week, month, year.
func IsValidSearchRecency(value string) bool {
	return ValidSearchRecency[value]
}

// IsValidSearchMode validates a search mode against the Perplexity API specification.
// Returns true if the value is one of: web, academic.
func IsValidSearchMode(value string) bool {
	return ValidSearchModes[value]
}

// IsValidContextSize validates a context size against the Perplexity API specification.
// Returns true if the value is one of: low, medium, high.
func IsValidContextSize(value string) bool {
	return ValidContextSizes[value]
}

// IsValidReasoningEffort validates a reasoning effort level against the Perplexity API specification.
// Returns true if the value is one of: low, medium, high.
func IsValidReasoningEffort(value string) bool {
	return ValidReasoningEfforts[value]
}

// IsValidImageFormat validates an image format against the Perplexity API specification.
// Returns true if the format is one of: jpg, jpeg, png, gif, webp, svg, bmp.
func IsValidImageFormat(value string) bool {
	return ValidImageFormats[value]
}

// Slice getters for iteration (needed by config wizard and display purposes)

// GetValidSearchRecencyValues returns all valid search recency values as a slice.
// Useful for iterating over valid options in CLI prompts or documentation.
func GetValidSearchRecencyValues() []string {
	return []string{"hour", "day", "week", "month", "year"}
}

// GetValidSearchModeValues returns all valid search mode values as a slice.
// Useful for iterating over valid options in CLI prompts or documentation.
func GetValidSearchModeValues() []string {
	return []string{"web", "academic"}
}

// GetValidContextSizeValues returns all valid context size values as a slice.
// Useful for iterating over valid options in CLI prompts or documentation.
func GetValidContextSizeValues() []string {
	return []string{"low", "medium", "high"}
}

// GetValidReasoningEffortValues returns all valid reasoning effort values as a slice.
// Useful for iterating over valid options in CLI prompts or documentation.
func GetValidReasoningEffortValues() []string {
	return []string{"low", "medium", "high"}
}

// GetValidImageFormatValues returns all valid image format values as a slice.
// Useful for iterating over valid options in CLI prompts or documentation.
func GetValidImageFormatValues() []string {
	return []string{"jpg", "jpeg", "png", "gif", "webp", "svg", "bmp"}
}
