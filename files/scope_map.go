package files

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/vanclief/ez"
)

type ScopeMap map[string]interface{}

// LoadScopeMap loads the file map with the scope configuration from a JSON file
func LoadScopeMap(path string) (ScopeMap, error) {
	const op = "files.LoadScopeMap"

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, ez.New(op, ez.ENOTFOUND, "Failed to read scope file", err)
	}

	var scopeMap ScopeMap
	if err := json.Unmarshal(data, &scopeMap); err != nil {
		return nil, ez.New(op, ez.EINVALID, "Failed to parse scope file", err)
	}

	return scopeMap, nil
}

func (sm ScopeMap) Save(outputPath string) error {
	const op = "ScopeMap.Save"

	data, err := json.MarshalIndent(sm, "", "  ")
	if err != nil {
		return ez.New(op, ez.EINTERNAL, "Error marshaling scope", err)
	}

	err = os.WriteFile(outputPath, data, 0644)
	if err != nil {
		return ez.New(op, ez.EINTERNAL, "Error writing scope file", err)
	}

	return nil
}

func (sm ScopeMap) addToMap(parts []string, isFile bool) {
	if len(parts) == 0 {
		return
	}

	current := parts[0]

	if len(parts) == 1 && isFile {
		sm[current] = true
		return
	}

	if _, exists := sm[current]; !exists {
		sm[current] = make(ScopeMap)
	}

	if subMap, ok := sm[current].(ScopeMap); ok {
		subMap.addToMap(parts[1:], isFile)
	}
}

// PrintTree prints the ScopeMap in a tree-like structure, showing only files in scope
func (sm ScopeMap) PrintTree() {
	printTreeNode(sm, "", "")
}

// printTreeNode recursively prints the tree structure
func printTreeNode(node interface{}, prefix string, path string) {
	// First try to convert to ScopeMap
	if sm, ok := node.(ScopeMap); ok {
		// Get sorted keys for consistent output
		keys := make([]string, 0, len(sm))
		for k, val := range sm {
			// For files (bool values), only include if true
			// For directories (ScopeMap), check if they have any true files
			switch v := val.(type) {
			case ScopeMap:
				if hasInScopeFiles(v) {
					keys = append(keys, k)
				}
			case map[string]interface{}:
				// Convert to ScopeMap and check
				subMap := ScopeMap(v)
				if hasInScopeFiles(subMap) {
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

			printTreeNode(sm[key], newPrefix, newPath)
		}
		return
	}

	// Try to convert to bool (for files)
	if b, ok := node.(bool); ok && b {
		fmt.Println(getLastComponent(path))
		return
	}

	// Try to convert map[string]interface{} to ScopeMap
	if m, ok := node.(map[string]interface{}); ok {
		printTreeNode(ScopeMap(m), prefix, path)
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
func hasInScopeFiles(sm ScopeMap) bool {
	for _, v := range sm {
		switch val := v.(type) {
		case bool:
			if val {
				return true
			}
		case ScopeMap:
			if hasInScopeFiles(val) {
				return true
			}
		case map[string]interface{}:
			if hasInScopeFiles(ScopeMap(val)) {
				return true
			}
		}
	}
	return false
}

// GetFilesContent reads and returns the content of all files marked as true
func (sm ScopeMap) GetFilesContent() (map[string]string, error) {
	const op = "Scanner.GetFilesContent"

	contents := make(map[string]string)
	paths := make([]string, 0)

	// First collect all paths with true value
	collectPaths(sm, "", &paths)

	// Then read each file
	for _, path := range paths {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, ez.New(op, ez.EINTERNAL, "Failed to read file: "+path, err)
		}

		if IsBinaryFile(content) {
			continue // Skip binary files
		}
		contents[path] = string(content)
	}

	return contents, nil
}

// collectPaths recursively collects all file paths that have value true
func collectPaths(node interface{}, currentPath string, paths *[]string) {
	switch v := node.(type) {
	case bool:
		if v {
			*paths = append(*paths, currentPath)
		}
	case ScopeMap:
		for k, val := range v {
			newPath := k
			if currentPath != "" {
				newPath = filepath.Join(currentPath, k)
			}
			collectPaths(val, newPath, paths)
		}
	case map[string]interface{}:
		collectPaths(ScopeMap(v), currentPath, paths)
	}
}
