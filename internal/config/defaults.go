package config

// DefaultIncludePatterns are file patterns included by default in snapshots.
var DefaultIncludePatterns = []string{
	"settings.json",
	"settings.local.json",
	"remote-settings.json",
	"plugins/blocklist.json",
	"projects/*/memory/**",
}

// DefaultExcludePatterns are file patterns excluded by default from snapshots.
var DefaultExcludePatterns = []string{
	"debug/**",
	"projects/*/*.jsonl",
	"plugins/cache/**",
	"file-history/**",
	"todos/**",
	"tasks/**",
	"plans/**",
	"shell-snapshots/**",
	"cache/**",
	"session-env/**",
	"ide/**",
	"usage-data/**",
	"paste-cache/**",
	"statsig/**",
	"telemetry/**",
	"history.jsonl",
	"backups/**",
}

// DefaultProjectIncludePatterns include all files for project-scope snapshots.
var DefaultProjectIncludePatterns = []string{"**"}

// DefaultProjectExcludePatterns exclude OS junk from project-scope snapshots.
var DefaultProjectExcludePatterns = []string{"**/.DS_Store"}
