package claude

import (
	"fmt"
	"os"
	"path/filepath"
)

// DotClaudeDir returns the path to ~/.claude/ directory.
func DotClaudeDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude"), nil
}

// ClaudeJSONPath returns the path to ~/.claude.json file.
func ClaudeJSONPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude.json"), nil
}

// FindMarkerRootFrom walks up from the given directory looking for Claude
// marker files (.claude/ dir, CLAUDE.md file, or .claudectx/ dir). Returns the
// nearest ancestor containing any marker, or an error if none found.
func FindMarkerRootFrom(start string) (string, error) {
	dir := start
	for {
		// Check for .claude/ directory
		if info, err := os.Stat(filepath.Join(dir, ".claude")); err == nil && info.IsDir() {
			return dir, nil
		}
		// Check for CLAUDE.md file
		if info, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); err == nil && !info.IsDir() {
			return dir, nil
		}
		// Check for .claudectx/ directory
		if info, err := os.Stat(filepath.Join(dir, ".claudectx")); err == nil && info.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no Claude marker files found (.claude/, CLAUDE.md, or .claudectx/)")
		}
		dir = parent
	}
}

// FindMarkerRoot walks up from the current working directory looking for Claude
// marker files. Convenience wrapper around FindMarkerRootFrom.
func FindMarkerRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return FindMarkerRootFrom(dir)
}

// IsGitRepo returns true if the given directory contains a .git/ directory.
// Note: git worktrees use a .git file (not a directory); this intentionally
// checks only for a directory to match standard repository roots.
func IsGitRepo(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil && info.IsDir()
}

// ProjectRoot walks up from the current working directory looking for a .git
// directory. Returns the directory containing .git, or an error if not found.
func ProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if info, err := os.Stat(filepath.Join(dir, ".git")); err == nil && info.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not inside a git repository")
		}
		dir = parent
	}
}
