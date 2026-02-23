package cli

import (
	"fmt"
	"path/filepath"

	"github.com/pfldy2850/claudectx/internal/context"
	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show <name>",
		Short: "Show context details",
		Args:  cobra.ExactArgs(1),
		RunE:  runShow,
	}
}

func runShow(cmd *cobra.Command, args []string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	slug := context.Slugify(args[0])
	m, err := context.ReadManifest(filepath.Join(cfg.ContextsDir(), slug))
	if err != nil {
		return fmt.Errorf("context %q not found", slug)
	}

	current, _ := context.GetCurrent(cfg)
	activeMarker := ""
	if slug == current {
		activeMarker = " (active)"
	}

	fmt.Printf("Context: %s%s\n", m.Name, activeMarker)
	if m.Description != "" {
		fmt.Printf("Description: %s\n", m.Description)
	}
	if m.Scope != "" {
		fmt.Printf("Scope: %s\n", m.Scope)
	}
	if m.OAuthEmail != "" {
		fmt.Printf("OAuth Email: %s\n", m.OAuthEmail)
	}
	fmt.Printf("Created: %s\n", m.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Updated: %s\n", m.UpdatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("Files: %d\n", len(m.Files))
	fmt.Printf("Total Size: %s\n", formatSize(m.TotalSize))
	fmt.Printf("Checksum: %s\n", m.Checksum[:12]+"...")

	fmt.Println("\nFiles:")
	for _, f := range m.Files {
		fmt.Printf("  %s (%s) [%s]\n", f.RelPath, formatSize(f.Size), f.Source)
	}

	return nil
}
