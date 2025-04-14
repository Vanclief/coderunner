package scopes

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/vanclief/coderunner/files"
	"github.com/vanclief/ez"
)

// Scope represents a scope with its metadata and content map
type Scope struct {
	Name         string                 `json:"name"`
	BaseCommit   string                 `json:"baseCommit"`
	TargetCommit string                 `json:"targetCommit,omitempty"`
	Files        map[string]interface{} `json:"files"`
}

// NewScope creates a new Scope with the given base commit
func NewScope(name, targetCommit string) *Scope {
	return &Scope{
		Name:         name,
		TargetCommit: targetCommit,
		Files:        make(map[string]interface{}),
	}
}

// SetTargetCommit sets the target commit for this scope
func (s *Scope) SetTargetCommit(commit string) {
	s.TargetCommit = commit
}

// LoadSelectedScope loads the currently selected scope
func LoadSelectedScope() (*Scope, error) {
	const op = "scopes.LoadSelectedScope"

	commitContext, err := LoadCommitContext()
	if err != nil {
		return nil, ez.Wrap(op, err)
	}

	return LoadScope(commitContext.SelectedScope)
}

// LoadScope loads a scope by name
func LoadScope(scopeName string) (*Scope, error) {
	const op = "scopes.LoadScope"

	filePath, err := files.GetScopeFilePath(scopeName)
	if err != nil {
		return nil, ez.Wrap(op, err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		errMsg := fmt.Sprintf("Scope %s doesn't exist", scopeName)
		return nil, ez.New(op, ez.ENOTFOUND, errMsg, err)
	}

	var scope Scope
	if err := json.Unmarshal(data, &scope); err != nil {
		return nil, ez.New(op, ez.EINVALID, "Failed to parse scope file", err)
	}

	return &scope, nil
}

// Save persists the scope to a file
func (s *Scope) Save(outputPath string) error {
	const op = "Scope.Save"

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return ez.New(op, ez.EINTERNAL, "Error marshaling scope", err)
	}

	err = os.WriteFile(outputPath, data, 0644)
	if err != nil {
		return ez.New(op, ez.EINTERNAL, "Error writing scope file", err)
	}

	return nil
}

// AddToMap adds a path to the scope map
func (s *Scope) AddToMap(path string, isFile bool) {
	parts := strings.Split(path, "/")
	addToMap(s.Files, parts, isFile)
}

// addToMap adds path parts to the map
func addToMap(fileMap map[string]interface{}, parts []string, isFile bool) {
	if len(parts) == 0 {
		return
	}

	current := parts[0]

	if len(parts) == 1 && isFile {
		fileMap[current] = true
		return
	}

	if _, exists := fileMap[current]; !exists {
		fileMap[current] = make(map[string]interface{})
	}

	if subMap, ok := fileMap[current].(map[string]interface{}); ok {
		addToMap(subMap, parts[1:], isFile)
	}
}

// PrintTree prints the scope in a tree-like structure
func (s *Scope) PrintTree() {
	fmt.Printf("Scope (Base: %s", s.BaseCommit)
	if s.TargetCommit != "" {
		fmt.Printf(", Target: %s", s.TargetCommit)
	}
	fmt.Println(")")
	printTreeNode(s.Files, "", "")
}

// printTreeNode recursively prints the tree structure
func printTreeNode(node interface{}, prefix string, path string) {
	// Try to convert to map[string]interface{}
	if fileMap, ok := node.(map[string]interface{}); ok {
		// Get sorted keys for consistent output
		keys := make([]string, 0, len(fileMap))
		for k, val := range fileMap {
			// For files (bool values), only include if true
			// For directories (maps), check if they have any true files
			switch v := val.(type) {
			case map[string]interface{}:
				if hasInScopeFiles(v) {
					keys = append(keys, k)
				}
			case bool:
				if v {
					keys = append(keys, k)
				}
			}
		}
		sort.Strings(keys)

		// If this is a non-root path and we have files to show, print the directory
		if path != "" && len(keys) > 0 {
			fmt.Println(getLastComponent(path))
		}

		// Process each child
		for i, key := range keys {
			isLastItem := i == len(keys)-1
			newPrefix := prefix
			newPath := key

			if path != "" {
				newPath = path + "/" + key
				if isLastItem {
					newPrefix = prefix + "    "
				} else {
					newPrefix = prefix + "│   "
				}
			}

			// Add the appropriate prefix character
			if path != "" {
				if isLastItem {
					fmt.Print(prefix + "└── ")
				} else {
					fmt.Print(prefix + "├── ")
				}
			}

			printTreeNode(fileMap[key], newPrefix, newPath)
		}
		return
	}

	// Try to convert to bool (for files)
	if b, ok := node.(bool); ok && b {
		fmt.Println(getLastComponent(path))
		return
	}
}

func getLastComponent(path string) string {
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		return path[idx+1:]
	}
	return path
}

// hasInScopeFiles recursively checks if a directory has any files with value true
func hasInScopeFiles(fileMap map[string]interface{}) bool {
	for _, v := range fileMap {
		switch val := v.(type) {
		case bool:
			if val {
				return true
			}
		case map[string]interface{}:
			if hasInScopeFiles(val) {
				return true
			}
		}
	}
	return false
}

// GetFilesContent reads and returns the content of all files marked as true
func (s *Scope) GetFilesContent() (map[string]string, error) {
	const op = "Scope.GetFilesContent"

	contents := make(map[string]string)
	paths := make([]string, 0)

	// First collect all paths with true value
	collectPaths(s.Files, "", &paths)

	// Then read each file
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, ez.New(op, ez.EINTERNAL, "Failed to read file: "+path, err)
		}

		if files.IsBinaryFile(content) {
			continue // Skip binary files
		}
		contents[path] = string(content)
	}

	return contents, nil
}

// GetAllFilePaths returns all file paths in the scope that are marked as true
func (s *Scope) GetAllFilePaths() []string {
	paths := make([]string, 0)
	collectPaths(s.Files, "", &paths)
	return paths
}

// collectPaths recursively collects all file paths that have value true
func collectPaths(node interface{}, currentPath string, paths *[]string) {
	switch v := node.(type) {
	case bool:
		if v {
			*paths = append(*paths, currentPath)
		}
	case map[string]interface{}:
		for k, val := range v {
			newPath := k
			if currentPath != "" {
				newPath = filepath.Join(currentPath, k)
			}
			collectPaths(val, newPath, paths)
		}
	}
}
