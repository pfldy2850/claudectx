package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaultConfig(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.IncludePatterns) == 0 {
		t.Error("expected default include patterns")
	}
	if len(cfg.ExcludePatterns) == 0 {
		t.Error("expected default exclude patterns")
	}
}

func TestLoadCustomConfig(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	err := os.WriteFile(cfgPath, []byte(`{"storageDir":"/tmp/test-claudectx"}`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.StorageDir != "/tmp/test-claudectx" {
		t.Errorf("got storageDir=%s, want /tmp/test-claudectx", cfg.StorageDir)
	}
	if len(cfg.IncludePatterns) == 0 {
		t.Error("expected default include patterns to be set")
	}
}

func TestConfigPaths(t *testing.T) {
	cfg := &Config{StorageDir: "/tmp/claudectx"}
	if cfg.ContextsDir() != "/tmp/claudectx/contexts" {
		t.Errorf("unexpected contexts dir: %s", cfg.ContextsDir())
	}
	if cfg.BackupsDir() != "/tmp/claudectx/backups" {
		t.Errorf("unexpected backups dir: %s", cfg.BackupsDir())
	}
	if cfg.CurrentFile() != "/tmp/claudectx/current" {
		t.Errorf("unexpected current file: %s", cfg.CurrentFile())
	}
}
