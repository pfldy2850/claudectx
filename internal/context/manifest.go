package context

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ReadManifest reads the manifest.json from a context directory.
func ReadManifest(contextDir string) (*Manifest, error) {
	path := filepath.Join(contextDir, "manifest.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	return &m, nil
}

// WriteManifest writes the manifest.json to a context directory atomically.
func WriteManifest(contextDir string, m *Manifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal manifest: %w", err)
	}
	data = append(data, '\n')

	path := filepath.Join(contextDir, "manifest.json")
	if err := os.MkdirAll(contextDir, 0755); err != nil {
		return fmt.Errorf("create context dir: %w", err)
	}

	tmpFile, err := os.CreateTemp(contextDir, ".manifest-*")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}
	tmpPath := tmpFile.Name()

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("write temp: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("close temp: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename: %w", err)
	}

	return nil
}

// ListContexts returns all context names in the storage directory.
func ListContexts(contextsDir string) ([]string, error) {
	entries, err := os.ReadDir(contextsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var names []string
	for _, e := range entries {
		if e.IsDir() {
			// Verify it has a manifest
			manifestPath := filepath.Join(contextsDir, e.Name(), "manifest.json")
			if _, err := os.Stat(manifestPath); err == nil {
				names = append(names, e.Name())
			}
		}
	}
	return names, nil
}

// ContextExists checks if a named context exists.
func ContextExists(contextsDir, name string) bool {
	manifestPath := filepath.Join(contextsDir, name, "manifest.json")
	_, err := os.Stat(manifestPath)
	return err == nil
}

// DeleteContext removes a saved context directory.
func DeleteContext(contextsDir, name string) error {
	dir := filepath.Join(contextsDir, name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("context %q not found", name)
	}
	return os.RemoveAll(dir)
}
