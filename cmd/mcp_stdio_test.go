package cmd

import (
	"errors"
	"os"
	"strings"
	"testing"

	clerrors "github.com/sgaunet/pplx/pkg/clerrors"
)

func TestMcpStdioCmd_MissingAPIKey(t *testing.T) {
	t.Setenv("PPLX_API_KEY", "")

	err := mcpStdioCmd.RunE(mcpStdioCmd, []string{})
	if err == nil {
		t.Fatal("expected error when PPLX_API_KEY is not set, got nil")
	}

	var configErr *clerrors.ConfigError
	if !errors.As(err, &configErr) {
		t.Errorf("expected clerrors.ConfigError, got %T: %v", err, err)
	}
}

func TestMcpStdioCmd_MissingAPIKey_Message(t *testing.T) {
	t.Setenv("PPLX_API_KEY", "")

	err := mcpStdioCmd.RunE(mcpStdioCmd, []string{})
	if err == nil {
		t.Fatal("expected error when PPLX_API_KEY is not set, got nil")
	}

	if !strings.Contains(err.Error(), "PPLX_API_KEY") {
		t.Errorf("expected error message to contain 'PPLX_API_KEY', got: %s", err.Error())
	}
}

func TestMcpStdioCmd_WithAPIKey_ReachesStart(t *testing.T) {
	t.Setenv("PPLX_API_KEY", "test-api-key-for-mcp")

	// Replace os.Stdin with a pipe whose write end is immediately closed.
	// This causes ServeStdio to read EOF and return, unblocking the test.
	origStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	w.Close() // EOF immediately
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = origStdin
		r.Close()
	})

	err = mcpStdioCmd.RunE(mcpStdioCmd, []string{})

	// If we got past config checks, the error (if any) should NOT be a ConfigError.
	// ServeStdio may return nil or an error, but it should not be a config error.
	if err != nil {
		var configErr *clerrors.ConfigError
		if errors.As(err, &configErr) {
			t.Errorf("expected to pass config checks, but got ConfigError: %v", err)
		}
	}
}

func TestMcpStdioCmd_CommandRegistered(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "mcp-stdio" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'mcp-stdio' command to be registered on rootCmd")
	}
}
