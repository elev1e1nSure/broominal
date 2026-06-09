<div align="center">

# broominal

safe, transparent, undoable windows cleanup from the terminal.

<br>

[![go](https://img.shields.io/badge/go-1.26.3-00ADD8?logo=go\&logoColor=white)](https://go.dev)
[![ci](https://github.com/elev1e1nSure/broominal/actions/workflows/ci.yml/badge.svg)](https://github.com/elev1e1nSure/broominal/actions/workflows/ci.yml)
[![release](https://img.shields.io/github/v/release/elev1e1nSure/broominal?label=release)](https://github.com/elev1e1nSure/broominal/releases)
[![license](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![platform](https://img.shields.io/badge/platform-Windows-0078D4?logo=windows\&logoColor=white)](https://github.com/elev1e1nSure/broominal)

<br>

[english](README.md) · [русский](README.ru.md) · [releases](../../releases) · [contributing](CONTRIBUTING.md)

</div>

---

## overview

**broominal** is a windows cleanup cli/tui built around one rule:

> cleanup must be reversible.

instead of permanently deleting files, broominal moves selected items into a local quarantine, stores json manifests, and makes every cleanup inspectable and restorable.

no fake boost magic. no hidden system tweaking. no “trust me bro” cleanup.

---

## why broominal

| principle           | what it means                                                  |
| ------------------- | -------------------------------------------------------------- |
| **safe by default** | files are moved to quarantine instead of being deleted         |
| **transparent**     | scan results, reports, and manifests are plain json            |
| **undoable**        | restore a cleanup by id or restore the latest cleanup          |
| **predictable**     | categories, risk levels, and exclusions are explicit           |
| **interactive**     | bubbletea tui for scan, preview, dry-run, cleanup, and restore |
| **multilingual**    | english and russian with first-run auto-detection              |

---

## features

| feature                | description                                                      |
| ---------------------- | ---------------------------------------------------------------- |
| **smart scan**         | 25+ cleanup categories for windows and common apps               |
| **risk levels**        | `safe`, `review`, and `danger` classification                    |
| **undoable cleanup**   | every cleanup gets a restore id                                  |
| **dry-run mode**       | preview cleanup without moving files                             |
| **json config**        | thresholds, exclusions, toggles, risk overrides, and language    |
| **doctor**             | checks permissions, directories, manifests, and quarantine state |
| **quarantine cleanup** | removes old quarantines with preview and confirmation            |
| **tui**                | interactive terminal interface powered by bubbletea              |
| **i18n**               | english and russian localization                                 |

---

## cleanup categories

broominal scans common temporary and cache locations, including:

| group                 | examples                                                                   |
| --------------------- | -------------------------------------------------------------------------- |
| **windows**           | temp, recycle bin, thumbnails, directx shader cache, delivery optimization |
| **system reports**    | windows error reports, crash dumps, old logs                               |
| **browsers**          | edge, chrome, firefox caches                                               |
| **apps**              | discord, steam, vscode, telegram caches                                    |
| **developer tools**   | npm cache, pip cache                                                       |
| **files**             | old `.tmp`, `.log`, `.bak`, installers, large old files, empty folders     |
| **drivers / updates** | windows update cache, nvidia leftovers                                     |

---

## safety model

> [!IMPORTANT]
> safe cleanup is selected by default. review cleanup requires manual choice. danger items are never cleaned automatically.

| risk     | default behavior            | examples                                                    |
| -------- | --------------------------- | ----------------------------------------------------------- |
| `safe`   | selected by default         | temp files, thumbnails, shader cache, common app caches     |
| `review` | user must select manually   | downloads, dumps, windows update cache, telegram cache      |
| `danger` | never cleaned automatically | system paths, protected extensions, unknown risky locations |

files are moved to:

```text
%LOCALAPPDATA%\broominal\quarantine\<restore-id>
```

each cleanup stores a `manifest.json`, mapping original paths to quarantined paths.

---

## quick start

### install from source

```powershell
go install github.com/elev1e1nSure/broominal/cmd/broominal@latest
```

requires **go 1.26.3+**.

or download the latest `.exe` from [releases](../../releases).

---

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

# remove quarantines older than 7 days
broominal quarantine-cleanup --force --max-age-days 7
```

---

## build from source

```powershell
git clone https://github.com/elev1e1nSure/broominal.git
cd broominal

go build -o broominal.exe ./cmd/broominal

.\broominal.exe ui
```

---

## architecture

```text
cmd/broominal/
  cli entrypoint

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

---

## design philosophy

broominal is intentionally boring.

it does not promise performance miracles, registry magic, or hidden optimization. it only finds cleanup candidates, classifies risk, shows what it found, and moves selected files into quarantine so the operation can be reversed.

small packages. explicit responsibilities. no hidden cleanup magic.

---

## development

### githooks

enable shared hooks:

```powershell
git config core.hooksPath githooks
```

included hooks:

| hook         | purpose                                                |
| ------------ | ------------------------------------------------------ |
| `pre-commit` | warns when code changes may need documentation updates |
| `commit-msg` | enforces conventional commits                          |

supported commit types:

```text
feat, fix, chore, refactor, docs, test, build, ci, perf, style, revert
```

---

## ci / cd

pushes and pull requests to `main` run:

```text
gofmt
go vet
golangci-lint
go test ./...
windows build artifact upload
```

---

## release

the release workflow:

```text
1. generates release notes from conventional commits via git-cliff
2. builds broominal.exe
3. creates a signed tag
4. publishes a github release
5. uploads checksums.txt
```

---

## contributing

bug reports, cleanup-category ideas, safety improvements, and windows edge cases are welcome.

see [CONTRIBUTING.md](CONTRIBUTING.md).

---

## license

[mit](LICENSE) © elev1e1nSure
