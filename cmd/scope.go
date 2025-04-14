package cmd

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"github.com/vanclief/coderunner/files"
	"github.com/vanclief/coderunner/scanner"
	"github.com/vanclief/coderunner/scopes"
	"github.com/vanclief/ez"
)

func ScopeCmd() *cli.Command {
	return &cli.Command{
		Name:  "scope",
		Usage: "Manage scopes of a codebase",
		Subcommands: []*cli.Command{
			scopeListCmd(),
			scopeCreateCmd(),
			scopeSelectedCmd(),
			scopeSelectCmd(),
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
			return scopes.List()
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
				Usage:   "Commit or branch to compare to use as base",
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

			var scope *scopes.Scope
			var err error

			if c.String("base") != "" {
				scope, err = s.ScanGitDiffAndCreateScope(c.String("scope"), c.String("base"))
				if err != nil {
					return ez.Wrap(op, err)
				}
			} else {
				scope, err = s.ScanAndCreateScope(c.String("scope"))
				if err != nil {
					return ez.Wrap(op, err)
				}
			}

			if c.String("scope") == "context" {
				return ez.New(op, ez.EINVALID, "Scope name 'context' is reserved", nil)
			}

			filePath, err := files.GetScopeFilePath(scope.Name)
			if err != nil {
				return ez.Wrap(op, err)
			}

			err = scope.Save(filePath)
			if err != nil {
				return ez.Wrap(op, err)
			}

			selectedScope := scopes.NewCommitContext(scope.Name)
			err = selectedScope.Save()
			if err != nil {
				return ez.Wrap(op, err)
			}

			fmt.Printf("Created scope file %s You can edit the file to change the scope.\n", filePath)

			return nil
		},
	}
}

func scopeSelectedCmd() *cli.Command {
	return &cli.Command{
		Name:  "selected",
		Usage: "Show the selected scope",
		Action: func(c *cli.Context) error {
			const op = "cli.scopeCopyCmd"

			commitContext, err := scopes.LoadCommitContext()
			if err != nil {
				fmt.Printf("No selected scope")
			} else {
				fmt.Printf("%s", commitContext.SelectedScope)
			}

			return nil
		},
	}
}

func scopeSelectCmd() *cli.Command {
	return &cli.Command{
		Name:  "select",
		Usage: "Select a scope",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "scope",
				Usage:    "Name of the scope",
				Aliases:  []string{"s", "n"},
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			const op = "cli.scopeCopyCmd"

			if c.String("scope") == "selected" {
				return ez.New(op, ez.EINVALID, "Scope name 'selected' is reserved", nil)
			}

			_, err := scopes.LoadScope(c.String("scope"))
			if err != nil {
				return ez.Wrap(op, err)
			}

			selectedScope := scopes.NewCommitContext(c.String("scope"))
			err = selectedScope.Save()
			if err != nil {
				return ez.Wrap(op, err)
			}

			fmt.Printf("Selected scope %s", c.String("scope"))

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
				Name:     "copy",
				Usage:    "Name of the scope copy",
				Aliases:  []string{"o", "c"},
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			const op = "cli.scopeCopyCmd"

			if c.String("copy") == "selected" {
				return ez.New(op, ez.EINVALID, "Scope name 'selected' is reserved", nil)
			}

			commitContext, err := scopes.LoadCommitContext()
			if err != nil {
				return ez.Wrap(op, err)
			}

			sourceFilePath, err := files.GetScopeFilePath(commitContext.SelectedScope)
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
				Name:    "editor",
				Usage:   "Editor to open the file",
				Aliases: []string{"e"},
			},
		},
		Action: func(c *cli.Context) error {
			const op = "cli.scopeEditCmd"

			commitContext, err := scopes.LoadCommitContext()
			if err != nil {
				return ez.Wrap(op, err)
			}

			filePath, err := files.GetScopeFilePath(commitContext.SelectedScope)
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

			commitContext, err := scopes.LoadCommitContext()
			if err != nil {
				return ez.Wrap(op, err)
			}

			if commitContext.SelectedScope == c.String("context") {
				return ez.New(op, ez.EINVALID, "Cannot delete context scope", nil)
			}

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
		Action: func(c *cli.Context) error {
			const op = "cmd.scopeTreeCmd"

			selectedScope, err := scopes.LoadSelectedScope()
			if err != nil {
				return ez.Wrap(op, err)
			}

			selectedScope.PrintTree()

			return nil
		},
	}
}

func scopePrintCmd() *cli.Command {
	return &cli.Command{
		Name:  "print",
		Usage: "Print the raw scope",
		Action: func(c *cli.Context) error {
			const op = "cmd.scopePrintCmd"

			selectedScope, err := scopes.LoadSelectedScope()
			if err != nil {
				return ez.Wrap(op, err)
			}

			contents, err := selectedScope.GetFilesContent()
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
