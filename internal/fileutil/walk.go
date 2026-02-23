package fileutil

import (
	"os"
	"path/filepath"
)

// WalkResult holds information about a walked file.
type WalkResult struct {
	RelPath string
	AbsPath string
	Info    os.FileInfo
}

// WalkFiltered walks a directory and returns files that match include patterns
// but not exclude patterns.
func WalkFiltered(root string, includes, excludes []string) ([]WalkResult, error) {
	var results []WalkResult

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}

		// Check include patterns
		if !MatchesAny(relPath, includes) {
			return nil
		}

		// Check exclude patterns
		if MatchesAny(relPath, excludes) {
			return nil
		}

		results = append(results, WalkResult{
			RelPath: relPath,
			AbsPath: path,
			Info:    info,
		})
		return nil
	})

	return results, err
}
