<div align="center">

# Broominal

**Safe, transparent, undoable Windows cleanup from the terminal.**

[![Go](https://img.shields.io/badge/Go-1.26.3-00ADD8?logo=go)](https://go.dev)
[![CI](https://github.com/elev1e1nSure/broominal/actions/workflows/ci.yml/badge.svg)](https://github.com/elev1e1nSure/broominal/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/elev1e1nSure/broominal?label=release)](https://github.com/elev1e1nSure/broominal/releases)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows-blue?logo=windows)](https://github.com/elev1e1nSure/broominal)

[Russian](README.ru.md)

</div>

---

<table>
<tr>
<td width="6" bgcolor="#00ADD8"></td>
<td>

**Broominal is a Windows cleanup CLI/TUI tool built around one rule: cleanup must be reversible.**

It moves selected files into a local quarantine with JSON manifests and reports, so every cleanup can be inspected and restored.

</td>
</tr>
</table>

It is designed for boring, predictable cleanup — not fake “PC boost” magic.

---

## Why Broominal

<table>
<tr>
<td width="6" bgcolor="#2EA043"></td>
<td><strong>Safe by default</strong><br>Files are moved to quarantine instead of being permanently deleted.</td>
</tr>
<tr>
<td width="6" bgcolor="#00ADD8"></td>
<td><strong>Transparent</strong><br>Scan results, reports, and manifests are plain JSON.</td>
</tr>
<tr>
<td width="6" bgcolor="#8957E5"></td>
<td><strong>Undoable</strong><br>Restore a cleanup by ID or restore the latest cleanup.</td>
</tr>
<tr>
<td width="6" bgcolor="#F0883E"></td>
<td><strong>Interactive</strong><br>Bubbletea TUI for category selection, previews, dry-run, and restore flow.</td>
</tr>
<tr>
<td width="6" bgcolor="#8B949E"></td>
<td><strong>Multilingual</strong><br>English and Russian with first-run auto-detection.</td>
</tr>
</table>

---

## Features

| Feature | Description |
|--------|-------------|
| **Smart Scan** | Temp, Downloads, Browser Cache, Recycle Bin, Logs, Old Installers, Large Old Files, Thumbnails, DirectX Shader Cache, Delivery Optimization, Windows Error Reports, Discord Cache, Steam Cache, VSCode Cache, Edge Code Cache, Chrome Code Cache, Firefox Cache2, Old Temp Files, Old .tmp/.log/.bak, Empty Folders, npm Cache, pip Cache, Windows Update Cache, Crash & Memory Dumps, Nvidia Installer Leftovers, Telegram Desktop Cache |
| **Risk Levels** | `safe` / `review` / `danger` — system paths and protected extensions are never cleaned automatically |
| **Undoable Cleanup** | Every cleanup gets a restore ID; `restore <id>` moves files back |
| **Dry-Run** | `--dry-run` in CLI and `T` in TUI show what would happen without moving files |
| **Config-Driven** | JSON config for thresholds, exclusions, category toggles, risk overrides, and language |
| **Doctor** | Health checks for permissions, directories, manifests, and quarantine stats |
| **Quarantine Cleanup** | Remove old quarantines with dry-run preview and explicit confirmation |
| **TUI** | Interactive Bubbletea interface with Main Menu, scan flow, restore picker, doctor, config viewer, cleanup, and language selector |
| **i18n** | English / Russian. Auto-detects language by IP on first launch |

---

## Safety model

<table>
<tr>
<td width="6" bgcolor="#2EA043"></td>
<td>

**Safe cleanup is selected by default. Review cleanup requires manual choice. Danger items are not cleaned automatically.**

</td>
</tr>
</table>

| Risk | Default behavior | Examples |
|------|------------------|----------|
| `safe` | selected by default | Temp files, thumbnails, shader cache, common app caches |
| `review` | user must select manually | Downloads, dumps, Windows Update cache, Telegram cache |
| `danger` | never cleaned automatically | system paths, protected extensions, unknown risky locations |

Files are moved to:

```text
%LOCALAPPDATA%\broominal\quarantine\<restore-id>
```

Each cleanup stores a `manifest.json`, which maps original paths to quarantined paths for restore.

---

## Quick Start

### Install

```powershell
# From source (requires Go 1.26.3+)
go install github.com/elev1e1nSure/broominal/cmd/broominal@latest
```

Or download the latest `.exe` from [Releases](../../releases).

### Usage

```powershell
# Scan safe zones
broominal scan

# Launch interactive TUI
broominal ui

# Clean safe items only
broominal clean --safe

# Simulate cleanup without moving files
broominal clean --dry-run

# Restore a specific cleanup
broominal restore <id>

# Restore with overwrite
broominal restore <id> --force-overwrite

# Run health checks
broominal doctor

# Show config
broominal config

# Remove quarantines older than 30 days
broominal quarantine-cleanup --dry-run
broominal quarantine-cleanup --force
broominal quarantine-cleanup --force --max-age-days 7
```

---

## Build from source

Requires **Go 1.26.3+**.

```powershell
# Clone and build
go build -o broominal.exe ./cmd/broominal

# Run the TUI
.\broominal.exe ui
```

---

## Architecture

<table>
<tr>
<td width="6" bgcolor="#8957E5"></td>
<td>

**Small packages, explicit responsibilities, and no hidden cleanup magic.**

</td>
</tr>
</table>

```text
cmd/broominal/      CLI entrypoint (Cobra)
pkg/
  scanner/          File discovery by cleanup category
  cleaner/          Quarantine move + report save pipeline
  quarantine/       Move / Restore / Cleanup with JSON manifests
  report/           JSON report generation
  risk/             Risk classification from paths, extensions, and config
  config/           JSON configuration and defaults
  doctor/           Health checks for runtime state
  i18n/             EN/RU localization and language detection
  style/            ANSI color helpers for CLI output
  util/             Size formatting and shared helpers
  types/            Shared domain types
internal/
  tui/              Bubbletea interactive interface
```

---

## Development

### Githooks

Enable shared hooks to enforce code style and commit conventions:

```powershell
git config core.hooksPath githooks
```

Hooks included:

- `pre-commit` — warns when code changes may need doc updates
- `commit-msg` — enforces [Conventional Commits](https://www.conventionalcommits.org/) (`feat|fix|chore|refactor|docs|test|build|ci|perf|style|revert`)

### CI / CD

All pushes and pull requests to `main` trigger:

- `gofmt` check
- `go vet`
- `golangci-lint`
- `go test ./...`
- Windows build artifact upload

### Releasing

Run the **Release** workflow from GitHub Actions. It will:

1. Generate release notes from Conventional Commits via `git-cliff`
2. Build `broominal.exe`
3. Create a signed tag and GitHub Release with `checksums.txt`

---

## Contributing

Bug reports, cleanup-category ideas, safety improvements, and Windows edge cases are welcome. See [CONTRIBUTING.md](CONTRIBUTING.md).

---

## License

[MIT](LICENSE) © elev1e1nSure
