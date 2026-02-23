package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds user configuration for claudectx.
type Config struct {
	StorageDir      string   `json:"storageDir,omitempty"`
	IncludePatterns []string `json:"includePatterns,omitempty"`
	ExcludePatterns []string `json:"excludePatterns,omitempty"`
	Scope           *Scope   `json:"-"` // runtime only, set by LoadWithScope
}

// DefaultStorageDir returns the default ~/.claudectx/ path.
func DefaultStorageDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claudectx"), nil
}

// Load reads config from the given path, falling back to defaults.
// Uses user scope by default for backward compatibility.
func Load(path string) (*Config, error) {
	scope, err := UserScope()
	if err != nil {
		return nil, err
	}
	return LoadWithScope(path, scope)
}

// LoadWithScope reads config from the given path and applies the provided scope.
func LoadWithScope(path string, scope *Scope) (*Config, error) {
	cfg := &Config{
		IncludePatterns: scope.IncludePatterns,
		ExcludePatterns: scope.ExcludePatterns,
		StorageDir:      scope.StorageDir,
		Scope:           scope,
	}

	if path == "" {
		path = filepath.Join(scope.StorageDir, "config.json")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Re-apply defaults for any empty fields
	if len(cfg.IncludePatterns) == 0 {
		cfg.IncludePatterns = scope.IncludePatterns
	}
	if len(cfg.ExcludePatterns) == 0 {
		cfg.ExcludePatterns = scope.ExcludePatterns
	}
	if cfg.StorageDir == "" {
		cfg.StorageDir = scope.StorageDir
	}

	// Always keep scope reference
	cfg.Scope = scope

	return cfg, nil
}

// ContextsDir returns the path to the contexts directory.
func (c *Config) ContextsDir() string {
	return filepath.Join(c.StorageDir, "contexts")
}

// BackupsDir returns the path to the backups directory.
func (c *Config) BackupsDir() string {
	return filepath.Join(c.StorageDir, "backups")
}

// CurrentFile returns the path to the 'current' marker file.
func (c *Config) CurrentFile() string {
	return filepath.Join(c.StorageDir, "current")
}
