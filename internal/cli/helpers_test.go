package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pfldy2850/claudectx/internal/config"
)

// setupGitDir creates a .git directory so ensureGitignore doesn't skip.
func setupGitDir(t *testing.T, dir string) {
	t.Helper()
	os.Mkdir(filepath.Join(dir, ".git"), 0755)
}

func TestEnsureGitignore_CreatesNewFile(t *testing.T) {
	tmpDir := t.TempDir()
	setupGitDir(t, tmpDir)
	scope := &config.Scope{
		DotClaudeDir: filepath.Join(tmpDir, ".claude"),
	}

	ensureGitignore(scope, false)

	data, err := os.ReadFile(filepath.Join(tmpDir, ".gitignore"))
	if err != nil {
		t.Fatalf("expected .gitignore to be created: %v", err)
	}
	if strings.TrimSpace(string(data)) != ".claudectx/" {
		t.Errorf("expected '.claudectx/', got %q", string(data))
	}
}

func TestEnsureGitignore_AppendsToExisting(t *testing.T) {
	tmpDir := t.TempDir()
	setupGitDir(t, tmpDir)
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	os.WriteFile(gitignorePath, []byte("node_modules/\n.env\n"), 0644)

	scope := &config.Scope{
		DotClaudeDir: filepath.Join(tmpDir, ".claude"),
	}

	ensureGitignore(scope, false)

	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, ".claudectx/") {
		t.Errorf("expected .claudectx/ to be appended, got %q", content)
	}
	if !strings.HasPrefix(content, "node_modules/\n.env\n") {
		t.Errorf("expected existing content preserved, got %q", content)
	}
}

func TestEnsureGitignore_AppendsToExistingWithoutTrailingNewline(t *testing.T) {
	tmpDir := t.TempDir()
	setupGitDir(t, tmpDir)
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	os.WriteFile(gitignorePath, []byte("node_modules/\n.env"), 0644)

	scope := &config.Scope{
		DotClaudeDir: filepath.Join(tmpDir, ".claude"),
	}

	ensureGitignore(scope, false)

	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if content != "node_modules/\n.env\n.claudectx/\n" {
		t.Errorf("expected newline added before .claudectx/, got %q", content)
	}
}

func TestEnsureGitignore_SkipsIfAlreadyPresent(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"exact match", "node_modules/\n.claudectx/\n"},
		{"without slash", "node_modules/\n.claudectx\n"},
		{"with glob", "node_modules/\n.claudectx/**\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			setupGitDir(t, tmpDir)
			gitignorePath := filepath.Join(tmpDir, ".gitignore")
			os.WriteFile(gitignorePath, []byte(tt.content), 0644)

			scope := &config.Scope{
				DotClaudeDir: filepath.Join(tmpDir, ".claude"),
			}

			ensureGitignore(scope, false)

			data, _ := os.ReadFile(gitignorePath)
			if string(data) != tt.content {
				t.Errorf("expected content unchanged %q, got %q", tt.content, string(data))
			}
		})
	}
}

func TestEnsureGitignore_DryRun(t *testing.T) {
	tmpDir := t.TempDir()
	setupGitDir(t, tmpDir)
	scope := &config.Scope{
		DotClaudeDir: filepath.Join(tmpDir, ".claude"),
	}

	ensureGitignore(scope, true)

	// .gitignore should NOT be created in dry-run mode
	_, err := os.Stat(filepath.Join(tmpDir, ".gitignore"))
	if !os.IsNotExist(err) {
		t.Error("expected .gitignore to not be created in dry-run mode")
	}
}

func TestEnsureGitignore_DryRunExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	setupGitDir(t, tmpDir)
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	original := "node_modules/\n"
	os.WriteFile(gitignorePath, []byte(original), 0644)

	scope := &config.Scope{
		DotClaudeDir: filepath.Join(tmpDir, ".claude"),
	}

	ensureGitignore(scope, true)

	// .gitignore should remain unchanged in dry-run mode
	data, _ := os.ReadFile(gitignorePath)
	if string(data) != original {
		t.Errorf("expected content unchanged in dry-run, got %q", string(data))
	}
}

func TestEnsureGitignore_PreservesFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	setupGitDir(t, tmpDir)
	gitignorePath := filepath.Join(tmpDir, ".gitignore")
	os.WriteFile(gitignorePath, []byte("node_modules/\n"), 0600)

	scope := &config.Scope{
		DotClaudeDir: filepath.Join(tmpDir, ".claude"),
	}

	ensureGitignore(scope, false)

	info, err := os.Stat(gitignorePath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected permission 0600, got %o", info.Mode().Perm())
	}
}

func TestEnsureGitignore_SkipsNonGit(t *testing.T) {
	tmpDir := t.TempDir()
	// No .git directory — ensureGitignore should be a no-op

	scope := &config.Scope{
		DotClaudeDir: filepath.Join(tmpDir, ".claude"),
	}

	ensureGitignore(scope, false)

	// .gitignore should NOT be created
	_, err := os.Stat(filepath.Join(tmpDir, ".gitignore"))
	if !os.IsNotExist(err) {
		t.Error("expected .gitignore to not be created in non-git directory")
	}
}
