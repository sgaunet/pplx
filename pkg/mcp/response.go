package mcp

import (
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sgaunet/perplexity-go/v2"
)

// ResponseFormatter formats Perplexity API responses for MCP.
type ResponseFormatter struct{}

// NewResponseFormatter creates a new response formatter.
func NewResponseFormatter() *ResponseFormatter {
	return &ResponseFormatter{}
}

// Format converts a Perplexity response to an MCP tool result.
func (f *ResponseFormatter) Format(response *perplexity.CompletionResponse) (*mcp.CallToolResult, error) {
	if response == nil {
		return mcp.NewToolResultError("No response received"), nil
	}

	if len(response.Choices) == 0 {
		return mcp.NewToolResultError("Response contains no choices"), nil
	}

	// Build response object
	result := f.buildResponse(response)

	// Convert to JSON
	jsonData, err := json.Marshal(result)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// buildResponse creates the response map with all available data.
func (f *ResponseFormatter) buildResponse(response *perplexity.CompletionResponse) map[string]any {
	result := map[string]any{
		"content": response.Choices[0].Message.Content,
		"model":   response.Model,
		"usage":   response.Usage,
	}

	// Add search results if available
	if response.SearchResults != nil && len(*response.SearchResults) > 0 {
		result["search_results"] = *response.SearchResults
	}

	// Add images if available
	if response.Images != nil && len(*response.Images) > 0 {
		result["images"] = *response.Images
	}

	// Add related questions if available
	if response.RelatedQuestions != nil && len(*response.RelatedQuestions) > 0 {
		result["related_questions"] = *response.RelatedQuestions
	}

	return result
}

// FormatError creates an MCP error result from a Go error.
func FormatError(err error) *mcp.CallToolResult {
	return mcp.NewToolResultError(fmt.Sprintf("Request failed: %v", err))
}
