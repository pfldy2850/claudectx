package context

import (
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"My Work", "my-work"},
		{"personal", "personal"},
		{"Test Context 123", "test-context-123"},
		{"  Spaces  ", "spaces"},
		{"UPPER_CASE", "upper-case"},
		{"special!@#chars", "special-chars"},
		{"multiple---dashes", "multiple---dashes"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := Slugify(tt.input)
			if got != tt.want {
				t.Errorf("Slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestManifestChecksum(t *testing.T) {
	files := []FileEntry{
		{RelPath: "a.json", Checksum: "abc123"},
		{RelPath: "b.json", Checksum: "def456"},
	}
	sum1 := ManifestChecksum(files)
	if sum1 == "" {
		t.Error("expected non-empty checksum")
	}

	// Same files, same checksum
	sum2 := ManifestChecksum(files)
	if sum1 != sum2 {
		t.Error("expected deterministic checksum")
	}

	// Different files, different checksum
	files2 := []FileEntry{
		{RelPath: "a.json", Checksum: "abc123"},
		{RelPath: "c.json", Checksum: "ghi789"},
	}
	sum3 := ManifestChecksum(files2)
	if sum1 == sum3 {
		t.Error("expected different checksum for different files")
	}
}
