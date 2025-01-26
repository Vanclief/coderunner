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

			callback := llmCallback()

			scopeMap.ProcessFilesWithLLM(c.String("prompt"), callback)

			return nil
		},
	}
}

// llmCallback creates a callback function that updates the WeightedFilesMap
func llmCallback() func(string, string) error {
	return func(path string, response string) error {
		const op = "llmCallback"

		outputPath := path + ".llm.md"

		// Write response to file
		if err := os.WriteFile(outputPath, []byte(response), 0644); err != nil {
			return ez.Wrap(op, fmt.Errorf("failed to write response file: %w", err))
		}

		fmt.Printf("Response written to: %s\n", outputPath)

		return nil
	}
}
