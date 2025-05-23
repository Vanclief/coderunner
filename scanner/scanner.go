package scanner

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vanclief/coderunner/git"
	"github.com/vanclief/coderunner/scopes"
	"github.com/vanclief/ez"
)

type Scanner struct {
	rootDir           string
	ignoreRules       []string
	allowedExtensions map[string]bool
	debug             bool
}

func New(rootDir string, allowedExtensions []string) *Scanner {
	s := &Scanner{
		rootDir:           rootDir,
		ignoreRules:       make([]string, 0, len(defaultIgnorePatterns)),
		allowedExtensions: make(map[string]bool),
	}

	// Add default ignore patterns
	s.ignoreRules = append(s.ignoreRules, defaultIgnorePatterns...)

	// Convert allowed extensions to a map
	for _, ext := range allowedExtensions {
		// Ensure extension starts with a dot
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		s.allowedExtensions[ext] = true
	}

	return s
}

func (s *Scanner) SetDebug(debug bool) {
	s.debug = debug
}

func (s *Scanner) ScanAndCreateScope(scopeName string) (*scopes.Scope, error) {
	const op = "Scanner.ScanAndCreateScope"

	// Get git info to use as base commit
	gitInfo, err := git.GetInfo()
	if err != nil {
		return nil, ez.Wrap(op, err)
	}

	// Create a new scope with the current commit
	scope := scopes.NewScope(scopeName, gitInfo.CurrentCommit)

	if err := s.loadGitIgnore(); err != nil {
		return nil, ez.Wrap(op, err)
	}

	err = filepath.Walk(s.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return ez.New(op, ez.EINTERNAL, "Error accessing path", err)
		}

		// Skip root directory
		if path == s.rootDir {
			return nil
		}

		// Check if path should be ignored
		if s.shouldIgnore(path) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Get path relative to root
		relPath, err := filepath.Rel(s.rootDir, path)
		if err != nil {
			return ez.New(op, ez.EINTERNAL, "Error getting relative path", err)
		}

		// Use forward slashes for consistency
		relPath = filepath.ToSlash(relPath)

		// Add to scope
		scope.AddToMap(relPath, !info.IsDir())

		return nil
	})
	if err != nil {
		return nil, ez.Wrap(op, err)
	}

	return scope, nil
}

// ScanGitDiffAndCreateScope scans only the scopes that changed in the git diff
// from and to can be either commit hashes or branch names
// if to is empty, it will show changes in the working directory
func (s *Scanner) ScanGitDiffAndCreateScope(scopeName, baseCommit string) (*scopes.Scope, error) {
	const op = "Scanner.ScanGitDiffAndCreateScope"

	gitInfo, err := git.GetInfo()
	if err != nil {
		return nil, ez.Wrap(op, err)
	}

	// Get changed scopes from git diff
	changedFiles, err := s.getGitDiffFiles(baseCommit, gitInfo.CurrentCommit)
	if err != nil {
		return nil, ez.Wrap(op, err)
	}

	if err := s.loadGitIgnore(); err != nil {
		return nil, ez.Wrap(op, err)
	}

	// Create a new scope with the target commit
	scope := scopes.NewScope(scopeName, gitInfo.CurrentCommit)
	scope.BaseCommit = baseCommit

	// Process each changed file
	for _, file := range changedFiles {

		// Skip empty lines
		if file == "" {
			continue
		}

		// Use forward slashes for consistency
		file = strings.TrimSpace(file)
		file = strings.ReplaceAll(file, "\\", "/")

		// Check if file should be ignored
		if s.shouldIgnore(file) {
			continue
		}

		scope.AddToMap(file, true)
	}

	return scope, nil
}

// loadGitIgnore loads and parses .gitignore scopes
func (s *Scanner) loadGitIgnore() error {
	const op = "Scanner.loadGitIgnore"

	gitignorePath := filepath.Join(s.rootDir, ".gitignore")
	file, err := os.Open(gitignorePath)
	if os.IsNotExist(err) {
		return nil // No .gitignore file is fine
	}
	if err != nil {
		return ez.New(op, ez.EINTERNAL, "Failed to open .gitignore", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			// Only remove trailing comments if # is preceded by whitespace
			// This prevents treating # in patterns like ".#*" as comments
			if idx := strings.Index(line, " #"); idx >= 0 {
				line = strings.TrimSpace(line[:idx])
			} else if idx := strings.Index(line, "\t#"); idx >= 0 {
				line = strings.TrimSpace(line[:idx])
			}

			if line != "" {
				s.ignoreRules = append(s.ignoreRules, line)
			}
		}
	}

	return nil
}

// shouldIgnore checks if a path should be ignored based on .gitignore rules
// and allowed extensions
func (s *Scanner) shouldIgnore(path string) bool {
	// Get the base name of the file/directory
	base := filepath.Base(path)

	// Check against exact base name matches first (for scopes like .DS_Store)
	for _, pattern := range s.ignoreRules {
		if pattern == base {
			return true
		}
	}

	// Get relative path
	relPath, err := filepath.Rel(s.rootDir, path)
	if err != nil {
		return false
	}
	relPath = filepath.ToSlash(relPath)

	// Check if it's a file (not a directory)
	info, err := os.Stat(path)
	if err != nil {
		return true // If we can't stat the file, ignore it
	}

	if !info.IsDir() {
		// For scopes, check extension only if we have a whitelist
		if len(s.allowedExtensions) > 0 {
			ext := filepath.Ext(path)
			if !s.allowedExtensions[ext] {
				return true // Ignore scopes with non-allowed extensions
			}
		}
	}

	// Apply gitignore rules
	for _, rule := range s.ignoreRules {
		if rule == "" {
			continue
		}

		// Handle basic glob patterns
		rule = strings.TrimPrefix(rule, "*/")
		if strings.HasSuffix(rule, "/*") {
			// Check if the directory matches
			rule = rule[:len(rule)-2]
			if strings.HasPrefix(relPath, rule) {
				return true
			}
			continue
		}

		// Remove leading ./
		rule = strings.TrimPrefix(rule, "./")

		// Exact match
		if rule == relPath {
			return true
		}

		// Handle file extensions (e.g., *.go)
		if strings.HasPrefix(rule, "*.") {
			ext := rule[1:] // Include the dot
			if strings.HasSuffix(relPath, ext) || strings.HasSuffix(base, ext) {
				return true
			}
			continue
		}

		// Handle directory matches
		if strings.HasSuffix(rule, "/") {
			rule = rule[:len(rule)-1]
			if strings.HasPrefix(relPath, rule) {
				return true
			}
			continue
		}

		// Simple wildcard handling
		if strings.Contains(rule, "*") {
			parts := strings.Split(rule, "*")
			matches := true
			remainingPath := relPath

			for _, part := range parts {
				if part == "" {
					continue
				}
				idx := strings.Index(remainingPath, part)
				if idx == -1 {
					matches = false
					break
				}
				remainingPath = remainingPath[idx+len(part):]
			}
			if matches {
				return true
			}
		}

		// Direct path match
		if strings.HasPrefix(relPath, rule) {
			return true
		}
	}

	return false
}

// AddAllowedExtension adds a new allowed extension at runtime
func (s *Scanner) AddAllowedExtension(ext string) {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	s.allowedExtensions[ext] = true
}

// RemoveAllowedExtension removes an extension from the allowed list
func (s *Scanner) RemoveAllowedExtension(ext string) {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	delete(s.allowedExtensions, ext)
}

// getGitDiffFiles returns a list of scopes that were changed in the git diff
func (s *Scanner) getGitDiffFiles(from, to string) ([]string, error) {
	const op = "Scanner.getGitDiffFiles"

	var cmd *exec.Cmd

	if to == "" {
		// If no 'to' is specified, show changes in working directory
		// Include added, deleted, and modified files
		cmd = exec.Command("git", "diff", "--name-only", "--diff-filter=ADM", from)
	} else {
		cmd = exec.Command("git", "diff", "--name-only", "--diff-filter=ADM", from+".."+to)
	}

	if s.debug {
		fmt.Println("Running git diff command:", cmd.String())
	}

	cmd.Dir = s.rootDir

	output, err := cmd.Output()
	if err != nil {
		return nil, ez.New(op, ez.EINTERNAL, "Failed to get git diff", err)
	}

	scopes := strings.Split(strings.TrimSpace(string(output)), "\n")
	return scopes, nil
}
