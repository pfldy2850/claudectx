package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pfldy2850/claudectx/internal/config"
	"github.com/pfldy2850/claudectx/internal/context"
	"github.com/pfldy2850/claudectx/internal/fileutil"
	"github.com/spf13/cobra"
)

var (
	createFromScratch bool
	createCopyFrom    string
	createDescription string
)

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new context",
		Long: "Create a new context from the current state, from scratch, or by copying an existing context.\n\n" +
			"By default, snapshots the current live files as the new context.",
		Args: cobra.ExactArgs(1),
		RunE: runCreate,
	}

	cmd.Flags().BoolVar(&createFromScratch, "from-scratch", false, "Create an empty context (no files)")
	cmd.Flags().StringVar(&createCopyFrom, "copy-from", "", "Copy from an existing context")
	cmd.Flags().StringVar(&createDescription, "description", "", "Description for this context")

	return cmd
}

func runCreate(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	name := args[0]
	slug := context.Slugify(name)
	if slug == "" {
		return fmt.Errorf("invalid context name: %q", name)
	}

	if context.ContextExists(cfg.ContextsDir(), slug) {
		return fmt.Errorf("context %q already exists", slug)
	}

	if createFromScratch && createCopyFrom != "" {
		return fmt.Errorf("--from-scratch and --copy-from cannot be used together")
	}

	switch {
	case createFromScratch:
		return doCreateFromScratch(cfg, slug)
	case createCopyFrom != "":
		return doCreateCopyFrom(cfg, slug, createCopyFrom)
	default:
		return doCreateFromCurrent(cfg, name, slug)
	}
}

// doCreateFromCurrent saves the current live state as a new context.
// Auto-saves the previous current context first so edits aren't lost.
func doCreateFromCurrent(cfg *config.Config, name, slug string) error {
	context.AutoSaveCurrent(cfg, slug)

	saveResult, err := context.Save(context.SaveOptions{
		Name:        name,
		Description: createDescription,
		DryRun:      dryRun,
		Verbose:     verbose,
		Config:      cfg,
	})
	if err != nil {
		return err
	}

	if dryRun {
		fmt.Printf("[dry-run] Would create context %q (%d files, %s)\n",
			saveResult.Name, saveResult.Files, formatSize(saveResult.TotalSize))
	} else {
		fmt.Printf("Context %q created (%d files, %s)\n",
			saveResult.Name, saveResult.Files, formatSize(saveResult.TotalSize))
	}

	if cfg.Scope != nil && cfg.Scope.Type == config.ScopeProject {
		ensureGitignore(cfg.Scope, dryRun)
	}
	return nil
}

// doCreateFromScratch creates an empty context and switches to it.
func doCreateFromScratch(cfg *config.Config, slug string) error {
	if dryRun {
		fmt.Printf("[dry-run] Would create empty context %q\n", slug)
		return nil
	}

	// Auto-save current context before clearing
	context.AutoSaveCurrent(cfg, slug)

	contextDir := filepath.Join(cfg.ContextsDir(), slug)
	if err := os.MkdirAll(filepath.Join(contextDir, "dotclaude"), 0755); err != nil {
		return fmt.Errorf("create context dir: %w", err)
	}

	now := time.Now()
	manifest := &context.Manifest{
		Name:        slug,
		Description: createDescription,
		CreatedAt:   now,
		UpdatedAt:   now,
		Files:       []context.FileEntry{},
		Checksum:    context.ManifestChecksum(nil),
		Scope:       string(cfg.Scope.Type),
	}
	if err := context.WriteManifest(contextDir, manifest); err != nil {
		return err
	}

	// Clear all managed files for a clean slate
	if err := context.ClearManagedFiles(cfg); err != nil {
		return fmt.Errorf("clear managed files: %w", err)
	}

	// Update current marker to the new empty context
	if err := context.SetCurrent(cfg, slug); err != nil {
		return err
	}

	fmt.Printf("Context %q created from scratch (clean slate)\n", slug)

	if cfg.Scope != nil && cfg.Scope.Type == config.ScopeProject {
		ensureGitignore(cfg.Scope, false)
	}
	return nil
}

// doCreateCopyFrom copies an existing context under a new name and switches to it.
func doCreateCopyFrom(cfg *config.Config, slug, srcName string) error {
	srcSlug := context.Slugify(srcName)
	if !context.ContextExists(cfg.ContextsDir(), srcSlug) {
		return fmt.Errorf("source context %q not found", srcSlug)
	}

	if dryRun {
		fmt.Printf("[dry-run] Would create context %q from %q\n", slug, srcSlug)
		return nil
	}

	srcDir := filepath.Join(cfg.ContextsDir(), srcSlug)
	dstDir := filepath.Join(cfg.ContextsDir(), slug)

	if err := fileutil.CopyDir(srcDir, dstDir); err != nil {
		return fmt.Errorf("copy context: %w", err)
	}

	// Update manifest with new name
	manifest, err := context.ReadManifest(dstDir)
	if err != nil {
		return fmt.Errorf("read copied manifest: %w", err)
	}
	manifest.Name = slug
	manifest.Description = createDescription
	manifest.UpdatedAt = time.Now()
	if err := context.WriteManifest(dstDir, manifest); err != nil {
		return err
	}

	// Switch to the copied context
	result, err := context.Restore(context.RestoreOptions{
		Name:   slug,
		Force:  true,
		Config: cfg,
	})
	if err != nil {
		return err
	}

	fmt.Printf("Context %q created from %q (%d files)\n", slug, srcSlug, result.FilesRestored)

	if cfg.Scope != nil && cfg.Scope.Type == config.ScopeProject {
		ensureGitignore(cfg.Scope, false)
	}
	return nil
}
