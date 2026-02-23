package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pfldy2850/claudectx/internal/claude"
	"github.com/pfldy2850/claudectx/internal/config"
)

func ensureGitignore(scope *config.Scope, isDryRun bool) {
	projectRoot := filepath.Dir(scope.DotClaudeDir) // .claude is at project root

	// Skip if not a git repository
	if !claude.IsGitRepo(projectRoot) {
		return
	}
	gitignorePath := filepath.Join(projectRoot, ".gitignore")

	data, err := os.ReadFile(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Warning: could not read .gitignore: %v\n", err)
		return
	}

	// Check if already present
	if err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == ".claudectx" || line == ".claudectx/" || line == ".claudectx/**" {
				return
			}
		}
	}

	if isDryRun {
		fmt.Println("[dry-run] Would add .claudectx/ to .gitignore")
		return
	}

	// Append .claudectx/ to .gitignore
	var content []byte
	if len(data) > 0 {
		content = append(content, data...)
		// Ensure trailing newline before appending
		if content[len(content)-1] != '\n' {
			content = append(content, '\n')
		}
	}
	content = append(content, []byte(".claudectx/\n")...)

	// Preserve existing file permissions
	mode := os.FileMode(0644)
	if info, statErr := os.Stat(gitignorePath); statErr == nil {
		mode = info.Mode()
	}

	if err := os.WriteFile(gitignorePath, content, mode); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not update .gitignore: %v\n", err)
		return
	}
	fmt.Println("Added .claudectx/ to .gitignore")
}

func formatSize(bytes int64) string {
	const (
		kb = 1024
		mb = 1024 * kb
	)
	switch {
	case bytes >= mb:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
