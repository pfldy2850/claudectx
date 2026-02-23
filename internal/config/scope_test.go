package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUserScope(t *testing.T) {
	scope, err := UserScope()
	if err != nil {
		t.Fatal(err)
	}
	if scope.Type != ScopeUser {
		t.Errorf("expected user scope, got %s", scope.Type)
	}
	if len(scope.ExtraFiles) != 1 {
		t.Fatalf("expected 1 extra file, got %d", len(scope.ExtraFiles))
	}
	if scope.ExtraFiles[0].Tag != "claudejson" {
		t.Errorf("expected claudejson tag, got %s", scope.ExtraFiles[0].Tag)
	}
	if filepath.Base(scope.ExtraFiles[0].Path) != ".claude.json" {
		t.Errorf("expected .claude.json, got %s", filepath.Base(scope.ExtraFiles[0].Path))
	}
}

func TestProjectScope(t *testing.T) {
	// Create a temp dir with .git
	tmp := t.TempDir()
	os.Mkdir(filepath.Join(tmp, ".git"), 0755)

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(tmp)

	scope, err := ProjectScope()
	if err != nil {
		t.Fatal(err)
	}
	if scope.Type != ScopeProject {
		t.Errorf("expected project scope, got %s", scope.Type)
	}
	if len(scope.ExtraFiles) != 2 {
		t.Fatalf("expected 2 extra files, got %d", len(scope.ExtraFiles))
	}

	claudemd := scope.ExtraFileByTag("claudemd")
	if claudemd == nil {
		t.Fatal("expected claudemd extra file")
	}
	if filepath.Base(claudemd.Path) != "CLAUDE.md" {
		t.Errorf("expected CLAUDE.md, got %s", filepath.Base(claudemd.Path))
	}

	mcpjson := scope.ExtraFileByTag("mcpjson")
	if mcpjson == nil {
		t.Fatal("expected mcpjson extra file")
	}
	if filepath.Base(mcpjson.Path) != ".mcp.json" {
		t.Errorf("expected .mcp.json, got %s", filepath.Base(mcpjson.Path))
	}

	if filepath.Base(scope.StorageDir) != ".claudectx" {
		t.Errorf("expected .claudectx storage dir, got %s", filepath.Base(scope.StorageDir))
	}
}

func TestExtraFileByTag(t *testing.T) {
	scope := &Scope{
		ExtraFiles: []ExtraFile{
			{Path: "/a/CLAUDE.md", Tag: "claudemd"},
			{Path: "/a/.mcp.json", Tag: "mcpjson"},
		},
	}

	if ef := scope.ExtraFileByTag("claudemd"); ef == nil || ef.Path != "/a/CLAUDE.md" {
		t.Errorf("expected to find claudemd extra file")
	}
	if ef := scope.ExtraFileByTag("mcpjson"); ef == nil || ef.Path != "/a/.mcp.json" {
		t.Errorf("expected to find mcpjson extra file")
	}
	if ef := scope.ExtraFileByTag("nonexistent"); ef != nil {
		t.Errorf("expected nil for nonexistent tag, got %v", ef)
	}
}

func TestProjectScopeAt(t *testing.T) {
	root := "/tmp/my-project"
	scope := ProjectScopeAt(root)

	if scope.Type != ScopeProject {
		t.Errorf("expected project scope, got %s", scope.Type)
	}
	if scope.DotClaudeDir != filepath.Join(root, ".claude") {
		t.Errorf("expected DotClaudeDir %s, got %s", filepath.Join(root, ".claude"), scope.DotClaudeDir)
	}
	if scope.StorageDir != filepath.Join(root, ".claudectx") {
		t.Errorf("expected StorageDir %s, got %s", filepath.Join(root, ".claudectx"), scope.StorageDir)
	}
	if len(scope.ExtraFiles) != 2 {
		t.Fatalf("expected 2 extra files, got %d", len(scope.ExtraFiles))
	}
	if scope.ExtraFileByTag("claudemd") == nil {
		t.Error("expected claudemd extra file")
	}
	if scope.ExtraFileByTag("mcpjson") == nil {
		t.Error("expected mcpjson extra file")
	}
}

func TestDetectProjectRoot(t *testing.T) {
	t.Run("marker only (no git)", func(t *testing.T) {
		tmp := t.TempDir()
		tmp, _ = filepath.EvalSymlinks(tmp)
		os.Mkdir(filepath.Join(tmp, ".claude"), 0755)

		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		os.Chdir(tmp)

		root, err := DetectProjectRoot()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if root != tmp {
			t.Errorf("got %s, want %s", root, tmp)
		}
	})

	t.Run("git only (no markers)", func(t *testing.T) {
		tmp := t.TempDir()
		tmp, _ = filepath.EvalSymlinks(tmp)
		os.Mkdir(filepath.Join(tmp, ".git"), 0755)

		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		os.Chdir(tmp)

		root, err := DetectProjectRoot()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if root != tmp {
			t.Errorf("got %s, want %s", root, tmp)
		}
	})

	t.Run("both marker and git", func(t *testing.T) {
		tmp := t.TempDir()
		tmp, _ = filepath.EvalSymlinks(tmp)
		os.Mkdir(filepath.Join(tmp, ".git"), 0755)
		os.Mkdir(filepath.Join(tmp, ".claude"), 0755)

		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		os.Chdir(tmp)

		root, err := DetectProjectRoot()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if root != tmp {
			t.Errorf("got %s, want %s", root, tmp)
		}
	})

	t.Run("neither marker nor git", func(t *testing.T) {
		tmp := t.TempDir()

		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		os.Chdir(tmp)

		_, err := DetectProjectRoot()
		if err == nil {
			t.Fatal("expected error when no project detected")
		}
	})
}

func TestResolveScopeWithRoot(t *testing.T) {
	t.Run("explicit root", func(t *testing.T) {
		tmp := t.TempDir()
		tmp, _ = filepath.EvalSymlinks(tmp)

		scope, err := ResolveScopeWithRoot("", tmp)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if scope.Type != ScopeProject {
			t.Errorf("expected project scope, got %s", scope.Type)
		}
		if scope.StorageDir != filepath.Join(tmp, ".claudectx") {
			t.Errorf("expected storage dir at %s, got %s", filepath.Join(tmp, ".claudectx"), scope.StorageDir)
		}
	})

	t.Run("root with scope user is error", func(t *testing.T) {
		tmp := t.TempDir()

		_, err := ResolveScopeWithRoot("user", tmp)
		if err == nil {
			t.Fatal("expected error for --root with --scope user")
		}
	})

	t.Run("root non-existent is error", func(t *testing.T) {
		_, err := ResolveScopeWithRoot("", "/nonexistent/path/that/does/not/exist")
		if err == nil {
			t.Fatal("expected error for non-existent root")
		}
	})

	t.Run("explicit project with CWD fallback", func(t *testing.T) {
		tmp := t.TempDir()
		tmp, _ = filepath.EvalSymlinks(tmp)
		// No .git, no markers — should fall back to CWD

		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		os.Chdir(tmp)

		scope, err := ResolveScopeWithRoot("project", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if scope.Type != ScopeProject {
			t.Errorf("expected project scope, got %s", scope.Type)
		}
		if scope.StorageDir != filepath.Join(tmp, ".claudectx") {
			t.Errorf("expected CWD-based storage dir, got %s", scope.StorageDir)
		}
	})

	t.Run("auto-detect marker wins", func(t *testing.T) {
		tmp := t.TempDir()
		tmp, _ = filepath.EvalSymlinks(tmp)
		os.Mkdir(filepath.Join(tmp, ".claude"), 0755)

		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		os.Chdir(tmp)

		scope, err := ResolveScopeWithRoot("", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if scope.Type != ScopeProject {
			t.Errorf("expected project scope, got %s", scope.Type)
		}
	})

	t.Run("auto-detect user fallback", func(t *testing.T) {
		tmp := t.TempDir()
		// No markers, no git

		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		os.Chdir(tmp)

		scope, err := ResolveScopeWithRoot("", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if scope.Type != ScopeUser {
			t.Errorf("expected user fallback, got %s", scope.Type)
		}
	})
}

func TestResolveScope(t *testing.T) {
	t.Run("explicit user", func(t *testing.T) {
		scope, err := ResolveScope("user")
		if err != nil {
			t.Fatal(err)
		}
		if scope.Type != ScopeUser {
			t.Errorf("expected user, got %s", scope.Type)
		}
	})

	t.Run("explicit project in git repo", func(t *testing.T) {
		tmp := t.TempDir()
		os.Mkdir(filepath.Join(tmp, ".git"), 0755)

		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		os.Chdir(tmp)

		scope, err := ResolveScope("project")
		if err != nil {
			t.Fatal(err)
		}
		if scope.Type != ScopeProject {
			t.Errorf("expected project, got %s", scope.Type)
		}
	})

	t.Run("invalid scope", func(t *testing.T) {
		_, err := ResolveScope("invalid")
		if err == nil {
			t.Fatal("expected error for invalid scope")
		}
	})

	t.Run("auto-detect in git repo", func(t *testing.T) {
		tmp := t.TempDir()
		os.Mkdir(filepath.Join(tmp, ".git"), 0755)

		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)
		os.Chdir(tmp)

		scope, err := ResolveScope("")
		if err != nil {
			t.Fatal(err)
		}
		if scope.Type != ScopeProject {
			t.Errorf("expected auto-detected project scope, got %s", scope.Type)
		}
	})
}
