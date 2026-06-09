# Broominal — Agent Context

## What it is
Broominal is a safe, transparent, undoable Windows cleanup CLI/TUI tool. It never permanently deletes files — instead it moves them to a local quarantine with a manifest, so any cleanup can be restored.

## Core idea
- **Safety first**: quarantine instead of delete, every cleanup is reversible
- **Risk-based**: files are classified as `safe`, `review`, or `danger` based on path, extension, and category
- **Transparency**: scan results, reports, and manifests are plain JSON
- **User control**: interactive TUI lets you inspect and toggle every category before cleanup

## Architecture
```
cmd/broominal/     CLI entrypoint (cobra)
  main.go          scan, ui, clean, restore, report, config, doctor,
                   quarantine-cleanup commands

pkg/
  scanner/         Scan() walks safe zones (Temp, Downloads, Browser Cache,
                   Recycle Bin, Logs, Old Installers, Large Old Files)
  quarantine/      Move() -> quarantine dir + manifest.json
                   Restore() -> move back, handles conflicts
                   Cleanup() -> delete old quarantines
  report/          Save() JSON report with scan + optional clean result
  risk/            Classify() risk level from path/category/config overrides
  config/          Load/Save JSON config: enabled categories, thresholds,
                   exclusions, auto risk overrides (e.g. .git -> review)
  types/           Shared structs: Item, CategorySummary, ScanResult,
                   Manifest, CleanResult, ReportData

internal/
  tui/             Bubbletea TUI: Dashboard -> Categories -> Details ->
                   Confirm -> Cleaning -> Result (with dry-run toggle and
                   restore-conflict screen)
```

## Key design decisions
1. **Quarantine pattern**: files are renamed/moved to `%LOCALAPPDATA%\broominal\quarantine\<uuid>`. A `manifest.json` records original -> quarantined mappings. Restore reverses the mapping.
2. **Dry-run everywhere**: `clean --dry-run`, TUI `T` toggle, `quarantine-cleanup --dry-run`.
3. **Config-driven scanning**: `config.json` controls which categories are enabled, age/size thresholds, exclusions, and risk overrides. Missing config auto-creates with defaults.
4. **Conflict handling on restore**: if the original file already exists, CLI offers `--force-overwrite`; TUI shows an interactive conflict screen.
5. **Doctor command**: lightweight health checks (permissions, manifest integrity, disk space) without heavy dependencies.

## Extension points
- New scanner categories: add to `scanner.go` + config `EnabledCategories` + TUI auto-select logic
- New risk rules: add to `risk.Classify()` or `config.AutoRiskOverrides`
- New TUI screens: add to `Screen` enum, `handleKey()`, and `View()`

## CLI style guide
All terminal output uses ANSI colors via `pkg/style` (`Bold`, `Green`, `Yellow`, `Red`, `Cyan`, `Gray`, `Pass/Warn/Fail`). Cobra help/usage templates are overridden with the same palette.

| Element | Style | Example |
|---------|-------|---------|
| Command names in help | Cyan | `clean`, `scan`, `doctor` |
| Section headers (Usage, Flags, Available Commands) | Bold | **Usage:**, **Flags:** |
| Positive results / success | Green + Bold | `[PASS]`, `Cleaned`, `Restored` |
| Warnings / dry-run | Yellow | `[WARN]`, `[dry-run]` |
| Errors / danger | Red + Bold | `[FAIL]`, `danger` |
| Quantities (size, count, IDs) | Cyan | `4.6 MB`, `restore-id` |
| Descriptions / secondary text | Gray | `Scan safe zones and show summary` |
| Tool name in help | Bold | **broominal** |

Rules:
- Every new CLI output must go through `pkg/style` helpers.
- Always append `Reset` after each color block (handled by helpers).
- Do not add colors to JSON or manifest files.

## Tech stack
- Go 1.26
- Cobra (CLI)
- Bubbletea + Bubbles + Lipgloss (TUI)
