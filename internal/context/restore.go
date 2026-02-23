package context

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pfldy2850/claudectx/internal/config"
	"github.com/pfldy2850/claudectx/internal/fileutil"
)

// RestoreOptions configures the restore (switch) operation.
type RestoreOptions struct {
	Name    string
	DryRun  bool
	Force   bool
	Verbose bool
	Config  *config.Config
}

// RestoreResult holds the result of a restore operation.
type RestoreResult struct {
	Name          string
	FilesRestored int
	BackupDir     string
}

// Restore applies a saved context to the current Claude Code state.
// If another context is currently active, it auto-saves the live state
// back to that context's snapshot before switching.
func Restore(opts RestoreOptions) (*RestoreResult, error) {
	slug := Slugify(opts.Name)
	cfg := opts.Config
	scope := cfg.Scope

	contextDir := filepath.Join(cfg.ContextsDir(), slug)
	manifest, err := ReadManifest(contextDir)
	if err != nil {
		return nil, fmt.Errorf("context %q not found: %w", slug, err)
	}

	// Validate manifest scope matches current scope
	if manifest.Scope != "" && manifest.Scope != string(scope.Type) {
		return nil, fmt.Errorf("context %q was saved with %s scope, but current scope is %s", slug, manifest.Scope, scope.Type)
	}

	if opts.DryRun {
		return &RestoreResult{
			Name:          slug,
			FilesRestored: len(manifest.Files),
		}, nil
	}

	// 1. Auto-save current context before switching
	AutoSaveCurrent(cfg, slug)

	// 2. Create backup of current state before switching
	backupDir, err := createBackup(cfg, scope)
	if err != nil && !opts.Force {
		return nil, fmt.Errorf("backup failed: %w (use --force to skip)", err)
	}

	// 3. Clear managed files so stale files from previous context don't linger
	if err := ClearManagedFiles(cfg); err != nil {
		return nil, fmt.Errorf("clear before restore: %w", err)
	}

	// 4. Restore files (copy from snapshot to live paths)
	restored, err := restoreCopy(contextDir, scope, manifest)
	if err != nil {
		return nil, err
	}

	// 5. Update current marker
	if err := SetCurrent(cfg, slug); err != nil {
		return nil, err
	}

	return &RestoreResult{
		Name:          slug,
		FilesRestored: restored,
		BackupDir:     backupDir,
	}, nil
}

// AutoSaveCurrent saves the current live state back to the active context's
// snapshot before switching away. targetSlug is excluded to avoid saving over
// the context we're about to switch to. A warning is printed to stderr if the
// save fails, but the switch is not aborted.
func AutoSaveCurrent(cfg *config.Config, targetSlug string) {
	current, err := GetCurrent(cfg)
	if err != nil || current == "" || current == targetSlug {
		return
	}
	if !ContextExists(cfg.ContextsDir(), current) {
		return
	}
	// Save live state back to current context (overwrite)
	if _, err := Save(SaveOptions{
		Name:      current,
		Overwrite: true,
		Config:    cfg,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: auto-save of context %q failed: %v\n", current, err)
	}
}

// restoreCopy copies files from snapshot to live paths (additive overlay).
func restoreCopy(contextDir string, scope *config.Scope, manifest *Manifest) (int, error) {
	restored := 0
	for _, entry := range manifest.Files {
		srcPath := filepath.Join(contextDir, entry.RelPath)

		var dstPath string
		if isExtraFileSource(entry.Source) {
			ef := scope.ExtraFileByTag(entry.Source)
			if ef == nil {
				continue // tag not recognized in current scope, skip gracefully
			}
			dstPath = ef.Path
		} else {
			relToDotClaude := strings.TrimPrefix(entry.RelPath, "dotclaude/")
			dstPath = filepath.Join(scope.DotClaudeDir, relToDotClaude)
		}

		if err := fileutil.CopyFile(srcPath, dstPath); err != nil {
			return 0, fmt.Errorf("restore %s: %w", entry.RelPath, err)
		}
		restored++
	}
	return restored, nil
}

// ClearManagedFiles removes all managed files for the current scope.
// This includes the extra file (CLAUDE.md or claude.json) and matched files
// inside the .claude/ directory. Used by --from-scratch to start clean.
func ClearManagedFiles(cfg *config.Config) error {
	scope := cfg.Scope

	// Remove extra files (CLAUDE.md/.mcp.json for project, claude.json for user)
	for _, ef := range scope.ExtraFiles {
		if err := os.Remove(ef.Path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove %s: %w", filepath.Base(ef.Path), err)
		}
	}

	// Remove managed files from .claude/ directory
	if _, err := os.Stat(scope.DotClaudeDir); err == nil {
		walked, err := fileutil.WalkFiltered(
			scope.DotClaudeDir,
			scope.IncludePatterns,
			scope.ExcludePatterns,
		)
		if err != nil {
			return fmt.Errorf("walk %s: %w", scope.DotClaudeDir, err)
		}
		for _, w := range walked {
			if err := os.Remove(w.AbsPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove %s: %w", w.RelPath, err)
			}
		}

		// Remove empty directories bottom-up (but keep .claude/ itself)
		removeEmptyDirs(scope.DotClaudeDir)
	}

	return nil
}

// removeEmptyDirs walks the directory bottom-up and removes empty subdirectories.
// The root directory itself is preserved.
func removeEmptyDirs(root string) {
	// Collect directories bottom-up
	var dirs []string
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() || path == root {
			return nil
		}
		dirs = append(dirs, path)
		return nil
	})
	// Remove in reverse order (deepest first)
	for i := len(dirs) - 1; i >= 0; i-- {
		os.Remove(dirs[i]) // only succeeds if empty
	}
}

func createBackup(cfg *config.Config, scope *config.Scope) (string, error) {
	backupName := fmt.Sprintf("pre-switch-%s", time.Now().Format("20060102-150405"))
	backupDir := filepath.Join(cfg.BackupsDir(), backupName)

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", err
	}

	// Backup extra files (claude.json, CLAUDE.md, .mcp.json, etc.)
	for _, ef := range scope.ExtraFiles {
		if _, err := os.Stat(ef.Path); err == nil {
			storedName := filepath.Base(ef.Path)
			if err := fileutil.CopyFile(ef.Path, filepath.Join(backupDir, storedName)); err != nil {
				return backupDir, err
			}
		}
	}

	// Backup managed files from .claude
	if _, err := os.Stat(scope.DotClaudeDir); err == nil {
		walked, err := fileutil.WalkFiltered(
			scope.DotClaudeDir,
			scope.IncludePatterns,
			scope.ExcludePatterns,
		)
		if err != nil {
			return backupDir, err
		}
		for _, w := range walked {
			dst := filepath.Join(backupDir, "dotclaude", w.RelPath)
			if err := fileutil.CopyFile(w.AbsPath, dst); err != nil {
				return backupDir, err
			}
		}
	}

	return backupDir, nil
}
