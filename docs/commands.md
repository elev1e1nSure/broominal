# CLI Commands

## scan

Scan the system for cleanup candidates across all enabled categories.

```powershell
broominal scan
```

Results are saved as JSON. No files are modified.

## clean

Execute cleanup — move selected items to quarantine (or delete if quarantine is disabled).

```powershell
broominal clean                # safe + review (default)
broominal clean --safe          # safe only
broominal clean --danger        # safe + review + danger
```

## restore

Restore files from quarantine.

```powershell
broominal restore <id>         # restore by ID
broominal restore last         # restore the latest cleanup
broominal restore <id> --force-overwrite  # overwrite existing files
```

## ui

Launch the interactive TUI. Walk through scan → select → clean with visual feedback.

```powershell
broominal ui
```

## doctor

Run environment health checks. Verifies that broominal can write to its directories and that quarantines are intact.

```powershell
broominal doctor
```

## config

Show current configuration.

```powershell
broominal config
```

## quarantine-cleanup

Permanently delete quarantine entries.

```powershell
broominal quarantine-cleanup                  # preview
broominal quarantine-cleanup --force          # delete all
broominal quarantine-cleanup --max-age-days 7 --force  # older than N days
```

## report

Generate a JSON report from a scan. Includes per-category breakdown.

```powershell
broominal report
```

## path

Add or remove broominal from the system PATH.

```powershell
broominal path add
broominal path remove
```

## Global Flags

| Flag | Description |
|------|-------------|
| `--json-logs` | Output logs in JSON format (for structured logging pipelines) |
