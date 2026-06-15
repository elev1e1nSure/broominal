# Developing

## Prerequisites

- Go 1.26+
- [just](https://github.com/casey/just) (command runner)

## Quick Start

```powershell
git clone https://github.com/elev1e1nSure/broominal.git
cd broominal
just build
just run ui
```

## Just Commands

| Command | Description |
|---------|-------------|
| `just build` | Build `broominal.exe` |
| `just build-release v1.0.0` | Build with embedded version for release |
| `just run` | Build and run (pass args: `just run scan --safe`) |
| `just test` | Run all tests |
| `just test-pkg` | Run package tests only (no TUI) |
| `just lint` | Run golangci-lint |
| `just vet` | Run `go vet` |
| `just fmt` | Format code with `gofmt` |
| `just check` | Full check: fmt → vet → lint → test-pkg |
| `just clean` | Remove build artifacts |
| `just changelog-raw` | Preview raw commits since last tag (no AI) |
| `just changelog` | Generate beautified changelog via OpenRouter |
| `just changelog-save FILE` | Save AI-polished changelog to file |

### AI Changelog

Release notes are auto-generated from conventional commits via [git-cliff](../cliff.toml). The `scripts/ai-changelog.ps1` script optionally pipes the output through an LLM (OpenRouter) to produce polished, human-readable release notes.

```powershell
$env:OPENROUTER_API_KEY = "sk-or-..."
just changelog
```

Set `$env:OPENROUTER_MODEL` to override the model (default: `deepseek/deepseek-v4-flash`).

## Git Hooks

```powershell
git config core.hooksPath githooks
```
- `pre-commit` — warns on undocumented category or i18n changes
- `commit-msg` — enforces conventional commits (types: `feat`, `fix`, `chore`, `refactor`, `docs`, `test`, `build`, `ci`, `perf`, `style`, `revert`)

## Architecture

```
cmd/broominal/     CLI entrypoint (Cobra) — all commands
pkg/
  categories/      Single registry of all scanner categories
  scanner/         File discovery by cleanup category
  cleaner/         Quarantine move + report save pipeline
  quarantine/      Move, restore, cleanup, JSON manifests
  config/          JSON configuration, defaults, presets
  doctor/          Health checks for runtime environment
  report/          JSON report generation
  i18n/            EN/RU translations, auto-detect by IP
  style/           ANSI color helpers for CLI
  util/            Size formatting, error helpers
  types/           Shared structs: Item, ScanResult, Manifest, etc.
  update/          Check for and install GitHub releases
  pathman/         PATH manipulation
  taskscheduler/   Windows Task Scheduler integration
internal/
  tui/             Bubbletea interactive interface
```

## Adding a Scanner Category

4 files to edit:

1. **`pkg/categories/categories.go`** — add a `Def{Name, InternalKey, Risk, MinPreset}`  
2. **`pkg/scanner/scanner.go`** — implement `scanXxx(ctx, cfg) ([]Item, error)`  
3. **`pkg/scanner/scanner_registry.go`** — wire `"Category Name": scanXxx` in `scanFuncs`  
4. **`pkg/i18n/i18n.go`** — add `cat_*` and `cat_desc_*` for EN and RU

A missing registration causes a startup panic — safe by design.

## Config

Default config is auto-created on first run at `%LOCALAPPDATA%\broominal\config.json`. Missing categories in existing configs are merged from defaults.

## Presets

`pkg/categories/categories.go` defines `MinPreset` for each category. `pkg/config/config.go` maps presets to enabled categories when `ActivePreset` is not `custom`.

## Testing

Tests create temp directories — no real files are touched. Scanner tests use `os.TempDir()` with known test fixtures. Quarantine tests create and clean up temp quarantine directories.

## CI

```
gofmt → go vet → golangci-lint → go test ./... → Windows build
```

Release workflow (on `v*` tag push):
```
git-cliff → [AI beautifier] → build → publish GitHub Release
```
