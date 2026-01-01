package mcp

import (
	"errors"
	"testing"
)

func TestNewServer(t *testing.T) {
	t.Run("creates server with valid config", func(t *testing.T) {
		config := ServerConfig{
			APIKey:  "test-key",
			Version: "1.0.0",
			Name:    "Test Server",
		}

		server, err := NewServer(config)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if server == nil {
			t.Fatal("Expected server, got nil")
		}

		if server.apiKey != "test-key" {
			t.Errorf("Expected apiKey %q, got %q", "test-key", server.apiKey)
		}

		if server.version != "1.0.0" {
			t.Errorf("Expected version %q, got %q", "1.0.0", server.version)
		}

		if server.handler == nil {
			t.Error("Expected handler to be initialized")
		}

		if server.extractor == nil {
			t.Error("Expected extractor to be initialized")
		}

		if server.formatter == nil {
			t.Error("Expected formatter to be initialized")
		}

		if server.server == nil {
			t.Error("Expected MCP server to be initialized")
		}
	})

	t.Run("applies default name", func(t *testing.T) {
		config := ServerConfig{
			APIKey:  "test-key",
			Version: "1.0.0",
		}

		server, err := NewServer(config)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Default name should be applied by the server
		if server == nil {
			t.Fatal("Expected server, got nil")
		}
	})

	t.Run("applies default version", func(t *testing.T) {
		config := ServerConfig{
			APIKey: "test-key",
			Name:   "Test",
		}

		server, err := NewServer(config)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if server.version != "1.0.0" {
			t.Errorf("Expected default version %q, got %q", "1.0.0", server.version)
		}
	})

	t.Run("rejects empty API key", func(t *testing.T) {
		config := ServerConfig{
			Version: "1.0.0",
			Name:    "Test",
		}

		_, err := NewServer(config)
		if err == nil {
			t.Fatal("Expected error for empty API key, got nil")
		}

		var paramErr *ParameterError
		if !errors.As(err, &paramErr) {
			t.Errorf("Expected ParameterError, got %T", err)
		}
	})

	t.Run("creates all components", func(t *testing.T) {
		config := ServerConfig{
			APIKey:  "test-key",
			Version: "2.0.0",
			Name:    "Full Test Server",
		}

		server, err := NewServer(config)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Verify all components are properly initialized
		if server.handler == nil {
			t.Error("Handler should be initialized")
		}

		if server.extractor == nil {
			t.Error("Extractor should be initialized")
		}

		if server.formatter == nil {
			t.Error("Formatter should be initialized")
		}

		if server.server == nil {
			t.Error("Underlying MCP server should be initialized")
		}
	})
}

func TestMCPServer_AddQueryTool(t *testing.T) {
	t.Run("adds query tool successfully", func(t *testing.T) {
		config := ServerConfig{
			APIKey:  "test-key",
			Version: "1.0.0",
			Name:    "Test",
		}

		server, err := NewServer(config)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		err = server.AddQueryTool()
		if err != nil {
			t.Fatalf("Unexpected error adding tool: %v", err)
		}
	})

	t.Run("can add tool multiple times", func(t *testing.T) {
		config := ServerConfig{
			APIKey:  "test-key",
			Version: "1.0.0",
			Name:    "Test",
		}

		server, err := NewServer(config)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Adding tool multiple times should not error
		err = server.AddQueryTool()
		if err != nil {
			t.Fatalf("Unexpected error on first add: %v", err)
		}

		err = server.AddQueryTool()
		if err != nil {
			t.Fatalf("Unexpected error on second add: %v", err)
		}
	})
}
