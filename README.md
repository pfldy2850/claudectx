# claudectx

A CLI tool for managing Claude Code configuration contexts. Save, switch, and diff snapshots of your Claude Code config files as switchable profiles.

Unlike tools that only manage OAuth/account profiles, **claudectx** manages your entire Claude Code configuration — settings, memory, plugin configs, and MCP server definitions — as switchable contexts.

Supports two scopes:

- **User scope** — `~/.claude/` + `~/.claude.json` (default outside git repos)
- **Project scope** — `<git-root>/.claude/` + `<git-root>/CLAUDE.md` (default inside git repos)

## Installation

### From Source

```bash
git clone https://github.com/pfldy2850/claudectx.git
cd claudectx
make build
# Binary is at ./bin/claudectx
```

### Go Install

```bash
go install github.com/pfldy2850/claudectx/cmd/claudectx@latest
```

## Usage

### Save Current State

```bash
claudectx save work
claudectx save personal --description "Personal account settings"
claudectx save full-backup --include-all  # Include all files (even debug, cache, etc.)
```

### Switch Context

```bash
claudectx work       # Switch to 'work' context
claudectx personal   # Switch to 'personal' context
```

### Interactive Selection

```bash
claudectx            # Opens TUI picker
```

### List Contexts

```bash
claudectx list
# Scope: project (/path/to/project/.claudectx)
#
# * work (5 files, 2.3 KB) - Work account
#   personal (3 files, 1.1 KB) - Personal account

claudectx list --json   # JSON output for scripting
```

### Show Context Details

```bash
claudectx show work
# Context: work (active)
# Description: Work account
# Scope: user
# OAuth Email: user@company.com
# Created: 2025-01-15 10:30:00
# Files: 5
# Total Size: 2.3 KB
#
# Files:
#   claude.json (1.2 KB) [claudejson]
#   dotclaude/settings.json (256 B) [dotclaude]
#   ...
```

### Diff Against Saved Context

```bash
claudectx diff work
#   M settings.json (256 B → 312 B)
#   + plugins/blocklist.json (48 B)
#   - projects/old/memory/MEMORY.md (128 B)
```

### Show Active Context

```bash
claudectx current
# work
```

### Delete Context

```bash
claudectx delete old-context
claudectx delete old-context --force   # Skip confirmation
```

### Scope Override

By default, claudectx auto-detects scope based on whether you're inside a git repo. Use `--scope` to override:

```bash
claudectx save my-settings --scope user      # Force user scope
claudectx list --scope project               # Force project scope
```

## What Gets Saved

### User Scope

Saves configuration and memory files while excluding large temporary data:

**Included:**
- `~/.claude.json` — Core settings (OAuth, MCP servers, feature flags)
- `settings.json`, `settings.local.json`, `remote-settings.json`
- `plugins/blocklist.json`
- `projects/*/memory/**` — Project memory files

**Excluded:**
- `debug/`, `cache/`, `plugins/cache/` — Temporary/cached data
- `projects/*/*.jsonl` — Session logs
- `file-history/`, `todos/`, `tasks/`, `shell-snapshots/` — Ephemeral state
- `statsig/`, `telemetry/`, `usage-data/` — Analytics

### Project Scope

Saves all project-level Claude config:

**Included:**
- `<git-root>/CLAUDE.md` — Project instructions
- `<git-root>/.claude/**` — All files in the project's `.claude/` directory

**Excluded:**
- `.DS_Store` files

Use `--include-all` to override excludes and snapshot everything.

## Global Flags

| Flag | Description |
|------|-------------|
| `--scope <user\|project>` | Override auto-detected scope |
| `--verbose`, `-v` | Verbose output |
| `--dry-run` | Show what would happen without making changes |
| `--force`, `-f` | Skip confirmations |
| `--config <path>` | Custom config file path |

## Storage

### User Scope

Stored in `~/.claudectx/`:

```
~/.claudectx/
├── config.json        # User configuration
├── current            # Active context name
├── contexts/          # Saved context snapshots
│   ├── work/
│   │   ├── manifest.json
│   │   ├── claude.json
│   │   └── dotclaude/
│   └── personal/
└── backups/           # Pre-switch backups
```

### Project Scope

Stored in `<git-root>/.claudectx/`:

```
<git-root>/.claudectx/
├── config.json        # Project configuration
├── current            # Active context name
├── contexts/          # Saved context snapshots
│   ├── experiment-a/
│   │   ├── manifest.json
│   │   ├── CLAUDE.md
│   │   └── dotclaude/
│   └── experiment-b/
└── backups/           # Pre-switch backups
```

> **Note:** Add `.claudectx/` to your `.gitignore` when using project scope.

## Architecture

```
cmd/claudectx/         Entry point
internal/
├── cli/               Cobra CLI commands (--scope flag, gitignore warning)
├── context/           Core operations (save, restore, diff, manifest)
├── fileutil/          File copy, glob filtering, directory walking
├── config/            Configuration, scope resolution, and defaults
├── claude/            Claude Code path resolution and git root detection
└── ui/                Interactive TUI (bubbletea) and formatted output
```

## Development

```bash
make build       # Build binary
make test        # Run tests
make test-cover  # Run tests with coverage
make clean       # Clean build artifacts
```

## License

MIT
