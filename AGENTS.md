# Broominal — Agent Context

## What it is
Broominal is a safe, transparent, undoable Windows cleanup CLI/TUI tool. It never permanently deletes files — instead it moves them to a local quarantine with a manifest, so any cleanup can be restored.

## Core idea
- **Safety first**: quarantine instead of delete, every cleanup is reversible
- **Risk-based**: files are classified as `safe`, `review`, or `danger` based on path, extension, and category
- **Transparency**: scan results, reports, and manifests are plain JSON
- **User control**: interactive TUI lets you inspect and toggle every category before cleanup
- **Multilingual**: English and Russian with auto-detection by IP on first launch

## Architecture
```
cmd/broominal/     CLI entrypoint (cobra)
  main.go          scan, ui, clean, restore, report, config, doctor,
                   quarantine-cleanup commands

pkg/
  scanner/         Scan() walks 25+ safe zones (Temp, Downloads, Browser Cache,
                   Recycle Bin, Logs, Old Installers, Large Old Files,
                   Thumbnails, DirectX Shader Cache, Delivery Optimization,
                   WER, Discord Cache, Steam Cache, VSCode Cache, Edge Code Cache,
                   Chrome Code Cache, Firefox Cache2, Old Temp Files,
                   Old .tmp/.log/.bak, Empty Folders, npm Cache, pip Cache,
                   Windows Update Cache, Crash & Memory Dumps,
                   Nvidia Installer Leftovers, Telegram Desktop Cache)
  quarantine/      Move() -> quarantine dir + manifest.json
                   Restore() -> move back, handles conflicts
                   Cleanup() -> delete old quarantines
  report/          Save() JSON report with scan + optional clean result
  risk/            Classify() risk level from path/category/config overrides
  config/          Load/Save JSON config: enabled categories, thresholds
                   (old_installer_months, large_file_min_size_mb, large_file_months,
                   old_temp_days, old_extension_days), exclusions, auto risk
                   overrides, language, quarantine_max_age_days. Missing config
                   auto-creates with defaults and merges missing categories.
  doctor/          Run() health checks: admin, dirs, manifests, quarantine stats
  i18n/            SetLanguage(), T(key), DetectFromIP(), SupportedLanguages()
  style/           ANSI color helpers: Boldf, Greenf, Yellowf, Redf, Cyanf,
                   Grayf, Passf, Warnf, Failf
  types/           Shared structs: Item, CategorySummary, ScanResult,
                   Manifest, CleanResult, ReportData

internal/
  tui/             Bubbletea TUI: Main Menu -> Scan & Clean / Restore /
                   Doctor / Config / Quarantine Cleanup / Language
                   Scan flow: Dashboard -> Categories -> Details ->
                   Confirm -> Cleaning -> Result (with dry-run toggle,
                   restore-conflict screen, and error screen)
```

## Key design decisions
1. **Quarantine pattern**: files are renamed/moved to `%LOCALAPPDATA%\broominal\quarantine\<uuid>`. A `manifest.json` records original -> quarantined mappings. Restore reverses the mapping.
2. **Dry-run everywhere**: `clean --dry-run`, TUI `T` toggle, `quarantine-cleanup --dry-run`.
3. **Config-driven scanning**: `config.json` controls which categories are enabled, age/size thresholds, exclusions, risk overrides, and language. Missing config auto-creates with defaults and merges missing categories into existing configs.
4. **Conflict handling on restore**: if the original file already exists, CLI offers `--force-overwrite`; TUI shows an interactive conflict screen.
5. **Doctor command**: lightweight health checks (admin rights, directory write access, manifest integrity, quarantine stats) without heavy dependencies.
6. **i18n**: embedded translation map (EN/RU). First run auto-detects language via ipapi.co and persists choice in config.
7. **TUI error handling**: errors do not quit the app immediately. A dedicated `ScreenError` displays the message; user presses Q/Esc to exit.

## Extension points
- New scanner categories: add to `scanner.go` + config `EnabledCategories` + TUI auto-select logic
- New risk rules: add to `risk.Classify()` or `config.AutoRiskOverrides`
- New TUI screens: add to `Screen` enum, `handleKey()`, and `View()`
- New i18n strings: add to `pkg/i18n/i18n.go` translations map for all supported languages

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
