package console

import (
	"bufio"
	"fmt"
	"io"
	"os"

	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/sgaunet/perplexity-go/v2"
)

const DefaultLineLength = 80
const DefaultLeftMargin = 6

func Input(label string) (string, error) {
	fmt.Printf("%s: (set an empty line to validate the entry)\n", label)

	scanner := bufio.NewScanner(os.Stdin)
	var lines string
	for {
		scanner.Scan()
		line := scanner.Text()
		if len(line) == 0 {
			break
		}
		lines = lines + " " + line
	}

	err := scanner.Err()
	return lines, err
}

func RenderAsMarkdown(pplxResponse *perplexity.CompletionResponse, output io.Writer) error {
	result := markdown.Render(pplxResponse.GetLastContent(), DefaultLineLength, DefaultLeftMargin)
	_, err := fmt.Fprintln(output, string(result))
	if err != nil {
		return fmt.Errorf("error writing markdown to output: %w", err)
	}
	return nil
}

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
		_, err := fmt.Fprintf(output, "[%d]: %s (origin: %s) - %dx%d\n", i+1, img.ImageURL, img.OriginURL, img.Width, img.Height)
		if err != nil {
			return fmt.Errorf("error writing image to output: %w", err)
		}
	}
	return nil
}

func RenderRelatedQuestions(pplxResponse *perplexity.CompletionResponse, output io.Writer) error {
	// TODO: Implement when the correct API is available
	// related := pplxResponse.GetRelatedQuestions()
	// if len(related) == 0 {
	// 	return nil
	// }
	
	// _, err := fmt.Fprintf(output, "\n❓ Related Questions:\n")
	// if err != nil {
	// 	return fmt.Errorf("error writing related questions header to output: %w", err)
	// }
	
	// for i, question := range related {
	// 	_, err := fmt.Fprintf(output, "%d. %s\n", i+1, question)
	// 	if err != nil {
	// 		return fmt.Errorf("error writing related question to output: %w", err)
	// 	}
	// }
	return nil
}

// StreamingRenderer handles incremental rendering of streaming content
type StreamingRenderer struct {
	lastContentLength int
	output           io.Writer
}

// NewStreamingRenderer creates a new streaming renderer
func NewStreamingRenderer(output io.Writer) *StreamingRenderer {
	return &StreamingRenderer{
		lastContentLength: 0,
		output:           output,
	}
}

// RenderIncremental renders only the new content since last render
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

// RenderStreamingContent renders streaming content as it arrives (for backward compatibility)
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
