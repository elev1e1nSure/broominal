<div align="center">

# 🧹 broominal

**safe, transparent, undoable Windows cleanup from the terminal**

[![go](https://img.shields.io/badge/go-1.26.3-00ADD8?logo=go\&logoColor=white)](https://go.dev)
[![ci](https://github.com/elev1e1nSure/broominal/actions/workflows/ci.yml/badge.svg)](https://github.com/elev1e1nSure/broominal/actions/workflows/ci.yml)
[![release](https://img.shields.io/github/v/release/elev1e1nSure/broominal?label=release)](https://github.com/elev1e1nSure/broominal/releases)
[![license](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![platform](https://img.shields.io/badge/platform-Windows-0078D4?logo=windows\&logoColor=white)](https://github.com/elev1e1nSure/broominal)

</div>

---

## What

**broominal** cleans Windows by moving files into **quarantine** — not deleting them. Every cleanup is inspectable and restorable.

60 scanner categories across three risk levels (Safe / Review / Danger), interactive TUI, JSON manifests, multilingual (EN/RU).

No fake boost magic. No hidden tweaks. No "trust me bro" cleanup.

---

## Install

```powershell
go install github.com/elev1e1nSure/broominal/cmd/broominal@latest
```

Or download from [releases](https://github.com/elev1e1nSure/broominal/releases).

## Quick Start

```powershell
broominal scan              # find cleanup candidates
broominal clean --safe      # clean safe items
broominal restore last      # undo last cleanup
broominal ui                # interactive TUI
```

## How It Works

```
Scan → Review → Clean → Quarantine → Restore
```

Files go to `%LOCALAPPDATA%\broominal\quarantine\<id>\` with a `manifest.json`. Restore reads the manifest and moves files back.

## Safety

| Level | Behavior |
|-------|----------|
| **Safe** | Auto-selected. Temp, caches, auto-rebuilt data |
| **Review** | Manual select. Downloads, dumps, update cache |
| **Danger** | Never auto-selected. System paths |

Three presets: **Quick** (19 safe categories) → **Standard** (42 total) → **Deep** (60 total). Review items require manual selection regardless of preset.

## Commands

`scan` · `clean` · `restore` · `ui` · `doctor` · `config` · `quarantine-cleanup` · `report` · `path`

See [docs/commands.md](docs/commands.md) for full reference.

## Develop

```powershell
git clone https://github.com/elev1e1nSure/broominal.git
cd broominal
just build
just run ui
```

See [docs/developing.md](docs/developing.md).

## Docs

- [commands](docs/commands.md) — CLI reference
- [categories](docs/categories.md) — full list with paths and risks
- [safety](docs/safety.md) — quarantine, risk model, security
- [developing](docs/developing.md) — setup, architecture, contributing

## License

[MIT](LICENSE) © elev1e1nSure
