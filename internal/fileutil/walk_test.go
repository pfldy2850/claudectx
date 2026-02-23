package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWalkFiltered(t *testing.T) {
	root := t.TempDir()

	// Create a fake .claude directory structure
	files := map[string]string{
		"settings.json":                       `{"theme":"dark"}`,
		"settings.local.json":                 `{"local":true}`,
		"debug/log.txt":                       "debug data",
		"plugins/blocklist.json":              `[]`,
		"plugins/cache/big.json":              "cached",
		"projects/myproj/memory/MEMORY.md":    "# Memory",
		"projects/myproj/session.jsonl":       "{}",
		"projects/myproj/memory/patterns.md":  "patterns",
	}

	for relPath, content := range files {
		absPath := filepath.Join(root, relPath)
		os.MkdirAll(filepath.Dir(absPath), 0755)
		os.WriteFile(absPath, []byte(content), 0644)
	}

	includes := []string{
		"settings.json",
		"settings.local.json",
		"plugins/blocklist.json",
		"projects/*/memory/**",
	}
	excludes := []string{
		"debug/**",
		"plugins/cache/**",
		"projects/*/*.jsonl",
	}

	results, err := WalkFiltered(root, includes, excludes)
	if err != nil {
		t.Fatal(err)
	}

	expected := map[string]bool{
		"settings.json":                      true,
		"settings.local.json":                true,
		"plugins/blocklist.json":             true,
		"projects/myproj/memory/MEMORY.md":   true,
		"projects/myproj/memory/patterns.md": true,
	}

	if len(results) != len(expected) {
		var paths []string
		for _, r := range results {
			paths = append(paths, r.RelPath)
		}
		t.Fatalf("got %d results %v, want %d", len(results), paths, len(expected))
	}

	for _, r := range results {
		if !expected[r.RelPath] {
			t.Errorf("unexpected file: %s", r.RelPath)
		}
	}
}
