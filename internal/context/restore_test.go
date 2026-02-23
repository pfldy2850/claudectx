package context

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pfldy2850/claudectx/internal/config"
)

// TestAutoSaveOnSwitch verifies that switching contexts auto-saves
// the current context's live state before restoring the target.
func TestAutoSaveOnSwitch(t *testing.T) {
	projectRoot := t.TempDir()
	storageDir := filepath.Join(projectRoot, ".claudectx")
	dotClaudeDir := filepath.Join(projectRoot, ".claude")
	os.MkdirAll(dotClaudeDir, 0755)

	claudeMDPath := filepath.Join(projectRoot, "CLAUDE.md")

	cfg := &config.Config{
		StorageDir:      storageDir,
		IncludePatterns: config.DefaultProjectIncludePatterns,
		ExcludePatterns: config.DefaultProjectExcludePatterns,
		Scope: &config.Scope{
			Type:            config.ScopeProject,
			DotClaudeDir:    dotClaudeDir,
			ExtraFiles: []config.ExtraFile{
				{Path: claudeMDPath, Tag: "claudemd"},
			},
			StorageDir:      storageDir,
			IncludePatterns: config.DefaultProjectIncludePatterns,
			ExcludePatterns: config.DefaultProjectExcludePatterns,
		},
	}

	// Create context A with initial content
	os.WriteFile(filepath.Join(dotClaudeDir, "settings.json"), []byte(`{"ctx":"a"}`), 0644)
	os.WriteFile(claudeMDPath, []byte("# Context A"), 0644)

	_, err := Save(SaveOptions{Name: "ctx-a", Config: cfg})
	if err != nil {
		t.Fatalf("Save ctx-a failed: %v", err)
	}

	// Create context B with different content
	os.WriteFile(filepath.Join(dotClaudeDir, "settings.json"), []byte(`{"ctx":"b"}`), 0644)
	os.WriteFile(claudeMDPath, []byte("# Context B"), 0644)

	_, err = Save(SaveOptions{Name: "ctx-b", Config: cfg})
	if err != nil {
		t.Fatalf("Save ctx-b failed: %v", err)
	}
	// Current is now ctx-b

	// Modify live files while on ctx-b
	os.WriteFile(claudeMDPath, []byte("# Context B Modified"), 0644)

	// Switch to ctx-a — should auto-save ctx-b first
	_, err = Restore(RestoreOptions{Name: "ctx-a", Config: cfg})
	if err != nil {
		t.Fatalf("Restore ctx-a failed: %v", err)
	}

	// Verify we're now on ctx-a
	data, _ := os.ReadFile(claudeMDPath)
	if string(data) != "# Context A" {
		t.Errorf("expected ctx-a content, got %q", data)
	}

	// Verify ctx-b's snapshot was auto-saved with the modification
	ctxBDir := filepath.Join(cfg.ContextsDir(), "ctx-b")
	snapshotData, _ := os.ReadFile(filepath.Join(ctxBDir, "CLAUDE.md"))
	if string(snapshotData) != "# Context B Modified" {
		t.Errorf("expected ctx-b snapshot to have modified content, got %q", snapshotData)
	}
}

// TestAutoSaveSkipsWhenSameContext verifies that switching to the
// same context doesn't trigger an auto-save loop.
func TestAutoSaveSkipsWhenSameContext(t *testing.T) {
	projectRoot := t.TempDir()
	storageDir := filepath.Join(projectRoot, ".claudectx")
	dotClaudeDir := filepath.Join(projectRoot, ".claude")
	os.MkdirAll(dotClaudeDir, 0755)

	claudeMDPath := filepath.Join(projectRoot, "CLAUDE.md")
	os.WriteFile(filepath.Join(dotClaudeDir, "settings.json"), []byte(`{}`), 0644)
	os.WriteFile(claudeMDPath, []byte("# Hello"), 0644)

	cfg := &config.Config{
		StorageDir:      storageDir,
		IncludePatterns: config.DefaultProjectIncludePatterns,
		ExcludePatterns: config.DefaultProjectExcludePatterns,
		Scope: &config.Scope{
			Type:            config.ScopeProject,
			DotClaudeDir:    dotClaudeDir,
			ExtraFiles: []config.ExtraFile{
				{Path: claudeMDPath, Tag: "claudemd"},
			},
			StorageDir:      storageDir,
			IncludePatterns: config.DefaultProjectIncludePatterns,
			ExcludePatterns: config.DefaultProjectExcludePatterns,
		},
	}

	_, err := Save(SaveOptions{Name: "same-ctx", Config: cfg})
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Modify and switch to same context — should not panic or error
	os.WriteFile(claudeMDPath, []byte("# Modified"), 0644)

	_, err = Restore(RestoreOptions{Name: "same-ctx", Config: cfg})
	if err != nil {
		t.Fatalf("Restore same context failed: %v", err)
	}

	// Content should be the original snapshot (not modified)
	data, _ := os.ReadFile(claudeMDPath)
	if string(data) != "# Hello" {
		t.Errorf("expected original content, got %q", data)
	}
}

// TestAutoSaveSkipsWhenNoCurrent verifies that auto-save doesn't
// error when there's no current context (first switch).
func TestAutoSaveSkipsWhenNoCurrent(t *testing.T) {
	projectRoot := t.TempDir()
	storageDir := filepath.Join(projectRoot, ".claudectx")
	dotClaudeDir := filepath.Join(projectRoot, ".claude")
	os.MkdirAll(dotClaudeDir, 0755)

	claudeMDPath := filepath.Join(projectRoot, "CLAUDE.md")
	os.WriteFile(filepath.Join(dotClaudeDir, "settings.json"), []byte(`{}`), 0644)
	os.WriteFile(claudeMDPath, []byte("# Hello"), 0644)

	cfg := &config.Config{
		StorageDir:      storageDir,
		IncludePatterns: config.DefaultProjectIncludePatterns,
		ExcludePatterns: config.DefaultProjectExcludePatterns,
		Scope: &config.Scope{
			Type:            config.ScopeProject,
			DotClaudeDir:    dotClaudeDir,
			ExtraFiles: []config.ExtraFile{
				{Path: claudeMDPath, Tag: "claudemd"},
			},
			StorageDir:      storageDir,
			IncludePatterns: config.DefaultProjectIncludePatterns,
			ExcludePatterns: config.DefaultProjectExcludePatterns,
		},
	}

	// Create context without setting current (manually)
	contextDir := filepath.Join(cfg.ContextsDir(), "first-ctx")
	os.MkdirAll(filepath.Join(contextDir, "dotclaude"), 0755)
	os.WriteFile(filepath.Join(contextDir, "CLAUDE.md"), []byte("# First"), 0644)
	os.WriteFile(filepath.Join(contextDir, "dotclaude", "settings.json"), []byte(`{"first":true}`), 0644)

	settingsSum, _ := FileChecksum(filepath.Join(contextDir, "dotclaude", "settings.json"))
	extraSum, _ := FileChecksum(filepath.Join(contextDir, "CLAUDE.md"))
	manifest := &Manifest{
		Name:  "first-ctx",
		Scope: "project",
		Files: []FileEntry{
			{RelPath: "CLAUDE.md", Size: 7, Mode: 0644, Checksum: extraSum, Source: "claudemd"},
			{RelPath: "dotclaude/settings.json", Size: 14, Mode: 0644, Checksum: settingsSum, Source: "dotclaude"},
		},
	}
	WriteManifest(contextDir, manifest)

	// Switch with no current context set — should not error
	_, err := Restore(RestoreOptions{Name: "first-ctx", Config: cfg})
	if err != nil {
		t.Fatalf("Restore with no current context failed: %v", err)
	}

	data, _ := os.ReadFile(claudeMDPath)
	if string(data) != "# First" {
		t.Errorf("expected first-ctx content, got %q", data)
	}
}

// TestClearManagedFiles verifies that ClearManagedFiles removes the extra
// file and managed files inside .claude/ while leaving unmanaged files alone.
func TestClearManagedFiles(t *testing.T) {
	projectRoot := t.TempDir()
	storageDir := filepath.Join(projectRoot, ".claudectx")
	dotClaudeDir := filepath.Join(projectRoot, ".claude")
	os.MkdirAll(dotClaudeDir, 0755)

	claudeMDPath := filepath.Join(projectRoot, "CLAUDE.md")

	cfg := &config.Config{
		StorageDir:      storageDir,
		IncludePatterns: config.DefaultProjectIncludePatterns,
		ExcludePatterns: config.DefaultProjectExcludePatterns,
		Scope: &config.Scope{
			Type:            config.ScopeProject,
			DotClaudeDir:    dotClaudeDir,
			ExtraFiles: []config.ExtraFile{
				{Path: claudeMDPath, Tag: "claudemd"},
			},
			StorageDir:      storageDir,
			IncludePatterns: config.DefaultProjectIncludePatterns,
			ExcludePatterns: config.DefaultProjectExcludePatterns,
		},
	}

	// Create managed files
	os.WriteFile(claudeMDPath, []byte("# My Context"), 0644)
	os.WriteFile(filepath.Join(dotClaudeDir, "settings.json"), []byte(`{"key":"val"}`), 0644)
	os.MkdirAll(filepath.Join(dotClaudeDir, "rules"), 0755)
	os.WriteFile(filepath.Join(dotClaudeDir, "rules", "my-rule.md"), []byte("rule content"), 0644)

	// Clear
	if err := ClearManagedFiles(cfg); err != nil {
		t.Fatalf("ClearManagedFiles failed: %v", err)
	}

	// Extra file should be gone
	if _, err := os.Stat(claudeMDPath); !os.IsNotExist(err) {
		t.Errorf("expected CLAUDE.md to be removed")
	}

	// Managed files in .claude/ should be gone
	if _, err := os.Stat(filepath.Join(dotClaudeDir, "settings.json")); !os.IsNotExist(err) {
		t.Errorf("expected settings.json to be removed")
	}
	if _, err := os.Stat(filepath.Join(dotClaudeDir, "rules", "my-rule.md")); !os.IsNotExist(err) {
		t.Errorf("expected rules/my-rule.md to be removed")
	}

	// .claude/ directory itself should still exist
	if _, err := os.Stat(dotClaudeDir); err != nil {
		t.Errorf("expected .claude/ directory to still exist")
	}

	// Empty subdirectory (rules/) should be removed
	if _, err := os.Stat(filepath.Join(dotClaudeDir, "rules")); !os.IsNotExist(err) {
		t.Errorf("expected empty rules/ directory to be removed")
	}
}

// TestClearManagedFilesUserScope verifies clearing works for user scope
// (extra file is claude.json, include patterns are more selective).
func TestClearManagedFilesUserScope(t *testing.T) {
	homeDir := t.TempDir()
	storageDir := filepath.Join(homeDir, ".claudectx")
	dotClaudeDir := filepath.Join(homeDir, ".claude")
	os.MkdirAll(dotClaudeDir, 0755)

	claudeJSONPath := filepath.Join(homeDir, ".claude.json")

	cfg := &config.Config{
		StorageDir:      storageDir,
		IncludePatterns: config.DefaultIncludePatterns,
		ExcludePatterns: config.DefaultExcludePatterns,
		Scope: &config.Scope{
			Type:            config.ScopeUser,
			DotClaudeDir:    dotClaudeDir,
			ExtraFiles: []config.ExtraFile{
				{Path: claudeJSONPath, Tag: "claudejson"},
			},
			StorageDir:      storageDir,
			IncludePatterns: config.DefaultIncludePatterns,
			ExcludePatterns: config.DefaultExcludePatterns,
		},
	}

	// Create files
	os.WriteFile(claudeJSONPath, []byte(`{"oauthEmail":"test@example.com"}`), 0644)
	os.WriteFile(filepath.Join(dotClaudeDir, "settings.json"), []byte(`{"theme":"dark"}`), 0644)
	// Create an excluded file (should NOT be removed)
	os.MkdirAll(filepath.Join(dotClaudeDir, "cache"), 0755)
	os.WriteFile(filepath.Join(dotClaudeDir, "cache", "data.bin"), []byte("cached"), 0644)

	if err := ClearManagedFiles(cfg); err != nil {
		t.Fatalf("ClearManagedFiles failed: %v", err)
	}

	// Extra file should be gone
	if _, err := os.Stat(claudeJSONPath); !os.IsNotExist(err) {
		t.Errorf("expected .claude.json to be removed")
	}

	// Managed file should be gone
	if _, err := os.Stat(filepath.Join(dotClaudeDir, "settings.json")); !os.IsNotExist(err) {
		t.Errorf("expected settings.json to be removed")
	}

	// Excluded file should still exist
	if _, err := os.Stat(filepath.Join(dotClaudeDir, "cache", "data.bin")); err != nil {
		t.Errorf("expected excluded cache/data.bin to still exist, got: %v", err)
	}
}

// TestSwitchClearsStaleFiles verifies that switching contexts removes files
// from the previous context that don't exist in the target context.
func TestSwitchClearsStaleFiles(t *testing.T) {
	projectRoot := t.TempDir()
	storageDir := filepath.Join(projectRoot, ".claudectx")
	dotClaudeDir := filepath.Join(projectRoot, ".claude")
	os.MkdirAll(dotClaudeDir, 0755)

	claudeMDPath := filepath.Join(projectRoot, "CLAUDE.md")

	cfg := &config.Config{
		StorageDir:      storageDir,
		IncludePatterns: config.DefaultProjectIncludePatterns,
		ExcludePatterns: config.DefaultProjectExcludePatterns,
		Scope: &config.Scope{
			Type:            config.ScopeProject,
			DotClaudeDir:    dotClaudeDir,
			ExtraFiles: []config.ExtraFile{
				{Path: claudeMDPath, Tag: "claudemd"},
			},
			StorageDir:      storageDir,
			IncludePatterns: config.DefaultProjectIncludePatterns,
			ExcludePatterns: config.DefaultProjectExcludePatterns,
		},
	}

	// Context A: has CLAUDE.md and .claude/CLAUDE.md
	os.WriteFile(claudeMDPath, []byte("# Root CLAUDE"), 0644)
	os.WriteFile(filepath.Join(dotClaudeDir, "CLAUDE.md"), []byte("# Dot CLAUDE A"), 0644)

	_, err := Save(SaveOptions{Name: "ctx-a", Config: cfg})
	if err != nil {
		t.Fatalf("Save ctx-a failed: %v", err)
	}

	// Context B: has CLAUDE.md only (no .claude/CLAUDE.md)
	os.Remove(filepath.Join(dotClaudeDir, "CLAUDE.md"))
	os.WriteFile(claudeMDPath, []byte("# Root CLAUDE B"), 0644)

	_, err = Save(SaveOptions{Name: "ctx-b", Config: cfg})
	if err != nil {
		t.Fatalf("Save ctx-b failed: %v", err)
	}
	// Current is now ctx-b

	// Switch back to ctx-a (has .claude/CLAUDE.md)
	_, err = Restore(RestoreOptions{Name: "ctx-a", Config: cfg})
	if err != nil {
		t.Fatalf("Restore ctx-a failed: %v", err)
	}

	// .claude/CLAUDE.md should exist (from ctx-a)
	data, err := os.ReadFile(filepath.Join(dotClaudeDir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("expected .claude/CLAUDE.md to exist after switching to ctx-a: %v", err)
	}
	if string(data) != "# Dot CLAUDE A" {
		t.Errorf("expected ctx-a .claude/CLAUDE.md content, got %q", data)
	}

	// Now switch to ctx-b (does NOT have .claude/CLAUDE.md)
	_, err = Restore(RestoreOptions{Name: "ctx-b", Config: cfg})
	if err != nil {
		t.Fatalf("Restore ctx-b failed: %v", err)
	}

	// .claude/CLAUDE.md should NOT exist (ctx-b doesn't have it)
	if _, err := os.Stat(filepath.Join(dotClaudeDir, "CLAUDE.md")); !os.IsNotExist(err) {
		t.Errorf("expected .claude/CLAUDE.md to be removed after switching to ctx-b, but it still exists")
	}

	// Root CLAUDE.md should have ctx-b content
	data, _ = os.ReadFile(claudeMDPath)
	if string(data) != "# Root CLAUDE B" {
		t.Errorf("expected ctx-b root CLAUDE.md content, got %q", data)
	}
}

// TestMcpJsonSaveRestoreRoundTrip verifies that .mcp.json is correctly
// saved and restored as part of project scope contexts.
func TestMcpJsonSaveRestoreRoundTrip(t *testing.T) {
	projectRoot := t.TempDir()
	storageDir := filepath.Join(projectRoot, ".claudectx")
	dotClaudeDir := filepath.Join(projectRoot, ".claude")
	os.MkdirAll(dotClaudeDir, 0755)

	claudeMDPath := filepath.Join(projectRoot, "CLAUDE.md")
	mcpJSONPath := filepath.Join(projectRoot, ".mcp.json")

	cfg := &config.Config{
		StorageDir:      storageDir,
		IncludePatterns: config.DefaultProjectIncludePatterns,
		ExcludePatterns: config.DefaultProjectExcludePatterns,
		Scope: &config.Scope{
			Type:         config.ScopeProject,
			DotClaudeDir: dotClaudeDir,
			ExtraFiles: []config.ExtraFile{
				{Path: claudeMDPath, Tag: "claudemd"},
				{Path: mcpJSONPath, Tag: "mcpjson"},
			},
			StorageDir:      storageDir,
			IncludePatterns: config.DefaultProjectIncludePatterns,
			ExcludePatterns: config.DefaultProjectExcludePatterns,
		},
	}

	// Context A: has CLAUDE.md, .mcp.json, and settings.json
	os.WriteFile(filepath.Join(dotClaudeDir, "settings.json"), []byte(`{"ctx":"a"}`), 0644)
	os.WriteFile(claudeMDPath, []byte("# Context A"), 0644)
	os.WriteFile(mcpJSONPath, []byte(`{"mcpServers":{"a":{"command":"node"}}}`), 0644)

	_, err := Save(SaveOptions{Name: "mcp-a", Config: cfg})
	if err != nil {
		t.Fatalf("Save mcp-a failed: %v", err)
	}

	// Context B: different .mcp.json
	os.WriteFile(filepath.Join(dotClaudeDir, "settings.json"), []byte(`{"ctx":"b"}`), 0644)
	os.WriteFile(claudeMDPath, []byte("# Context B"), 0644)
	os.WriteFile(mcpJSONPath, []byte(`{"mcpServers":{"b":{"command":"python"}}}`), 0644)

	_, err = Save(SaveOptions{Name: "mcp-b", Config: cfg})
	if err != nil {
		t.Fatalf("Save mcp-b failed: %v", err)
	}

	// Switch to mcp-a — should restore context A's .mcp.json
	_, err = Restore(RestoreOptions{Name: "mcp-a", Config: cfg})
	if err != nil {
		t.Fatalf("Restore mcp-a failed: %v", err)
	}

	data, _ := os.ReadFile(mcpJSONPath)
	if string(data) != `{"mcpServers":{"a":{"command":"node"}}}` {
		t.Errorf("expected mcp-a .mcp.json, got %q", data)
	}

	data, _ = os.ReadFile(claudeMDPath)
	if string(data) != "# Context A" {
		t.Errorf("expected mcp-a CLAUDE.md, got %q", data)
	}

	// Switch back to mcp-b
	_, err = Restore(RestoreOptions{Name: "mcp-b", Config: cfg})
	if err != nil {
		t.Fatalf("Restore mcp-b failed: %v", err)
	}

	data, _ = os.ReadFile(mcpJSONPath)
	if string(data) != `{"mcpServers":{"b":{"command":"python"}}}` {
		t.Errorf("expected mcp-b .mcp.json, got %q", data)
	}
}

// TestFromScratchClearsAndCreatesEmpty verifies that creating an empty manifest
// then clearing managed files leaves a clean slate with no managed files.
func TestFromScratchClearsAndCreatesEmpty(t *testing.T) {
	projectRoot := t.TempDir()
	storageDir := filepath.Join(projectRoot, ".claudectx")
	dotClaudeDir := filepath.Join(projectRoot, ".claude")
	os.MkdirAll(dotClaudeDir, 0755)

	claudeMDPath := filepath.Join(projectRoot, "CLAUDE.md")

	cfg := &config.Config{
		StorageDir:      storageDir,
		IncludePatterns: config.DefaultProjectIncludePatterns,
		ExcludePatterns: config.DefaultProjectExcludePatterns,
		Scope: &config.Scope{
			Type:         config.ScopeProject,
			DotClaudeDir: dotClaudeDir,
			ExtraFiles: []config.ExtraFile{
				{Path: claudeMDPath, Tag: "claudemd"},
			},
			StorageDir:      storageDir,
			IncludePatterns: config.DefaultProjectIncludePatterns,
			ExcludePatterns: config.DefaultProjectExcludePatterns,
		},
	}

	// Create initial context with files
	os.WriteFile(filepath.Join(dotClaudeDir, "settings.json"), []byte(`{"key":"val"}`), 0644)
	os.WriteFile(claudeMDPath, []byte("# Old Context"), 0644)

	_, err := Save(SaveOptions{Name: "old-ctx", Config: cfg})
	if err != nil {
		t.Fatalf("Save old-ctx failed: %v", err)
	}

	// Simulate from-scratch: auto-save current, create empty manifest, clear files
	AutoSaveCurrent(cfg, "clean-ctx")

	contextDir := filepath.Join(cfg.ContextsDir(), "clean-ctx")
	os.MkdirAll(filepath.Join(contextDir, "dotclaude"), 0755)
	manifest := &Manifest{
		Name:  "clean-ctx",
		Scope: "project",
		Files: []FileEntry{},
	}
	if err := WriteManifest(contextDir, manifest); err != nil {
		t.Fatalf("WriteManifest failed: %v", err)
	}

	if err := ClearManagedFiles(cfg); err != nil {
		t.Fatalf("ClearManagedFiles failed: %v", err)
	}

	if err := SetCurrent(cfg, "clean-ctx"); err != nil {
		t.Fatalf("SetCurrent failed: %v", err)
	}

	// CLAUDE.md should be gone
	if _, err := os.Stat(claudeMDPath); !os.IsNotExist(err) {
		t.Errorf("expected CLAUDE.md to be removed after from-scratch")
	}

	// settings.json should be gone
	if _, err := os.Stat(filepath.Join(dotClaudeDir, "settings.json")); !os.IsNotExist(err) {
		t.Errorf("expected settings.json to be removed after from-scratch")
	}

	// .claude/ dir should still exist
	if _, err := os.Stat(dotClaudeDir); err != nil {
		t.Errorf("expected .claude/ directory to still exist")
	}

	// old-ctx snapshot should still have original content (from auto-save)
	oldDir := filepath.Join(cfg.ContextsDir(), "old-ctx")
	oldManifest, err := ReadManifest(oldDir)
	if err != nil {
		t.Fatalf("ReadManifest old-ctx failed: %v", err)
	}
	if len(oldManifest.Files) == 0 {
		t.Errorf("expected old-ctx to have files in snapshot")
	}

	// Switch back to old-ctx should restore files
	_, err = Restore(RestoreOptions{Name: "old-ctx", Config: cfg})
	if err != nil {
		t.Fatalf("Restore old-ctx failed: %v", err)
	}

	data, _ := os.ReadFile(claudeMDPath)
	if string(data) != "# Old Context" {
		t.Errorf("expected old-ctx CLAUDE.md content, got %q", data)
	}
}

// TestCopyFromCreatesIndependentContext verifies that copying a context
// creates an independent snapshot that doesn't affect the original.
func TestCopyFromCreatesIndependentContext(t *testing.T) {
	projectRoot := t.TempDir()
	storageDir := filepath.Join(projectRoot, ".claudectx")
	dotClaudeDir := filepath.Join(projectRoot, ".claude")
	os.MkdirAll(dotClaudeDir, 0755)

	claudeMDPath := filepath.Join(projectRoot, "CLAUDE.md")

	cfg := &config.Config{
		StorageDir:      storageDir,
		IncludePatterns: config.DefaultProjectIncludePatterns,
		ExcludePatterns: config.DefaultProjectExcludePatterns,
		Scope: &config.Scope{
			Type:         config.ScopeProject,
			DotClaudeDir: dotClaudeDir,
			ExtraFiles: []config.ExtraFile{
				{Path: claudeMDPath, Tag: "claudemd"},
			},
			StorageDir:      storageDir,
			IncludePatterns: config.DefaultProjectIncludePatterns,
			ExcludePatterns: config.DefaultProjectExcludePatterns,
		},
	}

	// Create source context
	os.WriteFile(filepath.Join(dotClaudeDir, "settings.json"), []byte(`{"source":true}`), 0644)
	os.WriteFile(claudeMDPath, []byte("# Source Context"), 0644)

	_, err := Save(SaveOptions{Name: "source", Config: cfg})
	if err != nil {
		t.Fatalf("Save source failed: %v", err)
	}

	// Copy context using CopyDir + manifest update (simulating --copy-from)
	srcDir := filepath.Join(cfg.ContextsDir(), "source")
	dstDir := filepath.Join(cfg.ContextsDir(), "copy")

	if err := copyDir(srcDir, dstDir); err != nil {
		t.Fatalf("CopyDir failed: %v", err)
	}

	copyManifest, err := ReadManifest(dstDir)
	if err != nil {
		t.Fatalf("ReadManifest copy failed: %v", err)
	}
	copyManifest.Name = "copy"
	if err := WriteManifest(dstDir, copyManifest); err != nil {
		t.Fatalf("WriteManifest copy failed: %v", err)
	}

	// Switch to copy
	_, err = Restore(RestoreOptions{Name: "copy", Force: true, Config: cfg})
	if err != nil {
		t.Fatalf("Restore copy failed: %v", err)
	}

	// Verify copy has source content
	data, _ := os.ReadFile(claudeMDPath)
	if string(data) != "# Source Context" {
		t.Errorf("expected source content in copy, got %q", data)
	}

	// Modify copy's live state
	os.WriteFile(claudeMDPath, []byte("# Modified Copy"), 0644)

	// Switch to source — auto-saves copy, restores source
	_, err = Restore(RestoreOptions{Name: "source", Config: cfg})
	if err != nil {
		t.Fatalf("Restore source failed: %v", err)
	}

	// Source should have original content
	data, _ = os.ReadFile(claudeMDPath)
	if string(data) != "# Source Context" {
		t.Errorf("expected source content, got %q", data)
	}

	// Switch back to copy — should have modified content (auto-saved)
	_, err = Restore(RestoreOptions{Name: "copy", Config: cfg})
	if err != nil {
		t.Fatalf("Restore copy failed: %v", err)
	}

	data, _ = os.ReadFile(claudeMDPath)
	if string(data) != "# Modified Copy" {
		t.Errorf("expected modified copy content, got %q", data)
	}

	// Source snapshot should be unaffected by copy modifications
	sourceManifest, _ := ReadManifest(srcDir)
	if sourceManifest.Name != "source" {
		t.Errorf("expected source manifest name unchanged, got %q", sourceManifest.Name)
	}
}

// TestSaveDuplicateNameErrors verifies that Save returns an error when
// a context with the same name already exists and Overwrite is false.
func TestSaveDuplicateNameErrors(t *testing.T) {
	projectRoot := t.TempDir()
	storageDir := filepath.Join(projectRoot, ".claudectx")
	dotClaudeDir := filepath.Join(projectRoot, ".claude")
	os.MkdirAll(dotClaudeDir, 0755)

	claudeMDPath := filepath.Join(projectRoot, "CLAUDE.md")
	os.WriteFile(claudeMDPath, []byte("# Hello"), 0644)

	cfg := &config.Config{
		StorageDir:      storageDir,
		IncludePatterns: config.DefaultProjectIncludePatterns,
		ExcludePatterns: config.DefaultProjectExcludePatterns,
		Scope: &config.Scope{
			Type:         config.ScopeProject,
			DotClaudeDir: dotClaudeDir,
			ExtraFiles: []config.ExtraFile{
				{Path: claudeMDPath, Tag: "claudemd"},
			},
			StorageDir:      storageDir,
			IncludePatterns: config.DefaultProjectIncludePatterns,
			ExcludePatterns: config.DefaultProjectExcludePatterns,
		},
	}

	_, err := Save(SaveOptions{Name: "dup", Config: cfg})
	if err != nil {
		t.Fatalf("first Save failed: %v", err)
	}

	// Second save with same name should fail
	_, err = Save(SaveOptions{Name: "dup", Config: cfg})
	if err == nil {
		t.Fatal("expected error on duplicate save, got nil")
	}

	// With Overwrite, it should succeed
	_, err = Save(SaveOptions{Name: "dup", Overwrite: true, Config: cfg})
	if err != nil {
		t.Fatalf("overwrite Save failed: %v", err)
	}
}

// copyDir is a test helper that recursively copies a directory.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, data, info.Mode())
	})
}
