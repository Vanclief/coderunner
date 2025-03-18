package scopes

import (
	"fmt"
	"os"
	"strings"

	"github.com/vanclief/coderunner/files"
	"github.com/vanclief/coderunner/git"
	"github.com/vanclief/ez"
)

func List() error {
	const op = "scopes.List"

	// Check if directory exists
	if _, err := os.Stat(files.CODERUNNER_DIR); os.IsNotExist(err) {
		return ez.New(op, ez.ENOTFOUND, "No scopes found", err)
	}

	// Get the git info
	gitInfo, err := git.GetInfo()
	if err != nil {
		return ez.Wrap(op, err)
	}

	// Read all files in directory
	files, err := os.ReadDir(files.CODERUNNER_DIR)
	if err != nil {
		return ez.New(op, ez.EINTERNAL, "Error reading coderunner directory", err)
	}

	foundScopes := false
	for _, file := range files {
		fileName := file.Name()
		// Only process non-directory files that have .scope extension
		if !file.IsDir() && strings.HasSuffix(fileName, ".json") && strings.HasPrefix(fileName, gitInfo.CurrentCommit) {
			// Get name without .scope extension
			scopeName := strings.TrimPrefix(fileName, gitInfo.CurrentCommit+".")
			scopeName = strings.TrimSuffix(scopeName, ".json")
			// Skip if the name is empty or contains only whitespace
			if scopeName != "" && strings.TrimSpace(scopeName) != "" && scopeName != "context" {
				fmt.Println(scopeName)
				foundScopes = true
			}
		}
	}

	if !foundScopes {
		fmt.Println("No scopes for this commit found")
	}

	return nil
}
