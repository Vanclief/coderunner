package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"github.com/vanclief/coderunner/files"
	"github.com/vanclief/coderunner/scanner"
	"github.com/vanclief/ez"
)

func ScopeCmd() *cli.Command {
	return &cli.Command{
		Name:  "scope",
		Usage: "Manage scopes of a codebase",
		Subcommands: []*cli.Command{
			scopeListCmd(),
			scopeCreateCmd(),
			scopeCopyCmd(),
			scopeEditCmd(),
			scopeDeleteCmd(),
			scopeTreeCmd(),
			scopePrintCmd(),
		},
	}
}

func scopeListCmd() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List existing scopes",
		Action: func(c *cli.Context) error {
			return files.ListScopeFiles()
		},
	}
}

func scopeCreateCmd() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create a new scope",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "scope",
				Usage:    "Name of the scope",
				Aliases:  []string{"s", "n"},
				Required: true,
			},
			&cli.StringFlag{
				Name:    "base",
				Usage:   "Starting commit or branch",
				Aliases: []string{"b"},
			},
			&cli.StringFlag{
				Name:    "target",
				Usage:   "Commit or branch to compare to",
				Aliases: []string{"t"},
			},
			&cli.StringSliceFlag{
				Name:    "extensions",
				Usage:   "File extensions to include (e.g., --extensions .go,.js,.ts)",
				Aliases: []string{"e", "ext"},
			},
		},
		Action: func(c *cli.Context) error {
			const op = "cli.scopeCreateCmd"

			s := scanner.New(".", c.StringSlice("extensions"))

			var scopeMap files.ScopeMap
			var err error

			if c.String("base") == "" {
				scopeMap, err = s.ScanAndCreateScope()
				if err != nil {
					return ez.Wrap(op, err)
				}
			} else {
				scopeMap, err = s.ScanGitDiffAndCreateScope(c.String("base"), c.String("target"))
				if err != nil {
					return ez.Wrap(op, err)
				}
			}

			filePath, err := files.GetScopeFilePath(c.String("scope"))
			if err != nil {
				return ez.Wrap(op, err)
			}

			err = scopeMap.Save(filePath)
			if err != nil {
				return ez.Wrap(op, err)
			}

			fmt.Printf("Created scope file %s You can edit the file to change the scope.\n", filePath)

			return nil
		},
	}
}

func scopeCopyCmd() *cli.Command {
	return &cli.Command{
		Name:  "copy",
		Usage: "Copy a scope",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "scope",
				Usage:    "Name of the scope",
				Aliases:  []string{"s", "n"},
				Required: true,
			},
			&cli.StringFlag{
				Name:     "copy",
				Usage:    "Name of the scope copy",
				Aliases:  []string{"o", "c"},
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			const op = "cli.scopeCopyCmd"

			sourceFilePath, err := files.GetScopeFilePath(c.String("scope"))
			if err != nil {
				return ez.Wrap(op, err)
			}

			targetFilePath, err := files.GetScopeFilePath(c.String("copy"))
			if err != nil {
				return ez.Wrap(op, err)
			}

			err = files.CopyFile(sourceFilePath, targetFilePath)
			if err != nil {
				fmt.Println("error", err)
				return ez.Wrap(op, err)
			}

			fmt.Printf("Copied scope file to %s", targetFilePath)

			return nil
		},
	}
}

func scopeEditCmd() *cli.Command {
	return &cli.Command{
		Name:  "edit",
		Usage: "Edit a scope",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "scope",
				Usage:    "Name of the scope",
				Aliases:  []string{"s", "n"},
				Required: true,
			},
			&cli.StringFlag{
				Name:    "editor",
				Usage:   "Editor to open the file",
				Aliases: []string{"e"},
			},
		},
		Action: func(c *cli.Context) error {
			const op = "cli.scopeEditCmd"

			filePath, err := files.GetScopeFilePath(c.String("scope"))
			if err != nil {
				return ez.Wrap(op, err)
			}

			err = files.OpenFile(filePath, c.String("editor"))
			if err != nil {
				return ez.Wrap(op, err)
			}

			return nil
		},
	}
}

func scopeDeleteCmd() *cli.Command {
	return &cli.Command{
		Name:  "delete",
		Usage: "Delete a scope",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "scope",
				Usage:    "Name of the scope",
				Aliases:  []string{"s", "n"},
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			const op = "cli.scopeDeleteCmd"

			filePath, err := files.GetScopeFilePath(c.String("scope"))
			if err != nil {
				return ez.Wrap(op, err)
			}

			err = files.DeleteFile(filePath)
			if err != nil {
				return ez.Wrap(op, err)
			}

			fmt.Printf("Deleted scope file %s", filePath)

			return nil
		},
	}
}

func scopeTreeCmd() *cli.Command {
	return &cli.Command{
		Name:  "tree",
		Usage: "Display the tree of the files in scope",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "scope",
				Usage:    "Name of the scope",
				Aliases:  []string{"s", "n"},
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			const op = "cmd.scopeTreeCmd"

			filePath, err := files.GetScopeFilePath(c.String("scope"))
			if err != nil {
				return ez.Wrap(op, err)
			}

			scopeMap, err := files.LoadScopeMap(filePath)
			if err != nil {
				return ez.Wrap(op, err)
			}

			scopeMap.PrintTree()

			return nil
		},
	}
}

func scopePrintCmd() *cli.Command {
	return &cli.Command{
		Name:  "print",
		Usage: "Print the raw scope",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "scope",
				Usage:    "Name of the scope",
				Aliases:  []string{"s", "n"},
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			const op = "cmd.scopePrintCmd"

			filePath, err := files.GetScopeFilePath(c.String("scope"))
			if err != nil {
				return ez.Wrap(op, err)
			}

			scopeMap, err := files.LoadScopeMap(filePath)
			if err != nil {
				return ez.Wrap(op, err)
			}

			contents, err := scopeMap.GetFilesContent()
			if err != nil {
				return ez.Wrap(op, err)
			}

			// Now contents map has all files
			for path, content := range contents {
				fmt.Printf("File: %s\nContent: %s\n", path, content)
			}

			return nil
		},
	}
}
