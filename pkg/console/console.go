// Package console provides console input/output utilities for the Perplexity CLI.
package console

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/sgaunet/perplexity-go/v2"
)

// DefaultLineLength is the default line length for markdown rendering.
const DefaultLineLength = 80

// DefaultLeftMargin is the default left margin for markdown rendering.
const DefaultLeftMargin = 6

// Input prompts the user for input and returns the entered text.
func Input(label string) (string, error) {
	fmt.Printf("%s: (set an empty line to validate the entry)\n", label)

	scanner := bufio.NewScanner(os.Stdin)
	var buf strings.Builder
	for {
		scanner.Scan()
		line := scanner.Text()
		if len(line) == 0 {
			break
		}
		if buf.Len() > 0 {
			buf.WriteString(" ")
		}
		buf.WriteString(line)
	}

	if err := scanner.Err(); err != nil {
		return buf.String(), fmt.Errorf("error reading input: %w", err)
	}
	return buf.String(), nil
}

// RenderAsMarkdown renders the response content as markdown.
func RenderAsMarkdown(pplxResponse *perplexity.CompletionResponse, output io.Writer) error {
	result := markdown.Render(pplxResponse.GetLastContent(), DefaultLineLength, DefaultLeftMargin)
	_, err := fmt.Fprintln(output, string(result))
	if err != nil {
		return fmt.Errorf("error writing markdown to output: %w", err)
	}
	return nil
}

// RenderCitations renders the citations from the response.
func RenderCitations(pplxResponse *perplexity.CompletionResponse, output io.Writer) error {
	searchResults := pplxResponse.GetSearchResults()
	if len(searchResults) == 0 {
		// Fall back to deprecated citations if no search results
		for i, c := range pplxResponse.GetCitations() {
			_, err := fmt.Fprintf(output, "[%d]: %s\n", i, c)
			if err != nil {
				return fmt.Errorf("error writing citations as markdown to output: %w", err)
			}
		}
		return nil
	}
	
	for i, sr := range searchResults {
		dateInfo := ""
		if sr.Date != nil {
			dateInfo = fmt.Sprintf(" (date: %s)", *sr.Date)
		} else if sr.LastUpdated != nil {
			dateInfo = fmt.Sprintf(" (updated: %s)", *sr.LastUpdated)
		}
		_, err := fmt.Fprintf(output, "[%d]: %s - %s%s\n", i, sr.Title, sr.URL, dateInfo)
		if err != nil {
			return fmt.Errorf("error writing search results to output: %w", err)
		}
	}
	return nil
}

// RenderImages renders the images from the response.
func RenderImages(pplxResponse *perplexity.CompletionResponse, output io.Writer) error {
	images := pplxResponse.GetImages()
	if len(images) == 0 {
		return nil
	}
	
	_, err := fmt.Fprintf(output, "\n📸 Images:\n")
	if err != nil {
		return fmt.Errorf("error writing images header to output: %w", err)
	}
	
	for i, img := range images {
		_, err := fmt.Fprintf(output, "[%d]: %s (origin: %s) - %dx%d\n",
			i+1, img.ImageURL, img.OriginURL, img.Width, img.Height)
		if err != nil {
			return fmt.Errorf("error writing image to output: %w", err)
		}
	}
	return nil
}

// RenderRelatedQuestions renders the related questions from the response.
// Currently not implemented as the Perplexity API does not provide related questions functionality.
// This function exists for future compatibility when the API feature becomes available.
func RenderRelatedQuestions(_ *perplexity.CompletionResponse, _ io.Writer) error {
	// NOTE: Related questions functionality is not available in the current Perplexity API.
	// Implementation will be added when the API supports this feature.
	//
	// Expected implementation when available:
	// related := pplxResponse.GetRelatedQuestions()
	// if len(related) == 0 {
	// 	return nil
	// }
	//
	// _, err := fmt.Fprintf(output, "\n❓ Related Questions:\n")
	// if err != nil {
	// 	return fmt.Errorf("error writing related questions header to output: %w", err)
	// }
	//
	// for i, question := range related {
	// 	_, err := fmt.Fprintf(output, "%d. %s\n", i+1, question)
	// 	if err != nil {
	// 		return fmt.Errorf("error writing related question to output: %w", err)
	// 	}
	// }
	return nil
}

// StreamingRenderer handles incremental rendering of streaming content.
type StreamingRenderer struct {
	lastContentLength int
	output           io.Writer
}

// NewStreamingRenderer creates a new streaming renderer.
func NewStreamingRenderer(output io.Writer) *StreamingRenderer {
	return &StreamingRenderer{
		lastContentLength: 0,
		output:           output,
	}
}

// RenderIncremental renders only the new content since last render.
func (sr *StreamingRenderer) RenderIncremental(pplxResponse *perplexity.CompletionResponse) error {
	content := pplxResponse.GetLastContent()
	if content == "" {
		return nil
	}
	
	// Since Perplexity sends cumulative content, we need to track what we've already rendered
	contentLength := len(content)
	if contentLength > sr.lastContentLength {
		// Extract only the new content
		newContent := content[sr.lastContentLength:]
		_, err := fmt.Fprint(sr.output, newContent)
		if err != nil {
			return fmt.Errorf("error writing streaming content to output: %w", err)
		}
		sr.lastContentLength = contentLength
	}
	
	return nil
}

// RenderStreamingContent renders streaming content as it arrives (for backward compatibility).
func RenderStreamingContent(pplxResponse *perplexity.CompletionResponse, output io.Writer) error {
	// This function is kept for backward compatibility but shouldn't be used directly
	// Use StreamingRenderer instead
	content := pplxResponse.GetLastContent()
	if content == "" {
		return nil
	}

	_, err := fmt.Fprint(output, content)
	if err != nil {
		return fmt.Errorf("error writing streaming content to output: %w", err)
	}

	return nil
}

// RenderJSON formats and outputs the response as JSON.
func RenderJSON(pplxResponse *perplexity.CompletionResponse, output io.Writer) error {
	result := buildJSONResponse(pplxResponse)

	// Convert to JSON with indentation for readability
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format JSON output: %w", err)
	}

	_, err = fmt.Fprintln(output, string(jsonData))
	if err != nil {
		return fmt.Errorf("error writing JSON to output: %w", err)
	}

	return nil
}

// buildJSONResponse creates a structured JSON response from the Perplexity API response.
func buildJSONResponse(pplxResponse *perplexity.CompletionResponse) map[string]any {
	result := map[string]any{
		"content": pplxResponse.Choices[0].Message.Content,
		"model":   pplxResponse.Model,
		"usage":   pplxResponse.Usage,
	}

	// Add search results if available
	if pplxResponse.SearchResults != nil && len(*pplxResponse.SearchResults) > 0 {
		result["search_results"] = *pplxResponse.SearchResults
	}

	// Add images if available
	if pplxResponse.Images != nil && len(*pplxResponse.Images) > 0 {
		result["images"] = *pplxResponse.Images
	}

	// Add related questions if available
	if pplxResponse.RelatedQuestions != nil && len(*pplxResponse.RelatedQuestions) > 0 {
		result["related_questions"] = *pplxResponse.RelatedQuestions
	}

	return result
}

// RenderResponse renders the response in the specified format (JSON or console).
// This is a unified rendering function that handles both output formats.
func RenderResponse(pplxResponse *perplexity.CompletionResponse, output io.Writer, asJSON bool) error {
	if asJSON {
		return RenderJSON(pplxResponse, output)
	}

	// Render as console output (markdown + citations + images + related questions)
	if err := RenderAsMarkdown(pplxResponse, output); err != nil {
		return err
	}
	if err := RenderCitations(pplxResponse, output); err != nil {
		return err
	}
	if err := RenderImages(pplxResponse, output); err != nil {
		return err
	}
	if err := RenderRelatedQuestions(pplxResponse, output); err != nil {
		return err
	}

	return nil
}
