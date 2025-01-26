package git

import (
	"os/exec"
	"strings"

	"github.com/vanclief/ez"
)

const (
	op = "git"
)

// Info holds git repository information
type Info struct {
	IsRepo        bool   // Whether the current directory is in a git repository
	CurrentCommit string // The current commit hash
	CurrentBranch string // The current branch name
}

// GetInfo returns information about the current git repository
func GetInfo() (*Info, error) {
	const op = "git.GetInfo"

	isRepo, err := isGitRepo()
	if err != nil {
		return nil, ez.Wrap(op, err)
	}

	if !isRepo {
		return &Info{IsRepo: false}, nil
	}

	commit, err := getCurrentCommitShort()
	if err != nil {
		return nil, ez.Wrap(op, err)
	}

	branch, err := getCurrentBranch()
	if err != nil {
		return nil, ez.Wrap(op, err)
	}

	return &Info{
		IsRepo:        isRepo,
		CurrentCommit: commit,
		CurrentBranch: branch,
	}, nil
}

// isGitRepo checks if the current directory is within a git repository
func isGitRepo() (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	err := cmd.Run()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return false, nil
		}
		return false, ez.New(op, ez.EINTERNAL, "failed to check git repository", err)
	}
	return true, nil
}

// getCurrentCommit returns the current commit hash
func getCurrentCommit() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", ez.New(op, ez.EINTERNAL, "failed to get current commit", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// getCurrentCommitShort returns the abbreviated commit hash (usually 7 characters)
func getCurrentCommitShort() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", ez.New(op, ez.EINTERNAL, "failed to get short commit hash", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// getCurrentBranch returns the current branch name
func getCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", ez.New(op, ez.EINTERNAL, "failed to get current branch", err)
	}
	return strings.TrimSpace(string(output)), nil
}
