package context

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"
)

// Manifest holds metadata for a saved context.
type Manifest struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
	Files       []FileEntry `json:"files"`
	TotalSize   int64       `json:"totalSize"`
	Checksum    string      `json:"checksum"`
	OAuthEmail  string      `json:"oauthEmail,omitempty"`
	Scope       string      `json:"scope,omitempty"`
}

// FileEntry represents a single file within a context snapshot.
type FileEntry struct {
	RelPath  string `json:"relPath"`
	Size     int64  `json:"size"`
	Mode     uint32 `json:"mode"`
	Checksum string `json:"checksum"`
	Source   string `json:"source"` // "dotclaude", "claudejson", "claudemd", or "mcpjson"
}

var slugRe = regexp.MustCompile(`[^a-z0-9-]+`)

// Slugify normalizes a context name: lowercase, replace non-alphanumeric with hyphens.
func Slugify(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = slugRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// FileChecksum computes the SHA-256 checksum of a file.
func FileChecksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// isExtraFileSource returns true if the source tag represents an extra file
// (claude.json for user scope, CLAUDE.md/.mcp.json for project scope) rather
// than a file from the .claude/ directory.
func isExtraFileSource(source string) bool {
	return source == "claudejson" || source == "claudemd" || source == "mcpjson"
}

// ManifestChecksum computes a combined checksum from all file entries.
func ManifestChecksum(files []FileEntry) string {
	h := sha256.New()
	for _, f := range files {
		fmt.Fprintf(h, "%s:%s\n", f.RelPath, f.Checksum)
	}
	return hex.EncodeToString(h.Sum(nil))
}
