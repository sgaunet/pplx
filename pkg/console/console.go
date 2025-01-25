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
	for i, c := range pplxResponse.GetCitations() {
		_, err := fmt.Fprintf(output, "[%d]: %s\n", i, c)
		if err != nil {
			return fmt.Errorf("error writing citations as markdown to output: %w", err)
		}
	}
	return nil
}
