package fileutil

import "testing"

func TestMatchesAny(t *testing.T) {
	tests := []struct {
		relPath  string
		patterns []string
		want     bool
	}{
		{"settings.json", []string{"settings.json"}, true},
		{"settings.local.json", []string{"settings.local.json"}, true},
		{"other.json", []string{"settings.json"}, false},
		{"debug/foo.log", []string{"debug/**"}, true},
		{"debug/sub/bar.log", []string{"debug/**"}, true},
		{"plugins/cache/data.json", []string{"plugins/cache/**"}, true},
		{"plugins/blocklist.json", []string{"plugins/blocklist.json"}, true},
		{"projects/myproj/memory/MEMORY.md", []string{"projects/*/memory/**"}, true},
		{"projects/myproj/session.jsonl", []string{"projects/*/*.jsonl"}, true},
		{"projects/myproj/sub/deep.jsonl", []string{"projects/*/*.jsonl"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.relPath, func(t *testing.T) {
			got := MatchesAny(tt.relPath, tt.patterns)
			if got != tt.want {
				t.Errorf("MatchesAny(%q, %v) = %v, want %v", tt.relPath, tt.patterns, got, tt.want)
			}
		})
	}
}
