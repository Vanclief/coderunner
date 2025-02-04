package cmd

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/vanclief/coderunner/files"
	"github.com/vanclief/ez"
)

func LLMCmd() *cli.Command {
	return &cli.Command{
		Name:  "llm",
		Usage: "Call a llm on each file of scope",
		Subcommands: []*cli.Command{
			promptCmd(),
		},
	}
}

func promptCmd() *cli.Command {
	return &cli.Command{
		Name:  "prompt",
		Usage: "Run a prompt on each file of a scope",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "scope",
				Usage:    "Name of the scope",
				Aliases:  []string{"s", "n"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     "prompt",
				Usage:    "The actual prompt",
				Aliases:  []string{"p"},
				Required: true,
			},
			&cli.StringFlag{
				Name:    "model",
				Usage:   "The model to use (o1, o1-mini, 4o, sonnet)",
				Aliases: []string{"m"},
				Value:   "sonnet",
			},
			&cli.BoolFlag{
				Name:  "save",
				Usage: "Should the model save the response next to the file",
			},
		},
		Action: func(c *cli.Context) error {
			const op = "cli.promptCmd"

			scopeFilePath, err := files.GetScopeFilePath(c.String("scope"))
			if err != nil {
				return ez.Wrap(op, err)
			}

			scopeMap, err := files.LoadScopeMap(scopeFilePath)
			if err != nil {
				return ez.Wrap(op, err)
			}

			callback := llmCallback(c.Bool("save"))

			api, err := files.NewLLM(c.String("model"))
			if err != nil {
				return ez.Wrap(op, err)
			}

			scopeMap.RunPromptOnFiles(api, c.String("prompt"), callback)

			return nil
		},
	}
}

// llmCallback creates a callback function
func llmCallback(save bool) func(string, string) error {
	return func(path string, response string) error {
		const op = "llmCallback"

		if save {
			outputPath := path + ".llm.md"

			// Write response to file
			if err := os.WriteFile(outputPath, []byte(response), 0644); err != nil {
				return ez.Wrap(op, fmt.Errorf("failed to write response file: %w", err))
			}

			fmt.Println("Response written to: %s\n", outputPath)
		} else {
			fmt.Println(response)
		}

		return nil
	}
}
