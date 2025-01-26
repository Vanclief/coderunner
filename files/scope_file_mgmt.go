package files

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/vanclief/coderunner/git"
	"github.com/vanclief/ez"
)

func ListScopeFiles() error {
	const op = "files.ListScopes"

	// Check if directory exists
	if _, err := os.Stat(CODERUNNER_DIR); os.IsNotExist(err) {
		return ez.New(op, ez.ENOTFOUND, "No scopes found", err)
	}

	// Get the git info
	gitInfo, err := git.GetInfo()
	if err != nil {
		return ez.Wrap(op, err)
	}

	// Read all files in directory
	files, err := os.ReadDir(CODERUNNER_DIR)
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
			if scopeName != "" && strings.TrimSpace(scopeName) != "" {
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

func CopyFile(sourceFilePath, targetFilePath string) error {
	const op = "files.CopyFile"

	// Check if file exists
	_, err := os.Stat(sourceFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			errMsg := fmt.Sprintf("File %s not found", sourceFilePath)
			return ez.Root(op, ez.ENOTFOUND, errMsg)
		}
		return ez.Wrap(op, err)
	}

	// Check if target file already exists
	_, err = os.Stat(targetFilePath)
	if err == nil {
		errMsg := fmt.Sprintf("Target file '%s' already exists", targetFilePath)
		return ez.Root(op, ez.ENOTFOUND, errMsg)
	}

	// Open source file
	source, err := os.Open(sourceFilePath)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to open source file '%s'", sourceFilePath)
		return ez.Root(op, ez.ENOTFOUND, errMsg)
	}
	defer source.Close()

	// Create target file
	target, err := os.Create(targetFilePath)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to create target file '%s'", targetFilePath)
		return ez.Root(op, ez.ENOTFOUND, errMsg)
	}
	defer target.Close()

	// Copy the contents
	if _, err = io.Copy(target, source); err != nil {
		return ez.New(op, ez.EINTERNAL, "Failed to copy scope contents", err)
	}

	return nil
}

func DeleteFile(filePath string) error {
	const op = "files.DeleteFile"

	// Check if file exists
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			errMsg := fmt.Sprintf("File %s not found", filePath)
			return ez.Root(op, ez.ENOTFOUND, errMsg)
		}
		return ez.Wrap(op, err)
	}

	// Delete the file
	err = os.Remove(filePath)
	if err != nil {
		return ez.Wrap(op, err)
	}

	return nil
}
