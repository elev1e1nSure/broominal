# Safety Model

## Core Rule

> Cleanup must be **reversible**.

Files are never deleted permanently by default. They are moved to a local **quarantine** directory with a JSON manifest. Any cleanup can be inspected and restored.

## Risk Levels

Every scanned file gets a risk level based on its category.

| Level | Behavior |
|-------|----------|
| **Safe** | Auto-selected for cleanup. Temp files, caches, and automatically-rebuilt data. |
| **Review** | Must be manually selected. Downloads, dumps, update caches — check before removing. |
| **Danger** | Never auto-selected. System paths, protected extensions. |

Review categories exist in Standard and Deep presets. Quick preset contains only Safe categories.

## Quarantine

Files are moved to `%LOCALAPPDATA%\broominal\quarantine\<restore-id>\` with a `manifest.json`:

```json
{
  "id": "2025-06-01-120000",
  "created_at": "...",
  "items": [
    {
      "original": "C:\\Users\\...\\cache\\f_000001",
      "quarantined": "C:\\Users\\...\\quarantine\\2025-06-01-120000\\f_000001",
      "size": 4096
    }
  ]
}
```

Restore reverses the mapping. If the original path already exists, the TUI shows a conflict screen (overwrite / skip).

### When Quarantine Is Disabled

If `QuarantineEnabled = false` in config, `os.RemoveAll` is called directly. **No manifest is created, no restoration is possible.** A warning is shown in the TUI.

## Security Guards

- **Symlinks are skipped** — never followed during quarantine
- **Path traversal blocked** — restore IDs validated, paths checked against allowed roots
- **Atomic writes** — manifests written to `.tmp` then renamed
- **No telemetry** — the only outbound call is GitHub release check
- **No silent actions** — every destructive action requires explicit confirmation
- **Config local only** — all data lives under `%LOCALAPPDATA%\broominal\`

## Age Filters

Some categories apply age-based filters to avoid touching active data:

- **Crash & Memory Dumps** — only files older than 7 days
- **Printer Spooler** — only print jobs older than 1 hour

## Presets

| Preset | Categories | Behavior |
|--------|-----------|----------|
| **Quick** | 19 | Temp + browser + GPU caches. All Safe. |
| **Standard** | 42 total | Adds app caches, logs, dev tool caches. Some Review. |
| **Deep** | 62 total | Adds user data areas, system cleanup. More Review items. |

All Review categories require manual selection regardless of preset. The preset only controls which categories appear — it does not auto-select Review items.
