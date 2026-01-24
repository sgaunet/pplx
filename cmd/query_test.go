package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// TestSearchRecencyImagesIncompatibility tests that when both search-recency
// and return-images are specified, a clear message is shown explaining the
// auto-correction behavior (Option 2 from issue #87)
func TestSearchRecencyImagesIncompatibility(t *testing.T) {
	// Skip if API key is not set
	if os.Getenv("PPLX_API_KEY") == "" {
		t.Skip("Skipping test: PPLX_API_KEY not set")
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set flags that trigger the incompatibility
	globalOpts.SearchRecency = "week"
	globalOpts.ReturnImages = true
	globalOpts.UserPrompt = "test query"

	// Create a goroutine to capture output
	outputChan := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outputChan <- buf.String()
	}()

	// Run the command (this will make an API call, so we need the key)
	// Note: This is an integration test and will fail if the API is down
	// or if the key is invalid
	err := queryCmd.RunE(queryCmd, []string{})

	// Restore stdout
	_ = w.Close()
	os.Stdout = oldStdout
	output := <-outputChan

	// The command might fail due to API issues, but we're testing the warning message
	// so we check the output regardless of the error
	_ = err

	// Check that the expected message is present
	expectedMessage := "Note: When using --return-images, search-recency is automatically disabled"
	if !strings.Contains(output, expectedMessage) {
		t.Errorf("Expected output to contain: %q\nGot: %s", expectedMessage, output)
	}

	// Also check for the second line
	expectedMessage2 := "Proceeding with image search..."
	if !strings.Contains(output, expectedMessage2) {
		t.Errorf("Expected output to contain: %q\nGot: %s", expectedMessage2, output)
	}

	// Reset flags
	globalOpts.SearchRecency = ""
	globalOpts.ReturnImages = false
	globalOpts.UserPrompt = ""
}

// TestSearchRecencyWithoutImages tests that when search-recency is used
// without return-images, no warning is shown
func TestSearchRecencyWithoutImages(t *testing.T) {
	// Skip if API key is not set
	if os.Getenv("PPLX_API_KEY") == "" {
		t.Skip("Skipping test: PPLX_API_KEY not set")
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set flags without the incompatibility
	globalOpts.SearchRecency = "week"
	globalOpts.ReturnImages = false
	globalOpts.UserPrompt = "test query"

	// Create a goroutine to capture output
	outputChan := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outputChan <- buf.String()
	}()

	// Run the command
	err := queryCmd.RunE(queryCmd, []string{})

	// Restore stdout
	_ = w.Close()
	os.Stdout = oldStdout
	output := <-outputChan

	// The command might fail due to API issues, but we're testing the warning message
	_ = err

	// Check that the warning message is NOT present
	unexpectedMessage := "Note: When using --return-images"
	if strings.Contains(output, unexpectedMessage) {
		t.Errorf("Did not expect output to contain: %q\nGot: %s", unexpectedMessage, output)
	}

	// Reset flags
	globalOpts.SearchRecency = ""
	globalOpts.ReturnImages = false
	globalOpts.UserPrompt = ""
}
