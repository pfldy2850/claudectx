package context

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pfldy2850/claudectx/internal/config"
)

// TestSaveModifyRestoreRoundTrip tests the full save → modify → restore workflow
// using isolated temp directories instead of real ~/.claude paths.
func TestSaveModifyRestoreRoundTrip(t *testing.T) {
	// Setup isolated environment
	tmpHome := t.TempDir()
	storageDir := filepath.Join(tmpHome, ".claudectx")

	// Create fake ~/.claude directory
	dotClaudeDir := filepath.Join(tmpHome, ".claude")
	os.MkdirAll(filepath.Join(dotClaudeDir, "projects", "myproj", "memory"), 0755)

	// Create fake ~/.claude.json
	claudeJSONPath := filepath.Join(tmpHome, ".claude.json")

	// Set initial state
	originalSettings := `{"theme":"dark","editor":"vim"}`
	originalClaudeJSON := `{"oauthEmail":"user@example.com","apiKey":"sk-test-123"}`
	originalMemory := "# Project Memory\nKey patterns here."

	os.WriteFile(filepath.Join(dotClaudeDir, "settings.json"), []byte(originalSettings), 0644)
	os.WriteFile(claudeJSONPath, []byte(originalClaudeJSON), 0644)
	os.WriteFile(filepath.Join(dotClaudeDir, "projects", "myproj", "memory", "MEMORY.md"), []byte(originalMemory), 0644)

	cfg := &config.Config{
		StorageDir:      storageDir,
		IncludePatterns: config.DefaultIncludePatterns,
		ExcludePatterns: config.DefaultExcludePatterns,
		Scope: &config.Scope{
			Type:         config.ScopeUser,
			DotClaudeDir: dotClaudeDir,
			ExtraFiles: []config.ExtraFile{
				{Path: claudeJSONPath, Tag: "claudejson"},
			},
			StorageDir:      storageDir,
			IncludePatterns: config.DefaultIncludePatterns,
			ExcludePatterns: config.DefaultExcludePatterns,
		},
	}

	// Override claude paths for testing by directly calling snapshot internals
	// We'll test using the low-level functions since Save/Restore use real home paths

	// 1. Save: manually create a context snapshot
	contextDir := filepath.Join(cfg.ContextsDir(), "test-ctx")
	os.MkdirAll(filepath.Join(contextDir, "dotclaude", "projects", "myproj", "memory"), 0755)

	// Copy files to snapshot
	copyTestFile(t, claudeJSONPath, filepath.Join(contextDir, "claude.json"))
	copyTestFile(t, filepath.Join(dotClaudeDir, "settings.json"), filepath.Join(contextDir, "dotclaude", "settings.json"))
	copyTestFile(t, filepath.Join(dotClaudeDir, "projects", "myproj", "memory", "MEMORY.md"),
		filepath.Join(contextDir, "dotclaude", "projects", "myproj", "memory", "MEMORY.md"))

	// Create manifest
	settingsSum, _ := FileChecksum(filepath.Join(dotClaudeDir, "settings.json"))
	claudeJSONSum, _ := FileChecksum(claudeJSONPath)
	memorySum, _ := FileChecksum(filepath.Join(dotClaudeDir, "projects", "myproj", "memory", "MEMORY.md"))

	files := []FileEntry{
		{RelPath: "claude.json", Size: int64(len(originalClaudeJSON)), Mode: 0644, Checksum: claudeJSONSum, Source: "claudejson"},
		{RelPath: "dotclaude/settings.json", Size: int64(len(originalSettings)), Mode: 0644, Checksum: settingsSum, Source: "dotclaude"},
		{RelPath: "dotclaude/projects/myproj/memory/MEMORY.md", Size: int64(len(originalMemory)), Mode: 0644, Checksum: memorySum, Source: "dotclaude"},
	}

	manifest := &Manifest{
		Name:       "test-ctx",
		Files:      files,
		TotalSize:  int64(len(originalSettings) + len(originalClaudeJSON) + len(originalMemory)),
		Checksum:   ManifestChecksum(files),
		OAuthEmail: "user@example.com",
		Scope:      "user",
	}
	if err := WriteManifest(contextDir, manifest); err != nil {
		t.Fatal(err)
	}

	// 2. Verify context exists
	names, err := ListContexts(cfg.ContextsDir())
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "test-ctx" {
		t.Fatalf("expected [test-ctx], got %v", names)
	}

	// 3. Modify current state
	modifiedSettings := `{"theme":"light","editor":"code"}`
	os.WriteFile(filepath.Join(dotClaudeDir, "settings.json"), []byte(modifiedSettings), 0644)

	// 4. Verify the file changed
	data, _ := os.ReadFile(filepath.Join(dotClaudeDir, "settings.json"))
	if string(data) != modifiedSettings {
		t.Fatal("settings.json not modified")
	}

	// 5. Restore from snapshot (manually copy files back)
	readManifest, err := ReadManifest(contextDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range readManifest.Files {
		srcPath := filepath.Join(contextDir, entry.RelPath)
		var dstPath string
		if entry.Source == "claudejson" {
			dstPath = claudeJSONPath
		} else {
			relToDotClaude := entry.RelPath[len("dotclaude/"):]
			dstPath = filepath.Join(dotClaudeDir, relToDotClaude)
		}

		srcData, err := os.ReadFile(srcPath)
		if err != nil {
			t.Fatalf("read snapshot file %s: %v", entry.RelPath, err)
		}
		os.MkdirAll(filepath.Dir(dstPath), 0755)
		if err := os.WriteFile(dstPath, srcData, os.FileMode(entry.Mode)); err != nil {
			t.Fatalf("write restored file %s: %v", dstPath, err)
		}
	}

	// 6. Verify original state restored
	data, _ = os.ReadFile(filepath.Join(dotClaudeDir, "settings.json"))
	if string(data) != originalSettings {
		t.Errorf("settings not restored: got %q, want %q", data, originalSettings)
	}

	data, _ = os.ReadFile(claudeJSONPath)
	if string(data) != originalClaudeJSON {
		t.Errorf("claude.json not restored: got %q, want %q", data, originalClaudeJSON)
	}

	data, _ = os.ReadFile(filepath.Join(dotClaudeDir, "projects", "myproj", "memory", "MEMORY.md"))
	if string(data) != originalMemory {
		t.Errorf("memory not restored: got %q, want %q", data, originalMemory)
	}

	// 7. Delete context
	if err := DeleteContext(cfg.ContextsDir(), "test-ctx"); err != nil {
		t.Fatal(err)
	}
	names, _ = ListContexts(cfg.ContextsDir())
	if len(names) != 0 {
		t.Errorf("expected no contexts after delete, got %v", names)
	}
}

// TestProjectScopeRoundTrip tests save → modify → restore for project scope
// using isolated temp directories simulating a git project root.
func TestProjectScopeRoundTrip(t *testing.T) {
	// Setup isolated project root
	projectRoot := t.TempDir()
	storageDir := filepath.Join(projectRoot, ".claudectx")

	// Create fake .claude/ directory at project root
	dotClaudeDir := filepath.Join(projectRoot, ".claude")
	os.MkdirAll(dotClaudeDir, 0755)

	// Create fake CLAUDE.md at project root
	claudeMDPath := filepath.Join(projectRoot, "CLAUDE.md")

	// Set initial state
	originalSettings := `{"projectSettings": true}`
	originalClaudeMD := "# CLAUDE.md\nProject instructions here."

	os.WriteFile(filepath.Join(dotClaudeDir, "settings.json"), []byte(originalSettings), 0644)
	os.WriteFile(claudeMDPath, []byte(originalClaudeMD), 0644)

	// Create fake .mcp.json at project root
	mcpJSONPath := filepath.Join(projectRoot, ".mcp.json")

	// Set initial state
	originalMcpJSON := `{"mcpServers":{"server1":{"command":"node","args":["server.js"]}}}`
	os.WriteFile(mcpJSONPath, []byte(originalMcpJSON), 0644)

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

	// 1. Save using Save() with the project scope config
	result, err := Save(SaveOptions{
		Name:   "proj-ctx",
		Config: cfg,
	})
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if result.Files != 3 {
		t.Fatalf("expected 3 files saved (settings.json + CLAUDE.md + .mcp.json), got %d", result.Files)
	}

	// 2. Verify manifest has project scope
	contextDir := filepath.Join(cfg.ContextsDir(), "proj-ctx")
	m, err := ReadManifest(contextDir)
	if err != nil {
		t.Fatal(err)
	}
	if m.Scope != "project" {
		t.Errorf("expected scope 'project', got %q", m.Scope)
	}
	if m.OAuthEmail != "" {
		t.Errorf("expected no OAuth email for project scope, got %q", m.OAuthEmail)
	}

	// Verify CLAUDE.md is stored with claudemd source tag
	foundClaudeMD := false
	foundMcpJSON := false
	for _, f := range m.Files {
		if f.Source == "claudemd" {
			foundClaudeMD = true
			if f.RelPath != "CLAUDE.md" {
				t.Errorf("expected relPath 'CLAUDE.md', got %q", f.RelPath)
			}
		}
		if f.Source == "mcpjson" {
			foundMcpJSON = true
			if f.RelPath != ".mcp.json" {
				t.Errorf("expected relPath '.mcp.json', got %q", f.RelPath)
			}
		}
	}
	if !foundClaudeMD {
		t.Error("expected a file with source 'claudemd'")
	}
	if !foundMcpJSON {
		t.Error("expected a file with source 'mcpjson'")
	}

	// 3. Modify current state
	modifiedClaudeMD := "# CLAUDE.md\nModified instructions."
	os.WriteFile(claudeMDPath, []byte(modifiedClaudeMD), 0644)
	modifiedMcpJSON := `{"mcpServers":{"server2":{"command":"python","args":["srv.py"]}}}`
	os.WriteFile(mcpJSONPath, []byte(modifiedMcpJSON), 0644)

	// 4. Restore from snapshot
	restoreResult, err := Restore(RestoreOptions{
		Name:   "proj-ctx",
		Config: cfg,
	})
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}
	if restoreResult.FilesRestored != 3 {
		t.Errorf("expected 3 files restored, got %d", restoreResult.FilesRestored)
	}

	// 5. Verify original state restored
	data, _ := os.ReadFile(claudeMDPath)
	if string(data) != originalClaudeMD {
		t.Errorf("CLAUDE.md not restored: got %q, want %q", data, originalClaudeMD)
	}

	data, _ = os.ReadFile(mcpJSONPath)
	if string(data) != originalMcpJSON {
		t.Errorf(".mcp.json not restored: got %q, want %q", data, originalMcpJSON)
	}

	data, _ = os.ReadFile(filepath.Join(dotClaudeDir, "settings.json"))
	if string(data) != originalSettings {
		t.Errorf("settings not restored: got %q, want %q", data, originalSettings)
	}

	// 6. Delete context
	if err := DeleteContext(cfg.ContextsDir(), "proj-ctx"); err != nil {
		t.Fatal(err)
	}
}

// TestCrossScopeRestoreBlocked verifies that restoring a context saved with
// one scope into a different scope produces an error.
func TestCrossScopeRestoreBlocked(t *testing.T) {
	tmpDir := t.TempDir()
	storageDir := filepath.Join(tmpDir, ".claudectx")

	// Create a context with "project" scope manifest
	contextDir := filepath.Join(storageDir, "contexts", "proj-ctx")
	os.MkdirAll(contextDir, 0755)

	manifest := &Manifest{
		Name:  "proj-ctx",
		Scope: "project",
		Files: []FileEntry{},
	}
	if err := WriteManifest(contextDir, manifest); err != nil {
		t.Fatal(err)
	}

	// Try to restore it with a user-scope config
	userCfg := &config.Config{
		StorageDir:      storageDir,
		IncludePatterns: config.DefaultIncludePatterns,
		ExcludePatterns: config.DefaultExcludePatterns,
		Scope: &config.Scope{
			Type:         config.ScopeUser,
			DotClaudeDir: filepath.Join(tmpDir, ".claude"),
			ExtraFiles: []config.ExtraFile{
				{Path: filepath.Join(tmpDir, ".claude.json"), Tag: "claudejson"},
			},
			StorageDir:      storageDir,
			IncludePatterns: config.DefaultIncludePatterns,
			ExcludePatterns: config.DefaultExcludePatterns,
		},
	}

	_, err := Restore(RestoreOptions{
		Name:   "proj-ctx",
		Config: userCfg,
	})
	if err == nil {
		t.Fatal("expected error when restoring project-scope context with user scope")
	}
	if !strings.Contains(err.Error(), "scope") {
		t.Errorf("expected scope mismatch error, got: %v", err)
	}
}

func copyTestFile(t *testing.T, src, dst string) {
	t.Helper()
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	os.MkdirAll(filepath.Dir(dst), 0755)
	if err := os.WriteFile(dst, data, 0644); err != nil {
		t.Fatal(err)
	}
}
