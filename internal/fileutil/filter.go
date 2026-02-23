package fileutil

import (
	"path/filepath"
	"strings"
)

// MatchesAny checks if the given relative path matches any of the glob patterns.
func MatchesAny(relPath string, patterns []string) bool {
	relPath = filepath.ToSlash(relPath)
	for _, pattern := range patterns {
		pattern = filepath.ToSlash(pattern)
		if matchGlob(relPath, pattern) {
			return true
		}
	}
	return false
}

// matchGlob performs glob matching supporting ** for recursive directory matching.
func matchGlob(path, pattern string) bool {
	if strings.Contains(pattern, "**") {
		return matchDoubleGlob(path, pattern)
	}
	matched, _ := filepath.Match(pattern, path)
	return matched
}

// matchDoubleGlob handles patterns with ** (match any number of directories).
func matchDoubleGlob(path, pattern string) bool {
	parts := strings.SplitN(pattern, "**", 2)
	if len(parts) != 2 {
		matched, _ := filepath.Match(pattern, path)
		return matched
	}

	prefix := strings.TrimSuffix(parts[0], "/")
	suffix := strings.TrimPrefix(parts[1], "/")

	pathParts := strings.Split(path, "/")

	// Match prefix segments (may contain globs like *)
	if prefix == "" {
		// Pattern like "**" or "**/suffix"
		if suffix == "" {
			return true
		}
		// Try matching suffix against every sub-path
		for i := range pathParts {
			subPath := strings.Join(pathParts[i:], "/")
			if matched, _ := filepath.Match(suffix, subPath); matched {
				return true
			}
		}
		return false
	}

	// prefix is non-empty — match prefix segments against path segments
	prefixParts := strings.Split(prefix, "/")

	if len(prefixParts) > len(pathParts) {
		return false
	}

	// Check if prefix segments match
	if !matchSegments(pathParts[:len(prefixParts)], prefixParts) {
		return false
	}

	// Remaining path after prefix
	remaining := pathParts[len(prefixParts):]

	if suffix == "" {
		// Pattern like "prefix/**" — match anything under prefix
		return len(remaining) > 0 || len(pathParts) == len(prefixParts)
	}

	// Pattern like "prefix/**/suffix" — match suffix against any tail of remaining
	for i := range remaining {
		subPath := strings.Join(remaining[i:], "/")
		if matched, _ := filepath.Match(suffix, subPath); matched {
			return true
		}
	}
	return false
}

// matchSegments checks if path segments match pattern segments (supporting single * glob).
func matchSegments(pathSegs, patternSegs []string) bool {
	if len(pathSegs) != len(patternSegs) {
		return false
	}
	for i := range pathSegs {
		matched, _ := filepath.Match(patternSegs[i], pathSegs[i])
		if !matched {
			return false
		}
	}
	return true
}
