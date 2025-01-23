package cmd

import (
	"fmt"
	"os"

	"github.com/pterm/pterm"
	"github.com/sgaunet/pplx/pkg/chat"
	"github.com/sgaunet/pplx/pkg/console"
	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
	Use:   "chat",
	Short: "chat subcommand is an interactive chat with the Perplexity API",
	Long: `With chat subcommand you can interactively chat with the Perplexity API.
You can ask questions and get answers from the API. As long as you don't enter an empty question,
 the chat will continue.`,
	Run: func(cmd *cobra.Command, args []string) {
		systemMessage, err := console.Input("system message (optional - enter to skip)")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error reading system message: %v\n", err)
			os.Exit(1)
		}
		c := chat.NewChat(systemMessage)

		// discussion loop
	loop:
		for {
			prompt, err := console.Input("Ask anything (enter to quit)")
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading prompt: %v\n", err)
				os.Exit(1)
			}
			if prompt == "" {
				break loop
			}
			err = c.AddUserMessage(prompt)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error adding user message: %v\n", err)
				os.Exit(1)
			}
			// Print spinner while waiting for the response
			spinnerInfo, _ := pterm.DefaultSpinner.Start("Waiting after the response from perplexity...")
			response, err := c.Run()
			if err != nil {
				fmt.Fprintf(os.Stderr, "error running chat: %v\n", err)
				os.Exit(1)
			}
			spinnerInfo.Success("Response received")

			err = c.AddAgentMessage(response.GetLastContent())
			if err != nil {
				fmt.Fprintf(os.Stderr, "error adding agent message: %v\n", err)
				os.Exit(1)
			}
			err = console.RenderAsMarkdown(response, os.Stdout)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			err = console.RenderCitations(response, os.Stdout)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		}
	},
}
