package scopes

import (
	"encoding/json"
	"os"

	"github.com/vanclief/coderunner/files"
	"github.com/vanclief/ez"
)

// CommitContext holds information about the current commit we are working with
// including which scope is currently selected
type CommitContext struct {
	SelectedScope string `json:"selectedScope"`
}

// NewCommitContext creates a new CommitContext instance
func NewCommitContext(scope string) *CommitContext {
	return &CommitContext{
		SelectedScope: scope,
	}
}

// LoadCommitContext reads the commit context  from the current commit
func LoadCommitContext() (*CommitContext, error) {
	const op = "scopes.LoadSelectedScope"

	filePath, err := files.GetCommitFilePath("context", "json")
	if err != nil {
		return nil, ez.New(op, ez.EINTERNAL, "Error getting commit context file path", err)
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, ez.New(op, ez.ENOTFOUND, "No commit context, create a new scope or select an existing one", err)
	}

	// Read the file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, ez.New(op, ez.EINTERNAL, "Error reading commit context file", err)
	}

	// Create a new commit context
	commitContext := &CommitContext{}

	// Parse the JSON
	err = json.Unmarshal(data, commitContext)
	if err != nil {
		return nil, ez.New(op, ez.EINTERNAL, "Error parsing commit context file", err)
	}

	return commitContext, nil
}

// Save writes the currently selected scope to the specified file
func (s *CommitContext) Save() error {
	const op = "CommitContext.Save"

	filePath, err := files.GetCommitFilePath("context", "json")
	if err != nil {
		return ez.New(op, ez.EINTERNAL, "Error getting selected scope file path", err)
	}

	// Convert the structure to JSON
	data, err := json.Marshal(s)
	if err != nil {
		return ez.New(op, ez.EINTERNAL, "Error marshaling scope", err)
	}

	// Write to file
	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return ez.New(op, ez.EINTERNAL, "Error writing selected scope to file", err)
	}

	return nil
}
