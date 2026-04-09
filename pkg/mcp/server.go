package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// MCPServer wraps the MCP server with Perplexity query functionality.
//nolint:revive // MCPServer is the appropriate name for this MCP-specific server type
type MCPServer struct {
	server    *server.MCPServer
	handler   *QueryHandler
	extractor *ParameterExtractor
	formatter *ResponseFormatter
	apiKey    string
	version   string
}

// ServerConfig contains configuration for the MCP server.
type ServerConfig struct {
	APIKey  string
	Version string
	Name    string
}

// NewServer creates a new MCP server instance.
func NewServer(config ServerConfig) (*MCPServer, error) {
	if config.APIKey == "" {
		return nil, NewParameterError("api_key", nil, "API key is required")
	}

	if config.Name == "" {
		config.Name = "Perplexity MCP Server"
	}

	if config.Version == "" {
		config.Version = "1.0.0"
	}

	// Create MCP server
	s := server.NewMCPServer(
		config.Name,
		config.Version,
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(false, false),
	)

	return &MCPServer{
		server:    s,
		handler:   NewQueryHandler(),
		extractor: NewParameterExtractor(),
		formatter: NewResponseFormatter(),
		apiKey:    config.APIKey,
		version:   config.Version,
	}, nil
}

// AddQueryTool registers the query tool with the server.
func (s *MCPServer) AddQueryTool() error {
	tool := BuildQueryTool()

	s.server.AddTool(*tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := request.GetArguments()

		// Extract parameters
		params, err := s.extractor.Extract(args)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Handle query
		response, err := s.handler.Handle(ctx, s.apiKey, *params)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// Format response
		return s.formatter.Format(response)
	})

	return nil
}

// Start begins serving stdio requests.
func (s *MCPServer) Start() error {
	if err := server.ServeStdio(s.server); err != nil {
		return fmt.Errorf("failed to serve stdio: %w", err)
	}
	return nil
}
