package context

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pfldy2850/claudectx/internal/config"
	"github.com/pfldy2850/claudectx/internal/fileutil"
)

// toSlash normalizes a path to use forward slashes for portable manifest storage.
func toSlash(p string) string { return filepath.ToSlash(p) }

// SaveOptions configures the save operation.
type SaveOptions struct {
	Name        string
	Description string
	Overwrite   bool
	DryRun      bool
	Verbose     bool
	Config      *config.Config
}

// SaveResult holds the result of a save operation.
type SaveResult struct {
	Name      string
	Dir       string
	Files     int
	TotalSize int64
}

// Save creates a snapshot of the current Claude Code state.
func Save(opts SaveOptions) (*SaveResult, error) {
	slug := Slugify(opts.Name)
	if slug == "" {
		return nil, fmt.Errorf("invalid context name: %q", opts.Name)
	}

	cfg := opts.Config
	scope := cfg.Scope
	contextDir := filepath.Join(cfg.ContextsDir(), slug)

	if ContextExists(cfg.ContextsDir(), slug) && !opts.Overwrite {
		return nil, fmt.Errorf("context %q already exists", slug)
	}

	dotClaudeDir := scope.DotClaudeDir

	if opts.DryRun {
		return dryRunSave(slug, dotClaudeDir, cfg, opts)
	}

	// Clear existing context if overwriting
	if opts.Overwrite {
		os.RemoveAll(contextDir)
	}

	if err := os.MkdirAll(contextDir, 0755); err != nil {
		return nil, fmt.Errorf("create context dir: %w", err)
	}

	var files []FileEntry
	var totalSize int64

	// 1. Snapshot extra files (claude.json for user scope; CLAUDE.md, .mcp.json for project scope)
	for _, ef := range scope.ExtraFiles {
		info, err := os.Stat(ef.Path)
		if err != nil {
			continue // file doesn't exist, skip
		}
		storedName := filepath.Base(ef.Path)
		dstPath := filepath.Join(contextDir, storedName)
		if err := fileutil.CopyFile(ef.Path, dstPath); err != nil {
			return nil, fmt.Errorf("copy %s: %w", storedName, err)
		}
		checksum, _ := FileChecksum(ef.Path)
		files = append(files, FileEntry{
			RelPath:  storedName,
			Size:     info.Size(),
			Mode:     uint32(info.Mode()),
			Checksum: checksum,
			Source:   ef.Tag,
		})
		totalSize += info.Size()
	}

	// 2. Snapshot .claude/ directory (filtered)
	if _, err := os.Stat(dotClaudeDir); err == nil {
		walked, err := fileutil.WalkFiltered(
			dotClaudeDir,
			cfg.IncludePatterns,
			cfg.ExcludePatterns,
		)
		if err != nil {
			return nil, fmt.Errorf("walk .claude: %w", err)
		}

		dotClaudeDst := filepath.Join(contextDir, "dotclaude")
		for _, w := range walked {
			dstPath := filepath.Join(dotClaudeDst, w.RelPath)
			if err := fileutil.CopyFile(w.AbsPath, dstPath); err != nil {
				return nil, fmt.Errorf("copy %s: %w", w.RelPath, err)
			}
				checksum, _ := FileChecksum(w.AbsPath)
			files = append(files, FileEntry{
				RelPath:  toSlash(filepath.Join("dotclaude", w.RelPath)),
				Size:     w.Info.Size(),
				Mode:     uint32(w.Info.Mode()),
				Checksum: checksum,
				Source:   "dotclaude",
			})
			totalSize += w.Info.Size()
		}
	}

	// 3. Extract OAuth email from claude.json (user scope only)
	var oauthEmail string
	if scope.Type == config.ScopeUser {
		if ef := scope.ExtraFileByTag("claudejson"); ef != nil {
			oauthEmail = extractOAuthEmail(ef.Path)
		}
	}

	// 4. Write manifest
	now := time.Now()
	manifest := &Manifest{
		Name:        slug,
		Description: opts.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
		Files:       files,
		TotalSize:   totalSize,
		Checksum:    ManifestChecksum(files),
		OAuthEmail:  oauthEmail,
		Scope:       string(scope.Type),
	}

	if err := WriteManifest(contextDir, manifest); err != nil {
		return nil, err
	}

	// 5. Update current marker
	if err := SetCurrent(cfg, slug); err != nil {
		return nil, err
	}

	return &SaveResult{
		Name:      slug,
		Dir:       contextDir,
		Files:     len(files),
		TotalSize: totalSize,
	}, nil
}

func dryRunSave(slug, dotClaudeDir string, cfg *config.Config, opts SaveOptions) (*SaveResult, error) {
	var fileCount int
	var totalSize int64

	for _, ef := range cfg.Scope.ExtraFiles {
		if info, err := os.Stat(ef.Path); err == nil {
			fileCount++
			totalSize += info.Size()
		}
	}

	if _, err := os.Stat(dotClaudeDir); err == nil {
		walked, err := fileutil.WalkFiltered(
			dotClaudeDir,
			cfg.IncludePatterns,
			cfg.ExcludePatterns,
		)
		if err != nil {
			return nil, err
		}
		for _, w := range walked {
			fileCount++
			totalSize += w.Info.Size()
		}
	}

	return &SaveResult{
		Name:      slug,
		Files:     fileCount,
		TotalSize: totalSize,
	}, nil
}

func extractOAuthEmail(claudeJSONPath string) string {
	data, err := os.ReadFile(claudeJSONPath)
	if err != nil {
		return ""
	}
	var obj map[string]any
	if err := json.Unmarshal(data, &obj); err != nil {
		return ""
	}
	if email, ok := obj["oauthEmail"].(string); ok {
		return email
	}
	return ""
}

// SetCurrent writes the active context name to the current marker file.
func SetCurrent(cfg *config.Config, name string) error {
	if err := os.MkdirAll(cfg.StorageDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(cfg.CurrentFile(), []byte(name+"\n"), 0644)
}

// GetCurrent reads the active context name.
func GetCurrent(cfg *config.Config) (string, error) {
	data, err := os.ReadFile(cfg.CurrentFile())
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	name := string(data)
	name = name[:len(name)-1] // trim newline
	return name, nil
}
