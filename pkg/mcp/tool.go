package mcp

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sgaunet/perplexity-go/v2"
)

// BuildQueryTool creates the MCP tool definition for Perplexity queries.
//nolint:funlen // Function length appropriate for defining 30+ parameters
func BuildQueryTool() *mcp.Tool {
	tool := mcp.NewTool("query",
		mcp.WithDescription("Query Perplexity AI with extensive search and filtering options"),
		// Required parameters
		mcp.WithString("user_prompt",
			mcp.Required(),
			mcp.Description("The user query/prompt"),
		),
		// Optional core parameters
		mcp.WithString("system_prompt",
			mcp.Description("System prompt to guide the AI response"),
		),
		mcp.WithString("model",
			mcp.Description("Model to use (default: "+perplexity.DefaultModel+")"),
		),
		mcp.WithNumber("frequency_penalty",
			mcp.Description("Frequency penalty for response generation"),
		),
		mcp.WithNumber("max_tokens",
			mcp.Description("Maximum number of tokens in response"),
		),
		mcp.WithNumber("presence_penalty",
			mcp.Description("Presence penalty for response generation"),
		),
		mcp.WithNumber("temperature",
			mcp.Description("Temperature for response generation"),
		),
		mcp.WithNumber("top_k",
			mcp.Description("Top-K sampling parameter"),
		),
		mcp.WithNumber("top_p",
			mcp.Description("Top-P sampling parameter"),
		),
		mcp.WithNumber("timeout",
			mcp.Description("HTTP timeout in seconds"),
		),
		// Search/Web options
		mcp.WithArray("search_domains",
			mcp.Description("Filter search results to specific domains"),
		),
		mcp.WithString("search_recency",
			mcp.Description("Filter by time: day, week, month, year, hour"),
		),
		mcp.WithNumber("location_lat",
			mcp.Description("User location latitude"),
		),
		mcp.WithNumber("location_lon",
			mcp.Description("User location longitude"),
		),
		mcp.WithString("location_country",
			mcp.Description("User location country code"),
		),
		// Response enhancement options
		mcp.WithBoolean("return_images",
			mcp.Description("Include images in response"),
		),
		mcp.WithBoolean("return_related",
			mcp.Description("Include related questions"),
		),
		mcp.WithBoolean("stream",
			mcp.Description("Enable streaming responses (will be collected and returned as complete response)"),
		),
		// Image filtering options
		mcp.WithArray("image_domains",
			mcp.Description("Filter images by domains"),
		),
		mcp.WithArray("image_formats",
			mcp.Description("Filter images by formats (jpg, png, etc.)"),
		),
		// Response format options
		mcp.WithString("response_format_json_schema",
			mcp.Description("JSON schema for structured output (sonar model only)"),
		),
		mcp.WithString("response_format_regex",
			mcp.Description("Regex pattern for structured output (sonar model only)"),
		),
		// Search mode options
		mcp.WithString("search_mode",
			mcp.Description("Search mode: web (default) or academic"),
		),
		mcp.WithString("search_context_size",
			mcp.Description("Search context size: low, medium, or high"),
		),
		// Date filtering options
		mcp.WithString("search_after_date",
			mcp.Description("Filter results published after date (MM/DD/YYYY)"),
		),
		mcp.WithString("search_before_date",
			mcp.Description("Filter results published before date (MM/DD/YYYY)"),
		),
		mcp.WithString("last_updated_after",
			mcp.Description("Filter results last updated after date (MM/DD/YYYY)"),
		),
		mcp.WithString("last_updated_before",
			mcp.Description("Filter results last updated before date (MM/DD/YYYY)"),
		),
		// Deep research options
		mcp.WithString("reasoning_effort",
			mcp.Description("Reasoning effort for sonar-deep-research: low, medium, or high"),
		),
	)
	return &tool
}
