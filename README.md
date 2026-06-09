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

---

## Features

| Feature | Description |
|--------|-------------|
| 🧹 **Smart Scan** | Temp, Downloads, Browser Cache, Recycle Bin, Logs, Old Installers, Large Old Files |
| 🛡️ **Risk Levels** | `safe` / `review` / `danger` — system paths and `.sys`/`.dll` files are never touched |
| 🔄 **Undoable** | Every cleanup gets a restore ID; `restore last` brings files back |
| ⚡ **Dry-Run** | `--dry-run` on CLI and `T` key in TUI — see what would be freed without touching files |
| ⚙️ **Config-Driven** | JSON config for thresholds, exclusions, category toggles, and risk overrides |
| 🩺 **Doctor** | Built-in health checks for permissions, manifests, and disk space |
| 🗑️ **Quarantine Cleanup** | Auto-delete old quarantines with `--dry-run` preview |
| 🖥️ **TUI** | Interactive Bubbletea dashboard with per-category file inspection |

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

# Launch interactive TUI
broominal ui

# Clean safe items only
broominal clean --safe

# Simulate cleanup
broominal clean --dry-run

# Restore last cleanup
broominal restore last

# Restore with overwrite
broominal restore last --force-overwrite

# Run health checks
broominal doctor

# Show config
broominal config

# Remove quarantines older than 30 days
broominal quarantine-cleanup --dry-run
broominal quarantine-cleanup --force
```

---

## Architecture

```
cmd/broominal/      CLI entrypoint (Cobra)
pkg/
  scanner/          File discovery by category
  quarantine/       Move / Restore / Cleanup with JSON manifests
  report/           JSON report generation
  risk/             Risk classification (path, extension, config)
  config/           JSON configuration (thresholds, exclusions, overrides)
  types/            Shared domain types
internal/
  tui/              Bubbletea interactive interface
```

## License

[MIT](LICENSE) © elev1e1nSure
