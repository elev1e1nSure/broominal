<div align="center">

# 🧹 broominal

**safe, transparent, undoable windows cleanup from the terminal**

[![go](https://img.shields.io/badge/go-1.26.3-00ADD8?logo=go\&logoColor=white)](https://go.dev)
[![ci](https://github.com/elev1e1nSure/broominal/actions/workflows/ci.yml/badge.svg)](https://github.com/elev1e1nSure/broominal/actions/workflows/ci.yml)
[![release](https://img.shields.io/github/v/release/elev1e1nSure/broominal?label=release)](https://github.com/elev1e1nSure/broominal/releases)
[![license](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![platform](https://img.shields.io/badge/platform-Windows-0078D4?logo=windows\&logoColor=white)](https://github.com/elev1e1nSure/broominal)

[english](README.md) · [русский](README.ru.md)

</div>

---

## What is it

**broominal** — a Windows cleanup CLI/TUI built around one rule:

> cleanup must be **reversible**

Instead of permanently deleting files, broominal moves selected items into a local **quarantine**, stores JSON manifests, and makes every cleanup inspectable and restorable.

No fake boost magic. No hidden system tweaking. No "trust me bro" cleanup.

<img alt="Main menu" src="screenshots/screenshot_main.png" width="600" />

<img alt="Scan results" src="screenshots/screenshot_scan.png" width="600" />

---

## Installation

```powershell
go install github.com/elev1e1nSure/broominal/cmd/broominal@latest
```

Or grab the latest `.exe` from [releases][releases].

[releases]: https://github.com/elev1e1nSure/broominal/releases

<details>
<summary>Build from source</summary>

```powershell
git clone https://github.com/elev1e1nSure/broominal.git
cd broominal

go build -o broominal.exe ./cmd/broominal
.\broominal.exe ui
```

</details>

---

## Quick Start

Typical workflow: scan → preview → clean → restore if needed.

```powershell
# find cleanup candidates
broominal scan

# preview what would be cleaned (dry-run)
broominal clean --dry-run

# clean safe items only
broominal clean --safe

# something went wrong? restore the last cleanup
broominal restore last
```

For an interactive experience, run `broominal ui`.

---

## Safety Model

> **safe** cleanup is selected by default. **review** requires manual choice. **danger** items are never cleaned automatically.

```
┌─────────────────────────────────────────────────────────────┐
│  safe    ▸ selected by default                              │
│           temp files, thumbnails, shader cache, app caches  │
├─────────────────────────────────────────────────────────────┤
│  review  ▸ user must select manually                        │
│           downloads, dumps, windows update cache, telegram  │
├─────────────────────────────────────────────────────────────┤
│  danger  ▸ never cleaned automatically                      │
│           system paths, protected extensions                │
└─────────────────────────────────────────────────────────────┘
```

Files are moved to `%LOCALAPPDATA%\broominal\quarantine\<restore-id>` with a `manifest.json` mapping original paths to quarantined paths.

---

## Highlights

- **safe by default** — files are quarantined, not deleted
- **transparent** — scan results, reports, and manifests are plain JSON
- **undoable** — restore any cleanup by ID or restore the latest one
- **predictable** — explicit categories, risk levels, and exclusions
- **interactive** — Bubbletea TUI for scan, preview, and restore
- **multilingual** — English and Russian with first-run auto-detection
- **30+ categories** — temp, caches, logs, browser data, dev tools, and more
- **doctor** — lightweight health checks for permissions, manifests, and state

---

## Commands

### scan

Scan your system for cleanup candidates across 30+ categories.

```powershell
broominal scan
```

Scan results are saved as JSON for transparency and can be reviewed before any cleanup.

### clean

Clean selected items. By default, safe and review items are cleaned. Danger items require explicit confirmation.

```powershell
# clean safe and review items (default)
broominal clean

# allow cleaning danger items
broominal clean --danger

# clean only safe items
broominal clean --safe

# preview what would be cleaned without actually cleaning
broominal clean --dry-run
```

### restore

Restore a previous cleanup. Every cleanup gets a unique ID that can be used to restore files.

```powershell
# restore a specific cleanup by ID
broominal restore <id>

# restore the latest cleanup
broominal restore last

# restore with overwrite if file already exists
broominal restore <id> --force-overwrite
```

### ui

Launch the interactive TUI for a guided cleanup experience.

```powershell
broominal ui
```

The TUI lets you:
- Browse scan results by category
- Toggle items for cleanup
- Preview total size before cleaning
- Handle restore conflicts interactively

### doctor

Run health checks to verify broominal is working correctly.

```powershell
broominal doctor
```

Checks:
- Admin rights
- Directory write access
- Manifest integrity
- Quarantine statistics

### config

View and edit configuration.

```powershell
# show current config
broominal config
```

Config options include:
- Enabled categories
- Age/size thresholds
- Exclusions
- Risk overrides
- Language preference
- Quarantine max age

### quarantine-cleanup

Clean up old quarantines to free up space.

```powershell
# preview old quarantine cleanup (shows what will be removed)
broominal quarantine-cleanup

# remove quarantines older than 30 days
broominal quarantine-cleanup --force

# remove quarantines older than N days
broominal quarantine-cleanup --max-age-days 7 --force
```

### report

Generate a cleanup report from the last scan.

```powershell
# generate report (runs a fresh scan)
broominal report
```

The report is saved as JSON and includes scan results and cleanup statistics.

---

## Configuration

You can customize broominal behavior with flags.

```powershell
# clean only safe items
broominal clean --safe

# allow cleaning danger items
broominal clean --danger

# preview what would be cleaned
broominal clean --dry-run
```

Config file (`%LOCALAPPDATA%\broominal\config.json`):

```json
{
  "enabledCategories": ["temp", "thumbnails", "logs"],
  "oldInstallerMonths": 6,
  "largeFileMinSizeMb": 100,
  "largeFileMonths": 6,
  "oldTempDays": 7,
  "oldExtensionDays": 30,
  "exclusions": [],
  "autoRiskOverrides": {},
  "language": "en",
  "quarantineMaxAgeDays": 30
}
```

---

## Architecture

```
cmd/broominal/   CLI entrypoint (Cobra)

pkg/
  scanner/       file discovery by cleanup category
  cleaner/       quarantine move + report save pipeline
  quarantine/    move, restore, cleanup, JSON manifests
  report/        JSON report generation
  update/        check for and install updates
  config/        JSON configuration and defaults
  doctor/        runtime health checks
  i18n/          English/Russian localization
  style/         ANSI color helpers for CLI output
  util/          size formatting and shared helpers
  types/         shared domain types

internal/
  tui/           Bubbletea interactive interface
```

---

## Philosophy

broominal is intentionally boring. It does not promise performance miracles, registry magic, or hidden optimization. It finds cleanup candidates, classifies risk, shows what it found, and moves selected files into quarantine so the operation can be reversed.

Small packages. Explicit responsibilities. No hidden cleanup magic.

---

## Development

> Enable shared githooks before committing:
> ```powershell
> git config core.hooksPath githooks
> ```

**Hooks**
- `pre-commit` — warns when code changes may need documentation updates
- `commit-msg` — enforces conventional commits

**CI on every push / PR to `main`**
```
gofmt → go vet → golangci-lint → go test ./... → Windows build artifact
```

**Release workflow**
```
git-cliff → build broominal.exe → signed tag → GitHub release + checksums
```

---

## Contributing

Bug reports, cleanup-category ideas, safety improvements, and Windows edge cases are welcome.

See [CONTRIBUTING.md](CONTRIBUTING.md).

---

## License

[MIT](LICENSE) © elev1e1nSure
