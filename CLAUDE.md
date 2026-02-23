# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Project Is

claudectx is a Go CLI tool that manages Claude Code configuration contexts — create, switch, and list snapshots of Claude config files. It supports two scopes:

- **User scope** (default outside git repos): `~/.claude/` + `~/.claude.json`, stored in `~/.claudectx/`
- **Project scope** (default inside git repos): `<git-root>/.claude/` + `<git-root>/CLAUDE.md` + `<git-root>/.mcp.json`, stored in `<git-root>/.claudectx/`

Auto-detects scope based on git repo presence. Override with `--scope user|project`.

## Claude Code File Paths Reference

Settings and memory files that this tool manages. See upstream docs for full details:
- Settings scopes: https://code.claude.com/docs/en/settings#available-scopes
- Memory types: https://code.claude.com/docs/en/memory#determine-memory-type

### Settings files by scope

| Scope | Settings | Subagents | MCP servers | Plugins |
|-------|----------|-----------|-------------|---------|
| User | `~/.claude/settings.json` | `~/.claude/agents/` | `~/.claude.json` | `~/.claude/settings.json` |
| Project | `.claude/settings.json` | `.claude/agents/` | `.mcp.json` | `.claude/settings.json` |
| Local | `.claude/settings.local.json` | — | — | `.claude/settings.local.json` |

### Memory files by scope

| Type | Location | Shared? |
|------|----------|---------|
| Project memory | `./CLAUDE.md` or `./.claude/CLAUDE.md` | Team (via git) |
| Project rules | `./.claude/rules/*.md` | Team (via git) |
| User memory | `~/.claude/CLAUDE.md` | Just you (all projects) |
| Local memory | `./CLAUDE.local.md` | Just you (current project) |
| Auto memory | `~/.claude/projects/<project>/memory/` | Just you (per project) |

## Build & Development Commands

```bash
make build            # Build binary to ./bin/claudectx (injects git version via ldflags)
make test             # Run all tests with verbose output
make test-cover       # Run tests with coverage report
go test ./internal/context/ -run TestAutoSaveOnSwitch -v   # Run a single test
make lint             # Lint with golangci-lint
make clean            # Remove bin/ and coverage artifacts
```

## Architecture

The CLI uses **Cobra** for command routing and **Bubbletea** for the interactive TUI picker.

### Key data flow

1. **Create** (`create`): Walks the scope's `.claude/` dir with include/exclude glob filters → copies matched files + the extra files (`claude.json` for user scope; `CLAUDE.md` + `.mcp.json` for project scope) into `<storage>/.claudectx/contexts/<name>/` → writes a `manifest.json` with checksums and metadata → sets current marker
2. **Switch** (`claudectx <name>`): Auto-saves current context's live files back to its snapshot → creates a timestamped backup → copies target context's snapshot to live paths → updates current marker
3. **Auto-save on switch**: Before restoring a different context, the current context's live state is automatically saved back to its snapshot, so edits are never lost

### Package responsibilities

- `internal/cli/` — One file per Cobra subcommand (`create.go`, `list.go`, `delete.go`, `show.go`, `current.go`, `version.go`). Global flags (`--verbose`, `--dry-run`, `--force`, `--config`, `--scope`) live in `root.go`. Shared helpers (`formatSize`, `ensureGitignore`) in `helpers.go`.
- `internal/context/` — Core domain logic. `snapshot.go` (save), `restore.go` (switch + auto-save), `manifest.go` (read/write/list manifests), `context.go` (shared types like Manifest, FileEntry, and helpers like Slugify, FileChecksum, isExtraFileSource).
- `internal/fileutil/` — File operations: `filter.go` (glob matching with `**` support), `walk.go` (filtered directory walking), `copy.go` (file copying with mkdir).
- `internal/config/` — Config loading with defaults. `scope.go` defines the `Scope` struct (`UserScope`, `ProjectScope`, `ResolveScope`). `defaults.go` defines include/exclude patterns for each scope. `config.go` provides `Load()` and `LoadWithScope()`.
- `internal/claude/` — Resolves `~/.claude/`, `~/.claude.json` paths and `ProjectRoot()` (git root detection).
- `internal/ui/` — Bubbletea-based interactive picker and formatted output helpers.

### Important conventions

- Context names are slugified (lowercase, alphanumeric + hyphens only) via `context.Slugify()`.
- File entries track their `Source` field as `"dotclaude"`, `"claudejson"`, `"claudemd"`, or `"mcpjson"` to know where to restore them. Use `isExtraFileSource()` to check for the extra file tags.
- Manifests include a `Scope` field (`"user"` or `"project"`). Restore validates scope match to prevent cross-scope accidents.
- Restore **clears managed files** before applying the snapshot, so stale files from the previous context don't linger. Unmanaged files in `.claude/` are untouched.
- All domain functions (`Save`, `Restore`) read paths from `cfg.Scope` — they never call `claude.DotClaudeDir()` or `claude.ClaudeJSONPath()` directly.
- Tests in `internal/context/` use `t.TempDir()` for isolation and construct `Config` with explicit `Scope` to avoid touching real paths.
