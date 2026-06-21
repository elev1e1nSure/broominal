# Safety Model

## Core Rule

> [!IMPORTANT]
> Cleanup must be **reversible**.

Files are never deleted permanently by default. They are moved to a local **quarantine** directory with a JSON manifest. Any cleanup can be inspected and restored.

## Risk Levels

Every scanned file gets a risk level based on its category. These levels serve as visual guides in the TUI (green/yellow/red) to help the user identify potentially sensitive areas. All categories enabled by the active preset are scanned and selected for cleanup by default, but the user can always review and deselect them before execution.

| Level | Behavior / Description |
|-------|----------|
| **Safe** | Temp files, caches, and automatically-rebuilt data. |
| **Review** | Downloads, dumps, update caches — check before removing. |
| **Danger** | System paths, protected extensions. |

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

> [!CAUTION]
> If `QuarantineEnabled = false` in config, `os.RemoveAll` is called directly.
> **No manifest is created, and restoration is not possible.**
> A warning is shown in the TUI.

## Security Guards

- **Write-Ahead Log (WAL)** — every quarantine operation is recorded to `journal.jsonl`.
- **Background Auto-Repair** — if the application crashes or loses power during cleanup, the background repair system (`RepairState()`) automatically verifies journal transactions on next launch, purging corrupted partial data and reverting incomplete manifests.
- **Strict Atomic Copying** — disk space is pre-checked via `GetDiskFreeSpaceEx` and copied files are verified against `expectedSize` before the original file is removed.
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
| **Standard** | 43 total | Adds app caches, logs, dev tool caches. Some Review. |
| **Deep** | 58 total | Adds user data areas, system cleanup. More Review items. |

All categories enabled by the preset are selected by default. The user can review the category details and manually deselect any items before initiating the cleanup.
