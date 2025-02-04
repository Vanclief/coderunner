package main

import (
	"bufio"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"github.com/vanclief/coderunner/cmd"
	"github.com/vanclief/coderunner/files"
	"github.com/vanclief/ez"
)

func main() {
	app := cli.NewApp()
	app.Name = "coderunner"
	app.Usage = "A context-aware code extraction tool to run LLMs in your codebase."
	app.Version = "0.1.2"

	app.Commands = []*cli.Command{
		cmd.ScopeCmd(),
		cmd.LLMCmd(),
	}

	err := Init()
	if err != nil {
		color.Red(ez.ErrorMessage(err))
		return
	}

	err = app.Run(os.Args)
	if err != nil {
		errorCode := ez.ErrorCode(err)
		if errorCode == ez.EINTERNAL {
			color.Red(err.Error())
		} else {
			color.Red(ez.ErrorMessage(err))
		}
	}
}

func Init() error {
	const op = "coderunner.Init"

	err := AddDirToGitignore(files.CODERUNNER_DIR)
	if err != nil {
		return ez.Wrap(op, err)
	}

	err = files.EnsureDirectoryExists(files.CODERUNNER_DIR)
	if err != nil {
		return ez.Wrap(op, err)
	}

	return nil
}

func AddDirToGitignore(dir string) error {
	const op = "coderunner.AddDirToGitignore"

	if strings.TrimSpace(dir) == "" {
		return ez.New(op, ez.EINVALID, "directory name cannot be empty", nil)
	}

	// Check if .gitignore exists
	file, err := os.OpenFile(".gitignore", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return ez.New(op, ez.EINTERNAL, "failed to open .gitignore file", err)
	}
	defer file.Close()

	// Read the content
	scanner := bufio.NewScanner(file)
	existingContent := make([]string, 0)
	hasDirEntry := false

	for scanner.Scan() {
		line := scanner.Text()
		existingContent = append(existingContent, line)
		if strings.TrimSpace(line) == dir {
			hasDirEntry = true
		}
	}

	if err := scanner.Err(); err != nil {
		return ez.New(op, ez.EINTERNAL, "failed to read .gitignore content", err)
	}

	// If directory is not in the file, add it
	if !hasDirEntry {
		// Ensure there's a newline before adding directory if file is not empty
		if len(existingContent) > 0 && existingContent[len(existingContent)-1] != "" {
			existingContent = append(existingContent, "")
		}
		existingContent = append(existingContent, dir)

		// Truncate the file and write the updated content
		if err := file.Truncate(0); err != nil {
			return ez.New(op, ez.EINTERNAL, "failed to truncate .gitignore file", err)
		}

		if _, err := file.Seek(0, 0); err != nil {
			return ez.New(op, ez.EINTERNAL, "failed to seek to beginning of file", err)
		}

		content := strings.Join(existingContent, "\n")
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}

		if _, err := file.WriteString(content); err != nil {
			return ez.New(op, ez.EINTERNAL, "failed to write to .gitignore file", err)
		}
	}

	return nil
}
