package claude

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDotClaudeDir(t *testing.T) {
	dir, err := DotClaudeDir()
	if err != nil {
		t.Fatal(err)
	}
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".claude")
	if dir != expected {
		t.Errorf("got %s, want %s", dir, expected)
	}
}

func TestClaudeJSONPath(t *testing.T) {
	p, err := ClaudeJSONPath()
	if err != nil {
		t.Fatal(err)
	}
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".claude.json")
	if p != expected {
		t.Errorf("got %s, want %s", p, expected)
	}
}

func TestFindMarkerRootFrom(t *testing.T) {
	t.Run("finds .claude dir", func(t *testing.T) {
		tmp := t.TempDir()
		tmp, _ = filepath.EvalSymlinks(tmp)

		os.Mkdir(filepath.Join(tmp, ".claude"), 0755)
		subDir := filepath.Join(tmp, "a", "b")
		os.MkdirAll(subDir, 0755)

		root, err := FindMarkerRootFrom(subDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if root != tmp {
			t.Errorf("got %s, want %s", root, tmp)
		}
	})

	t.Run("finds CLAUDE.md file", func(t *testing.T) {
		tmp := t.TempDir()
		tmp, _ = filepath.EvalSymlinks(tmp)

		os.WriteFile(filepath.Join(tmp, "CLAUDE.md"), []byte("# test"), 0644)
		subDir := filepath.Join(tmp, "sub")
		os.MkdirAll(subDir, 0755)

		root, err := FindMarkerRootFrom(subDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if root != tmp {
			t.Errorf("got %s, want %s", root, tmp)
		}
	})

	t.Run("finds .claudectx dir", func(t *testing.T) {
		tmp := t.TempDir()
		tmp, _ = filepath.EvalSymlinks(tmp)

		os.Mkdir(filepath.Join(tmp, ".claudectx"), 0755)

		root, err := FindMarkerRootFrom(tmp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if root != tmp {
			t.Errorf("got %s, want %s", root, tmp)
		}
	})

	t.Run("no markers returns error", func(t *testing.T) {
		tmp := t.TempDir()

		_, err := FindMarkerRootFrom(tmp)
		if err == nil {
			t.Fatal("expected error when no markers found")
		}
	})

	t.Run("closest ancestor wins", func(t *testing.T) {
		tmp := t.TempDir()
		tmp, _ = filepath.EvalSymlinks(tmp)

		// Parent has .claude
		os.Mkdir(filepath.Join(tmp, ".claude"), 0755)
		// Child also has .claude
		child := filepath.Join(tmp, "child")
		os.MkdirAll(filepath.Join(child, ".claude"), 0755)
		grandchild := filepath.Join(child, "deep")
		os.MkdirAll(grandchild, 0755)

		root, err := FindMarkerRootFrom(grandchild)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should find child (closest), not tmp (parent)
		if root != child {
			t.Errorf("got %s, want %s (closest ancestor)", root, child)
		}
	})
}

func TestFindMarkerRoot(t *testing.T) {
	tmp := t.TempDir()
	tmp, _ = filepath.EvalSymlinks(tmp)
	os.Mkdir(filepath.Join(tmp, ".claude"), 0755)

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmp)

	root, err := FindMarkerRoot()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if root != tmp {
		t.Errorf("got %s, want %s", root, tmp)
	}
}

func TestIsGitRepo(t *testing.T) {
	t.Run("true with .git dir", func(t *testing.T) {
		tmp := t.TempDir()
		os.Mkdir(filepath.Join(tmp, ".git"), 0755)

		if !IsGitRepo(tmp) {
			t.Error("expected true for directory with .git/")
		}
	})

	t.Run("false without .git dir", func(t *testing.T) {
		tmp := t.TempDir()

		if IsGitRepo(tmp) {
			t.Error("expected false for directory without .git/")
		}
	})
}

func TestProjectRoot(t *testing.T) {
	t.Run("finds git root", func(t *testing.T) {
		tmp := t.TempDir()
		// Resolve symlinks (macOS /var -> /private/var)
		tmp, _ = filepath.EvalSymlinks(tmp)

		// Create nested dirs with a .git at root
		gitDir := filepath.Join(tmp, ".git")
		os.Mkdir(gitDir, 0755)
		subDir := filepath.Join(tmp, "a", "b", "c")
		os.MkdirAll(subDir, 0755)

		// Change to nested dir
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		os.Chdir(subDir)

		root, err := ProjectRoot()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if root != tmp {
			t.Errorf("got %s, want %s", root, tmp)
		}
	})

	t.Run("error when no git dir", func(t *testing.T) {
		tmp := t.TempDir()

		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		os.Chdir(tmp)

		_, err := ProjectRoot()
		if err == nil {
			t.Fatal("expected error for non-git directory")
		}
	})
}
