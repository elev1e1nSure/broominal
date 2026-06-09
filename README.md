<div align="center">

# 🧹 broominal

**safe, transparent, undoable windows cleanup from the terminal**

[![go](https://img.shields.io/badge/go-1.26.3-00ADD8?logo=go\&logoColor=white)](https://go.dev)
[![ci](https://github.com/elev1e1nSure/broominal/actions/workflows/ci.yml/badge.svg)](https://github.com/elev1e1nSure/broominal/actions/workflows/ci.yml)
[![release](https://img.shields.io/github/v/release/elev1e1nSure/broominal?label=release)](https://github.com/elev1e1nSure/broominal/releases)
[![license](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![platform](https://img.shields.io/badge/platform-Windows-0078D4?logo=windows\&logoColor=white)](https://github.com/elev1e1nSure/broominal)

[english](README.md) · [русский](README.ru.md) · [releases](../../releases) · [contributing](CONTRIBUTING.md)

</div>

## what is it

**broominal** is a windows cleanup cli/tui built around one rule:

> cleanup must be **reversible**.

instead of permanently deleting files, broominal moves selected items into a local **quarantine**, stores json manifests, and makes every cleanup inspectable and restorable.

no fake boost magic. no hidden system tweaking. no “trust me bro” cleanup.

## highlights

- 🛡️ **safe by default** — files are quarantined, not deleted
- 🔍 **transparent** — scan results, reports, and manifests are plain json
- ↩️ **undoable** — restore any cleanup by id or restore the latest one
- 📋 **predictable** — explicit categories, risk levels, and exclusions
- 🎛️ **interactive** — bubbletea tui for scan, preview, dry-run, and restore
- 🌍 **multilingual** — english and russian with first-run auto-detection
- ⚡ **25+ categories** — temp, caches, logs, browser data, dev tools, and more
- 🧪 **dry-run everywhere** — preview before touching a single file
- 🩺 **doctor** — lightweight health checks for permissions, manifests, and state

## safety model

> [!IMPORTANT]
> safe cleanup is selected by default. review cleanup requires manual choice. danger items are never cleaned automatically.

| risk | default behavior | examples |
|------|------------------|----------|
| `safe` | selected by default | temp files, thumbnails, shader cache, common app caches |
| `review` | user must select manually | downloads, dumps, windows update cache, telegram cache |
| `danger` | never cleaned automatically | system paths, protected extensions, unknown risky locations |

files are moved to `%LOCALAPPDATA%\broominal\quarantine\<restore-id>` with a `manifest.json` mapping original paths to quarantined paths.

## quick start

```powershell
# install from source (requires go 1.26.3+)
go install github.com/elev1e1nSure/broominal/cmd/broominal@latest

# ...or grab the latest .exe from releases
```

## usage

```powershell
# scan safe zones
broominal scan

# launch interactive tui
broominal ui

# clean safe items only
broominal clean --safe

# simulate cleanup without moving files
broominal clean --dry-run

# restore a specific cleanup
broominal restore <id>

# restore with overwrite
broominal restore <id> --force-overwrite

# run health checks
broominal doctor

# show config
broominal config

# preview old quarantine cleanup
broominal quarantine-cleanup --dry-run

# remove quarantines older than 30 days
broominal quarantine-cleanup --force
```

## build from source

```powershell
git clone https://github.com/elev1e1nSure/broominal.git
cd broominal

go build -o broominal.exe ./cmd/broominal
.\broominal.exe ui
```

## architecture

```text
cmd/broominal/   cli entrypoint

pkg/
  scanner/       file discovery by cleanup category
  cleaner/       quarantine move + report save pipeline
  quarantine/    move, restore, cleanup, json manifests
  report/        json report generation
  risk/          risk classification from paths, extensions, config
  config/        json configuration and defaults
  doctor/        runtime health checks
  i18n/          english/russian localization
  style/         ansi color helpers for cli output
  util/          size formatting and shared helpers
  types/         shared domain types

internal/
  tui/           bubbletea interactive interface
```

## philosophy

broominal is intentionally boring. it does not promise performance miracles, registry magic, or hidden optimization. it finds cleanup candidates, classifies risk, shows what it found, and moves selected files into quarantine so the operation can be reversed.

small packages. explicit responsibilities. no hidden cleanup magic.

## development

> [!TIP]
> enable shared githooks before committing:
>
> ```powershell
> git config core.hooksPath githooks
> ```

**hooks:**
- `pre-commit` — warns when code changes may need documentation updates
- `commit-msg` — enforces conventional commits (`feat`, `fix`, `chore`, `refactor`, `docs`, `test`, `build`, `ci`, `perf`, `style`, `revert`)

**ci on every push / pr to `main`:**
```text
gofmt → go vet → golangci-lint → go test ./... → windows build artifact
```

**release workflow:**
git-cliff release notes → build `broominal.exe` → signed tag → github release + checksums

## contributing

bug reports, cleanup-category ideas, safety improvements, and windows edge cases are welcome.

see [CONTRIBUTING.md](CONTRIBUTING.md).

## license

[mit](LICENSE) © elev1e1nSure
