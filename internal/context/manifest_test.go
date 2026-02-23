package context

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWriteAndReadManifest(t *testing.T) {
	dir := t.TempDir()

	now := time.Now().Truncate(time.Second)
	manifest := &Manifest{
		Name:        "test-ctx",
		Description: "A test context",
		CreatedAt:   now,
		UpdatedAt:   now,
		Files: []FileEntry{
			{
				RelPath:  "claude.json",
				Size:     100,
				Mode:     0644,
				Checksum: "abc123",
				Source:   "claudejson",
			},
		},
		TotalSize:  100,
		Checksum:   "manifest-sum",
		OAuthEmail: "test@example.com",
	}

	if err := WriteManifest(dir, manifest); err != nil {
		t.Fatal(err)
	}

	// Verify file exists
	if _, err := os.Stat(filepath.Join(dir, "manifest.json")); err != nil {
		t.Fatal("manifest.json not created")
	}

	got, err := ReadManifest(dir)
	if err != nil {
		t.Fatal(err)
	}

	if got.Name != "test-ctx" {
		t.Errorf("got name %q, want %q", got.Name, "test-ctx")
	}
	if got.Description != "A test context" {
		t.Errorf("got description %q", got.Description)
	}
	if len(got.Files) != 1 {
		t.Fatalf("got %d files, want 1", len(got.Files))
	}
	if got.Files[0].RelPath != "claude.json" {
		t.Errorf("got relpath %q", got.Files[0].RelPath)
	}
	if got.OAuthEmail != "test@example.com" {
		t.Errorf("got email %q", got.OAuthEmail)
	}
}

func TestListContexts(t *testing.T) {
	dir := t.TempDir()

	// No contexts yet
	names, err := ListContexts(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 0 {
		t.Errorf("expected no contexts, got %d", len(names))
	}

	// Create a context with manifest
	ctxDir := filepath.Join(dir, "work")
	os.MkdirAll(ctxDir, 0755)
	now := time.Now()
	WriteManifest(ctxDir, &Manifest{Name: "work", CreatedAt: now, UpdatedAt: now})

	// Create a directory without manifest (should be ignored)
	os.MkdirAll(filepath.Join(dir, "orphan"), 0755)

	names, err = ListContexts(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "work" {
		t.Errorf("expected [work], got %v", names)
	}
}

func TestDeleteContext(t *testing.T) {
	dir := t.TempDir()
	ctxDir := filepath.Join(dir, "test")
	os.MkdirAll(ctxDir, 0755)
	now := time.Now()
	WriteManifest(ctxDir, &Manifest{Name: "test", CreatedAt: now, UpdatedAt: now})

	if err := DeleteContext(dir, "test"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(ctxDir); !os.IsNotExist(err) {
		t.Error("expected context to be deleted")
	}
}
