<div align="center">

<h1>Broominal</h1>

<p><strong>Safe, transparent, undoable Windows cleanup</strong></p>

[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Windows-blue?logo=windows)](https://github.com/elev1e1nSure/broominal)

[🇷🇺 Русский](README.ru.md)

</div>

---

## Overview

Broominal is a **Windows cleanup tool** that never permanently deletes your files. Instead of shredding data, it moves items into a local **quarantine** with a JSON manifest — so you can **restore anything** at any time.

- **Safe**: quarantine, not delete
- **Transparent**: plain JSON reports and manifests
- **Reversible**: one command to restore your last cleanup
- **Interactive**: beautiful TUI with category selection and previews
- **Multilingual**: English and Russian with auto-detection

---

## Features

| Feature | Description |
|--------|-------------|
| 🧹 **Smart Scan** | Temp, Downloads, Browser Cache, Recycle Bin, Logs, Old Installers, Large Old Files, Thumbnails, DirectX Shader Cache, Delivery Optimization, Windows Error Reports, Discord Cache, Steam Cache, VSCode Cache, Edge Code Cache, Chrome Code Cache, Firefox Cache2, Old Temp Files, Old .tmp/.log/.bak, Empty Folders, npm Cache, pip Cache, Windows Update Cache, Crash & Memory Dumps, Nvidia Installer Leftovers, Telegram Desktop Cache |
| 🛡️ **Risk Levels** | `safe` / `review` / `danger` — system paths and `.sys`/`.dll` files are never touched |
| 🔄 **Undoable** | Every cleanup gets a restore ID; `restore <id>` brings files back |
| ⚡ **Dry-Run** | `--dry-run` on CLI and `T` key in TUI — see what would be freed without touching files |
| ⚙️ **Config-Driven** | JSON config for thresholds, exclusions, category toggles, risk overrides, and language |
| 🩺 **Doctor** | Built-in health checks for permissions, manifests, and disk space |
| 🗑️ **Quarantine Cleanup** | Auto-delete old quarantines with `--dry-run` preview |
| 🖥️ **TUI** | Interactive Bubbletea interface with Main Menu, per-category inspection, restore picker, doctor, config viewer, language selector |
| 🌐 **i18n** | English / Russian. Auto-detects language by IP on first launch |

---

## Quick Start

### Install

```powershell
# From source (requires Go 1.26+)
go install github.com/elev1e1nSure/broominal/cmd/broominal@latest
```

Or download the latest `.exe` from [Releases](../../releases).

### Usage

```powershell
# Scan safe zones
broominal scan

# Launch interactive TUI (Main Menu → Scan & Clean / Restore / Doctor / Config / Cleanup / Language)
broominal ui

# Clean safe items only
broominal clean --safe

# Simulate cleanup
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

Requires **Go 1.26+**.

```powershell
# Clone and build
go build -o broominal.exe ./cmd/broominal

# Run
.\broominal.exe ui
```

---

## Architecture

```
cmd/broominal/      CLI entrypoint (Cobra)
pkg/
  scanner/          File discovery by category (25+ scan targets)
  cleaner/          Orchestrates quarantine move + report save pipeline
  quarantine/       Move / Restore / Cleanup with JSON manifests
  report/           JSON report generation
  risk/             Risk classification (path, extension, config)
  config/           JSON configuration (thresholds, exclusions, overrides, language)
  doctor/           Health checks (admin, dirs, manifests, stats)
  i18n/             Localization (EN/RU, auto-detect, T-key lookups)
  style/            ANSI color helpers for CLI output
  util/             Size formatting and shared helpers
  types/            Shared domain types
internal/
  tui/              Bubbletea interactive interface (Main Menu → multiple screens)
```

## License

[MIT](LICENSE) © elev1e1nSure
