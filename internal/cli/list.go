package cli

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/pfldy2850/claudectx/internal/config"
	"github.com/pfldy2850/claudectx/internal/context"
	"github.com/spf13/cobra"
)

var listJSON bool

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all saved contexts",
		Args:    cobra.NoArgs,
		RunE:    runList,
	}

	cmd.Flags().BoolVar(&listJSON, "json", false, "Output as JSON")

	return cmd
}

// scopeContexts holds listing data for a single scope.
type scopeContexts struct {
	cfg     *config.Config
	names   []string
	current string
}

func runList(cmd *cobra.Command, args []string) error {
	// When no --scope/--root flags and project scope is available, show both scopes.
	if scopeFlag == "" && rootFlag == "" {
		if err := tryListBothScopes(); err == nil {
			return nil
		}
	}

	// Single scope: either explicit --scope or auto-detected (user only).
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	return runListSingleScope(cfg)
}

// tryListBothScopes attempts to display both project and user scopes.
// Returns nil on success, or an error if dual-scope display is not applicable.
func tryListBothScopes() error {
	root, err := config.DetectProjectRoot()
	if err != nil {
		return err
	}
	projectScope := config.ProjectScopeAt(root)

	// If project and user scope share the same storage dir (e.g. project root
	// is ~), skip dual-scope display and show user scope only.
	userScope, err := config.UserScope()
	if err == nil && projectScope.StorageDir == userScope.StorageDir {
		cfg, err := config.LoadWithScope(configPath, userScope)
		if err != nil {
			return err
		}
		return runListSingleScope(cfg)
	}
	return runListBothScopes(projectScope)
}

func runListBothScopes(projectScope *config.Scope) error {
	// Build configs for both scopes.
	projectCfg, err := config.LoadWithScope(configPath, projectScope)
	if err != nil {
		return err
	}

	userScope, err := config.UserScope()
	if err != nil {
		return err
	}
	userCfg, err := config.LoadWithScope(configPath, userScope)
	if err != nil {
		return err
	}

	// Collect contexts from both scopes.
	var scopes []scopeContexts
	for _, cfg := range []*config.Config{projectCfg, userCfg} {
		names, err := context.ListContexts(cfg.ContextsDir())
		if err != nil {
			return err
		}
		current, _ := context.GetCurrent(cfg)
		scopes = append(scopes, scopeContexts{cfg: cfg, names: names, current: current})
	}

	if listJSON {
		return listBothScopesAsJSON(scopes)
	}

	// Check if both scopes are empty.
	total := 0
	for _, s := range scopes {
		total += len(s.names)
	}
	if total == 0 {
		fmt.Println("No saved contexts.")
		return nil
	}

	for i, sc := range scopes {
		if i > 0 && len(sc.names) > 0 {
			fmt.Println()
		}
		if len(sc.names) == 0 {
			continue
		}
		printScopeSection(sc)
	}

	return nil
}

func runListSingleScope(cfg *config.Config) error {
	names, err := context.ListContexts(cfg.ContextsDir())
	if err != nil {
		return err
	}

	current, _ := context.GetCurrent(cfg)

	if listJSON {
		return listAsJSON(cfg, names, current)
	}

	if len(names) == 0 {
		fmt.Println("No saved contexts.")
		return nil
	}

	sc := scopeContexts{cfg: cfg, names: names, current: current}
	printScopeSection(sc)
	return nil
}

func printScopeSection(sc scopeContexts) {
	scope := sc.cfg.Scope
	fmt.Printf("Scope: %s (%s)\n", scope.Type, scope.StorageDir)

	for _, name := range sc.names {
		marker := "  "
		if name == sc.current {
			marker = "* "
		}

		m, err := context.ReadManifest(filepath.Join(sc.cfg.ContextsDir(), name))
		if err != nil {
			fmt.Printf("%s%s (error reading manifest)\n", marker, name)
			continue
		}

		desc := ""
		if m.Description != "" {
			desc = fmt.Sprintf(" - %s", m.Description)
		}
		fmt.Printf("%s%s (%d files, %s)%s\n",
			marker, name, len(m.Files), formatSize(m.TotalSize), desc)
	}
}

type listOutput struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Files       int    `json:"files"`
	TotalSize   int64  `json:"totalSize"`
	IsCurrent   bool   `json:"isCurrent"`
	Scope       string `json:"scope,omitempty"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

func buildListItems(cfg *config.Config, names []string, current string) []listOutput {
	items := []listOutput{}
	for _, name := range names {
		m, err := context.ReadManifest(filepath.Join(cfg.ContextsDir(), name))
		if err != nil {
			continue
		}
		items = append(items, listOutput{
			Name:        m.Name,
			Description: m.Description,
			Files:       len(m.Files),
			TotalSize:   m.TotalSize,
			IsCurrent:   m.Name == current,
			Scope:       m.Scope,
			CreatedAt:   m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   m.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return items
}

func listBothScopesAsJSON(scopes []scopeContexts) error {
	allItems := []listOutput{}
	for _, sc := range scopes {
		allItems = append(allItems, buildListItems(sc.cfg, sc.names, sc.current)...)
	}

	data, err := json.MarshalIndent(allItems, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func listAsJSON(cfg *config.Config, names []string, current string) error {
	items := buildListItems(cfg, names, current)

	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
