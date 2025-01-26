package files

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/vanclief/ez"
)

func EnsureDirectoryExists(dirName string) error {
	// Get the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %v", err)
	}

	// Construct the full path for the output directory
	dirPath := filepath.Join(currentDir, dirName)

	// Check if the directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		// Directory doesn't exist, so create it
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
	}

	return nil
}

// OpenFile opens the specified file with the system's default text editor
func OpenFile(filePath, editor string) error {
	const op = "files.OpenFile"

	// Check if file exists
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			errMsg := fmt.Sprintf("File %s not found", filePath)
			return ez.Root(op, ez.ENOTFOUND, errMsg)
		}
		return ez.Wrap(op, err)
	}

	// If a specific editor is specified, use it
	if editor == "vim" || editor == "nano" {
		// Check if the editor is available in the system
		editorPath, err := exec.LookPath(editor)
		if err != nil {
			errMsg := fmt.Sprintf("%s is not installed in the system", editor)
			return ez.New(op, ez.EUNAVAILABLE, errMsg, err)
		}

		// Create the command
		cmd := exec.Command(editorPath, filePath)

		// Set the command to run in the current terminal
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Execute the command
		err = cmd.Run()
		if err != nil {
			errMsg := fmt.Sprintf("Failed to open file %s with %s", filePath, editor)
			return ez.New(op, ez.EINTERNAL, errMsg, err)
		}

		return nil
	}

	// Determine the command to use based on the operating system
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin": // macOS
		cmd = exec.Command("open", filePath)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", filePath)
	case "linux":
		cmd = exec.Command("xdg-open", filePath)
	default:
		errMsg := fmt.Sprintf("Unsupported operating system: %s", runtime.GOOS)
		return ez.Root(op, ez.ENOTIMPLEMENTED, errMsg)
	}

	// Execute the command
	err = cmd.Run()
	if err != nil {
		errMsg := fmt.Sprintf("Failed to open file %s", filePath)
		return ez.New(op, ez.EINTERNAL, errMsg, err)
	}

	return nil
}

// IsBinaryFile uses Git's approach to detect binary files
// It looks for null bytes in the first 8000 bytes of the file
func IsBinaryFile(content []byte) bool {
	// Empty files are not binary
	if len(content) == 0 {
		return false
	}

	// Check up to first 8000 bytes like Git does
	size := 8000
	if len(content) < size {
		size = len(content)
	}

	// Look for null byte
	return bytes.IndexByte(content[:size], 0) != -1
}
