package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pfldy2850/claudectx/internal/config"
	"github.com/pfldy2850/claudectx/internal/context"
	"github.com/pfldy2850/claudectx/internal/ui"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	verbose    bool
	dryRun     bool
	force      bool
	configPath string
	scopeFlag  string
	rootFlag   string
)

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "claudectx [context]",
		Short: "Claude Code context management tool",
		Long: "Manage Claude Code configuration contexts by creating and switching snapshots.\n\n" +
			"Supports two scopes:\n" +
			"  user    — manages ~/.claude/ and ~/.claude.json (default outside git repos)\n" +
			"  project — manages <project-root>/.claude/ and <project-root>/CLAUDE.md (default inside projects)\n\n" +
			"Project root detection (highest priority first):\n" +
			"  1. --root flag (explicit path)\n" +
			"  2. Claude marker files (.claude/, CLAUDE.md, .claudectx/)\n" +
			"  3. Git repository root (.git/)\n" +
			"  4. Current directory fallback (with --scope project)\n\n" +
			"Use --scope to override auto-detection, --root to set an explicit project root.",
		Args: cobra.MaximumNArgs(1),
		RunE: runRoot,
	}

	root.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	root.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes")
	root.PersistentFlags().BoolVarP(&force, "force", "f", false, "Force operation (skip confirmations)")
	root.PersistentFlags().StringVar(&configPath, "config", "", "Path to config file")
	root.PersistentFlags().StringVar(&scopeFlag, "scope", "", "Scope: 'user' or 'project' (auto-detects if omitted)")
	root.PersistentFlags().StringVar(&rootFlag, "root", "", "Explicit project root directory (implies project scope)")

	root.AddCommand(
		newCreateCmd(),
		newListCmd(),
		newShowCmd(),
		newDeleteCmd(),
		newCurrentCmd(),
		newVersionCmd(),
	)

	return root
}

// Execute runs the root command.
func Execute() error {
	return newRootCmd().Execute()
}

func runRoot(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	// If a context name is provided, switch to it
	if len(args) == 1 {
		return switchContext(cfg, args[0])
	}

	// No arguments — interactive selection
	return interactiveSelect(cfg)
}

func switchContext(cfg *config.Config, name string) error {
	slug := context.Slugify(name)
	if !context.ContextExists(cfg.ContextsDir(), slug) {
		return fmt.Errorf("context %q not found; use 'claudectx create %s' to create it", slug, slug)
	}

	current, _ := context.GetCurrent(cfg)
	if current == slug && !force {
		fmt.Printf("Already on context %q\n", slug)
		return nil
	}

	result, err := context.Restore(context.RestoreOptions{
		Name:    name,
		DryRun:  dryRun,
		Force:   force,
		Verbose: verbose,
		Config:  cfg,
	})
	if err != nil {
		return err
	}

	if dryRun {
		fmt.Printf("[dry-run] Would switch to context %q (%d files)\n", result.Name, result.FilesRestored)
		return nil
	}

	fmt.Printf("Switched to context %q (%d files)\n", result.Name, result.FilesRestored)
	return nil
}

func interactiveSelect(cfg *config.Config) error {
	names, err := context.ListContexts(cfg.ContextsDir())
	if err != nil {
		return err
	}
	if len(names) == 0 {
		fmt.Println("No saved contexts. Use 'claudectx create <name>' to create one.")
		return nil
	}

	current, _ := context.GetCurrent(cfg)

	// Load manifests for display
	var items []ui.PickerItem
	for _, name := range names {
		m, err := context.ReadManifest(filepath.Join(cfg.ContextsDir(), name))
		if err != nil {
			continue
		}
		items = append(items, ui.PickerItem{
			Name:        m.Name,
			Description: m.Description,
			Files:       len(m.Files),
			TotalSize:   m.TotalSize,
			UpdatedAt:   m.UpdatedAt,
			IsCurrent:   m.Name == current,
		})
	}

	selected, err := ui.RunPicker(items)
	if err != nil {
		return err
	}
	if selected == "" {
		return nil // user cancelled
	}

	return switchContext(cfg, selected)
}

func loadConfig() (*config.Config, error) {
	scope, err := config.ResolveScopeWithRoot(scopeFlag, rootFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving scope: %v\n", err)
		return nil, err
	}
	cfg, err := config.LoadWithScope(configPath, scope)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		return nil, err
	}
	return cfg, nil
}
