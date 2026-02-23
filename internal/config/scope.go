package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pfldy2850/claudectx/internal/claude"
)

// ScopeType represents the scope of a context (user or project).
type ScopeType string

const (
	ScopeUser    ScopeType = "user"
	ScopeProject ScopeType = "project"
)

// ExtraFile represents a standalone file outside .claude/ that is part of a scope
// (e.g. ~/.claude.json for user scope, CLAUDE.md and .mcp.json for project scope).
type ExtraFile struct {
	Path string // absolute path to the file
	Tag  string // source tag: "claudejson", "claudemd", "mcpjson"
}

// Scope defines where Claude config files live and how they are stored.
type Scope struct {
	Type            ScopeType
	DotClaudeDir    string      // source dir to walk
	ExtraFiles      []ExtraFile // standalone files outside .claude/
	StorageDir      string      // where .claudectx data lives
	IncludePatterns []string
	ExcludePatterns []string
}

// ExtraFileByTag returns the ExtraFile matching the given tag, or nil if not found.
func (s *Scope) ExtraFileByTag(tag string) *ExtraFile {
	for i := range s.ExtraFiles {
		if s.ExtraFiles[i].Tag == tag {
			return &s.ExtraFiles[i]
		}
	}
	return nil
}

// UserScope builds a Scope for user-level config (~/.claude/ + ~/.claude.json).
func UserScope() (*Scope, error) {
	dotClaudeDir, err := claude.DotClaudeDir()
	if err != nil {
		return nil, err
	}
	claudeJSONPath, err := claude.ClaudeJSONPath()
	if err != nil {
		return nil, err
	}
	storageDir, err := DefaultStorageDir()
	if err != nil {
		return nil, err
	}
	return &Scope{
		Type:         ScopeUser,
		DotClaudeDir: dotClaudeDir,
		ExtraFiles: []ExtraFile{
			{Path: claudeJSONPath, Tag: "claudejson"},
		},
		StorageDir:      storageDir,
		IncludePatterns: DefaultIncludePatterns,
		ExcludePatterns: DefaultExcludePatterns,
	}, nil
}

// ProjectScopeAt builds a project Scope rooted at the given directory (no detection).
func ProjectScopeAt(root string) *Scope {
	return &Scope{
		Type:         ScopeProject,
		DotClaudeDir: filepath.Join(root, ".claude"),
		ExtraFiles: []ExtraFile{
			{Path: filepath.Join(root, "CLAUDE.md"), Tag: "claudemd"},
			{Path: filepath.Join(root, ".mcp.json"), Tag: "mcpjson"},
		},
		StorageDir:      filepath.Join(root, ".claudectx"),
		IncludePatterns: DefaultProjectIncludePatterns,
		ExcludePatterns: DefaultProjectExcludePatterns,
	}
}

// ProjectScope builds a Scope for project-level config (<git-root>/.claude/ + <git-root>/CLAUDE.md).
func ProjectScope() (*Scope, error) {
	root, err := claude.ProjectRoot()
	if err != nil {
		return nil, err
	}
	return ProjectScopeAt(root), nil
}

// DetectProjectRoot returns the project root using the composite detection chain:
// Claude marker files first, then git root. Returns an error if neither is found.
func DetectProjectRoot() (string, error) {
	if root, err := claude.FindMarkerRoot(); err == nil {
		return root, nil
	}
	return claude.ProjectRoot()
}

// ResolveScopeWithRoot returns a Scope using the composite detection chain.
// rootOverride takes highest priority, then scopeOverride, then auto-detection.
func ResolveScopeWithRoot(scopeOverride, rootOverride string) (*Scope, error) {
	// Validate scopeOverride early
	if scopeOverride != "" && scopeOverride != "user" && scopeOverride != "project" {
		return nil, fmt.Errorf("invalid scope %q: must be 'user' or 'project'", scopeOverride)
	}

	// --root flag: explicit project root
	if rootOverride != "" {
		if scopeOverride == "user" {
			return nil, fmt.Errorf("--root cannot be used with --scope user")
		}
		absRoot, err := filepath.Abs(rootOverride)
		if err != nil {
			return nil, fmt.Errorf("resolve root path: %w", err)
		}
		info, err := os.Stat(absRoot)
		if err != nil {
			return nil, fmt.Errorf("root directory %q does not exist", rootOverride)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("root path %q is not a directory", rootOverride)
		}
		return ProjectScopeAt(absRoot), nil
	}

	switch scopeOverride {
	case "user":
		return UserScope()
	case "project":
		// Explicit project: composite detection with CWD fallback
		if root, err := DetectProjectRoot(); err == nil {
			return ProjectScopeAt(root), nil
		}
		// Fallback to CWD
		cwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("get working directory: %w", err)
		}
		return ProjectScopeAt(cwd), nil
	default: // "" — auto-detect: composite detection, user scope fallback
		if root, err := DetectProjectRoot(); err == nil {
			return ProjectScopeAt(root), nil
		}
		return UserScope()
	}
}

// ResolveScope returns a Scope based on the override flag value.
// If override is empty, auto-detects: project scope if detected, user scope otherwise.
func ResolveScope(override string) (*Scope, error) {
	return ResolveScopeWithRoot(override, "")
}
