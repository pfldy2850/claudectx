# claudectx

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/pfldy2850/claudectx)](https://github.com/pfldy2850/claudectx/releases/latest)

A CLI tool for managing [Claude Code](https://code.claude.com/) configuration contexts. Create, switch, and list snapshots of your Claude Code config files as switchable profiles.

Unlike tools that only manage OAuth/account profiles, **claudectx** manages your entire Claude Code configuration — settings, memory, plugin configs, and MCP server definitions — as switchable contexts.

## Scopes

claudectx supports two scopes and auto-detects which to use:

| Scope | Source files | Storage | Default when |
|-------|-------------|---------|-------------|
| **User** | `~/.claude/` + `~/.claude.json` | `~/.claudectx/` | Outside a project |
| **Project** | `<root>/.claude/` + `CLAUDE.md` + `.mcp.json` | `<root>/.claudectx/` | Inside a project |

> **Windows:** `~` refers to `%USERPROFILE%` (typically `C:\Users\<username>`). All paths work the same way.

Project root is detected by (highest priority first):
1. `--root` flag (explicit path)
2. Claude marker files (`.claude/`, `CLAUDE.md`, `.claudectx/`)
3. Git repository root (`.git/`)
4. Current directory fallback (with `--scope project`)

## Installation

### Homebrew (macOS / Linux)

```bash
brew install pfldy2850/tap/claudectx
```

### Scoop (Windows)

```powershell
scoop bucket add pfldy2850 https://github.com/pfldy2850/scoop-bucket
scoop install claudectx
```

### Download Binary

Pre-built binaries for Linux, macOS, and Windows are available on the [Releases](https://github.com/pfldy2850/claudectx/releases) page.

```bash
# macOS (Apple Silicon)
curl -sL https://github.com/pfldy2850/claudectx/releases/latest/download/claudectx_$(curl -s https://api.github.com/repos/pfldy2850/claudectx/releases/latest | grep tag_name | cut -d'"' -f4 | sed 's/^v//')_darwin_arm64.tar.gz | tar xz
sudo mv claudectx /usr/local/bin/
```

```powershell
# Windows (PowerShell)
$version = (Invoke-RestMethod https://api.github.com/repos/pfldy2850/claudectx/releases/latest).tag_name -replace '^v',''
Invoke-WebRequest -Uri "https://github.com/pfldy2850/claudectx/releases/latest/download/claudectx_${version}_windows_amd64.zip" -OutFile claudectx.zip
Expand-Archive claudectx.zip -DestinationPath .
Move-Item claudectx.exe "$env:LOCALAPPDATA\Microsoft\WindowsApps\"
Remove-Item claudectx.zip
```

### Go Install

```bash
go install github.com/pfldy2850/claudectx/cmd/claudectx@latest
```

### From Source

```bash
git clone https://github.com/pfldy2850/claudectx.git
cd claudectx
make build       # macOS / Linux
# or
go build -o claudectx.exe ./cmd/claudectx   # Windows
```

## Usage

### Create a Context

Snapshot the current live files as a new context:

```bash
claudectx create work
claudectx create personal --description "Personal account settings"
```

Create an empty context (clean slate):

```bash
claudectx create blank --from-scratch
```

Copy from an existing context:

```bash
claudectx create work-v2 --copy-from work
```

### Switch Context

```bash
claudectx work       # Switch to 'work' context
claudectx personal   # Switch to 'personal' context
```

Edits to the current context are auto-saved before switching, so changes are never lost.

### Interactive Selection

```bash
claudectx            # Opens TUI picker when no arguments given
```

### List Contexts

```bash
claudectx list
# Scope: project (/path/to/project/.claudectx)
#
# * work (5 files, 2.3 KB) - Work account
#   personal (3 files, 1.1 KB) - Personal account

claudectx ls --json   # JSON output for scripting
```

When inside a project without `--scope`, both project and user contexts are shown.

### Show Context Details

```bash
claudectx show work
# Context: work (active)
# Description: Work account
# Scope: user
# OAuth Email: user@company.com
# Created: 2025-01-15 10:30:00
# Updated: 2025-01-20 14:22:00
# Files: 5
# Total Size: 2.3 KB
# Checksum: a1b2c3d4e5f6...
#
# Files:
#   claude.json (1.2 KB) [claudejson]
#   dotclaude/settings.json (256 B) [dotclaude]
#   ...
```

### Show Active Context

```bash
claudectx current
# work
```

### Delete Context

```bash
claudectx delete old-context
claudectx rm old-context --force   # Skip confirmation
```

The active context cannot be deleted — switch to another context first.

### Scope Override

```bash
claudectx create my-settings --scope user      # Force user scope
claudectx list --scope project                  # Force project scope
claudectx work --root /path/to/project          # Explicit project root
```

## What Gets Saved

### User Scope (`~/.claude/`)

Snapshots configuration and memory files while excluding large ephemeral data:

**Included:**
- `~/.claude.json` — Core settings (OAuth, MCP servers, feature flags)
- `settings.json`, `settings.local.json`, `remote-settings.json`
- `plugins/blocklist.json`
- `projects/*/memory/**` — Project memory files

**Excluded:**
- `debug/**`, `cache/**`, `plugins/cache/**` — Temporary/cached data
- `projects/*/*.jsonl` — Session logs
- `file-history/**`, `todos/**`, `tasks/**`, `plans/**` — Ephemeral state
- `shell-snapshots/**`, `session-env/**`, `ide/**` — Runtime data
- `statsig/**`, `telemetry/**`, `usage-data/**` — Analytics
- `paste-cache/**`, `history.jsonl`, `backups/**`

### Project Scope (`<root>/.claude/`)

Snapshots all project-level Claude config:

**Included:**
- `<root>/CLAUDE.md` — Project instructions
- `<root>/.mcp.json` — MCP server configuration
- `<root>/.claude/**` — All files in the project's `.claude/` directory

**Excluded:**
- `.DS_Store` files

## Global Flags

| Flag | Description |
|------|-------------|
| `--scope <user\|project>` | Override auto-detected scope |
| `--root <path>` | Explicit project root directory (implies project scope) |
| `--verbose`, `-v` | Verbose output |
| `--dry-run` | Show what would happen without making changes |
| `--force`, `-f` | Skip confirmations |
| `--config <path>` | Custom config file path |

## Storage Layout

```
<storage-dir>/
├── config.json          # Configuration
├── current              # Active context name
├── contexts/            # Saved context snapshots
│   ├── work/
│   │   ├── manifest.json
│   │   ├── claude.json        # (user scope) or CLAUDE.md (project scope)
│   │   └── dotclaude/         # .claude/ directory snapshot
│   │       ├── settings.json
│   │       └── ...
│   └── personal/
└── backups/             # Pre-switch backups
```

> **Note:** `.claudectx/` is automatically added to `.gitignore` when using project scope.

## Architecture

```
cmd/claudectx/         Entry point
internal/
├── cli/               Cobra commands & global flags
├── context/           Core operations (save, restore, manifest)
├── fileutil/          File copy, glob filtering, directory walking
├── config/            Configuration, scope resolution, defaults
├── claude/            Claude Code path resolution, project root detection
└── ui/                Interactive TUI (Bubbletea) and formatted output
```

## Development

```bash
make build       # Build binary (injects git version via ldflags)
make test        # Run all tests
make test-cover  # Run tests with coverage report
make lint        # Lint with golangci-lint
make clean       # Clean build artifacts
```

## License

[MIT](LICENSE)
