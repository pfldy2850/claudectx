package fileutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CopyFile copies a single file from src to dst atomically (write to temp + rename).
// It preserves file permissions.
func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer srcFile.Close()

	info, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("stat source: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("create dest dir: %w", err)
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(dst), ".claudectx-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	defer func() {
		if err != nil {
			os.Remove(tmpPath)
		}
	}()

	if _, err = io.Copy(tmpFile, srcFile); err != nil {
		tmpFile.Close()
		return fmt.Errorf("copy data: %w", err)
	}

	if err = tmpFile.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}

	if err = os.Chmod(tmpPath, info.Mode()); err != nil {
		return fmt.Errorf("chmod: %w", err)
	}

	if err = os.Rename(tmpPath, dst); err != nil {
		return fmt.Errorf("rename: %w", err)
	}

	return nil
}

// CopyDir recursively copies a directory from src to dst.
func CopyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return CopyFile(path, dstPath)
	})
}
