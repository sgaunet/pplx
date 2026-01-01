package cmd

import (
	"os"

	"github.com/sgaunet/pplx/pkg/clerrors"
	"github.com/sgaunet/pplx/pkg/mcp"
	"github.com/spf13/cobra"
)

var mcpStdioCmd = &cobra.Command{
	Use:   "mcp-stdio",
	Short: "Start MCP server in stdio mode",
	Long:  `Start an MCP (Model Context Protocol) server that exposes Perplexity query functionality`,
	RunE: func(_ *cobra.Command, _ []string) error {
		// Check env var PPLX_API_KEY exists
		apiKey := os.Getenv("PPLX_API_KEY")
		if apiKey == "" {
			return clerrors.NewConfigError("PPLX_API_KEY environment variable is not set", nil)
		}

		// Create server configuration
		config := mcp.ServerConfig{
			APIKey:  apiKey,
			Version: version,
			Name:    "Perplexity MCP Server",
		}

		// Create MCP server
		server, err := mcp.NewServer(config)
		if err != nil {
			return clerrors.NewConfigError("Failed to create MCP server", err)
		}

		// Add query tool
		if err := server.AddQueryTool(); err != nil {
			return clerrors.NewConfigError("Failed to add query tool", err)
		}

		// Start the stdio server
		if err := server.Start(); err != nil {
			return clerrors.NewAPIError("MCP server error", err)
		}

		return nil
	},
}
